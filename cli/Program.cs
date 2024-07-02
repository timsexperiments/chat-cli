using System.CommandLine;
using System.CommandLine.Builder;
using System.CommandLine.Invocation;
using System.CommandLine.Parsing;
using System.Net;

using Spectre.Console;

using TimsExperiments.ChatCli.Chat;

using var client = new ConversationClient();

RootCommand rootCommand = new("ChatCLI - an AI chat application.");

var idArgument = new Argument<string?>("id", () => null, "The ID of the chat.");

var noInteractiveOption = new Option<bool>("--no-interactive", "Disable interactive mode.");
noInteractiveOption.AddAlias("--no-interaction");
noInteractiveOption.AddAlias("-ni");

var messageOption = new Option<string?>("--message", "The message to send to the chat.");
messageOption.AddAlias("-m");

var titleArgument = new Argument<string>("title", "The title of the chat.");

var tokenOption = new Option<string>("--token", "Sets the OpenAI API token.");

var messageArgument = new Argument<string>("message", "The message to send to the chat.");

Command chatCommand = new("chat", "Chat with an AI.");
chatCommand.AddArgument(idArgument);
chatCommand.AddOption(noInteractiveOption);
chatCommand.SetHandler(ChatHandler, idArgument, noInteractiveOption);

Command newChatCommand = new("new", "Creates a new Chat with the given title.");
newChatCommand.AddArgument(titleArgument);
newChatCommand.SetHandler(NewChatHandler, titleArgument);

Command messagesCommand = new("messages", "Adds a message to a chat");
messagesCommand.AddArgument(idArgument);
messagesCommand.AddOption(messageOption);
messagesCommand.SetHandler(MessagesHandler, idArgument, messageOption);

Command listMessagesCommand = new("list", "List all messages in a chat.");
listMessagesCommand.AddAlias("ls");
listMessagesCommand.AddArgument(idArgument);
listMessagesCommand.SetHandler(ListMessagesHandler, idArgument);

Command listChatCommand = new("list", "List all chats available on the server.");
listChatCommand.AddAlias("ls");
listChatCommand.SetHandler(ListChatHandler);

rootCommand.AddGlobalOption(tokenOption);
rootCommand.AddCommand(chatCommand);
chatCommand.AddCommand(listChatCommand);
chatCommand.AddCommand(messagesCommand);
chatCommand.AddCommand(newChatCommand);
messagesCommand.AddCommand(listMessagesCommand);

var builder = new CommandLineBuilder(rootCommand);
builder.UseDefaults();
builder.AddMiddleware(CheckToken);

var parser = builder.Build();
await parser.InvokeAsync(args);

async Task ChatHandler(string? id, bool noInteractive)
{
    if (noInteractive)
    {
        var strId = GetStringId(id);
        var conv = await client.Get(strId);
        AnsiConsole.Write(Cli.CreateConversationTable([conv], true));
        return;
    }

    Conversation conversation = id != null ? await client.Get(id.ToString()!) : await ChooseConversation();
    var quit = false;
    await client.Connect(conversation.Id.ToString(), new ConnectionCallbacks
    {
        OnMessage = async (connection, ev) =>
        {
            switch (ev.Type)
            {
                case ChatEvent.Types.Type.Message:
                    AnsiConsole.WriteLine();
                    AnsiConsole.MarkupLine($"[bold]AI:[/] {ev.Message.Body}\n");
                    break;
                case ChatEvent.Types.Type.Error:
                    AnsiConsole.WriteLine($"[red]Error: {ev.Error.Message}[/]\n");
                    break;
            }
            await PromptUser(connection);
        },
        OnConnect = async (connection) =>
        {
            AnsiConsole.MarkupLine($"You are now connected to [bold]{conversation.Title}[/]. Type 'exit' to exit the chat.");
            AnsiConsole.MarkupLine("Type your message and press [bold]Enter[/] to send it.");
            AnsiConsole.WriteLine();
            if (!string.IsNullOrWhiteSpace(conversation.Context))
            {
                AnsiConsole.MarkupLine($"Previously on [bold]{conversation.Title}[/]: {conversation.Context}");
                AnsiConsole.WriteLine();
            }
            await PromptUser(connection);
        },
        OnDisconnected = () =>
        {
            AnsiConsole.WriteLine("Disconnected from WebSocket");
            quit = true;
        },
        OnError = (_, error) =>
        {
            AnsiConsole.MarkupLine($"[red]Error: {error.Message}[/]");
            quit = true;
        }
    });

    while (!quit)
    {
        await Task.Delay(1000);
    }

    async Task PromptUser(ConversationClient.WebSocketConnection connection)
    {
        var input = AnsiConsole.Prompt(new TextPrompt<string>("[bold]You:[/]"));
        if (input == "exit")
        {
            await connection.Disconnect();
        }
        else
        {
            await connection.SendMessage(input);
        }
    }
}

async Task ListChatHandler()
{
    var response = await client.List();
    var table = Cli.CreateConversationTable(response.Conversations);
    AnsiConsole.Write(table);
}

async Task MessagesHandler(string? id, string? messageBody)
{
    Conversation conversation = id != null ? await client.Get(id) : await ChooseConversation();
    messageBody = GetMessageBody(messageBody);
    var message = await client.CreateMessage(conversation.Id.ToString(), messageBody);
    AnsiConsole.MarkupLine($"Message was successfully sent.");
    AnsiConsole.Write(Cli.CreateMessagesTable([message]));
}

async Task ListMessagesHandler(string? id)
{
    var conversation = id != null ? await client.Get(id) : await ChooseConversation();
    var messages = await client.ListMessages(conversation.Id.ToString());
    AnsiConsole.Write(Cli.CreateMessagesTable(messages));
}

async Task NewChatHandler(string? title)
{
    var conversation = await client.Create(title ?? AnsiConsole.Ask<string>("Enter the title of the chat:"));
    AnsiConsole.MarkupLine("Chat created with ID [bold]{0}[/]", conversation.Id);
    AnsiConsole.Write(Cli.CreateConversationTable([conversation]));
}

async Task CheckToken(InvocationContext context, Func<InvocationContext, Task> next)
{
    if (context.ParseResult.HasOption(tokenOption))
    {
        var token = context.ParseResult.GetValueForOption(tokenOption);
        ConfigurationManager.SetToken(token ?? "");
    }
    await next(context);
}

static string GetStringId(string? id)
{
    return id?.ToString() ?? AnsiConsole.Ask<string>("Enter the Title or ID of the chat (use [italic bold]chatcli chat ls[/] to see a list of chats):");
}

static string GetMessageBody(string? messageBody)
{
    return messageBody ?? AnsiConsole.Ask<string>("Enter the message to send:");
}

async Task<Conversation> ChooseConversation()
{
    var response = await client.List();

    var choices = response.Conversations.Select(c => new ConversationListItem(c)).ToList();
    var option = new ConversationListItem(null) { Display = "Create a new chat" };
    choices.Add(option);

    var conversationListItem = AnsiConsole.Prompt(new SelectionPrompt<ConversationListItem>()
        .Title("Select the conversation you want to connect with:")
        .PageSize(10)
        .MoreChoicesText("[grey](Move up and down to reveal more conversations)[/]")
        .AddChoices(choices));
    AnsiConsole.MarkupLine("Connecting to [bold]{0}[/]", conversationListItem);
    return conversationListItem.Conversation ?? await client.Create(AnsiConsole.Ask<string>("Enter the title of the chat to create:"));
}

internal class ConversationListItem(Conversation? conversation)
{
    private string? _display;

    public string Display { set => _display = value; }

    public Conversation? Conversation { get; } = conversation;

    public override string ToString()
    {
        return _display ?? Conversation?.Title ?? "";
    }
}
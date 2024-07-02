using Spectre.Console;
using TimsExperiments.ChatCli.Chat;

internal static class Cli
{
    public static Table CreateConversationTable(IEnumerable<Conversation> conversations)
    {
        return CreateConversationTable(conversations, null);
    }

    public static Table CreateConversationTable(IEnumerable<Conversation> conversations, bool? withMessages)
    {
        Table table = new();

        table.AsciiBorder();

        table.AddColumn("ID")
            .AddColumn("Title")
            .AddColumn("Context")
            .AddColumn("Created At");

        if (withMessages == true)
        {
            table.AddColumn("Messages");
        }

        foreach (var conversation in conversations)
        {
            var rowValues = new List<string>
            {
                conversation.Id.ToString(),
                conversation.Title,
                conversation.Context,
                conversation.CreatedAt.ToDateTime().ToString("yyyy-MM-dd HH:mm:ss")
            };
            if (withMessages == true)
            {
                rowValues.Add(conversation.Messages.Count.ToString());
            }
            table.AddRow(rowValues.ToArray());
        }
        return table;
    }

    public static Table CreateMessagesTable(IEnumerable<Message> messages)
    {
        Table table = new();

        table.AsciiBorder();

        table.AddColumn("ID")
            .AddColumn("Sender")
            .AddColumn("Body")
            .AddColumn("Created At");

        foreach (var message in messages)
        {
            table.AddRow(
                message.Id.ToString(),
                message.Sender.ToString(),
                message.Body,
                message.CreatedAt.ToDateTime().ToString("yyyy-MM-dd HH:mm:ss"));
        }
        return table;

    }
}
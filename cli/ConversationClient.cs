using System.Net.WebSockets;
using Google.Protobuf;
using TimsExperiments.ChatCli.Chat;
using Spectre.Console;
using System.Net.Http.Headers;

internal class ConversationClient : IDisposable
{
    private static readonly string DefaultApiUrl = "http://localhost:8080";

    private readonly HttpClient _httpClient;

    private readonly Uri _baseUri;

    public ConversationClient()
    {
        var baseAddress = Environment.GetEnvironmentVariable("CHAT_CLI_API_URL");
        if (string.IsNullOrEmpty(baseAddress))
        {
            baseAddress = DefaultApiUrl;
        }
        try
        {
            _baseUri = new Uri(baseAddress);
        }
        catch (UriFormatException)
        {
            AnsiConsole.WriteLine("Invalid API_URL: {0}", baseAddress);
            AnsiConsole.WriteLine("Using default API_URL: {0}", DefaultApiUrl);
            _baseUri = new Uri(DefaultApiUrl);
        }
        _httpClient = new()
        {
            BaseAddress = _baseUri,
            DefaultRequestHeaders = {
                { "Accept", "application/protobuf" }
            }
        };
    }

    public async Task<ListConversationsResponse> List()
    {
        var response = await _httpClient.GetAsync("/conversations");
        response.EnsureSuccessStatusCode();

        var responseStream = await response.Content.ReadAsStreamAsync();
        return ListConversationsResponse.Parser.ParseFrom(responseStream);
    }

    public async Task<Message[]> ListMessages(string conversationId)
    {
        var conversation = await Get(conversationId);
        return [.. conversation.Messages];
    }

    public async Task<Conversation> Create(string title)
    {
        CreateConversationRequest request = new()
        {
            Title = title
        };
        var requestBytes = request.ToByteArray();
        var response = await _httpClient.PostAsync("/conversations", new ByteArrayContent(requestBytes)
        {
            Headers = { ContentType = new MediaTypeHeaderValue("application/protobuf") }
        });
        response.EnsureSuccessStatusCode();

        var responseStream = await response.Content.ReadAsStreamAsync();
        return Conversation.Parser.ParseFrom(responseStream);
    }

    public async Task<Message> CreateMessage(string conversationId, string message)
    {
        CreateMessageRequest request = new()
        {
            Body = message
        };
        var requestBytes = request.ToByteArray();
        var token = ConfigurationManager.EnsureToken();
        var requestMessage = new HttpRequestMessage(HttpMethod.Post, $"/conversations/{conversationId}/messages")
        {
            Content = new ByteArrayContent(requestBytes)
            {
                Headers = { ContentType = new MediaTypeHeaderValue("application/protobuf") }
            },
            Headers = { Authorization = new("Bearer", token) }
        };
        var response = await _httpClient.SendAsync(requestMessage);
        response.EnsureSuccessStatusCode();

        var responseStream = await response.Content.ReadAsStreamAsync();
        return Message.Parser.ParseFrom(responseStream);
    }

    public async Task<Conversation> Get(string id)
    {
        var request = new HttpRequestMessage(HttpMethod.Get, $"/conversations/{id}");
        request.Headers.Authorization = new("Bearer", ConfigurationManager.EnsureToken());
        var response = await _httpClient.SendAsync(request);
        response.EnsureSuccessStatusCode();
        var responseStream = await response.Content.ReadAsStreamAsync();
        return Conversation.Parser.ParseFrom(responseStream);
    }

    public async Task Connect(string id, ConnectionCallbacks callbacks)
    {
        var baseWebsocketUri = new UriBuilder(_baseUri)
        {
            Scheme = _baseUri.Scheme == Uri.UriSchemeHttp ? Uri.UriSchemeWs : Uri.UriSchemeWss,
        }.Uri;
        var token = ConfigurationManager.EnsureToken();
        var websocketUri = new Uri(baseWebsocketUri, $"/conversations/{id}?api_secret={token}");
        var clientWebSocket = new ClientWebSocket();

        try
        {
            await clientWebSocket.ConnectAsync(websocketUri, CancellationToken.None);
            _ = ReceiveMessages(clientWebSocket, callbacks?.OnMessage, callbacks?.OnConnect, callbacks?.OnDisconnected, callbacks?.OnError);
        }
        catch (Exception ex)
        {
            clientWebSocket.Dispose();
            throw new WebsocketConnectionException("Failed to connect to WebSocket", ex);
        }
    }

    private static async Task ReceiveMessages(ClientWebSocket clientWebSocket, MessageCallback? onMessage, ConnectCallback? onConnect, DisconnectCallback? onDisconnected, ErrorCallback? onError)
    {
        var connection = new WebSocketConnection(clientWebSocket);

        if (clientWebSocket.State == WebSocketState.Open)
        {
            onConnect?.Invoke(connection);
        }

        while (true)
        {
            try
            {
                var segment = new ArraySegment<byte>(new byte[1024 * 4]);
                WebSocketReceiveResult result;
                int totalBytes = 0;
                do
                {
                    result = await clientWebSocket.ReceiveAsync(segment, CancellationToken.None);
                    totalBytes += result.Count;
                } while (!result.EndOfMessage);

                if (result.MessageType == WebSocketMessageType.Close)
                {
                    await clientWebSocket.CloseAsync(WebSocketCloseStatus.NormalClosure, string.Empty, CancellationToken.None);
                    onDisconnected?.Invoke();
                    clientWebSocket.Dispose();
                }
                else
                {
                    var message = ChatEvent.Parser.ParseFrom(segment[0..totalBytes]);
                    onMessage?.Invoke(connection, message);
                }
            }
            catch (Exception ex)
            {
                onError?.Invoke(connection, ex);
                if (clientWebSocket.State != WebSocketState.Open)
                {
                    onDisconnected?.Invoke();
                    clientWebSocket.Dispose();
                    connection.Dispose();
                }
                break;
            }
        }
    }

    public void Dispose()
    {
        _httpClient.Dispose();
    }

    public sealed class WebSocketConnection(ClientWebSocket webSocket) : IDisposable
    {
        public async Task SendMessage(string message)
        {
            if (webSocket.State == WebSocketState.Open)
            {
                var messageEvent = new MessageEvent
                {
                    Body = message
                };
                var messageBytes = messageEvent.ToByteArray();
                var messageSegment = new ArraySegment<byte>(messageBytes);
                await webSocket.SendAsync(messageSegment, WebSocketMessageType.Text, true, CancellationToken.None);
            }
            else
            {
                throw new InvalidOperationException("WebSocket connection is not open.");
            }
        }

        public async Task Disconnect()
        {
            await webSocket.CloseAsync(WebSocketCloseStatus.NormalClosure, "Done", CancellationToken.None);
        }

        public void Dispose()
        {
            webSocket.Dispose();
        }
    }

}

internal class ConnectionCallbacks
{
    public MessageCallback? OnMessage { get; set; }
    public ConnectCallback? OnConnect { get; set; }
    public DisconnectCallback? OnDisconnected { get; set; }
    public ErrorCallback? OnError { get; set; }
}

internal delegate void MessageCallback(ConversationClient.WebSocketConnection connection, ChatEvent ev);
internal delegate Task MessageCallbackAsync(ConversationClient.WebSocketConnection connection, ChatEvent ev);

internal delegate void ConnectCallback(ConversationClient.WebSocketConnection connection);
internal delegate Task ConnectCallbackAsync(ConversationClient.WebSocketConnection connection);

internal delegate void DisconnectCallback();
internal delegate Task DisconnectCallbackAsync();

internal delegate void ErrorCallback(ConversationClient.WebSocketConnection connection, Exception ex);
internal delegate Task ErrorCallbackAsync(ConversationClient.WebSocketConnection connection, Exception ex);
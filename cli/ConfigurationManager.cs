internal class ConfigurationManager
{
    private static readonly Lazy<ConfigurationManager> LazyInstance = new(() => new ConfigurationManager());

    public static ConfigurationManager Instance => LazyInstance.Value;

    public string? Token { get; set; } = Environment.GetEnvironmentVariable("CHAT_CLI_OPEN_AI_TOKEN");

    public static void SetToken(string token)
    {
        Instance.Token = token;
    }

    public static string? EnsureToken()
    {
        return Instance.Token ?? throw new FailedPreconditionException("Token is not set. Please call the chatcli token xxxxxxx method or set the CHAT_CLI_OPEN_AI_TOKEN environment variable.");
    }
}
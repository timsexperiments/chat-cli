internal class FailedPreconditionException(string? message) : ApplicationException(message)
{
}

internal class WebsocketConnectionException(string? message, Exception? cause) : ApplicationException(message, cause)
{
}
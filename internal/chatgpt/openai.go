package chatgpt

import (
	"fmt"
	"log"
	"strings"

	"github.com/timsexperiments/chat-cli/internal/proto/chat"
)

func SendMessage(message, token string, messages []*chat.Message) (responseMessage, context, completionId string, err error) {
	if message == "" || token == "" {
		return "", "", "", fmt.Errorf("message and token must be provided")
	}

	contextMessage := ChatMessage{
		Role:    "system",
		Content: "Provide a response to the previous messages. In a separate paragraph at the end, summarize the entire chat including your response.",
	}

	requestMessages := append(messagesToChatMessages(messages), contextMessage)
	response, err := MakeChatRequest(requestMessages, 0.3, token)
	if err != nil {
		return "", "", "", fmt.Errorf("unable to make chat request: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", "", "", fmt.Errorf("no choices returned")
	}

	messageParts := strings.Split(response.Choices[0].Message.Content, "\n\n")

	if len(messageParts) < 2 {
		log.Default().Println("Expected at least 2 parts in the response, a response and a context, only found 1 part. Both context and message will be the full message content.")
	}

	return strings.Join(messageParts[0:len(messageParts)-1], "\n\n"), messageParts[len(messageParts)-1], response.ID, nil
}

type messages = []*chat.Message

func messageToChatMessage(message *chat.Message) ChatMessage {
	role := "system"
	if message.Sender == chat.Message_USER {
		role = "user"
	}
	if message.Sender == chat.Message_BOT {
		role = "assistant"
	}
	return ChatMessage{
		Role:    role,
		Content: message.Body,
	}
}

func messagesToChatMessages(m messages) []ChatMessage {
	chatMessages := make([]ChatMessage, len(m))
	for i, message := range m {
		chatMessages[i] = messageToChatMessage(message)
	}
	return chatMessages
}

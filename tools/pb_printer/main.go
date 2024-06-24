package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/timsexperiments/chat-cli/internal/proto/chat"
	"github.com/timsexperiments/chat-cli/internal/proto/errors"
	"google.golang.org/protobuf/proto"
)

var (
	title string
	body  string
)

func main() {
	flag.StringVar(&body, "body", "This is a test message...", "body of the message")
	flag.StringVar(&title, "title", "Test Title", "title of the conversation")
	flag.Parse()
	flagArgs := flag.Args()
	if len(flagArgs) < 1 {
		panic("missing message type")
	}

	messageType := flagArgs[0]
	command := "build"
	if len(flagArgs) > 1 {
		command = flagArgs[0]
		messageType = flagArgs[1]
	}

	switch command {
	case "build":
		runSerialize(messageType)
	case "parse":
		serializedMessage, err := os.ReadFile("bin/input.txt")
		if err != nil {
			panic(fmt.Errorf("unable to read input file: %w", err))
		}
		runParse(string(serializedMessage), messageType)
	default:
		panic(fmt.Errorf("command %s does not exist", command))
	}
}

func runSerialize(messageType string) {
	var serialized []byte
	var err error
	switch messageType {
	case "CreateConversationRequest":
		message := &chat.CreateConversationRequest{Title: title}
		serialized, err = proto.Marshal(message)
		if err != nil {
			panic(fmt.Errorf("unable to serialize message: %w", err))
		}
	case "CreateMessageRequest":
		message := &chat.CreateMessageRequest{Body: body}
		serialized, err = proto.Marshal(message)
		if err != nil {
			panic(fmt.Errorf("unable to serialize message: %w", err))
		}
	case "MessageEvent":
		message := &chat.MessageEvent{Body: body}
		serialized, err = proto.Marshal(message)
		if err != nil {
			panic(fmt.Errorf("unable to serialize message: %w", err))
		}
	case "ChatEvent":
		fallthrough
	case "ErrorEvent":
		fallthrough
	case "ListConversationsResponse":
		fallthrough
	case "Conversation":
		fallthrough
	case "Message":
		fallthrough
	case "Error":
		fallthrough
	default:
		panic(fmt.Errorf("serialization not supported for message type: %s", messageType))
	}

	if err != nil {
		panic(fmt.Errorf("unable to serialize message: %w", err))
	}
	file := "bin/serialization_output.txt"
	os.WriteFile(file, serialized, 0777)
	fmt.Printf("Succesfully wrote output to %s\n", file)
}

func runParse(serializedMessage, messageType string) {
	var message proto.Message
	switch messageType {
	case "Conversation":
		message = &chat.Conversation{}
	case "Message":
		message = &chat.Message{}
	case "CreateConversationRequest":
		message = &chat.CreateConversationRequest{}
	case "CreateMessageRequest":
		message = &chat.CreateMessageRequest{}
	case "ListConversationsResponse":
		message = &chat.ListConversationsResponse{}
	case "ChatEvent":
		message = &chat.ChatEvent{}
	case "MessageEvent":
		message = &chat.MessageEvent{}
	case "ErrorEvent":
		message = &chat.ErrorEvent{}
	case "Error":
		message = &errors.Error{}

	default:
		panic(fmt.Errorf("deserialization not supported for message type: %s", messageType))
	}
	err := proto.Unmarshal([]byte(serializedMessage), message)
	if err != nil {
		panic(fmt.Errorf("unable to parse conversation from input [%s]: %w", serializedMessage, err))
	}
	fmt.Println(message)
}

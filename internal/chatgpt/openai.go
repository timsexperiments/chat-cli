package chatgpt

import "fmt"

func SendMessage(message, token string) (string, error) {
	if message == "" || token == "" {
		return "", fmt.Errorf("message and token must be provided")
	}
	return "This is a fake chatgpt message.", nil
}

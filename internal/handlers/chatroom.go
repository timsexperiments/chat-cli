package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/timsexperiments/chat-cli/internal/chatgpt"
	"github.com/timsexperiments/chat-cli/internal/constants"
	"github.com/timsexperiments/chat-cli/internal/database"
	"github.com/timsexperiments/chat-cli/internal/proto/chat"
	"google.golang.org/protobuf/proto"
)

func handleConversation(c echo.Context) error {
	token, ok := c.Get(constants.OPEN_API_TOKEN_KEY).(string)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "missing open api token")
	}
	idParam := c.Param("id")
	conversationId, err := strconv.Atoi(idParam)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("invalid conversation id [%s]: %w", idParam, err).Error())
	}

	db, ok := c.Get(constants.DB_KEY).(*database.DB)
	if !ok {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "unable to get database")
	}

	conversation, err := db.GetConversation(conversationId)
	if (err != nil) || (conversation == nil) {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusNotFound, fmt.Errorf("unable to get conversation with id [%d]: %w", conversationId, err).Error())
	}

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	context := conversation.Context
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			c.Logger().Error(err)
		}

		eventMsg := &chat.MessageEvent{}
		if err := proto.Unmarshal(msg, eventMsg); err != nil {
			c.Logger().Error(err)
			errorEvent, err := buildErrorResposne(chat.ErrorEvent_INPUT_VALIDATION_ERROR, "unable to parse message")
			if err != nil {
				c.Logger().Error(err)
				ws.CloseHandler()(websocket.CloseInternalServerErr, "Internal Server Error")
				return nil
			}
			if err := ws.WriteMessage(websocket.BinaryMessage, errorEvent); err != nil {
				c.Logger().Error(err)
			}
			continue
		}

		if _, err := db.CreateMessage(eventMsg.Body, chat.Message_USER, conversation.Id); err != nil {
			c.Logger().Error(err)
		}

		var chatEvent *chat.ChatEvent
		chatEvent, context, err = askChatGpt(eventMsg.Body, token, context)
		if err != nil {
			c.Logger().Error(err)
			errorEvent, err := buildErrorResposne(chat.ErrorEvent_SERVER_ERROR, "unable to ask chatgpt")
			if err != nil {
				c.Logger().Error(err)
				ws.CloseHandler()(websocket.CloseInternalServerErr, "Internal Server Error")
				return nil
			}
			if err := ws.WriteMessage(websocket.BinaryMessage, errorEvent); err != nil {
				c.Logger().Error(err)
			}
			continue
		}

		if _, err := db.CreateMessage(chatEvent.GetMessage().Body, chat.Message_BOT, conversation.Id); err != nil {
			c.Logger().Error(err)
		}

		outMsg, err := proto.Marshal(chatEvent)
		if err != nil {
			c.Logger().Error(err)
			errorEvent, err := buildErrorResposne(chat.ErrorEvent_SERVER_ERROR, "unable to serialize chat event")
			if err != nil {
				c.Logger().Error(err)
				ws.CloseHandler()(websocket.CloseInternalServerErr, "Internal Server Error")
				return nil
			}
			if err := ws.WriteMessage(websocket.BinaryMessage, errorEvent); err != nil {
				c.Logger().Error(err)
			}
			continue
		}

		if err := ws.WriteMessage(websocket.BinaryMessage, outMsg); err != nil {
			c.Logger().Error(err)
		}
	}
}

func buildErrorResposne(errType chat.ErrorEvent_Type, message string) ([]byte, error) {
	event := &chat.ChatEvent{
		Type: chat.ChatEvent_ERROR,
		Event: &chat.ChatEvent_Error{
			Error: &chat.ErrorEvent{Type: errType, Message: message},
		},
	}
	return proto.Marshal(event)
}

func askChatGpt(message string, token, context string) (*chat.ChatEvent, string, error) {
	response, err := chatgpt.SendMessage(message, token)
	if err != nil {
		return nil, "", fmt.Errorf("unable to ask chatgpt: %w", err)
	}
	event := &chat.ChatEvent{
		Type: chat.ChatEvent_MESSAGE,
		Event: &chat.ChatEvent_Message{
			Message: &chat.MessageEvent{Body: response},
		},
	}
	return event, context, nil
}

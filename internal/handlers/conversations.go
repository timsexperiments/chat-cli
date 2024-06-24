package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/timsexperiments/chat-cli/internal/constants"
	"github.com/timsexperiments/chat-cli/internal/database"
	"github.com/timsexperiments/chat-cli/internal/middleware"
	"github.com/timsexperiments/chat-cli/internal/proto/chat"
	"github.com/timsexperiments/chat-cli/internal/response"
	"google.golang.org/protobuf/proto"
)

var (
	upgrader = websocket.Upgrader{}
)

func RegisterConversationsHandlers(e *echo.Echo) {
	conversations := e.Group("/conversations")
	conversations.Use(middleware.ProtobufBodyChecker)
	conversations.Use(middleware.ProtobufHeader)
	conversations.POST("", createConversationHandler)
	conversations.GET("", listConversationsHandler)
	conversationGroup := conversations.Group("/:id")
	conversationGroup.Use(middleware.AuthChecker)
	conversationGroup.GET("", conversationHandler)
	messages := conversationGroup.Group("/messages")
	messages.POST("", createMessage)
}

func createConversationHandler(c echo.Context) error {
	db := c.Get(constants.DB_KEY).(*database.DB)
	body, ok := c.Get(constants.BODY_KEY).([]byte)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "missing body")
	}
	request := &chat.CreateConversationRequest{}
	if err := proto.Unmarshal([]byte(body), request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("unable to parse request: %w", err).Error())
	}
	conversation, err := db.CreateConversation(request.Title)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("unable to create conversation: %w", err).Error())
	}
	return response.Protobuf(c, http.StatusCreated, conversation)
}

func listConversationsHandler(c echo.Context) error {
	db := c.Get(constants.DB_KEY).(*database.DB)
	conversationsList, err := db.ListConversations()
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("unable to list conversations: %w", err).Error())
	}
	conversations := &chat.ListConversationsResponse{Conversations: conversationsList}
	return response.Protobuf(c, http.StatusOK, conversations)
}

func conversationHandler(c echo.Context) error {
	if !websocket.IsWebSocketUpgrade(c.Request()) {
		return getConversationHandler(c)
	}

	return handleConversation(c)
}

func getConversationHandler(c echo.Context) error {
	db := c.Get(constants.DB_KEY).(*database.DB)
	strId := c.Param("id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("invalid conversation id [%s]: %w", strId, err).Error())
	}
	conversation, err := db.GetConversation(id)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("unable to get conversation: %w", err).Error())
	}
	return response.Protobuf(c, http.StatusOK, conversation)
}

func createMessage(c echo.Context) error {
	db := c.Get(constants.DB_KEY).(*database.DB)
	strId := c.Param("id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("invalid conversation id [%s]: %w", strId, err).Error())
	}
	body := c.Get(constants.BODY_KEY).([]byte)
	request := chat.CreateMessageRequest{}
	proto.Unmarshal(body, &request)
	message, err := db.CreateMessage(request.Body, chat.Message_USER, int64(id))
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("unable to create conversation: %w", err).Error())
	}
	return response.Protobuf(c, http.StatusCreated, message)
}

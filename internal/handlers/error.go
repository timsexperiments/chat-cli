package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/timsexperiments/chat-cli/internal/proto/errors"
	"github.com/timsexperiments/chat-cli/internal/response"
)

func ErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	message := "an unknown message occurred"
	if he, ok := err.(*echo.HTTPError); ok {
		if messageStr, ok := he.Message.(string); ok {
			message = messageStr
		}
		code = he.Code
	}
	c.Logger().Error(err)
	responseErr := &errors.Error{Message: message}
	response.Protobuf(c, code, responseErr)
}

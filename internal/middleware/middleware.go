package middleware

import (
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/timsexperiments/chat-cli/internal/config"
	"github.com/timsexperiments/chat-cli/internal/database"
)

func ProtobufHeader(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "application/protobuf")
		return next(c)
	}
}

func AuthChecker(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		var token string
		if len(authHeader) > len("Bearer ") {
			token = authHeader[len("Bearer "):]
		}
		if len(token) == 0 {
			token = c.QueryParams().Get("api_secret")
		}
		if len(token) == 0 {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing open api token")
		}
		c.Set(config.OPEN_AI_TOKEN_KEY, token)
		return next(c)
	}
}

func ProtobufBodyChecker(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("unable to read body: %w", err).Error())
		}
		if len(body) == 0 {
			return next(c)
		}
		contentType := c.Request().Header.Get("Content-Type")
		if contentType != "application/protobuf" {
			return echo.NewHTTPError(http.StatusUnsupportedMediaType, fmt.Errorf("invalid content type. Expected application/protobuf: %s", contentType).Error())
		}
		c.Set(config.BODY_KEY, body)
		return next(c)
	}
}

func ContextDB(db *database.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(config.DB_KEY, db)
			return next(c)
		}
	}
}

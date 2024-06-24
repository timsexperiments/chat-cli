package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"google.golang.org/protobuf/proto"
)

func Protobuf(c echo.Context, code int, message proto.Message) error {
	serialized, err := proto.Marshal(message)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.Blob(code, "application/protobuf", serialized)
}

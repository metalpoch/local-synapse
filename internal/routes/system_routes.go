package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/metalpoch/local-synapse/internal/handler"
)

func NewSystemRoutes(server *echo.Echo) {
	hdlr := handler.NewSystemHandler()

	route := server.Group("/api/v1/system")
	route.GET("/stats", hdlr.Stats)
}

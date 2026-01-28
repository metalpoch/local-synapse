package router

import (
	"github.com/labstack/echo/v4"
	"github.com/metalpoch/local-synapse/internal/handler"
)

func SetupSystemRouter(e *echo.Echo) {
	h := handler.NewSystemHandler()

	router := e.Group("/api/v1/system")
	router.GET("/stats", h.Stats)
}

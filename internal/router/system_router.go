package router

import (
	"github.com/labstack/echo/v4"
	"github.com/metalpoch/local-synapse/internal/handler"
	"github.com/metalpoch/local-synapse/internal/middleware"
	"github.com/metalpoch/local-synapse/internal/pkg/authentication"
)

func SetupSystemRouter(e *echo.Echo, am authentication.AuthManager) {
	h := handler.NewSystemHandler()

	router := e.Group("/api/v1/system", middleware.AuthMiddleware(am))
	router.GET("/stats", h.Stats)
}

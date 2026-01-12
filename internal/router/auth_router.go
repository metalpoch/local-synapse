package router

import (
	"github.com/labstack/echo/v4"
	"github.com/metalpoch/local-synapse/internal/handler"
	"github.com/metalpoch/local-synapse/internal/pkg/authentication"
	"github.com/metalpoch/local-synapse/internal/repository"
	"github.com/metalpoch/local-synapse/internal/usecase/auth"
)

func SetupAuthRouter(e *echo.Echo, am authentication.AuthManager, ur repository.UserRepository) {
	h := handler.NewAuthHandler(
		auth.NewUserLogin(am, ur),
		auth.NewUserRegister(am, ur),
	)

	router := e.Group("/api/v1/auth")
	router.POST("/login", h.Login)
	router.POST("/register", h.Register)
}

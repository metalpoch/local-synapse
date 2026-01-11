package router

import (
	"github.com/labstack/echo/v4"
	"github.com/metalpoch/local-synapse/internal/handler"
	"github.com/metalpoch/local-synapse/internal/repository"
	"github.com/metalpoch/local-synapse/internal/usecase/auth"
)

func SetupAuthRouter(e *echo.Echo, ur repository.UserRepository) {
	h := handler.NewAuthHandler(
		auth.NewUserLogin(ur),
		auth.NewUserRegister(ur),
	)

	router := e.Group("/api/v1/auth")
	router.POST("/login", h.Login)
	router.POST("/register", h.Register)
}

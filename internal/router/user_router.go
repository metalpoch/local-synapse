package router

import (
	"github.com/labstack/echo/v4"
	"github.com/metalpoch/local-synapse/internal/handler"
	"github.com/metalpoch/local-synapse/internal/middleware"
	"github.com/metalpoch/local-synapse/internal/pkg/authentication"
	"github.com/metalpoch/local-synapse/internal/repository"
	"github.com/metalpoch/local-synapse/internal/usecase/user"
)

func SetupAuthRouter(e *echo.Echo, am authentication.AuthManager, ur repository.UserRepository) {
	h := handler.NewAuthHandler(
		user.NewUserLogin(am, ur),
		user.NewUserRegister(am, ur),
		user.NewGetUser(ur),
		user.NewUserLogout(am),
		user.NewRefreshToken(am),
		user.NewUpdateUser(ur),
	)

	authRouter := e.Group("/api/v1/auth")
	authRouter.POST("/login", h.Login)
	authRouter.POST("/register", h.Register)
	authRouter.POST("/refresh", h.Refresh)
	authRouter.GET("/me", h.Me, middleware.AuthMiddleware(am))
	authRouter.POST("/logout", h.Logout, middleware.AuthMiddleware(am))

	userRouter := e.Group("/api/v1/user", middleware.AuthMiddleware(am))
	userRouter.PUT("/profile", h.UpdateProfile, middleware.AuthMiddleware(am))
}

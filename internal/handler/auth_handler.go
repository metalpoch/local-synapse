package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/metalpoch/local-synapse/internal/dto"
	"github.com/metalpoch/local-synapse/internal/middleware"
	"github.com/metalpoch/local-synapse/internal/usecase/auth"
	"github.com/metalpoch/local-synapse/internal/usecase/user"
)

type authHandler struct {
	userLogin    *auth.UserLogin
	userRegister *auth.UserRegister
	getUser      *user.GetUser
}

func NewAuthHandler(ul *auth.UserLogin, ur *auth.UserRegister, gu *user.GetUser) *authHandler {
	return &authHandler{ul, ur, gu}
}

func (h *authHandler) Register(c echo.Context) error {
	var input dto.UserRegisterRequest
	err := c.Bind(&input)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	user, err := h.userRegister.Execute(input)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, user)
}

func (h *authHandler) Login(c echo.Context) error {
	var input dto.UserLoginRequest
	err := c.Bind(&input)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	user, err := h.userLogin.Execute(input)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, user)
}

func (h *authHandler) Me(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	user, err := h.getUser.Execute(userID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, user)
}

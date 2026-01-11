package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/metalpoch/local-synapse/internal/dto"
	"github.com/metalpoch/local-synapse/internal/usecase/auth"
)

type authHandler struct {
	userLogin    *auth.UserLogin
	userRegister *auth.UserRegister
}

func NewAuthHandler(ul *auth.UserLogin, ur *auth.UserRegister) *authHandler {
	return &authHandler{ul, ur}
}

func (h *authHandler) Register(c echo.Context) error {
	var input dto.UserRegister
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
	var input dto.UserLogin
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

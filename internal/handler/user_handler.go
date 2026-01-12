package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/metalpoch/local-synapse/internal/dto"
	"github.com/metalpoch/local-synapse/internal/middleware"
	"github.com/metalpoch/local-synapse/internal/usecase/user"
)

type authHandler struct {
	userLogin    *user.UserLogin
	userRegister *user.UserRegister
	getUser      *user.GetUser
	userLogout   *user.UserLogout
	refreshToken *user.RefreshToken
	updateUser   *user.UpdateUser
}

func NewAuthHandler(
	ul *user.UserLogin,
	ur *user.UserRegister,
	gu *user.GetUser,
	ulo *user.UserLogout,
	rt *user.RefreshToken,
	uu *user.UpdateUser,
) *authHandler {
	return &authHandler{ul, ur, gu, ulo, rt, uu}
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

func (h *authHandler) Logout(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	accessToken := ""
	if len(authHeader) > 7 {
		accessToken = authHeader[7:]
	}

	var input dto.RefreshTokenRequest
	_ = c.Bind(&input) // Refresh token is optional for logout but recommended

	err := h.userLogout.Execute(accessToken, input.RefreshToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "logged out successfully"})
}

func (h *authHandler) Refresh(c echo.Context) error {
	var input dto.RefreshTokenRequest
	err := c.Bind(&input)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
	}

	tokens, err := h.refreshToken.Execute(input)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, tokens)
}

func (h *authHandler) UpdateProfile(c echo.Context) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	var input dto.UpdateProfileRequest
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid form data"})
	}

	file, _ := c.FormFile("image")

	err := h.updateUser.Execute(c.Request().Context(), userID, input.Name, file)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "profile updated successfully"})
}

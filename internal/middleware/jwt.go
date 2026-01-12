package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/metalpoch/local-synapse/internal/pkg/authentication"
)

const ContextUserIDKey string = "user_id"

var ErrMissingAuth string = "missing authentication token"

func AuthMiddleware(authManager authentication.AuthManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")

			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": ErrMissingAuth})
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": ErrMissingAuth})
			}

			accessToken := strings.TrimPrefix(authHeader, "Bearer ")

			userID, err := authAssistant(authManager, c, accessToken)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
			}

			c.Set(ContextUserIDKey, userID)

			return next(c)
		}
	}
}

func authAssistant(authManager authentication.AuthManager, c echo.Context, token string) (string, error) {
	userID, err := authManager.ValidateAccessToken(c.Request().Context(), token)
	if err != nil {
		return "", err
	}

	if userID == "" {
		return "", errors.New("user ID missing in token")
	}

	return userID, nil
}

func GetUserID(c echo.Context) (string, bool) {
	userID := c.Get(ContextUserIDKey)
	if userID == nil {
		return "", false
	}
	userIDStr, ok := userID.(string)
	return userIDStr, ok
}

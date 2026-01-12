package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/metalpoch/local-synapse/internal/pkg/authentication"
)

type UserLogout struct {
	authManager authentication.AuthManager
}

func NewUserLogout(am authentication.AuthManager) *UserLogout {
	return &UserLogout{am}
}

func (uc *UserLogout) Execute(accessToken, refreshToken string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if accessToken == "" {
		return fmt.Errorf("access token is required for logout")
	}

	err := uc.authManager.Logout(ctx, accessToken, refreshToken)
	if err != nil {
		return fmt.Errorf("failed to logout: %v", err)
	}

	return nil
}

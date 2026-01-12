package user

import (
	"context"
	"fmt"
	"time"

	"github.com/metalpoch/local-synapse/internal/dto"
	"github.com/metalpoch/local-synapse/internal/pkg/authentication"
)

type RefreshToken struct {
	authManager authentication.AuthManager
}

func NewRefreshToken(am authentication.AuthManager) *RefreshToken {
	return &RefreshToken{am}
}

func (uc *RefreshToken) Execute(input dto.RefreshTokenRequest) (*dto.Tokens, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if input.RefreshToken == "" {
		return nil, fmt.Errorf("refresh token is required")
	}

	accessToken, newRefreshToken, err := uc.authManager.RefreshToken(ctx, input.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh tokens: %v", err)
	}

	return &dto.Tokens{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

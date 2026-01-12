package user

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/metalpoch/local-synapse/internal/dto"
	"github.com/metalpoch/local-synapse/internal/pkg/authentication"
	"github.com/metalpoch/local-synapse/internal/pkg/validator"
	"github.com/metalpoch/local-synapse/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserLogin struct {
	authManager authentication.AuthManager
	userRepo    repository.UserRepository
}

func NewUserLogin(am authentication.AuthManager, ur repository.UserRepository) *UserLogin {
	return &UserLogin{am, ur}
}

func (uc *UserLogin) Execute(input dto.UserLoginRequest) (*dto.UserResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if input.Password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	if err := validator.Email(input.Email); err != nil {
		return nil, err
	}

	user, err := uc.userRepo.Login(ctx, input.Email)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invalid username or password")
	} else if err != nil {
		return nil, fmt.Errorf("searching for user: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	accessToken, refreshToken, err := uc.authManager.GenerateTokens(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %v", err)
	}

	userLogin := dto.UserResponse{
		ID:           user.ID,
		Name:         user.Name,
		Email:        user.Email,
		AuthProvider: user.AuthProvider,
		CodeProvider: user.CodeProvider,
		ImageURL:     user.ImageURL,
		Tokens: dto.Tokens{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}

	return &userLogin, nil
}

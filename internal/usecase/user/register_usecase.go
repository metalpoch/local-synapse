package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/metalpoch/local-synapse/internal/dto"
	"github.com/metalpoch/local-synapse/internal/entity"
	"github.com/metalpoch/local-synapse/internal/pkg/authentication"
	"github.com/metalpoch/local-synapse/internal/pkg/validator"
	"github.com/metalpoch/local-synapse/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserRegister struct {
	authManager authentication.AuthManager
	userRepo repository.UserRepository
}

func NewUserRegister(am authentication.AuthManager, ur repository.UserRepository) *UserRegister {
	return &UserRegister{am, ur}
}

func (uc *UserRegister) Execute(user dto.UserRegisterRequest) (*dto.UserResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := validator.Email(user.Email); err != nil {
		return nil, err
	}

	if err := validator.Password(user.Password); err != nil {
		return nil, fmt.Errorf("error validating password: %v", err)
	}

	if user.Password != user.ConfirmPassword {
		return nil, validator.ErrPasswordsMismatch
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error processing the password: %v", err)
	}

	newUser := &dto.UserResponse{
		Name:         strings.TrimSpace(user.FirstName + " " + user.Lastname),
		Email:        user.Email,
		AuthProvider: "email",
	}

	id, err := uc.userRepo.Register(ctx, entity.User{
		Name:         newUser.Name,
		Email:        newUser.Email,
		AuthProvider: newUser.AuthProvider,
		Password:     string(hashedPassword),
	})
	if err != nil {
		return nil, fmt.Errorf("error trying register a user: %v", err)
	}

	accessToken, refreshToken, err := uc.authManager.GenerateTokens(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %v", err)
	}

	newUser.ID = id
	newUser.AccessToken = accessToken
	newUser.RefreshToken = refreshToken

	return newUser, nil
}

package user

import (
	"context"
	"fmt"
	"time"

	"github.com/metalpoch/local-synapse/internal/dto"
	"github.com/metalpoch/local-synapse/internal/repository"
)

type GetUser struct {
	userRepo repository.UserRepository
}

func NewGetUser(ur repository.UserRepository) *GetUser {
	return &GetUser{ur}
}

func (uc *GetUser) Execute(id string) (*dto.UserResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("searching for user: %v", err)
	}
	return &user, nil
}

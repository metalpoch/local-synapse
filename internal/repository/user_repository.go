package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/metalpoch/local-synapse/internal/dto"
	"github.com/metalpoch/local-synapse/internal/entity"
)

type UserRepository interface {
	Register(ctx context.Context, user entity.User) (string, error)
	Login(ctx context.Context, email string) (entity.User, error)
	GetByID(ctx context.Context, id string) (dto.UserResponse, error)
}

type userRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *userRepo {
	return &userRepo{db}
}

// Register creates a new user record in the database.
func (repo *userRepo) Register(ctx context.Context, user entity.User) (string, error) {
	id := uuid.New().String()
	query := "INSERT INTO users(id, name, email, password, auth_provider, code_provider) VALUES(?, ?, ?, ?, ?, ?)"

	_, err := repo.db.ExecContext(
		ctx,
		query,
		id,
		user.Name,
		user.Email,
		user.Password,
		user.AuthProvider,
		user.CodeProvider,
	)

	if err != nil {
		return "", err
	}

	return id, nil
}

// GetByID retrieves user details for a given ID.
func (repo *userRepo) GetByID(ctx context.Context, id string) (dto.UserResponse, error) {
	var user dto.UserResponse

	query := "SELECT id, name, email, image_url FROM users WHERE id = ?"

	err := repo.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Name, &user.Email, &user.ImageURL)

	return user, err
}

// Login fetches user authentication data by email.
func (repo *userRepo) Login(ctx context.Context, email string) (entity.User, error) {
	var user entity.User

	query := "SELECT id, name, email, image_url, password, auth_provider FROM users WHERE email = ?"

	err := repo.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Name, &user.Email, &user.ImageURL, &user.Password, &user.AuthProvider)

	return user, err
}

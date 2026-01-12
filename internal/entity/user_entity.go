package entity

import "time"

type User struct {
	ID           string    `db:"id"`
	Name         string    `db:"name"`
	Email        string    `db:"email"`
	Password     string    `db:"password"`
	AuthProvider string    `db:"auth_provider"`
	CodeProvider *int32    `db:"code_provider"`
	ImageURL     *string   `db:"image_url"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

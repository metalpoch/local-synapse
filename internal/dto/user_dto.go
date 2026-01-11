package dto

type User struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Email        string  `json:"email"`
	AuthProvider string  `json:"auth_provider"`
	CodeProvider *int32  `json:"code_provider"`
	ImageURL     *string `json:"image_url"`
}

type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserRegister struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	FirstName       string `json:"first_name"`
	Lastname        string `json:"last_name"`
}

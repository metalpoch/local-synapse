package authentication

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	Sub string `json:"sub"`
	jwt.RegisteredClaims
}

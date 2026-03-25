package authentication

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type Service struct {
	db        *gorm.DB
	jwtSecret []byte
	jwtExpiry time.Duration
}

type CustomClaims struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	ID       uint64 `json:"id"`
	jwt.RegisteredClaims
}

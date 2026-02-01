package auth

import (
	"gorm.io/gorm"
	"time"
	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	db *gorm.DB
	jwtSecret []byte
	jwtExpiry time.Duration 
}

type CustomClaims struct {
	Username string `json:"username"`
    Email string `json:"email"`
    Role  string `json:"role"`
	ID uint64 `json:"id"`
    jwt.RegisteredClaims
}
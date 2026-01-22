package auth

import (
	"errors"
	"time"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirkartik/cloud_drive_2.0/internal/config"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"strconv"
)

type Service struct {
	db *gorm.DB
	jwtSecret []byte
	jwtExpiry time.Duration 
}

type CustomClaims struct {
	Usernam string `json:"username"`
    Email string `json:"email"`
    Role  string `json:"role"`
    jwt.RegisteredClaims
}

func NewService(DB *gorm.DB, cfg config.Config) *Service {
	DB.AutoMigrate(&User{})
	return &Service{
		db : DB,
		jwtSecret: []byte(cfg.JWT.Secret),
		jwtExpiry: time.Hour * time.Duration(cfg.JWT.ExpiryHour),
	}
}

func (svc *Service) RegisterService(email string, username string, password string) error {
	if email == "" || username == "" || password == "" {
		return errors.New("email, username, and password cannot be empty")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := User{
		Email:    email,
		Username: username,
		Password: string(hashedPassword),
	}

	result := svc.db.Create(&user)
	return result.Error
}

func (svc *Service) LoginService(email string, password string) (*User, error) {
	if email == "" || password == "" {
		return nil, errors.New("email and password cannot be empty")
	}

	var user User
	err := svc.db.Where("email = ?", email).First(&user).Error

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("NotRegistered")
	}
	passwordError := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if passwordError != nil {
		return nil, errors.New("InvalidPassword")
	}
	return &user, nil
}

func (svc *Service) GenerateToken(user *User) (string, error) {
    claims := jwt.MapClaims{
        "sub": strconv.Itoa(int(user.ID)),  
		"username" : user.Username,
        "email": user.Email,
        "role":  user.Role,
        "iat":   time.Now().Unix(),
        "exp":   time.Now().Add(svc.jwtExpiry).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    signedToken, err := token.SignedString(svc.jwtSecret)
    if err != nil {
        return "", err
    }

    return signedToken, nil
}

func (svc *Service) DecodeToken(token *string, secret []byte) (*CustomClaims, error) {
	tokenActual, err := jwt.ParseWithClaims(
        *token,
        &CustomClaims{},
        func(t *jwt.Token) (any, error) {
            if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, errors.New("unexpected signing method")
            }
            return secret, nil
        },
    )

    if err != nil {
        return nil, err
    }

    if !tokenActual.Valid {
        return nil, errors.New("invalid token")
    }

    claims, ok := tokenActual.Claims.(*CustomClaims)
    if !ok {
        return nil, errors.New("invalid token claims")
    }
    return claims, nil
}

package auth

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(DB *gorm.DB) *Service {
	DB.AutoMigrate(&User{})
	return &Service{
		db : DB,
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

func (svc *Service) LoginService(email string, password string) (error, *User) {
	if email == "" || password == "" {
		return errors.New("email and password cannot be empty"), nil
	}

	var user User
	err := svc.db.Where("email = ?", email).First(&user).Error

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("NotRegistered"), nil
	}
	passwordError := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if passwordError != nil {
		return errors.New("InvalidPassword"), nil
	}
	return nil, &user
}
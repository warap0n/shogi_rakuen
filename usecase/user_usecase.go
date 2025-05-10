package usecase

import (
	"errors"
	"fmt"
	"os"
	"shogi-rakuen/model"
	"shogi-rakuen/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type IUserUsecase interface {
	SignUp(user *model.User) (model.UserResponse, error)
	Login(user *model.User) (string, error)
	GetUserByEmail(email string) (model.UserResponse, error)
}

type UserUsecase struct {
	ur repository.IUserRepository
}

func NewUserUsecase(ur repository.IUserRepository) IUserUsecase {
	return &UserUsecase{ur: ur}
}

func (uu *UserUsecase) SignUp(user *model.User) (model.UserResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.UserResponse{}, err
	}

	newUser := &model.User{
		Email:    user.Email,
		Username: user.Username,
		Password: string(hash),
	}
	createdUser, err := uu.ur.Create(newUser)
	if err != nil {
		return model.UserResponse{}, err
	}

	response := model.UserResponse{
		ID:       createdUser.ID,
		Email:    createdUser.Email,
		Username: createdUser.Username,
		Rank:     createdUser.Rank,
	}

	return response, nil
}

func (uu *UserUsecase) Login(user *model.User) (string, error) {
	// DBからユーザー取得（Emailで）
	storedUser, err := uu.ur.GetUserByEmail(user.Email)
	if err != nil {
		return "", ErrEmailNotFound
	}

	// パスワード検証
	err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password))
	if err != nil {
		return "", ErrInvalidPassword
	}

	// JWT生成
	claims := jwt.RegisteredClaims{
		Subject:   fmt.Sprint(storedUser.ID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", ErrJWTSecretUnset
	}

	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return t, nil
}

func (uu *UserUsecase) GetUserByEmail(email string) (model.UserResponse, error) {
	user, err := uu.ur.GetUserByEmail(email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return model.UserResponse{}, ErrEmailNotFound
	}
	if err != nil {
		return model.UserResponse{}, err
	}
	response := model.UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		Rank:     user.Rank,
	}
	return response, nil
}

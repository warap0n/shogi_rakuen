package usecase

import (
	"errors"
	"fmt"
	"os"
	"shogi-rakuen/model"
	"shogi-rakuen/repository"
	"shogi-rakuen/usecase/input"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type IUserUsecase interface {
	SignUp(input input.SignupInput) (model.UserResponse, error)
	Login(input input.LoginInput) (string, error)
	GetUserById(id uint) (model.UserResponse, error)
}

type UserUsecase struct {
	ur repository.IUserRepository
}

func NewUserUsecase(ur repository.IUserRepository) IUserUsecase {
	return &UserUsecase{ur: ur}
}

func (uu *UserUsecase) SignUp(input input.SignupInput) (model.UserResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.UserResponse{}, err
	}

	newUser := &model.User{
		Email:    input.Email,
		Username: input.Username,
		Password: string(hash),
	}

	createdUser, err := uu.ur.Create(newUser)
	if err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyExists) {
			return model.UserResponse{}, ErrEmailAlreadyExists
		}
		return model.UserResponse{}, err
	}

	return model.UserResponse{
		ID:       createdUser.ID,
		Email:    createdUser.Email,
		Username: createdUser.Username,
		Rank:     createdUser.Rank,
	}, nil
}

func (uu *UserUsecase) Login(input input.LoginInput) (string, error) {
	storedUser, err := uu.ur.GetUserByEmail(input.Email)
	if err != nil {
		return "", ErrEmailNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(input.Password)); err != nil {
		return "", ErrInvalidPassword
	}

	claims := jwt.RegisteredClaims{
		Subject:   fmt.Sprint(storedUser.ID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", ErrJWTSecretUnset
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (uu *UserUsecase) GetUserById(id uint) (model.UserResponse, error) {
	user, err := uu.ur.GetUserById(id)
	if errors.Is(err, repository.ErrUserNotFound) {
		return model.UserResponse{}, ErrUserNotFound
	}
	if err != nil {
		return model.UserResponse{}, err
	}

	return model.UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		Rank:     user.Rank,
	}, nil
}

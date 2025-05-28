package usecase

import (
	"errors"
	"fmt"
	"os"
	"shogi-rakuen/model"
	"shogi-rakuen/repository"
	"shogi-rakuen/usecase/input"
	"shogi-rakuen/usecase/output"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type IUserUsecase interface {
	SignUp(input input.SignupInput) (output.SignupOutput, error)
	Login(input input.LoginInput) (output.LoginOutput, error)
	GetUserById(id uint) (output.GetUserByIdOutput, error)
}

type UserUsecase struct {
	ur repository.IUserRepository
}

func NewUserUsecase(ur repository.IUserRepository) IUserUsecase {
	return &UserUsecase{ur: ur}
}

func (uu *UserUsecase) SignUp(input input.SignupInput) (output.SignupOutput, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return output.SignupOutput{}, err
	}

	newUser := &model.User{
		Email:    input.Email,
		Username: input.Username,
		Password: string(hash),
	}

	createdUser, err := uu.ur.Create(newUser)
	if err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyExists) {
			return output.SignupOutput{}, ErrEmailAlreadyExists
		}
		return output.SignupOutput{}, err
	}

	return output.SignupOutput{
		ID:        createdUser.ID,
		Email:     createdUser.Email,
		Username:  createdUser.Username,
		RankType:  createdUser.RankType,
		RankLevel: createdUser.RankLevel,
	}, nil
}

func (uu *UserUsecase) Login(input input.LoginInput) (output.LoginOutput, error) {
	storedUser, err := uu.ur.GetUserByEmail(input.Email)
	if err != nil {
		return output.LoginOutput{}, ErrEmailNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(input.Password)); err != nil {
		return output.LoginOutput{}, ErrInvalidPassword
	}

	claims := jwt.RegisteredClaims{
		Subject:   fmt.Sprint(storedUser.ID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return output.LoginOutput{}, ErrJWTSecretUnset
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return output.LoginOutput{}, err
	}

	return output.LoginOutput{Token: signedToken}, nil
}

func (uu *UserUsecase) GetUserById(id uint) (output.GetUserByIdOutput, error) {
	user, err := uu.ur.GetUserById(id)
	if errors.Is(err, repository.ErrUserNotFound) {
		return output.GetUserByIdOutput{}, ErrUserNotFound
	}
	if err != nil {
		return output.GetUserByIdOutput{}, err
	}

	return output.GetUserByIdOutput{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		RankType:  user.RankType,
		RankLevel: user.RankLevel,
	}, nil
}

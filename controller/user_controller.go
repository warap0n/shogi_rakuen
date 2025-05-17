package controller

import (
	"errors"
	"net/http"
	"os"
	"shogi-rakuen/controller/dto"
	"shogi-rakuen/usecase"
	"shogi-rakuen/usecase/input"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type IUserController interface {
	SignUp(c echo.Context) error
	Login(c echo.Context) error
	Me(c echo.Context) error
}
type UserController struct {
	uu usecase.IUserUsecase
}

func NewUserController(uu usecase.IUserUsecase) IUserController {
	return &UserController{uu}
}

// POST /signup
func (uc *UserController) SignUp(c echo.Context) error {
	var user dto.SignupRequest
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid input"})
	}

	if err := c.Validate(&user); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "validation failed"})
	}
	input := input.SignupInput{
		Email:    user.Email,
		Username: user.Username,
		Password: user.Password,
	}
	created, err := uc.uu.SignUp(input)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrEmailAlreadyExists):
			return c.JSON(http.StatusConflict, echo.Map{"error": "email already registered"})

		default:
			// 本番では詳細を返さずログに残す
			// log.Error(err)
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal server error"})
		}
	}

	return c.JSON(http.StatusCreated, created)
}

// POST /login
func (uc *UserController) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid input"})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "validation failed"})
	}

	input := input.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}
	token, err := uc.uu.Login(input)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrEmailNotFound):
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "email not found"})
		case errors.Is(err, usecase.ErrInvalidPassword):
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid password"})
		case errors.Is(err, usecase.ErrJWTSecretUnset):
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "jwt secret not set"})
		default:
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "internal server error"})
		}
	}

	// Cookie にセット
	cookie := new(http.Cookie)
	cookie.Name = "access_token"
	cookie.Value = token
	cookie.Path = "/"
	cookie.Domain = os.Getenv("API_DOMAIN")
	cookie.Expires = time.Now().Add(24 * time.Hour)
	cookie.HttpOnly = true // JS から参照不可
	// cookie.Secure = true   // Todo:本番用
	cookie.Secure = false // 開発用

	cookie.SameSite = http.SameSiteLaxMode // 必要に応じて Strict も可
	c.SetCookie(cookie)

	// JSON でも返したければ併記
	return c.JSON(http.StatusOK, echo.Map{})

}

// GET /me
func (uc *UserController) Me(c echo.Context) error {
	// 1. JWT ミドルウェアでセットされたトークンを取得
	token := c.Get("user").(*jwt.Token)

	mc, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token claims"})
	}
	sub, ok := mc["sub"].(string)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid subject claim"})
	}
	userId, err := strconv.Atoi(sub)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid user id in token"})
	}

	resp, err := uc.uu.GetUserById(uint(userId))
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "user not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}

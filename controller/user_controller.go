package controller

import (
	"errors"
	"net/http"
	"shogi-rakuen/model"
	"shogi-rakuen/usecase"

	"github.com/labstack/echo/v4"
)

type UserController struct {
	uu usecase.IUserUsecase
}

func NewUserController(uu usecase.IUserUsecase) *UserController {
	return &UserController{uu}
}

// POST /signup
func (uc *UserController) SignUp(c echo.Context) error {
	var user model.User
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid input"})
	}

	if err := c.Validate(&user); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "validation failed"})
	}

	created, err := uc.uu.SignUp(&user)
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

// GET /users/:email
func (uc *UserController) GetUserByEmail(c echo.Context) error {
	email := c.Param("email")

	resp, err := uc.uu.GetUserByEmail(email)
	if err != nil {
		if err == usecase.ErrEmailNotFound {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "user not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}

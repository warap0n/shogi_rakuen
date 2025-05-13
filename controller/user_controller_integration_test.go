package controller_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"shogi-rakuen/controller"
	"shogi-rakuen/controller/dto"
	"shogi-rakuen/model"
	"shogi-rakuen/usecase/input"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// --- テスト内だけで使うバリデータ ---
type customValidator struct {
	validator *validator.Validate
}

func (cv *customValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// --- ダミー usecase ---
type DummyUserUsecase struct{}

func (d *DummyUserUsecase) SignUp(in input.SignupInput) (model.UserResponse, error) {
	return model.UserResponse{
		ID:       1,
		Email:    in.Email,
		Username: in.Username,
		Rank:     "未設定",
	}, nil
}

func (d *DummyUserUsecase) Login(input input.LoginInput) (string, error) {
	return "", nil
}

func (d *DummyUserUsecase) GetUserByEmail(email string) (model.UserResponse, error) {
	return model.UserResponse{}, nil
}

// --- テスト本体 ---
func TestSignUp_InvalidInput(t *testing.T) {
	e := echo.New()
	e.Validator = &customValidator{validator: validator.New()}

	uc := controller.NewUserController(&DummyUserUsecase{})
	e.POST("/signup", uc.SignUp)

	// 無効なリクエスト：Email が空
	body := dto.SignupRequest{
		Email:    "",
		Username: "testuser",
		Password: "secret123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 実行
	if assert.NoError(t, uc.SignUp(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "validation failed")
	}
}

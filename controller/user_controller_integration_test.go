package controller_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"shogi-rakuen/controller"
	"shogi-rakuen/model"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// ダミーusecase（正常な登録を想定）
type DummyUserUsecase struct{}

type CustomValidator struct {
	Validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.Validator.Struct(i)
}

func (d *DummyUserUsecase) SignUp(user *model.User) (model.UserResponse, error) {
	return model.UserResponse{
		ID:       1,
		Email:    user.Email,
		Username: user.Username,
		Rank:     "未設定",
	}, nil
}
func (d *DummyUserUsecase) Login(user *model.User) (string, error) { return "", nil }
func (d *DummyUserUsecase) GetUserByEmail(email string) (model.UserResponse, error) {
	return model.UserResponse{}, nil
}

func TestSignUp_InvalidInput(t *testing.T) {
	e := echo.New()

	// validator登録
	e.Validator = &CustomValidator{Validator: validator.New()}

	// コントローラ設定
	uc := controller.NewUserController(&DummyUserUsecase{})
	e.POST("/signup", uc.SignUp)

	// 無効な入力（空のemail）
	body := map[string]interface{}{
		"email":    "",
		"username": "testuser",
		"password": "secret123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	// 実行
	if assert.NoError(t, uc.SignUp(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

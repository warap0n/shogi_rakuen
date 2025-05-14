package controller_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"shogi-rakuen/controller"
	"shogi-rakuen/controller/dto"
	"shogi-rakuen/model"
	"shogi-rakuen/usecase"
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

func (d *DummyUserUsecase) Login(in input.LoginInput) (string, error) {
	if in.Email == "no-user@example.com" {
		return "", usecase.ErrEmailNotFound
	}
	if in.Password == "wrong" {
		return "", usecase.ErrInvalidPassword
	}
	if in.Password == "no-secret" {
		os.Unsetenv("JWT_SECRET")
		return "", usecase.ErrJWTSecretUnset
	}
	os.Setenv("JWT_SECRET", "test-secret")
	return "dummytoken", nil
}

func (d *DummyUserUsecase) GetUserByEmail(email string) (model.UserResponse, error) {
	switch email {
	case "notfound@example.com":
		return model.UserResponse{}, usecase.ErrEmailNotFound
	case "error@example.com":
		return model.UserResponse{}, errors.New("something went wrong")
	default:
		return model.UserResponse{
			ID:       42,
			Email:    email,
			Username: "user42",
			Rank:     "未設定",
		}, nil
	}
}

// --- SignUp ---
func TestSignUp_InvalidInput(t *testing.T) {
	e := echo.New()
	e.Validator = &customValidator{validator: validator.New()}
	uc := controller.NewUserController(&DummyUserUsecase{})
	e.POST("/signup", uc.SignUp)

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

	if assert.NoError(t, uc.SignUp(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "validation failed")
	}
}

func TestSignUp_Success(t *testing.T) {
	e := echo.New()
	e.Validator = &customValidator{validator: validator.New()}
	uc := controller.NewUserController(&DummyUserUsecase{})
	e.POST("/signup", uc.SignUp)

	body := dto.SignupRequest{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "secret123",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, uc.SignUp(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Body.String(), "\"email\":\"test@example.com\"")
	}
}

// --- Login ---
func TestLogin_InvalidInput(t *testing.T) {
	e := echo.New()
	e.Validator = &customValidator{validator: validator.New()}
	uc := controller.NewUserController(&DummyUserUsecase{})
	e.POST("/login", uc.Login)

	body := dto.LoginRequest{
		Email:    "",
		Password: "",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, uc.Login(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "validation failed")
	}
}

func TestLogin_EmailNotFound(t *testing.T) {
	e := echo.New()
	e.Validator = &customValidator{validator: validator.New()}
	uc := controller.NewUserController(&DummyUserUsecase{})
	e.POST("/login", uc.Login)

	body := dto.LoginRequest{
		Email:    "no-user@example.com",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, uc.Login(c)) {
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "email not found")
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	e := echo.New()
	e.Validator = &customValidator{validator: validator.New()}
	uc := controller.NewUserController(&DummyUserUsecase{})
	e.POST("/login", uc.Login)

	body := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "wrong",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, uc.Login(c)) {
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "invalid password")
	}
}

func TestLogin_JWTSecretUnset(t *testing.T) {
	e := echo.New()
	e.Validator = &customValidator{validator: validator.New()}
	uc := controller.NewUserController(&DummyUserUsecase{})
	e.POST("/login", uc.Login)

	body := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "no-secret",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, uc.Login(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "internal server error")
	}
}

// --- GetUserByEmail ---

// --- 成功ケース ---
func TestGetUserByEmail_Success(t *testing.T) {
	e := echo.New()
	uc := controller.NewUserController(&DummyUserUsecase{})
	e.GET("/users/:email", uc.GetUserByEmail)

	req := httptest.NewRequest(http.MethodGet, "/users/test@example.com", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("email")
	c.SetParamValues("test@example.com")

	// ハンドラ実行
	if assert.NoError(t, uc.GetUserByEmail(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp model.UserResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.EqualValues(t, 42, resp.ID)
		assert.Equal(t, "test@example.com", resp.Email)
		assert.Equal(t, "user42", resp.Username)
	}
}

// --- ユーザー未発見 (404) ケース ---
func TestGetUserByEmail_NotFound(t *testing.T) {
	e := echo.New()
	uc := controller.NewUserController(&DummyUserUsecase{})
	e.GET("/users/:email", uc.GetUserByEmail)

	req := httptest.NewRequest(http.MethodGet, "/users/notfound@example.com", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("email")
	c.SetParamValues("notfound@example.com")

	if assert.NoError(t, uc.GetUserByEmail(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "user not found")
	}
}

// --- 内部エラー (500) ケース ---
func TestGetUserByEmail_InternalError(t *testing.T) {
	e := echo.New()
	uc := controller.NewUserController(&DummyUserUsecase{})
	e.GET("/users/:email", uc.GetUserByEmail)

	req := httptest.NewRequest(http.MethodGet, "/users/error@example.com", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("email")
	c.SetParamValues("error@example.com")

	if assert.NoError(t, uc.GetUserByEmail(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "something went wrong")
	}
}

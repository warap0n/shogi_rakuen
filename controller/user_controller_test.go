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
	"github.com/golang-jwt/jwt/v5"
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
	return model.UserResponse{ID: 1, Email: in.Email, Username: in.Username, Rank: "未設定"}, nil
}

func (d *DummyUserUsecase) Login(in input.LoginInput) (string, error) {
	switch {
	case in.Email == "no-user@example.com":
		return "", usecase.ErrEmailNotFound
	case in.Password == "wrong":
		return "", usecase.ErrInvalidPassword
	case in.Password == "no-secret":
		os.Unsetenv("JWT_SECRET")
		return "", usecase.ErrJWTSecretUnset
	default:
		os.Setenv("JWT_SECRET", "test-secret")
		return "dummytoken", nil
	}
}

func (d *DummyUserUsecase) GetUserById(id uint) (model.UserResponse, error) {
	switch id {
	case 42:
		return model.UserResponse{ID: 42, Email: "me@example.com", Username: "meuser", Rank: "段位"}, nil
	case 100:
		return model.UserResponse{}, usecase.ErrUserNotFound
	default:
		return model.UserResponse{}, errors.New("db error")
	}
}

// setupEcho は共通の Echo インスタンスとバリデータを返す
func setupEcho() *echo.Echo {
	e := echo.New()
	e.Validator = &customValidator{validator: validator.New()}
	return e
}

// --- SignUp ---
func TestSignUp_InvalidInput(t *testing.T) {
	e := setupEcho()
	uc := controller.NewUserController(&DummyUserUsecase{})
	e.POST("/signup", uc.SignUp)

	body := dto.SignupRequest{Email: "", Username: "testuser", Password: "secret123"}
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
	e := setupEcho()
	uc := controller.NewUserController(&DummyUserUsecase{})
	e.POST("/signup", uc.SignUp)

	body := dto.SignupRequest{Email: "test@example.com", Username: "testuser", Password: "secret123"}
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

func TestLogin_Success(t *testing.T) {
	e := setupEcho()
	uc := controller.NewUserController(&DummyUserUsecase{})
	e.POST("/login", uc.Login)

	// DummyUserUsecase の Login では、Email != "no-user" && Password != "wrong" && Password != "no-secret" の場合
	// os.Setenv("JWT_SECRET","test-secret") → "dummytoken" を返すようになっています
	body := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "any-other-password",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 実行
	if assert.NoError(t, uc.Login(c)) {
		// ステータスとボディ
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{}`, rec.Body.String())

		// Cookie がセットされているか
		result := rec.Result()
		cookies := result.Cookies()
		assert.Len(t, cookies, 1)
		cookie := cookies[0]
		assert.Equal(t, "access_token", cookie.Name)
		assert.Equal(t, "dummytoken", cookie.Value)

		// HttpOnly / Path など属性もチェックしたければ追加で
		assert.True(t, cookie.HttpOnly)
		assert.Equal(t, "/", cookie.Path)
	}
}

func TestLogin_InvalidInput(t *testing.T) {
	e := setupEcho()
	uc := controller.NewUserController(&DummyUserUsecase{})
	e.POST("/login", uc.Login)

	body := dto.LoginRequest{Email: "", Password: ""}
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
	e := setupEcho()
	uc := controller.NewUserController(&DummyUserUsecase{})
	e.POST("/login", uc.Login)

	body := dto.LoginRequest{Email: "no-user@example.com", Password: "password123"}
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
	e := setupEcho()
	uc := controller.NewUserController(&DummyUserUsecase{})
	e.POST("/login", uc.Login)

	body := dto.LoginRequest{Email: "test@example.com", Password: "wrong"}
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
	e := setupEcho()
	uc := controller.NewUserController(&DummyUserUsecase{})
	e.POST("/login", uc.Login)

	body := dto.LoginRequest{Email: "test@example.com", Password: "no-secret"}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, uc.Login(c)) {
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "jwt secret not set")
	}
}

// --- Me ---
func TestMe_Success(t *testing.T) {
	e := setupEcho()
	uc := controller.NewUserController(&DummyUserUsecase{})
	e.GET("/me", uc.Me)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// 1) MapClaims で sub=42 のトークンを作成
	claims := jwt.MapClaims{"sub": "42"}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 2) ミドルウェアの代わりに直接コンテキストにセット
	c.Set("user", token)

	// 3) ハンドラ実行＆検証
	if assert.NoError(t, uc.Me(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp model.UserResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, uint(42), resp.ID)
		assert.Equal(t, "me@example.com", resp.Email)
	}
}

func TestMe_NotFound(t *testing.T) {
	e := setupEcho()
	uc := controller.NewUserController(&DummyUserUsecase{})
	e.GET("/me", uc.Me)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// MapClaims で存在しない ID=100 のトークンを作成
	claims := jwt.MapClaims{"sub": "100"}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	c.Set("user", token)

	if assert.NoError(t, uc.Me(c)) {
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "user not found")
	}
}

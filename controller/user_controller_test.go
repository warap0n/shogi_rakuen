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
	"shogi-rakuen/usecase/output"
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

func (d *DummyUserUsecase) SignUp(in input.SignupInput) (output.SignupOutput, error) {
	return output.SignupOutput{ID: 1, Email: in.Email, Username: in.Username, RankType: model.RankTypeDan, RankLevel: 2}, nil
}

func (d *DummyUserUsecase) Login(in input.LoginInput) (output.LoginOutput, error) {
	switch {
	case in.Email == "no-user@example.com":
		return output.LoginOutput{}, usecase.ErrEmailNotFound
	case in.Password == "wrong":
		return output.LoginOutput{}, usecase.ErrInvalidPassword
	case in.Password == "no-secret":
		os.Unsetenv("JWT_SECRET")
		return output.LoginOutput{}, usecase.ErrJWTSecretUnset
	default:
		os.Setenv("JWT_SECRET", "test-secret")
		return output.LoginOutput{Token: "dummytoken"}, nil
	}
}

func (d *DummyUserUsecase) GetUserById(id uint) (output.GetUserByIdOutput, error) {
	switch id {
	case 42:
		return output.GetUserByIdOutput{ID: 42, Email: "me@example.com", Username: "meuser", RankType: model.RankTypeDan, RankLevel: 2}, nil
	case 100:
		return output.GetUserByIdOutput{}, usecase.ErrUserNotFound
	default:
		return output.GetUserByIdOutput{}, errors.New("db error")
	}
}

// EmailExistsUsecase は SignUp 時に ErrEmailAlreadyExists を返すだけのモック
type EmailExistsUsecase struct{}

func (d *EmailExistsUsecase) SignUp(in input.SignupInput) (output.SignupOutput, error) {
	return output.SignupOutput{}, usecase.ErrEmailAlreadyExists
}
func (d *EmailExistsUsecase) Login(in input.LoginInput) (output.LoginOutput, error) {
	return output.LoginOutput{}, nil
}
func (d *EmailExistsUsecase) GetUserById(id uint) (output.GetUserByIdOutput, error) {
	return output.GetUserByIdOutput{}, nil
}

// setupEcho は共通の Echo インスタンスとバリデータを返す
func setupEcho() *echo.Echo {
	e := echo.New()
	e.Validator = &customValidator{validator: validator.New()}
	return e
}

// --- SignUp ---

func TestSignUp_BindError(t *testing.T) {
	e := setupEcho()
	uc := controller.NewUserController(&DummyUserUsecase{}) // バインドだけ失敗させる
	e.POST("/signup", uc.SignUp)

	// あえて JSON じゃない文字列を投げる
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader([]byte("xxx")))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NoError(t, uc.SignUp(c))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid input")
}

func TestSignUp_EmailAlreadyExists(t *testing.T) {
	e := setupEcho()
	uc := controller.NewUserController(&EmailExistsUsecase{})
	e.POST("/signup", uc.SignUp)

	body := dto.SignupRequest{
		Email:    "dup@example.com",
		Username: "dup",
		Password: "secret123",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NoError(t, uc.SignUp(c))
	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Contains(t, rec.Body.String(), "email already registered")
}

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
		assert.Contains(t, rec.Body.String(), "\"Email\":\"test@example.com\"")
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

package usecase_test

import (
	"errors"
	"os"
	"shogi-rakuen/model"
	"shogi-rakuen/repository"
	"shogi-rakuen/usecase"
	"shogi-rakuen/usecase/input"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// --- モック定義 ---

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUserById(id uint) (*model.User, error) {
	args := m.Called(id)
	u := args.Get(0)
	if u == nil {
		return nil, args.Error(1)
	}
	return u.(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	u := args.Get(0)
	if u == nil {
		return nil, args.Error(1)
	}
	return u.(*model.User), args.Error(1)
}

func (m *MockUserRepository) Create(user *model.User) (*model.User, error) {
	args := m.Called(user)
	u := args.Get(0)
	if u == nil {
		return nil, args.Error(1)
	}
	return u.(*model.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// --- SignUp ---

func TestSignUp_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := usecase.NewUserUsecase(mockRepo)

	input := input.SignupInput{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "plaintext123",
	}

	createdUser := &model.User{
		ID:       1,
		Email:    input.Email,
		Username: input.Username,
		Password: "hashedpass",
		Rank:     "未設定",
	}

	mockRepo.On("Create", mock.AnythingOfType("*model.User")).Return(createdUser, nil)

	result, err := uc.SignUp(input)

	assert.NoError(t, err)
	assert.Equal(t, createdUser.ID, result.ID)
	assert.Equal(t, createdUser.Email, result.Email)
	assert.Equal(t, createdUser.Username, result.Username)
	assert.Equal(t, createdUser.Rank, result.Rank)
	mockRepo.AssertExpectations(t)
}

func TestSignUp_CreateFails(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := usecase.NewUserUsecase(mockRepo)

	input := input.SignupInput{
		Email:    "fail@example.com",
		Username: "failuser",
		Password: "failpass",
	}

	mockRepo.On("Create", mock.AnythingOfType("*model.User")).Return(nil, errors.New("db error"))

	_, err := uc.SignUp(input)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSignUp_EmailAlreadyExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := usecase.NewUserUsecase(mockRepo)

	input := input.SignupInput{
		Email:    "test@example.com",
		Username: "tester",
		Password: "plaintext123",
	}

	mockRepo.On("Create", mock.AnythingOfType("*model.User")).
		Return(nil, repository.ErrEmailAlreadyExists)

	_, err := uc.SignUp(input)

	assert.ErrorIs(t, err, usecase.ErrEmailAlreadyExists)
	mockRepo.AssertExpectations(t)
}

// --- Login ---

func TestLogin_Success(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	mockRepo := new(MockUserRepository)
	uc := usecase.NewUserUsecase(mockRepo)

	plain := "correct-password"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)

	user := &model.User{
		ID:       1,
		Email:    "test@example.com",
		Password: string(hashed),
	}

	mockRepo.On("GetUserByEmail", "test@example.com").Return(user, nil)

	input := input.LoginInput{
		Email:    "test@example.com",
		Password: plain,
	}

	token, err := uc.Login(input)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	mockRepo.AssertExpectations(t)
}

func TestLogin_EmailNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := usecase.NewUserUsecase(mockRepo)

	mockRepo.On("GetUserByEmail", "nope@example.com").Return(nil, errors.New("not found"))

	input := input.LoginInput{
		Email:    "nope@example.com",
		Password: "any",
	}

	token, err := uc.Login(input)
	assert.Error(t, err)
	assert.Equal(t, "", token)
	assert.Equal(t, usecase.ErrEmailNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestLogin_InvalidPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := usecase.NewUserUsecase(mockRepo)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	user := &model.User{
		ID:       1,
		Email:    "test@example.com",
		Password: string(hashed),
	}

	mockRepo.On("GetUserByEmail", "test@example.com").Return(user, nil)

	input := input.LoginInput{
		Email:    "test@example.com",
		Password: "wrong-password",
	}

	token, err := uc.Login(input)
	assert.Error(t, err)
	assert.Equal(t, "", token)
	assert.Equal(t, usecase.ErrInvalidPassword, err)
	mockRepo.AssertExpectations(t)
}

func TestLogin_JWTSignFail(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := usecase.NewUserUsecase(mockRepo)

	plain := "pass"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)

	user := &model.User{
		ID:       1,
		Email:    "test@example.com",
		Password: string(hashed),
	}

	mockRepo.On("GetUserByEmail", "test@example.com").Return(user, nil)

	os.Unsetenv("JWT_SECRET")

	input := input.LoginInput{
		Email:    "test@example.com",
		Password: plain,
	}

	token, err := uc.Login(input)
	assert.Error(t, err)
	assert.Equal(t, "", token)
	assert.Equal(t, usecase.ErrJWTSecretUnset, err)
	mockRepo.AssertExpectations(t)
}

// --- GetUserById のテスト ---

func TestGetUserById_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := usecase.NewUserUsecase(mockRepo)

	user := &model.User{
		ID:       42,
		Email:    "foo@bar",
		Username: "foobar",
		Rank:     "三段",
	}

	// モックは ID=42 でこの user を返す
	mockRepo.On("GetUserById", uint(42)).Return(user, nil)

	resp, err := uc.GetUserById(42)
	assert.NoError(t, err)
	assert.Equal(t, uint(42), resp.ID)
	assert.Equal(t, "foo@bar", resp.Email)
	assert.Equal(t, "foobar", resp.Username)
	assert.Equal(t, "三段", resp.Rank)

	mockRepo.AssertExpectations(t)
}

func TestGetUserById_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := usecase.NewUserUsecase(mockRepo)

	// 存在しない ID を渡すとリポジトリは ErrUserNotFound
	mockRepo.On("GetUserById", uint(999)).Return(nil, repository.ErrUserNotFound)

	_, err := uc.GetUserById(999)
	assert.ErrorIs(t, err, usecase.ErrUserNotFound)

	mockRepo.AssertExpectations(t)
}

func TestGetUserById_OtherError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	uc := usecase.NewUserUsecase(mockRepo)

	mockErr := errors.New("db error")
	mockRepo.On("GetUserById", uint(1)).Return(nil, mockErr)

	_, err := uc.GetUserById(1)
	assert.Equal(t, mockErr, err)

	mockRepo.AssertExpectations(t)
}

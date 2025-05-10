package repository_test

import (
	"testing"

	"shogi-rakuen/model"
	"shogi-rakuen/repository"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUserRepository_Create(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	sqlDB, err := db.DB()
	assert.NoError(t, err)
	defer sqlDB.Close()

	err = db.AutoMigrate(&model.User{})
	assert.NoError(t, err)

	repo := repository.NewUserRepository(db)

	user := &model.User{
		Email:    "test@example.com",
		Username: "tester",
		Password: "hashed123",
	}

	created, err := repo.Create(user)
	assert.NoError(t, err)
	assert.NotZero(t, created.ID)
	assert.Equal(t, "test@example.com", created.Email)
}

func TestUserRepository_GetUserByEmail(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	sqlDB, err := db.DB()
	assert.NoError(t, err)
	defer sqlDB.Close()

	err = db.AutoMigrate(&model.User{})
	assert.NoError(t, err)

	repo := repository.NewUserRepository(db)

	// --- 正常系: 先にユーザーを作成 ---
	user := &model.User{
		Email:    "test@example.com",
		Username: "tester",
		Password: "hashed123",
	}
	_, err = repo.Create(user)
	assert.NoError(t, err)

	// 実際に GetUserByEmail を使う
	found, err := repo.GetUserByEmail("test@example.com")
	assert.NoError(t, err)
	assert.Equal(t, "tester", found.Username)
	assert.Equal(t, user.Email, found.Email)
}

func TestUserRepository_GetUserByEmail_NotFound(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	sqlDB, err := db.DB()
	assert.NoError(t, err)
	defer sqlDB.Close()

	err = db.AutoMigrate(&model.User{})
	assert.NoError(t, err)

	repo := repository.NewUserRepository(db)

	// 該当するメールアドレスなし
	_, err = repo.GetUserByEmail("noone@example.com")
	assert.ErrorIs(t, err, repository.ErrUserNotFound)
}

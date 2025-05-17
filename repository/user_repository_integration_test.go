package repository_test

import (
	"shogi-rakuen/model"
	"shogi-rakuen/repository"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUserRepository_Create(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	err = db.AutoMigrate(&model.User{})
	require.NoError(t, err)

	repo := repository.NewUserRepository(db)

	user := &model.User{
		Email:    "test@example.com",
		Username: "tester",
		Password: "hashed123",
	}

	created, err := repo.Create(user)
	require.NoError(t, err)
	assert.NotZero(t, created.ID)
	assert.Equal(t, "test@example.com", created.Email)
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	err = db.AutoMigrate(&model.User{})
	require.NoError(t, err)

	repo := repository.NewUserRepository(db)

	// 最初のユーザーを作成
	user1 := &model.User{
		Email:    "duplicate@example.com",
		Username: "user1",
		Password: "pass1",
	}
	_, err = repo.Create(user1)
	require.NoError(t, err)

	// 同じメールアドレスで2人目のユーザーを作成
	user2 := &model.User{
		Email:    "duplicate@example.com",
		Username: "user2",
		Password: "pass2",
	}
	_, err = repo.Create(user2)

	// repository.ErrEmailAlreadyExists が返ってくることを期待
	assert.ErrorIs(t, err, repository.ErrEmailAlreadyExists)
}

func TestUserRepository_GetUserById_Success(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	require.NoError(t, db.AutoMigrate(&model.User{}))

	repo := repository.NewUserRepository(db)

	// ユーザー作成
	u := &model.User{
		Email:    "foo@example.com",
		Username: "foo",
		Password: "pw",
	}
	created, err := repo.Create(u)
	require.NoError(t, err)
	require.NotZero(t, created.ID)

	// ID で取得
	found, err := repo.GetUserById(created.ID)
	assert.NoError(t, err)
	assert.Equal(t, created.Email, found.Email)
	assert.Equal(t, created.Username, found.Username)
}

func TestUserRepository_GetUserById_NotFound(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	require.NoError(t, db.AutoMigrate(&model.User{}))

	repo := repository.NewUserRepository(db)

	// 存在しない ID を指定
	_, err = repo.GetUserById(999)
	assert.ErrorIs(t, err, repository.ErrUserNotFound)
}

func TestUserRepository_GetUserByEmail(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	require.NoError(t, db.AutoMigrate(&model.User{}))

	repo := repository.NewUserRepository(db)

	// --- 正常系: 先にユーザーを作成 ---
	user := &model.User{
		Email:    "test@example.com",
		Username: "tester",
		Password: "hashed123",
	}
	_, err = repo.Create(user)
	require.NoError(t, err)

	// 実際に GetUserByEmail を使う
	found, err := repo.GetUserByEmail("test@example.com")
	assert.NoError(t, err)
	assert.Equal(t, "tester", found.Username)
	assert.Equal(t, user.Email, found.Email)
}

func TestUserRepository_GetUserByEmail_NotFound(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	defer sqlDB.Close()

	require.NoError(t, db.AutoMigrate(&model.User{}))

	repo := repository.NewUserRepository(db)

	// 該当するメールアドレスなし
	_, err = repo.GetUserByEmail("noone@example.com")
	assert.ErrorIs(t, err, repository.ErrUserNotFound)
}

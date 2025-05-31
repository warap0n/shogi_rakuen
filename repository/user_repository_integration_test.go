package repository_test

import (
	"testing"

	"shogi-rakuen/model"
	"shogi-rakuen/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB はインメモリ SQLite DB を初期化し、マイグレーションまで行います。
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { sqlDB.Close() })

	require.NoError(t, db.AutoMigrate(&model.User{}))
	return db
}

func createTestUser(t *testing.T, repo repository.IUserRepository, email, username, password string) *model.User {
	u := &model.User{Email: email, Username: username, Password: password}
	created, err := repo.Create(u)
	require.NoError(t, err)
	require.NotZero(t, created.ID)
	return created
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	u := createTestUser(t, repo, "test@example.com", "tester", "hashed123")
	assert.Equal(t, "test@example.com", u.Email)
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	createTestUser(t, repo, "dup@example.com", "user1", "pass1")

	// 同じメールで再度作成 ⇒ ErrEmailAlreadyExists
	_, err := repo.Create(&model.User{
		Email:    "dup@example.com",
		Username: "user2",
		Password: "pass2",
	})
	assert.ErrorIs(t, err, repository.ErrEmailAlreadyExists)
}

func TestUserRepository_GetUserById_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	orig := createTestUser(t, repo, "foo@example.com", "foo", "pw")

	found, err := repo.GetUserById(orig.ID)
	assert.NoError(t, err)
	assert.Equal(t, orig.Email, found.Email)
	assert.Equal(t, orig.Username, found.Username)
}

func TestUserRepository_GetUserById_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	_, err := repo.GetUserById(999) // 存在しないID
	assert.ErrorIs(t, err, repository.ErrUserNotFound)
}

func TestUserRepository_GetUserByEmail_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	createTestUser(t, repo, "bar@example.com", "bar", "pw123")

	found, err := repo.GetUserByEmail("bar@example.com")
	assert.NoError(t, err)
	assert.Equal(t, "bar", found.Username)
}

func TestUserRepository_GetUserByEmail_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	_, err := repo.GetUserByEmail("noone@example.com")
	assert.ErrorIs(t, err, repository.ErrUserNotFound)
}

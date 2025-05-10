package repository

import (
	"shogi-rakuen/model"

	"gorm.io/gorm"
)

type IUserRepository interface {
	GetUserByEmail(email string) (*model.User, error)
	Create(user *model.User) (*model.User, error)
	Update(user *model.User) error
	Delete(id uint) error
}
type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) IUserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *model.User) (*model.User, error) {
	// 本番実装が未定ならとりあえず nil 返すだけでOK
	return nil, nil
}

func (r *userRepository) GetUserByEmail(email string) (*model.User, error) {
	return nil, nil
}

func (r *userRepository) Update(user *model.User) error {
	return nil
}

func (r *userRepository) Delete(id uint) error {
	return nil
}

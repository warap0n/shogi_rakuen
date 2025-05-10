package repository

import (
	"errors"
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
	if err := r.db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *model.User) error {
	return nil
}

func (r *userRepository) Delete(id uint) error {
	return nil
}

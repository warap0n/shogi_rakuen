package repository

import "errors"

var (
	ErrUserNotFound = errors.New("does not exist email")
)

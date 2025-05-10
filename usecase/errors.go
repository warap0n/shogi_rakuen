package usecase

import "errors"

var (
	ErrEmailNotFound      = errors.New("email not found")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrJWTSecretUnset     = errors.New("JWT_SECRET is not set")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

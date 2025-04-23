package repository

import (
	"auth-service/api/calltypes"
)

type Repository interface {
	GetAll() ([]*calltypes.User, error)
	GetByEmail(email string) (*calltypes.User, error)
	GetOne(id int) (*calltypes.User, error)
	Update(user calltypes.User) error
	Insert(user calltypes.User) (int, error)
	PasswordMatches(plainText string, user calltypes.User) (bool, error)
	EmailCheck(email string) (*calltypes.User, error)
	StoreRefreshToken(id int, hashedToken string) error
	ValidateRefreshToken(rawToken, clientIP string, id int) (bool, error)
	UpdateRefreshToken(id int, rawToken string) error
}

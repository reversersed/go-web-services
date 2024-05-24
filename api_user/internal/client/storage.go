package client

import "context"

type Storage interface {
	FindByLogin(ctx context.Context, login string) (*User, error)
	FindById(ctx context.Context, id string) (*User, error)
	AddUser(ctx context.Context, user *User) (string, error)
}

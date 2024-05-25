package client

import "context"

type Storage interface {
	FindByLogin(ctx context.Context, login string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindById(ctx context.Context, id string) (*User, error)
	ApproveUserEmail(ctx context.Context, id string) error
	AddUser(ctx context.Context, user *User) (string, error)
}

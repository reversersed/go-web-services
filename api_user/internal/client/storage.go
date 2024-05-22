package client

import "context"

type Storage interface {
	FindByLogin(ctx context.Context, login string) (*User, error)
}

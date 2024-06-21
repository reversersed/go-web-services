package client

import "context"

//go:generate mockgen -source=storage.go -destination=mocks/storage.go

type Storage interface {
	AddBook(ctx context.Context, book *Book) (string, error)
}

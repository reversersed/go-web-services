package client

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

//go:generate mockgen -source=storage.go -destination=mocks/storage.go

type Storage interface {
	AddBook(ctx context.Context, book *Book) (string, error)
	GetBookByName(ctx context.Context, name string) (*Book, error)
	GetBookById(ctx context.Context, id primitive.ObjectID) (*Book, error)
	GetByFilter(ctx context.Context, filter map[string]string, offset, limit int) ([]*Book, error)
}

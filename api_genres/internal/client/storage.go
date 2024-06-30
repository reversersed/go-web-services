package client

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

//go:generate mockgen -source=storage.go -destination=mocks/storage.go

type Storage interface {
	GetGenre(ctx context.Context, id []primitive.ObjectID) ([]*Genre, error)
	AddGenre(ctx context.Context, genre *Genre) (*Genre, error)
	GetAllGenres(ctx context.Context) ([]*Genre, error)
}

package client

import "context"

//go:generate mockgen -source=storage.go -destination=mocks/storage.go

type Storage interface {
	GetGenre(ctx context.Context, id string) (*Genre, error)
	AddGenre(ctx context.Context, genre *Genre) (*Genre, error)
	GetAllGenres(ctx context.Context) ([]*Genre, error)
}

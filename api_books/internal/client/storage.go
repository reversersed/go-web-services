package client

//go:generate mockgen -source=storage.go -destination=mocks/storage.go

type Storage interface {
}

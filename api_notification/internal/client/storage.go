package client

import "context"

//go:generate mockgen -source=storage.go -destination=mocks/storage.go

type Storage interface {
	SendNotification(ctx context.Context, notif *Notification, user_id string) error
	IsUserExists(ctx context.Context, user_id string) (bool, error)
	CreateUser(ctx context.Context, user_id, login string) error
	DeleteUser(ctx context.Context, user_id string) error
	ChangeUserLogin(ctx context.Context, user_id string, newLogin string) error
}

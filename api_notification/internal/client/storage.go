package client

import "context"

type Storage interface {
	SendNotification(ctx context.Context, notif *Notification, user_id string) error
	IsUserExists(ctx context.Context, user_id string) (bool, error)
	CreateUser(ctx context.Context, user_id, login string) error
}

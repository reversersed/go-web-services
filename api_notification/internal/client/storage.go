package client

import "context"

type Storage interface {
	SendNotification(ctx context.Context, notif *Notification, user_id string) error
}

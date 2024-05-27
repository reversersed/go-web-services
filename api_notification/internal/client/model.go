package client

import "go.mongodb.org/mongo-driver/bson/primitive"

type NotificationType string

const (
	Info     NotificationType = "info"
	Warning  NotificationType = "warn"
	Security NotificationType = "security"
)

type Notification struct {
	Sended  primitive.Timestamp `json:"sended,omitempty" bson:"sended"`
	Content string              `json:"content" bson:"content"`
	Type    NotificationType    `json:"type" bson:"type"`
}
type User struct {
	Id            primitive.ObjectID `json:"id" bson:"id"`
	Login         string             `json:"login" bson:"login"`
	Notifications []*Notification    `json:"notifications" bson:"notifications"`
}

type SendNotificationMessage struct {
	UserId  string           `json:"userid" validate:"required,primitiveid"`
	Content string           `json:"content" validate:"required"`
	Type    NotificationType `json:"type" validate:"required,oneof=info warn security"`
}
type UserLoginChangedMessage struct {
	UserId   string `json:"userid" validate:"required,primitiveid"`
	NewLogin string `json:"newlogin" validate:"required"`
}

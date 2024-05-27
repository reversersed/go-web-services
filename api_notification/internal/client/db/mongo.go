package db

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/reversersed/go-web-services/tree/main/api_notification/internal/client"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/logging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type db struct {
	*sync.Mutex
	collection *mongo.Collection
	logger     *logging.Logger
}

func NewStorage(storage *mongo.Database, collection string, logger *logging.Logger) client.Storage {
	db := &db{
		collection: storage.Collection(collection),
		logger:     logger,
	}
	return db
}
func (d *db) CreateUser(ctx context.Context, user_id, login string) error {
	d.Lock()
	defer d.Unlock()
	id, err := primitive.ObjectIDFromHex(user_id)
	if err != nil {
		return err
	}
	user := &client.User{
		Id:            id,
		Login:         login,
		Notifications: []*client.Notification{},
	}
	_, err = d.collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	return nil
}
func (d *db) IsUserExists(ctx context.Context, user_id string) (bool, error) {
	d.Lock()
	defer d.Unlock()
	id, err := primitive.ObjectIDFromHex(user_id)
	if err != nil {
		return false, err
	}
	filter := bson.M{"id": id}
	result := d.collection.FindOne(ctx, filter)
	if err = result.Err(); err != nil {
		return false, nil
	}
	return true, nil
}
func (d *db) SendNotification(ctx context.Context, notif *client.Notification, user_id string) error {
	d.Lock()
	defer d.Unlock()
	id, err := primitive.ObjectIDFromHex(user_id)
	if err != nil {
		return err
	}
	filter := bson.M{"id": id}
	result := d.collection.FindOne(ctx, filter)
	var u client.User
	if err = result.Err(); err != nil {
		return fmt.Errorf("user does not exist: %v", err)
	}
	err = result.Decode(u)
	if err != nil {
		return err
	}
	notif.Sended = primitive.Timestamp{T: uint32(time.Now().UTC().Unix()), I: 0}

	u.Notifications = append([]*client.Notification{notif}, u.Notifications...)

	upd_result, err := d.collection.UpdateByID(ctx, id, u)
	if err != nil {
		return err
	}
	if upd_result.MatchedCount == 0 || upd_result.ModifiedCount == 0 {
		return fmt.Errorf("user has not been updated")
	}
	return nil
}

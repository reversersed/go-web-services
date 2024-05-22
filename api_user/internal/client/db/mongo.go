package db

import (
	"context"
	"time"

	"github.com/reversersed/go-web-services/tree/main/api_user/internal/client"
	"github.com/reversersed/go-web-services/tree/main/api_user/pkg/logging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type db struct {
	collection *mongo.Collection
	logger     *logging.Logger
}

func NewStorage(storage *mongo.Database, collection string, logger *logging.Logger) client.Storage {
	return &db{
		collection: storage.Collection(collection),
		logger:     logger,
	}
}

func (d *db) FindByLogin(ctx context.Context, login string) (*client.User, error) {
	filter := bson.M{"login": login}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result := d.collection.FindOne(ctx, filter)
	if err := result.Err(); err != nil {
		d.logger.Warnf("error while fetching user from db: %v", err)
		return nil, err
	}
	var u client.User
	if err := result.Decode(&u); err != nil {
		return nil, err
	}
	return &u, nil
}

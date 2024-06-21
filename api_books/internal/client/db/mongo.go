package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/reversersed/go-web-services/tree/main/api_books/internal/client"
	"github.com/reversersed/go-web-services/tree/main/api_books/pkg/logging"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type db struct {
	sync.Mutex
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

func (d *db) AddBook(ctx context.Context, book *client.Book) (string, error) {
	context, cancel := context.WithCancel(ctx)
	defer cancel()

	result, err := d.collection.InsertOne(context, book)
	if err != nil {
		return "", err
	}
	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("error while accessing user id")
	}

	return id.Hex(), nil
}

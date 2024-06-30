package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/reversersed/go-web-services/tree/main/api_books/internal/client"
	"github.com/reversersed/go-web-services/tree/main/api_books/pkg/logging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type db struct {
	sync.RWMutex
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
func (d *db) GetBookByName(ctx context.Context, name string) (*client.Book, error) {
	d.RLock()
	defer d.RUnlock()

	result := d.collection.FindOne(ctx, bson.M{"name": name})

	if err := result.Err(); err != nil {
		return nil, err
	}
	var book client.Book
	err := result.Decode(&book)
	if err != nil {
		return nil, err
	}
	return &book, nil
}
func (d *db) AddBook(ctx context.Context, book *client.Book) (string, error) {
	d.Lock()
	defer d.Unlock()

	result, err := d.collection.InsertOne(ctx, book)
	if err != nil {
		return "", err
	}
	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("error while accessing user id")
	}

	return id.Hex(), nil
}

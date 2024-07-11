package db

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/reversersed/go-web-services/tree/main/api_books/internal/client"
	"github.com/reversersed/go-web-services/tree/main/api_books/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_books/pkg/logging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
func (d *db) GetBookById(ctx context.Context, id primitive.ObjectID) (*client.Book, error) {
	d.RLock()
	defer d.RUnlock()

	filter := bson.M{"_id": id}
	result := d.collection.FindOne(ctx, filter)
	if err := result.Err(); errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errormiddleware.NotFoundError([]string{"book not exists"}, err.Error())
	} else if err != nil {
		return nil, err
	}

	var book client.Book
	result.Decode(&book)

	return &book, nil
}
func (d *db) GetByFilter(ctx context.Context, filter map[string]string, offset, limit int) ([]*client.Book, error) {
	d.RLock()
	defer d.RUnlock()

	options := options.Find().SetSkip(int64(offset)).SetLimit(int64(limit)).SetSort(bson.M{"name": -1})
	filters := make(bson.D, 0)

	for i, v := range filter {
		switch i {
		case "year":
			filters = append(filters, bson.E{Key: "year", Value: bson.M{"$gte": v}})
		default:
			return nil, errormiddleware.BadRequestError([]string{"invalid filter received"}, fmt.Sprintf("filter %s: %s is unsupported", i, v))
		}
	}

	result, err := d.collection.Find(ctx, filters, options)
	if err != nil {
		return nil, err
	}
	if err := result.Err(); errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errormiddleware.NotFoundError([]string{"no books found"}, result.Err().Error())
	} else if err != nil {
		return nil, err
	}

	var books []*client.Book
	err = result.All(ctx, &books)
	if err != nil {
		return nil, err
	}
	if len(books) == 0 {
		return nil, errormiddleware.NotFoundError([]string{"no books found"}, "document slice containes 0 items")
	}
	return books, nil
}

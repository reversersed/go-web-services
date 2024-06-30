package db

import (
	"context"
	"errors"
	"sync"

	"github.com/reversersed/go-web-services/tree/main/api_genres/internal/client"
	"github.com/reversersed/go-web-services/tree/main/api_genres/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_genres/pkg/logging"
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

func (d *db) GetGenre(ctx context.Context, id string) (*client.Genre, error) {
	d.RLock()
	defer d.RUnlock()

	filter := bson.M{"_id": id}
	result := d.collection.FindOne(ctx, filter)

	if err := result.Err(); err != nil {
		return nil, errormiddleware.NotFoundError([]string{"no genre with provided id"}, err.Error())
	}

	var genre client.Genre
	err := result.Decode(&genre)
	if err != nil {
		return nil, err
	}

	return &genre, nil
}
func (d *db) AddGenre(ctx context.Context, genre *client.Genre) (*client.Genre, error) {
	d.Lock()
	defer d.Unlock()

	result, err := d.collection.InsertOne(ctx, genre)
	if err != nil {
		return nil, err
	}

	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, errors.New("cannot get inserted document id")
	}
	genre.Id = id

	return genre, nil
}
func (d *db) GetAllGenres(ctx context.Context) ([]*client.Genre, error) {
	d.RLock()
	defer d.RUnlock()

	result, err := d.collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, errormiddleware.NotFoundError([]string{"there's no genres"}, err.Error())
	}

	var genres []*client.Genre
	err = result.All(ctx, &genres)
	if err != nil {
		return nil, err
	}

	return genres, err
}

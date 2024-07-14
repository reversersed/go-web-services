package db

import (
	"context"
	"errors"
	"sync"
	"time"

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
	defer db.seedGenres()
	return db
}
func (d *db) seedGenres() {
	d.Lock()
	defer d.Unlock()

	genreCount, _ := d.collection.CountDocuments(context.Background(), bson.D{})
	if genreCount > 0 {
		d.logger.Infof("there is %d genres, seeding canceled", genreCount)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	preinstalledGenres := []string{"Детектив", "Фантастика", "Комиксы", "Бизнес-менеджмент", "Хобби", "История", "Легкое чтение", "Серьезное чтение"}

	d.logger.Infof("trying to seed %d genres...", len(preinstalledGenres))
	for _, v := range preinstalledGenres {
		d.logger.Infof("seeding genre %s...", v)
		genre := &client.Genre{
			Name: v,
		}
		response, err := d.collection.InsertOne(ctx, genre)
		if err != nil {
			d.logger.Fatalf("cannot seed genre: %v", err)
		}
		id, ok := response.InsertedID.(primitive.ObjectID)
		if !ok {
			d.logger.Fatalf("can't create id for genre")
		}
		d.logger.Infof("genre %s seeded with id %v", v, id.Hex())
	}
	d.logger.Infof("seeding manual genres...")
	hex, _ := primitive.ObjectIDFromHex("6690e6dcfd658345b06c2a25")
	d.collection.InsertOne(ctx, &client.Genre{Id: hex, Name: "Детские книги"})
	hex, _ = primitive.ObjectIDFromHex("6690e6dcfd658345b06c2a12")
	d.collection.InsertOne(ctx, &client.Genre{Id: hex, Name: "Фентези"})
	d.logger.Infof("genres seeded")
}
func (d *db) GetGenre(ctx context.Context, id []primitive.ObjectID) ([]*client.Genre, error) {
	d.RLock()
	defer d.RUnlock()

	filter := bson.M{"_id": bson.M{"$in": id}}
	result, err := d.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	if err := result.Err(); err != nil {
		return nil, errormiddleware.NotFoundError([]string{"no genre with provided id"}, err.Error())
	}

	var genre []*client.Genre
	err = result.All(ctx, &genre)
	if err != nil {
		return nil, err
	}

	return genre, nil
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
		return nil, err
	}

	var genres []*client.Genre
	err = result.All(ctx, &genres)
	if err != nil {
		return nil, err
	}
	if len(genres) == 0 {
		return nil, errormiddleware.NotFoundError([]string{"there's no genres"}, "marshalled array contained 0 elements")
	}
	return genres, nil
}

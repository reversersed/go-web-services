package db

import (
	"sync"

	"github.com/reversersed/go-web-services/tree/main/api_genres/internal/client"
	"github.com/reversersed/go-web-services/tree/main/api_genres/pkg/logging"
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

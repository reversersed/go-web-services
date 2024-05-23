package db

import (
	"context"
	"time"

	"github.com/reversersed/go-web-services/tree/main/api_user/internal/client"
	"github.com/reversersed/go-web-services/tree/main/api_user/pkg/logging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type db struct {
	collection *mongo.Collection
	logger     *logging.Logger
}

func NewStorage(storage *mongo.Database, collection string, logger *logging.Logger) client.Storage {
	db := &db{
		collection: storage.Collection(collection),
		logger:     logger,
	}
	defer db.seedAdminAccount()
	return db
}
func (d *db) seedAdminAccount() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := d.collection.FindOne(ctx, bson.M{"login": "admin"})
	if err := result.Err(); err != nil {
		d.logger.Info("starting seeding admin account...")
		pass, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.MinCost)
		admin := &client.User{
			Login:    "admin",
			Password: pass,
			Roles:    []string{"user", "admin"},
		}
		response, err := d.collection.InsertOne(ctx, admin)
		if err != nil {
			d.logger.Fatalf("cannot seed admin account: %v", err)
		}
		id, ok := response.InsertedID.(primitive.ObjectID)
		if !ok {
			d.logger.Fatalf("can't create id for admin document")
		}
		d.logger.Infof("admin account seeded with id %v", id.Hex())
		return
	}
	d.logger.Info("admin account exists. seed not executed")
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

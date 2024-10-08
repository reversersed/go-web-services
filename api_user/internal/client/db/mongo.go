package db

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/reversersed/go-web-services/tree/main/api_user/internal/client"
	"github.com/reversersed/go-web-services/tree/main/api_user/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_user/pkg/logging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
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
	defer db.seedAdminAccount()
	return db
}
func (d *db) seedAdminAccount() {
	d.Lock()
	defer d.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := d.collection.FindOne(ctx, bson.M{"login": "admin"})
	if err := result.Err(); err != nil {
		d.logger.Info("starting seeding admin account...")
		pass, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.MinCost)
		admin := &client.User{
			Login:          "admin",
			Password:       pass,
			Roles:          []string{"user", "admin"},
			Email:          "admin@example.com",
			EmailConfirmed: true,
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
func (d *db) ApproveUserEmail(ctx context.Context, id string) error {
	d.Lock()
	defer d.Unlock()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	obj_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	result, err := d.collection.UpdateByID(ctx, obj_id, bson.M{"$set": bson.M{"emailconfirmed": true}})
	if err != nil {
		return err
	}
	if result.ModifiedCount == 0 {
		if result.MatchedCount == 0 {
			return errormiddleware.NotFoundError([]string{"user does not exists"}, "database returned no matching for provided id")
		}
		return errormiddleware.NotFoundError([]string{"user's email already approved"}, "database found user, but didn't update it")
	}
	return nil
}
func (d *db) FindByLogin(ctx context.Context, login string) (*client.User, error) {
	d.RLock()
	defer d.RUnlock()

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
func (d *db) FindByEmail(ctx context.Context, email string) (*client.User, error) {
	d.RLock()
	defer d.RUnlock()

	filter := bson.M{"email": email}

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
func (d *db) AddUser(ctx context.Context, user *client.User) (string, error) {
	d.Lock()
	defer d.Unlock()

	contx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	result, err := d.collection.InsertOne(contx, user)
	if err != nil {
		d.logger.Warnf("error while user creation: %v", err)
		return "", err
	}
	oid, ok := result.InsertedID.(primitive.ObjectID)
	if ok {
		return oid.Hex(), nil
	}
	d.logger.Warnf("cant get created user id: %v (%v)", oid.Hex(), oid)
	return "", fmt.Errorf("cant resolve user id")
}
func (d *db) FindById(ctx context.Context, id string) (*client.User, error) {
	d.RLock()
	defer d.RUnlock()

	primitive_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	filter := bson.M{"_id": primitive_id}

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
func (d *db) DeleteUser(ctx context.Context, userId string) error {
	d.Lock()
	defer d.Unlock()
	
	primitive_id, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": primitive_id}

	result, err := d.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("user with provided id does not exist")
	}
	return nil
}
func (d *db) ChangeUserLogin(ctx context.Context, userId, newLogin string) error {
	d.Lock()
	defer d.Unlock()
	
	primitive_id, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return err
	}
	update := bson.M{"$set": bson.M{"login": newLogin, "logincooldown": time.Now().UTC().Unix() + (31 * 86400)}}
	result, err := d.collection.UpdateByID(ctx, primitive_id, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errormiddleware.NotFoundError([]string{"user with provided id not found"}, fmt.Sprintf("matched count was == %d", result.MatchedCount))
	}
	if result.ModifiedCount == 0 {
		return errormiddleware.NotFoundError([]string{"no user was updated"}, fmt.Sprintf("modified count was == %d", result.ModifiedCount))
	}
	return nil
}

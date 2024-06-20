package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/reversersed/go-web-services/tree/main/api_books/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewClient(ctx context.Context, cfg *config.DatabaseConfig) (*mongo.Database, error) {
	var mongoURL string
	var anonymous bool

	if cfg.Db_Name == "" || cfg.Db_Pass == "" {
		anonymous = true
		mongoURL = fmt.Sprintf("mongodb://%s:%d", cfg.Db_Host, cfg.Db_Port)
	} else {
		mongoURL = fmt.Sprintf("mongodb://%s:%s@%s:%d", cfg.Db_Name, cfg.Db_Pass, cfg.Db_Host, cfg.Db_Port)
	}
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	clientOptions := options.Client().ApplyURI(mongoURL)
	if !anonymous {
		clientOptions.SetAuth(options.Credential{
			Username:    cfg.Db_Name,
			Password:    cfg.Db_Pass,
			PasswordSet: true,
			AuthSource:  cfg.Db_Auth,
		})
	}
	client, err := mongo.Connect(reqCtx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb: %w", err)
	}
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	return client.Database(cfg.Db_Base), nil
}

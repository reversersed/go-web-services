package client

import (
	"context"

	"github.com/reversersed/go-web-services/tree/main/api_genres/pkg/cache"
	"github.com/reversersed/go-web-services/tree/main/api_genres/pkg/logging"
	valid "github.com/reversersed/go-web-services/tree/main/api_genres/pkg/validator"
)

type service struct {
	storage   Storage
	logger    *logging.Logger
	cache     cache.Cache
	validator *valid.Validator
}

func NewService(storage Storage, logger *logging.Logger, cache cache.Cache, validator *valid.Validator) *service {
	return &service{storage: storage, logger: logger, cache: cache, validator: validator}
}

func (s *service) GetAllGenres(ctx context.Context) ([]*Genre, error) {
	
}

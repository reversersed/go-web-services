package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/reversersed/go-web-services/tree/main/api_genres/pkg/cache"
	"github.com/reversersed/go-web-services/tree/main/api_genres/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_genres/pkg/logging"
	valid "github.com/reversersed/go-web-services/tree/main/api_genres/pkg/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
func (s *service) GetGenre(ctx context.Context, id string) ([]*Genre, error) {
	cntx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	ids := strings.Split(id, ",")
	primitives := make([]primitive.ObjectID, 0, len(ids))
	genre := make([]*Genre, 0, len(ids))
	var err error
	for _, cnvrt := range ids {
		hex, err := primitive.ObjectIDFromHex(cnvrt)
		if err != nil {
			return nil, errormiddleware.BadRequestError([]string{"wrong request params"}, fmt.Sprintf("can't convert value %s to object hex. Must be primitive id. %v", cnvrt, err))
		}
		bytes, err := s.cache.Get([]byte(cnvrt))
		if err == nil {
			var g Genre
			json.Unmarshal(bytes, &g)
			genre = append(genre, &g)
		}
		primitives = append(primitives, hex)
	}
	if len(genre) < len(ids) {
		genre, err = s.storage.GetGenre(cntx, primitives)
		if err != nil {
			return nil, err
		}
		for _, g := range genre {
			data, _ := json.Marshal(g)
			s.cache.Set([]byte(g.Id.Hex()), data, int((time.Hour*6)/time.Second))
		}
		s.logger.Infof("added %d items in cache", len(genre))
	} else {
		s.logger.Infof("got %d items from cache", len(genre))
	}

	return genre, nil
}
func (s *service) GetAllGenres(ctx context.Context) ([]*Genre, error) {
	cntx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	genres, err := s.storage.GetAllGenres(cntx)
	if err != nil {
		return nil, err
	}

	return genres, nil
}
func (s *service) AddGenre(ctx context.Context, query *AddGenreQuery) (*Genre, error) {
	if err := s.validator.Struct(query); err != nil {
		return nil, errormiddleware.ValidationError(err, "wrong query")
	}
	cntx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	genre := &Genre{
		Name: query.Name,
	}
	response, err := s.storage.AddGenre(cntx, genre)
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(response)
	s.cache.Set([]byte(response.Id.Hex()), data, int((time.Hour*6)/time.Second))
	s.logger.Infof("created new genre: %v", response)
	return response, nil
}

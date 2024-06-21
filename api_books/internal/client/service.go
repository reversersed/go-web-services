package client

import (
	"context"

	"github.com/reversersed/go-web-services/tree/main/api_books/pkg/cache"
	"github.com/reversersed/go-web-services/tree/main/api_books/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_books/pkg/logging"
	valid "github.com/reversersed/go-web-services/tree/main/api_books/pkg/validator"
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
func (s *service) AddBook(ctx context.Context, query *InsertBookQuery) (*Book, error) {

	if err := s.validator.Struct(&query); err != nil {
		return nil, errormiddleware.ValidationError(err, "wrong query")
	}
	book := &Book{
		Name:     query.Name,
		AuthorId: query.AuthorId,
		GenresId: query.GenresId,
		Year:     query.Year,
	}
	id, err := s.storage.AddBook(ctx, book)
	if err != nil {
		return nil, err
	}

	book.Id, err = primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return book, nil
}

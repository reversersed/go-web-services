package client

import (
	"context"
	"encoding/json"
	"time"

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
func (s *service) IsBookExists(ctx context.Context, name string) bool {
	cntx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	book, err := s.storage.GetBookByName(cntx, name)
	s.logger.Infof("find book: %v with err %v", book, err)
	return (err == nil) && (book != nil)
}
func (s *service) AddBook(ctx context.Context, query *InsertBookQuery) (*Book, error) {

	if err := s.validator.Struct(query); err != nil {
		return nil, errormiddleware.ValidationError(err, "wrong query")
	}
	book := &Book{
		Name:      query.Name,
		AuthorId:  query.AuthorId,
		GenresId:  query.GenresId,
		Pages:     query.Pages,
		Year:      query.Year,
		FilePath:  query.FilePath,
		CoverPath: query.CoverPath,
	}
	id, err := s.storage.AddBook(ctx, book)
	if err != nil {
		return nil, err
	}

	book.Id, err = primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	data, _ := json.Marshal(book)
	s.cache.Set([]byte(book.Id.Hex()), data, int((time.Hour*6)/time.Second))
	s.logger.Infof("created new book: %v", book)
	return book, nil
}

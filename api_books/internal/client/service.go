package client

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/reversersed/go-web-services/tree/main/api_books/internal/config"
	"github.com/reversersed/go-web-services/tree/main/api_books/pkg/cache"
	"github.com/reversersed/go-web-services/tree/main/api_books/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_books/pkg/logging"
	"github.com/reversersed/go-web-services/tree/main/api_books/pkg/rest"
	valid "github.com/reversersed/go-web-services/tree/main/api_books/pkg/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type service struct {
	storage   Storage
	logger    *logging.Logger
	cache     cache.Cache
	validator *valid.Validator
	genreApi  *BaseClient
	authorApi *BaseClient
}

func NewService(storage Storage, logger *logging.Logger, cache cache.Cache, validator *valid.Validator, cfg *config.UrlConfig) *service {
	return &service{storage: storage, logger: logger, cache: cache, validator: validator,
		genreApi: &BaseClient{
			Base: &rest.RestClient{BaseURL: cfg.GenreApiAdress, HttpClient: &http.Client{Timeout: 5 * time.Second}, Logger: logger},
			Path: "/genres",
		},
		authorApi: &BaseClient{
			Base: &rest.RestClient{BaseURL: cfg.AuthorApiAdress, HttpClient: &http.Client{Timeout: 5 * time.Second}, Logger: logger},
			Path: "/authors",
		}}
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
func (s *service) FindBooks(ctx context.Context, filters map[string]string, offset, limit int) ([]*Book, error) {
	cntx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	books, err := s.storage.GetByFilter(cntx, filters, offset, limit)
	if err != nil {
		return nil, err
	}
	wg := sync.WaitGroup{}
	wg.Add(len(books) * 2)
	missed_genre := false
	for _, v := range books {
		go func(book *Book, wg *sync.WaitGroup) {
			defer wg.Done()
			paramRequest := make([]string, 0)
			var genres []*Genre
			for _, v := range book.GenresId {
				genre, err := s.cache.Get([]byte(v.Hex()))
				if err != nil {
					missed_genre = true
				} else {
					var genreObject Genre
					json.Unmarshal(genre, &genreObject)
					genres = append(genres, &genreObject)
				}
				paramRequest = append(paramRequest, v.Hex())
			}
			if missed_genre {
				genreBytes, err := s.genreApi.SendGetGeneric(ctx, "", map[string][]string{"id": {strings.Join(paramRequest, ",")}})
				if err != nil {
					s.logger.Errorf("error occured while fetching genres for book %v: %v", book, err)
					book.Genres = nil
					return
				}
				genres = make([]*Genre, 0)
				json.Unmarshal(genreBytes, &genres)
				for _, v := range genres {
					_, ok := s.cache.Get([]byte(v.Id.Hex()))
					if ok != nil {
						bytes, _ := json.Marshal(v)
						s.cache.Set([]byte(v.Id.Hex()), bytes, int(12*time.Hour))
					}
				}
			}

			book.Genres = genres
		}(v, &wg)

		go func(book *Book, wg *sync.WaitGroup) {
			defer wg.Done()
			var author Author
			bytes, err := s.cache.Get([]byte(book.AuthorId.Hex()))
			if err == nil {
				json.Unmarshal(bytes, &author)
				book.Author = &author
				return
			}

			authorBytes, err := s.authorApi.SendGetGeneric(ctx, "", map[string][]string{"id": {book.AuthorId.Hex()}})
			if err != nil {
				s.logger.Errorf("error occured while fetching author for book %v: %v", book, err)
				book.Author = nil
				return
			}
			json.Unmarshal(authorBytes, &author)
			if _, ok := s.cache.Get([]byte(v.Id.Hex())); ok != nil {
				bytes, _ := json.Marshal(v)
				s.cache.Set([]byte(v.Id.Hex()), bytes, int(12*time.Hour))
			}

			book.Author = &author
		}(v, &wg)
	}
	wg.Wait()
	return books, nil
}

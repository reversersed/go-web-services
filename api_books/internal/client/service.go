package client

import (
	"context"
	"encoding/json"
	"fmt"
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
func (s *service) GetBook(ctx context.Context, id string) (*Book, error) {
	cntx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if byteBook, err := s.cache.Get([]byte(fmt.Sprintf("book_%s", id))); err == nil {
		var unBook Book
		json.Unmarshal(byteBook, &unBook)
		return &unBook, nil
	}

	pId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errormiddleware.BadRequestError([]string{"bad request"}, err.Error())
	}

	book, err := s.storage.GetBookById(cntx, pId)
	if err != nil {
		return nil, err
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	go s.findBookGenres(cntx, book, &wg)
	go s.findBookAuthor(cntx, book, &wg)

	wg.Wait()
	if _, err := s.cache.Get([]byte(fmt.Sprintf("book_%s", book.Id.Hex()))); err != nil {
		bytes, _ := json.Marshal(book)
		s.cache.Set([]byte(fmt.Sprintf("book_%s", book.Id.Hex())), bytes, int(24*time.Hour))
	}
	return book, nil
}
func (s *service) IsBookExists(ctx context.Context, name string) bool {
	cntx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	book, err := s.storage.GetBookByName(cntx, name)
	s.logger.Infof("find book: %v with err %v", book, err)
	return (err == nil) && (book != nil)
}
func (s *service) AddBook(ctx context.Context, query *InsertBookQuery) (*Book, error) {
	cntx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

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
	id, err := s.storage.AddBook(cntx, book)
	if err != nil {
		return nil, err
	}

	book.Id, err = primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	go s.findBookGenres(cntx, book, &wg)
	go s.findBookAuthor(cntx, book, &wg)

	wg.Wait()

	data, _ := json.Marshal(book)
	s.cache.Set([]byte(fmt.Sprintf("book_%s", book.Id.Hex())), data, int((time.Hour*6)/time.Second))
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
	for _, v := range books {
		go s.findBookGenres(cntx, v, &wg)
		go s.findBookAuthor(cntx, v, &wg)
	}
	wg.Wait()
	for _, v := range books {
		if _, err := s.cache.Get([]byte(fmt.Sprintf("book_%s", v.Id.Hex()))); err != nil {
			bytes, _ := json.Marshal(v)
			s.cache.Set([]byte(fmt.Sprintf("book_%s", v.Id.Hex())), bytes, int(24*time.Hour))
		}
	}
	return books, nil
}
func (s *service) findBookGenres(ctx context.Context, book *Book, wg *sync.WaitGroup) {
	defer wg.Done()
	missed_genre := false
	paramRequest := make([]string, 0)
	var genres []*Genre
	for _, v := range book.GenresId {
		genre, err := s.cache.Get([]byte(fmt.Sprintf("genre_%s", v.Hex())))
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
			if _, ok := s.cache.Get([]byte(fmt.Sprintf("genre_%s", v.Id.Hex()))); ok != nil {
				bytes, _ := json.Marshal(v)
				s.cache.Set([]byte(fmt.Sprintf("genre_%s", v.Id.Hex())), bytes, int(12*time.Hour))
			}
		}
	}

	book.Genres = genres
}
func (s *service) findBookAuthor(ctx context.Context, book *Book, wg *sync.WaitGroup) {
	defer wg.Done()
	var author Author
	bytes, err := s.cache.Get([]byte(fmt.Sprintf("author_%s", book.AuthorId.Hex())))
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
	if _, ok := s.cache.Get([]byte(fmt.Sprintf("author_%s", author.Id.Hex()))); ok != nil {
		s.cache.Set([]byte(fmt.Sprintf("author_%s", author.Id.Hex())), authorBytes, int(12*time.Hour))
	}

	book.Author = &author
}

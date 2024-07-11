package book

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/julienschmidt/httprouter"
	model "github.com/reversersed/go-web-services/tree/main/api_gateway/internal/client/book"
	mw "github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
	valid "github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/validator"
)

const (
	url_add_book = "/api/v1/books"
	url_get_book = "/api/v1/books"
)

//go:generate mockgen -source=handler.go -destination=mocks/service_mock.go

type BookService interface {
	AddBook(ctx context.Context, body io.Reader, contentType string) (*model.Book, error)
	FindBooks(ctx context.Context, params url.Values) ([]*model.Book, error)
}
type JwtService interface {
	Middleware(h http.HandlerFunc, roles ...string) http.HandlerFunc
}
type Handler struct {
	Logger      *logging.Logger
	BookService BookService
	JwtService  JwtService
	Validator   *valid.Validator
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, url_add_book, h.JwtService.Middleware(h.Logger.Middleware(mw.Middleware(h.AddBook)), "admin"))
	router.HandlerFunc(http.MethodGet, url_get_book, h.Logger.Middleware(mw.Middleware(h.FindBooks)))
	h.Logger.Info("book handlers registered")
}

// @Summary Creates a new book
// @Description Requires admin role to use
// @Tags books
// @Produce json
// @Param Book formData model.InsertBookQuery true "Book's name must be unique"
// @Success 201 {object} model.Book "Successful response. Added book"
// @Failure 400 {object} errormiddleware.Error "Return's if handler received wrong content-type"
// @Failure 401 {object} errormiddleware.Error "User is not authorized"
// @Failure 403 {object} errormiddleware.Error "Returns when user has no rights to use this handler"
// @Failure 500 {object} errormiddleware.Error "Returns when there's some internal error that needs to be fixed or smtp server is not responding"
// @Failure 501 {object} errormiddleware.Error "Returns if query was incorrect"
// @Security ApiKeyAuth
// @Router /books [post]
func (h *Handler) AddBook(w http.ResponseWriter, r *http.Request) error {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		return mw.BadRequestError([]string{"wrong request"}, fmt.Sprintf("Content Type %s not validated. Must be form-data.", r.Header.Get("Content-Type")))
	}
	book, err := h.BookService.AddBook(r.Context(), r.Body, r.Header.Get("Content-Type"))

	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusCreated)
	bookByte, _ := json.Marshal(book)
	w.Write(bookByte)
	return nil
}

// @Summary Finds a books by filters
// @Description Author and genres are fetching from another microservices and then storing in cache
// @Description If it's impossible to fetch author or genres, the field will be null
// @Tags books
// @Produce json
// @Param offset query string true "Offset to books. Must be present, starting with 0"
// @Param limit query string true "Max amount of docs to return. Must be greater than 0"
// @Success 200 {array} model.Book "Successful response"
// @Failure 400 {object} errormiddleware.Error "Returns if query was incorrect"
// @Failure 404 {object} errormiddleware.Error "Returns if there are no documents found"
// @Failure 500 {object} errormiddleware.Error "Returns when there's some internal error that needs to be fixed or smtp server is not responding"
// @Router /books [get]
func (h *Handler) FindBooks(w http.ResponseWriter, r *http.Request) error {
	books, err := h.BookService.FindBooks(r.Context(), r.URL.Query())
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	bookByte, _ := json.Marshal(books)
	w.Write(bookByte)
	return nil
}

package book

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	model "github.com/reversersed/go-web-services/tree/main/api_gateway/internal/client/book"
	mw "github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
	valid "github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/validator"
)

const (
	url_add_book = "/api/v1/books"
)

//go:generate mockgen -source=handler.go -destination=mocks/service_mock.go

type BookService interface {
	AddBook(ctx context.Context, body io.Reader, contentType string) (*model.Book, error)
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
	h.Logger.Info("book service registered")
}

// @Summary Creates a new book
// @Tags books
// @Produce json
// @Param Book formData model.InsertBookQuery true "Book's name must be unique"
// @Success 201 {object} model.Book "Successful response. Added book"
// @Failure 400 {object} errormiddleware.Error "Return's if some fields was missing"
// @Failure 401 {object} errormiddleware.Error "Return's if service can't authorize user or user's rights"
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

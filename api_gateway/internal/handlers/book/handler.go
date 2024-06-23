package book

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	mw "github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
	valid "github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/validator"
)

const (
	url_add_book = "/api/v1/books"
)

//go:generate mockgen -source=handler.go -destination=mocks/service_mock.go

type BookService interface {
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
func (h *Handler) AddBook(w http.ResponseWriter, r *http.Request) error {

	return nil
}

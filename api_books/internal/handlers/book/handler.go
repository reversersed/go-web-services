package book

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/reversersed/go-web-services/tree/main/api_books/internal/client"
	"github.com/reversersed/go-web-services/tree/main/api_books/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_books/pkg/logging"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//go:generate mockgen -source=handler.go -destination=mocks/service_mock.go

const (
	url_add_book = "/books"
)

type Service interface {
	AddBook(ctx context.Context, query *client.InsertBookQuery) (*client.Book, error)
}
type Handler struct {
	Logger      *logging.Logger
	BookService Service
}

func (h *Handler) Register(route *httprouter.Router) {
	route.HandlerFunc(http.MethodPost, url_add_book, h.Logger.Middleware(errormiddleware.Middleware(h.AddBookHandler)))
}

func (h *Handler) AddBookHandler(w http.ResponseWriter, r *http.Request) error {
	r.ParseMultipartForm(10 << 20) //10 Mb
	file, handler, err := r.FormFile("file")
	if err != nil {
		return errormiddleware.BadRequestError([]string{"you need to upload file"}, err.Error())
	}

	author, err := primitive.ObjectIDFromHex(r.FormValue("authorid"))
	if err != nil {
		return errormiddleware.BadRequestError([]string{"invalid author id"}, err.Error())
	}

	query := &client.InsertBookQuery{
		Name:     r.FormValue("name"),
		AuthorId: author,
	}
}

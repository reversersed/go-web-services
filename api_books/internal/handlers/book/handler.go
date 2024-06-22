package book

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

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
	file, header, err := r.FormFile("file")
	if err != nil {
		return errormiddleware.BadRequestError([]string{"you need to upload file"}, err.Error())
	}

	author, err := primitive.ObjectIDFromHex(r.FormValue("authorid"))
	if err != nil {
		return errormiddleware.BadRequestError([]string{"invalid author id"}, err.Error())
	}
	genres := make([]primitive.ObjectID, len(strings.Split(r.FormValue("genres"), ",")))
	wg := sync.WaitGroup{}
	errchan := make(chan error, 1)

	for idx, val := range strings.Split(r.FormValue("genres"), ",") {
		wg.Add(1)
		go func(idx int, value string, wg *sync.WaitGroup) {
			genres[idx], err = primitive.ObjectIDFromHex(value)
			if err != nil {
				errchan <- err
			}
			wg.Done()
		}(idx, val, &wg)
	}
	wg.Wait()
	close(errchan)
	select {
	case err = <-errchan:
		return errormiddleware.BadRequestError([]string{"invalid genre id"}, err.Error())
	default:
		break
	}
	year, err := strconv.Atoi(r.FormValue("year"))
	if err != nil {
		return errormiddleware.BadRequestError([]string{"year must be a number"}, err.Error())
	}
	query := &client.InsertBookQuery{
		Name:     r.FormValue("name"),
		AuthorId: author,
		GenresId: genres,
		Year:     year,
		File:     file,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	book, err := h.BookService.AddBook(ctx, query)
	if err != nil {
		return err
	}
	//save file and write path

	w.WriteHeader(http.StatusCreated)
	bytes, _ := json.Marshal(&book)
	w.Write(bytes)
}

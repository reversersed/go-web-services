package book

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
	url_add_book        = "/books"
	url_get_books       = "/books"
	url_find_book_by_id = "/books/:id"
)

type Service interface {
	IsBookExists(ctx context.Context, name string) bool
	AddBook(ctx context.Context, query *client.InsertBookQuery) (*client.Book, error)
	FindBooks(ctx context.Context, filters map[string]string, offset, limit int) ([]*client.Book, error)
	GetBook(ctx context.Context, id string) (*client.Book, error)
}
type Handler struct {
	Logger      *logging.Logger
	BookService Service
}

func (h *Handler) Register(route *httprouter.Router) {
	route.HandlerFunc(http.MethodPost, url_add_book, h.Logger.Middleware(errormiddleware.Middleware(h.AddBookHandler)))
	route.HandlerFunc(http.MethodGet, url_get_books, h.Logger.Middleware(errormiddleware.Middleware(h.GetBooks)))
	route.HandlerFunc(http.MethodGet, url_find_book_by_id, h.Logger.Middleware(errormiddleware.Middleware(h.FindBook)))
}
func (h *Handler) FindBook(w http.ResponseWriter, r *http.Request) error {
	params := httprouter.ParamsFromContext(r.Context())

	id := params.ByName("id")
	if len(id) == 0 {
		return errormiddleware.BadRequestError([]string{"id: parameter is required"}, "id path is not present")
	}
	book, err := h.BookService.GetBook(r.Context(), id)
	if err != nil {
		return err
	}
	bytes, _ := json.Marshal(book)
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
	return nil
}
func (h *Handler) GetBooks(w http.ResponseWriter, r *http.Request) error {
	filters := make(map[string]string, 0)
	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		return errormiddleware.BadRequestError([]string{"bad query request", "offset must be present"}, err.Error())
	}
	if offset < 0 {
		return errormiddleware.BadRequestError([]string{"bad query request"}, "offset must be greater than -1")
	}
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		return errormiddleware.BadRequestError([]string{"bad query request", "offset must be present"}, err.Error())
	}
	if limit <= 0 {
		return errormiddleware.BadRequestError([]string{"bad query request"}, "limit must be greater than 0")
	}
	AllowedParams := []string{}

	for _, v := range AllowedParams {
		if r.URL.Query().Has(v) {
			filters[v] = r.URL.Query().Get(v)
		}
	}
	books, err := h.BookService.FindBooks(r.Context(), filters, offset, limit)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	bytes, _ := json.Marshal(books)
	w.Write(bytes)
	return nil
}
func (h *Handler) AddBookHandler(w http.ResponseWriter, r *http.Request) error {
	r.ParseMultipartForm(10 << 20) //10 Mb
	if h.BookService.IsBookExists(r.Context(), r.FormValue("name")) {
		return errormiddleware.NotUniqueError([]string{fmt.Sprintf("name %s already taken", r.FormValue("name"))}, "book with provided name already in database")
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return errormiddleware.BadRequestError([]string{"file: you need to upload file"}, err.Error())
	}
	defer file.Close()
	cover, coverHeader, err := r.FormFile("cover")
	if err != nil {
		return errormiddleware.BadRequestError([]string{"cover: you need to upload file"}, err.Error())
	}
	defer cover.Close()

	if filepath.Ext(header.Filename) != ".pdf" {
		return errormiddleware.BadRequestError([]string{"file must have a .pdf extension"}, "wrong file extension")
	}
	switch filepath.Ext(coverHeader.Filename) {
	case ".jpg", ".png", ".jpeg":
		break
	default:
		return errormiddleware.BadRequestError([]string{"cover has a wrong extension", "available extensions: .jpg, .png, .jpeg"}, "wrong file extension")
	}

	author, err := primitive.ObjectIDFromHex(r.FormValue("authorid"))
	if err != nil {
		return errormiddleware.BadRequestError([]string{"authorid: invalid author id"}, err.Error())
	}
	genres := make([]primitive.ObjectID, len(strings.Split(r.FormValue("genres"), ",")))
	wg := sync.WaitGroup{}
	errchan := make(chan error, 1)
	defer close(errchan)

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
	pages, err := strconv.Atoi(r.FormValue("pages"))
	if err != nil {
		return errormiddleware.BadRequestError([]string{"pages must be a number"}, err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	file_path := fmt.Sprintf("./files/books/%s/book_%s%s", r.FormValue("name"), primitive.NewObjectID().Hex(), filepath.Ext(header.Filename))
	cover_path := fmt.Sprintf("./files/books/%s/cover_%s%s", r.FormValue("name"), primitive.NewObjectID().Hex(), filepath.Ext(coverHeader.Filename))

	query := &client.InsertBookQuery{
		Name:      r.FormValue("name"),
		AuthorId:  author,
		GenresId:  genres,
		Year:      year,
		FilePath:  filepath.Base(file_path),
		CoverPath: filepath.Base(cover_path),
		Pages:     pages,
	}
	book, err := h.BookService.AddBook(ctx, query)
	if err != nil {
		return err
	}
	err = os.MkdirAll("./files/books", 0644)
	if err != nil {
		return err
	}
	err = os.MkdirAll(fmt.Sprintf("./files/books/%s", query.Name), 0644)
	if err != nil {
		return err
	}
	destination, err := os.Create(file_path)
	if err != nil {
		return err
	}
	defer destination.Close()
	if _, err := io.Copy(destination, file); err != nil {
		return err
	}

	coverDest, err := os.Create(cover_path)
	if err != nil {
		return err
	}
	defer coverDest.Close()
	if _, err := io.Copy(coverDest, cover); err != nil {
		return err
	}

	w.WriteHeader(http.StatusCreated)
	bytes, _ := json.Marshal(&book)
	w.Write(bytes)
	return nil
}

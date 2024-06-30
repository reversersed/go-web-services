package genre

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/reversersed/go-web-services/tree/main/api_genres/internal/client"
	"github.com/reversersed/go-web-services/tree/main/api_genres/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_genres/pkg/logging"
)

//go:generate mockgen -source=handler.go -destination=mocks/service_mock.go

const (
	url_genres    = "/genres"
	url_all_genre = "/genres/all"
)

type Service interface {
	GetGenre(ctx context.Context, id string) ([]*client.Genre, error)
	AddGenre(ctx context.Context, genre *client.AddGenreQuery) (*client.Genre, error)
	GetAllGenres(ctx context.Context) ([]*client.Genre, error)
}
type Handler struct {
	Logger  *logging.Logger
	Service Service
}

func (h *Handler) Register(route *httprouter.Router) {
	route.HandlerFunc(http.MethodPost, url_genres, h.Logger.Middleware(errormiddleware.Middleware(h.AddGenre)))
	route.HandlerFunc(http.MethodGet, url_genres, h.Logger.Middleware(errormiddleware.Middleware(h.GetGenre)))
	route.HandlerFunc(http.MethodGet, url_all_genre, h.Logger.Middleware(errormiddleware.Middleware(h.GetAll)))
}

func (h *Handler) AddGenre(w http.ResponseWriter, r *http.Request) error {
	var query client.AddGenreQuery
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	genre, err := h.Service.AddGenre(ctx, &query)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusCreated)
	body, _ := json.Marshal(genre)
	w.Write(body)
	return nil
}
func (h *Handler) GetGenre(w http.ResponseWriter, r *http.Request) error {
	code := r.URL.Query().Get("id")
	if len(code) == 0 {
		return errormiddleware.BadRequestError([]string{"code param must contain at least 1 id"}, "wrong query (try use ?code=id1,id2,id3...)")
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	genres, err := h.Service.GetGenre(ctx, code)
	if err != nil {
		return err
	}

	body, _ := json.Marshal(&genres)
	w.WriteHeader(http.StatusOK)
	w.Write(body)
	return nil
}
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	genres, err := h.Service.GetAllGenres(ctx)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	body, _ := json.Marshal(&genres)
	w.Write(body)
	return nil
}

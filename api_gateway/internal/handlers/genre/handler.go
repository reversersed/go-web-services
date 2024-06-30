package genre

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	_ "github.com/reversersed/go-web-services/tree/main/api_gateway/internal/client/genre"
	mw "github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
	valid "github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/validator"
)

const (
	url_genres     = "/api/v1/genres"
	url_all_genres = "/api/v1/genres/all"
)

//go:generate mockgen -source=handler.go -destination=mocks/service_mock.go

type Service interface {
	SendPostGeneric(ctx context.Context, path string, body []byte) ([]byte, error)
	SendGetGeneric(ctx context.Context, path string, params map[string][]string) ([]byte, error)
}
type JwtService interface {
	Middleware(h http.HandlerFunc, roles ...string) http.HandlerFunc
}
type Handler struct {
	Logger       *logging.Logger
	JwtService   JwtService
	GenreService Service
	Validator    *valid.Validator
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, url_genres, h.JwtService.Middleware(h.Logger.Middleware(mw.Middleware(h.AddGenre)), "admin"))
	router.HandlerFunc(http.MethodGet, url_genres, h.Logger.Middleware(mw.Middleware(h.GetGenre)))
	router.HandlerFunc(http.MethodGet, url_all_genres, h.Logger.Middleware(mw.Middleware(h.GetAllGenre)))
	h.Logger.Info("genre handlers registered")
}

// @Summary Adds a genre
// @Description Requires admin role to use
// @Tags genres
// @Produce json
// @Param Genre body genre.AddGenreQuery true "Genre name"
// @Success 201 {object} genre.Genre "Successful response. Added genre"
// @Failure 400 {object} errormiddleware.Error "Return's if request body was empty"
// @Failure 401 {object} errormiddleware.Error "User is not authorized"
// @Failure 403 {object} errormiddleware.Error "Returns when user has no rights to use this handler"
// @Failure 500 {object} errormiddleware.Error "Returns when there's some internal error that needs to be fixed or smtp server is not responding"
// @Failure 501 {object} errormiddleware.Error "Returns if query was incorrect"
// @Security ApiKeyAuth
// @Router /genres [post]
func (h *Handler) AddGenre(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return mw.BadRequestError([]string{"can't read request body"}, err.Error())
	}

	response, err := h.GenreService.SendPostGeneric(ctx, "", body)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(response)
	return nil
}

// @Summary Get genres by id
// @Description You can use multiple ids in query using , separator
// @Description Example: ?id=id1,id2,id3...
// @Tags genres
// @Produce json
// @Param id query string true "Genre IDs"
// @Success 200 {array} genre.Genre "Successful response"
// @Failure 400 {object} errormiddleware.Error "Return's if received bad request"
// @Failure 404 {object} errormiddleware.Error "Return's if genre was not found"
// @Failure 500 {object} errormiddleware.Error "Returns when there's some internal error that needs to be fixed or smtp server is not responding"
// @Router /genres [get]
func (h *Handler) GetGenre(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	code := r.URL.Query().Get("id")
	if len(code) == 0 {
		return mw.BadRequestError([]string{"wrong request received"}, "id param is required")
	}

	response, err := h.GenreService.SendGetGeneric(ctx, "", map[string][]string{
		"id": {code},
	})
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)
	return nil
}

// @Summary Get all genres stored in database
// @Tags genres
// @Produce json
// @Success 200 {array} genre.Genre "Successful response"
// @Failure 404 {object} errormiddleware.Error "Return's if service does not have data"
// @Failure 500 {object} errormiddleware.Error "Returns when there's some internal error that needs to be fixed or smtp server is not responding"
// @Router /genres/all [get]
func (h *Handler) GetAllGenre(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	response, err := h.GenreService.SendGetGeneric(ctx, "/all", nil)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)
	return nil
}

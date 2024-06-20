package book

import (
	"github.com/julienschmidt/httprouter"
	"github.com/reversersed/go-web-services/tree/main/api_genres/pkg/logging"
)

//go:generate mockgen -source=handler.go -destination=mocks/service_mock.go

const (
// url_auth              = "/users/auth"
)

type Service interface {
}
type Handler struct {
	Logger      *logging.Logger
	BookService Service
}

func (h *Handler) Register(route *httprouter.Router) {
	//route.HandlerFunc(http.MethodPost, url_auth, h.Logger.Middleware(errormiddleware.Middleware(h.AuthUser)))
}

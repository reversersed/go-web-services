package notification

import (
	"github.com/julienschmidt/httprouter"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/logging"
)

const (
	url_auth = "/users/auth"
)

type Service interface {
}

type Handler struct {
	Logger  *logging.Logger
	Service Service
}

func (h *Handler) Register(route *httprouter.Router) {

}

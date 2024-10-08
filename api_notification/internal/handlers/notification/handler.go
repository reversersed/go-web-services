package notification

import (
	"github.com/julienschmidt/httprouter"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/logging"
	valid "github.com/reversersed/go-web-services/tree/main/api_notification/pkg/validator"
)

const ()

//go:generate mockgen -source=handler.go -destination=mocks/service.go

type Service interface {
}

type Handler struct {
	Logger    *logging.Logger
	Service   Service
	Validator *valid.Validator
}

func (h *Handler) Register(route *httprouter.Router) {

}

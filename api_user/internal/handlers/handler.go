package handlers

import "github.com/julienschmidt/httprouter"

type Handler interface {
	Register(route *httprouter.Router)
}

package errormiddleware

import (
	"errors"
	"net/http"
)

type Handler func(w http.ResponseWriter, r *http.Request) error

func Middleware(h Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var internal_err *Error
		err := h(w, r)
		if err != nil {
			if errors.As(err, &internal_err) {
				err := err.(*Error)
				w.WriteHeader(http.StatusBadRequest)
				w.Write(err.Marshall())
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(sysError(err.Error()).Marshall())
		}
	}
}

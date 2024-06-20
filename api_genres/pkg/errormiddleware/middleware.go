package errormiddleware

import (
	"errors"
	"net/http"
)

type Handler func(w http.ResponseWriter, r *http.Request) error

func Middleware(h Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		var internal_err *Error
		err := h(w, r)
		if err != nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			if errors.As(err, &internal_err) {
				err := err.(*Error)
				switch err.Code {
				case InternalErrorCode:
					w.WriteHeader(http.StatusInternalServerError)
				case NotFoundErrorCode:
					w.WriteHeader(http.StatusNotFound)
				case BadRequestErrorCode:
					w.WriteHeader(http.StatusBadRequest)
				case ValidationErrorCode:
					w.WriteHeader(http.StatusNotImplemented)
				case UnauthorizedErrorCode:
					w.WriteHeader(http.StatusUnauthorized)
				case NotUniqueErrorCode:
					w.WriteHeader(http.StatusConflict)
				case ForbiddenErrorCode:
					w.WriteHeader(http.StatusForbidden)
				default:
					w.WriteHeader(http.StatusBadRequest)
				}
				w.Write(err.Marshall())
				return err
			}
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(sysError([]string{err.Error()}).Marshall())
			return err
		}
		return nil
	}
}

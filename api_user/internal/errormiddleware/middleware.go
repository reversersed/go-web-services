package errormiddleware

import (
	"errors"
	"net/http"

	"github.com/reversersed/go-web-services/tree/main/api_user/pkg/logging"
)

type Handler func(w http.ResponseWriter, r *http.Request) error

func Middleware(h Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var internal_err *Error
		w.Header().Add("Access-Control-Allow-Origin", "*")
		err := h(w, r)
		logger := logging.GetLogger()
		if err != nil {
			if errors.As(err, &internal_err) {
				err := err.(*Error)
				switch err.Code {
				case "IE-0001":
					w.WriteHeader(http.StatusInternalServerError)
				case "IE-0002":
					w.WriteHeader(http.StatusNotFound)
				case "IE-0004":
					w.WriteHeader(http.StatusNotImplemented)
				case "IE-0005":
					w.WriteHeader(http.StatusUnauthorized)
				case "IE-0006":
					w.WriteHeader(http.StatusConflict)
				case "IE-0007":
					w.WriteHeader(http.StatusForbidden)
				default:
					w.WriteHeader(http.StatusBadRequest)
				}
				logger.Warnf("Error %s occured: %s (%s)", err.Code, err.Message, err.DeveloperMessage)
				w.Header().Add("Content-Type", "application/json")
				w.Write(err.Marshall())
				return
			}
			logger.Errorf("Undefined error occured: %v", err)
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(sysError([]string{err.Error()}).Marshall())
		}
	}
}

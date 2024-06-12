package logging

import (
	"errors"
	"net/http"
	"time"

	mw "github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/errormiddleware"
)

func (logger *Logger) Middleware(h func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Infof("received %s request %s", r.Method, r.RequestURI)
		start := time.Now()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		err := h(w, r)
		logger.Infof("handler fininshed working after %dms", time.Since(start).Milliseconds())

		if err != nil {
			var internal_err *mw.Error
			if errors.As(err, &internal_err) {
				err := err.(*mw.Error)
				logger.Warnf("Error %s occured: %s (%s)", err.Code, err.Message, err.DeveloperMessage)
				return
			}
			logger.Errorf("Undefined error occured: %v", err)
			return
		}
	}
}

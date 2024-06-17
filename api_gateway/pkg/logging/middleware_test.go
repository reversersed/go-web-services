package logging

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	errmw "github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/errormiddleware"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

var testCases = []struct {
	Name            string
	Err             error
	ExceptedLastLog string
}{
	{"Undefined error log message", errors.New("error happened"), "Undefined error occured: error happened"},
	{"Custom error log message", errmw.NotFoundError([]string{"message not found"}, "internal error"), fmt.Sprintf("Error %s occured: [message not found] (internal error)", errmw.NotFoundErrorCode)},
}

func TestMiddleware(t *testing.T) {
	log, hook := test.NewNullLogger()
	logger := &Logger{Entry: logrus.NewEntry(log)}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			handler := logger.Middleware(func(w http.ResponseWriter, r *http.Request) error {
				return testCase.Err
			})

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "http://test", nil)
			handler(w, r)

			entry := hook.LastEntry()
			assert.Equal(t, entry.Message, testCase.ExceptedLastLog)
		})
	}
}

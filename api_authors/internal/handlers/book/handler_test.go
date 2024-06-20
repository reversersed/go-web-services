package book

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/julienschmidt/httprouter"
	mock "github.com/reversersed/go-web-services/tree/main/api_authors/internal/handlers/book/mocks"
	"github.com/reversersed/go-web-services/tree/main/api_authors/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_authors/pkg/logging"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

var h *Handler

func TestMain(m *testing.M) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	h = &Handler{Logger: logger}

	os.Exit(m.Run())
}
func TestRegister(t *testing.T) {
	var registerCases = []struct {
		Name   string
		Path   string
		Method string
	}{}
	router := httprouter.New()
	h.Register(router)
	for _, registerCase := range registerCases {
		t.Run(registerCase.Name, func(t *testing.T) {
			handler, _, _ := router.Lookup(registerCase.Method, registerCase.Path)
			assert.NotNil(t, handler, "handler %s (%s) with method %s not found", registerCase.Name, registerCase.Path, registerCase.Method)
		})
	}
}

func TestHandlers(t *testing.T) {
	type handlerOptions struct {
		Name           string
		MockBehaviour  func(s *mock.MockService)
		InputJson      func() *[]byte
		ExceptedStatus int
		ExceptedError  error
		ExceptedBody   string
	}
	var testTable = []struct {
		HandlerName string
		Handler     func(w http.ResponseWriter, r *http.Request) error
		Method      string
		Options     []handlerOptions
	}{}
	for _, tt := range testTable {
		for _, testCase := range tt.Options {
			t.Run(fmt.Sprintf("%s %s", tt.HandlerName, testCase.Name), func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				service := mock.NewMockService(ctrl)
				if testCase.MockBehaviour != nil {
					testCase.MockBehaviour(service)
				}
				h.BookService = service

				w := httptest.NewRecorder()
				var r *http.Request
				if testCase.InputJson != nil && testCase.InputJson() != nil {
					r = httptest.NewRequest(tt.Method, "http://test", bytes.NewBuffer(*testCase.InputJson()))
				} else {
					r = httptest.NewRequest(tt.Method, "http://test", nil)
				}
				err := errormiddleware.Middleware(tt.Handler)(w, r)
				assert.Equal(t, testCase.ExceptedStatus, w.Result().StatusCode)
				assert.Equal(t, testCase.ExceptedError, err)

				body := w.Body.String()
				if assert.Len(t, body, len(testCase.ExceptedBody)) {
					assert.Equal(t, testCase.ExceptedBody, body)
				}
			})
		}
	}
}

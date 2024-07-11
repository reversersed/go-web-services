package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/reversersed/go-web-services/tree/main/api_books/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_books/pkg/logging"
	"github.com/reversersed/go-web-services/tree/main/api_books/pkg/rest"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

var intErr *errormiddleware.Error = errormiddleware.NotFoundError([]string{"not found"}, "can't find data")
var c BaseClient

type testResponse struct {
	Method string
	Query  string
	Path   string
	Body   []byte
}

func TestMain(m *testing.M) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var test testResponse
		test.Method = r.Method
		test.Query = r.URL.Query().Encode()
		test.Path = r.URL.Path
		defer r.Body.Close()
		body, _ := io.ReadAll(r.Body)
		test.Body = body

		if test.Path == "/test/error" {
			w.WriteHeader(http.StatusNotFound)
			w.Write(intErr.Marshall())
			return
		}
		testByte, _ := json.Marshal(test)
		w.WriteHeader(http.StatusOK)
		w.Write(testByte)
	}))
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}

	c = BaseClient{
		Path: "/test",
		Base: &rest.RestClient{
			BaseURL:    server.URL,
			HttpClient: &http.Client{},
			Logger:     logger,
		},
	}
	defer c.Base.Close()

	os.Exit(m.Run())
}

func TestPostGeneric(t *testing.T) {
	testTable := []struct {
		Name         string
		Body         string
		Path         string
		ExceptedPath string
		Error        error
	}{
		{
			Name:         "successful empty body sending",
			Body:         "",
			Path:         "",
			ExceptedPath: "/test",
			Error:        nil,
		},
		{
			Name:         "successful error response",
			Body:         "",
			Path:         "/error",
			ExceptedPath: "",
			Error:        intErr,
		},
	}

	for _, tt := range testTable {
		t.Run(tt.Name, func(t *testing.T) {
			body, err := c.SendPostGeneric(context.Background(), tt.Path, []byte(tt.Body))

			assert.Equal(t, tt.Error, err)
			var response testResponse
			json.Unmarshal(body, &response)

			assert.Equal(t, tt.Body, string(response.Body))
			assert.Equal(t, tt.ExceptedPath, response.Path)
		})
	}
}

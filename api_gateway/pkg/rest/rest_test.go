package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

// URL Builder Tests

var urlBilderCases = []struct {
	Name     string
	Url      string
	Path     string
	Filters  []FilterOptions
	Err      error
	Excepted string
}{
	{
		Name:     "Empty url test",
		Url:      "http://localhost:0000",
		Path:     "",
		Filters:  []FilterOptions{},
		Excepted: "http://localhost:0000",
	},
	{
		Name:     "Empty filter test",
		Url:      "http://localhost:0000",
		Path:     "/testing",
		Filters:  []FilterOptions{},
		Excepted: "http://localhost:0000/testing",
	},
	{
		Name:     "Empty filter test without slash",
		Url:      "http://localhost:0000",
		Path:     "testing",
		Filters:  []FilterOptions{},
		Excepted: "http://localhost:0000/testing",
	},
	{
		Name: "Single filter test",
		Url:  "http://localhost:0000",
		Path: "/testing",
		Filters: []FilterOptions{
			{
				Field:  "id",
				Values: []string{"test"},
			},
		},
		Excepted: "http://localhost:0000/testing?id=test",
	},
	{
		Name: "Single filter test with multiple values",
		Url:  "http://localhost:0000",
		Path: "/testing",
		Filters: []FilterOptions{
			{
				Field:  "id",
				Values: []string{"test", "second", "any"},
			},
		},
		Excepted: "http://localhost:0000/testing?id=test%2Csecond%2Cany",
	},
	{
		Name: "Multiple filter test with multiple values",
		Url:  "http://localhost:0000",
		Path: "/testing",
		Filters: []FilterOptions{
			{
				Field:  "id",
				Values: []string{"test", "second", "any"},
			},
			{
				Field:  "name",
				Values: []string{"Alice", "Gray"},
			},
		},
		Excepted: "http://localhost:0000/testing?id=test%2Csecond%2Cany&name=Alice%2CGray",
	},
	{
		Name:    "Wrong http url",
		Url:     "wrongurl",
		Path:    "testing",
		Filters: []FilterOptions{},
		Err:     errors.New("failed to parse url: parse \"wrongurl\": invalid URI for request"),
	},
}

func TestUrlBuilder(t *testing.T) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}

	for _, urlCase := range urlBilderCases {
		t.Run(urlCase.Name, func(t *testing.T) {
			client := &RestClient{
				BaseURL: urlCase.Url,
				Logger:  logger,
			}
			url, err := client.BuildURL(urlCase.Path, urlCase.Filters)
			assert.Equal(t, err, urlCase.Err)
			assert.Equal(t, url, urlCase.Excepted)
		})
	}
}

func TestClientClose(t *testing.T) {
	client := &RestClient{
		HttpClient: &http.Client{},
	}
	client.Close()

	assert.Nil(t, client.HttpClient)
}

// Send Request Test
var requestCases = []struct {
	Name     string
	Excepted string
	Code     int
	Method   string
	Err      error
	Body     io.Reader
}{
	{
		Name:     "successful response",
		Method:   http.MethodGet,
		Body:     nil,
		Excepted: "hello world",
		Code:     http.StatusOK,
	},
	{
		Name:     "successful response with request body",
		Method:   http.MethodPut,
		Body:     strings.NewReader("tester"),
		Excepted: "hello, tester",
		Code:     http.StatusCreated,
	},
	{
		Name:   "error returned",
		Method: http.MethodDelete,
		Body:   nil,
		Err: CustomError{
			Message:          []string{"hi"},
			ErrorCode:        "IE-1111",
			DeveloperMessage: "bad request",
		},
		Code: http.StatusBadRequest,
	},
}

func TestSendRequest(t *testing.T) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("hello world"))
		case http.MethodDelete:
			w.WriteHeader(http.StatusBadRequest)
			err := &CustomError{
				Message:          []string{"hi"},
				ErrorCode:        "IE-1111",
				DeveloperMessage: "bad request",
			}
			errBody, _ := json.Marshal(err)
			w.Write(errBody)
		case http.MethodPut:
			defer r.Body.Close()
			name, _ := io.ReadAll(r.Body)
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(fmt.Sprintf("hello, %s", string(name))))
		}
	}))

	client := &RestClient{
		BaseURL:    server.URL,
		Logger:     logger,
		HttpClient: &http.Client{Timeout: 5 * time.Second},
	}
	for _, requestCase := range requestCases {
		t.Run(requestCase.Name, func(t *testing.T) {

			req, err := http.NewRequest(requestCase.Method, server.URL, requestCase.Body)
			assert.NoError(t, err)

			response, err := client.SendRequest(req)
			assert.NoError(t, err)

			defer response.Body().Close()
			assert.Equal(t, response.StatusCode(), requestCase.Code)
			if !response.Valid {
				assert.Equal(t, response.Error, requestCase.Err)
			} else {
				assert.NoError(t, requestCase.Err)
				body, err := io.ReadAll(response.Body())
				assert.NoError(t, err)
				assert.Equal(t, string(body), requestCase.Excepted)
			}
		})
	}
}

func TestNilHttpClient(t *testing.T) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	client := &RestClient{
		BaseURL:    "",
		Logger:     logger,
		HttpClient: nil,
	}
	_, err := client.SendRequest(nil)
	assert.Error(t, err)
}

func TestEmptyRequest(t *testing.T) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	client := &RestClient{
		BaseURL:    "",
		Logger:     logger,
		HttpClient: &http.Client{},
	}
	request, _ := http.NewRequest("", "", nil)
	_, err := client.SendRequest(request)
	assert.Error(t, err)
}

func TestUserHeader(t *testing.T) {
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := json.Marshal(struct{ User string }{User: r.Header.Get("User")})
		w.Write(body)
	}))

	client := &RestClient{
		BaseURL:    server.URL,
		Logger:     logger,
		HttpClient: &http.Client{Timeout: 5 * time.Second},
	}
	ctx := context.WithValue(context.Background(), UserIdKey, "userKeyId")
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	assert.NoError(t, err)
	response, err := client.SendRequest(request)
	assert.NoError(t, err)

	type body struct{ User string }
	var Body body
	err = json.NewDecoder(response.Body()).Decode(&Body)
	assert.NoError(t, err)
	assert.Equal(t, Body.User, "userKeyId")
}

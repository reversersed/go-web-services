package rest

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

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
		Name:   "error returned",
		Method: http.MethodDelete,
		Body:   nil,
		Err:    errors.New("Error code: IE-1111, Error: hi, Dev message: bad request"),
		Code:   http.StatusBadRequest,
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
			if err != nil {
				t.Fatalf("excepted response but got error %v", err)
			}
			response, err := client.SendRequest(req)
			if err != nil {
				t.Fatalf("excepted response but got error %v", err)
			}
			defer response.Body().Close()
			if response.StatusCode() != requestCase.Code {
				t.Fatalf("excepted status code %d but got %d", requestCase.Code, response.StatusCode())
			}
			if !response.Valid {
				if requestCase.Err == nil {
					t.Fatalf("excepted response but got error %v", response.Error)
				}
				if response.Error.Error() != requestCase.Err.Error() {
					t.Fatalf("excepted error %v but got %v", requestCase.Err.Error(), response.Error.Error())
				}
			} else {
				if requestCase.Err != nil {
					t.Fatalf("excepted error %v but got response", requestCase.Err)
				}
				body, err := io.ReadAll(response.Body())
				if err != nil {
					t.Fatalf("excepted body but got error %v", err)
				}
				if string(body) != requestCase.Excepted {
					t.Fatalf("excepted body %s but got %s", requestCase.Excepted, string(body))
				}
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
	if err == nil {
		t.Fatal("excepted error but got nil")
	}
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
	if err == nil {
		t.Fatal("excepted error but got nil")
	}
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
	if err != nil {
		t.Fatalf("excepted request but got %v", err)
	}
	response, err := client.SendRequest(request)
	if err != nil {
		t.Fatalf("excepted response but got %v", err)
	}
	type body struct{ User string }
	var Body body
	err = json.NewDecoder(response.Body()).Decode(&Body)
	if err != nil {
		t.Fatalf("excepted body but got %v", err)
	}
	if Body.User != "userKeyId" {
		t.Fatalf("excepted response body userKeyId but got %s", Body.User)
	}
}

package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/reversersed/go-web-services/tree/main/api_books/pkg/logging"
)

type key string

const (
	UserIdKey key = "user_id"
)

type RestClient struct {
	BaseURL    string
	HttpClient *http.Client
	Logger     *logging.Logger
}

func (c *RestClient) SendRequest(r *http.Request) (*CustomResponse, error) {
	if c.HttpClient == nil {
		return nil, errors.New("no http client registered")
	}
	r.Header.Set("Accept", "*/*")
	if r.Header.Get("Content-Type") == "" {
		r.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	c.Logger.Infof("sending request to %s", r.URL)
	//reading userid from context and adding it to header
	userId, valid := r.Context().Value(UserIdKey).(string)
	if valid && len(userId) > 0 {
		r.Header.Add("User", userId)
	}

	response, err := c.HttpClient.Do(r)
	if err != nil {
		c.Logger.Errorf("error while sending rest request: %s", err)
		return nil, err
	}

	resp := CustomResponse{
		Valid:    true,
		response: response,
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusBadRequest {
		resp.Valid = false
		defer response.Body.Close()

		var errs CustomError
		if err = json.NewDecoder(response.Body).Decode(&errs); err == nil {
			resp.Error = errs
		}
	}
	return &resp, nil
}
func (c *RestClient) BuildURL(way string, filters map[string][]string) (string, error) {
	parsed, err := url.ParseRequestURI(c.BaseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse url: %v", err)
	}
	parsed.Path = path.Join(parsed.Path, way)

	if len(filters) > 0 {
		q := parsed.Query()
		for key, values := range filters {
			q.Set(key, strings.Join(values, ","))
		}
		parsed.RawQuery = q.Encode()
	}

	c.Logger.Infof("built url: %s", parsed.String())
	return parsed.String(), nil
}
func (c *RestClient) Close() error {
	c.HttpClient = nil
	return nil
}

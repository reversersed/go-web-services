package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
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
	r.Header.Set("Accept", "application/json; charset=utf-8")
	r.Header.Set("Content-Type", "application/json; charset=utf-8")

	c.Logger.Infof("sending requiest to %s", r.URL)
	response, err := c.HttpClient.Do(r)
	if err != nil {
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
func (c *RestClient) BuildURL(way string, filters []FilterOptions) (string, error) {
	parsed, err := url.ParseRequestURI(c.BaseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse url: %v", err)
	}
	parsed.Path = path.Join(parsed.Path, way)

	if len(filters) > 0 {
		q := parsed.Query()
		for _, opt := range filters {
			q.Set(opt.Field, opt.ToString())
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

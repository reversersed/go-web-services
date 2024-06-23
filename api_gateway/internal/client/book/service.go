package book

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/rest"
)

type client struct {
	base *rest.RestClient
	Path string
}

func NewService(baseURL, path string, logger *logging.Logger) *client {
	return &client{
		Path: path,
		base: &rest.RestClient{
			BaseURL: baseURL,
			HttpClient: &http.Client{
				Timeout: 10 * time.Second,
			},
			Logger: logger,
		},
	}
}

func (c *client) AddBook(ctx context.Context, body io.Reader, contentType string) (*Book, error) {
	uri, err := c.base.BuildURL(c.Path, nil)
	if err != nil {
		return nil, err
	}
	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, uri, body)
	if err != nil {
		return nil, fmt.Errorf("failed while request creation: %v", err)
	}
	req.Header.Add("Content-Type", contentType)
	response, err := c.base.SendRequest(req)
	if err != nil {
		return nil, err
	}
	if response.Valid {
		defer response.Body().Close()
		var b Book
		if err = json.NewDecoder(response.Body()).Decode(&b); err != nil {
			return nil, fmt.Errorf("failed to unmarshall response body: %v", err)
		}
		return &b, nil
	}
	return nil, errormiddleware.NewError(response.Error.Message, response.Error.ErrorCode, response.Error.DeveloperMessage)
}

package book

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	base "github.com/reversersed/go-web-services/tree/main/api_gateway/internal/client"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/rest"
)

type client struct {
	base.BaseClient
}

func NewService(baseURL, path string, logger *logging.Logger) *client {
	return &client{BaseClient: base.BaseClient{
		Path: path,
		Base: &rest.RestClient{
			BaseURL: baseURL,
			HttpClient: &http.Client{
				Timeout: 10 * time.Second,
			},
			Logger: logger,
		},
	}}
}
func (c *client) GetBook(ctx context.Context, id string) (*Book, error) {
	cntx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	bookByte, err := c.SendGetGeneric(cntx, id, nil)
	if err != nil {
		return nil, err
	}
	var book Book
	json.Unmarshal(bookByte, &book)

	return &book, nil
}
func (c *client) FindBooks(ctx context.Context, params url.Values) ([]*Book, error) {
	cntx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	AllowedParams := []string{"offset", "limit"}

	filters := make(map[string][]string, 0)
	for _, v := range AllowedParams {
		if params.Has(v) {
			filters[v] = []string{params.Get(v)}
		}
	}
	bookBytes, err := c.SendGetGeneric(cntx, "", filters)
	if err != nil {
		return nil, err
	}
	var books []*Book
	json.Unmarshal(bookBytes, &books)
	return books, nil
}
func (c *client) AddBook(ctx context.Context, body io.Reader, contentType string) (*Book, error) {
	uri, err := c.Base.BuildURL(c.Path, nil)
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
	response, err := c.Base.SendRequest(req)
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

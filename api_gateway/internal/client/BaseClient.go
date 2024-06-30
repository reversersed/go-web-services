package client

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/rest"
)

type BaseClient struct {
	Base *rest.RestClient
	Path string
}

func (c *BaseClient) SendPostGeneric(ctx context.Context, path string, body []byte) ([]byte, error) {
	uri, err := c.Base.BuildURL(c.Path+path, nil)
	if err != nil {
		return nil, err
	}
	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, uri, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed while request creation: %v", err)
	}
	response, err := c.Base.SendRequest(req)
	if err != nil {
		return nil, err
	}
	if response.Valid {
		defer response.Body().Close()
		return response.ReadBody()
	}
	return nil, errormiddleware.NewError(response.Error.Message, response.Error.ErrorCode, response.Error.DeveloperMessage)
}
func (c *BaseClient) SendGetGeneric(ctx context.Context, path string, params map[string][]string) ([]byte, error) {
	uri, err := c.Base.BuildURL(c.Path+path, params)
	if err != nil {
		return nil, err
	}
	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("failed while request creation: %v", err)
	}
	response, err := c.Base.SendRequest(req)
	if err != nil {
		return nil, err
	}
	if response.Valid {
		defer response.Body().Close()
		return response.ReadBody()
	}
	return nil, errormiddleware.NewError(response.Error.Message, response.Error.ErrorCode, response.Error.DeveloperMessage)
}

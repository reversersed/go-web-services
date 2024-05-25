package user

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fatih/structs"

	"github.com/reversersed/go-web-services/tree/main/api_gateway/internal/errormiddleware"
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
func (c *client) UserEmailConfirmation(ctx context.Context, code string) (int, error) {
	c.base.Logger.Info("building request url...")
	var uri string
	var err error
	if len(code) > 0 {
		filter := []rest.FilterOptions{
			{
				Field:  "code",
				Values: []string{code},
			},
		}
		uri, err = c.base.BuildURL(c.Path+"/email", filter)
	} else {
		uri, err = c.base.BuildURL(c.Path+"/email", nil)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to build url: %v", err)
	}
	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, uri, nil)
	if err != nil {
		return 0, fmt.Errorf("failed while request creation: %v", err)
	}
	response, err := c.base.SendRequest(req)
	if err != nil {
		return 0, err
	}
	if response.Valid {
		return response.StatusCode(), nil
	}
	return 0, errormiddleware.NewError(response.Error.Message, response.Error.ErrorCode, response.Error.DeveloperMessage)
}
func (c *client) AuthByLoginAndPassword(ctx context.Context, query *UserAuthQuery) (*User, error) {
	c.base.Logger.Info("building request url...")
	uri, err := c.base.BuildURL(c.Path+"/auth", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build url: %v", err)
	}
	structs.DefaultTagName = "json"
	body, err := json.Marshal(structs.Map(query))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed while request creation: %v", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	response, err := c.base.SendRequest(req.WithContext(reqCtx))
	if err != nil {
		return nil, err
	}
	if response.Valid {
		var u User
		if err = json.NewDecoder(response.Body()).Decode(&u); err != nil {
			return nil, fmt.Errorf("failed to unmarshall response body: %v", err)
		}
		return &u, nil
	}
	return nil, errormiddleware.NewError(response.Error.Message, response.Error.ErrorCode, response.Error.DeveloperMessage)
}
func (c *client) RegisterUser(ctx context.Context, query *UserRegisterQuery) (*User, error) {
	c.base.Logger.Info("building request url...")
	uri, err := c.base.BuildURL(c.Path+"/register", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build url: %v", err)
	}
	structs.DefaultTagName = "json"
	body, err := json.Marshal(structs.Map(query))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed while request creation: %v", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req = req.WithContext(reqCtx)
	response, err := c.base.SendRequest(req)
	if err != nil {
		return nil, err
	}
	if response.Valid {
		var u User
		if err = json.NewDecoder(response.Body()).Decode(&u); err != nil {
			return nil, fmt.Errorf("failed to unmarshall response body: %v", err)
		}
		return &u, nil
	}
	return nil, errormiddleware.NewError(response.Error.Message, response.Error.ErrorCode, response.Error.DeveloperMessage)
}

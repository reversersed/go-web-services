package user

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fatih/structs"

	Base "github.com/reversersed/go-web-services/tree/main/api_gateway/internal/client"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/rest"
)

type client struct {
	Base.BaseClient
}

func NewService(BaseURL, path string, logger *logging.Logger) *client {
	return &client{BaseClient: Base.BaseClient{
		Path: path,
		Base: &rest.RestClient{
			BaseURL: BaseURL,
			HttpClient: &http.Client{
				Timeout: 10 * time.Second,
			},
			Logger: logger,
		},
	}}
}
func (c *client) FindUser(ctx context.Context, userid string, login string) (*User, error) {
	filter := make(map[string][]string, 1)
	if len(userid) > 0 {
		filter["id"] = []string{userid}
	} else if len(login) > 0 {
		filter["login"] = []string{login}
	} else {
		return nil, errormiddleware.BadRequestError([]string{"query has to have one of parameters", "login: user login", "id: user id"}, "bad request provided")
	}

	uri, err := c.Base.BuildURL(c.Path, filter)
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
		var u User
		if err = json.NewDecoder(response.Body()).Decode(&u); err != nil {
			return nil, fmt.Errorf("failed to unmarshall response body: %v", err)
		}
		return &u, nil
	}
	return nil, errormiddleware.NewError(response.Error.Message, response.Error.ErrorCode, response.Error.DeveloperMessage)
}
func (c *client) UserEmailConfirmation(ctx context.Context, code string) (int, error) {
	c.Base.Logger.Info("building request url...")
	var uri string
	var err error
	if len(code) > 0 {
		filter := map[string][]string{
			"code": {code},
		}
		uri, err = c.Base.BuildURL(c.Path+"/email", filter)
	} else {
		uri, err = c.Base.BuildURL(c.Path+"/email", nil)
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
	response, err := c.Base.SendRequest(req)
	if err != nil {
		return 0, err
	}
	if response.Valid {
		return response.StatusCode(), nil
	}
	return 0, errormiddleware.NewError(response.Error.Message, response.Error.ErrorCode, response.Error.DeveloperMessage)
}
func (c *client) AuthByLoginAndPassword(ctx context.Context, query *UserAuthQuery) (*User, error) {
	c.Base.Logger.Info("building request url...")
	uri, err := c.Base.BuildURL(c.Path+"/auth", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build url: %v", err)
	}
	if query == nil {
		return nil, errormiddleware.BadRequestError([]string{"wrong login or password"}, "received nil query")
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

	response, err := c.Base.SendRequest(req.WithContext(reqCtx))
	if err != nil {
		return nil, err
	}
	if response.Valid {
		var u User
		defer response.Body().Close()
		if err = json.NewDecoder(response.Body()).Decode(&u); err != nil {
			return nil, fmt.Errorf("failed to unmarshall response body: %v", err)
		}
		return &u, nil
	}
	return nil, errormiddleware.NewError(response.Error.Message, response.Error.ErrorCode, response.Error.DeveloperMessage)
}
func (c *client) RegisterUser(ctx context.Context, query *UserRegisterQuery) (*User, error) {
	c.Base.Logger.Info("building request url...")
	uri, err := c.Base.BuildURL(c.Path+"/register", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build url: %v", err)
	}
	if query == nil {
		return nil, errormiddleware.BadRequestError([]string{"wrong login or password"}, "received nil query")
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
	response, err := c.Base.SendRequest(req)
	if err != nil {
		return nil, err
	}
	if response.Valid {
		var u User
		defer response.Body().Close()
		if err = json.NewDecoder(response.Body()).Decode(&u); err != nil {
			return nil, fmt.Errorf("failed to unmarshall response body: %v", err)
		}
		return &u, nil
	}
	return nil, errormiddleware.NewError(response.Error.Message, response.Error.ErrorCode, response.Error.DeveloperMessage)
}
func (c *client) DeleteUser(ctx context.Context, query *DeleteUserQuery) error {
	c.Base.Logger.Info("building request url...")
	uri, err := c.Base.BuildURL(c.Path+"/delete", nil)
	if err != nil {
		return fmt.Errorf("failed to build url: %v", err)
	}
	if query == nil {
		return errormiddleware.BadRequestError([]string{"wrong login or password"}, "received nil query")
	}
	structs.DefaultTagName = "json"
	body, err := json.Marshal(structs.Map(query))
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodDelete, uri, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed while request creation: %v", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req = req.WithContext(reqCtx)
	response, err := c.Base.SendRequest(req)
	if err != nil {
		return err
	}
	if response.Valid {
		return nil
	}
	return errormiddleware.NewError(response.Error.Message, response.Error.ErrorCode, response.Error.DeveloperMessage)
}
func (c *client) UpdateUserLogin(ctx context.Context, query *UpdateUserLoginQuery) (*User, error) {
	c.Base.Logger.Info("building request url...")
	uri, err := c.Base.BuildURL(c.Path+"/changename", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build url: %v", err)
	}
	if query == nil {
		return nil, errormiddleware.BadRequestError([]string{"wrong login or password"}, "received nil query")
	}
	structs.DefaultTagName = "json"
	body, err := json.Marshal(structs.Map(query))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPatch, uri, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed while request creation: %v", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req = req.WithContext(reqCtx)
	response, err := c.Base.SendRequest(req)
	if err != nil {
		return nil, err
	}
	if response.Valid {
		defer response.Body().Close()
		var u User
		if err = json.NewDecoder(response.Body()).Decode(&u); err != nil {
			return nil, fmt.Errorf("failed to unmarshall response body: %v", err)
		}
		return &u, nil
	}
	return nil, errormiddleware.NewError(response.Error.Message, response.Error.ErrorCode, response.Error.DeveloperMessage)
}

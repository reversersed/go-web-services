package client

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/reversersed/go-web-services/tree/main/api_notification/internal/config"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/cache"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/errormiddleware"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/logging"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/rest"
	valid "github.com/reversersed/go-web-services/tree/main/api_notification/pkg/validator"
)

type service struct {
	storage    Storage
	logger     *logging.Logger
	cache      cache.Cache
	validator  *valid.Validator
	restClient *rest.RestClient
}

func NewService(storage Storage, logger *logging.Logger, cache cache.Cache, validator *valid.Validator) *service {
	return &service{storage: storage, logger: logger, cache: cache, validator: validator, restClient: &rest.RestClient{
		BaseURL: config.GetConfig().Urls.Url_User_Service,
		HttpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		Logger: logger,
	},
	}
}

func (s *service) SendNotification(ctx context.Context, query *SendNotificationMessage) {
	cntx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := s.validator.Struct(query); err != nil {
		s.logger.Errorf("received wrong notification query: %v", errormiddleware.ValidationError(err.(validator.ValidationErrors), "").Error())
		return
	}

	exist, err := s.storage.IsUserExists(cntx, query.UserId)
	if err != nil {
		s.logger.Error(err)
	}
	if !exist {
		filter := []rest.FilterOptions{
			{
				Field:  "id",
				Values: []string{query.UserId},
			},
		}
		uri, err := s.restClient.BuildURL("/users", filter)
		if err != nil {
			s.logger.Error(err)
			return
		}
		request, err := http.NewRequestWithContext(cntx, http.MethodGet, uri, nil)
		if err != nil {
			s.logger.Error(err)
			return
		}
		response, err := s.restClient.SendRequest(request)
		if err != nil {
			s.logger.Error(err)
			return
		}
		if !response.Valid {
			s.logger.Error(response.Error)
			return
		}
		type UserResponse struct {
			Login string `json:"login"`
		}
		var u UserResponse
		defer response.Body().Close()
		if err = json.NewDecoder(response.Body()).Decode(&u); err != nil {
			s.logger.Error(err)
			return
		}
		err = s.storage.CreateUser(cntx, query.UserId, u.Login)
		if err != nil {
			s.logger.Errorf("Error while creating user: %v", err)
		}
	}
	err = s.storage.SendNotification(cntx, &Notification{Content: query.Content, Type: query.Type}, query.UserId)
	if err != nil {
		s.logger.Errorf("Error sending notification: %v", err)
		return
	}
	s.logger.Infof("Notification %s sended to user %s (Content: %s)", query.Type, query.UserId, query.Content)
}

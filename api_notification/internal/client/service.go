package client

import (
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/cache"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/logging"
)

type service struct {
	storage Storage
	logger  *logging.Logger
	cache   cache.Cache
}

func NewService(storage Storage, logger *logging.Logger, cache cache.Cache) *service {
	return &service{storage: storage, logger: logger, cache: cache}
}

/*func (s *service) SendNotification(ctx context.Context, userId string, query *SendNotificationQuery) {

}*/

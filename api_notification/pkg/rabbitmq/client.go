package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/reversersed/go-web-services/tree/main/api_notification/internal/config"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/logging"
)

type RabbitClient struct {
	*amqp.Connection
}

func New(config *config.RabbitConfig, logger *logging.Logger) *RabbitClient {
	connection, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", config.Rabbit_User, config.Rabbit_Pass, config.Rabbit_Host, config.Rabbit_Port))
	if err != nil {
		logger.Fatal(err)
	}
	return &RabbitClient{connection}
}

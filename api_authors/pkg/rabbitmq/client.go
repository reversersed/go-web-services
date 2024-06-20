package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/reversersed/go-web-services/tree/main/api_authors/internal/config"
	"github.com/reversersed/go-web-services/tree/main/api_authors/pkg/logging"
)

type RabbitClient struct {
	*amqp.Connection
}

func New(config *config.RabbitConfig, logger *logging.Logger) (*RabbitClient, error) {
	connection, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", config.Rabbit_User, config.Rabbit_Pass, config.Rabbit_Host, config.Rabbit_Port))
	if err != nil {
		return nil, err
	}
	return &RabbitClient{connection}, nil
}

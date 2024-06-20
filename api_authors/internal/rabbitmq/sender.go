package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/reversersed/go-web-services/tree/main/api_authors/pkg/logging"
)

type Sender struct {
	connection *amqp.Connection
	logger     *logging.Logger
}

func NewSender(connection *amqp.Connection, logger *logging.Logger) *Sender {
	return &Sender{connection: connection, logger: logger}
}
func (s *Sender) Close() error {
	return nil
}

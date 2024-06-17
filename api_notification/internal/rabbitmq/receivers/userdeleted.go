package receivers

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/reversersed/go-web-services/tree/main/api_notification/internal/rabbitmq"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/logging"
	valid "github.com/reversersed/go-web-services/tree/main/api_notification/pkg/validator"
)

//go:generate mockgen -source=userdeleted.go -destination=mocks/userdeleted.go

type user_deleted_service interface {
	OnUserDeleted(ctx context.Context, userid string)
}
type UserDeletedReceiver struct {
	connection *amqp.Connection
	validator  *valid.Validator
	logger     *logging.Logger
	channel    *amqp.Channel
	service    user_deleted_service
}

func NewUserDeletedReceiver(connection *amqp.Connection, validator *valid.Validator, logger *logging.Logger, service user_deleted_service) rabbitmq.Receiver {
	return &UserDeletedReceiver{
		connection: connection,
		validator:  validator,
		logger:     logger,
		service:    service,
	}
}
func (r *UserDeletedReceiver) Start() {
	ch, err := r.connection.Channel()
	if err != nil {
		r.logger.Fatal(err)
	}
	r.channel = ch

	queue, err := r.channel.QueueDeclare("UserDeletedQueue", false, false, false, false, nil)
	if err != nil {
		r.logger.Fatal(err)
	}
	err = ch.ExchangeDeclare("UserDeletedExchange", "fanout", false, false, false, false, nil)
	if err != nil {
		r.logger.Fatal(err)
	}

	err = r.channel.QueueBind(queue.Name, "#", "UserDeletedExchange", false, nil)
	if err != nil {
		r.logger.Fatal(err)
	}

	messages, err := r.channel.Consume(queue.Name, "NotificationAPI", true, false, false, false, nil)
	if err != nil {
		r.logger.Fatal(err)
	}
	go func() {
		for message := range messages {
			if r.channel.IsClosed() || r.connection.IsClosed() {
				return
			}
			r.logger.Info("Received user deleted message")
			r.service.OnUserDeleted(context.Background(), string(message.Body))
		}
	}()
	r.logger.Infof("Waiting for deleted users...")
}

func (r *UserDeletedReceiver) Close() error {
	return r.channel.Close()
}

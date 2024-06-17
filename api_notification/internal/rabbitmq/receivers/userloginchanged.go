package receivers

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/reversersed/go-web-services/tree/main/api_notification/internal/client"
	"github.com/reversersed/go-web-services/tree/main/api_notification/internal/rabbitmq"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/logging"
	valid "github.com/reversersed/go-web-services/tree/main/api_notification/pkg/validator"
)

//go:generate mockgen -source=userloginchanged.go -destination=mocks/userloginchanged.go

type user_login_changed_service interface {
	OnUserLoginChanged(ctx context.Context, query *client.UserLoginChangedMessage)
}
type UserLoginChangedReceiver struct {
	connection *amqp.Connection
	validator  *valid.Validator
	logger     *logging.Logger
	channel    *amqp.Channel
	service    user_login_changed_service
}

func NewUserLoginChangedReceiver(connection *amqp.Connection, validator *valid.Validator, logger *logging.Logger, service user_login_changed_service) rabbitmq.Receiver {
	return &UserLoginChangedReceiver{
		connection: connection,
		validator:  validator,
		logger:     logger,
		service:    service,
	}
}
func (r *UserLoginChangedReceiver) Start() {
	ch, err := r.connection.Channel()
	if err != nil {
		r.logger.Fatal(err)
	}
	r.channel = ch

	queue, err := r.channel.QueueDeclare("UserLoginChangedQueue", false, false, false, false, nil)
	if err != nil {
		r.logger.Fatal(err)
	}
	err = ch.ExchangeDeclare("UserLoginChangedExchange", "fanout", false, false, false, false, nil)
	if err != nil {
		r.logger.Fatal(err)
	}

	err = r.channel.QueueBind(queue.Name, "#", "UserLoginChangedExchange", false, nil)
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
			r.logger.Info("Received user login changed message")
			var msg client.UserLoginChangedMessage
			err := json.Unmarshal(message.Body, &msg)
			if err != nil {
				r.logger.Errorf("Unable to unmarshal message: %v", string(message.Body))
			} else {
				r.service.OnUserLoginChanged(context.Background(), &msg)
			}
		}
	}()
	r.logger.Infof("Waiting for users that changed login...")
}

func (r *UserLoginChangedReceiver) Close() error {
	return r.channel.Close()
}

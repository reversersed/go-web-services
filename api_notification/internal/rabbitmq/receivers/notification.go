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

//go:generate mockgen -source=notification.go -destination=mocks/notification.go

type notification_service interface {
	SendNotification(ctx context.Context, query *client.SendNotificationMessage)
}
type NotificationReceiver struct {
	connection *amqp.Connection
	validator  *valid.Validator
	logger     *logging.Logger
	service    notification_service
	channel    *amqp.Channel
}

func NewNotificationReceiver(connection *amqp.Connection, validator *valid.Validator, logger *logging.Logger, service notification_service) rabbitmq.Receiver {
	return &NotificationReceiver{
		connection: connection,
		validator:  validator,
		logger:     logger,
		service:    service,
	}
}
func (r *NotificationReceiver) Start() {
	ch, err := r.connection.Channel()
	if err != nil {
		r.logger.Fatal(err)
	}
	r.channel = ch

	queue, err := r.channel.QueueDeclare("NotificationReceiverQueue", false, false, false, false, nil)
	if err != nil {
		r.logger.Fatal(err)
	}
	err = ch.ExchangeDeclare("NotificationExchange", "fanout", false, false, false, false, nil)
	if err != nil {
		r.logger.Fatal(err)
	}

	err = r.channel.QueueBind(queue.Name, "#", "NotificationExchange", false, nil)
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
			r.logger.Info("Received new notification message")
			var query client.SendNotificationMessage
			err := json.Unmarshal(message.Body, &query)
			if err != nil {
				r.logger.Errorf("Unable to unmarshal message: %v", string(message.Body))
			} else {
				r.service.SendNotification(context.Background(), &query)
			}
		}
	}()
	r.logger.Infof("Waiting for new notifications...")
}

func (r *NotificationReceiver) Close() error {
	return r.channel.Close()
}

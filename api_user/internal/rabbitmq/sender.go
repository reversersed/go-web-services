package rabbitmq

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/reversersed/go-web-services/tree/main/api_user/pkg/logging"
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
func (s *Sender) SendUserLoginChangedMessage(ctx context.Context, userId string, newLogin string) error {
	ch, err := s.connection.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	queue, err := ch.QueueDeclare("UserLoginChangedQueue", false, false, false, false, nil)
	if err != nil {
		return err
	}
	err = ch.ExchangeDeclare("UserLoginChangedExchange", "fanout", false, false, false, false, nil)
	if err != nil {
		return err
	}
	err = ch.QueueBind(queue.Name, "#", "UserLoginChangedExchange", false, nil)
	if err != nil {
		return err
	}
	cntx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	type query struct {
		UserId   string `json:"userid"`
		Newlogin string `json:"newlogin"`
	}

	body, err := json.Marshal(&query{UserId: userId, Newlogin: newLogin})
	if err != nil {
		return err
	}
	err = ch.PublishWithContext(cntx, "UserLoginChangedExchange", "#", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	if err != nil {
		s.logger.Errorf("Error sending user login changed message: %v", err)
		return err
	}
	s.logger.Warnf("Sended user (%s) login changed to %s message", userId, newLogin)
	return nil
}
func (s *Sender) SendUserDeletedMessage(ctx context.Context, userId string) error {
	ch, err := s.connection.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	queue, err := ch.QueueDeclare("UserDeletedQueue", false, false, false, false, nil)
	if err != nil {
		return err
	}
	err = ch.ExchangeDeclare("UserDeletedExchange", "fanout", false, false, false, false, nil)
	if err != nil {
		return err
	}
	err = ch.QueueBind(queue.Name, "#", "UserDeletedExchange", false, nil)
	if err != nil {
		return err
	}
	cntx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	body, err := json.Marshal(userId)
	if err != nil {
		return err
	}
	err = ch.PublishWithContext(cntx, "UserDeletedExchange", "#", false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        body,
	})
	if err != nil {
		s.logger.Errorf("Error sending user deleted message: %v", err)
		return err
	}
	s.logger.Warnf("Sended user (%s) deleted meesage", userId)
	return nil
}

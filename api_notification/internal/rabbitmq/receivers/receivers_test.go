package receivers

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/reversersed/go-web-services/tree/main/api_notification/internal/client"
	"github.com/reversersed/go-web-services/tree/main/api_notification/internal/config"
	mock "github.com/reversersed/go-web-services/tree/main/api_notification/internal/rabbitmq/receivers/mocks"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/logging"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/rabbitmq"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/validator"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var conn *rabbitmq.RabbitClient
var logger *logging.Logger

func TestMain(m *testing.M) {
	req := testcontainers.ContainerRequest{
		Image:        "rabbitmq:3.10.7-management",
		ExposedPorts: []string{"5672/tcp"},
		WaitingFor:   wait.ForLog("Server startup complete"),
		Env:          map[string]string{"RABBITMQ_DEFAULT_USER": "user", "RABBITMQ_DEFAULT_PASS": "password"},
	}
	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := container.Terminate(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}()
	port, err := container.MappedPort(ctx, "5672")
	if err != nil {
		log.Fatal(err)
	}
	cfg := &config.RabbitConfig{Rabbit_Port: port.Port(), Rabbit_Pass: "password", Rabbit_User: "user"}
	cfg.Rabbit_Host, err = container.Host(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log, _ := test.NewNullLogger()
	logger = &logging.Logger{Entry: logrus.NewEntry(log)}

	conn, err = rabbitmq.New(cfg, logger)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		conn.Close()
	}()

	os.Exit(m.Run())
}

func TestNotification(t *testing.T) {
	if testing.Short() {
		t.Skip("integration tests are not run in short mode")
	}
	ctrl := gomock.NewController(t)
	service := mock.NewMocknotification_service(ctrl)
	receiver := NewNotificationReceiver(conn.Connection, validator.New(), logger, service)

	Query := make(chan *client.SendNotificationMessage, 1)
	service.EXPECT().SendNotification(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, query *client.SendNotificationMessage) {
		Query <- query
	})

	receiver.Start()
	defer receiver.Close()

	sendQuery := &client.SendNotificationMessage{
		UserId:  "userId",
		Content: "hello world",
		Type:    client.Info,
	}

	ch, err := conn.Channel()
	if !assert.NoError(t, err) {
		return
	}
	body, err := json.Marshal(sendQuery)
	if !assert.NoError(t, err) {
		return
	}

	err = ch.PublishWithContext(context.Background(), "NotificationExchange", "#", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	if assert.NoError(t, err) {
		notification := <-Query
		assert.Equal(t, sendQuery, notification)
	}
}

func TestUserDeleted(t *testing.T) {
	if testing.Short() {
		t.Skip("integration tests are not run in short mode")
	}
	ctrl := gomock.NewController(t)
	service := mock.NewMockuser_deleted_service(ctrl)
	receiver := NewUserDeletedReceiver(conn.Connection, validator.New(), logger, service)

	Query := make(chan string, 1)
	service.EXPECT().OnUserDeleted(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, userId string) {
		Query <- userId
	})

	receiver.Start()
	defer receiver.Close()

	ch, err := conn.Channel()
	if !assert.NoError(t, err) {
		return
	}

	err = ch.PublishWithContext(context.Background(), "UserDeletedExchange", "#", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        []byte("userId"),
	})
	if assert.NoError(t, err) {
		userid := <-Query
		assert.Equal(t, "userId", userid)
	}
}

func TestUserLoginChanged(t *testing.T) {
	if testing.Short() {
		t.Skip("integration tests are not run in short mode")
	}
	ctrl := gomock.NewController(t)
	service := mock.NewMockuser_login_changed_service(ctrl)
	receiver := NewUserLoginChangedReceiver(conn.Connection, validator.New(), logger, service)

	Query := make(chan *client.UserLoginChangedMessage, 1)
	service.EXPECT().OnUserLoginChanged(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, query *client.UserLoginChangedMessage) {
		Query <- query
	})

	receiver.Start()
	defer receiver.Close()

	sendQuery := &client.UserLoginChangedMessage{
		UserId:   "userId",
		NewLogin: "new_user",
	}

	ch, err := conn.Channel()
	if !assert.NoError(t, err) {
		return
	}
	body, err := json.Marshal(sendQuery)
	if !assert.NoError(t, err) {
		return
	}

	err = ch.PublishWithContext(context.Background(), "UserLoginChangedExchange", "#", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	if assert.NoError(t, err) {
		notification := <-Query
		assert.Equal(t, sendQuery, notification)
	}
}

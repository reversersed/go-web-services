package rabbitmq

import (
	"context"
	"testing"

	"github.com/reversersed/go-web-services/tree/main/api_notification/internal/config"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/logging"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestNewClient(t *testing.T) {
	if testing.Short() {
		t.Skip("integration tests are not running in short mode")
	}
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
	assert.NoError(t, err)
	defer func() {
		err := container.Terminate(ctx)
		assert.NoError(t, err)
	}()
	port, err := container.MappedPort(ctx, "5672")
	if !assert.NoError(t, err) {
		return
	}
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	cfg := &config.RabbitConfig{Rabbit_Port: port.Port(), Rabbit_Pass: "password", Rabbit_User: "user"}
	cfg.Rabbit_Host, err = container.Host(ctx)
	assert.NoError(t, err)
	_, err = New(cfg, logger)
	assert.NoError(t, err)
}
func TestWrongClient(t *testing.T) {
	if testing.Short() {
		t.Skip("integration tests are not running in short mode")
	}
	log, _ := test.NewNullLogger()
	logger := &logging.Logger{Entry: logrus.NewEntry(log)}
	_, err := New(&config.RabbitConfig{Rabbit_Host: "localhost", Rabbit_Port: "123213", Rabbit_User: "user", Rabbit_Pass: "pass"}, logger)

	assert.Error(t, err)
}

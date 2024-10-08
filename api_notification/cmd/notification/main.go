package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/reversersed/go-web-services/tree/main/api_notification/internal/client"
	"github.com/reversersed/go-web-services/tree/main/api_notification/internal/client/db"
	"github.com/reversersed/go-web-services/tree/main/api_notification/internal/config"
	"github.com/reversersed/go-web-services/tree/main/api_notification/internal/handlers/notification"
	"github.com/reversersed/go-web-services/tree/main/api_notification/internal/rabbitmq/receivers"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/cache/freecache"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/logging"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/mongo"
	rabbitClient "github.com/reversersed/go-web-services/tree/main/api_notification/pkg/rabbitmq"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/shutdown"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/validator"
)

func main() {
	logger := logging.GetLogger()
	logger.Info("logger initialized")

	logger.Info("config initializing...")
	config := config.GetConfig()

	logger.Info("router initializing...")
	router := httprouter.New()

	logger.Info("cache initializing...")
	cache := freecache.NewCache(104857600) // 100 mb

	logger.Info("database initializing...")
	db_client, err := mongo.NewClient(context.Background(), config.Database)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Info("validator initializing...")
	validator := validator.New()

	logger.Info("services initializing...")
	storage := db.NewStorage(db_client, config.Database.Db_Base, logger)
	service := client.NewService(storage, logger, cache, validator, config.Urls)

	logger.Info("rabbitmq initializing...")
	rabbit, err := rabbitClient.New(config.Rabbit, logger)
	if err != nil {
		logger.Fatal(err)
	}
	notifReceiver := receivers.NewNotificationReceiver(rabbit.Connection, validator, logger, service)
	notifReceiver.Start()

	userDeletedReceiver := receivers.NewUserDeletedReceiver(rabbit.Connection, validator, logger, service)
	userDeletedReceiver.Start()

	userLoginChangedReceiver := receivers.NewUserLoginChangedReceiver(rabbit.Connection, validator, logger, service)
	userLoginChangedReceiver.Start()

	logger.Info("handlers registration...")
	handler := notification.Handler{Service: service, Logger: logger, Validator: validator}
	handler.Register(router)

	logger.Info("starting application...")
	start(router, logger, config.Server, rabbit, notifReceiver, userDeletedReceiver, userLoginChangedReceiver)
}
func start(router *httprouter.Router, logger *logging.Logger, cfg *config.ServerConfig, closers ...io.Closer) {
	var server *http.Server
	var listener net.Listener

	logger.Infof("bind application to host: %s and port: %d", cfg.ListenAddress, cfg.ListenPort)

	var err error

	listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.ListenAddress, cfg.ListenPort))
	if err != nil {
		logger.Fatal(err)
	}

	server = &http.Server{
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go shutdown.Graceful(logger, []os.Signal{syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGHUP, os.Interrupt, syscall.SIGTERM},
		append(closers, server)...)

	logger.Infof("application initialized and started as %s", cfg.Environment)

	if err := server.Serve(listener); err != nil {
		switch {
		case errors.Is(err, http.ErrServerClosed):
			logger.Info("server shutdown")
		default:
			logger.Fatal(err)
		}
	}
}

package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/reversersed/go-web-services/tree/main/api_books/internal/client"
	"github.com/reversersed/go-web-services/tree/main/api_books/internal/client/db"
	"github.com/reversersed/go-web-services/tree/main/api_books/internal/config"
	"github.com/reversersed/go-web-services/tree/main/api_books/internal/handlers/book"
	"github.com/reversersed/go-web-services/tree/main/api_books/internal/rabbitmq"
	"github.com/reversersed/go-web-services/tree/main/api_books/pkg/cache/freecache"
	"github.com/reversersed/go-web-services/tree/main/api_books/pkg/logging"
	"github.com/reversersed/go-web-services/tree/main/api_books/pkg/mongo"
	RabbitClient "github.com/reversersed/go-web-services/tree/main/api_books/pkg/rabbitmq"
	"github.com/reversersed/go-web-services/tree/main/api_books/pkg/shutdown"
	"github.com/reversersed/go-web-services/tree/main/api_books/pkg/validator"
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

	logger.Info("rabbitmq initializing...")
	rabbit, err := RabbitClient.New(config.Rabbit, logger)
	if err != nil {
		logger.Fatal(err)
	}
	rabbitSender := rabbitmq.NewSender(rabbit.Connection, logger)

	logger.Info("services initializing...")
	bookStorage := db.NewStorage(db_client, "books", logger)
	bookService := client.NewService(bookStorage, logger, cache, validator.New())

	logger.Info("handlers registration...")
	handler := book.Handler{Logger: logger, BookService: bookService}
	handler.Register(router)

	logger.Info("starting application...")
	start(router, logger, config.Server, rabbit, rabbitSender)
}
func start(router *httprouter.Router, logger *logging.Logger, cfg *config.ServerConfig, rabbit *RabbitClient.RabbitClient, rabbitSender *rabbitmq.Sender) {
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
		server, rabbit, rabbitSender)

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

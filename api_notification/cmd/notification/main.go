package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/reversersed/go-web-services/tree/main/api_notification/internal/client"
	"github.com/reversersed/go-web-services/tree/main/api_notification/internal/client/db"
	"github.com/reversersed/go-web-services/tree/main/api_notification/internal/config"
	"github.com/reversersed/go-web-services/tree/main/api_notification/internal/handlers/notification"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/cache/freecache"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/logging"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/mongo"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/shutdown"
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
	debug.SetGCPercent(40)

	logger.Info("database initializing...")
	db_client, err := mongo.NewClient(context.Background(), config)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Info("services initializing...")
	storage := db.NewStorage(db_client, config.Db_Base, logger)
	service := client.NewService(storage, logger, cache)

	logger.Info("handlers registration...")
	handler := notification.Handler{Service: service, Logger: logger}
	handler.Register(router)

	logger.Info("starting application...")
	start(router, logger, config)
}
func start(router *httprouter.Router, logger *logging.Logger, cfg *config.Config) {
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

	go shutdown.Graceful([]os.Signal{syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGHUP, os.Interrupt, syscall.SIGTERM},
		server)

	logger.Info("application initialized and started")

	if err := server.Serve(listener); err != nil {
		switch {
		case errors.Is(err, http.ErrServerClosed):
			logger.Info("server shutdown")
		default:
			logger.Fatal(err)
		}
	}
}

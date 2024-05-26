package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/julienschmidt/httprouter"
	_ "github.com/reversersed/go-web-services/tree/main/api_gateway/docs"
	user "github.com/reversersed/go-web-services/tree/main/api_gateway/internal/client/user"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/internal/config"
	auth "github.com/reversersed/go-web-services/tree/main/api_gateway/internal/handlers/user"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/cache/freecache"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/jwt"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/shutdown"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/validator"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title API
// @version 1.0

// @host localhost:9000
// @BasePath /api/v1/

// @scheme http
// @accept json

// @securityDefinitions.apiKey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	logger := logging.GetLogger()
	logger.Info("logger initialized")

	logger.Info("config initializing...")
	config := config.GetConfig()

	logger.Info("router initializing...")
	router := httprouter.New()

	logger.Info("cache initializing...")
	cache := freecache.NewCache(104857600) // 100 mb

	logger.Info("validator initializing...")
	validator := validator.New()

	logger.Info("services initializing....")
	jwtService := jwt.NewService(cache, logger, validator)

	logger.Info("handlers registration...")
	router.GET("/swagger/:any", swaggerHandler)

	user_service := user.NewService(config.UserServiceURL, "/users", logger)
	user_handler := auth.Handler{Logger: logger, UserService: user_service, JwtService: jwtService, Validator: validator}
	user_handler.Register(router)

	logger.Info("starting application...")
	start(router, logger, config)
}
func swaggerHandler(res http.ResponseWriter, req *http.Request, p httprouter.Params) {
	httpSwagger.WrapHandler(res, req)
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

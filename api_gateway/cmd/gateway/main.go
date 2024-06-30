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
	"github.com/reversersed/go-web-services/tree/main/api_gateway/internal/client/book"
	user "github.com/reversersed/go-web-services/tree/main/api_gateway/internal/client/user"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/internal/config"
	bh "github.com/reversersed/go-web-services/tree/main/api_gateway/internal/handlers/book"
	gh "github.com/reversersed/go-web-services/tree/main/api_gateway/internal/handlers/genre"
	auth "github.com/reversersed/go-web-services/tree/main/api_gateway/internal/handlers/user"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/cache/freecache"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/jwt"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/shutdown"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/validator"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title API
// @version 1.0

// @host localhost:9000
// @BasePath /api/v1/

// @scheme http
// @accept json

// @securityDefinitions.apiKey ApiKeyAuth
// @in Cookie
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
	jwtService := jwt.NewService(cache, logger, validator, config.Jwt.SecretToken)

	logger.Info("handlers registration...")
	//swagger
	if config.Server.Environment == "debug" {
		logger.Info("swagger registration...")
		router.GET("/swagger/:any", func(res http.ResponseWriter, req *http.Request, p httprouter.Params) {
			httpSwagger.WrapHandler(res, req)
		})
	}

	user_service := user.NewService(config.Urls.UserServiceURL, "/users", logger)
	user_handler := &auth.Handler{Logger: logger, UserService: user_service, JwtService: jwtService, Validator: validator}
	user_handler.Register(router)

	book_service := book.NewService(config.Urls.BookServiceURL, "/books", logger)
	book_handler := &bh.Handler{Logger: logger, BookService: book_service, JwtService: jwtService, Validator: validator}
	book_handler.Register(router)

	genre_service := book.NewService(config.Urls.GenresServiceURL, "/genres", logger)
	genre_handler := &gh.Handler{Logger: logger, GenreService: genre_service, JwtService: jwtService, Validator: validator}
	genre_handler.Register(router)

	logger.Info("starting application...")
	start(router, logger, config.Server)
}

func start(router *httprouter.Router, logger *logging.Logger, cfg *config.ServerConfig) {
	var server *http.Server
	var listener net.Listener

	logger.Infof("bind application to host: %s and port: %d", cfg.ListenAddress, cfg.ListenPort)

	var err error

	listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.ListenAddress, cfg.ListenPort))
	if err != nil {
		logger.Fatal(err)
	}

	server = &http.Server{
		Handler:      cors.AllowAll().Handler(router),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go shutdown.Graceful(logger, []os.Signal{syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGHUP, os.Interrupt, syscall.SIGTERM},
		server)

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

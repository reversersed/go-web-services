package main

import (
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
)

func main() {
	logger := logging.NewLogger()
	logger.Println("logger initialized")
}
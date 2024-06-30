package genre

import (
	"net/http"
	"time"

	base "github.com/reversersed/go-web-services/tree/main/api_gateway/internal/client"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/rest"
)

type client struct {
	base.BaseClient
}

func NewService(baseURL, path string, logger *logging.Logger) *client {
	return &client{BaseClient: base.BaseClient{
		Path: path,
		Base: &rest.RestClient{
			BaseURL: baseURL,
			HttpClient: &http.Client{
				Timeout: 10 * time.Second,
			},
			Logger: logger,
		},
	}}
}

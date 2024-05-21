package config

import (
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
)

type Config struct {
	SecretToken    string `env:"jwt_secret" env-required:"true"`
	ListenAddress  string `env:"ip_addr" env-required:"true"`
	ListenPort     int    `env:"port" env-required:"true"`
	UserServiceURL string `env:"srv_url_user" env-required:"true"`
}

var cfg *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		logger := logging.GetLogger()
		logger.Info("reading api config...")
		cfg = &Config{}

		if err := cleanenv.ReadConfig("config/.env", cfg); err != nil {
			desc, _ := cleanenv.GetDescription(cfg, nil)
			logger.Info(desc)
			logger.Fatal(err)
		}
	})
	return cfg
}

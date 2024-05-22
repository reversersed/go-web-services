package config

import (
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
)

type Config struct {
	ListenAddress  string `env:"IP_ADDR" env-required:"true"`
	ListenPort     int    `env:"PORT" env-required:"true"`
	UserServiceURL string `env:"SRV_URL_USER" env-required:"true"`
	SecretToken    string `env:"JWT_SECRET" env-required:"true"`
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
			logger.Error(desc)
			logger.Fatal(err)
		}
	})
	return cfg
}

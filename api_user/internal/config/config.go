package config

import (
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/reversersed/go-web-services/tree/main/api_user/pkg/logging"
)

type Config struct {
	ListenAddress string `env:"HOST" env-required:"true"`
	ListenPort    int    `env:"PORT" env-required:"true"`
	Db_Host       string `env:"DB_HOST" env-required:"true"`
	Db_Base       string `env:"DB_BASE" env-required:"true"`
	Db_Port       int    `env:"DB_PORT" env-required:"true"`
	Db_Name       string `env:"DB_NAME"`
	Db_Pass       string `env:"DB_PASS"`
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

package config

import (
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/reversersed/go-web-services/tree/main/api_notification/pkg/logging"
)

type ServerConfig struct {
	ListenAddress string `env:"HOST" env-required:"true"`
	ListenPort    int    `env:"PORT" env-required:"true"`
	Environment   string `env:"ENVIRONMENT"`
}
type DbConfig struct {
	Db_Host string `env:"DB_HOST" env-required:"true"`
	Db_Base string `env:"DB_BASE" env-required:"true"`
	Db_Port int    `env:"DB_PORT" env-required:"true"`
	Db_Name string `env:"DB_NAME"`
	Db_Pass string `env:"DB_PASS"`
	Db_Auth string `env:"DB_AUTHDB"`
}
type UrlConfig struct {
	Url_User_Service string `env:"SRV_URL_USER" env-required:"true"`
}
type RabbitConfig struct {
	Rabbit_Host string `env:"RABBITMQ_HOST" env-required:"true"`
	Rabbit_Port string `env:"RABBITMQ_PORT" env-required:"true"`
	Rabbit_User string `env:"RABBITMQ_USER" env-required:"true"`
	Rabbit_Pass string `env:"RABBITMQ_PASS" env-required:"true"`
}
type Config struct {
	Server   *ServerConfig
	Database *DbConfig
	Urls     *UrlConfig
	Rabbit   *RabbitConfig
}

var cfg *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		logger := logging.GetLogger()
		logger.Info("reading api config...")
		srvCfg := &ServerConfig{}
		dbCfg := &DbConfig{}
		urlCfg := &UrlConfig{}
		rabbitCfg := &RabbitConfig{}

		if err := cleanenv.ReadConfig("config/.env", srvCfg); err != nil {
			desc, _ := cleanenv.GetDescription(cfg, nil)
			logger.Error(desc)
			logger.Fatal(err)
		}
		if len(srvCfg.Environment) == 0 {
			srvCfg.Environment = "debug"
		}
		if err := cleanenv.ReadConfig("config/.env", dbCfg); err != nil {
			desc, _ := cleanenv.GetDescription(cfg, nil)
			logger.Error(desc)
			logger.Fatal(err)
		}
		if err := cleanenv.ReadConfig("config/.env", urlCfg); err != nil {
			desc, _ := cleanenv.GetDescription(cfg, nil)
			logger.Error(desc)
			logger.Fatal(err)
		}
		if err := cleanenv.ReadConfig("config/.env", rabbitCfg); err != nil {
			desc, _ := cleanenv.GetDescription(cfg, nil)
			logger.Error(desc)
			logger.Fatal(err)
		}
		cfg = &Config{
			Server:   srvCfg,
			Database: dbCfg,
			Urls:     urlCfg,
			Rabbit:   rabbitCfg,
		}
	})
	return cfg
}

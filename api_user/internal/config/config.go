package config

import (
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/reversersed/go-web-services/tree/main/api_user/pkg/logging"
)

type ServerConfig struct {
	ListenAddress string `env:"HOST" env-required:"true"`
	ListenPort    int    `env:"PORT" env-required:"true"`
}
type DatabaseConfig struct {
	Db_Host string `env:"DB_HOST" env-required:"true"`
	Db_Base string `env:"DB_BASE" env-required:"true"`
	Db_Port int    `env:"DB_PORT" env-required:"true"`
	Db_Name string `env:"DB_NAME"`
	Db_Pass string `env:"DB_PASS"`
	Db_Auth string `env:"DB_AUTHDB"`
}
type SmtpConfig struct {
	SmtpHost     string `env:"SMTP_HOST"`
	SmtpPort     int    `env:"SMTP_PORT"`
	SmtpLogin    string `env:"SMTP_LOGIN"`
	SmtpPassword string `env:"SMTP_PASS"`
}
type Config struct {
	Server   *ServerConfig
	Database *DatabaseConfig
	SMTP     *SmtpConfig
}

var cfg *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		logger := logging.GetLogger()
		logger.Info("reading api config...")
		srvCfg := &ServerConfig{}
		dbCfg := &DatabaseConfig{}
		smtpCfg := &SmtpConfig{}

		if err := cleanenv.ReadConfig("config/.env", srvCfg); err != nil {
			desc, _ := cleanenv.GetDescription(cfg, nil)
			logger.Error(desc)
			logger.Fatal(err)
		}
		if err := cleanenv.ReadConfig("config/.env", dbCfg); err != nil {
			desc, _ := cleanenv.GetDescription(cfg, nil)
			logger.Error(desc)
			logger.Fatal(err)
		}
		if err := cleanenv.ReadConfig("config/.env", smtpCfg); err != nil {
			desc, _ := cleanenv.GetDescription(cfg, nil)
			logger.Error(desc)
			logger.Fatal(err)
		}
		cfg = &Config{
			Server:   srvCfg,
			Database: dbCfg,
			SMTP:     smtpCfg,
		}
	})
	return cfg
}

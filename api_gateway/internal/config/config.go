package config

import (
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/reversersed/go-web-services/tree/main/api_gateway/pkg/logging"
)

type ServerConfig struct {
	ListenAddress string `env:"IP_ADDR" env-required:"true"`
	ListenPort    int    `env:"PORT" env-required:"true"`
}
type UrlConfig struct {
	UserServiceURL string `env:"SRV_URL_USER" env-required:"true"`
}
type JwtConfig struct {
	SecretToken string `env:"JWT_SECRET" env-required:"true"`
}
type Config struct {
	Server *ServerConfig
	Urls   *UrlConfig
	Jwt    *JwtConfig
}

var cfg *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		logger := logging.GetLogger()
		logger.Info("reading api config...")
		srvCfg := &ServerConfig{}
		urlCfg := &UrlConfig{}
		jwtCfg := &JwtConfig{}

		if err := cleanenv.ReadConfig("config/.env", srvCfg); err != nil {
			desc, _ := cleanenv.GetDescription(cfg, nil)
			logger.Error(desc)
			logger.Fatal(err)
		}
		if err := cleanenv.ReadConfig("config/.env", urlCfg); err != nil {
			desc, _ := cleanenv.GetDescription(cfg, nil)
			logger.Error(desc)
			logger.Fatal(err)
		}
		if err := cleanenv.ReadConfig("config/.env", jwtCfg); err != nil {
			desc, _ := cleanenv.GetDescription(cfg, nil)
			logger.Error(desc)
			logger.Fatal(err)
		}
		cfg = &Config{
			Server: srvCfg,
			Urls:   urlCfg,
			Jwt:    jwtCfg,
		}
	})
	return cfg
}

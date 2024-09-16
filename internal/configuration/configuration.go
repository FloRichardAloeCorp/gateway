package configuration

import (
	"fmt"
	"strings"
	"time"

	"github.com/Aloe-Corporation/logs"
	"github.com/FloRichardAloeCorp/gateway/internal/middlewares/auth"
	"github.com/FloRichardAloeCorp/gateway/internal/service"
	"github.com/spf13/viper"
)

var (
	log = logs.Get()
)

type Config struct {
	Logger      *logs.Conf       `mapstructure:"logger"`
	Server      ServerConfig     `mapstructure:"server"`
	Services    []service.Config `mapstructure:"services"`
	Middlewares MiddlewaresConf  `mapstructure:"middlewares"`
}

type ServerConfig struct {
	Port int        `mapstructure:"port"`
	Cors CorsConfig `mapstructure:"cors"`
}

type CorsConfig struct {
	AllowOrigins     []string      `mapstructure:"allow_origins"`
	AllowMethods     []string      `mapstructure:"allow_methods"`
	AllowHeaders     []string      `mapstructure:"allow_headers"`
	ExposeHeaders    []string      `mapstructure:"expose_headers"`
	AllowCredentials bool          `mapstructure:"allow_credentials"`
	MaxAge           time.Duration `mapstructure:"max_age"`
}

type MiddlewaresConf struct {
	Auth auth.AuthMiddlewareConfig `mapstructure:"auth"`
}

// LoadConf load the configuration from the file at the given path.
func LoadConf(path, prefix string) (*Config, error) {
	conf := new(Config)

	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix(prefix)
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("can't load gateway configuration : %w", err)
	}

	err = viper.Unmarshal(&conf)
	if err != nil {
		return nil, fmt.Errorf("can't unmarshall gateway configuration : %w", err)
	}

	mergeAuthMiddlewareConfig(conf)
	return conf, nil
}

func mergeAuthMiddlewareConfig(conf *Config) {
	globalMiddlewareConfig := conf.Middlewares.Auth

	for i := 0; i < len(conf.Services); i++ {
		for j := 0; j < len(conf.Services[i].Endpoints); j++ {
			conf.Services[i].Middlewares.Auth.AuthMiddlewareConfig = globalMiddlewareConfig
		}
	}
}

package config

import (
	"sync"

	"github.com/delonce/apishop/pkg/logging"
	"github.com/spf13/viper"
)

type Config struct {
	Host string `mapstructure:"HOST"`
	Port uint16 `mapstructure:"HOST_PORT"`

	DBAddr   string `mapstructure:"DB_ADDR"`
	DBPort   string `mapstructure:"DB_PORT"`
	DBName   string `mapstructure:"DB_NAME"`
	DBLogin  string `mapstructure:"DB_LOGIN"`
	DBPasswd string `mapstructure:"DB_PASSWD"`
}

var instance Config
var once sync.Once

func GetConfig(logger *logging.Logger) *Config {
	// Making sure the code is executed once
	once.Do(func() {
		logger.Info("Getting config from .env file")

		viper.SetConfigFile(".env")
		viper.AutomaticEnv()

		// Just reading our config
		err := viper.ReadInConfig()

		if err != nil {
			logger.Panicf("Error reading config file, %v", err)
		}

		// Uses struct Config to handy access
		err = viper.Unmarshal(&instance)

		if err != nil {
			logger.Panicf("Error unmarshalling config file, %v", err)
		}
	})

	logger.Info(instance)
	return &instance
}

package main

import (
	"context"

	"github.com/delonce/apishop/internal/app"
	"github.com/delonce/apishop/internal/config"
	"github.com/delonce/apishop/pkg/logging"
)

func main() {
	// Point of entry
	// Getting logger, config, starting application
	logger := logging.GetLogger()
	logger.Info("Logger had started")
	mainConfig := loadConfig(logger)
	logger.Info("Config had loaded")

	mainApp := app.NewApp(context.Background(), logger, mainConfig)
	mainApp.StartConsumerApplication()
}

func loadConfig(logger *logging.Logger) *config.Config {
	return config.GetConfig(logger)
}

package logger

import (
	"os"

	"go.uber.org/zap"
)

var log *zap.Logger

func initLogger() {

	var err error
	if os.Getenv("env") == "test" {
		log, err = zap.NewDevelopment(zap.AddCaller())
	} else {
		log, err = zap.NewProduction(zap.AddCaller())
	}
	log.Info("Logger initialized")

	if err != nil {
		panic(err)
	}
}

// GetLogger godoc
// Get Global Logger
func GetLogger() *zap.Logger {
	if log == nil {
		initLogger()
	}
	return log
}

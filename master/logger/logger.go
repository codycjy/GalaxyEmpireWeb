package logger

import (
	"os"

	"go.uber.org/zap"
)

var log *zap.Logger

func init() {

	var err error
	if os.Getenv("env") == "test" {
		log, err = zap.NewDevelopment(zap.AddCaller())
	} else {
		log, err = zap.NewProduction(zap.AddCaller())
	}

	if err != nil {
		panic(err)
	}
}

func GetLogger() *zap.Logger {
	return log
}

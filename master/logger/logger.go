package logger

import (
	"go.uber.org/zap"
)

var log *zap.Logger

func init() {
	var err error
	log, err = zap.NewProduction(zap.AddCaller())
	if err != nil {
		panic(err)
	}
}

func GetLogger() *zap.Logger {
	return log
}

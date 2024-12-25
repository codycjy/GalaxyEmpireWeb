package logger

import (
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	log  *zap.Logger
	once sync.Once
)

func initLogger() {
	var err error

	// 基础配置
	var config zap.Config
	if os.Getenv("ENV") == "test" {
		fmt.Println("****************** test ******************")
		config = zap.NewDevelopmentConfig()
		// 设置日志级别为 Debug
		config.Level.SetLevel(zap.DebugLevel)
	} else {
		config = zap.NewProductionConfig()
		// 生产环境使用 Info 级别
		config.Level.SetLevel(zap.InfoLevel)
	}

	// 通用配置
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.StacktraceKey = "stacktrace"

	// 从环境变量获取日志级别（如果设置了的话）
	if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
		var level zapcore.Level
		if err := level.UnmarshalText([]byte(lvl)); err == nil {
			config.Level.SetLevel(level)
		}
	}

	log, err = config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	if err != nil {
		panic(err)
	}

	// 替换全局 logger
	zap.ReplaceGlobals(log)

	log.Info("Logger initialized",
		zap.String("env", os.Getenv("env")),
		zap.String("level", config.Level.String()),
	)
}

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	once.Do(initLogger)
	return log
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal logs a fatal message and then calls os.Exit(1)
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// WithFields creates a child logger with the given fields
func WithFields(fields ...zap.Field) *zap.Logger {
	return GetLogger().With(fields...)
}

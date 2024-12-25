package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	log  *zap.Logger
	once sync.Once
)

// getLogPath returns the log file path with date
func getLogPath() string {
	// 确保日志目录存在
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(err)
	}

	// 使用日期作为日志文件名
	date := time.Now().Format("2006-01-02")
	return filepath.Join(logDir, fmt.Sprintf("%s.log", date))
}

func initLogger() {
	var err error

	// 基础配置
	var config zap.Config
	if os.Getenv("ENV") == "test" {
		fmt.Println("****************** test ******************")
		config = zap.NewDevelopmentConfig()
		config.Level.SetLevel(zap.DebugLevel)
	} else {
		config = zap.NewProductionConfig()
		config.Level.SetLevel(zap.InfoLevel)
	}

	// 同时输出到文件和标准输出
	logPath := getLogPath()
	config.OutputPaths = []string{
		"stdout",
		logPath,
	}
	// 错误日志同时输出到错误文件和标准错误
	errorLogPath := logPath + ".error"
	config.ErrorOutputPaths = []string{
		"stderr",
		errorLogPath,
	}

	// JSON编码器配置
	config.Encoding = "json"
	config.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

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
		zap.String("env", os.Getenv("ENV")),
		zap.String("level", config.Level.String()),
		zap.String("logPath", logPath),
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

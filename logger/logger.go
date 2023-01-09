package logger

import (
	"fmt"
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// https://pkg.go.dev/go.uber.org/zap
// https://betterstack.com/community/guides/logging/logging-in-go/

const (
	filename = "test.log"
)

func Development() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	config.Development = true
	config.Encoding = "console"
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("Jan-02-15:04:05.00")

	logger, err := config.Build()
	if err != nil {
		log.Fatalln(fmt.Errorf("could not initialize the zap logger: %w", err))
	}
	return logger
}

func Production() *zap.Logger {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	config.Development = false
	//config.Encoding = "console"

	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("Jan-02-15:04:05.00")
	console := zapcore.NewConsoleEncoder(config.EncoderConfig)

	// logger, err := config.Build()
	// if err != nil {
	// 	log.Fatalln(fmt.Errorf("could not initialize the zap logger: %w", err))
	// }
	core := zapcore.NewTee(
		File(),
		zapcore.NewCore(console, zapcore.AddSync(os.Stdout), zapcore.InfoLevel),
	)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return logger
}

func File() zapcore.Core {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	logFile, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	fileEncoder := zapcore.NewJSONEncoder(config)
	writer := zapcore.AddSync(logFile)
	defaultLogLevel := zapcore.DebugLevel
	core := zapcore.NewCore(fileEncoder, writer, defaultLogLevel)
	return core
}

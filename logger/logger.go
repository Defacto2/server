package logger

import (
	"fmt"
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// https://pkg.go.dev/go.uber.org/zap
// https://betterstack.com/community/guides/logging/logging-in-go/
// https://github.com/uber-go/zap/blob/master/FAQ.md
// https://pkg.go.dev/gopkg.in/natefinch/lumberjack.v2@v2.0.0#section-readme

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
		FileRotation(),
		zapcore.NewCore(console, zapcore.AddSync(os.Stdout), zapcore.InfoLevel),
	)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return logger
}

func FileRotation() zapcore.Core {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filename,
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	})
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(config),
		w,
		zap.InfoLevel,
	)
	return core
}

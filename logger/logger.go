// Package logger uses the zap logging library for application logs.
package logger

import (
	"fmt"
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	ServerLog  = "server.log"
	HTTP500Log = "http500.log"
	HTTP400Log = "http400.log"
	HTTP300Log = "http300.log"
)

/*
	https://github.com/uber-go/zap
	https://pkg.go.dev/go.uber.org/zap
	https://pkg.go.dev/go.uber.org/zap@v1.24.0/zapcore
	https://pkg.go.dev/gopkg.in/natefinch/lumberjack.v2@v2.0.0

	Using Zap - Simple use cases
	https://blog.sandipb.net/2018/05/02/using-zap-simple-use-cases/

	Structured Logging in Golang with Zap
	https://codewithmukesh.com/blog/structured-logging-in-golang-with-zap/
*/

/*
const (
	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	DebugLevel = zapcore.DebugLevel
	// InfoLevel is the default logging priority.
	InfoLevel = zapcore.InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel = zapcore.WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel = zapcore.ErrorLevel
	// DPanicLevel logs are particularly important errors. In development the
	// logger panics after writing the message.
	DPanicLevel = zapcore.DPanicLevel
	// PanicLevel logs a message, then panics.
	PanicLevel = zapcore.PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel = zapcore.FatalLevel
)
*/

func Development() *zap.Logger {
	config := zap.NewDevelopmentEncoderConfig()
	config.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(config)
	defaultLogLevel := zapcore.DebugLevel
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel),
	)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

// https://codewithmukesh.com/blog/structured-logging-in-golang-with-zap/
// https://blog.sandipb.net/2018/05/02/using-zap-simple-use-cases/
func Production() *zap.Logger {
	config := zap.NewProductionConfig()
	// config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	// config.Development = false
	// //config.Encoding = "console"

	// config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("Jan-02-15:04:05.00")
	// console := zapcore.NewConsoleEncoder(config.EncoderConfig)

	logger, err := config.Build()
	if err != nil {
		log.Fatalln(fmt.Errorf("could not initialize the zap logger: %w", err))
	}
	// core := zapcore.NewTee(
	// 	FileRotation(),
	// 	zapcore.NewCore(console, zapcore.AddSync(os.Stdout), zapcore.InfoLevel),
	// )
	// logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return logger
}

// func FileRotation() zapcore.Core {
// 	config := zap.NewProductionEncoderConfig()
// 	config.EncodeTime = zapcore.ISO8601TimeEncoder

// 	w := zapcore.AddSync(&lumberjack.Logger{
// 		Filename:   filename,
// 		MaxSize:    500, // megabytes
// 		MaxBackups: 3,
// 		MaxAge:     28, // days
// 	})
// 	core := zapcore.NewCore(
// 		zapcore.NewJSONEncoder(config),
// 		w,
// 		zap.InfoLevel,
// 	)
// 	return core
// }

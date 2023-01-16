// Package logger uses the zap logging library for application logs.
package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	ServerLog  = "server.log" // Filename of the Error, Panic and Fatal level log.
	InfoLog    = "info.log"   // Filename of the Warn and Info level log.
	MaxSizeMB  = 100          // Maximum file size in megabytes before a log rotation.
	MaxBackups = 5            // Maximum number of rotated logs to keep, older logs are deleted.
	MaxDays    = 45           // Maximum days a log is kept before a log rotation.
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

// Development logger prints all log levels to stdout.
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

// Production logger prints all info and higher log levels to files.
func Production(root string) *zap.Logger {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.TimeEncoderOfLayout("Jan-02-15:04:05.00")
	jsonEncoder := zapcore.NewJSONEncoder(config)

	// server breakage log
	servWr := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(root, ServerLog),
		MaxSize:    MaxSizeMB,
		MaxBackups: MaxBackups,
		MaxAge:     MaxDays,
	})

	// information and warning log
	infoWr := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(root, InfoLog),
		MaxSize:    MaxSizeMB,
		MaxBackups: MaxBackups,
		MaxAge:     MaxDays,
	})

	core := zapcore.NewTee(
		zapcore.NewCore(jsonEncoder, servWr, zapcore.FatalLevel),
		zapcore.NewCore(jsonEncoder, servWr, zapcore.PanicLevel),
		zapcore.NewCore(jsonEncoder, servWr, zapcore.ErrorLevel),
		zapcore.NewCore(jsonEncoder, infoWr, zapcore.WarnLevel),
		zapcore.NewCore(jsonEncoder, infoWr, zapcore.InfoLevel),
	)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

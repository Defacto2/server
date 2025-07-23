// Package zaplog uses the zap logging library for application logs.
// There are two logging modes, development and production.
// The production mode saves the logs to file and automatically rotates
// older files. While the development mode prints all feedback to stdout.
package zaplog

import (
	"context"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	// ServerLog is the filename of the Error, Panic and Fatal level log.
	ServerLog = "defacto2_server_panics.log"
	// InfoLog is the filename of the Warn and Info level log.
	InfoLog = "defacto2_server_info.log"
	// MaxSizeMB is the maximum file size in megabytes before a log rotation is triggered.
	MaxSizeMB = 100
	// MaxBackups is the maximum number of rotated logs to keep, older logs are deleted.
	MaxBackups = 5
	// MaxDays is the maximum days a log is kept before a log rotation.
	MaxDays = 45
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

// Status logger prints all log levels to stdout but without callers.
func Status() *zap.Logger {
	enc := TextNoTime()
	defaultLogLevel := zapcore.InfoLevel
	core := zapcore.NewTee(
		zapcore.NewCore(
			enc,
			zapcore.AddSync(os.Stdout),
			defaultLogLevel,
		),
	)
	return zap.New(core)
}

// Timestamp logger prints all log levels to stdout but without callers.
func Timestamp() *zap.Logger {
	enc := Text()
	defaultLogLevel := zapcore.InfoLevel
	core := zapcore.NewTee(
		zapcore.NewCore(
			enc,
			zapcore.AddSync(os.Stdout),
			defaultLogLevel,
		),
	)
	return zap.New(core)
}

// Debug logger prints all log levels to stdout.
func Debug() *zap.Logger {
	enc := Text()
	defaultLogLevel := zapcore.DebugLevel
	core := zapcore.NewTee(
		zapcore.NewCore(
			enc,
			zapcore.AddSync(os.Stdout),
			defaultLogLevel,
		),
	)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

// Store logger prints all info and higher log levels to files using the encoder.
// Either the zaplog.JSON, zaplog.Text, or zaplog.TextNoTime encoders can be used.
// Fatal and Panics are also returned to os.Stderr.
func Store(enc zapcore.Encoder, logPath string) *zap.Logger {
	errs := Text()
	// server breakage log
	serverWr := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(logPath, ServerLog),
		MaxSize:    MaxSizeMB,
		MaxBackups: MaxBackups,
		MaxAge:     MaxDays,
	})
	// server breakage print errors
	errWr := zapcore.AddSync(os.Stderr)

	// information and warning log
	infoWr := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filepath.Join(logPath, InfoLog),
		MaxSize:    MaxSizeMB,
		MaxBackups: MaxBackups,
		MaxAge:     MaxDays,
	})

	core := zapcore.NewTee(
		// log to stderr
		zapcore.NewCore(errs, errWr, zapcore.FatalLevel),
		zapcore.NewCore(errs, errWr, zapcore.PanicLevel),
		// log to "server.log"
		zapcore.NewCore(enc, serverWr, zapcore.FatalLevel),
		zapcore.NewCore(enc, serverWr, zapcore.PanicLevel),
		zapcore.NewCore(enc, serverWr, zapcore.ErrorLevel),
		// log to "info.log"
		zapcore.NewCore(enc, infoWr, zapcore.WarnLevel),
		zapcore.NewCore(enc, infoWr, zapcore.InfoLevel),
	)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

// JSON returns a logger in JSON format.
func JSON() zapcore.Encoder { //nolint:ireturn
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.TimeEncoderOfLayout("Jan-02-15:04:05.00")
	return zapcore.NewJSONEncoder(config)
}

// Text returns a logger in color and time.
func Text() zapcore.Encoder { //nolint:ireturn
	config := zap.NewDevelopmentEncoderConfig()
	config.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return zapcore.NewConsoleEncoder(config)
}

// TextNoTime returns a logger in color but without the time.
func TextNoTime() zapcore.Encoder { //nolint:ireturn
	config := zap.NewDevelopmentEncoderConfig()
	// config.EncodeTime = nil  // use nil to remove the leading console separator
	config.EncodeTime = zapcore.TimeEncoderOfLayout("")
	config.ConsoleSeparator = "  "
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return zapcore.NewConsoleEncoder(config)
}

type contextKey string

const LoggerKey contextKey = "logger"

// Logger returns the logger from the context.
// If the logger is not found, it panics.
//
// zaplog.Logger was previously referenced as helper.Logger
func Logger(ctx context.Context) *zap.SugaredLogger {
	logger, loggerExists := ctx.Value(LoggerKey).(*zap.SugaredLogger)
	if !loggerExists {
		panic("zaplog context named '" + LoggerKey + "' does not exist")
	}
	return logger
}

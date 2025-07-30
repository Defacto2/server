// Package out uses the slog and tint packages for the application logs.
// There are two logging modes, development and production.
// The production mode saves the logs to file and automatically rotates
// older files. While the development mode prints all feedback to stdout.
package out

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
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

type Logging slog.Logger

// LevelDebug Level = -4
// LevelInfo  Level = 0
// LevelWarn  Level = 4
// LevelError Level = 8

const (
	LevelDebug   = slog.LevelDebug
	LevelInfo    = slog.LevelInfo
	LevelNotice  = slog.Level(2)
	LevelWarning = slog.LevelWarn
	LevelError   = slog.LevelError
	LevelFatal   = slog.Level(12) // More severe than ERROR

	FatalRed    = 196 // #ff0000
	TracePurple = 206 // #ff5fff
)

// options is a placeholder
func options() tint.Options {
	w := os.Stdout
	return tint.Options{
		AddSource:  true,
		Level:      slog.LevelDebug,
		TimeFormat: time.Kitchen,
		NoColor:    !isatty.IsTerminal(w.Fd()),
	}
}

func Startup() *slog.Logger {
	opts := startup()
	sl := slog.New(tint.NewHandler(
		os.Stdout,
		&opts,
	))
	return sl
}

func startup() tint.Options {
	w := os.Stdout
	return tint.Options{
		AddSource:  false,
		Level:      slog.LevelInfo,
		TimeFormat: time.Kitchen,
		NoColor:    !isatty.IsTerminal(w.Fd()),
	}
}

// Discard all logger output.
func Discard() *slog.Logger {
	sl := slog.New(slog.DiscardHandler)
	return sl
}

// Raw displays all logs to stdout but without timestamps.
func Raw() *slog.Logger {
	opts := raw()
	sl := slog.New(slog.NewTextHandler(
		os.Stdout,
		&opts,
	))
	return sl
}

func raw() slog.HandlerOptions {
	return slog.HandlerOptions{
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			a = notime(groups, a)
			return a
		},
	}
}

// notime removes the time key from the logs.
func notime(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey && len(groups) == 0 {
		return slog.Attr{}
	}
	return a
}

func nosrcdir(a slog.Attr) slog.Attr {
	// Remove the directory from the source's filename.
	if a.Key == slog.SourceKey {
		source := a.Value.Any().(*slog.Source)
		source.File = filepath.Base(source.File)
	}
	return a
}

// Devel is the default development mode ogger that prints to the
// termminal standard error (stdout). The date/time and source path
// are included as are levels using colored three letter codes.
func Devel() *slog.Logger {
	w := os.Stdout
	// Create a new logger
	opts := options()
	opts.ReplaceAttr = defaultAttr
	logger := slog.New(tint.NewHandler(w, &opts))
	return logger
}

func defaultAttr(groups []string, a slog.Attr) slog.Attr {
	// Remove time.
	// if a.Key == slog.TimeKey && len(groups) == 0 {
	// 	return slog.Attr{}
	// }
	//
	var level slog.Level
	if a.Key == slog.LevelKey {
		level = a.Value.Any().(slog.Level)
	}
	// Remove the directory from the source's filename.
	if a.Key == slog.SourceKey {
		source := a.Value.Any().(*slog.Source)
		switch level {
		case LevelDebug:
			// skip
		default:
			source.File = filepath.Base(source.File)
		}
	}
	return custom(a)
}

// Printout is a text only logger for terminal output using the
// terminal standard output. It gets used for displaying the server
// startup configuration and the output of flags and application commands.
func Printout(w io.Writer) *slog.Logger {
	if w == nil {
		w = os.Stdout
	}
	// Create a new logger
	opts := options()
	opts.AddSource = false
	opts.ReplaceAttr = printAttr
	logger := slog.New(tint.NewHandler(w, &opts))
	return logger
}

func printAttr(groups []string, a slog.Attr) slog.Attr {
	const unset = ",unset"
	key := a.Key
	value := a.Value.String()
	// comp := strings.ToLower(value)
	// fix keys
	if strings.HasSuffix(key, unset) {
		fix := strings.TrimSuffix(key, unset)
		key = fix
		a.Key = fix
	}
	// cases for skipping
	switch key {
	case "":
		return slog.Attr{}
	case "help":
		if value == "" {
			return slog.Attr{}
		}
	case slog.TimeKey:
		if len(groups) == 0 {
			return slog.Attr{}
		}
	case "msg":
		switch strings.ToLower(value) {
		case "googleaccounts":
			return slog.Attr{}
		case "postgres", "repair":
			a.Value = slog.StringValue(strings.ToUpper(a.Value.String()))
		}
		a.Value = slog.StringValue(fmt.Sprintf("%s\n      ", a.Value))
	}
	// formatting
	if key == slog.LevelKey {
		switch key {
		case "postgres", "repair":
			a.Key = strings.ToUpper(a.Key)
		}
		a = tint.Attr(6, slog.String(a.Key, "INF  "))
	}
	if key == "issue" {
		if value == "" {
			return slog.Attr{}
		}
		a.Key = "ISSUE"
		a = tint.Attr(9, slog.String(key, value))
		return tint.Attr(9, a)
	}
	return a
}

// TODO: this maybe removed and instead consolidated with Devel()?
func Tracer() *slog.Logger {
	w := os.Stdout
	// Create a new logger
	opts := options()
	opts.ReplaceAttr = traceAttr
	logger := slog.New(tint.NewHandler(w, &opts))
	return logger
}

func traceAttr(groups []string, a slog.Attr) slog.Attr {
	return custom(a)
}

func custom(a slog.Attr) slog.Attr {
	// Format the custom level keys to use color
	if a.Key == slog.LevelKey {
		level := a.Value.Any().(slog.Level)
		switch level {
		case LevelDebug:
			a = tint.Attr(TracePurple, slog.String(a.Key, "DEBUG"))
		case LevelFatal:
			a = tint.Attr(FatalRed, slog.String(a.Key, "FATAL"))
		}
	}
	return a
}

// Fatal logs any issues and exits to the operating system.
func Fatal(l *slog.Logger, msg string, args ...slog.Attr) {
	l.LogAttrs(context.Background(), LevelFatal, msg, args...)
	os.Exit(1)
}

func Trace(l *slog.Logger, msg string, args ...slog.Attr) {
	l.LogAttrs(context.Background(), LevelDebug, msg, args...)
}

// 	logFile, err := os.OpenFile("application.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer logFile.Close()
//
// var logLevel = new(slog.LevelVar)
// 	logger := slog.NewJSONHandler(logFile, &slog.HandlerOptions{Level: logLevel})
// 	slog.SetDefault(slog.New(logger))
// 	logLevel.Set(slog.LevelDebug)

// func x() {
// 	wl := slog.LevelDebug
//
// 	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
// 		Level: wl,
// 	})
// 	logger := slog.New(handler)
//
// 	if logger.Enabled(context.Background(), slog.LevelDebug) {
// 		// This code will not run when the logger's level is INFO or greater
// 		logger.Debug("operation complete", "data", getExpensiveDebugData())
// 	}
//
// 	var logLevel slog.LevelVar // INFO is the zero value
// 	// the initial value is set from the environment and you can call Set() anytime
// 	// to update this value
// 	logLevel.Set(getLogLevelFromEnv())
//
// 	loggerx := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
// 		Level: &logLevel,
// 	}))
//
//
// 	func getLogLevelFromEnv() slog.Level {
//     levelStr := os.Getenv("LOG_LEVEL")
//
//     switch strings.ToLower(levelStr) {
//     case "debug":
//         return slog.LevelDebug
//     case "warn":
//         return slog.LevelWarn
//     case "error":
//         return slog.LevelError
//     default:
//         return slog.LevelInfo
//     }
// }
// 	    loggery := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
//         Level: getLogLevelFromEnv(),
//     }))
//
// var logLevel slog.LevelVar // INFO is the zero value
// // the initial value is set from the environment and you can call Set() anytime
// // to update this value
// logLevel.Set(getLogLevelFromEnv())
//
// logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
//     Level: &logLevel,
// }))
//
// jsonLogger := slog.New(slog.NewJSONHandler(os.Stdout,nil))
// textLogger := slog.New(slog.NewTextHandler(os.Stdout,nil))
//
// jsonLogger.Info("database connected","db_host","localhost","port",5432)
// textLogger.Info("database connected","db_host","localhost","port",5432)
//
// // 	While source information is handy to have, it comes with a performance penalty because slog must call
// // runtime.Caller()
// //  to get the source code information, so keep that in mind.
// opts := &slog.HandlerOptions{
//     AddSource: true,
// }
// logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
//
// logger.Warn("storage space is low")
//
// logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
// if err != nil {
//     panic(err)
// }
//
// defer logFile.Close()
//
// logger := slog.New(slog.NewJSONHandler(logFile, nil))
//
// logger.Info("Starting server...", "port", 8080)
// logger.Warn("Storage space is low", "remaining_gb", 15)
// logger.Error("Database connection failed", "db_host", "10.0.0.5")
//
// }

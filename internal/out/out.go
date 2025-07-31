// Package out uses the slog and tint packages for the application logs.
// There are two logging modes, development and production.
// The production mode saves the logs to file and automatically rotates
// older files. While the development mode prints all feedback to stdout.
package out

import (
	"context"
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

// Color returns true when the w io writer is an os.file type
// and its file descriptor is a terminal.
func Color(w io.Writer) bool {
	if w == nil {
		return false
	}
	if _, ok := w.(*os.File); ok {
		return isatty.IsTerminal(w.(*os.File).Fd())
	}
	return false
}

// Default is the recommended logger that returns ...
func Default(w io.Writer) *slog.Logger {
	w, opts := defaultOptions(w)
	sl := slog.New(tint.NewHandler(w, &opts))
	return sl
}

// Discard all logger output.
func Discard() *slog.Logger {
	sl := slog.New(slog.DiscardHandler)
	return sl
}

// Quiet is intended to be used with the Confg.Quiet configuration.
// It only returns error and fatal log levels and does not display
// any source file information.
func Quiet(w io.Writer) *slog.Logger {
	w, opts := defaultOptions(w)
	sl := slog.New(tint.NewHandler(w, &opts))
	return sl
}

func defaultOptions(w io.Writer) (io.Writer, tint.Options) {
	if w == nil {
		w = os.Stdout
	}
	return w, tint.Options{
		AddSource: false,
		Level:     slog.LevelInfo,
		// ReplaceAttr: startAttr,
		TimeFormat: time.Kitchen,
		NoColor:    !Color(w),
	}
}

// Start is used to print the configurations and loading information
// during the launch of the application. Unlike a normal logger it has
// custom formatting for certain keys and their values used in the
// configuration environment variables and displayed during the
// server startup.
//
// The w io.Writer can usually be left as nil, as the standard output will
// be used with automatic color detection.
func Start(w io.Writer) *slog.Logger {
	w, opts := startOptions(w)
	sl := slog.New(tint.NewHandler(w, &opts))
	return sl
}

// StartCustom runs [Start] using the provided configurations.
//   - quiet sets the log level to error
//   - prod also saves the output to a file
func StartCustom(w io.Writer, quiet, prod bool) *slog.Logger {
	w, opts := startOptions(w)
	if quiet {
		opts.Level = LevelError
	}
	if prod {
		// TODO:
		//
		// w := io.MultiWriter(&buf1, &buf2)
	}
	sl := slog.New(tint.NewHandler(w, &opts))
	return sl
}

// StartDebug runs [Start].
//
// However it has
//   - color output disabled
//   - the log level is set to debug
//   - add source filename is enabled
func StartDebug(w io.Writer) *slog.Logger {
	w, opts := startOptions(w)
	opts.NoColor = true
	opts.AddSource = true
	opts.Level = LevelDebug
	opts.ReplaceAttr = func(groups []string, attr slog.Attr) slog.Attr {
		attr = addsourceNoDirectory(attr)
		return attr
	}
	sl := slog.New(tint.NewHandler(w, &opts))
	return sl
}

// StartMono runs [Start], however it has color output disabled.
func StartMono(w io.Writer) *slog.Logger {
	w, opts := startOptions(w)
	opts.NoColor = true
	sl := slog.New(tint.NewHandler(w, &opts))
	return sl
}

// startOptions returns the options and a writer for a text logger handler.
//
// The following are configured:
//   - add source is false
//   - log level is set to info
//   - time format is set to kitchen aka 08:43PM
//   - nocolor is automatically determined by the value of w
//
// If the given w io writer is nil then the stdout is returned.
func startOptions(w io.Writer) (io.Writer, tint.Options) {
	if w == nil {
		w = os.Stdout
	}
	return w, tint.Options{
		AddSource:   false,
		Level:       slog.LevelInfo,
		ReplaceAttr: startAttr,
		TimeFormat:  time.Kitchen,
		NoColor:     !Color(w),
	}
}

func startAttr(groups []string, a slog.Attr) slog.Attr {
	// fmt.Printf("\t%+v [%q]\n", a, a.Key)
	a = configUnsetAttr(a)
	switch strings.ToLower(a.Key) {
	case "":
		return slog.Attr{}
	case "postgres", "repair":
		a.Key = strings.ToUpper(a.Key)
	case "help":
		return configHelpAttr(a)
	case "issue":
		return configIssueAttr(a)
	case "msg":
		return configMsgAttr(a)
	}
	return a
}

func configUnsetAttr(a slog.Attr) slog.Attr {
	const unset = ",unset"
	if !strings.HasSuffix(a.Key, unset) {
		return a
	}
	a.Key = strings.TrimSuffix(a.Key, unset)
	return a
}

func configHelpAttr(a slog.Attr) slog.Attr {
	if a.Value.String() == "" {
		return slog.Attr{}
	}
	a.Key = "Help"
	return a
}

func configIssueAttr(a slog.Attr) slog.Attr {
	if a.Value.String() == "" {
		return slog.Attr{}
	}
	a.Key = strings.ToUpper(a.Key)
	a = tint.Attr(9, slog.String(a.Key, a.Value.String()))
	return tint.Attr(9, a)
}

func configMsgAttr(a slog.Attr) slog.Attr {
	switch strings.ToLower(a.Value.String()) {
	case "googleaccounts":
		return slog.Attr{}
	default:
		return a
	}
}

// Raw displays all logs to stdout but without timestamps.
// It is intended for use when troubleshooting.
func Raw() *slog.Logger {
	opts := slog.HandlerOptions{
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			a = timeformatRemove(groups, a)
			return a
		},
	}
	sl := slog.New(slog.NewTextHandler(os.Stdout, &opts))
	return sl
}

// addsourceNoDirectory removes the directory path from the log output
// when AddSource is true.
//
// For example, before and after.
//
//	`08:43PM /github/server/server.go:501 INF Log stuff`
//	`08:43PM server.go:501 INF Log stuff`
func addsourceNoDirectory(a slog.Attr) slog.Attr {
	// Remove the directory from the source's filename.
	if a.Key == slog.SourceKey {
		source := a.Value.Any().(*slog.Source)
		source.File = filepath.Base(source.File)
	}
	return a
}

// timeformatRemove removes the time key from the logs.
//
// For example, before and after.
//
//	`08:43PM INF Log stuff`
//	`INF Log stuff`
func timeformatRemove(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey && len(groups) == 0 {
		return slog.Attr{}
	}
	return a
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

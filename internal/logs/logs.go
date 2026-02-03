// Package logs uses the slog and tint packages for the application logs.
// There are two logging modes, development and production.
// The production mode saves the logs to file and automatically rotates
// older files. While the development mode prints all feedback to stdout.
package logs

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/panics"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

const (
	LevelDebug   = slog.LevelDebug
	LevelInfo    = slog.LevelInfo
	LevelWarning = slog.LevelWarn
	LevelError   = slog.LevelError
	LevelFatal   = slog.Level(12) // A more serious ERROR that aborts the application.
)

const (
	Ldate      = 1 << iota // the date as year/month/day: 2009/01/23
	Ltime                  // the time using US syntax: 1:23PM
	Lseconds               // the time using 24-hour syntax: 13:23:23. overrides Ltime
	Llongfile              // full file name and line number: /a/b/c/d.go:23
	Lshortfile             // final file name element and line number: d.go:23. overrides Llongfile
	Lcolor                 // use color output by providing ansi escape codes
	Lstdout                // output to the standard output (stdout)
	Lstderr                // output to the standard error (stderr)
	FlagAttr               // an internal flag to toggle a custom output for the environment configurations
)

const (
	// Quiets flag for the config.Quiet toggle that outputs to stderr but without color or a source file.
	Quiets = Ltime | Lstderr
	// Defaults flag for the Default slog logger that outputs to stderr and includes the time and source filename.
	Defaults = Lcolor | Lseconds | Lshortfile | Lstderr
	// Configurations flag for the Config.Print method to style the log output with a time but no source file.
	Configurations = Lcolor | Lstdout | Ltime | FlagAttr
	// Flags flag for the flags package to style the log output without the date and time.
	Flags = Lcolor | Lstdout | FlagAttr

	ErrorRed    = 9
	FatalRed    = 196 // #ff0000
	DebugPurple = 206 // #ff5fff
)

// Default is a general logger intended for use before the configurations
// environment variables have been read and parsed.
//
// It prints all log levels including debugging to the stdout and
// does not write to any files.
func Default() *slog.Logger {
	lf := NoFiles()
	sl := lf.New(LevelDebug, Defaults)
	return sl
}

// Discard is a logger intended for use with tests and discards all log output.
func Discard() *slog.Logger {
	sl := slog.New(slog.DiscardHandler)
	return sl
}

// Fatal logs any issues and exits to the operating system.
func Fatal(sl *slog.Logger, msg string, args ...slog.Attr) {
	FatalTx(context.Background(), sl, msg, args...)
}

// FatalTx logs any issues and exits to the operating system.
func FatalTx(ctx context.Context, sl *slog.Logger, msg string, args ...slog.Attr) {
	if sl == nil {
		panic(fmt.Errorf("fatal logger: %w", panics.ErrNoSlog))
	}
	sl.LogAttrs(ctx, LevelFatal, msg, args...)
	os.Exit(1)
}

// Color returns true when the w io writer is an os.file type
// and its file descriptor is a terminal.
func Color(w io.Writer) bool {
	if w == nil {
		return false
	}
	if descriptor, ok := w.(*os.File); ok {
		return isatty.IsTerminal(descriptor.Fd())
	}
	return false
}

// tintOptions applies the flag toggles and rewrites the slog attributes before they're returned.
//
// tint is a package that colors the log output.
func tintOptions(minimum slog.Level, flag int) tint.Options {
	attr := func(groups []string, a slog.Attr) slog.Attr {
		if flag&(Lshortfile) != 0 {
			a = addsourceNoDirectory(a)
		}
		if flag&(Ldate) == 0 && flag&(Ltime) == 0 && flag&(Lseconds) == 0 {
			a = timeformatRemove(groups, a)
		}
		if flag&(FlagAttr) != 0 {
			return flagAttr(a)
		}
		a = replaceAttr(a)
		return a
	}
	return tint.Options{
		AddSource:   addsource(flag),
		Level:       minimum,
		ReplaceAttr: attr,
		TimeFormat:  timeformat(flag),
		NoColor:     nocolor(flag),
	}
}

// replaceAttr formats specific keys and values for readability.
func replaceAttr(a slog.Attr) slog.Attr {
	key := strings.ToLower(a.Key)
	switch key {
	case "":
		return slog.Attr{}
	case "help", "problem":
		a.Key = helper.Capitalize(a.Key)
	case "error":
		a.Key = strings.ToUpper(a.Key)
		val := a.Value.Any()
		if err, ok := val.(error); ok {
			a = tint.Attr(ErrorRed, slog.String(a.Key, err.Error()))
		}
	case "postgres":
		a.Key = "PostgreSQL"
	}
	a = levelAttr(a)
	return a
}

// levelAttr formats the custom log levels.
// It also rewrites the debug level to give a greater emphasis of its verboseness.
func levelAttr(a slog.Attr) slog.Attr {
	// Format the custom level keys to use color
	if a.Key == slog.LevelKey {
		level, ok := a.Value.Any().(slog.Level)
		if !ok {
			return a
		}
		switch level {
		case LevelDebug:
			a = tint.Attr(DebugPurple, slog.String(a.Key, "debug"))
		case LevelFatal:
			a = tint.Attr(FatalRed, slog.String(a.Key, "FATAL"))
		}
	}
	return a
}

// addsource returns true if the AddSource Option is flagged for use.
func addsource(flag int) bool {
	return flag&(Llongfile|Lshortfile) != 0
}

// nocolor returns true if the NoColor Option is flagged for use.
func nocolor(flag int) bool {
	return flag&(Lcolor) == 0
}

// timeformat customizes the date and time of the log output based on the flag sets.
func timeformat(flag int) string {
	if flag&(Ldate) != 0 {
		if flag&(Lseconds) != 0 {
			return "2006-01-02 15:04:05"
		}
		if flag&(Ltime) != 0 {
			return "2006-01-02 15:04"
		}
		return "2006-01-02"
	}
	if flag&(Lseconds) != 0 {
		return "15:04:05"
	}
	if flag&(Ltime) != 0 {
		return "15:04"
	}
	return ""
}

// flagAttr formats the keys and values used by the config.Config struct.
func flagAttr(a slog.Attr) slog.Attr {
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

// configUnsetAttr drops the ',unset' suffix found in some keys.
func configUnsetAttr(a slog.Attr) slog.Attr {
	const unset = ",unset"
	if trimmed, ok := strings.CutSuffix(a.Key, unset); ok {
		a.Key = trimmed
	}
	return a
}

func configHelpAttr(a slog.Attr) slog.Attr {
	if a.Value.String() == "" {
		return slog.Attr{}
	}
	a.Key = "Help"
	return a
}

// configIssueAttr applies formatting for the optional "issue" key.
func configIssueAttr(a slog.Attr) slog.Attr {
	if a.Value.String() == "" {
		return slog.Attr{}
	}
	a.Key = strings.ToUpper(a.Key)
	return tint.Attr(ErrorRed, slog.String(a.Key, a.Value.String()))
}

// configMsgAttr drops values that are not intended for logging.
func configMsgAttr(a slog.Attr) slog.Attr {
	switch strings.ToLower(a.Value.String()) {
	case "googleaccounts":
		return slog.Attr{}
	default:
		return a
	}
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
		source, ok := a.Value.Any().(*slog.Source)
		if !ok {
			return a
		}
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

// Package out uses the slog and tint packages for the application logs.
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

	FatalRed    = 196 // #ff0000
	DebugPurple = 206 // #ff5fff
)

const (
	Ldate      = 1 << iota // the date in the local time zone: 2009/01/23
	Ltime                  // the time in the local time zone: 1:23AM
	Lseconds               // second resolution: 01:23:23.  assumes Ltime.
	Llongfile              // full file name and line number: /a/b/c/d.go:23
	Lshortfile             // final file name element and line number: d.go:23. overrides Llongfile
	LUTC                   // if Ldate or Ltime is set, use UTC rather than the local time zone
	Lcolor
	Lstdout
	Lstderr
	FlagAttr
)

const (
	Quiets         = Ltime | Lstderr
	Defaults       = Lcolor | Lseconds | Lshortfile | Lstderr
	Configurations = Lcolor | Lstdout | Ltime | FlagAttr
	Flags          = Lcolor | Lstdout | FlagAttr
)

// Default is a general logger intended for use before the configurations
// environment variables have been read and parsed.
func Default() *slog.Logger {
	sl := New(LevelDebug, nil, Defaults)
	return sl
}

// Discard is a logger intended for use with tests and discards all log output.
func Discard() *slog.Logger {
	sl := slog.New(slog.DiscardHandler)
	return sl
}

// New creates a new slog logger.
//
// The logf LogFile is an optional opened log file writer or can be
// set to nil to leave unused.
//
// The flag provide customizations including the options Lstdout
// and Lstderr which are not set. For terminal output, at least one
// of these must be provided.
func New(level slog.Level, logf *LogFile, flag int) *slog.Logger {
	sl := slog.New(slog.DiscardHandler)
	if flag&(Lstdout) == 0 && flag&(Lstderr) == 0 && logf == nil {
		return sl
	}
	w := writers(logf, flag)
	opts := tintOptions(level, flag)
	sl = slog.New(tint.NewHandler(w, &opts))
	return sl
}

// Fatal logs any issues and exits to the operating system.
func Fatal(sl *slog.Logger, msg string, args ...slog.Attr) {
	if sl == nil {
		panic(fmt.Errorf("fatal logger: %w", panics.ErrNoSlog))
	}
	sl.LogAttrs(context.Background(), LevelFatal, msg, args...)
	os.Exit(1)
}

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

func writers(logf *LogFile, flag int) io.Writer {
	if logf != nil {
		switch {
		case flag&(Lstderr) != 0 && flag&(Lstdout) != 0:
			return io.MultiWriter(os.Stderr, os.Stdout, logf.file)
		case flag&(Lstderr) != 0:
			return io.MultiWriter(os.Stderr, logf.file)
		case flag&(Lstdout) != 0:
			return io.MultiWriter(os.Stdout, logf.file)
		default:
			return logf.file
		}
	}
	switch {
	case flag&(Lstderr) != 0 && flag&(Lstdout) != 0:
		return io.MultiWriter(os.Stderr, os.Stdout)
	case flag&(Lstderr) != 0:
		return os.Stderr
	case flag&(Lstdout) != 0:
		return os.Stdout
	default:
		return io.Discard
	}
}

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

func replaceAttr(a slog.Attr) slog.Attr {
	switch strings.ToLower(a.Key) {
	case "":
		return slog.Attr{}
	case "help", "problem":
		a.Key = helper.Capitalize(a.Key)
	case "error":
		a.Key = strings.ToUpper(a.Key)
		val := a.Value.Any()
		if err, ok := val.(error); ok {
			a = tint.Attr(9, slog.String(a.Key, err.Error()))
		}
	case "postgres":
		a.Key = "PostgreSQL"
	}
	a = levelAttr(a)
	return a
}

func levelAttr(a slog.Attr) slog.Attr {
	// Format the custom level keys to use color
	if a.Key == slog.LevelKey {
		level := a.Value.Any().(slog.Level)
		switch level {
		case LevelDebug:
			a = tint.Attr(DebugPurple, slog.String(a.Key, "debug"))
		case LevelFatal:
			a = tint.Attr(FatalRed, slog.String(a.Key, "FATAL"))
		}
	}
	return a
}

func addsource(flag int) bool {
	return flag&(Llongfile|Lshortfile) != 0
}

func nocolor(flag int) bool {
	return flag&(Lcolor) == 0
}

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

//	func DefaultMore(w io.Writer, level *slog.LevelVar) *slog.Logger {
//		buildInfo, _ := debug.ReadBuildInfo()
//		sl := Default(w, level)
//		child := sl.With(
//			slog.Group("program",
//				slog.Int("pid", os.Getpid()),
//				slog.String("go version", buildInfo.GoVersion)))
//		return child
//	}

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

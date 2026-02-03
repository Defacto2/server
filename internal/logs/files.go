package logs

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	slogmulti "github.com/samber/slog-multi"
)

const (
	// NameErr is the filename of the Error and Fatal levels log.
	NameErr = pname + "_error.json.log"
	// NameInfo is the filename of the Warning and Info level log.
	NameInfo = pname + "_info.json.log"
	// NameDebug is the filename of the Debug level log.
	NameDebug = pname + "_debug.json.log"
	pname     = "defacto2_serve" // prefix name for log files
)

// Files is used to write log files to multiple locations in parallel.
// The primary use of this is to log the different error types into
// separate files and to also permit writing to files and the terminal
// at the same time.
type Files struct {
	errlevel   *os.File // for fatal and error levels
	infolevel  *os.File // for warn and info levels
	debuglevel *os.File // for debug level
}

// Close the open file descriptors in use by Files.
// Any errors will be joined and returned.
func (f Files) Close() error {
	const msg = "logs files close"
	var errs []error
	if f.errlevel != nil {
		if err := f.errlevel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("error level: %w", err))
		}
	}
	if f.infolevel != nil {
		if err := f.infolevel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("info level: %w", err))
		}
	}
	if f.debuglevel != nil {
		if err := f.debuglevel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("debug level: %w", err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s: %w", msg, errors.Join(errs...))
	}
	return nil
}

// New creates a slog logger that can write to multiple writers.
// The stdmin slog level can be used to limit the minimum log level
// for the stdout and stderr loggers.
// The flag int is used to configure the stdout and stderr logging.
//
// Opened files descriptors can be provided using the Files struct
// and should be closed after use. Multiple open files can be used
// for different log levels. For example, you can use the
// Files.errolevel descriptor to save only errors and fatal issues.
// While using infolevel to also log info, warning, errors and fatal
// to another file.
//
// File descriptors ignore the provided stdmin slog level.
func (f Files) New(stdmin slog.Level, flag int) *slog.Logger {
	useStdout, useStderr := flag&(Lstdout) != 0, flag&(Lstderr) != 0
	if f.errlevel == nil && f.infolevel == nil && f.debuglevel == nil && !useStderr && !useStdout {
		return Discard()
	}
	handlers := make([]slog.Handler, 0, 5)
	if f.errlevel != nil {
		handlers = append(handlers, slog.NewJSONHandler(f.errlevel, &slog.HandlerOptions{
			Level: LevelError, AddSource: true,
		}))
	}
	if f.infolevel != nil {
		handlers = append(handlers, slog.NewJSONHandler(f.infolevel, &slog.HandlerOptions{
			Level: LevelInfo, AddSource: false,
		}))
	}
	if f.debuglevel != nil {
		handlers = append(handlers, slog.NewJSONHandler(f.debuglevel, &slog.HandlerOptions{
			Level: LevelDebug, AddSource: false,
		}))
	}
	if !useStdout && !useStderr {
		sl := slog.New(slogmulti.Fanout(handlers...))
		return sl
	}
	opts := tintOptions(stdmin, flag)
	if useStdout {
		handlers = append(handlers, tint.NewHandler(os.Stdout, &opts))
	}
	if useStderr {
		handlers = append(handlers, tint.NewHandler(os.Stderr, &opts))
	}
	sl := slog.New(slogmulti.Fanout(handlers...))
	return sl
}

// NoFiles returns an empty Files struct and is available to show usage and intention.
func NoFiles() Files {
	return Files{}
}

// OpenFiles creates or opens the named log files for use with the
// [Files.New] method. Multiple files can be opened together and all
// files must closed a after use using the [Files.Close] method.
//
//   - ename will be used to write fatal and error reports.
//   - iname will be used to write fatal, error, warnings and info reports.
//   - dname will be used to write all reports including debug level reports.
//
// The root should be the named directory to store the logs. If root is left empty
// the home directory of the user account will be used.
//
// If any errors occur they will be returned as a wrapped error and
// must be handled appropriately.
func OpenFiles(root string, ename, iname, dname string) (Files, error) {
	const msg = "logs open file"
	const flag = os.O_CREATE | os.O_APPEND | os.O_WRONLY
	const perm = 0o666
	none, files := Files{}, Files{}
	// handle the root directory
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return none, fmt.Errorf("%s user home dir: %w", msg, err)
		}
		root = home
	}
	r, err := os.OpenRoot(root)
	if err != nil {
		return none, fmt.Errorf("%s open root: %w", msg, err)
	}
	defer r.Close()
	// open files
	var errr error
	if ename != "" {
		files.errlevel, errr = r.OpenFile(ename, flag, perm)
	}
	var erri error
	if iname != "" {
		files.infolevel, erri = r.OpenFile(iname, flag, perm)
	}
	var errd error
	if dname != "" {
		files.debuglevel, errd = r.OpenFile(dname, flag, perm)
	}
	err = errors.Join(errr, erri, errd)
	if err != nil {
		files.Close()
		return none, fmt.Errorf("%s: %w", msg, err)
	}
	return files, nil
}

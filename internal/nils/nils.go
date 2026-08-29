// Package nils provides checks to help avoid panics caused by named functions
// using empty arguments, or nil pointers.
//
//nolint:exhaustive
package nils

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"mime/multipart"
	"reflect"

	"github.com/Defacto2/server/handler/fulltext"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/labstack/echo/v5"
)

var (
	ErrArgument    = errors.New("nil argument provided")
	ErrBoil        = errors.New("boil context executor is nil or invalid")
	ErrContext     = errors.New("ctx context is nil or invalid")
	ErrEchoHandler = errors.New("echo handler func is nil or invalid")
	ErrEchoRoutes  = errors.New("echo routes is nil or invalid")
	ErrEmbedFS     = errors.New("embed file system instance is empty")
	ErrFSFS        = errors.New("file system instance is empty")

	ErrBytesBuf     = errors.New("bytes buffer pointer is nil")
	ErrEchoContext  = errors.New("echo context pointer is nil")
	ErrEchoEcho     = errors.New("echo echo pointer is nil") //nolint:dupword
	ErrEchoGroup    = errors.New("echo group pointer is nil")
	ErrFileHeader   = errors.New("multipart file header pointer is nil")
	ErrModels       = errors.New("models file pointer is nil")
	ErrModelSummary = errors.New("model summary pointer is nil")
	ErrSQLDB        = errors.New("sql db database pointer is nil")
	ErrSQLTx        = errors.New("sql tx transaction pointer is nil")
	ErrSlogLogger   = errors.New("slog logger pointer is nil")

	ErrPointer = errors.New("pointer is nil")
)

const format = "argument %d: %w"

// Check looks for any nil pointers or values and returns an error if found.
// This is intended for the arguments of named functions and methods.
// And importantly to avoid unexpected panics caused by nil pointers.
func Check(args ...any) error {
	for n, arg := range args {
		if !IsNil(arg) {
			continue
		}
		if arg == nil {
			return fmt.Errorf(format, n, ErrArgument)
		}

		argType := reflect.TypeOf(arg)
		if err := exactInterface(n, argType); err != nil {
			return err
		}
		return pointers(n, argType)
	}

	return nil
}

// exactInterface match.
func exactInterface(n int, argType reflect.Type) error {
	switch {
	case argType.Implements(reflect.TypeFor[context.Context]()):
		return fmt.Errorf(format, n, ErrContext)
	case argType.Implements(reflect.TypeFor[boil.ContextExecutor]()):
		return fmt.Errorf(format, n, ErrBoil)
	}

	return nil
}

// pointers, interfaces, concrete types.
func pointers(n int, argType reflect.Type) error {
	switch argType {
	case reflect.TypeFor[echo.HandlerFunc]():
		return fmt.Errorf(format, n, ErrEchoHandler)
	case reflect.TypeFor[echo.Routes]():
		return fmt.Errorf(format, n, ErrEchoRoutes)
	case reflect.TypeFor[embed.FS]():
		return fmt.Errorf(format, n, ErrEmbedFS)
	case reflect.TypeFor[fs.FS]():
		return fmt.Errorf(format, n, ErrFSFS)
	case reflect.TypeFor[*slog.Logger]():
		return fmt.Errorf(format, n, ErrSlogLogger)
	case reflect.TypeFor[*sql.DB]():
		return fmt.Errorf(format, n, ErrSQLDB)
	case reflect.TypeFor[*sql.Tx]():
		return fmt.Errorf(format, n, ErrSQLTx)
	case reflect.TypeFor[*echo.Context]():
		return fmt.Errorf(format, n, ErrEchoContext)
	case reflect.TypeFor[*echo.Echo]():
		return fmt.Errorf(format, n, ErrEchoEcho)
	case reflect.TypeFor[*echo.Group]():
		return fmt.Errorf(format, n, ErrEchoGroup)
	case reflect.TypeFor[*models.File]():
		return fmt.Errorf(format, n, ErrModels)
	case reflect.TypeFor[*bytes.Buffer]():
		return fmt.Errorf(format, n, ErrBytesBuf)
	case reflect.TypeFor[*multipart.FileHeader]():
		return fmt.Errorf(format, n, ErrFileHeader)
	case reflect.TypeFor[*fulltext.Tidbits]():
		return fmt.Errorf(format, n, ErrPointer)
	default:
		const fallback = "argument %d using type %s: %w"
		return fmt.Errorf(fallback, n, argType, ErrArgument)
	}
}

// Slog looks for any nil pointers or values, and if found,
// logs an error using the [log.slog] package.
// This is intended for the arguments of named functions and methods that
// do not return errors.
//
// False is returned when all the args are valid.
func Slog(msg string, args ...any) bool {
	err := Check(args...)
	if err == nil {
		return false
	}

	var sl *slog.Logger
	for _, arg := range args {
		if logger, ok := arg.(*slog.Logger); ok && logger != nil {
			sl = logger
			break
		}
	}

	if sl == nil {
		sl = slog.Default()
	}
	sl.Error(msg, slog.Any("check", err))

	return true
}

// IsNil reports whether v is nil or a typed nil.
func IsNil(v any) bool {
	if v == nil {
		return true
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case
		reflect.Chan,
		reflect.Func,
		reflect.Map,
		reflect.Pointer,
		reflect.UnsafePointer,
		reflect.Interface,
		reflect.Slice:
		return val.IsNil()
	default:
		return false
	}
}

// BoilExec returns true if the database boil package, context executor is invalid.
func BoilExec(exec boil.ContextExecutor) bool {
	v := reflect.ValueOf(exec)
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			return true
		}
		return false
	}

	return true
}

// BoilExecCrash panics if the database boil package, context extractor is invalid.
// This is a lazy fallback function intended for the model packages
// to reduce programming boilerplate, however its use should generally be avoided.
func BoilExecCrash(exec boil.ContextExecutor) {
	if BoilExec(exec) {
		panic(ErrBoil)
	}
}

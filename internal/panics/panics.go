// Package panics provides checks to help avoid nil pointer panics
// caused by missing arguments provided to some funcs.
package panics

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"log/slog"
	"reflect"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/labstack/echo/v4"
)

var (
	ErrNoArtM    = errors.New("art models file is nil")
	ErrNoBoil    = errors.New("exec boil context executor is nil or invalid")
	ErrNoBuffer  = errors.New("bytes buffer pointer is nil")
	ErrNoContext = errors.New("ctx context is nil")
	ErrNoDB      = errors.New("db database pointer is nil")
	ErrNoEchoE   = errors.New("e echo pointer is nil")
	ErrNoEchoC   = errors.New("c echo context pointer is nil")
	ErrNoEmbed   = errors.New("embed file system instance is empty")
	ErrNoGroup   = errors.New("g echo group pointer is nil")
	ErrNoSlog    = errors.New("sl slog logger pointer is nil")
	ErrNoTx      = errors.New("tx transaction pointer is nil")
)

// BoilExec returns true if the database context executor is invalid such as nil.
func BoilExec(exec boil.ContextExecutor) bool {
	v := reflect.ValueOf(exec)
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return true
		}
		return false
	}
	return true
}

// BoilExecCrash panics if the exec boil context extractor is invalid.
// This is a fallback function intended for the model packages to reduce
// programmign boilerplate by requiring only the function without conditional statements.
func BoilExecCrash(exec boil.ContextExecutor) {
	if BoilExec(exec) {
		panic(ErrNoBoil)
	}
}

func GroupD(g *echo.Group, db *sql.DB) error {
	if g == nil {
		return ErrNoGroup
	}
	if db == nil {
		return ErrNoDB
	}
	return nil
}

func GroupDS(g *echo.Group, db *sql.DB, sl *slog.Logger) error {
	if g == nil {
		return ErrNoGroup
	}
	if db == nil {
		return ErrNoDB
	}
	if sl == nil {
		return ErrNoSlog
	}
	return nil
}

func EchoContextD(c echo.Context, db *sql.DB) error {
	if c == nil {
		return ErrNoEchoC
	}
	if db == nil {
		return ErrNoDB
	}
	return nil
}

func EchoContextS(c echo.Context, sl *slog.Logger) error {
	if c == nil {
		return ErrNoEchoC
	}
	if sl == nil {
		return ErrNoSlog
	}
	return nil
}

func EchoContextDS(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	if c == nil {
		return ErrNoEchoC
	}
	if db == nil {
		return ErrNoDB
	}
	if sl == nil {
		return ErrNoSlog
	}
	return nil
}

func ContextT(ctx context.Context, tx *sql.Tx) error {
	if ctx == nil {
		return ErrNoContext
	}
	if tx == nil {
		return ErrNoTx
	}
	return nil
}

func ContextB(ctx context.Context, exec boil.ContextExecutor) error {
	if ctx == nil {
		return ErrNoContext
	}
	if exec == nil {
		return ErrNoBoil
	}
	return nil
}

func ContextBS(ctx context.Context, exec boil.ContextExecutor, sl *slog.Logger) error {
	if ctx == nil {
		return ErrNoContext
	}
	if exec == nil {
		return ErrNoBoil
	}
	if sl == nil {
		return ErrNoSlog
	}
	return nil
}

func ContextDS(ctx context.Context, db *sql.DB, sl *slog.Logger) error {
	if ctx == nil {
		return ErrNoContext
	}
	if db == nil {
		return ErrNoDB
	}
	if sl == nil {
		return ErrNoSlog
	}
	return nil
}

func ContextDTS(ctx context.Context, db *sql.DB, tx *sql.Tx, sl *slog.Logger) error {
	if ctx == nil {
		return ErrNoContext
	}
	if db == nil {
		return ErrNoDB
	}
	if tx == nil {
		return ErrNoTx
	}
	if sl == nil {
		return ErrNoSlog
	}
	return nil
}

func ContextD(ctx context.Context, db *sql.DB) error {
	if ctx == nil {
		return ErrNoContext
	}
	if db == nil {
		return ErrNoDB
	}
	return nil
}

func ContextS(ctx context.Context, sl *slog.Logger) error {
	if ctx == nil {
		return ErrNoContext
	}
	if sl == nil {
		return ErrNoSlog
	}
	return nil
}

// EchoDSP checks the arguments for handler package.
// If an error is returned, the calling method or func should abort and return the error.
func EchoDSP(e *echo.Echo, db *sql.DB, sl *slog.Logger, public embed.FS) error {
	if e == nil {
		return ErrNoEchoE
	}
	if db == nil {
		return ErrNoDB
	}
	if sl == nil {
		return ErrNoSlog
	}
	if public == (embed.FS{}) {
		return ErrNoEmbed
	}
	return nil
}

// EchoP checks the arguments for handler package.
// If an error is returned, the calling method or func should abort and return the error.
func EchoP(e *echo.Echo, public embed.FS) error {
	if e == nil {
		return ErrNoEchoE
	}
	if public == (embed.FS{}) {
		return ErrNoEmbed
	}
	return nil
}

// EchoDS checks the arguments for handler package.
// If an error is returned, the calling method or func should abort and return the error.
func EchoDS(e *echo.Echo, db *sql.DB, sl *slog.Logger) error {
	if e == nil {
		return ErrNoEchoE
	}
	if db == nil {
		return ErrNoDB
	}
	if sl == nil {
		return ErrNoSlog
	}
	return nil
}

// EchoS checks the arguments for handler package.
// If an error is returned, the calling method or func should abort and return the error.
func EchoS(e *echo.Echo, sl *slog.Logger) error {
	if e == nil {
		return ErrNoEchoE
	}
	if sl == nil {
		return ErrNoSlog
	}
	return nil
}

func DS(db *sql.DB, sl *slog.Logger) error {
	if db == nil {
		return ErrNoDB
	}
	if sl == nil {
		return ErrNoSlog
	}
	return nil
}

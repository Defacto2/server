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
	"github.com/labstack/echo/v5"
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
	switch v.Kind() { //nolint:exhaustive
	case reflect.Pointer, reflect.Interface:
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

func CD(ctx context.Context, db *sql.DB) error {
	if ctx == nil {
		return ErrNoContext
	}
	if db == nil {
		return ErrNoDB
	}
	return nil
}

func CE(ctx context.Context, exec boil.ContextExecutor) error {
	if ctx == nil {
		return ErrNoContext
	}
	if exec == nil {
		return ErrNoBoil
	}
	return nil
}

func CS(ctx context.Context, sl *slog.Logger) error {
	if ctx == nil {
		return ErrNoContext
	}
	if sl == nil {
		return ErrNoSlog
	}
	return nil
}

func CSD(ctx context.Context, sl *slog.Logger, db *sql.DB) error {
	if ctx == nil {
		return ErrNoContext
	}
	if sl == nil {
		return ErrNoSlog
	}
	if db == nil {
		return ErrNoDB
	}
	return nil
}

func CSDTx(ctx context.Context, sl *slog.Logger, db *sql.DB, tx *sql.Tx) error {
	if ctx == nil {
		return ErrNoContext
	}
	if sl == nil {
		return ErrNoSlog
	}
	if db == nil {
		return ErrNoDB
	}
	if tx == nil {
		return ErrNoTx
	}
	return nil
}

func CSE(ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor) error {
	if ctx == nil {
		return ErrNoContext
	}
	if sl == nil {
		return ErrNoSlog
	}
	if exec == nil {
		return ErrNoBoil
	}
	return nil
}

func CTx(ctx context.Context, tx *sql.Tx) error {
	if ctx == nil {
		return ErrNoContext
	}
	if tx == nil {
		return ErrNoTx
	}
	return nil
}

func ECD(c *echo.Context, db *sql.DB) error {
	if c == nil {
		return ErrNoEchoC
	}
	if db == nil {
		return ErrNoDB
	}
	return nil
}

func EP(e *echo.Echo, public embed.FS) error {
	if e == nil {
		return ErrNoEchoE
	}
	if public == (embed.FS{}) {
		return ErrNoEmbed
	}
	return nil
}

func GD(g *echo.Group, db *sql.DB) error {
	if g == nil {
		return ErrNoGroup
	}
	if db == nil {
		return ErrNoDB
	}
	return nil
}

func SC(c *echo.Context, sl *slog.Logger) error {
	// leave these incorrectly ordered arguments unchanged for the moment
	if sl == nil {
		return ErrNoSlog
	}
	if c == nil {
		return ErrNoEchoC
	}
	return nil
}

func SCD(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	if sl == nil {
		return ErrNoSlog
	}
	if c == nil {
		return ErrNoEchoC
	}
	if db == nil {
		return ErrNoDB
	}
	return nil
}

func SD(sl *slog.Logger, db *sql.DB) error {
	if sl == nil {
		return ErrNoSlog
	}
	if db == nil {
		return ErrNoDB
	}
	return nil
}

func SDE(sl *slog.Logger, db *sql.DB, e *echo.Echo) error {
	if sl == nil {
		return ErrNoSlog
	}
	if db == nil {
		return ErrNoDB
	}
	if e == nil {
		return ErrNoEchoE
	}
	return nil
}

func SDEP(sl *slog.Logger, db *sql.DB, e *echo.Echo, public embed.FS) error {
	if sl == nil {
		return ErrNoSlog
	}
	if db == nil {
		return ErrNoDB
	}
	if e == nil {
		return ErrNoEchoE
	}
	if public == (embed.FS{}) {
		return ErrNoEmbed
	}
	return nil
}

func SE(sl *slog.Logger, e *echo.Echo) error {
	if sl == nil {
		return ErrNoSlog
	}
	if e == nil {
		return ErrNoEchoE
	}
	return nil
}

func SGD(sl *slog.Logger, g *echo.Group, db *sql.DB) error {
	if sl == nil {
		return ErrNoSlog
	}
	if g == nil {
		return ErrNoGroup
	}
	if db == nil {
		return ErrNoDB
	}
	return nil
}

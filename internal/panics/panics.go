// Package panics provides checks to help avoid nil pointer panics
// caused by missing arguments provided to some funcs.
package panics

import (
	"database/sql"
	"embed"
	"errors"
	"log/slog"

	"github.com/labstack/echo/v4"
)

var (
	ErrNoDB    = errors.New("db database pointer is nil")
	ErrNoEchoE = errors.New("e echo pointer is nil")
	ErrNoEchoC = errors.New("c echo context pointer is nil")
	ErrNoEmbed = errors.New("embed file system instance is empty")
	ErrNoSlog  = errors.New("sl slog logger pointer is nil")
)

func Db(c echo.Context, db *sql.DB) error {
	if c == nil {
		return ErrNoEchoC
	}
	if db == nil {
		return ErrNoDB
	}
	return nil
}

func Dbslog(c echo.Context, db *sql.DB, sl *slog.Logger) error {
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

func Slog(c echo.Context, sl *slog.Logger) error {
	if c == nil {
		return ErrNoEchoC
	}
	if sl == nil {
		return ErrNoSlog
	}
	return nil
}

// EchoEmbed checks the arguments for handler package.
// If an error is returned, the calling method or func should abort and return the error.
func EchoEmbed(e *echo.Echo, db *sql.DB, sl *slog.Logger, public embed.FS) error {
	if e == nil {
		return ErrNoEchoE
	}
	if db == nil {
		return ErrNoDB
	}
	if sl == nil {
		return ErrNoSlog
	}
	var empty embed.FS
	if public == empty {
		return ErrNoEmbed
	}
	return nil
}

// EchoHtml checks the arguments for handler package.
// If an error is returned, the calling method or func should abort and return the error.
func EchoHtml(e *echo.Echo, public embed.FS) error {
	if e == nil {
		return ErrNoEchoE
	}
	var empty embed.FS
	if public == empty {
		return ErrNoEmbed
	}
	return nil
}

// EchoDbslog checks the arguments for handler package.
// If an error is returned, the calling method or func should abort and return the error.
func EchoDbslog(e *echo.Echo, db *sql.DB, sl *slog.Logger) error {
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

// EchoSlog checks the arguments for handler package.
// If an error is returned, the calling method or func should abort and return the error.
func EchoSlog(e *echo.Echo, sl *slog.Logger) error {
	if e == nil {
		return ErrNoEchoE
	}
	if sl == nil {
		return ErrNoSlog
	}
	return nil
}

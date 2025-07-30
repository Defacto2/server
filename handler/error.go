package handler

import (
	"database/sql"
	"embed"
	"errors"
	"log/slog"

	"github.com/labstack/echo/v4"
)

var (
	ErrName    = errors.New("name is empty")
	ErrName404 = errors.New("named template cannot be found")
	ErrPorts   = errors.New("the server ports are not configured")
	ErrRoutes  = errors.New("echo instance is nil")
	ErrZap     = errors.New("zap logger instance is nil")

	ErrNoDB    = errors.New("database pointer db is nil")
	ErrNoEcho  = errors.New("echo pointer e is nil")
	ErrNoEmbed = errors.New("embed file system instance is empty")
	ErrNoSlog  = errors.New("logger pointer sl is nil")
)

func embedpanic(e *echo.Echo, db *sql.DB, sl *slog.Logger, public embed.FS) error {
	if e == nil {
		return ErrNoEcho
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

func htmlpanic(e *echo.Echo, public embed.FS) error {
	if e == nil {
		return ErrNoEcho
	}
	var empty embed.FS
	if public == empty {
		return ErrNoEmbed
	}
	return nil
}

func dbpanic(e *echo.Echo, db *sql.DB, sl *slog.Logger) error {
	if e == nil {
		return ErrNoEcho
	}
	if db == nil {
		return ErrNoDB
	}
	if sl == nil {
		return ErrNoSlog
	}
	return nil
}

func slpanic(e *echo.Echo, sl *slog.Logger) error {
	if e == nil {
		return ErrNoEcho
	}
	if sl == nil {
		return ErrNoSlog
	}
	return nil
}

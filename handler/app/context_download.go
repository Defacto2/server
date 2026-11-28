package app

// Package file context_download.go contains handlers for binary artifact downloads,
// inline previews, checksums, and js-dos zip delivery.

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Defacto2/server/handler/download"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/nils"
	"github.com/labstack/echo/v5"
)

const (
	dx    = "d"
	jsdos = "jsdos"
	vx    = "v"
)

// Checksum is the handler for the Checksum file record page.
func Checksum(sl *slog.Logger, c *echo.Context, db *sql.DB, id string) error {
	const format = "checksum context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	const uri = "sum"

	ctx := c.Request().Context()
	if err := download.Checksum(ctx, c, db, id); err != nil {
		if errors.Is(err, download.ErrStat) {
			return FileMissingErr(sl, c, uri, err)
		}
		return DownloadErr(sl, c, uri, err)
	}
	return nil
}

// DownloadJsDos is the handler for the js-dos emulator to download zip files that are then
// mounted as a C: hard drive in the emulation. js-dos only supports common zip compression methods,
// so this func first attempts to offer a re-archived zip file found in the extra directory, and
// only if that fails does it offer the original download file.
func DownloadJsDos(sl *slog.Logger, c *echo.Context, db *sql.DB, extra, downl dir.Directory,
) error {
	const format = "download jsdos context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	e := download.ExtraZip{
		Extra:    extra,
		Download: downl,
	}

	ctx := c.Request().Context()
	if err := e.HTTPSend(ctx, c, db); err != nil {
		if errors.Is(err, download.ErrStat) {
			return FileMissingErr(sl, c, jsdos, err)
		}
		return DownloadErr(sl, c, jsdos, err)
	}
	return nil
}

// Download is the handler for downloading a file artifact as an attachment.
func Download(sl *slog.Logger, c *echo.Context, db *sql.DB, path dir.Directory) error {
	const format = "download context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	d := download.Download{
		Inline: false,
		Dir:    path,
	}

	if id := c.Param("id"); id != "" {
		r := c.Response()
		r.Header().Set("Link", `<https://defacto2.net/`+dx+`/`+id+`>; rel="canonical"`)
	}

	if err := d.HTTPSend(sl, c, db); err != nil {
		if errors.Is(err, download.ErrStat) {
			return FileMissingErr(sl, c, dx, err)
		}
		return DownloadErr(sl, c, dx, err)
	}
	return nil
}

// Inline is the handler for serving a file artifact inline in the browser.
func Inline(sl *slog.Logger, c *echo.Context, db *sql.DB, path dir.Directory) error {
	const format = "inline context: %w"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	d := download.Download{
		Inline: true,
		Dir:    path,
	}

	if err := d.HTTPSend(sl, c, db); err != nil {
		if errors.Is(err, download.ErrStat) {
			return FileMissingErr(sl, c, vx, err)
		}
		return DownloadErr(sl, c, vx, err)
	}
	return nil
}

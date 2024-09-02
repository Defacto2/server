// Package download handles the client file downloads.
package download

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/sess"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var (
	ErrNone = errors.New("not found")
	ErrStat = errors.New("file download stored on this server cannot be found")
)

// Checksum serves the checksums for the requested file.
// The response is a text file named "checksums.txt" with the checksum and filename.
// The id string is the UID filename of the requested file.
func Checksum(c echo.Context, db *sql.DB, id string) error {
	ctx := context.Background()
	art, err := model.OneFileByKey(ctx, db, id)
	if err != nil {
		if errors.Is(err, model.ErrDB) && sess.Editor(c) {
			art, err = model.OneEditByKey(ctx, db, id)
		}
		if err != nil {
			return fmt.Errorf("file download checksum %w: %s", err, id)
		}
	}
	// an example checksum file body created by `shasum`
	// 72f8a29d75993487b7ad5ad3a17d2f65ed4c41be155adbda88258d0458fcfe29f55e2e31b0316f01d57f4427ca9e2422  sk8-01.jpg
	sum := strings.TrimSpace(art.FileIntegrityStrong.String)
	if sum == "" {
		return fmt.Errorf("file download checksum %w: %d", ErrNone, art.ID)
	}
	name := art.Filename.String
	body := []byte(sum + " " + name)

	file, err := os.CreateTemp(helper.TmpDir(), "checksum-server.*.txt")
	if err != nil {
		return fmt.Errorf("file download checksum create tempdir: %w", err)
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(body); err != nil {
		return fmt.Errorf("file download checksum write: %w", err)
	}
	err = c.Attachment(file.Name(), "checksums.txt")
	if err != nil {
		return fmt.Errorf("file download checksum attachment: %w", err)
	}
	return nil
}

// Download configuration.
type Download struct {
	Path   string // Path is the absolute path to the download directory.
	Inline bool   // Inline is true if the file should attempt to display in the browser.
}

// HTTPSend serves files to the client and prompts for a save location.
// The download relies on the URL ID parameter to determine the requested file.
func (d Download) HTTPSend(c echo.Context, db *sql.DB, logger *zap.SugaredLogger) error {
	id := c.Param("id")
	ctx := context.Background()
	art, err := model.OneFileByKey(ctx, db, id)
	switch {
	case err != nil && sess.Editor(c):
		art, err = model.OneEditByKey(ctx, db, id)
		if err != nil {
			return fmt.Errorf("http send, one edit by key: %w", err)
		}
	case err != nil:
		return fmt.Errorf("http send, one file by key: %w", err)
	}
	name := art.Filename.String
	uid := strings.TrimSpace(art.UUID.String)
	file := filepath.Join(d.Path, uid)
	if !helper.Stat(file) {
		logger.Warnf("The hosted file download %q, for record %d does not exist.\n"+
			"Absolute path: %q", art.Filename.String, art.ID, file)
		return fmt.Errorf("http send, %w: %s", ErrStat, name)
	}
	if name == "" {
		logger.Warnf("No filename exists for the record %d.", art.ID)
		name = file
	}
	if d.Inline && tags.IsText(art.Platform.String) {
		modernText, err := helper.UTF8(file)
		if err != nil {
			return fmt.Errorf("http send utf-8: %w", err)
		}
		if !modernText {
			c.Response().Header().Set(echo.HeaderContentType, "text/plain; charset=iso-8859-1")
		}
		if err := c.Inline(file, name); err != nil {
			return fmt.Errorf("http send text as inline: %w", err)
		}
		return nil
	}
	if d.Inline {
		if err := c.Inline(file, name); err != nil {
			return fmt.Errorf("http send inline: %w", err)
		}
		return nil
	}
	if err := c.Attachment(file, name); err != nil {
		return fmt.Errorf("http send attachment: %w", err)
	}
	return nil
}

// ExtraZip configuration.
type ExtraZip struct {
	ExtraPath    string // ExtraPath is the absolute path to the extra directory.
	DownloadPath string // DownloadPath is the absolute path to the download directory.
}

// HTTPSend looks for any zip files in the extra directory and serves them to the client,
// otherwise it will serve the standard download file.
//
// This is used for obsolute file types that have been rearchived into a standard zip file.
func (e ExtraZip) HTTPSend(c echo.Context, db *sql.DB) error {
	id := c.Param("id")
	ctx := context.Background()
	art, err := model.OneFileByKey(ctx, db, id)
	switch {
	case err != nil && sess.Editor(c):
		art, err = model.OneEditByKey(ctx, db, id)
		if err != nil {
			return fmt.Errorf("http extra send, one edit by key: %w", err)
		}
	case err != nil:
		return fmt.Errorf("http extra send, one file by key: %w", err)
	}
	ext := ".zip"
	name := filepath.Base(art.Filename.String) + ext
	uid := strings.TrimSpace(art.UUID.String)
	file := filepath.Join(e.ExtraPath, uid) + ext
	if !helper.Stat(file) {
		name = art.Filename.String
		file = filepath.Join(e.DownloadPath, uid)
	}
	if err := c.Attachment(file, name); err != nil {
		return fmt.Errorf("http extra send attachment: %w", err)
	}
	return nil
}

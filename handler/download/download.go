// Package download handles the client file downloads.
package download

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/sess"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/extensions"
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
	defer func() { _ = os.Remove(file.Name()) }()
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
	Dir    dir.Directory // Dir is the absolute path to the download directory.
	Inline bool          // Inline is true if the file should attempt to display in the browser.
}

// HTTPSend serves files to the client and prompts for a save location.
// The download relies on the URL ID parameter to determine the requested file.
func (d Download) HTTPSend(c echo.Context, db *sql.DB, logger *zap.SugaredLogger) error {
	key := c.Param("id")
	ctx := context.Background()
	art, err := model.OneFileByKey(ctx, db, key)
	switch {
	case err != nil && sess.Editor(c):
		art, err = model.OneEditByKey(ctx, db, key)
		if err != nil {
			return fmt.Errorf("http send, one edit by key: %w", err)
		}
	case err != nil:
		return fmt.Errorf("http send, one file by key: %w", err)
	}
	name := art.Filename.String
	uid := strings.TrimSpace(art.UUID.String)
	file := d.Dir.Join(uid)
	if !helper.Stat(file) {
		if logger != nil {
			logger.Warnf("The hosted file download %q, for record %d does not exist.\n"+
				"Absolute path: %q", art.Filename.String, art.ID, file)
		}
		return fmt.Errorf("http send, %w: %s", ErrStat, name)
	}
	if name == "" {
		if logger != nil {
			logger.Warnf("No filename exists for the record %d.", art.ID)
		}
		name = file
	}
	if d.Inline {
		text := tags.IsText(art.Platform.String)
		ext := filepath.Ext(art.Filename.String)
		return inline(c, text, file, name, ext)
	}
	if err := c.Attachment(file, name); err != nil {
		return fmt.Errorf("http send attachment: %w", err)
	}
	return nil
}

func inline(c echo.Context, text bool, file, name, ext string) error {
	if text && slices.Contains(extensions.Image(), ext) {
		text = false
	}
	if text && slices.Contains(extensions.Media(), ext) {
		text = false
	}
	if !text {
		if err := c.Inline(file, name); err != nil {
			return fmt.Errorf("http send inline: %w", err)
		}
		return nil
	}
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

// ExtraZip configuration.
type ExtraZip struct {
	Extra    dir.Directory // Extra is the absolute path to the extra directory.
	Download dir.Directory // Download is the absolute path to the download directory.
}

// HTTPSend looks for any zip files in the extra directory and serves them to the client,
// otherwise it will serve the standard download file.
//
// This is used for obsolete file types that have been re-archived into a standard zip file.
func (e ExtraZip) HTTPSend(c echo.Context, db *sql.DB) error {
	key := c.Param("id")
	ctx := context.Background()
	art, err := model.OneFileByKey(ctx, db, key)
	switch {
	case err != nil && sess.Editor(c):
		art, err = model.OneEditByKey(ctx, db, key)
		if err != nil {
			return fmt.Errorf("http extra send, one edit by key: %w", err)
		}
	case err != nil:
		return fmt.Errorf("http extra send, one file by key: %w", err)
	}
	ext := ".zip"
	name := filepath.Base(art.Filename.String) + ext
	uid := strings.TrimSpace(art.UUID.String)
	file := e.Extra.Join(uid + ext)
	if !helper.Stat(file) {
		name = art.Filename.String
		file = e.Download.Join(uid)
	}
	if err := c.Attachment(file, name); err != nil {
		return fmt.Errorf("http extra send attachment: %w", err)
	}
	return nil
}

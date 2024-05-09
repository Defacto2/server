// Package download handles the client file downloads.
package download

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Defacto2/server/handler/sess"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var (
	ErrSum  = errors.New("file download checksum was not found")
	ErrStat = errors.New("file download stored on this server cannot be found")
)

// Checksum serves the checksums for the requested file.
// The response is a text file named "checksums.txt" with the checksum and filename.
// The id string is the UID filename of the requested file.
func Checksum(c echo.Context, id string) error {
	art, err := model.FindObf(id)
	if err != nil {
		if errors.Is(err, model.ErrDB) && sess.Editor(c) {
			art, err = model.EditObf(id)
		}
		if err != nil {
			return fmt.Errorf("model.FindObf: %w", err)
		}
	}
	// an example checksum file body created by `shasum`
	// 72f8a29d75993487b7ad5ad3a17d2f65ed4c41be155adbda88258d0458fcfe29f55e2e31b0316f01d57f4427ca9e2422  sk8-01.jpg
	sum := strings.TrimSpace(art.FileIntegrityStrong.String)
	if sum == "" {
		return fmt.Errorf("%w: %d", ErrSum, art.ID)
	}
	name := art.Filename.String
	body := []byte(sum + " " + name)

	file, err := os.CreateTemp(os.TempDir(), "checksum-server.*.txt")
	if err != nil {
		return fmt.Errorf("os.CreateTemp: %w", err)
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(body); err != nil {
		return fmt.Errorf("file.Write: %w", err)
	}
	err = c.Attachment(file.Name(), "checksums.txt")
	if err != nil {
		return fmt.Errorf("c.Attachment: %w", err)
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
func (d Download) HTTPSend(c echo.Context, logger *zap.SugaredLogger) error {
	id := c.Param("id")
	art, err := model.FindObf(id)
	if err != nil {
		if errors.Is(err, model.ErrDB) && sess.Editor(c) {
			art, err = model.EditObf(id)
		}
		if err != nil {
			return fmt.Errorf("model.FindObf: %w", err)
		}
	}
	name := art.Filename.String
	uid := strings.TrimSpace(art.UUID.String)
	file := filepath.Join(d.Path, uid)
	if !helper.Stat(file) {
		logger.Warnf("The hosted file download %q, for record %d does not exist.\n"+
			"Absolute path: %q", art.Filename.String, art.ID, file)
		return fmt.Errorf("%w: %s", ErrStat, name)
	}
	if name == "" {
		logger.Warnf("No filename exists for the record %d.", art.ID)
		name = file
	}
	if d.Inline {
		if err := c.Inline(file, name); err != nil {
			return fmt.Errorf("c.Inline: %w", err)
		}
		return nil
	}
	if err := c.Attachment(file, name); err != nil {
		return fmt.Errorf("c.Attachment: %w", err)
	}
	return nil
}

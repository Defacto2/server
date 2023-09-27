// Package download handles the client file downloads.
package download

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
func Checksum(z *zap.SugaredLogger, c echo.Context, id string) error {
	res, err := model.OneRecord(z, c, id)
	if err != nil {
		return err
	}

	// an example checksum file body created by `shasum`
	// 72f8a29d75993487b7ad5ad3a17d2f65ed4c41be155adbda88258d0458fcfe29f55e2e31b0316f01d57f4427ca9e2422  sk8-01.jpg
	sum := res.FileIntegrityStrong.String
	if sum == "" {
		return fmt.Errorf("%w: %d", ErrSum, res.ID)
	}
	name := res.Filename.String
	body := fmt.Sprintf("%s  %s", sum, name)

	file, err := os.CreateTemp(os.TempDir(), "checksum-server.*.txt")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err := file.Write([]byte(body)); err != nil {
		return err
	}
	return c.Attachment(file.Name(), "checksums.txt")
}

// Download configuration.
type Download struct {
	Inline bool   // Inline is true if the file should attempt to display in the browser.
	Path   string // Path is the absolute path to the download directory.
}

// HTTPSend serves files to the user and prompts for a save location.
// The download relies on the URL ID parameter to determine the requested file.
func (d Download) HTTPSend(z *zap.SugaredLogger, c echo.Context) error {
	res, err := model.OneRecord(z, c, c.Param("id"))
	if err != nil {
		return err
	}
	// build the source filepath
	name := res.Filename.String
	uid := strings.TrimSpace(res.UUID.String)
	file := filepath.Join(d.Path, uid)
	if !helper.IsStat(file) {
		z.Warnf("The hosted file download %q, for record %d does not exist.\n"+
			"Absolute path: %q", res.Filename.String, res.ID, file)
		return fmt.Errorf("%w: %s", ErrStat, name)
	}
	// pass the original filename to the client browser
	if name == "" {
		z.Warnf("No filename exists for the record %d.", res.ID)
		name = file
	}
	if d.Inline {
		return c.Inline(file, name)
	}
	return c.Attachment(file, name)
}

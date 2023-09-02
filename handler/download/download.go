// Package download handles the client file downloads.
package download

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/pkg/helper"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Download configuration.
type Download struct {
	Inline bool   // Inline is true if the file should attempt to display in the browser.
	Path   string // Path is the absolute path to the download directory.
}

const startID = 1 // startID is the default, first ID value.

var (
	ErrCtx  = errors.New("echo context is nil")
	ErrDB   = errors.New("database is not available")
	ErrID   = errors.New("file download database id cannot be found")
	ErrStat = errors.New("file download stored on this server cannot be found")
	ErrZap  = errors.New("zap logger instance is nil")
)

// HTTPSend serves files to the user and prompts for a save location.
// The download relies on the URL ID parameter to determine the requested file.
func (d Download) HTTPSend(z *zap.SugaredLogger, c echo.Context) error {
	if z == nil {
		return ErrZap
	}
	if c == nil {
		return ErrCtx
	}
	// https://go.dev/src/net/http/status.go
	// get id
	id := helper.DeobfuscateID(c.Param("id"))
	if id < startID {
		return fmt.Errorf("%w: %d ~ %s", ErrID, id, c.Param("id"))
	}
	// get record id, filename, uuid
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return ErrDB
	}
	defer db.Close()
	res, err := model.One(ctx, db, id)
	if err != nil {
		return ErrDB
	}
	if res.ID != int64(id) {
		return fmt.Errorf("%w: %d ~ %s", ErrID, id, c.Param("id"))
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

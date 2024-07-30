package htmx

// THIS IS A PLACEHOLDER
// These funcs used to be under handler/app but have been moved here for now.

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type extract int // extract target format for the file archive extractor

const (
	picture  extract = iota // extract a picture or image
	ansitext                // extract ansilove compatible text
)

// AnsiLovePost handles the post submission for the Preview from text in archive.
func AnsiLovePost(c echo.Context, dir app.Dirs, logger *zap.SugaredLogger) error {
	return extractor(c, dir, logger, ansitext)
}

// PreviewDel handles the post submission for the Delete complementary images button.
func PreviewDel(c echo.Context, dir app.Dirs) error {
	var f app.Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return badRequest(c, err)
	}
	defer db.Close()
	r, err := model.One(ctx, db, true, f.ID)
	if err != nil {
		return badRequest(c, err)
	}
	if err = command.RemoveImgs(r.UUID.String, dir.Preview, dir.Thumbnail); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

// PreviewPost handles the post submission for the Preview from image in archive.
func PreviewPost(c echo.Context, dir app.Dirs, logger *zap.SugaredLogger) error {
	return extractor(c, dir, logger, picture)
}

// extractor is a helper function for the PreviewPost and AnsiLovePost handlers.
func extractor(c echo.Context, dir app.Dirs, logger *zap.SugaredLogger, p extract) error {
	var f app.Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return badRequest(c, err)
	}
	defer db.Close()
	r, err := model.One(ctx, db, true, f.ID)
	if err != nil {
		return badRequest(c, err)
	}

	list := strings.Split(r.FileZipContent.String, "\n")
	target := ""
	for _, x := range list {
		s := strings.TrimSpace(x)
		if s == "" {
			continue
		}
		if strings.EqualFold(s, f.Target) {
			target = s
		}
	}
	if target == "" {
		return badRequest(c, app.ErrTarget)
	}
	src := filepath.Join(dir.Download, r.UUID.String)
	cmd := command.Dirs{Download: dir.Download, Preview: dir.Preview, Thumbnail: dir.Thumbnail}
	ext := filepath.Ext(strings.ToLower(r.Filename.String))
	switch p {
	case picture:
		err = cmd.ExtractImage(logger, src, ext, r.UUID.String, target)
	case ansitext:
		err = cmd.ExtractAnsiLove(logger, src, ext, r.UUID.String, target)
	default:
		err1 := fmt.Errorf("%w: %d", app.ErrExtract, p)
		return app.InternalErr(c, "extractor", err1) //nolint:wrapcheck
	}
	if err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

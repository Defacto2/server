package htmx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/nils"
	"github.com/labstack/echo/v5"
)

type Copy int

const (
	FileID Copy = iota
	Text
	Helper
)

func (cp Copy) String() string {
	switch cp {
	case FileID:
		return "DIZ"
	case Text:
		return "Text"
	case Helper:
		return "Helper text"
	}
	return ""
}

func (cp Copy) Ext() string {
	switch cp {
	case FileID:
		return ".diz"
	case Text:
		return ".txt"
	case Helper:
		return ".hlp"
	}
	return ""
}

func (cp Copy) MkCopy(c *echo.Context, dirs command.Dirs) error {
	const format = "artifact copy duplicator: %w"
	if err := nils.Check(c); err != nil {
		return fmt.Errorf(format, err)
	}

	unid, name, err := Path(c)
	if err != nil {
		return badRequest(c, err)
	}

	// FIX: logic ?
	// I believe originally the temp directory was fixed based on the unid?
	tmp, err := dir.MkdirTemp(unid)
	if err != nil {
		return badRequest(c, err)
	}

	name = filepath.Clean(name)
	src := filepath.Join(tmp, name)
	st, err := os.Stat(src)
	if err != nil {
		return badRequest(c, err)
	}
	if st.Size() == 0 {
		return c.String(http.StatusOK, "The file is empty and was not copied.")
	}

	dst := filepath.Join(dirs.Extra.Path(), unid+cp.Ext())
	if _, err = helper.DuplicateOW(src, dst); err != nil {
		return badRequest(c, err)
	}

	c = PageReload(c)

	s := cp.String() + ` copied, the browser will refresh.`
	return c.String(http.StatusOK, s)
}

// Package file artifactfs.go has all the funcs that use htmx
// to modify files and artifact assets kept on the file system.
// Examples, include thumbnails, preview images, readme texts etc.

// FSCopyDIZ handles the htmx request to use the file_id.diz artifact as a preview.
func FSCopyDIZ(c *echo.Context, dirs command.Dirs) error {
	return FileID.MkCopy(c, dirs)
}

// FSCopyHelp handles the htmx request to use the helper textfile as a preview.
func FSCopyHelp(c *echo.Context, dirs command.Dirs) error {
	return Helper.MkCopy(c, dirs)
}

func FSCopyReadme(sl *slog.Logger, c *echo.Context, dirs command.Dirs) error {
	const format = "record readme copier: %w"
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf(format, err)
	}

	unid, name, err := Path(c)
	if err != nil {
		return badRequest(c, err)
	}

	tmp, err := dir.MkdirTemp(unid)
	if err != nil {
		return badRequest(c, err)
	}

	src := filepath.Join(tmp, name)
	st, err := os.Stat(src)
	if err != nil {
		return badRequest(c, err)
	}
	if st.Size() == 0 {
		return c.String(http.StatusOK, "The file is empty and was not copied.")
	}

	dst := filepath.Join(dirs.Extra.Path(), unid+Text.Ext())
	if _, err = helper.DuplicateOW(src, dst); err != nil {
		return badRequest(c, err)
	}

	missingPNG := !helper.File(filepath.Join(dirs.Thumbnail.Path(), unid+".png"))
	missingWebp := !helper.File(filepath.Join(dirs.Thumbnail.Path(), unid+".webp"))
	if missingPNG && missingWebp {
		const amigaFont = false
		ctx := c.Request().Context()
		if err := dirs.TextImager(ctx, sl, src, unid, amigaFont); err != nil {
			return badRequest(c, err)
		}
	}

	c = PageReload(c)

	return c.String(http.StatusOK,
		`Images copied, the browser will refresh.`)
}

// FSCrop handles the htmx request for the preview image cropping.
func FSCrop(sl *slog.Logger, c *echo.Context, crop command.Crop, dirs command.Dirs,
) error {
	const format = "artifact record image cropper: %w"
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf(format, err)
	}

	unid, err := UUID(c)
	if err != nil {
		return badRequest(c, err)
	}

	ctx := c.Request().Context()
	err = crop.Images(ctx, sl, unid, dirs.Preview)
	if errors.Is(err, command.ErrNoImages) {
		return c.String(http.StatusOK, fmt.Sprint(err))
	}
	if err != nil {
		return badRequest(c, err)
	}

	c = PageReload(c)

	return c.String(http.StatusOK,
		`Images cropped, the browser will refresh.`)
}

// FSAlign handles the htmx request for the thumbnail crop alignment.
func FSAlign(sl *slog.Logger, c *echo.Context, align command.Align, dirs command.Dirs) error {
	const format = "artifact record thumb alignment: %w"
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf(format, err)
	}

	unid, err := UUID(c)
	if err != nil {
		return badRequest(c, err)
	}

	ctx := c.Request().Context()
	err = align.Thumbs(ctx, sl, unid, dirs.Preview, dirs.Thumbnail)
	if errors.Is(err, command.ErrNoImages) {
		return c.String(http.StatusOK, fmt.Sprint(err))
	}
	if err != nil {
		return badRequest(c, err)
	}

	c = PageReload(c)

	return c.String(http.StatusOK,
		`Thumb realigned, the browser will refresh.`)
}

// FSThumb handles the htmx request for the thumbnail quality.
func FSThumb(sl *slog.Logger, c *echo.Context, thumb command.Generate, dirs command.Dirs) error {
	const format = "artifact record thumb: %w"
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf(format, err)
	}

	unid, err := UUID(c)
	if err != nil {
		return badRequest(c, err)
	}

	const timeout = command.CmdTimeout * 2
	ctx, cancel := context.WithTimeout(c.Request().Context(), timeout)
	defer cancel()

	err = dirs.Thumbs(ctx, sl, unid, thumb)
	if errors.Is(err, command.ErrNoImages) {
		return c.String(http.StatusOK, fmt.Sprint(err))
	}
	if err != nil {
		return badRequest(c, err)
	}

	c = PageReload(c)

	return c.String(http.StatusOK,
		`Thumb created, the browser will refresh.`)
}

// FSPixelate handles the htmx request to pixelate both the preview and
// thumbnails, if they are not suitable for a general audience. This also has an
// added benefit of reducing the file sizes of both images and reducing page load.
func FSPixelate(sl *slog.Logger, c *echo.Context, directory ...dir.Directory) error {
	const format = "record image pixelator: %w"
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf(format, err)
	}

	unid, err := UUID(c)
	if err != nil {
		return badRequest(c, err)
	}

	const timeout = command.CmdTimeout * 2
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	dirs := dir.Paths(directory...)
	if err := command.ImagesPixelate(ctx, sl, unid, dirs...); err != nil {
		return badRequest(c, err)
	}
	// NOTE: do not use pageRefresh as it returns an error
	// c = pageRefresh(c)

	return c.String(http.StatusOK,
		`Images pixelated, please refresh your browser.`)
}

// FSRemoveDIZ handles the request to remove the uuid named file_id.diz text file
// from the provided extra directory.
func FSRemoveDIZ(c *echo.Context, extra dir.Directory) error {
	return extrasDeleter(c, FileID.Ext(), extra)
}

// FSRemoveHelp handles the request to remove the uuid named helper (.hlp) text file
// from the provided extra directory.
func FSRemoveHelp(c *echo.Context, extra dir.Directory) error {
	return extrasDeleter(c, Helper.Ext(), extra)
}

// FSRemoveReadme handles the request to remove the uuid named readme text file
// from the provided extra directory.
func FSRemoveReadme(c *echo.Context, extra dir.Directory) error {
	return extrasDeleter(c, Text.Ext(), extra)
}

// TODO: struct?
func extrasDeleter(c *echo.Context, ext string, extra dir.Directory) error {
	const format = "extras deleter %s: %w"
	if err := nils.Check(c); err != nil {
		return fmt.Errorf(format, ext, err)
	}

	unid, err := UUID(c)
	if err != nil {
		return badRequest(c, err)
	}

	dst := filepath.Join(extra.Path(), unid+ext)
	dst = filepath.Clean(dst)
	st, err := os.Stat(dst)
	if errors.Is(err, os.ErrNotExist) {
		return c.NoContent(http.StatusOK) //nolint:wrapcheck
	}
	if err != nil {
		return badRequest(c, err)
	}
	if st.IsDir() {
		return badRequest(c, ErrIsDir)
	}

	if err := os.Remove(dst); err != nil {
		return badRequest(c, err)
	}

	c = PageReload(c)

	return c.NoContent(http.StatusOK) //nolint:wrapcheck
}

// FSRemoveImages handles the request to remove the uuid named
// image files from the directories provided.
func FSRemoveImages(c *echo.Context, directories ...dir.Directory) error {
	const format = "record images deleter: %w"
	if err := nils.Check(c); err != nil {
		return fmt.Errorf(format, err)
	}

	unid, err := UUID(c)
	if err != nil {
		return badRequest(c, err)
	}

	dirs := make([]string, len(directories))
	for i, directory := range directories {
		dirs[i] = directory.Path()
	}

	if err := command.ImagesDelete(unid, dirs...); err != nil {
		if errors.Is(err, command.ErrNoImages) {
			return c.String(http.StatusOK, err.Error())
		}
		return badRequest(c, err)
	}

	// NOTE: do not use pageRefresh as it returns an error
	// c = pageRefresh(c)

	return c.String(http.StatusOK, "Images are erased. "+
		"However, depending on the download type, the assets may recreate automatically. "+
		"Please refresh your browser.")
}

// FSUseReadme handles the htmx request to use the text file artifact as a preview.
func FSUseReadme(sl *slog.Logger, c *echo.Context, amigaFont bool, dirs command.Dirs,
) error {
	const format = "record readme imager: %w"
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf(format, err)
	}

	unid, name, err := Path(c)
	if err != nil {
		return badRequest(c, err)
	}

	tmp, err := dir.MkdirTemp(unid)
	if err != nil {
		return badRequest(c, err)
	}

	name = filepath.Clean(name)
	src := filepath.Join(tmp, name)
	st, err := os.Stat(src)
	if err != nil {
		return badRequest(c, err)
	}
	if st.Size() == 0 {
		return c.String(http.StatusOK, "The file is empty and was not used.")
	}

	ctx := c.Request().Context()
	if err := dirs.TextImager(ctx, sl, src, unid, amigaFont); err != nil {
		return badRequest(c, err)
	}

	c = PageReload(c)

	return c.String(http.StatusOK,
		`Text filed imaged, the browser will refresh.`)
}

type useProcess struct {
	format  string
	empty   string
	success string
}

// FSUseImage handles the htmx request to use an image file artifact as a preview.
func FSUseImage(sl *slog.Logger, c *echo.Context, dirs command.Dirs) error {
	up := useProcess{
		"record image copier",
		"The file is empty and was not copied.",
		"Images copied, the browser will refresh.",
	}
	return up.process(sl, c, dirs.PictureImager)
}

// FSUseBinText handles the htmx request to use the text file artifact as a preview.
func FSUseBinText(sl *slog.Logger, c *echo.Context, dirs command.Dirs) error {
	up := useProcess{
		"record binary text readme imager",
		"The file is empty and was not used.",
		"Binary text imaged, the browser will refresh.",
	}
	return up.process(sl, c, dirs.BinTextImager)
}

// process is a helper function that handles the common file processing logic
// for both image copying and binary text imaging operations.
func (fp useProcess) process(sl *slog.Logger, c *echo.Context,
	processFunc func(context.Context, *slog.Logger, string, string) error,
) error {
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf("%s: %w", fp.format, err)
	}

	unid, name, err := Path(c)
	if err != nil {
		return badRequest(c, err)
	}

	name = filepath.Clean(name)
	tmp, err := dir.MkdirTemp(unid)
	if err != nil {
		return badRequest(c, err)
	}

	src := filepath.Join(tmp, name)
	st, err := os.Stat(src)
	if err != nil {
		return badRequest(c, err)
	}
	if st.Size() == 0 {
		return c.String(http.StatusOK, fp.empty)
	}

	if err := processFunc(c.Request().Context(), sl, src, unid); err != nil {
		return badRequest(c, err)
	}

	c = PageReload(c)

	return c.String(http.StatusOK, fp.success)
}

package htmx

// Package file artifact.go provides functions for handling the HTMX requests for the artifact editor.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/handler/form"
	"github.com/Defacto2/server/handler/jsdos"
	"github.com/Defacto2/server/handler/pouet"
	"github.com/Defacto2/server/handler/releaser"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

var (
	ErrFileIsDir = errors.New("the file is a directory")
	ErrPath      = errors.New("the file path is invalid")
	ErrYouTube   = errors.New("youtube watch video id needs to be empty or 11 characters")
)

const (
	checkMark   = "&#x2713;"
	editorKey   = "artifact-editor-key"
	successSpan = `<span class="text-success">✓</span>`
)

// Validate checks that the path is not absolute and does not allow
// for any path traversal. It returns an error if the path is invalid.
func Validate(path string) error {
	const format = "%w: %q"
	path = strings.TrimSuffix(path, "/")
	if filepath.IsAbs(path) {
		return fmt.Errorf(format, ErrPath, path)
	}
	if s := filepath.Clean(path); s != path {
		return fmt.Errorf(format, ErrPath, path)
	}
	return nil
}

// Path returns the uuid and directory path, the named unid, plus the path from the URL parameters.
// It returns an error if the unid or name is invalid.
func Path(c *echo.Context) (string, string, error) {
	if err := nils.Check(c); err != nil {
		return "", "", fmt.Errorf("artifact path: %w", err)
	}
	unid := c.Param("unid")
	if err := form.Checkname(unid); err != nil {
		return "", "", fmt.Errorf("invalid unid format: %w", err)
	}
	if err := uuid.Validate(unid); err != nil {
		return "", "", fmt.Errorf("invalid uuid: %w", err)
	}
	path := c.Param("path")
	name, err := url.QueryUnescape(path)
	if err != nil {
		return "", "", fmt.Errorf("failed to unescape path: %w", err)
	}
	if err := Validate(name); err != nil {
		return "", "", err
	}
	return unid, name, nil
}

// UUID returns the uuid from the URL parameters and returns an error if it is invalid.
func UUID(c *echo.Context) (string, error) {
	if err := nils.Check(c); err != nil {
		return "", fmt.Errorf("artifact uuid: %w", err)
	}
	unid := c.Param("unid")
	if err := form.Checkname(unid); err != nil {
		return "", fmt.Errorf("invalid unid format: %w", err)
	}
	if err := uuid.Validate(unid); err != nil {
		return "", fmt.Errorf("invalid uuid: %w", err)
	}
	return unid, nil
}

// ID returns the id from the URL parameters and returns an error if it is invalid.
func ID(c *echo.Context) (int, error) {
	if err := nils.Check(c); err != nil {
		return 0, fmt.Errorf("artifact id: %w", err)
	}
	id, err := echo.PathParam[int](c, "id")
	if err != nil {
		const format = `%w: "%w"`
		return 0, fmt.Errorf(format, ErrKey, err)
	}
	return id, nil
}

// pageRefresh is a helper function to set the HTTP [HTMX header] for the browser to refresh the page.
//
// [HTMX header]: https://htmx.org/reference/#response_headers
func pageRefresh(c *echo.Context) *echo.Context {
	c.Response().Header().Set("HX-Refresh", "true")
	c.Response().WriteHeader(http.StatusFound)
	return c
}

// RecordThumb handles the htmx request for the thumbnail quality.
func RecordThumb(
	ctx context.Context, sl *slog.Logger, c *echo.Context, thumb command.Generate, dirs command.Dirs,
) error {
	const format = "artifact record thumb: %w"
	if err := nils.Check(ctx, sl, c); err != nil {
		return fmt.Errorf(format, err)
	}
	unid, err := UUID(c)
	if err != nil {
		return badRequest(c, err)
	}
	err = dirs.Thumbs(ctx, sl, unid, thumb)
	if errors.Is(err, command.ErrNoImages) {
		return c.String(http.StatusOK, fmt.Sprint(err))
	}
	if err != nil {
		return badRequest(c, err)
	}
	c = pageRefresh(c)
	return c.String(http.StatusOK,
		`Thumb created, the browser will refresh.`)
}

// RecordThumbAlignment handles the htmx request for the thumbnail crop alignment.
func RecordThumbAlignment(
	ctx context.Context, sl *slog.Logger, c *echo.Context, align command.Align, dirs command.Dirs,
) error {
	const format = "artifact record thumb alignment: %w"
	if err := nils.Check(ctx, sl, c); err != nil {
		return fmt.Errorf(format, err)
	}
	unid, err := UUID(c)
	if err != nil {
		return badRequest(c, err)
	}
	err = align.Thumbs(ctx, sl, unid, dirs.Preview, dirs.Thumbnail)
	if errors.Is(err, command.ErrNoImages) {
		return c.String(http.StatusOK, fmt.Sprint(err))
	}
	if err != nil {
		return badRequest(c, err)
	}
	c = pageRefresh(c)
	return c.String(http.StatusOK,
		`Thumb realigned, the browser will refresh.`)
}

// RecordImageCropper handles the htmx request for the preview image cropping.
func RecordImageCropper(
	ctx context.Context, sl *slog.Logger, c *echo.Context, crop command.Crop, dirs command.Dirs,
) error {
	const format = "artifact record image cropper: %w"
	if err := nils.Check(ctx, sl, c); err != nil {
		return fmt.Errorf(format, err)
	}
	unid, err := UUID(c)
	if err != nil {
		return badRequest(c, err)
	}
	err = crop.Images(ctx, sl, unid, dirs.Preview)
	if errors.Is(err, command.ErrNoImages) {
		return c.String(http.StatusOK, fmt.Sprint(err))
	}
	if err != nil {
		return badRequest(c, err)
	}
	c = pageRefresh(c)
	return c.String(http.StatusOK,
		`Images cropped, the browser will refresh.`)
}

// recordFileProcessor is a helper function that handles the common file processing logic
// for both image copying and binary text imaging operations.
func recordFileProcessor(ctx context.Context, sl *slog.Logger, c *echo.Context, _ command.Dirs,
	msg, emptyMsg, successMsg string,
	processFunc func(context.Context, *slog.Logger, string, string) error,
) error {
	if err := nils.Check(ctx, sl, c); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
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
		return c.String(http.StatusOK, emptyMsg)
	}
	if err := processFunc(ctx, sl, src, unid); err != nil {
		return badRequest(c, err)
	}
	c = pageRefresh(c)
	return c.String(http.StatusOK, successMsg)
}

// RecordImageCopier handles the htmx request to use an image file artifact as a preview.
func RecordImageCopier(ctx context.Context, sl *slog.Logger, c *echo.Context, dirs command.Dirs) error {
	return recordFileProcessor(ctx, sl, c, dirs,
		"record image copier",
		"The file is empty and was not copied.",
		"Images copied, the browser will refresh.",
		dirs.PictureImager)
}

// RecordBinTextImager handles the htmx request to use the text file artifact as a preview.
func RecordBinTextImager(ctx context.Context, sl *slog.Logger, c *echo.Context, dirs command.Dirs) error {
	return recordFileProcessor(ctx, sl, c, dirs,
		"record binary text readme imager",
		"The file is empty and was not used.",
		"Binary text imaged, the browser will refresh.",
		dirs.BinTextImager)
}

// RecordReadmeImager handles the htmx request to use the text file artifact as a preview.
func RecordReadmeImager(
	ctx context.Context, sl *slog.Logger, c *echo.Context, amigaFont bool, dirs command.Dirs,
) error {
	const format = "record readme imager: %w"
	if err := nils.Check(ctx, sl, c); err != nil {
		return fmt.Errorf(format, err)
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
		return c.String(http.StatusOK, "The file is empty and was not used.")
	}
	if err := dirs.TextImager(ctx, sl, src, unid, amigaFont); err != nil {
		return badRequest(c, err)
	}
	c = pageRefresh(c)
	return c.String(http.StatusOK,
		`Text filed imaged, the browser will refresh.`)
}

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

func (cp Copy) Duplicator(c *echo.Context, dirs command.Dirs) error {
	const format = "artifact copy duplicator: %w"
	if err := nils.Check(c); err != nil {
		return fmt.Errorf(format, err)
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
		return c.String(http.StatusOK, "The file is empty and was not copied.")
	}
	dst := filepath.Join(dirs.Extra.Path(), unid+cp.Ext())
	if _, err = helper.DuplicateOW(src, dst); err != nil {
		return badRequest(c, err)
	}
	c = pageRefresh(c)
	s := cp.String() + ` copied, the browser will refresh.`
	return c.String(http.StatusOK, s)
}

// RecordDizCopier handles the htmx request to use the file_id.diz artifact as a preview.
func RecordDizCopier(c *echo.Context, dirs command.Dirs) error {
	return FileID.Duplicator(c, dirs)
}

// RecordHlpCopier handles the htmx request to use the file_id.diz artifact as a preview.
func RecordHlpCopier(c *echo.Context, dirs command.Dirs) error {
	return Helper.Duplicator(c, dirs)
}

func RecordReadmeCopier(ctx context.Context, sl *slog.Logger, c *echo.Context, dirs command.Dirs) error {
	const format = "record readme copier: %w"
	if err := nils.Check(ctx, sl, c); err != nil {
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
	if !helper.File(filepath.Join(dirs.Thumbnail.Path(), unid+".png")) &&
		!helper.File(filepath.Join(dirs.Thumbnail.Path(), unid+".webp")) {
		const amigaFont = false
		if err := dirs.TextImager(ctx, sl, src, unid, amigaFont); err != nil {
			return badRequest(c, err)
		}
	}
	c = pageRefresh(c)
	return c.String(http.StatusOK,
		`Images copied, the browser will refresh.`)
}

// RecordReadmeDisable handles the htmx request to disable the in
// page display of both the text files readme and file_id.diz for the file artifact.
func RecordReadmeDisable(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record readme disable: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	id, err := ID(c)
	if err != nil {
		return badRequest(c, err)
	}
	value := c.FormValue("readme-is-off") != "on"

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err = model.ReadmeDisable.Update(ctx, tx, int64(id), value); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	return c.String(http.StatusOK, successSpan)
}

// RecordImagePixelator handles the htmx request to pixelate both the preview and
// thumbnails, if they are not suitable for a general audience. This also has an
// added benefit of reducing the file sizes of both images and reducing page load.
func RecordImagePixelator(ctx context.Context, sl *slog.Logger, c *echo.Context, directory ...dir.Directory) error {
	const format = "record image pixelator: %w"
	if err := nils.Check(ctx, sl, c); err != nil {
		return fmt.Errorf(format, err)
	}
	unid, err := UUID(c)
	if err != nil {
		return badRequest(c, err)
	}
	dirs := dir.Paths(directory...)
	if err := command.ImagesPixelate(ctx, sl, unid, dirs...); err != nil {
		return badRequest(c, err)
	}
	// do not use pageRefresh as it returns an error
	// c = pageRefresh(c)
	return c.String(http.StatusOK,
		`Images pixelated, the browser will refresh.`)
}

// RecordImagesDeleter handles the request to remove the uuid named
// image files from the directories provided.
func RecordImagesDeleter(c *echo.Context, directories ...dir.Directory) error {
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
	// do not use pageRefresh as it returns an error
	// c = pageRefresh(c)
	return c.String(http.StatusOK, "Images are gone, please refresh the tab. "+
		"However, depending on the download type, these assets may be recreated automatically.")
}

// RecordDizDeleter handles the request to remove the uuid named file_id.diz text file
// from the provided extra directory.
func RecordDizDeleter(c *echo.Context, extra dir.Directory) error {
	return extrasDeleter(c, FileID.Ext(), extra)
}

// RecordHlpDeleter handles the request to remove the uuid named helper (.hlp) text file
// from the provided extra directory.
func RecordHlpDeleter(c *echo.Context, extra dir.Directory) error {
	return extrasDeleter(c, Helper.Ext(), extra)
}

// RecordReadmeDeleter handles the request to remove the uuid named readme text file
// from the provided extra directory.
func RecordReadmeDeleter(c *echo.Context, extra dir.Directory) error {
	return extrasDeleter(c, Text.Ext(), extra)
}

func extrasDeleter(c *echo.Context, ext string, extra dir.Directory) error {
	const format = "extras deleter: %w"
	if err := nils.Check(c); err != nil {
		return fmt.Errorf(format, err)
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
		return badRequest(c, ErrFileIsDir)
	}
	if err := os.Remove(dst); err != nil {
		return badRequest(c, err)
	}
	c = pageRefresh(c)
	return c.NoContent(http.StatusOK) //nolint:wrapcheck
}

// RecordToggle handles the post submission for the file artifact record toggle.
// The return value is either "online" or "offline" depending on the state.
func RecordToggle(ctx context.Context, c *echo.Context, db *sql.DB, state bool) error {
	const format = "artifact record toggle %s: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, "check", err)
	}
	key := c.FormValue(editorKey)
	id, err := strconv.ParseInt(key, 10, 64)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, "begin tx", err)
	}
	if err := model.UpdateOnline(ctx, db, state, id); err != nil {
		return fmt.Errorf(format, "update", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf(format, "tx commit", err)
	}
	if state {
		return c.String(http.StatusOK, "online")
	}
	return c.String(http.StatusOK, "offline")
}

// RecordToggleByID handles the post submission for the file artifact record toggle.
// The key string is converted into an integer and used as the artifact id.
// The return value is either "online" or "offline" depending on the state.
func RecordToggleByID(ctx context.Context, c *echo.Context, db *sql.DB, key string, state bool) error {
	const format = "artifact record toggle by id %d, %s: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, 0, "check", err)
	}
	id, err := strconv.ParseInt(key, 10, 64)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, id, "begin tx", err)
	}
	if err := model.UpdateOnline(ctx, db, state, id); err != nil {
		return fmt.Errorf(format, id, "update", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf(format, id, "tx commit", err)
	}

	if state {
		return c.String(http.StatusOK, "Record is visible to the public.")
	}
	return c.String(http.StatusOK, "🚫 Record is disabled and hidden from public access. 🚫")
}

// RecordClassification handles the post submission for the file artifact classifications,
// such as the platform, operating system, section or category tags.
// The return value is either the humanized and counted classification or an error.
func RecordClassification(ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	const format = "record classification: %w"
	if err := nils.Check(ctx, sl, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	section := c.FormValue("artifact-editor-categories")
	platform := c.FormValue("artifact-editor-operatingsystem")
	key := c.FormValue(editorKey)
	if invalid := section == "" || platform == ""; invalid {
		html, err := form.HumanizeCount(ctx, db, section, platform)
		if err != nil {
			sl.Error("record classification", slog.Any("error", err))
			return badRequest(c, err)
		}
		return c.HTML(http.StatusOK, string(html)+" did not update")
	}
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	classification := model.Classification{
		ID: int64(id), Platform: platform, Tag: section,
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err := classification.Update(ctx, tx); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	html, err := form.HumanizeCount(ctx, db, section, platform)
	if err != nil {
		sl.Error("record classification", slog.Any("error", err))
		return badRequest(c, err)
	}
	return c.HTML(http.StatusOK, string(html))
}

// RecordFilename handles the post submission for the file artifact filename.
func RecordFilename(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record filename: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	name := c.FormValue("artifact-editor-filename")
	key := c.FormValue(editorKey)
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	name = form.SanitizeFilename(name)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.Filename.Update(ctx, tx, int64(id), name); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	return c.String(http.StatusOK, "Updated")
}

// RecordFilenameReset handles the post submission for the file artifact filename reset.
func RecordFilenameReset(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record filename reset: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	val := c.FormValue("artifact-editor-filename-undo")
	key := c.FormValue(editorKey)
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, err)
	}
	if err := model.Filename.Update(ctx, tx, int64(id), val); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	return c.String(http.StatusOK, val)
}

// RecordVirusTotal handles the post submission for the file artifact VirusTotal report link.
func RecordVirusTotal(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record virus total: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	link := c.FormValue("artifact-editor-virustotal")
	key := c.FormValue(editorKey)
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	if !form.ValidVT(link) {
		return c.NoContent(http.StatusNoContent) //nolint:wrapcheck
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.VirusTotal.Update(ctx, tx, int64(id), link); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	return c.String(http.StatusOK, "Updated")
}

// RecordTitle handles the post submission for the file artifact title.
func RecordTitle(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record title: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	title := c.FormValue("artifact-editor-title")
	key := c.FormValue(editorKey)
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, err)
	}
	if err := model.Title.Update(ctx, tx, int64(id), title); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf(format, err)
	}

	return c.String(http.StatusOK, "Updated")
}

// RecordTitleReset handles the post submission for the file artifact title reset.
func RecordTitleReset(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record title reset: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	val := c.FormValue("artifact-editor-titleundo")
	key := c.FormValue(editorKey)
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.Title.Update(ctx, tx, int64(id), val); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf(format, err)
	}

	return c.String(http.StatusOK, val)
}

// RecordComment handles the post submission for the file artifact comment.
func RecordComment(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record comment: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	comment := c.FormValue("artifact-editor-comment")
	key := c.FormValue(editorKey)
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.Comment.Update(ctx, tx, int64(id), comment); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	return c.String(http.StatusOK, "Updated")
}

// RecordCommentReset handles the post submission for the file artifact comment reset.
func RecordCommentReset(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record comment reset: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	val := c.FormValue("artifact-editor-comment-resetter")
	key := c.FormValue(editorKey)
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.Comment.Update(ctx, tx, int64(id), val); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	return c.String(http.StatusOK, "Undo comment")
}

// RecordReleasers handles the post submission for the file artifact releasers.
// It will only update the releaser1 and the releaser2 values if they have changed.
// The return value is either "Updated" or "Update" depending on if the values have changed.
func RecordReleasers(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record releasers: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	val1 := c.FormValue("releaser1")
	val2 := c.FormValue("releaser2")
	rel1 := c.FormValue("artifact-editor-releaser1")
	rel2 := c.FormValue("artifact-editor-releaser2")
	key := c.FormValue(editorKey)
	unchanged := (rel1 == val1 && rel2 == val2)
	if unchanged {
		return c.NoContent(http.StatusNoContent) //nolint:wrapcheck
	}
	if err := recordReleases(ctx, db, rel1, rel2, key); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Save")
}

// RecordReleasersReset handles the post submission for the file artifact releasers reset.
// It will always reset and save the releaser1 and the releaser2 values.
// The return value is always "Resetted" unless an error occurs.
func RecordReleasersReset(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record releaser reset: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	val1 := c.FormValue("releaser1")
	val2 := c.FormValue("releaser2")
	rel1 := c.FormValue("artifact-editor-releaser1")
	rel2 := c.FormValue("artifact-editor-releaser2")
	key := c.FormValue(editorKey)
	unchanged := (rel1 == val1 && rel2 == val2)
	if unchanged {
		return c.String(http.StatusNoContent, "")
	}
	if err := recordReleases(ctx, db, val1, val2, key); err != nil {
		return badRequest(c, err)
	}
	return c.HTML(http.StatusOK, checkMark)
}

func recordReleases(ctx context.Context, db *sql.DB, rel1, rel2, key string) error {
	const format = "record releases %s: %w"
	if err := nils.Check(ctx, db); err != nil {
		return fmt.Errorf(format, "check", err)
	}
	id, err := strconv.Atoi(key)
	if err != nil {
		return fmt.Errorf("%w: %w: %q", ErrKey, err, key)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(format, "begin tx", err)
	}
	if err := model.UpdateReleasers(ctx, db, int64(id), rel1, rel2); err != nil {
		return fmt.Errorf(format, "update releasers", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf(format, "tx commit", err)
	}

	return nil
}

// RecordDateIssued handles the post submission for the file artifact date of release.
func RecordDateIssued(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record date issued: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	year := c.FormValue("artifact-editor-year")
	month := c.FormValue("artifact-editor-month")
	day := c.FormValue("artifact-editor-day")
	key := c.FormValue(editorKey)
	yearval := c.FormValue("artifact-editor-yearval")
	monthval := c.FormValue("artifact-editor-monthval")
	dayval := c.FormValue("artifact-editor-dayval")
	if year == yearval && month == monthval && day == dayval {
		return c.NoContent(http.StatusNoContent) //nolint:wrapcheck
	}
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	y, m, d := form.ValidDate(year, month, day)
	if !y || !m || !d {
		const format = `%w, date failed to validate: Y %q %v ; M %q %v ; D %q %v `
		return badRequest(c, fmt.Errorf(format, ErrYMDFormat, year, y, month, m, day, d))
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.UpdateYMDS(ctx, tx, int64(id), year, month, day); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	return c.String(http.StatusOK, "Save")
}

// RecordDateIssuedReset handles the post submission for the file artifact date of release reset.
func RecordDateIssuedReset(ctx context.Context, c *echo.Context, db *sql.DB, elmID string) error {
	if err := nils.Check(ctx, c, db); err != nil {
		const format = "record date issued reset: %w"
		return fmt.Errorf(format, err)
	}
	reset := c.FormValue(elmID)
	key := c.FormValue(editorKey)
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	vals := strings.Split(reset, "-")
	const expected = 3
	const format = `%w, record date issued reset requires YYYY-MM-DD`
	if len(vals) != expected {
		return badRequest(c, fmt.Errorf(format, ErrYMDFormat))
	}
	year, month, day := vals[0], vals[1], vals[2]
	y, m, d := form.ValidDate(year, month, day)
	if invalid := !y || !m || !d; invalid {
		return badRequest(c, fmt.Errorf(format, ErrYMDFormat))
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.UpdateYMDS(ctx, tx, int64(id), year, month, day); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	return c.String(http.StatusOK, " "+checkMark)
}

// RecordCreatorText handles the post submission for the file artifact creator text.
func RecordCreatorText(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record creator text: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	creator := c.FormValue("artifact-editor-credittext")
	key := c.FormValue(editorKey)
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	val := creatorFix(creator)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.CreatorText.Update(ctx, tx, int64(id), val); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

// RecordCreatorIll handles the post submission for the file artifact creator illustrator.
func RecordCreatorIll(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record creator illustrator: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	creator := c.FormValue("artifact-editor-creditill")
	key := c.FormValue(editorKey)
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	val := creatorFix(creator)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.CreatorIll.Update(ctx, tx, int64(id), val); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

// RecordCreatorProg handles the post submission for the file artifact creator programmer.
func RecordCreatorProg(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record creator programmer: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	creator := c.FormValue("artifact-editor-creditprog")
	key := c.FormValue(editorKey)
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	val := creatorFix(creator)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.CreatorProg.Update(ctx, tx, int64(id), val); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	return c.String(http.StatusOK, "Updated")
}

// RecordCreatorAudio handles the post submission for the file artifact creator musician.
func RecordCreatorAudio(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record creator audio: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	creator := c.FormValue("artifact-editor-creditaudio")
	key := c.FormValue(editorKey)
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	val := creatorFix(creator)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.CreatorAudio.Update(ctx, tx, int64(id), val); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	return c.String(http.StatusOK, "Updated")
}

func creatorFix(s string) string {
	creators := strings.Split(s, ",")
	for i, c := range creators {
		creators[i] = releaser.Clean(c)
	}
	return strings.Join(creators, ",")
}

// RecordCreatorReset handles the post submission for the file artifact creators reset.
func RecordCreatorReset(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record creator reset: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	// form values must be the "name" value of html elements
	reset := c.FormValue("artifact-editor-creditsundo")
	textval := c.FormValue("artifact-editor-credittext")
	illval := c.FormValue("artifact-editor-creditill")
	progval := c.FormValue("artifact-editor-creditprog")
	audioval := c.FormValue("artifact-editor-creditaudio")
	key := c.FormValue(editorKey)
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	vals := strings.Split(reset, ";")
	const expected = 4
	if len(vals) != expected {
		const format = `%w, record creator reset requires string;string;string;string`
		return badRequest(c, fmt.Errorf(format, ErrYMDFormat))
	}

	text := vals[0]
	ill := vals[1]
	prog := vals[2]
	audio := vals[3]
	creators := model.Creators{
		ID:    int64(id),
		Text:  text,
		Ill:   ill,
		Prog:  prog,
		Audio: audio,
	}
	if textval == text && illval == ill && progval == prog && audioval == audio {
		return c.NoContent(http.StatusNoContent) //nolint:wrapcheck
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err := creators.Update(ctx, tx); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	return c.String(http.StatusOK, "Undo creators")
}

// RecordYouTube handles the post submission for the file artifact YouTube watch video link.
func RecordYouTube(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record youtube: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	key := c.FormValue(editorKey)
	newVideo := strings.TrimSpace(c.FormValue("artifact-editor-youtube"))
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	const requirement = 11
	if len(newVideo) != 0 && len(newVideo) != requirement {
		return c.NoContent(http.StatusNoContent) //nolint:wrapcheck
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.YouTube.Update(ctx, tx, int64(id), newVideo); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	return RecordLinks(c)
}

// RecordDemozoo handles the post submission for the file artifact Demozoo production link.
func RecordDemozoo(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record demozoo: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	key := c.FormValue(editorKey)
	newProd := c.FormValue("artifact-editor-demozoo")
	if newProd == "" {
		newProd = "0"
	}
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.Demozoo.Update(ctx, tx, int64(id), newProd); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	return RecordLinks(c)
}

// RecordPouet handles the post submission for the file artifact Pouet production link.
func RecordPouet(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record pouet: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	key := c.FormValue(editorKey)
	newProd := c.FormValue("artifact-editor-pouet")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.Pouet.Update(ctx, tx, int64(id), newProd); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	return RecordLinks(c)
}

// Record16Colors handles the post submission for the file artifact 16 Colors link.
func Record16Colors(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record 16colors: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	key := c.FormValue(editorKey)
	newURL := c.FormValue("artifact-editor-16colors")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	link := form.SanitizeURLPath(newURL)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.Colors16.Update(ctx, tx, int64(id), link); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	return RecordLinks(c)
}

// RecordGitHub handles the post submission for the file artifact GitHub repository link.
func RecordGitHub(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record github: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	key := c.FormValue(editorKey)
	newRepo := c.FormValue("artifact-editor-github")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	link := form.SanitizeGitHub(newRepo)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.GitHub.Update(ctx, tx, int64(id), link); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	return RecordLinks(c)
}

// RecordRelations handles the post submission for the file artifact releaser relationships.
func RecordRelations(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record relations: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	key := c.FormValue(editorKey)
	newRelations := c.FormValue("artifact-editor-relations")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.Relations.Update(ctx, tx, int64(id), newRelations); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	return RecordLinks(c)
}

// RecordSites handles the post submission for the file artifact website links.
func RecordSites(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record sites: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	key := c.FormValue(editorKey)
	newSites := c.FormValue("artifact-editor-websites")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.Sites.Update(ctx, tx, int64(id), newSites); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	return RecordLinks(c)
}

// RecordLinks handles the post submission for a form submission to provide the
// HTML formatted links for the "Links" section of the artifact editor.
func RecordLinks(c *echo.Context) error {
	youtube := c.FormValue("artifact-editor-youtube")
	demozoo := c.FormValue("artifact-editor-demozoo")
	pouet := c.FormValue("artifact-editor-pouet")
	colors16 := c.FormValue("artifact-editor-16colors")
	github := c.FormValue("artifact-editor-github")
	rels := c.FormValue("artifact-editor-relations")
	sites := c.FormValue("artifact-editor-websites")
	links := app.LinkPreviews(youtube, demozoo, pouet, colors16, github, rels, sites)
	for i, link := range links {
		links[i] = `<small><strong>Link to</strong></small> &nbsp; ` + link
	}
	return c.HTML(http.StatusOK, strings.Join(links, "<br>"))
}

// RecordLinksReset handles the post submission for the file artifact links reset.
func RecordLinksReset(ctx context.Context, c *echo.Context, db *sql.DB) error {
	if err := nils.Check(ctx, c, db); err != nil {
		const format = "record links reset: %w"
		return fmt.Errorf(format, err)
	}
	key := c.FormValue(editorKey)
	youtube := c.FormValue("artifact-editor-youtubeval")
	demozooVal := c.FormValue("artifact-editor-demozooval")
	pouetVal := c.FormValue("artifact-editor-pouetval")
	colors16 := c.FormValue("artifact-editor-16colorstval")
	github := c.FormValue("artifact-editor-githubval")
	rels := c.FormValue("artifact-editor-relationsval")
	sites := c.FormValue("artifact-editor-websitesval")
	const format = "record links reset %w: %q"
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf(format+": %q", ErrKey, err, key))
	}

	const requirement = 11
	if len(youtube) != 0 && len(youtube) != requirement {
		return badRequest(c, fmt.Errorf(format, ErrYouTube, youtube))
	}
	colors16 = form.SanitizeURLPath(colors16)
	github = form.SanitizeGitHub(github)

	var demozooID int64
	if demozooVal != "" {
		demozooID, err = strconv.ParseInt(demozooVal, 10, 64)
		const format = "the demozoo production id %s, %v: %w"
		if err != nil {
			return badRequest(c, fmt.Errorf(format, "must be an int", demozooVal, err))
		}
		if demozooID > demozoo.Sanity {
			return badRequest(c, fmt.Errorf(format, "does not exist", demozooID, err))
		}
	}

	var pouetID int64
	if pouetVal != "" {
		const format = "the pouet production id %s, %v: %w"
		pouetID, err = strconv.ParseInt(pouetVal, 10, 64)
		if err != nil {
			return badRequest(c, fmt.Errorf(format, "must be an int", pouetVal, err))
		}
		if pouetID > pouet.Sanity {
			return badRequest(c, fmt.Errorf(format, "does not exist", pouetID, err))
		}
	}

	lnks := model.Links{
		ID:        int64(id),
		Demozoo:   demozooID,
		Pouet:     pouetID,
		YouTube:   youtube,
		Colors16:  colors16,
		GitHub:    github,
		Relations: rels,
		Sites:     sites,
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err := lnks.Update(ctx, tx); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	links := app.LinkPreviews(youtube, demozooVal, pouetVal, colors16, github, rels, sites)
	for i, link := range links {
		links[i] = `<small><strong>Link to</strong></small> &nbsp; ` + link
	}
	return c.HTML(http.StatusOK, strings.Join(links, "<br>"))
}

func recordEmulateRAM(ctx context.Context, c *echo.Context, db *sql.DB, name string) error {
	const format = "record emulate ram: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	id, err := ID(c)
	if err != nil {
		return badRequest(c, err)
	}
	value := c.FormValue(name) == "on"

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	switch name {
	case "emulate-ram-umb":
		err = model.EmulateUMB.Update(ctx, tx, int64(id), value)
	case "emulate-ram-ems":
		err = model.EmulateEMS.Update(ctx, tx, int64(id), value)
	case "emulate-ram-xms":
		err = model.EmulateXMS.Update(ctx, tx, int64(id), value)
	}
	if err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	return c.String(http.StatusOK, successSpan)
}

func RecordEmulateUMB(ctx context.Context, c *echo.Context, db *sql.DB) error {
	return recordEmulateRAM(ctx, c, db, "emulate-ram-umb")
}

func RecordEmulateEMS(ctx context.Context, c *echo.Context, db *sql.DB) error {
	return recordEmulateRAM(ctx, c, db, "emulate-ram-ems")
}

func RecordEmulateXMS(ctx context.Context, c *echo.Context, db *sql.DB) error {
	return recordEmulateRAM(ctx, c, db, "emulate-ram-xms")
}

// RecordEmulateBroken handles the patch submission for the broken emulation for a file artifact.
func RecordEmulateBroken(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record emulate broken: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	id, err := ID(c)
	if err != nil {
		return badRequest(c, err)
	}
	value := c.FormValue("emulate-is-broken") != "on"

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err = model.EmulateBroken.Update(ctx, tx, int64(id), value); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	return c.String(http.StatusOK, successSpan)
}

// RecordEmulateRunProgram handles the patch submission for the run program emulation.
func RecordEmulateRunProgram(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record emulate run prgram: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	idValue, err := ID(c)
	if err != nil {
		return badRequest(c, err)
	}
	const id = `emulate-run-program-feedback`
	const invalid = `d-block invalid-feedback`
	const success = `text-success`
	const div = `<div id="%s" class="%s">%s</div>`
	toggleValue := strings.ToUpper(c.FormValue("emulate-run-program"))
	if !jsdos.Valid(toggleValue) {
		s := `The command, or name contains invalid characters; or is too long`
		return c.String(http.StatusOK,
			fmt.Sprintf(div, id, invalid, s))
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err = model.UpdateEmulateRunProgram(ctx, tx, int64(idValue), toggleValue); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	if toggleValue == "" {
		s := `✓ Custom command(s) removed`
		return c.String(http.StatusOK,
			fmt.Sprintf(div, id, success, s))
	}
	s := `✓ Custom command(s) saved`
	return c.String(http.StatusOK,
		fmt.Sprintf(div, id, success, s))
}

// RecordEmulateMachine handles the patch submission for the machine and graphic emulation for a file artifact.
func RecordEmulateMachine(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record emulate machine: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}

	id, err := ID(c)
	if err != nil {
		return badRequest(c, err)
	}

	value := c.FormValue("emulate-machine")

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.UpdateEmulateMachine(ctx, tx, int64(id), value); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	return c.String(http.StatusOK, successSpan)
}

// RecordEmulateCPU handles the patch submission for the CPU emulation for a file artifact.
func RecordEmulateCPU(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record emulate cpu: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	id, err := ID(c)
	if err != nil {
		return badRequest(c, err)
	}

	value := c.FormValue("emulate-cpu")

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.UpdateEmulateCPU(ctx, tx, int64(id), value); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	return c.String(http.StatusOK, successSpan)
}

// RecordEmulateSFX handles the patch submission for the audio emulation for a file artifact.
func RecordEmulateSFX(ctx context.Context, c *echo.Context, db *sql.DB) error {
	const format = "record emulate sfx: %w"
	if err := nils.Check(ctx, c, db); err != nil {
		return fmt.Errorf(format, err)
	}
	id, err := ID(c)
	if err != nil {
		return badRequest(c, err)
	}

	value := c.FormValue("emulate-sfx")

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.UpdateEmulateSfx(ctx, tx, int64(id), value); err != nil {
		return badRequest(c, err)
	}
	if err = tx.Commit(); err != nil {
		return badRequest(c, err)
	}

	return c.String(http.StatusOK, successSpan)
}

// badRequest returns an error response with a 400 status code,
// the server cannot or will not process the request due to something that is perceived to be a client error.
func badRequest(c *echo.Context, err error) error {
	return c.String(http.StatusBadRequest, err.Error())
}

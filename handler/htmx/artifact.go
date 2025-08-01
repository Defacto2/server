package htmx

// Package file artifact.go provides functions for handling the HTMX requests for the artifact editor.

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/handler/form"
	"github.com/Defacto2/server/handler/jsdos"
	"github.com/Defacto2/server/handler/pouet"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var (
	ErrIsDir = errors.New("the file is a directory")
	ErrName  = errors.New("the file name is invalid")
	ErrPath  = errors.New("the file path is invalid")
	ErrYT    = errors.New("youtube watch video id needs to be empty or 11 characters")
)

const (
	checkMark = "&#x2713;"
	editorKey = "artifact-editor-key"
)

// Validate checks that the path is not absolute and does not allow
// for any path traversal. It returns an error if the path is invalid.
func Validate(path string) error {
	path = strings.TrimSuffix(path, "/")
	if filepath.IsAbs(path) {
		return fmt.Errorf("%w: %q", ErrPath, path)
	}
	if s := filepath.Clean(path); s != path {
		return fmt.Errorf("%w: %q", ErrPath, path)
	}
	return nil
}

// Path returns the uuid and directory path, the named unid, plus the path from the URL parameters.
// It returns an error if the unid or name is invalid.
func Path(c echo.Context) (string, string, error) {
	unid := c.Param("unid")
	if err := uuid.Validate(unid); err != nil {
		return "", "", err
	}
	path := c.Param("path")
	name, err := url.QueryUnescape(path)
	if err != nil {
		return "", "", err
	}
	if err := Validate(name); err != nil {
		return "", "", err
	}
	return unid, name, nil
}

// UUID returns the uuid from the URL parameters and returns an error if it is invalid.
func UUID(c echo.Context) (string, error) {
	unid := c.Param("unid")
	if err := uuid.Validate(unid); err != nil {
		return "", err
	}
	return unid, nil
}

// ID returns the id from the URL parameters and returns an error if it is invalid.
func ID(c echo.Context) (int, error) {
	key := c.Param("id")
	id, err := strconv.Atoi(key)
	if err != nil {
		return 0, fmt.Errorf("%w: %w: %q", ErrKey, err, key)
	}
	return id, nil
}

// pageRefresh is a helper function to set the HTTP [HTMX header] for the browser to refresh the page.
//
// [HTMX header]: https://htmx.org/reference/#response_headers
func pageRefresh(c echo.Context) echo.Context { //nolint:ireturn
	c.Response().Header().Set("HX-Refresh", "true")
	c.Response().WriteHeader(http.StatusFound)
	return c
}

// RecordThumb handles the htmx request for the thumbnail quality.
func RecordThumb(c echo.Context, thumb command.Thumb, dirs command.Dirs) error {
	unid, err := UUID(c)
	if err != nil {
		return badRequest(c, err)
	}
	err = dirs.Thumbs(unid, thumb)
	if errors.Is(err, command.ErrNoImages) {
		return c.String(http.StatusOK, err.Error())
	}
	if err != nil {
		return badRequest(c, err)
	}
	c = pageRefresh(c)
	return c.String(http.StatusOK,
		`Thumb created, the browser will refresh.`)
}

// RecordThumbAlignment handles the htmx request for the thumbnail crop alignment.
func RecordThumbAlignment(c echo.Context, align command.Align, dirs command.Dirs) error {
	unid, err := UUID(c)
	if err != nil {
		return badRequest(c, err)
	}
	err = align.Thumbs(unid, dirs.Preview, dirs.Thumbnail)
	if errors.Is(err, command.ErrNoImages) {
		return c.String(http.StatusOK, err.Error())
	}
	if err != nil {
		return badRequest(c, err)
	}
	c = pageRefresh(c)
	return c.String(http.StatusOK,
		`Thumb realigned, the browser will refresh.`)
}

// RecordImageCropper handles the htmx request for the preview image cropping.
func RecordImageCropper(c echo.Context, crop command.Crop, dirs command.Dirs) error {
	unid, err := UUID(c)
	if err != nil {
		return badRequest(c, err)
	}
	err = crop.Images(unid, dirs.Preview)
	if errors.Is(err, command.ErrNoImages) {
		return c.String(http.StatusOK, err.Error())
	}
	if err != nil {
		return badRequest(c, err)
	}
	c = pageRefresh(c)
	return c.String(http.StatusOK,
		`Images cropped, the browser will refresh.`)
}

// RecordImageCopier handles the htmx request to use an image file artifact as a preview.
func RecordImageCopier(c echo.Context, debug *zap.SugaredLogger, dirs command.Dirs) error {
	unid, name, err := Path(c)
	if err != nil {
		return badRequest(c, err)
	}
	name = filepath.Clean(name)
	tmp, err := helper.MkContent(unid)
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
	if err := dirs.PictureImager(debug, src, unid); err != nil {
		return badRequest(c, err)
	}
	c = pageRefresh(c)
	return c.String(http.StatusOK,
		`Images copied, the browser will refresh.`)
}

// RecordReadmeImager handles the htmx request to use the text file artifact as a preview.
func RecordReadmeImager(c echo.Context, logger *zap.SugaredLogger, amigaFont bool, dirs command.Dirs) error {
	unid, name, err := Path(c)
	if err != nil {
		return badRequest(c, err)
	}
	name = filepath.Clean(name)
	tmp, err := helper.MkContent(unid)
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
	if err := dirs.TextImager(logger, src, unid, amigaFont); err != nil {
		return badRequest(c, err)
	}
	c = pageRefresh(c)
	return c.String(http.StatusOK,
		`Text filed imaged, the browser will refresh.`)
}

// RecordDizCopier handles the htmx request to use the file_id.diz artifact as a preview.
func RecordDizCopier(c echo.Context, dirs command.Dirs) error {
	unid, name, err := Path(c)
	if err != nil {
		return badRequest(c, err)
	}
	name = filepath.Clean(name)
	tmp, err := helper.MkContent(unid)
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
	dst := filepath.Join(dirs.Extra.Path(), unid+".diz")
	if _, err = helper.DuplicateOW(src, dst); err != nil {
		return badRequest(c, err)
	}
	c = pageRefresh(c)
	return c.String(http.StatusOK,
		`DIZ copied, the browser will refresh.`)
}

func RecordReadmeCopier(c echo.Context, dirs command.Dirs) error {
	unid, name, err := Path(c)
	if err != nil {
		return badRequest(c, err)
	}
	tmp, err := helper.MkContent(unid)
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
	dst := filepath.Join(dirs.Extra.Path(), unid+".txt")
	if _, err = helper.DuplicateOW(src, dst); err != nil {
		return badRequest(c, err)
	}
	if !helper.File(filepath.Join(dirs.Thumbnail.Path(), unid+".png")) &&
		!helper.File(filepath.Join(dirs.Thumbnail.Path(), unid+".webp")) {
		if err := dirs.TextImager(nil, src, unid, false); err != nil {
			return badRequest(c, err)
		}
	}
	c = pageRefresh(c)
	return c.String(http.StatusOK,
		`Images copied, the browser will refresh.`)
}

// RecordReadmeDisable handles the htmx request to disable the in
// page display of both the text files readme and file_id.diz for the file artifact.
func RecordReadmeDisable(c echo.Context, db *sql.DB) error {
	id, err := ID(c)
	if err != nil {
		return badRequest(c, err)
	}
	value := c.FormValue("readme-is-off") != "on"
	if err = model.UpdateReadmeDisable(db, int64(id), value); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "<span class=\"text-success\">✓</span>")
}

// RecordImagePixelator handles the htmx request to pixelate both the preview and
// thumbnails, if they are not suitable for a general audience. This also has an
// added benefit of reducing the file sizes of both images and reducing page load.
func RecordImagePixelator(c echo.Context, directory ...dir.Directory) error {
	unid, err := UUID(c)
	if err != nil {
		return badRequest(c, err)
	}
	dirs := dir.Paths(directory...)
	if err := command.ImagesPixelate(unid, dirs...); err != nil {
		return badRequest(c, err)
	}
	// do not use pageRefresh as it returns an error
	// c = pageRefresh(c)
	return c.String(http.StatusOK,
		`Images pixelated, the browser will refresh.`)
}

// RecordImagesDeleter handles the request to remove the uuid named
// image files from the directories provided.
func RecordImagesDeleter(c echo.Context, directories ...dir.Directory) error {
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
func RecordDizDeleter(c echo.Context, extra dir.Directory) error {
	return extrasDeleter(c, ".diz", extra)
}

// RecordReadmeDeleter handles the request to remove the uuid named readme text file
// from the provided extra directory.
func RecordReadmeDeleter(c echo.Context, extra dir.Directory) error {
	return extrasDeleter(c, ".txt", extra)
}

func extrasDeleter(c echo.Context, ext string, extra dir.Directory) error {
	unid, err := UUID(c)
	if err != nil {
		return badRequest(c, err)
	}
	dst := filepath.Join(extra.Path(), unid+ext)
	dst = filepath.Clean(dst)
	st, err := os.Stat(dst)
	if errors.Is(err, os.ErrNotExist) {
		return c.NoContent(http.StatusOK)
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
	c = pageRefresh(c)
	return c.NoContent(http.StatusOK)
}

// RecordToggle handles the post submission for the file artifact record toggle.
// The return value is either "online" or "offline" depending on the state.
func RecordToggle(c echo.Context, db *sql.DB, state bool) error {
	key := c.FormValue(editorKey)
	id, err := strconv.ParseInt(key, 10, 64)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	if state {
		if err := model.UpdateOnline(db, id); err != nil {
			return fmt.Errorf("artifact record toggle online: %w", err)
		}
		return c.String(http.StatusOK, "online")
	}
	if err := model.UpdateOffline(db, id); err != nil {
		return fmt.Errorf("artifact record toggle offline: %w", err)
	}
	return c.String(http.StatusOK, "offline")
}

// RecordToggleByID handles the post submission for the file artifact record toggle.
// The key string is converted into an integer and used as the artifact id.
// The return value is either "online" or "offline" depending on the state.
func RecordToggleByID(c echo.Context, db *sql.DB, key string, state bool) error {
	id, err := strconv.ParseInt(key, 10, 64)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	if state {
		if err := model.UpdateOnline(db, id); err != nil {
			return fmt.Errorf("artifact record toggle by id online: %w", err)
		}
		return c.String(http.StatusOK, "Record is visible to the public.")
	}
	if err := model.UpdateOffline(db, id); err != nil {
		return fmt.Errorf("artifact record toggle by id offline: %w", err)
	}
	return c.String(http.StatusOK, "🚫 Record is disabled and hidden from public access. 🚫")
}

// RecordClassification handles the post submission for the file artifact classifications,
// such as the platform, operating system, section or category tags.
// The return value is either the humanized and counted classification or an error.
func RecordClassification(c echo.Context, db *sql.DB, logger *zap.SugaredLogger) error {
	section := c.FormValue("artifact-editor-categories")
	platform := c.FormValue("artifact-editor-operatingsystem")
	key := c.FormValue(editorKey)
	if invalid := section == "" || platform == ""; invalid {
		html, err := form.HumanizeCount(db, section, platform)
		if err != nil {
			logger.Error(err)
			return badRequest(c, err)
		}
		return c.HTML(http.StatusOK, string(html)+" did not update")
	}
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	if err := model.UpdateClassification(db, int64(id), platform, section); err != nil {
		return badRequest(c, err)
	}
	html, err := form.HumanizeCount(db, section, platform)
	if err != nil {
		logger.Error(err)
		return badRequest(c, err)
	}
	return c.HTML(http.StatusOK, string(html))
}

// RecordFilename handles the post submission for the file artifact filename.
func RecordFilename(c echo.Context, db *sql.DB) error {
	name := c.FormValue("artifact-editor-filename")
	key := c.FormValue(editorKey)
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	name = form.SanitizeFilename(name)
	if err := model.UpdateFilename(db, int64(id), name); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

// RecordFilenameReset handles the post submission for the file artifact filename reset.
func RecordFilenameReset(c echo.Context, db *sql.DB) error {
	val := c.FormValue("artifact-editor-filename-undo")
	key := c.FormValue(editorKey)
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	if err := model.UpdateFilename(db, int64(id), val); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, val)
}

// RecordVirusTotal handles the post submission for the file artifact VirusTotal report link.
func RecordVirusTotal(c echo.Context, db *sql.DB) error {
	link := c.FormValue("artifact-editor-virustotal")
	key := c.FormValue(editorKey)
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	if !form.ValidVT(link) {
		return c.NoContent(http.StatusNoContent)
	}
	if err := model.UpdateVirusTotal(db, int64(id), link); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

// RecordTitle handles the post submission for the file artifact title.
func RecordTitle(c echo.Context, db *sql.DB) error {
	title := c.FormValue("artifact-editor-title")
	key := c.FormValue(editorKey)
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	if err := model.UpdateTitle(db, int64(id), title); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

// RecordTitleReset handles the post submission for the file artifact title reset.
func RecordTitleReset(c echo.Context, db *sql.DB) error {
	val := c.FormValue("artifact-editor-titleundo")
	key := c.FormValue(editorKey)
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	if err := model.UpdateTitle(db, int64(id), val); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, val)
}

// RecordComment handles the post submission for the file artifact comment.
func RecordComment(c echo.Context, db *sql.DB) error {
	comment := c.FormValue("artifact-editor-comment")
	key := c.FormValue(editorKey)
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	if err := model.UpdateComment(db, int64(id), comment); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

// RecordCommentReset handles the post submission for the file artifact comment reset.
func RecordCommentReset(c echo.Context, db *sql.DB) error {
	val := c.FormValue("artifact-editor-comment-resetter")
	key := c.FormValue(editorKey)
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	if err := model.UpdateComment(db, int64(id), val); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Undo comment")
}

// RecordReleasers handles the post submission for the file artifact releasers.
// It will only update the releaser1 and the releaser2 values if they have changed.
// The return value is either "Updated" or "Update" depending on if the values have changed.
func RecordReleasers(c echo.Context, db *sql.DB) error {
	val1 := c.FormValue("releaser1")
	val2 := c.FormValue("releaser2")
	rel1 := c.FormValue("artifact-editor-releaser1")
	rel2 := c.FormValue("artifact-editor-releaser2")
	key := c.FormValue(editorKey)
	unchanged := (rel1 == val1 && rel2 == val2)
	if unchanged {
		return c.NoContent(http.StatusNoContent)
	}
	if err := recordReleases(db, rel1, rel2, key); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Save")
}

// RecordReleasersReset handles the post submission for the file artifact releasers reset.
// It will always reset and save the releaser1 and the releaser2 values.
// The return value is always "Resetted" unless an error occurs.
func RecordReleasersReset(c echo.Context, db *sql.DB) error {
	val1 := c.FormValue("releaser1")
	val2 := c.FormValue("releaser2")
	rel1 := c.FormValue("artifact-editor-releaser1")
	rel2 := c.FormValue("artifact-editor-releaser2")
	key := c.FormValue(editorKey)
	unchanged := (rel1 == val1 && rel2 == val2)
	if unchanged {
		return c.String(http.StatusNoContent, "")
	}
	if err := recordReleases(db, val1, val2, key); err != nil {
		return badRequest(c, err)
	}
	return c.HTML(http.StatusOK, checkMark)
}

func recordReleases(db *sql.DB, rel1, rel2, key string) error {
	id, err := strconv.Atoi(key)
	if err != nil {
		return fmt.Errorf("%w: %w: %q", ErrKey, err, key)
	}
	val := rel1
	if rel2 != "" {
		val = rel1 + "+" + rel2
	}
	if err := model.UpdateReleasers(db, int64(id), val); err != nil {
		return fmt.Errorf("model.UpdateReleasers: %w", err)
	}
	return nil
}

// RecordDateIssued handles the post submission for the file artifact date of release.
func RecordDateIssued(c echo.Context, db *sql.DB) error {
	year := c.FormValue("artifact-editor-year")
	month := c.FormValue("artifact-editor-month")
	day := c.FormValue("artifact-editor-day")
	key := c.FormValue(editorKey)
	yearval := c.FormValue("artifact-editor-yearval")
	monthval := c.FormValue("artifact-editor-monthval")
	dayval := c.FormValue("artifact-editor-dayval")
	if year == yearval && month == monthval && day == dayval {
		return c.NoContent(http.StatusNoContent)
	}
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	y, m, d := form.ValidDate(year, month, day)
	if !y || !m || !d {
		return badRequest(c, fmt.Errorf("%w, date failed to validate: Y %q %v ; M %q %v ; D %q %v ",
			ErrFormat, year, y, month, m, day, d))
	}
	if err := model.UpdateDateIssued(db, int64(id), year, month, day); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Save")
}

// RecordDateIssuedReset handles the post submission for the file artifact date of release reset.
func RecordDateIssuedReset(c echo.Context, db *sql.DB, elmID string) error {
	reset := c.FormValue(elmID)
	key := c.FormValue(editorKey)
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	vals := strings.Split(reset, "-")
	const expected = 3
	if len(vals) != expected {
		return badRequest(c, fmt.Errorf("%w, record date issued reset requires YYYY-MM-DD", ErrFormat))
	}
	year, month, day := vals[0], vals[1], vals[2]
	y, m, d := form.ValidDate(year, month, day)
	if !y || !m || !d {
		return badRequest(c, fmt.Errorf("%w, record date issued reset requires YYYY-MM-DD", ErrFormat))
	}
	if err := model.UpdateDateIssued(db, int64(id), year, month, day); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, " "+checkMark)
}

// RecordCreatorText handles the post submission for the file artifact creator text.
func RecordCreatorText(c echo.Context, db *sql.DB) error {
	creator := c.FormValue("artifact-editor-credittext")
	key := c.FormValue(editorKey)
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	val := creatorFix(creator)
	if err := model.UpdateCreatorText(db, int64(id), val); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

// RecordCreatorIll handles the post submission for the file artifact creator illustrator.
func RecordCreatorIll(c echo.Context, db *sql.DB) error {
	creator := c.FormValue("artifact-editor-creditill")
	key := c.FormValue(editorKey)
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	val := creatorFix(creator)
	if err := model.UpdateCreatorIll(db, int64(id), val); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

// RecordCreatorProg handles the post submission for the file artifact creator programmer.
func RecordCreatorProg(c echo.Context, db *sql.DB) error {
	creator := c.FormValue("artifact-editor-creditprog")
	key := c.FormValue(editorKey)
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	val := creatorFix(creator)
	if err := model.UpdateCreatorProg(db, int64(id), val); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

// RecordCreatorAudio handles the post submission for the file artifact creator musician.
func RecordCreatorAudio(c echo.Context, db *sql.DB) error {
	creator := c.FormValue("artifact-editor-creditaudio")
	key := c.FormValue(editorKey)
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	val := creatorFix(creator)
	if err := model.UpdateCreatorAudio(db, int64(id), val); err != nil {
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
func RecordCreatorReset(c echo.Context, db *sql.DB) error {
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
		return badRequest(c, fmt.Errorf("%w, record creator reset requires string;string;string;string",
			ErrFormat))
	}
	text := vals[0]
	ill := vals[1]
	prog := vals[2]
	audio := vals[3]
	if textval == text && illval == ill && progval == prog && audioval == audio {
		return c.NoContent(http.StatusNoContent)
	}
	if err := model.UpdateCreators(db, int64(id), text, ill, prog, audio); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Undo creators")
}

// RecordYouTube handles the post submission for the file artifact YouTube watch video link.
func RecordYouTube(c echo.Context, db *sql.DB) error {
	key := c.FormValue(editorKey)
	newVideo := strings.TrimSpace(c.FormValue("artifact-editor-youtube"))
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	const requirement = 11
	if len(newVideo) != 0 && len(newVideo) != requirement {
		return c.NoContent(http.StatusNoContent)
	}
	if err := model.UpdateYouTube(db, int64(id), newVideo); err != nil {
		return badRequest(c, err)
	}
	return RecordLinks(c)
}

// RecordDemozoo handles the post submission for the file artifact Demozoo production link.
func RecordDemozoo(c echo.Context, db *sql.DB) error {
	key := c.FormValue(editorKey)
	newProd := c.FormValue("artifact-editor-demozoo")
	if newProd == "" {
		newProd = "0"
	}
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	if err := model.UpdateDemozoo(db, int64(id), newProd); err != nil {
		return badRequest(c, err)
	}
	return RecordLinks(c)
}

// RecordPouet handles the post submission for the file artifact Pouet production link.
func RecordPouet(c echo.Context, db *sql.DB) error {
	key := c.FormValue(editorKey)
	newProd := c.FormValue("artifact-editor-pouet")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	if err := model.UpdatePouet(db, int64(id), newProd); err != nil {
		return badRequest(c, err)
	}
	return RecordLinks(c)
}

// Record16Colors handles the post submission for the file artifact 16 Colors link.
func Record16Colors(c echo.Context, db *sql.DB) error {
	key := c.FormValue(editorKey)
	newURL := c.FormValue("artifact-editor-16colors")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	link := form.SanitizeURLPath(newURL)
	if err := model.Update16Colors(db, int64(id), link); err != nil {
		return badRequest(c, err)
	}
	return RecordLinks(c)
}

// RecordGitHub handles the post submission for the file artifact GitHub repository link.
func RecordGitHub(c echo.Context, db *sql.DB) error {
	key := c.FormValue(editorKey)
	newRepo := c.FormValue("artifact-editor-github")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	link := form.SanitizeGitHub(newRepo)
	if err := model.UpdateGitHub(db, int64(id), link); err != nil {
		return badRequest(c, err)
	}
	return RecordLinks(c)
}

// RecordRelations handles the post submission for the file artifact releaser relationships.
func RecordRelations(c echo.Context, db *sql.DB) error {
	key := c.FormValue(editorKey)
	newRelations := c.FormValue("artifact-editor-relations")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	if err := model.UpdateRelations(db, int64(id), newRelations); err != nil {
		return badRequest(c, err)
	}
	return RecordLinks(c)
}

// RecordSites handles the post submission for the file artifact website links.
func RecordSites(c echo.Context, db *sql.DB) error {
	key := c.FormValue(editorKey)
	newSites := c.FormValue("artifact-editor-websites")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	if err := model.UpdateSites(db, int64(id), newSites); err != nil {
		return badRequest(c, err)
	}
	return RecordLinks(c)
}

// RecordLinks handles the post submission for a form submission to provide the
// HTML formatted links for the "Links" section of the artifact editor.
func RecordLinks(c echo.Context) error {
	youtube := c.FormValue("artifact-editor-youtube")
	demozoo := c.FormValue("artifact-editor-demozoo")
	pouet := c.FormValue("artifact-editor-pouet")
	colors16 := c.FormValue("artifact-editor-16colors")
	github := c.FormValue("artifact-editor-github")
	rels := c.FormValue("artifact-editor-relations")
	sites := c.FormValue("artifact-editor-websites")
	links := app.LinkPreviews(youtube, demozoo, pouet, colors16, github, rels, sites)
	for i, link := range links {
		links[i] = "<small><strong>Link to</strong></small> &nbsp; " + link
	}
	return c.HTML(http.StatusOK, strings.Join(links, "<br>"))
}

// RecordLinksReset handles the post submission for the file artifact links reset.
func RecordLinksReset(c echo.Context, db *sql.DB) error {
	key := c.FormValue(editorKey)
	youtube := c.FormValue("artifact-editor-youtubeval")
	demozooS := c.FormValue("artifact-editor-demozooval")
	pouetS := c.FormValue("artifact-editor-pouetval")
	colors16 := c.FormValue("artifact-editor-16colorstval")
	github := c.FormValue("artifact-editor-githubval")
	rels := c.FormValue("artifact-editor-relationsval")
	sites := c.FormValue("artifact-editor-websitesval")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("record links reset, %w: %w: %q", ErrKey, err, key))
	}

	const requirement = 11
	if len(youtube) != 0 && len(youtube) != requirement {
		return badRequest(c, fmt.Errorf("record links reset, %w: %q", ErrYT, youtube))
	}
	colors16 = form.SanitizeURLPath(colors16)
	github = form.SanitizeGitHub(github)

	var demozooI int64
	if demozooS != "" {
		demozooI, err = strconv.ParseInt(demozooS, 10, 64)
		if err != nil {
			return badRequest(c, fmt.Errorf("the demozoo production id must be an int, %w: %q", err, demozooS))
		}
		if demozooI > demozoo.Sanity {
			return badRequest(c, fmt.Errorf("the demozoo production id doesn't exist, %w: %q", err, demozooI))
		}
	}

	var pouetI int64
	if pouetS != "" {
		pouetI, err = strconv.ParseInt(pouetS, 10, 64)
		if err != nil {
			return badRequest(c, fmt.Errorf("the pouet production id must be an int, %w: %q", err, pouetS))
		}
		if pouetI > pouet.Sanity {
			return badRequest(c, fmt.Errorf("the pouet production id doesn't exist, %w: %q", err, pouetI))
		}
	}

	if err := model.UpdateLinks(db,
		int64(id), youtube, colors16, github, rels, sites, demozooI, pouetI); err != nil {
		return badRequest(c, err)
	}
	links := app.LinkPreviews(youtube, demozooS, pouetS, colors16, github, rels, sites)
	for i, link := range links {
		links[i] = "<small><strong>Link to</strong></small> &nbsp; " + link
	}
	return c.HTML(http.StatusOK, strings.Join(links, "<br>"))
}

func recordEmulateRAM(c echo.Context, db *sql.DB, name string) error {
	id, err := ID(c)
	if err != nil {
		return badRequest(c, err)
	}
	value := c.FormValue(name) == "on"
	switch name {
	case "emulate-ram-umb":
		err = model.UpdateEmulateUMB(db, int64(id), value)
	case "emulate-ram-ems":
		err = model.UpdateEmulateEMS(db, int64(id), value)
	case "emulate-ram-xms":
		err = model.UpdateEmulateXMS(db, int64(id), value)
	}
	if err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "<span class=\"text-success\">✓</span>")
}

func RecordEmulateUMB(c echo.Context, db *sql.DB) error {
	return recordEmulateRAM(c, db, "emulate-ram-umb")
}

func RecordEmulateEMS(c echo.Context, db *sql.DB) error {
	return recordEmulateRAM(c, db, "emulate-ram-ems")
}

func RecordEmulateXMS(c echo.Context, db *sql.DB) error {
	return recordEmulateRAM(c, db, "emulate-ram-xms")
}

// RecordEmulateBroken handles the patch submission for the broken emulation for a file artifact.
func RecordEmulateBroken(c echo.Context, db *sql.DB) error {
	id, err := ID(c)
	if err != nil {
		return badRequest(c, err)
	}
	value := c.FormValue("emulate-is-broken") != "on"
	if err = model.UpdateEmulateBroken(db, int64(id), value); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "<span class=\"text-success\">✓</span>")
}

// RecordEmulateRunProgram handles the patch submission for the run program emulation.
func RecordEmulateRunProgram(c echo.Context, db *sql.DB) error {
	id, err := ID(c)
	if err != nil {
		return badRequest(c, err)
	}
	value := strings.ToUpper(c.FormValue("emulate-run-program"))
	if !jsdos.Valid(value) {
		return c.HTML(http.StatusOK, "<div id=\"emulate-run-program-feedback\" class=\"d-block invalid-feedback\">"+
			"The command or name contains invalid characters, syntax or is too long</div>")
	}
	if err = model.UpdateEmulateRunProgram(db, int64(id), value); err != nil {
		return badRequest(c, err)
	}
	if value == "" {
		return c.String(http.StatusOK, "<div id=\"emulate-run-program-feedback\" class=\"text-success\">"+
			"✓ Custom command(s) removed</div>")
	}
	return c.String(http.StatusOK, "<div id=\"emulate-run-program-feedback\" class=\"text-success\">"+
		"✓ Command(s) saved</div>")
}

// RecordEmulateMachine handles the patch submission for the machine and graphic emulation for a file artifact.
func RecordEmulateMachine(c echo.Context, db *sql.DB) error {
	id, err := ID(c)
	if err != nil {
		return badRequest(c, err)
	}
	value := c.FormValue("emulate-machine")
	if err := model.UpdateEmulateMachine(db, int64(id), value); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "<span class=\"text-success\">✓</span>")
}

// RecordEmulateCPU handles the patch submission for the CPU emulation for a file artifact.
func RecordEmulateCPU(c echo.Context, db *sql.DB) error {
	id, err := ID(c)
	if err != nil {
		return badRequest(c, err)
	}
	value := c.FormValue("emulate-cpu")
	if err := model.UpdateEmulateCPU(db, int64(id), value); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "<span class=\"text-success\">✓</span>")
}

// RecordEmulateSFX handles the patch submission for the audio emulation for a file artifact.
func RecordEmulateSFX(c echo.Context, db *sql.DB) error {
	id, err := ID(c)
	if err != nil {
		return badRequest(c, err)
	}
	value := c.FormValue("emulate-sfx")
	if err := model.UpdateEmulateSfx(db, int64(id), value); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "<span class=\"text-success\">✓</span>")
}

// badRequest returns an error response with a 400 status code,
// the server cannot or will not process the request due to something that is perceived to be a client error.
func badRequest(c echo.Context, err error) error {
	return c.String(http.StatusBadRequest, err.Error())
}

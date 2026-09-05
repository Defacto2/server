package htmx

// Package file submit.go provides functions for handling the HTMX requests for uploading files.

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/archive"
	"github.com/Defacto2/helper"
	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/handler/form"
	"github.com/Defacto2/server/handler/sess"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/model/fix"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// UPCount handles the post submission for the Uploader classification,
// such as the platform, operating system, section or category tags.
// The return value is either the humanized and counted classification or an error.
func UPCount(sl *slog.Logger, c *echo.Context, db *sql.DB, name string) error {
	const msg = "htmx upcount"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	section := c.FormValue(name + "-category")

	platform := c.FormValue(name + "-operatingsystem")
	if platform == "" {
		platform = c.FormValue(name + "-operating-system")
	}

	ctx := c.Request().Context()
	html, err := form.HumanizeCount(ctx, db, section, platform)
	if err != nil {
		sl.Error(msg+" could not create the html template", slog.Any("error", err))
		return badRequest(c, err)
	}

	return c.HTML(http.StatusOK, string(html))
}

// UPSHA384 is a handler for the /uploader/sha384 route. It checks the SHA-384 hash
// against the database to see if the file already exists, and returns the URI if it does.
// Otherwise, if it does not exist, it returns an empty string.
func UPSHA384(sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	const msg = "htmx upsha384"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	const code = http.StatusBadRequest
	hash := c.Param("hash")
	if hash == "" {
		return c.String(code, "empty hash error")
	}

	const pattern = "^[a-fA-F0-9]{96}$"
	match, err := regexp.MatchString(pattern, hash)
	if err != nil {
		slog.Error(msg+" could not run the regular expression pattern",
			slog.String("pattern", pattern), slog.String("hash", hash), slog.Any("error", err))
		return c.String(code, "regex match error")
	}
	if !match {
		return c.String(code, "invalid hash error: "+hash)
	}

	ctx := c.Request().Context()
	uri, err := model.OneByHash(ctx, db, hash)
	if err != nil {
		slog.Error(msg+" database could not lookup the hash", slog.Any("error", err))
		return c.String(http.StatusServiceUnavailable, "cannot confirm the hash with the database")
	}

	return c.String(http.StatusOK, uri)
}

// UPImage is a handler for the /uploader/image route.
func UPImage(sl *slog.Logger, c *echo.Context, tx *sql.Tx, download dir.Directory) error {
	const key = "uploader-image"
	c.Set(key+"-operating-system", tags.Image.String())
	return Transfer{Key: key, Download: download}.Submit(sl, c, tx)
}

// UPIntro is a handler for the /uploader/intro route.
func UPIntro(sl *slog.Logger, c *echo.Context, tx *sql.Tx, download dir.Directory) error {
	const key = "uploader-intro"
	c.Set(key+"-category", tags.Intro.String())
	return Transfer{Key: key, Download: download}.Submit(sl, c, tx)
}

// UPMagazine is a handler for the /uploader/magazine route.
func UPMagazine(sl *slog.Logger, c *echo.Context, tx *sql.Tx, download dir.Directory) error {
	const key = "uploader-magazine"
	c.Set(key+"-category", tags.Mag.String())
	return Transfer{Key: key, Download: download}.Submit(sl, c, tx)
}

// UPText is a handler for the /uploader/text route.
func UPText(sl *slog.Logger, c *echo.Context, tx *sql.Tx, download dir.Directory) error {
	const key = "uploader-trainer" // FIX: Incorrect?
	return Transfer{Key: key, Download: download}.Submit(sl, c, tx)
}

// UPTrainer is a handler for the /uploader/trainer route.
func UPTrainer(sl *slog.Logger, c *echo.Context, tx *sql.Tx, download dir.Directory) error {
	const key = "uploader-trainer"
	return Transfer{Key: key, Download: download}.Submit(sl, c, tx)
}

// UPAdvanced is a handler for the /uploader/advanced route.
func UPAdvanced(sl *slog.Logger, c *echo.Context, tx *sql.Tx, download dir.Directory) error {
	const key = "uploader-advanced"
	return Transfer{Key: key, Download: download}.Submit(sl, c, tx)
}

// Transfer is a generic file transfer handler that uploads and validates a chosen file upload.
// The provided key is the name of the form input field.
type Transfer struct {
	Key      string
	Download dir.Directory
}

func (t Transfer) Submit(sl *slog.Logger, c *echo.Context, tx *sql.Tx) error {
	const format = "htmx submit %s: %w"
	const msg = "htmx submit"
	if err := nils.Check(sl, c, tx); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	// check directory before transferring
	if err := t.Download.Check(sl); err != nil {
		return errUpload(c, err)
	}

	name := t.Key + "file"
	fileHeader, err := c.FormFile(name)
	if err != nil {
		return errFormHeader(sl, c, name, err)
	}

	fileMultipart, err := fileHeader.Open()
	if err != nil {
		return errMultipartFile(sl, c, name, err)
	}

	defer Close(sl, msg, fileMultipart)

	hasher := sha512.New384()
	const size = 4 * 1024
	buf := make([]byte, size)
	if _, err := io.CopyBuffer(hasher, fileMultipart, buf); err != nil {
		return errSHA384(sl, c, name, err)
	}

	sha384 := hasher.Sum(nil)
	ctx := c.Request().Context()
	exist, err := model.ExistSHA(ctx, tx, sha384)
	if err != nil {
		return errExistSHA(sl, c, err)
	}
	if exist {
		const thanks = "<p>Thanks, but the chosen file already exists on Defacto2.</p>"
		return c.HTML(http.StatusOK, thanks+html.EscapeString(fileHeader.Filename))
	}

	tempDest, err := SaveTemp(sl, c, t.Key, fileHeader)
	if err != nil {
		return fmt.Errorf(format, "save to temp", err)
	}
	if tempDest == "" {
		return c.HTML(http.StatusInternalServerError, "The temporary save cannot be created")
	}

	content := t.content(ctx, sl, msg, fileHeader.Filename, tempDest)

	readme := archive.Readme(fileHeader.Filename, content...)

	insert := Insert{ //nolint:exhaustruct_v5
		FileHeader: fileHeader,
		Readme:     readme,
		Key:        t.Key,
		Content:    content,
	}
	copy(insert.Checksum[:], sha384)
	id, uid, err := insert.Upload(sl, c, tx)
	if err != nil {
		// resync the files table sequence if the insert failed and try again
		if err := fix.SyncFilesIDSeq(tx); err != nil {
			return c.HTML(http.StatusInternalServerError, err.Error())
		}
		// second attempt
		id, uid, err = insert.Upload(sl, c, tx)
		if err != nil {
			return c.HTML(http.StatusInternalServerError, err.Error())
		}
	} else if id == 0 {
		return nil
	}

	defer t.CopyTemp(sl, uid, tempDest)

	return t.StatusOK(c, fileHeader.Filename, id)
}

func (t Transfer) StatusOK(c *echo.Context, fname string, id int64) error {
	if err := nils.Check(c); err != nil {
		return fmt.Errorf("htmx submit success: %w", err)
	}

	const code = http.StatusOK
	if sess.Editor(c) {
		html := `<div data-bs-toggle="tooltip" data-bs-placement="top" ` +
			`data-bs-title="ctrl + alt + enter"><a id="go-to-the-new-artifact-record" ` +
			`href="/f/` + helper.ObfuscateID(id) + `" autofocus>Go to the new artifact record</a>.</div>`
		return c.HTML(code, html)
	}

	html := `<div>Thanks, the chosen file submission was a success.<br> ` +
		`<span class="text-success">✓</span> <var>` + html.EscapeString(fname) + `</var></div>`
	return c.HTML(code, html)
}

// CopyTemp copies the uploaded file to the downloads directory store.
// The fname UUID needs be provided as a unique identifier filename.
// The tempDest is the saved, temporary file that was uploaded.
func (t Transfer) CopyTemp(sl *slog.Logger, fname uuid.UUID, tempDest string) {
	const msg = "htmx transfer duplication"
	if sl == nil {
		sl = logs.Discard()
	}

	logErr := func(s string, err error) {
		args := []any{slog.String("name", tempDest), slog.String("uuid", fname.String())}
		if err != nil {
			args = append(args, slog.Any("error", err))
		}
		sl.Error(msg+" "+s, args...)
	}

	if fname.String() == "" {
		sl.Error(msg+" uuid is in an invalid syntax or empty",
			slog.String("uuid", fname.String()))
		return
	}

	st, err := os.Stat(tempDest)
	if err != nil {
		logErr("cannot stat file", err)
		return
	}
	if st.IsDir() {
		logErr("source file is a directory", nil)
		return
	}

	newpath := t.Download.Join(fname.String())
	i, err := helper.Duplicate(tempDest, newpath)
	if err != nil {
		logErr("cannot duplicate file", err)
		return
	}

	sl.Info(msg, slog.String("okay", "uploader transfer to the destination directory"),
		slog.String("uuid", fname.String()), slog.Int64("bytes_tranfered", i))
}

func (t Transfer) content(ctx context.Context, sl *slog.Logger, msg, fname, tempDest string) []string {
	content, err := archive.Lists(ctx, tempDest)
	if err != nil {
		sl.Info(msg+" archive list caused an error",
			slog.String("src", tempDest),
			slog.String("filename", fname),
			slog.Any("error", err))
	}
	return content
}

// SaveTemp is a generic file writer that saves the chosen file upload to a temporary file.
// Note the multipart file header needs to be closed outside of this func.
func SaveTemp(sl *slog.Logger, c *echo.Context, key string, fileHeader *multipart.FileHeader) (string, error) {
	const msg = "htmx save temp"
	if err := nils.Check(sl, c, fileHeader); err != nil {
		return "", fmt.Errorf(msg+": %w", err)
	}

	internalErr := func(html string, err error) (string, error) {
		sl.Error(msg+" "+strings.ToLower(html), slog.String("named_file", key+"file"),
			slog.Any("error", err))
		return "", c.HTML(http.StatusInternalServerError, html)
	}

	// open uploaded file
	fileMultipart, err := fileHeader.Open()
	if err != nil {
		return internalErr("The chosen file input cannot be opened", err)
	}

	defer Close(sl, msg, fileMultipart)

	// create temporary destination file
	const pattern = "upload-*.zip"
	tempDest, err := dir.CreateTemp(pattern) // FIX:
	if err != nil {
		return internalErr("The temporary save cannot be created", err)
	}

	defer Close(sl, msg, tempDest)

	// buffer copier
	const size = 4 * 1024
	buf := make([]byte, size)
	if _, err = io.CopyBuffer(tempDest, fileMultipart, buf); err != nil {
		return internalErr("The temporary save cannot be written", err)
	}

	return tempDest.Name(), nil
}

type Insert struct {
	FileHeader *multipart.FileHeader
	Readme     string
	Key        string
	Content    []string
	Checksum   [48]byte
}

func (in Insert) Upload(sl *slog.Logger, c *echo.Context, tx *sql.Tx) (int64, uuid.UUID, error) {
	const format = "htmx insert upload %s: %w"
	if err := nils.Check(sl, c, tx, in.FileHeader); err != nil {
		return 0, uuid.UUID{}, fmt.Errorf(format, "check", err)
	}

	uploadErr := func(err, logErr error) (int64, uuid.UUID, error) {
		logFormErr(sl, "insert upload", in.Key, in.FileHeader.Filename, logErr)
		return 0, uuid.UUID{}, err
	}

	values, err := c.FormValues()
	if err != nil {
		return uploadErr(ErrFormRead, err)
	}

	prefix := in.Key
	values.Add(prefix+"-filename", in.FileHeader.Filename)
	values.Add(prefix+"-integrity", hex.EncodeToString(in.Checksum[:]))
	values.Add(prefix+"-size", strconv.FormatInt(in.FileHeader.Size, 10))
	values.Add(prefix+"-content", strings.Join(in.Content, "\n"))
	values.Add(prefix+"-readme", in.Readme)

	checks := [...]string{
		prefix + "-operating-system",
		prefix + "-category",
	}
	for _, key := range checks {
		if values.Get(key) != "" {
			continue
		}
		value, ok := c.Get(key).(string)
		if ok {
			values.Add(key, value)
		}
	}

	ctx := c.Request().Context()
	id, unid, err := model.InsertUpload(ctx, tx, values, prefix)
	if err != nil {
		return uploadErr(ErrFormInsert, err)
	}

	return id, unid, nil
}

type Submit struct {
	Download  dir.Directory
	Extra     dir.Directory
	Preview   dir.Directory
	Thumbnail dir.Directory
}

func (u Submit) Image(sl *slog.Logger, c *echo.Context) error {
	const msg = "htmx submit image"
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	if err := u.Preview.Check(sl); err != nil {
		return errUpload(c, err)
	}

	if err := u.Thumbnail.Check(sl); err != nil {
		return errUpload(c, err)
	}

	formData := Record{unid: "", key: "", platform: "", id: 0}
	if s := formData.FormData(c); s != "" {
		return c.HTML(http.StatusBadRequest, s)
	}

	const name = "artifact-editor-replace-preview"
	fileHeader, err := c.FormFile(name)
	if err != nil {
		return errFormHeader(sl, c, name, err)
	}

	tempDest, err := SaveTemp(sl, c, name, fileHeader)
	if err != nil || tempDest == "" {
		return c.HTML(http.StatusInternalServerError, "The temporary save cannot be copied")
	}

	dirs := command.Dirs{Download: "", Preview: u.Preview, Thumbnail: u.Thumbnail, Extra: ""}
	fileMultipart, err := fileHeader.Open()
	if err != nil {
		return errMultipartFile(sl, c, name, err)
	}
	defer Close(sl, msg, fileMultipart)

	magic := magicnumber.Find(fileMultipart)

	const timeout = command.CmdTimeout * 2
	ctx, cancel := context.WithTimeout(c.Request().Context(), timeout)
	defer cancel()

	if useImager(magic) {
		if err := dirs.PictureImager(ctx, sl, tempDest, formData.unid); err != nil {
			html := "The uploaded image file could not be converted. " +
				"Try converting it first on your local machine," +
				"either to a PNG or JPG format:\n" + err.Error()
			return c.HTML(http.StatusBadRequest, html)
		}

		return u.StatusOK(c, fileHeader.Filename)
	}

	if useTexter(magic) {
		useAmiga := strings.EqualFold(formData.platform, tags.TextAmiga.String())
		useConsole := strings.EqualFold(formData.platform, tags.Console.String())
		amigaFont := useAmiga || useConsole
		err = dirs.TextImager(ctx, sl, tempDest, formData.unid, amigaFont)
		if err != nil {
			return badRequest(c, err)
		}

		return u.StatusOK(c, fileHeader.Filename)
	}

	return c.HTML(http.StatusBadRequest, "The chosen file is not a valid image or text file")
}

func useImager(magic magicnumber.Signature) bool {
	x := magicnumber.Images()
	slices.Sort(x)
	return slices.Contains(x, magic)
}

func useTexter(magic magicnumber.Signature) bool {
	x := magicnumber.Texts()
	slices.Sort(x)
	return slices.Contains(x, magic)
}

func (u Submit) StatusOK(c *echo.Context, filename string) error {
	if err := nils.Check(c); err != nil {
		const format = "transfer reloader: %w"
		return fmt.Errorf(format, err)
	}

	const format = "The submitted image %s is in use, about to reload this page"
	return c.String(http.StatusOK, fmt.Sprintf(format, filename))
}

// Replacement is the file transfer handler that uploads, validates a new file upload
// and updates the existing artifact record with the new file information.
func (u Submit) Replacement(sl *slog.Logger, c *echo.Context, tx *sql.Tx) error { //nolint:funlen
	const msg = "htmx submix replacement"
	if err := nils.Check(sl, c, tx); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	if err := u.Download.Check(sl); err != nil {
		return errUpload(c, err)
	}

	upload := Record{unid: "", key: "", platform: "", id: 0}
	if s := upload.FormData(c); s != "" {
		return c.HTML(http.StatusBadRequest, s)
	}

	const name = "artifact-editor-replace-file"
	fileHeader, err := c.FormFile(name)
	if err != nil {
		return errFormHeader(sl, c, name, err)
	}

	fileMultipart, err := fileHeader.Open()
	if err != nil {
		return errMultipartFile(sl, c, name, err)
	}

	defer Close(sl, msg, fileMultipart)

	replacement := model.FileUpload{
		ID:          upload.id,
		Filename:    fileHeader.Filename,
		Filesize:    fileHeader.Size,
		LastMod:     time.Time{},
		Integrity:   "",
		MagicNumber: "",
		Content:     "",
	}

	hasher := sha512.New384()
	const size = 4 * 1024
	buf := make([]byte, size)
	if _, err := io.CopyBuffer(hasher, fileMultipart, buf); err != nil {
		return errSHA384(sl, c, name, err)
	}
	replacement.Integrity = hex.EncodeToString(hasher.Sum(nil))

	if _, err := fileMultipart.Seek(0, io.SeekStart); err != nil {
		return errMultipartFile(sl, c, name, err)
	}

	lastmod := c.FormValue("artifact-editor-lastmodified")
	lm, err := strconv.ParseInt(lastmod, 10, 64)
	if err == nil && lm >= 0 {
		lmod := time.UnixMilli(lm)
		replacement.LastMod = lmod
	}

	sign := magicnumber.Find(fileMultipart)
	replacement.MagicNumber = sign.Title()
	destTemp, err := SaveTemp(sl, c, upload.key, fileHeader)
	if err != nil || destTemp == "" {
		return c.HTML(http.StatusInternalServerError, "The temporary save cannot be copied")
	}

	ctx := c.Request().Context()
	if list, err := archive.Lists(ctx, destTemp); err == nil {
		replacement.Content = strings.Join(list, "\n")
	}

	if err := replacement.Update(ctx, tx); err != nil {
		return badRequest(c, fmt.Errorf("replacement upload %w: %w", ErrFormUpdate, err))
	}

	destDownload := filepath.Join(u.Download.Path(), upload.unid)
	if _, err = helper.DuplicateOW(destTemp, destDownload); err != nil {
		// TODO: test the name for a manual rollback
		return badRequest(c, err)
	}

	defer func() {
		// cleanup of the redundant, repacked zipfile stored as an extra
		extra, err := os.OpenRoot(u.Extra.Path())
		if err != nil {
			slog.Warn(msg+" cannot open extras path", slog.String("path", u.Extra.Path()))
			return
		}
		name := upload.unid + ".zip"
		err = extra.Remove(name)
		if err != nil {
			slog.Warn(msg+" cannot remove extras file", slog.String("name", name))
			return
		}
	}()

	s := "The replacement file " + fileHeader.Filename + " is ready, about to reload this page"
	return c.String(http.StatusOK, s)
}

func Close(sl *slog.Logger, msg string, file multipart.File) {
	if sl == nil {
		sl = slog.Default()
	}
	if file == nil {
		sl.Error(msg + " attempted to close an empty multipart file")
	}
	if err := file.Close(); err != nil {
		sl.Info(msg+" could not close file", slog.Any("file", file),
			slog.Any("error", err))
	}
}

type Record struct {
	unid     string
	key      string
	platform string
	id       int64
}

// FormData reads the form values from the context and validates the unique identifier and record key.
// The returned value should be blank, otherwise there is an error.
func (rec *Record) FormData(c *echo.Context) string {
	if err := nils.Check(c); err != nil {
		const format = "The editor file upload is broken, %s"
		return fmt.Sprintf(format, err)
	}

	const invalid = "The editor file upload unique identifier is invalid"

	s := c.FormValue("artifact-editor-unid")
	if err := form.Checkname(s); err != nil {
		return invalid
	}
	if err := uuid.Validate(s); err != nil {
		return invalid
	}

	key := c.FormValue("artifact-editor-record-key")
	id, err := strconv.ParseInt(key, 10, 64)
	if err != nil {
		return "The editor file upload record key is invalid"
	}

	rec.unid = s
	rec.key = key
	rec.id = id
	rec.platform = c.FormValue("artifact-editor-download-classify")

	return ""
}

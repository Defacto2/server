package htmx

// Package file transfer.go provides functions for handling the HTMX requests for uploading files.

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"errors"
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
	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/handler/form"
	"github.com/Defacto2/server/handler/pouet"
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

var (
	ErrFormRead   = errors.New("form parameters could not be read")
	ErrFormInsert = errors.New("form submission could not be inserted into the database")
	ErrFormUpdate = errors.New("form submission could not update the database record")
)

const (
	dz = "demozoo"
	pt = "pouet"
)

// HumanizeCount handles the post submission for the Uploader classification,
// such as the platform, operating system, section or category tags.
// The return value is either the humanized and counted classification or an error.
func HumanizeCount(ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB, name string) error {
	const msg = "transfer humanized count"
	if err := nils.Check(ctx, sl, c, db); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	section := c.FormValue(name + "-category")

	platform := c.FormValue(name + "-operatingsystem")
	if platform == "" {
		platform = c.FormValue(name + "-operating-system")
	}

	html, err := form.HumanizeCount(ctx, db, section, platform)
	if err != nil {
		sl.Error(msg+" could not create the html template", slog.Any("error", err))
		return badRequest(c, err)
	}
	return c.HTML(http.StatusOK, string(html))
}

// LookupSHA384 is a handler for the /uploader/sha384 route. It checks the SHA-384 hash
// against the database to see if the file already exists, and returns the URI if it does.
// Otherwise, if it does not exist, it returns an empty string.
func LookupSHA384(ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB) error {
	const msg = "transfer lookup sha384"
	if err := nils.Check(ctx, sl, c, db); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	hash := c.Param("hash")
	if hash == "" {
		return c.String(http.StatusBadRequest, "empty hash error")
	}

	const pattern = "^[a-fA-F0-9]{96}$"
	match, err := regexp.MatchString(pattern, hash)
	if err != nil {
		slog.Error(msg+" could not run the regular expression pattern",
			slog.String("pattern", pattern), slog.String("hash", hash), slog.Any("error", err))
		return c.String(http.StatusBadRequest, "regex match error")
	}
	if !match {
		return c.String(http.StatusBadRequest, "invalid hash error: "+hash)
	}

	uri, err := model.OneByHash(ctx, db, hash)
	if err != nil {
		slog.Error(msg+" database could not lookup the hash", slog.Any("error", err))
		return c.String(http.StatusServiceUnavailable, "cannot confirm the hash with the database")
	}

	return c.String(http.StatusOK, uri)
}

// ImageSubmit is a handler for the /uploader/image route.
func ImageSubmit(ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB, download dir.Directory) error {
	const key = "uploader-image"
	c.Set(key+"-operating-system", tags.Image.String())
	t := Transfer{Key: key, Download: download}
	return t.transfer(ctx, sl, c, db)
}

// IntroSubmit is a handler for the /uploader/intro route.
func IntroSubmit(ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB, download dir.Directory) error {
	const key = "uploader-intro"
	c.Set(key+"-category", tags.Intro.String())
	t := Transfer{Key: key, Download: download}
	return t.transfer(ctx, sl, c, db)
}

// MagazineSubmit is a handler for the /uploader/magazine route.
func MagazineSubmit(ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB, download dir.Directory) error {
	const key = "uploader-magazine"
	c.Set(key+"-category", tags.Mag.String())
	t := Transfer{Key: key, Download: download}
	return t.transfer(ctx, sl, c, db)
}

// TextSubmit is a handler for the /uploader/text route.
func TextSubmit(ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB, download dir.Directory) error {
	t := Transfer{Key: "uploader-trainer", Download: download}
	return t.transfer(ctx, sl, c, db)
}

// TrainerSubmit is a handler for the /uploader/trainer route.
func TrainerSubmit(ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB, download dir.Directory) error {
	t := Transfer{Key: "uploader-trainer", Download: download}
	return t.transfer(ctx, sl, c, db)
}

// AdvancedSubmit is a handler for the /uploader/advanced route.
func AdvancedSubmit(ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB, download dir.Directory) error {
	t := Transfer{Key: "uploader-advanced", Download: download}
	return t.transfer(ctx, sl, c, db)
}

func uploader(err error) string {
	if err != nil {
		return "The uploader cannot save your file to the host system"
	}

	return ""
}

// Transfer is a generic file transfer handler that uploads and validates a chosen file upload.
// The provided key is the name of the form input field.
type Transfer struct {
	Key      string
	Download dir.Directory
}

func (t Transfer) transfer(ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB) error { //nolint:funlen
	const msg = "transfer file handler"
	if err := nils.Check(ctx, sl, c, db); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	if err := t.Download.Check(sl); err != nil {
		return c.HTML(http.StatusInternalServerError, uploader(err))
	}

	name := t.Key + "file"
	formFile, err := c.FormFile(name)
	if err != nil {
		return checkFormFile(sl, c, name, err)
	}
	src, err := formFile.Open()
	if err != nil {
		return checkFileOpen(sl, c, name, err)
	}
	defer func() {
		logClose(sl, msg, src)
	}()

	hasher := sha512.New384()
	const size = 4 * 1024
	buf := make([]byte, size)
	if _, err := io.CopyBuffer(hasher, src, buf); err != nil {
		return checkHasher(sl, c, name, err)
	}
	checksum := hasher.Sum(nil)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return c.HTML(http.StatusInternalServerError, "The database transaction could not begin")
	}
	exist, err := model.ExistSHA(ctx, tx, checksum)
	if err != nil {
		return checkExist(sl, c, err)
	}
	if exist {
		return c.HTML(http.StatusOK,
			"<p>Thanks, but the chosen file already exists on Defacto2.</p>"+
				html.EscapeString(formFile.Filename))
	}

	dst, err := copier(sl, c, formFile, t.Key)
	if err != nil {
		return fmt.Errorf("copier: %w", err)
	}
	if dst == "" {
		return c.HTML(http.StatusInternalServerError, "The temporary save cannot be created")
	}

	content, err := archive.Lists(ctx, dst)
	if err != nil {
		sl.Info(msg+" archive list caused an error",
			slog.String("src", dst), slog.String("filename", formFile.Filename),
			slog.Any("error", err))
	}

	readme := archive.Readme(formFile.Filename, content...)

	creator := creator{
		file: formFile, readme: readme, key: t.Key, checksum: checksum, content: content,
	}
	id, uid, err := creator.insert(ctx, sl, c, tx)
	if err != nil {
		// resync the files table sequence if the insert failed and try again
		if err := fix.SyncFilesIDSeq(db); err != nil {
			return c.HTML(http.StatusInternalServerError, err.Error())
		}
		id, uid, err = creator.insert(ctx, sl, c, tx)
		if err != nil {
			return c.HTML(http.StatusInternalServerError, err.Error())
		}
	} else if id == 0 {
		return nil
	}

	defer Duplicate(sl, uid, dst, t.Download)

	return success(c, msg, formFile.Filename, id)
}

func success(c *echo.Context, msg, filename string, id int64,
) error {
	if err := nils.Check(c); err != nil {
		return fmt.Errorf("%s success: %w", msg, err)
	}

	const format = `<div>Thanks, the chosen file submission was a success.<br> ` +
		`<span class="text-success">✓</span> <var>%s</var></div>`
	html := fmt.Sprintf(format, html.EscapeString(filename))

	if sess.Editor(c) {
		const format = `<div data-bs-toggle="tooltip" data-bs-placement="top" ` +
			`data-bs-title="ctrl + alt + enter"><a id="go-to-the-new-artifact-record" ` +
			`href="/f/%s" autofocus>Go to the new artifact record</a>.</div>`
		html += fmt.Sprintf(format, helper.ObfuscateID(id))
	}

	return c.HTML(http.StatusOK, html)
}

// Duplicate copies the chosen file to the destination directory.
// The UUID needs be provided as a unique identifier for the filename.
// The source path is the temporary file that was uploaded.
// The destination directory is where the file will be copied to.
func Duplicate(sl *slog.Logger, uid uuid.UUID, src string, dst dir.Directory) {
	const msg = "htmx transfer duplication"
	if sl == nil {
		sl = logs.Discard()
	}

	logErr := func(s string, err error) {
		args := []any{slog.String("name", src), slog.String("uuid", uid.String())}
		if err != nil {
			args = append(args, slog.Any("error", err))
		}
		sl.Error(msg+" "+s, args...)
	}

	if uid.String() == "" {
		sl.Error(msg+" uuid is in an invalid syntax or empty",
			slog.String("uuid", uid.String()))
		return
	}

	st, err := os.Stat(src)
	if err != nil {
		logErr("cannot stat file", err)
		return
	}
	if st.IsDir() {
		logErr("source file is a directory", nil)
		return
	}

	newpath := dst.Join(uid.String())
	i, err := helper.Duplicate(src, newpath)
	if err != nil {
		logErr("cannot duplicate file", err)
		return
	}

	sl.Info(msg, slog.String("okay", "uploader transfer to the destination directory"),
		slog.String("uuid", uid.String()), slog.Int64("bytes_tranfered", i))
}

func formErr(sl *slog.Logger, msg, form, name string, err error) {
	if sl == nil {
		return
	}
	args := []any{
		slog.String("form", form),
	}
	if name != "" {
		args = append(args, slog.String("name", name))
	}
	if err != nil {
		args = append(args, slog.Any("error", err))
	}
	sl.Error("transfer check "+msg, args...)
}

func checkFormFile(sl *slog.Logger, c *echo.Context, name string, err error) error {
	const format = "check form file %s: %w"
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	formErr(sl, "form file", "file input caused an error", name, err)
	return c.HTML(http.StatusBadRequest,
		"The chosen file form input caused an error")
}

func checkFileOpen(sl *slog.Logger, c *echo.Context, name string, err error) error {
	const format = "check file open %s: %w"
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	formErr(sl, "file open", "file input cannot be opened", name, err)
	return c.HTML(http.StatusBadRequest,
		"The chosen file input cannot be opened")
}

func checkHasher(sl *slog.Logger, c *echo.Context, name string, err error) error {
	const format = "check hasher %s: %w"
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	formErr(sl, "hasher", "file input cannot be hashed", name, err)
	return c.HTML(http.StatusInternalServerError,
		"The chosen file input cannot be hashed")
}

func checkExist(sl *slog.Logger, c *echo.Context, err error) error {
	const format = "check exist %s: %w"
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	formErr(sl, "exist", "cannot connect", "", err)
	return c.HTML(http.StatusServiceUnavailable,
		"Cannot confirm the hash with the database")
}

// copier is a generic file writer that saves the chosen file upload to a temporary file.
func copier(sl *slog.Logger, c *echo.Context, file *multipart.FileHeader, key string) (string, error) {
	const msg = "transfer generic file copier"
	if err := nils.Check(sl, c, file); err != nil {
		return "", fmt.Errorf(msg+": %w", err)
	}

	logErr := func(task, name string, err error) {
		sl.Error(msg, slog.String("task", task),
			slog.String("named_file", name), slog.Any("error", err))
	}

	const code = http.StatusInternalServerError
	// open uploaded file
	name := key + "file"
	src, err := file.Open()
	if err != nil {
		logErr("cannot open file", name, err)
		return "", c.HTML(code, "The chosen file input cannot be opened")
	}
	defer func() {
		logClose(sl, msg, src)
	}()

	// create temporary destination file
	const pattern = "upload-*.zip"
	dst, err := dir.CreateTemp(pattern)
	if err != nil {
		logErr("cannot create temp file", name, err)
		return "", c.HTML(code, "The temporary save cannot be created")
	}
	defer func() {
		logClose(sl, msg, dst)
	}()

	// buffer copier
	const size = 4 * 1024
	buf := make([]byte, size)
	if _, err = io.CopyBuffer(dst, src, buf); err != nil {
		logErr("cannot copy to temp file", name, err)
		return "", c.HTML(code, "The temporary save cannot be written")
	}
	return dst.Name(), nil
}

type creator struct {
	file     *multipart.FileHeader
	readme   string
	key      string
	checksum []byte
	content  []string
}

func (cr creator) insert(ctx context.Context, sl *slog.Logger, c *echo.Context, tx *sql.Tx,
) (int64, uuid.UUID, error) {
	const format = "transfer creator insert %s: %w"
	empty := uuid.UUID{}
	if err := nils.Check(ctx, sl, c, tx, cr.file); err != nil {
		return 0, empty, fmt.Errorf(format, "check", err)
	}

	// form parameters
	values, err := c.FormValues()
	if err != nil {
		formErr(sl, "insert", "cannot obtain parameters", "", err)
		return 0, empty, ErrFormRead
	}

	values.Add(cr.key+"-filename", cr.file.Filename)
	values.Add(cr.key+"-integrity", hex.EncodeToString(cr.checksum))
	values.Add(cr.key+"-size", strconv.FormatInt(cr.file.Size, 10))
	values.Add(cr.key+"-content", strings.Join(cr.content, "\n"))
	values.Add(cr.key+"-readme", cr.readme)

	if os := values.Get(cr.key + "-operating-system"); os == "" {
		s, fallback := c.Get(cr.key + "-operating-system").(string)
		if fallback {
			values.Add(cr.key+"-operating-system", s)
		}
	}

	if cat := values.Get(cr.key + "-category"); cat == "" {
		s, fallback := c.Get(cr.key + "-category").(string)
		if fallback {
			values.Add(cr.key+"-category", s)
		}
	}

	// database record
	id, unid, err := model.InsertUpload(ctx, tx, values, cr.key)
	if err != nil {
		formErr(sl, "cannot insert upload", cr.key, cr.file.Filename, err)
		return 0, empty, ErrFormInsert
	}

	return id, unid, nil
}

type Submission int

const (
	Demozoo Submission = iota
	Pouet
)

func (prod Submission) String() string {
	return [...]string{dz, pt}[prod]
}

func (prod Submission) Submit( //nolint:funlen
	ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB, download dir.Directory,
) error {
	const msg = "htmx transfer submit"
	if err := nils.Check(ctx, sl, c, db); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	logErr := func(s string, err error) {
		sl.Error(msg, slog.String("problem", s), slog.Any("error", err))
	}

	name := strings.ToTitle(prod.String())
	id, err := sanitizeID(c, name, prod.String())
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logErr("database transaction cannot start", err)
		return c.String(http.StatusServiceUnavailable, "error, the database transaction could not begin")
	}

	var exist bool
	var eErr error
	switch prod {
	case Demozoo:
		exist, eErr = model.ExistDemozoo(ctx, tx, id)
	case Pouet:
		exist, eErr = model.ExistPouet(ctx, tx, id)
	}
	if eErr != nil {
		return c.String(http.StatusServiceUnavailable, "error, the database query failed")
	}
	if exist {
		return c.String(http.StatusForbidden, "error, the "+prod.String()+" key is already in use")
	}

	var key int64
	var unid string
	switch prod {
	case Demozoo:
		key, unid, err = model.InsertDemozoo(ctx, tx, id)
	case Pouet:
		key, unid, err = model.InsertPouet(ctx, tx, id)
	}
	if err != nil || key == 0 {
		logErr(fmt.Sprintf("cannot insert record id %d", id), err)
		return c.String(http.StatusServiceUnavailable, "error, the database insert failed")
	}
	if err := tx.Commit(); err != nil {
		logErr("database commit failed", err)
		return c.String(http.StatusServiceUnavailable, "error, the database commit failed")
	}

	const format = `<div class="text-success">Thanks for the submission of %s production, %d</div>`
	html := fmt.Sprintf(format, name, id)
	if sess.Editor(c) {
		uri := helper.ObfuscateID(key)
		const format = `<p data-bs-toggle="tooltip" data-bs-placement="top" data-bs-title="ctrl + alt + enter">` +
			`<a id="go-to-the-new-artifact-record" href="/f/%s" autofocus>Go to the new artifact record</a></p>`
		html += fmt.Sprintf(format, uri)
	}

	// see Download in handler/app/internal/remote/remote.go
	switch prod {
	case Demozoo:
		if err := app.GetDemozoo(ctx, sl, c, db, id, unid, download); err != nil {
			logErr("cannot fetch remote demozoo api", err)
			const format = `<p class="text-danger">error, cannot fetch the remote download linked by %s</p>`
			html += fmt.Sprintf(format, prod.String())
			return c.String(http.StatusServiceUnavailable, html)
		}
	case Pouet:
		if err := app.GetPouet(ctx, sl, c, db, id, unid, download); err != nil {
			logErr("cannot fetch remote pouet api", err)
			const format = `<p class="text-danger">error, cannot fetch the remote download linked by %s</p>`
			html += fmt.Sprintf(format, prod.String())
			return c.String(http.StatusServiceUnavailable, html)
		}
	}

	sl.Info(msg,
		slog.String("okay", "the production has been submitted"),
		slog.String("remote", name), slog.Int("new_id", id))
	return c.String(http.StatusOK, html)
}

// sanitizeID validates the production ID and ensures that it is a valid numeric value.
func sanitizeID(c *echo.Context, name, prod string) (int, error) {
	const format = "transfer sanitize id: %w"
	if err := nils.Check(c); err != nil {
		return 0, fmt.Errorf(format, err)
	}
	id, err := echo.PathParam[int](c, "id")
	if err != nil {
		return 0, c.String(http.StatusNotAcceptable,
			"The "+name+" production ID must be a numeric value")
	}
	var sanity int
	switch prod {
	case dz:
		sanity = demozoo.Sanity
	case pt:
		sanity = pouet.Sanity
	}
	if id < 1 || id > sanity {
		const format = `The %q production ID is invalid, %d`
		return 0, c.String(http.StatusNotAcceptable, fmt.Sprintf(format, name, id))
	}
	return id, nil
}

type Upload struct {
	Download  dir.Directory
	Extra     dir.Directory
	Preview   dir.Directory
	Thumbnail dir.Directory
}

func (u Upload) ImagePreview(ctx context.Context, sl *slog.Logger, c *echo.Context) error { //nolint:funlen
	const msg = "htmx upload preview"
	if err := nils.Check(ctx, sl, c); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	const code = http.StatusInternalServerError
	if err := u.Preview.Check(sl); err != nil {
		return c.HTML(code, uploader(err))
	}
	if err := u.Thumbnail.Check(sl); err != nil {
		return c.HTML(code, uploader(err))
	}
	upload := values{unid: "", key: "", platform: "", id: 0}
	if s := upload.validate(c); s != "" {
		return c.HTML(http.StatusBadRequest, s)
	}

	name := "artifact-editor-replace-preview"
	file, err := c.FormFile(name)
	if err != nil {
		return checkFormFile(nil, c, name, err)
	}
	src, err := file.Open()
	if err != nil {
		return checkFileOpen(nil, c, name, err)
	}
	defer func() {
		logClose(sl, msg, src)
	}()

	pattern := name + "-*"
	dst, err := dir.CreateTemp(pattern)
	if err != nil {
		return c.HTML(code, "The temporary save cannot be created")
	}
	defer func() {
		logClose(sl, msg, dst)
	}()

	const size = 4 * 1024
	buf := make([]byte, size)
	if _, err := io.CopyBuffer(dst, src, buf); err != nil {
		return c.HTML(code, "The temporary save cannot be written")
	}
	defer func() {
		logRemove(sl, msg, dst.Name())
	}()

	dirs := command.Dirs{Download: "", Preview: u.Preview, Thumbnail: u.Thumbnail, Extra: ""}
	src, err = file.Open()
	if err != nil {
		return checkFileOpen(nil, c, name, err)
	}
	defer func() {
		logClose(sl, msg, src)
	}()

	magic := magicnumber.Find(src)

	if imagers(magic) {
		if err := dirs.PictureImager(ctx, sl, dst.Name(), upload.unid); err != nil {
			const s = "\nThe uploaded image file could not be converted, " +
				"please try converting it on your local machine into a PNG or JPG file"
			return c.HTML(http.StatusBadRequest, err.Error()+s)
		}

		return okayReload(c, file.Filename)
	}

	if texters(magic) {
		amigaFont := strings.EqualFold(upload.platform, tags.TextAmiga.String()) ||
			strings.EqualFold(upload.platform, tags.Console.String())
		err = dirs.TextImager(ctx, nil, dst.Name(), upload.unid, amigaFont)
		if err != nil {
			return badRequest(c, err)
		}
		return okayReload(c, file.Filename)
	}

	return c.HTML(http.StatusBadRequest, "The chosen file is not a valid image or text file")
}

func imagers(magic magicnumber.Signature) bool {
	imgs := magicnumber.Images()
	slices.Sort(imgs)
	return slices.Contains(imgs, magic)
}

func texters(magic magicnumber.Signature) bool {
	txts := magicnumber.Texts()
	slices.Sort(txts)
	return slices.Contains(txts, magic)
}

func okayReload(c *echo.Context, filename string) error {
	if err := nils.Check(c); err != nil {
		const format = "transfer reloader: %w"
		return fmt.Errorf(format, err)
	}

	const format = "The new preview %s is in use, about to reload this page"
	return c.String(http.StatusOK, fmt.Sprintf(format, filename))
}

// Replacement is the file transfer handler that uploads, validates a new file upload
// and updates the existing artifact record with the new file information.
func (u Upload) Replacement(ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB) error { //nolint:funlen
	const msg = "htmx upload replacement"
	if err := nils.Check(ctx, sl, c, db); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	if err := u.Download.Check(sl); err != nil {
		return c.HTML(http.StatusInternalServerError, uploader(err))
	}
	upload := values{unid: "", key: "", platform: "", id: 0}
	if s := upload.validate(c); s != "" {
		return c.HTML(http.StatusBadRequest, s)
	}

	const name = "artifact-editor-replace-file"
	file, err := c.FormFile(name)
	if err != nil {
		return checkFormFile(sl, c, name, err)
	}
	src, err := file.Open()
	if err != nil {
		return checkFileOpen(sl, c, name, err)
	}
	defer func() {
		logClose(sl, msg, src)
	}()

	fu := model.FileUpload{
		ID:          upload.id,
		Filename:    file.Filename,
		Filesize:    file.Size,
		LastMod:     time.Time{},
		Integrity:   "",
		MagicNumber: "",
		Content:     "",
	}

	hasher := sha512.New384()
	const size = 4 * 1024
	buf := make([]byte, size)
	if _, err := io.CopyBuffer(hasher, src, buf); err != nil {
		return checkHasher(sl, c, name, err)
	}
	fu.Integrity = hex.EncodeToString(hasher.Sum(nil))
	src, err = file.Open()
	if err != nil {
		return checkFileOpen(sl, c, name, err)
	}
	defer func() {
		logClose(sl, msg, src)
	}()

	lastmod := c.FormValue("artifact-editor-lastmodified")
	lm, err := strconv.ParseInt(lastmod, 10, 64)
	if err == nil && lm >= 0 {
		lmod := time.UnixMilli(lm)
		fu.LastMod = lmod
	}

	sign := magicnumber.Find(src)
	fu.MagicNumber = sign.Title()
	dst, err := copier(sl, c, file, upload.key)
	if err != nil || dst == "" {
		return c.HTML(http.StatusInternalServerError, "The temporary save cannot be copied")
	}

	if list, err := archive.Lists(ctx, dst); err == nil {
		fu.Content = strings.Join(list, "\n")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return c.HTML(http.StatusInternalServerError, "The database transaction could not begin")
	}
	if err := fu.Update(ctx, tx); err != nil {
		return badRequest(c, fmt.Errorf("file upload update, %w: %w", ErrFormUpdate, err))
	}

	abs := filepath.Join(u.Download.Path(), upload.unid)
	if _, err = helper.DuplicateOW(dst, abs); err != nil {
		defer func() {
			rollback(sl, msg, upload.id, tx)
		}()
		return badRequest(c, err)
	}
	if err := tx.Commit(); err != nil {
		return c.HTML(http.StatusInternalServerError, "The database commit failed")
	}

	repack := filepath.Join(u.Extra.Path(), upload.unid+".zip")
	repack = filepath.Clean(repack)
	defer func() {
		logRemove(sl, msg, repack)
	}()

	const format = "The new file %s is in use, about to reload this page"
	return c.String(http.StatusOK, fmt.Sprintf(format, file.Filename))
}

func logClose(sl *slog.Logger, msg string, file multipart.File) {
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

func logRemove(sl *slog.Logger, msg, name string) {
	if sl == nil {
		sl = slog.Default()
	}
	if err := os.Remove(name); err != nil {
		sl.Info(msg+" could not remove named file", slog.String("name", name),
			slog.Any("error", err))
	}
}

type values struct {
	unid     string
	key      string
	platform string
	id       int64
}

// validate reads the form values from the context and validates the unique identifier and record key.
// The return value is an error message if the unique identifier or record key is invalid.
func (i *values) validate(c *echo.Context) string {
	if err := nils.Check(c); err != nil {
		const format = "The editor file upload is broken, %s"
		return fmt.Sprintf(format, err)
	}

	const invalid = "The editor file upload unique identifier is invalid"

	i.unid = c.FormValue("artifact-editor-unid")
	if err := form.Checkname(i.unid); err != nil {
		return invalid
	}
	if err := uuid.Validate(i.unid); err != nil {
		return invalid
	}

	i.key = c.FormValue("artifact-editor-record-key")
	id, err := strconv.ParseInt(i.key, 10, 64)
	if err != nil {
		return "The editor file upload record key is invalid"
	}

	i.id = id
	i.platform = c.FormValue("artifact-editor-download-classify")

	return ""
}

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
	"github.com/Defacto2/server/internal/panics"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/model/fix"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var (
	ErrFormRead     = errors.New("form parameters could not be read")
	ErrFormInsert   = errors.New("form submission could not be inserted into the database")
	ErrFormUpdate   = errors.New("form submission could not update the database record")
	ErrNoFileHeader = errors.New("multipart file header is nil")
)

const (
	dz       = "demozoo"
	pt       = "pouet"
	category = "-category"
)

// HumanizeCount handles the post submission for the Uploader classification,
// such as the platform, operating system, section or category tags.
// The return value is either the humanized and counted classification or an error.
func HumanizeCount(c echo.Context, db *sql.DB, sl *slog.Logger, name string) error {
	const msg = "transfer humanized count"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	section := c.FormValue(name + category)
	platform := c.FormValue(name + "-operatingsystem")
	if platform == "" {
		platform = c.FormValue(name + "-operating-system")
	}
	html, err := form.HumanizeCount(db, section, platform)
	if err != nil {
		sl.Error(msg,
			slog.String("issue", "could not create the html template"), slog.Any("error", err))
		return badRequest(c, err)
	}
	return c.HTML(http.StatusOK, string(html))
}

// LookupSHA384 is a handler for the /uploader/sha384 route. It checks the SHA-384 hash
// against the database to see if the file already exists, and returns the URI if it does.
// Otherwise, if it does not exist, it returns an empty string.
func LookupSHA384(c echo.Context, db *sql.DB, sl *slog.Logger) error {
	const msg = "transfer lookup sha384"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	hash := c.Param("hash")
	if hash == "" {
		return c.String(http.StatusBadRequest, "empty hash error")
	}
	const pattern = "^[a-fA-F0-9]{96}$"
	match, err := regexp.MatchString(pattern, hash)
	if err != nil {
		slog.Error(msg, slog.String("regexp", "could not run the pattern to string match"),
			slog.String("pattern", pattern), slog.String("hash", hash), slog.Any("error", err))
		return c.String(http.StatusBadRequest, "regex match error")
	}
	if !match {
		return c.String(http.StatusBadRequest, "invalid hash error: "+hash)
	}
	ctx := context.Background()
	uri, err := model.HashFind(ctx, db, hash)
	if err != nil {
		slog.Error(msg, slog.String("database", "could not lookup the hash"), slog.Any("error", err))
		return c.String(http.StatusServiceUnavailable,
			"cannot confirm the hash with the database")
	}
	return c.String(http.StatusOK, uri)
}

// ImageSubmit is a handler for the /uploader/image route.
func ImageSubmit(c echo.Context, db *sql.DB, sl *slog.Logger, download dir.Directory) error {
	const key = "uploader-image"
	c.Set(key+"-operating-system", tags.Image.String())
	return transfer(c, db, sl, key, download)
}

// IntroSubmit is a handler for the /uploader/intro route.
func IntroSubmit(c echo.Context, db *sql.DB, sl *slog.Logger, download dir.Directory) error {
	const key = "uploader-intro"
	c.Set(key+"-category", tags.Intro.String())
	return transfer(c, db, sl, key, download)
}

// MagazineSubmit is a handler for the /uploader/magazine route.
func MagazineSubmit(c echo.Context, db *sql.DB, sl *slog.Logger, download dir.Directory) error {
	const key = "uploader-magazine"
	c.Set(key+"-category", tags.Mag.String())
	return transfer(c, db, sl, key, download)
}

// TextSubmit is a handler for the /uploader/text route.
func TextSubmit(c echo.Context, db *sql.DB, sl *slog.Logger, download dir.Directory) error {
	const key = "uploader-text"
	return transfer(c, db, sl, key, download)
}

// TrainerSubmit is a handler for the /uploader/trainer route.
func TrainerSubmit(c echo.Context, db *sql.DB, sl *slog.Logger, download dir.Directory) error {
	const key = "uploader-trainer"
	return transfer(c, db, sl, key, download)
}

// AdvancedSubmit is a handler for the /uploader/advanced route.
func AdvancedSubmit(c echo.Context, db *sql.DB, sl *slog.Logger, download dir.Directory) error {
	const key = "uploader-advanced"
	return transfer(c, db, sl, key, download)
}

func uploader(err error) string {
	if err != nil {
		return "The uploader cannot save your file to the host system"
	}
	return ""
}

// Transfer is a generic file transfer handler that uploads and validates a chosen file upload.
// The provided name is that of the form input field. The logger is optional and if nil then
// the function will not log any debug information.
func transfer(c echo.Context, db *sql.DB, sl *slog.Logger, key string, download dir.Directory) error {
	const msg = "transfer file handler"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	if err := download.Check(sl); err != nil {
		return c.HTML(http.StatusInternalServerError, uploader(err))
	}
	name := key + "file"
	file, err := c.FormFile(name)
	if err != nil {
		return checkFormFile(c, sl, name, err)
	}
	src, err := file.Open()
	if err != nil {
		return checkFileOpen(c, sl, name, err)
	}
	defer func() { _ = src.Close() }()
	hasher := sha512.New384()
	const size = 4 * 1024
	buf := make([]byte, size)
	if _, err := io.CopyBuffer(hasher, src, buf); err != nil {
		return checkHasher(c, sl, name, err)
	}
	checksum := hasher.Sum(nil)
	ctx := context.Background()
	tx, err := db.Begin()
	if err != nil {
		return c.HTML(http.StatusInternalServerError, "The database transaction could not begin")
	}
	exist, err := model.SHA384Exists(ctx, tx, checksum)
	if err != nil {
		return checkExist(c, sl, err)
	}
	if exist {
		return c.HTML(http.StatusOK,
			"<p>Thanks, but the chosen file already exists on Defacto2.</p>"+
				html.EscapeString(file.Filename))
	}
	dst, err := copier(c, sl, file, key)
	if err != nil {
		return fmt.Errorf("copier: %w", err)
	}
	if dst == "" {
		return c.HTML(http.StatusInternalServerError, "The temporary save cannot be created")
	}
	content, _ := archive.List(dst, file.Filename)
	readme := archive.Readme(file.Filename, content...)
	creator := creator{
		file: file, readme: readme, key: key, checksum: checksum, content: content,
	}
	id, uid, err := creator.insert(ctx, c, tx, sl)
	if err != nil {
		// resync the files table sequence if the insert failed and try again
		if err := fix.SyncFilesIDSeq(db); err != nil {
			return c.HTML(http.StatusInternalServerError, err.Error())
		}
		id, uid, err = creator.insert(ctx, c, tx, sl)
		if err != nil {
			return c.HTML(http.StatusInternalServerError, err.Error())
		}
	} else if id == 0 {
		return nil
	}
	defer Duplicate(sl, uid, dst, download)
	return success(c, file.Filename, id)
}

func success(c echo.Context, filename string, id int64,
) error {
	html := fmt.Sprintf("<div>Thanks, the chosen file submission was a success.<br> "+
		"<span class=\"text-success\">✓</span> <var>%s</var></div>", html.EscapeString(filename))
	if sess.Editor(c) {
		html += fmt.Sprintf("<div data-bs-toggle=\"tooltip\" data-bs-placement=\"top\" "+
			"data-bs-title=\"ctrl + alt + enter\"><a id=\"go-to-the-new-artifact-record\" "+
			"href=\"/f/%s\" autofocus>Go to the new artifact record</a>.</div>",
			helper.ObfuscateID(id))
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
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}
	if uid.String() == "" {
		sl.Error(msg,
			slog.String("issue", "uuid is in an invalid syntax or empty"),
			slog.String("uuid", uid.String()))
		return
	}
	st, err := os.Stat(src)
	if err != nil {
		sl.Error(msg, slog.String("issue", "cannot stat the named source file"),
			slog.String("name", src),
			slog.String("uuid", uid.String()),
			slog.Any("error", err))
		return
	}
	if st.IsDir() {
		sl.Error(msg, slog.String("issue", "named source file cannot be a directory"),
			slog.String("uuid", uid.String()), slog.String("name", src))
		return
	}
	newPath := dst.Join(uid.String())
	i, err := helper.Duplicate(src, newPath)
	if err != nil {
		sl.Error(msg, slog.String("issue", "could not duplicate the file"),
			slog.String("uuid", uid.String()),
			slog.String("source file", src), slog.String("destination", newPath),
			slog.Any("error", err))
		return
	}
	sl.Info(msg,
		slog.String("success", "uploader transfer to the destination directory"),
		slog.String("uuid", uid.String()), slog.Int64("bytes tranfered", i))
}

func checkFormFile(c echo.Context, sl *slog.Logger, name string, err error) error {
	const msg = "check form file"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	sl.Error("transfer check form file",
		slog.String("form", "the file input caused an error"),
		slog.String("name", name), slog.Any("error", err))
	return c.HTML(http.StatusBadRequest,
		"The chosen file form input caused an error")
}

func checkFileOpen(c echo.Context, sl *slog.Logger, name string, err error) error {
	const msg = "check file open"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	sl.Error("transfer check file open",
		slog.String("form", "the file input could not be opened"),
		slog.String("named file", name),
		slog.Any("error", err))
	return c.HTML(http.StatusBadRequest,
		"The chosen file input cannot be opened")
}

func checkHasher(c echo.Context, sl *slog.Logger, name string, err error) error {
	const msg = "check hasher"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	sl.Error("transfer check hasher",
		slog.String("form", "the file input could not be hashed"),
		slog.String("named file", name),
		slog.Any("error", err))
	return c.HTML(http.StatusInternalServerError,
		"The chosen file input cannot be hashed")
}

func checkExist(c echo.Context, sl *slog.Logger, err error) error {
	const msg = "check exist"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	sl.Error("transfer check exist",
		slog.String("form", "could not connect"),
		slog.Any("error", err))
	return c.HTML(http.StatusServiceUnavailable,
		"Cannot confirm the hash with the database")
}

// copier is a generic file writer that saves the chosen file upload to a temporary file.
func copier(c echo.Context, sl *slog.Logger, file *multipart.FileHeader, key string) (string, error) {
	const msg = "transfer generic file copier"
	if err := panics.EchoContextS(c, sl); err != nil {
		return "", fmt.Errorf("%s: %w", msg, err)
	}
	if file == nil {
		return "", fmt.Errorf("%s: %w", msg, ErrNoFileHeader)
	}
	// open uploaded file
	const pattern = "upload-*.zip"
	name := key + "file"
	src, err := file.Open()
	if err != nil {
		sl.Error(msg,
			slog.String("task", "the file input could not be opened"),
			slog.String("named file", name),
			slog.Any("error", err))
		return "", c.HTML(http.StatusInternalServerError,
			"The chosen file input cannot be opened")
	}
	defer func() { _ = src.Close() }()
	// create temporary destination file
	dst, err := os.CreateTemp(helper.TmpDir(), pattern)
	if err != nil {
		sl.Error(msg,
			slog.String("task", "the file input could not create a temporary destination file"),
			slog.String("named file", name),
			slog.Any("error", err))
		return "", c.HTML(http.StatusInternalServerError,
			"The temporary save cannot be created")
	}
	defer func() { _ = dst.Close() }()
	// buffer copier
	const size = 4 * 1024
	buf := make([]byte, size)
	if _, err = io.CopyBuffer(dst, src, buf); err != nil {
		sl.Error(msg,
			slog.String("task", "the file input could not be copied to the temporary destination file"),
			slog.String("named file", name),
			slog.Any("error", err))
		return "", c.HTML(http.StatusInternalServerError,
			"The temporary save cannot be written")
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

func insertpanic(ctx context.Context, c echo.Context, tx *sql.Tx, sl *slog.Logger, file *multipart.FileHeader) error {
	if ctx == nil {
		return panics.ErrNoContext
	}
	if c == nil {
		return panics.ErrNoEchoC
	}
	if tx == nil {
		return panics.ErrNoTx
	}
	if sl == nil {
		return panics.ErrNoSlog
	}
	if file == nil {
		return ErrNoFileHeader
	}
	return nil
}

func (cr creator) insert(ctx context.Context, c echo.Context, tx *sql.Tx, sl *slog.Logger,
) (int64, uuid.UUID, error) {
	const msg = "transfer creator insert"
	empty := uuid.UUID{}
	if err := insertpanic(ctx, c, tx, sl, cr.file); err != nil {
		return 0, empty, fmt.Errorf("%s: %w", msg, err)
	}
	// form parameters
	values, err := c.FormParams()
	if err != nil {
		sl.Error(msg, slog.String("form", "could not obtain the form parameters"), slog.Any("error", err))
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
	id, uid, err := model.InsertUpload(ctx, tx, values, cr.key)
	if err != nil {
		sl.Error(msg, slog.String("form", "could not insert a new database record for the file upload"),
			slog.String("filename", cr.file.Filename), slog.String("cr key", cr.key),
			slog.Any("error", err))
		return 0, empty, ErrFormInsert
	}
	return id, uid, nil
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
	c echo.Context, db *sql.DB, sl *slog.Logger, download dir.Directory,
) error {
	const msg = "htmx transfer submit"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	name := strings.ToTitle(prod.String())
	id, err := sanitizeID(c, name, prod.String())
	if err != nil {
		return err
	}
	ctx := context.Background()
	tx, err := db.Begin()
	if err != nil {
		sl.Error(msg,
			slog.String("problem", "the database transaction could not start"), slog.Any("error", err))
		return c.String(http.StatusServiceUnavailable, "error, the database transaction could not begin")
	}
	var exist bool
	switch prod {
	case Demozoo:
		exist, err = model.DemozooExists(ctx, tx, id)
	case Pouet:
		exist, err = model.PouetExists(ctx, tx, id)
	}
	if err != nil {
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
		sl.Error(msg,
			slog.String("problem", "cannot insert a record to the database"),
			slog.Int("record id", id),
			slog.Any("error", err))
		return c.String(http.StatusServiceUnavailable,
			"error, the database insert failed")
	}
	if err := tx.Commit(); err != nil {
		sl.Error(msg,
			slog.String("problem", "the database transaction commit failed"), slog.Any("error", err))
		return c.String(http.StatusServiceUnavailable,
			"error, the database commit failed")
	}
	html := fmt.Sprintf("<div class=\"text-success\">Thanks for the submission of %s production, %d</div>", name, id)
	if sess.Editor(c) {
		uri := helper.ObfuscateID(key)
		html += fmt.Sprintf("<p data-bs-toggle=\"tooltip\" data-bs-placement=\"top\" data-bs-title=\"ctrl + alt + enter\">"+
			"<a id=\"go-to-the-new-artifact-record\" href=\"/f/%s\" autofocus>Go to the new artifact record</a></p>", uri)
	}
	// see Download in handler/app/internal/remote/remote.go
	switch prod {
	case Demozoo:
		if err := app.GetDemozoo(c, db, id, unid, download); err != nil {
			sl.Error(msg,
				slog.String("problem", "could not fetch the remote demozoo api"), slog.Any("error", err))
			html += fmt.Sprintf(`<p class="text-danger">error, cannot fetch the remote download linked by %s</p>`, prod.String())
			return c.String(http.StatusServiceUnavailable, html)
		}
	case Pouet:
		if err := app.GetPouet(c, db, id, unid, download); err != nil {
			sl.Error(msg,
				slog.String("problem", "could not fetch the remote pouet api"), slog.Any("error", err))
			html += fmt.Sprintf(`<p class="text-danger">error, cannot fetch the remote download linked by %s</p>`, prod.String())
			return c.String(http.StatusServiceUnavailable, html)
		}
	}
	sl.Info(msg,
		slog.String("success", "the production has been submitted"),
		slog.String("remote", name), slog.Int("new record id", id))
	return c.String(http.StatusOK, html)
}

// sanitizeID validates the production ID and ensures that it is a valid numeric value.
func sanitizeID(c echo.Context, name, prod string) (int, error) {
	if c == nil {
		return 0, panics.ErrNoEchoC
	}
	sid := c.Param("id")
	id, err := strconv.Atoi(sid)
	if err != nil {
		return 0, c.String(http.StatusNotAcceptable,
			"The "+name+" production ID must be a numeric value, "+sid)
	}
	var sanity int
	switch prod {
	case dz:
		sanity = demozoo.Sanity
	case pt:
		sanity = pouet.Sanity
	}
	if id < 1 || id > sanity {
		return 0, c.String(http.StatusNotAcceptable,
			"The "+name+" production ID is invalid, "+sid)
	}
	return id, nil
}

func UploadPreview(c echo.Context, sl *slog.Logger, preview, thumbnail dir.Directory) error {
	const msg = "htmx upload preview"
	if err := panics.EchoContextS(c, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	name := "artifact-editor-replace-preview"
	if err := preview.Check(sl); err != nil {
		return c.HTML(http.StatusInternalServerError, uploader(err))
	}
	if err := thumbnail.Check(sl); err != nil {
		return c.HTML(http.StatusInternalServerError, uploader(err))
	}
	upload := values{}
	if s := upload.formValues(c); s != "" {
		return c.HTML(http.StatusBadRequest, s)
	}
	file, err := c.FormFile(name)
	if err != nil {
		return checkFormFile(c, nil, name, err)
	}
	src, err := file.Open()
	if err != nil {
		return checkFileOpen(c, nil, name, err)
	}
	defer func() { _ = src.Close() }()
	pattern := name + "-*"
	dst, err := os.CreateTemp(helper.TmpDir(), pattern)
	if err != nil {
		return c.HTML(http.StatusInternalServerError,
			"The temporary save cannot be created")
	}
	defer func() { _ = dst.Close() }()
	const size = 4 * 1024
	buf := make([]byte, size)
	if _, err := io.CopyBuffer(dst, src, buf); err != nil {
		return c.HTML(http.StatusInternalServerError,
			"The temporary save cannot be written")
	}
	defer func() { _ = os.Remove(dst.Name()) }()
	dirs := command.Dirs{Preview: preview, Thumbnail: thumbnail}
	src, err = file.Open()
	if err != nil {
		return checkFileOpen(c, nil, name, err)
	}
	defer func() { _ = src.Close() }()
	magic := magicnumber.Find(src)
	if imagers(magic) {
		if err := dirs.PictureImager(sl, dst.Name(), upload.unid); err != nil {
			return c.HTML(http.StatusBadRequest,
				err.Error()+
					"\nThe uploaded image file could not be converted, "+
					"please try converting it on your local machine into a PNG or JPG file")
		}
		return reloader(c, file.Filename)
	}
	if texters(magic) {
		amigaFont := strings.EqualFold(upload.platform, tags.TextAmiga.String())
		err = dirs.TextImager(nil, dst.Name(), upload.unid, amigaFont)
		if err != nil {
			return badRequest(c, err)
		}
		return reloader(c, file.Filename)
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

func reloader(c echo.Context, filename string) error {
	if c == nil {
		return panics.ErrNoEchoC
	}
	return c.String(http.StatusOK,
		fmt.Sprintf("The new preview %s is in use, about to reload this page", filename))
}

// UploadReplacement is the file transfer handler that uploads, validates a new file upload
// and updates the existing artifact record with the new file information.
func UploadReplacement( //nolint:funlen
	c echo.Context, db *sql.DB, sl *slog.Logger,
	download, extra dir.Directory,
) error {
	const msg = "htmx upload replacement"
	if err := panics.EchoContextDS(c, db, sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	const name = "artifact-editor-replace-file"
	if err := download.Check(sl); err != nil {
		return c.HTML(http.StatusInternalServerError, uploader(err))
	}
	upload := values{}
	if s := upload.formValues(c); s != "" {
		return c.HTML(http.StatusBadRequest, s)
	}
	file, err := c.FormFile(name)
	if err != nil {
		return checkFormFile(c, sl, name, err)
	}
	src, err := file.Open()
	if err != nil {
		return checkFileOpen(c, sl, name, err)
	}
	defer func() { _ = src.Close() }()
	fu := model.FileUpload{Filename: file.Filename, Filesize: file.Size}
	hasher := sha512.New384()
	const size = 4 * 1024
	buf := make([]byte, size)
	if _, err := io.CopyBuffer(hasher, src, buf); err != nil {
		return checkHasher(c, sl, name, err)
	}
	fu.Integrity = hex.EncodeToString(hasher.Sum(nil))
	src, err = file.Open()
	if err != nil {
		return checkFileOpen(c, sl, name, err)
	}
	defer func() { _ = src.Close() }()
	lastmod := c.FormValue("artifact-editor-lastmodified")
	lm, err := strconv.ParseInt(lastmod, 10, 64)
	if err == nil && lm > 0 {
		lmod := time.UnixMilli(lm)
		fu.LastMod = lmod
	}
	sign := magicnumber.Find(src)
	fu.MagicNumber = sign.Title()
	dst, err := copier(c, sl, file, upload.key)
	if err != nil || dst == "" {
		return c.HTML(http.StatusInternalServerError, "The temporary save cannot be copied")
	}
	if list, err := archive.List(dst, file.Filename); err == nil {
		fu.Content = strings.Join(list, "\n")
	}
	tx, err := db.Begin()
	if err != nil {
		return c.HTML(http.StatusInternalServerError, "The database transaction could not begin")
	}
	if err := fu.Update(context.Background(), tx, upload.id); err != nil {
		return badRequest(c, fmt.Errorf("file upload update, %w: %w", ErrFormUpdate, err))
	}
	abs := filepath.Join(download.Path(), upload.unid)
	if _, err = helper.DuplicateOW(dst, abs); err != nil {
		_ = tx.Rollback()
		return badRequest(c, err)
	}
	if err := tx.Commit(); err != nil {
		return c.HTML(http.StatusInternalServerError, "The database commit failed")
	}
	repack := filepath.Join(extra.Path(), upload.unid+".zip")
	repack = filepath.Clean(repack)
	defer func() { _ = os.Remove(repack) }()
	if mkc, err := helper.MkContent(abs); err == nil {
		defer func() { _ = os.RemoveAll(mkc) }()
	}
	return c.String(http.StatusOK,
		fmt.Sprintf("The new file %s is in use, about to reload this page", file.Filename))
}

type values struct {
	unid     string
	key      string
	platform string
	id       int64
}

// formValues reads the form values from the context and validates the unique identifier and record key.
// The return value is an error message if the unique identifier or record key is invalid.
func (i *values) formValues(c echo.Context) string {
	if c == nil {
		return fmt.Sprintf("The editor file upload is broken, %s",
			panics.ErrNoEchoC)
	}
	const msg = "The editor file upload unique identifier is invalid"
	i.unid = c.FormValue("artifact-editor-unid")
	if err := form.Checkname(i.unid); err != nil {
		return msg
	}
	if err := uuid.Validate(i.unid); err != nil {
		return msg
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

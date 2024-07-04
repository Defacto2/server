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
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Defacto2/server/handler/sess"
	"github.com/Defacto2/server/internal/archive"
	"github.com/Defacto2/server/internal/demozoo"
	"github.com/Defacto2/server/internal/form"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/pouet"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var (
	ErrDir       = errors.New("cannot be a directory")
	ErrFile      = errors.New("cannot be a file")
	ErrSave      = errors.New("cannot save a file")
	ErrMultiHead = errors.New("multipart file header is nil")
	ErrUUID      = errors.New("invalid or an empty UUID")
)

const (
	dz       = "demozoo"
	pt       = "pouet"
	category = "-category"
)

// HumanizeAndCount handles the post submission for the Uploader classification,
// such as the platform, operating system, section or category tags.
// The return value is either the humanized and counted classification or an error.
func HumanizeAndCount(c echo.Context, logger *zap.SugaredLogger, name string) error {
	section := c.FormValue(name + category)
	platform := c.FormValue(name + "-operatingsystem")
	if platform == "" {
		platform = c.FormValue(name + "-operating-system")
	}
	html, err := form.HumanizeAndCount(section, platform)
	if err != nil {
		logger.Error(err)
		return badRequest(c, err)
	}
	return c.HTML(http.StatusOK, string(html))
}

// LookupSHA384 is a handler for the /uploader/sha384 route.
func LookupSHA384(c echo.Context, logger *zap.SugaredLogger) error {
	hash := c.Param("hash")
	if hash == "" {
		return c.String(http.StatusBadRequest, "empty hash error")
	}
	match, err := regexp.MatchString("^[a-fA-F0-9]{96}$", hash)
	if err != nil {
		logger.Error(err)
		return c.String(http.StatusBadRequest, "regex match error")
	}
	if !match {
		return c.String(http.StatusBadRequest, "invalid hash error: "+hash)
	}

	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		logger.Error(err)
		return c.String(http.StatusServiceUnavailable,
			"cannot connect to the database")
	}
	defer db.Close()
	exist, err := model.HashExists(ctx, db, hash)
	if err != nil {
		logger.Error(err)
		return c.String(http.StatusServiceUnavailable,
			"cannot confirm the hash with the database")
	}
	switch exist {
	case true:
		return c.String(http.StatusOK, "true")
	case false:
		return c.String(http.StatusOK, "false")
	}
	return c.String(http.StatusServiceUnavailable,
		"unexpected boolean error occurred")
}

// ImageSubmit is a handler for the /uploader/image route.
func ImageSubmit(c echo.Context, logger *zap.SugaredLogger, downloadDir string) error {
	const key = "uploader-image"
	c.Set(key+"-operating-system", tags.Image.String())
	return transfer(c, logger, key, downloadDir)
}

// IntroSubmit is a handler for the /uploader/intro route.
func IntroSubmit(c echo.Context, logger *zap.SugaredLogger, downloadDir string) error {
	const key = "uploader-intro"
	c.Set(key+"-category", tags.Intro.String())
	return transfer(c, logger, key, downloadDir)
}

// MagazineSubmit is a handler for the /uploader/magazine route.
func MagazineSubmit(c echo.Context, logger *zap.SugaredLogger, downloadDir string) error {
	const key = "uploader-magazine"
	c.Set(key+"-category", tags.Mag.String())
	return transfer(c, logger, key, downloadDir)
}

// TextSubmit is a handler for the /uploader/text route.
func TextSubmit(c echo.Context, logger *zap.SugaredLogger, downloadDir string) error {
	const key = "uploader-text"
	return transfer(c, logger, key, downloadDir)
}

// TrainerSubmit is a handler for the /uploader/trainer route.
func TrainerSubmit(c echo.Context, logger *zap.SugaredLogger, downloadDir string) error {
	const key = "uploader-trainer"
	return transfer(c, logger, key, downloadDir)
}

// AdvancedSubmit is a handler for the /uploader/advanced route.
func AdvancedSubmit(c echo.Context, logger *zap.SugaredLogger, downloadDir string) error {
	const key = "uploader-advanced"
	return transfer(c, logger, key, downloadDir)
}

// Transfer is a generic file transfer handler that uploads and validates a chosen file upload.
// The provided name is that of the form input field. The logger is optional and if nil then
// the function will not log any debug information.
func transfer(c echo.Context, logger *zap.SugaredLogger, key, downloadDir string) error {
	if s, err := checkDest(downloadDir); err != nil {
		return c.HTML(http.StatusInternalServerError, s)
	}
	name := key + "file"
	file, err := c.FormFile(name)
	if err != nil {
		return checkFormFile(c, logger, name, err)
	}
	src, err := file.Open()
	if err != nil {
		return checkFileOpen(c, logger, name, err)
	}
	defer src.Close()
	hasher := sha512.New384()
	if _, err := io.Copy(hasher, src); err != nil {
		return checkHasher(c, logger, name, err)
	}
	checksum := hasher.Sum(nil)
	ctx := context.Background()
	db, tx, err := postgres.ConnectTx()
	if err != nil {
		return c.HTML(http.StatusServiceUnavailable,
			"Cannot begin the database transaction")
	}
	defer db.Close()
	exist, err := model.SHA384Exists(ctx, tx, checksum)
	if err != nil {
		return checkExist(c, logger, err)
	}
	if exist {
		return c.HTML(http.StatusOK,
			"<p>Thanks, but the chosen file already exists on Defacto2.</p>"+
				html.EscapeString(file.Filename))
	}
	dst, err := copier(c, logger, file, key)
	if err != nil {
		return fmt.Errorf("copier: %w", err)
	}
	if dst == "" {
		return c.HTML(http.StatusInternalServerError,
			"The temporary save cannot be created")
	}
	content, _ := archive.List(dst, file.Filename)
	readme := archive.Readme(file.Filename, content...)
	creator := creator{
		file: file, readme: readme, key: key, checksum: checksum, content: content,
	}
	id, uid, err := creator.insert(ctx, c, tx, logger)
	if err != nil {
		return fmt.Errorf("creator.insert: %w", err)
	} else if id == 0 {
		return nil
	}
	defer Duplicate(logger, uid, dst, downloadDir)
	return success(c, file.Filename, id)
}

func success(c echo.Context, filename string, id int64,
) error {
	html := fmt.Sprintf("<div>Thanks, the chosen file submission was a success.<br> "+
		"<span class=\"text-success\">✓</span> <var>%s</var></div>", html.EscapeString(filename))
	if sess.Editor(c) {
		html += fmt.Sprintf("<div><a href=\"/f/%s\">Go to the new artifact record</a>.</div>",
			helper.ObfuscateID(id))
	}
	return c.HTML(http.StatusOK, html)
}

// Duplicate copies the chosen file to the destination directory.
// The UUID needs be provided as a unique identifier for the filename.
// The source path is the temporary file that was uploaded.
// The destination directory is where the file will be copied to.
func Duplicate(logger *zap.SugaredLogger, uid uuid.UUID, srcPath, dstDir string) {
	if uid.String() == "" {
		logger.Errorf("htmx transfer duplicate file: %w, %s", ErrUUID, uid)
		return
	}
	st, err := os.Stat(srcPath)
	if err != nil {
		logger.Errorf("htmx transfer duplicate file: %w, %s", err, srcPath)
		return
	}
	if st.IsDir() {
		logger.Errorf("htmx transfer duplicate file, %w: %s", ErrDir, srcPath)
		return
	}
	newPath := filepath.Join(dstDir, uid.String())
	i, err := helper.Duplicate(srcPath, newPath)
	if err != nil {
		logger.Errorf("htmx transfer duplicate file: %w,%q,  %s",
			err, uid.String(), srcPath)
		return
	}
	logger.Infof("Uploader copied %d bytes for %s, to the destination dir", i, uid.String())
}

func checkDest(dest string) (string, error) {
	st, err := os.Stat(dest)
	if err != nil {
		return "The uploader is misconfigured and cannot save your file",
			fmt.Errorf("invalid uploader destination, %w", err)
	}
	if !st.IsDir() {
		return "The uploader is misconfigured and cannot save your file",
			fmt.Errorf("invalid uploader destination, %w", ErrFile)
	}
	f, err := os.CreateTemp(dest, "uploader-*.zip")
	if err != nil {
		return "The uploader cannot save your file to the host system.",
			fmt.Errorf("%w: %w", ErrSave, err)
	}
	defer f.Close()
	defer os.Remove(f.Name())
	return "", nil
}

func checkFormFile(c echo.Context, logger *zap.SugaredLogger, name string, err error) error {
	if logger != nil {
		s := fmt.Sprintf("The chosen file input caused an error, %s: %s", name, err)
		logger.Error(s)
	}
	return c.HTML(http.StatusBadRequest,
		"The chosen file form input caused an error")
}

func checkFileOpen(c echo.Context, logger *zap.SugaredLogger, name string, err error) error {
	if logger != nil {
		s := fmt.Sprintf("The chosen file input could not be opened, %s: %s", name, err)
		logger.Error(s)
	}
	return c.HTML(http.StatusBadRequest,
		"The chosen file input cannot be opened")
}

func checkHasher(c echo.Context, logger *zap.SugaredLogger, name string, err error) error {
	if logger != nil {
		s := fmt.Sprintf("The chosen file input could not be hashed, %s: %s", name, err)
		logger.Error(s)
	}
	return c.HTML(http.StatusInternalServerError,
		"The chosen file input cannot be hashed")
}

func checkExist(c echo.Context, logger *zap.SugaredLogger, err error) error {
	if logger != nil {
		s := fmt.Sprintf("%s: %s", ErrDB, err)
		logger.Error(s)
	}
	return c.HTML(http.StatusServiceUnavailable,
		"Cannot confirm the hash with the database")
}

// copier is a generic file writer that saves the chosen file upload to a temporary file.
func copier(c echo.Context, logger *zap.SugaredLogger, file *multipart.FileHeader, key string) (string, error) {
	if file == nil {
		return "", fmt.Errorf("htmx copier: %w", ErrMultiHead)
	}
	const pattern = "upload-*.zip"
	name := key + "file"

	src, err := file.Open()
	if err != nil {
		if logger != nil {
			s := fmt.Sprintf("The chosen file input could not be opened, %s: %s", name, err)
			logger.Error(s)
		}
		return "", c.HTML(http.StatusInternalServerError,
			"The chosen file input cannot be opened")
	}
	defer src.Close()

	dst, err := os.CreateTemp("tmp", pattern)
	if err != nil {
		if logger != nil {
			s := fmt.Sprintf("Cannot create a temporary destination file, %s: %s", name, err)
			logger.Error(s)
		}
		return "", c.HTML(http.StatusInternalServerError,
			"The temporary save cannot be created")
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		if logger != nil {
			s := fmt.Sprintf("Cannot copy to the temporary destination file, %s: %s", name, err)
			logger.Error(s)
		}
		return "", c.HTML(http.StatusInternalServerError,
			"The temporary save cannot be written")
	}
	return dst.Name(), nil
}

func debug(c echo.Context, htm string) (string, error) {
	values, err := c.FormParams()
	if err != nil {
		return htm, fmt.Errorf("c.FormParams: %w", err)
	}
	htm += "<ul>"
	for k, v := range values {
		val := html.EscapeString(strings.Join(v, " "))
		htm += fmt.Sprintf("<li>%s: %s</li>", k, val)
	}
	htm += "</ul>"
	htm += "<small>The debug information is not shown in production.</small>"
	return htm, nil
}

type creator struct {
	file     *multipart.FileHeader
	readme   string
	key      string
	checksum []byte
	content  []string
}

func (cr creator) insert(ctx context.Context, c echo.Context, tx *sql.Tx, logger *zap.SugaredLogger,
) (int64, uuid.UUID, error) {
	noID := uuid.UUID{}
	values, err := c.FormParams()
	if err != nil {
		if logger != nil {
			logger.Error(err)
		}
		return 0, noID, c.HTML(http.StatusInternalServerError,
			"The form parameters could not be read")
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
	id, uid, err := model.InsertUpload(ctx, tx, values, cr.key)
	if err != nil {
		if logger != nil {
			logger.Error(err)
		}
		return 0, noID, c.HTML(http.StatusInternalServerError,
			"The form submission could not be inserted")
	}
	return id, uid, nil
}

func submit(c echo.Context, logger *zap.SugaredLogger, prod string) error {
	name := strings.ToTitle(prod)
	if logger == nil {
		return c.String(http.StatusInternalServerError,
			"error, "+prod+" submit logger is nil")
	}
	id, err := sanitizeID(c, name, prod)
	if err != nil {
		return err
	}
	ctx := context.Background()
	db, tx, err := postgres.ConnectTx()
	if err != nil {
		return fmt.Errorf("htmx submit: %w", err)
	}
	defer db.Close()
	var exist bool
	switch prod {
	case dz:
		exist, err = model.DemozooExists(ctx, tx, id)
	case pt:
		exist, err = model.PouetExists(ctx, tx, id)
	}
	if err != nil {
		return c.String(http.StatusServiceUnavailable,
			"error, the database query failed")
	}
	if exist {
		return c.String(http.StatusForbidden,
			"error, the "+prod+" key is already in use")
	}
	var key int64
	switch prod {
	case dz:
		key, err = model.InsertDemozoo(ctx, tx, id)
	case pt:
		key, err = model.InsertPouet(ctx, tx, id)
	}
	if err != nil || key == 0 {
		logger.Error(err, id)
		return c.String(http.StatusServiceUnavailable,
			"error, the database insert failed")
	}
	if err := tx.Commit(); err != nil {
		logger.Error(err)
		return c.String(http.StatusServiceUnavailable,
			"error, the database commit failed")
	}
	html := fmt.Sprintf("Thanks for the submission of %s production, %d", name, id)
	if sess.Editor(c) {
		uri := helper.ObfuscateID(key)
		html += fmt.Sprintf("<p><a href=\"/f/%s\">Go to the new artifact record</a></p>", uri)
	}
	return c.HTML(http.StatusOK, html)
}

func sanitizeID(c echo.Context, name, prod string) (int64, error) {
	sid := c.Param("id")
	id, err := strconv.ParseInt(sid, 10, 64)
	if err != nil {
		return 0, c.String(http.StatusNotAcceptable,
			"The "+name+" production ID must be a numeric value, "+sid)
	}
	var sanity uint64
	switch prod {
	case dz:
		sanity = demozoo.Sanity
	case pt:
		sanity = pouet.Sanity
	}
	if id < 1 || id > int64(sanity) {
		return 0, c.String(http.StatusNotAcceptable,
			"The "+name+" production ID is invalid, "+sid)
	}
	return id, nil
}

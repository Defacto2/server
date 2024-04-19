package htmx

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/Defacto2/server/internal/archive"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func ImageSubmit(c echo.Context, logger *zap.SugaredLogger, prod bool) error {
	if prod {
		logger = nil
	}
	const key = "uploader-image"
	c.Set(key+"-operating-system", tags.Image.String())
	return transfer(c, logger, key)
}

func IntroSubmit(c echo.Context, logger *zap.SugaredLogger, prod bool) error {
	if prod {
		logger = nil
	}
	const key = "uploader-intro"
	c.Set(key+"-category", tags.Intro.String())
	return transfer(c, logger, key)
}

func MagazineSubmit(c echo.Context, logger *zap.SugaredLogger, prod bool) error {
	if prod {
		logger = nil
	}
	const key = "uploader-magazine"
	c.Set(key+"-category", tags.Mag.String())
	return transfer(c, logger, key)
}

func TextSubmit(c echo.Context, logger *zap.SugaredLogger, prod bool) error {
	if prod {
		logger = nil
	}
	const key = "uploader-text"
	// TODO: Set the category tag.
	return transfer(c, logger, key)
}

// Transfer is a generic file transfer handler that uploads and validates a chosen file upload.
// The provided name is that of the form input field. The logger is optional and if nil then
// the function will not log any debug information.
func transfer(c echo.Context, logger *zap.SugaredLogger, key string) error {
	name := key + "file"
	file, err := c.FormFile(name)
	if err != nil {
		if logger != nil {
			s := fmt.Sprintf("The chosen file input caused an error, %s: %s", name, err)
			logger.Error(s)
		}
		return c.HTML(http.StatusBadRequest,
			"The chosen file form input caused an error.")
	}

	src, err := file.Open()
	if err != nil {
		if logger != nil {
			s := fmt.Sprintf("The chosen file input could not be opened, %s: %s", name, err)
			logger.Error(s)
		}
		return c.HTML(http.StatusBadRequest,
			"The chosen file input cannot be opened.")
	}
	defer src.Close()

	hasher := sha512.New384()
	if _, err := io.Copy(hasher, src); err != nil {
		if logger != nil {
			s := fmt.Sprintf("The chosen file input could not be hashed, %s: %s", name, err)
			logger.Error(s)
		}
		return c.HTML(http.StatusInternalServerError,
			"The chosen file input cannot be hashed.")
	}
	checksum := hasher.Sum(nil)

	db, err := postgres.ConnectDB()
	if err != nil {
		if logger != nil {
			s := fmt.Sprintf("%s: %s", ErrDB, err)
			logger.Error(s)
		}
		return c.HTML(http.StatusServiceUnavailable,
			"Cannot connect to the database.")
	}
	defer db.Close()

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return c.HTML(http.StatusServiceUnavailable,
			"Cannot begin the database transaction.")
	}
	defer tx.Rollback()

	exist, err := model.ExistSumHash(ctx, db, checksum)
	if err != nil {
		if logger != nil {
			s := fmt.Sprintf("%s: %s", ErrDB, err)
			logger.Error(s)
		}
		return c.HTML(http.StatusServiceUnavailable,
			"Cannot confirm the hash with the database.")
	}
	if exist {
		return c.HTML(http.StatusOK,
			"<p>Thanks, but the chosen file already exists on Defacto2.</p>"+
				html.EscapeString(file.Filename))
	}

	dst, err := copier(c, logger, file, key)
	if err != nil {
		return err
	}
	if dst == "" {
		return c.HTML(http.StatusInternalServerError,
			"The temporary save cannot be created.")
	}

	content, _ := archive.List(dst, file.Filename)
	readme := archive.Readme(file.Filename, content...)
	creator := creator{
		file:     file,
		readme:   readme,
		key:      key,
		checksum: checksum,
		content:  content,
	}
	id, err := creator.insert(c, logger, ctx, tx)
	if err != nil {
		return err
	}
	if id == 0 {
		return nil
	}
	return success(c, logger, file.Filename)
}

// copier is a generic file writer that saves the chosen file upload to a temporary file.
func copier(c echo.Context, logger *zap.SugaredLogger, file *multipart.FileHeader, key string) (string, error) {
	if file == nil {
		return "", ErrFileHead
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
			"The chosen file input cannot be opened.")
	}
	defer src.Close()

	dst, err := os.CreateTemp("tmp", pattern)
	if err != nil {
		if logger != nil {
			s := fmt.Sprintf("Cannot create a temporary destination file, %s: %s", name, err)
			logger.Error(s)
		}
		return "", c.HTML(http.StatusInternalServerError,
			"The temporary save cannot be created.")
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		if logger != nil {
			s := fmt.Sprintf("Cannot copy to the temporary destination file, %s: %s", name, err)
			logger.Error(s)
		}
		return "", c.HTML(http.StatusInternalServerError,
			"The temporary save cannot be written.")
	}
	return dst.Name(), nil
}

func debug(c echo.Context, html string) (string, error) {
	values, err := c.FormParams()
	if err != nil {
		return html, err
	}
	html += "<ul>"
	for k, v := range values {
		html += fmt.Sprintf("<li>%s: %s</li>", k, v)
	}
	html += "</ul>"
	html += "<small>The debug information is not shown in production.</small>"
	return html, nil
}

type creator struct {
	file     *multipart.FileHeader
	readme   string
	key      string
	checksum []byte
	content  []string
}

func (cr creator) insert(c echo.Context, logger *zap.SugaredLogger, ctx context.Context, tx *sql.Tx,
) (int64, error) {
	values, err := c.FormParams()
	if err != nil {
		if logger != nil {
			logger.Error(err)
		}
		return 0, c.HTML(http.StatusInternalServerError,
			"The form parameters could not be read.")
	}
	values.Add(cr.key+"-filename", cr.file.Filename)
	values.Add(cr.key+"-integrity", hex.EncodeToString(cr.checksum))
	values.Add(cr.key+"-size", fmt.Sprintf("%d", cr.file.Size))
	values.Add(cr.key+"-content", strings.Join(cr.content, "\n"))
	values.Add(cr.key+"-readme", cr.readme)

	id, err := model.InsertUpload(ctx, tx, values, cr.key)
	if err != nil {
		if logger != nil {
			logger.Error(err)
		}
		return 0, c.HTML(http.StatusInternalServerError,
			"The form submission could not be inserted.")
	}
	return id, nil
}

func success(c echo.Context, logger *zap.SugaredLogger, filename string) error {
	html := fmt.Sprintf("<p>Thanks, the chosen file submission was a success.<br> ✓ %s</p>",
		html.EscapeString(filename))
	if production := logger == nil; production {
		return c.HTML(http.StatusOK, html)
	}
	html, err := debug(c, html)
	if err != nil {
		return c.HTML(http.StatusOK,
			html+"<p>Could not show the form parameters and values.</p>")
	}
	return c.HTML(http.StatusOK, html)
}

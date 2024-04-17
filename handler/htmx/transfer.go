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

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Transfer is a generic file transfer handler that uploads and validates a chosen file upload.
// The provided name is that of the form input field. The logger is optional and if nil then
// the function will not log any debug information.
func transfer(c echo.Context, logger *zap.SugaredLogger, name string) error {
	file, err := c.FormFile(name)
	if err != nil {
		if logger != nil {
			logger.Error(fmt.Sprintf("The chosen file input caused an error, %s: %s", name, err))
		}
		return c.HTML(http.StatusBadRequest, "The chosen file form input caused an error.")
	}

	src, err := file.Open()
	if err != nil {
		if logger != nil {
			logger.Error(fmt.Sprintf("The chosen file input could not be opened, %s: %s", name, err))
		}
		return c.HTML(http.StatusBadRequest, "The chosen file input cannot be opened.")
	}
	defer src.Close()

	hasher := sha512.New384()
	if _, err := io.Copy(hasher, src); err != nil {
		if logger != nil {
			logger.Error(fmt.Sprintf("The chosen file input could not be hashed, %s: %s", name, err))
		}
		return c.HTML(http.StatusInternalServerError, "The chosen file input cannot be hashed.")
	}
	checksum := hasher.Sum(nil)
	fmt.Println("sha512.New384", hex.EncodeToString(checksum))

	db, err := postgres.ConnectDB()
	if err != nil {
		if logger != nil {
			logger.Error(fmt.Sprintf("%s: %s", ErrDB, err))
		}
		return c.HTML(http.StatusServiceUnavailable, "Cannot connect to the database.")
	}
	defer db.Close()

	ctx := context.Background()
	exist, err := model.ExistSumHash(ctx, db, checksum)
	if err != nil {
		if logger != nil {
			logger.Error(fmt.Sprintf("%s: %s", ErrDB, err))
		}
		return c.HTML(http.StatusServiceUnavailable, "Cannot confirm the hash with the database.")
	}
	if exist {
		return c.HTML(http.StatusOK, "<p>Thanks, but the chosen file already exists on Defacto2.</p>"+
			html.EscapeString(file.Filename))
	}
	fmt.Println("pre-copier")
	if err = copier(c, logger, file, name); err != nil {
		return err
	}
	fmt.Println("pre-creator")
	id, err := creator(c, logger, ctx, db, checksum, file.Filename)
	if err != nil {
		return err
	}
	if id == 0 {
		return nil
	}
	fmt.Println("pre-success", id, err)
	return success(c, logger, file.Filename)
}

// copier is a generic file writer that saves the chosen file upload to a temporary file.
func copier(c echo.Context, logger *zap.SugaredLogger, file *multipart.FileHeader, name string) error {
	if file == nil {
		return ErrFileHead
	}
	src, err := file.Open()
	if err != nil {
		if logger != nil {
			logger.Error(fmt.Sprintf("The chosen file input could not be opened, %s: %s", name, err))
		}
		return c.HTML(http.StatusInternalServerError, "The chosen file input cannot be opened.")
	}
	defer src.Close()

	dst, err := os.CreateTemp("tmp", "upload-*.zip")
	if err != nil {
		if logger != nil {
			logger.Error(fmt.Sprintf("Cannot create a temporary destination file, %s: %s", name, err))
		}
		return c.HTML(http.StatusInternalServerError, "The temporary save cannot be created.")
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		if logger != nil {
			logger.Error(fmt.Sprintf("Cannot copy to the temporary destination file, %s: %s", name, err))
		}
		return c.HTML(http.StatusInternalServerError, "The temporary save cannot be written.")
	}
	return nil
}

func debug(c echo.Context, html string) (string, error) {
	values, err := c.FormParams()
	if err != nil {
		return html, err
	}
	fmt.Println(values)
	html += "<ul>"
	for k, v := range values {
		html += fmt.Sprintf("<li>%s: %s</li>", k, v)
	}
	html += "</ul>"
	html += "<small>The debug information is not shown in production.</small>"
	return html, nil
}

func creator(c echo.Context, logger *zap.SugaredLogger, ctx context.Context, db *sql.DB, checksum []byte, filename string) (int64, error) {
	values, err := c.FormParams()
	if err != nil {
		if logger != nil {
			logger.Error(err)
		}
		return 0, c.HTML(http.StatusInternalServerError,
			"The form parameters could not be read.")
	}
	values.Add("filename", filename)
	values.Add("integrity", hex.EncodeToString(checksum))

	id, err := model.InsertUpload(ctx, db, values)
	if err != nil {
		if logger != nil {
			logger.Error(err)
		}
		return 0, c.HTML(http.StatusInternalServerError,
			"The form submission could not be inserted.")
	}
	// InsertUpload(ctx context.Context, db *sql.DB, values url.Values) (int64, error) {
	fmt.Println("creator", id, checksum, filename, id)
	fmt.Println(values)
	// todo insert values into form submission.
	// hex.EncodeToString(checksum) is the checksum / integrity of the file.

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
		return c.HTML(http.StatusOK, html+"<p>Could not show the form parameters and values.</p>")
	}
	return c.HTML(http.StatusOK, html)
}

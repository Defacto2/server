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
	sum := hasher.Sum(nil)
	fmt.Println("sha512.New384", hex.EncodeToString(sum))

	db, err := postgres.ConnectDB()
	if err != nil {
		if logger != nil {
			logger.Error(fmt.Sprintf("%s: %s", ErrDB, err))
		}
		return c.HTML(http.StatusServiceUnavailable, "Cannot connect to the database.")
	}
	defer db.Close()

	ctx := context.Background()
	exist, err := model.ExistSumHash(ctx, db, sum)
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
	if err = copier(c, logger, file, name); err != nil {
		return err
	}
	if err = creator(c, ctx, db, file.Filename); err != nil {
		return err
	}
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
	html += "<ul>"
	for k, v := range values {
		html += fmt.Sprintf("<li>%s: %s</li>", k, v)
	}
	html += "</ul>"
	html += "<small>The debug information is not shown in production.</small>"
	return html, nil
}

func creator(c echo.Context, _ context.Context, _ *sql.DB, filename string) error {
	_, err := c.FormParams()
	if err != nil {
		return err
	}
	return nil
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

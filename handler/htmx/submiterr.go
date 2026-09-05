package htmx

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Defacto2/server/internal/nils"
	"github.com/labstack/echo/v5"
)

func logFormErr(sl *slog.Logger, msg, form, name string, err error) {
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

func errUpload(c *echo.Context, err error) error {
	const code = http.StatusInternalServerError
	if c == nil || err == nil {
		return c.HTML(code, "unexpected error")
	}
	return c.HTML(code,
		"The uploader cannot save your submission")
}

func errFormHeader(sl *slog.Logger, c *echo.Context, name string, err error) error {
	const format = "error form header %s: %w"
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	logFormErr(sl, "form file", "file input caused an error", name, err)

	return c.HTML(http.StatusBadRequest,
		"The chosen file form input caused an error")
}

func errMultipartFile(sl *slog.Logger, c *echo.Context, name string, err error) error {
	const format = "check file open %s: %w"
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	logFormErr(sl, "file open", "file input cannot be opened", name, err)
	return c.HTML(http.StatusBadRequest,
		"The chosen file input cannot be opened")
}

func errSHA384(sl *slog.Logger, c *echo.Context, name string, err error) error {
	const format = "check hasher %s: %w"
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	logFormErr(sl, "hasher", "file input cannot be hashed", name, err)
	return c.HTML(http.StatusInternalServerError,
		"The chosen file input cannot be hashed")
}

func errExistSHA(sl *slog.Logger, c *echo.Context, err error) error {
	const format = "check exist %s: %w"
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	logFormErr(sl, "exist", "cannot connect", "", err)
	return c.HTML(http.StatusServiceUnavailable,
		"Cannot confirm the hash with the database")
}

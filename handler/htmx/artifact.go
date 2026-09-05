package htmx

// Package file artifact.go provides functions for handling the HTMX requests for the artifact editor.

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Defacto2/server/handler/form"
	"github.com/Defacto2/server/internal/nils"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

const EditorKey = "artifact-editor-key"

const (
	checkMark   = "&#x2713;"
	successSpan = `<span class="text-success">✓</span>`
)

// Path returns the uuid and directory path.
// The named unid and the path are sourced from the URL parameters.
// It returns an error if the unid or name is invalid.
func Path(c *echo.Context) (unid string, name string, err error) { //nolint:nonamedreturns
	const format = "htmx path %s: %w"
	if err := nils.Check(c); err != nil {
		return "", "", fmt.Errorf(format, "check", err)
	}

	unid = c.Param("unid")
	if err := form.Checkname(unid); err != nil {
		return "", "", fmt.Errorf(format, "invalid unid format", err)
	}
	if err := uuid.Validate(unid); err != nil {
		return "", "", fmt.Errorf(format, "invalid uuid", err)
	}

	path := c.Param("path")
	name, err = url.QueryUnescape(path)
	if err != nil {
		return "", "", fmt.Errorf(format, "failed to unescape path", err)
	}

	if err := Validate(name); err != nil {
		return "", "", err
	}

	return unid, name, nil
}

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

// UUID returns the uuid from the URL parameters and returns an error if it is invalid.
func UUID(c *echo.Context) (string, error) {
	const format = "htmx uuid %s: %w"
	if err := nils.Check(c); err != nil {
		return "", fmt.Errorf(format, "check", err)
	}

	unid := c.Param("unid")
	if err := form.Checkname(unid); err != nil {
		return "", fmt.Errorf(format, "invalid unid format", err)
	}

	if err := uuid.Validate(unid); err != nil {
		return "", fmt.Errorf(format, "invalid uuid", err)
	}

	return unid, nil
}

// PageReload is a helper function to set the HTTP [HTMX header] for the browser to refresh the page.
//
// [HTMX header]: https://htmx.org/reference/#response_headers
func PageReload(c *echo.Context) *echo.Context {
	c.Response().Header().Set("HX-Refresh", "true")
	c.Response().WriteHeader(http.StatusFound)

	return c
}

func CommitSanitize(c *echo.Context, tx *sql.Tx, name string,
	sanitize func(string) string,
	fn func(context.Context, boil.ContextExecutor, int64, string) error,
) error {
	const format = "commit sanitize %s: %w"
	if err := nils.Check(c, tx, fn); err != nil {
		return badRequest(c, fmt.Errorf(format, "check", err))
	}

	key, err := Key(c)
	if err != nil {
		return badRequest(c, fmt.Errorf(format, "key", err))
	}

	value := c.FormValue(name)
	value = sanitize(value)

	ctx := c.Request().Context()
	if err := fn(ctx, tx, key, value); err != nil {
		return badRequest(c, fmt.Errorf(format, "fn", err))
	}

	return c.String(http.StatusOK, successSpan)
}

func CommitNotOn(c *echo.Context, tx *sql.Tx, name string,
	fn func(context.Context, boil.ContextExecutor, int64, bool) error,
) error {
	const format = "commit not on %s: %w"
	if err := nils.Check(c, tx, fn); err != nil {
		return badRequest(c, fmt.Errorf(format, "check", err))
	}

	key, err := KeyParam(c)
	if err != nil {
		return badRequest(c, fmt.Errorf(format, "key", err))
	}

	value := c.FormValue(name) != "on"

	ctx := c.Request().Context()
	if err := fn(ctx, tx, key, value); err != nil {
		return badRequest(c, fmt.Errorf(format, "fn", err))
	}

	return c.String(http.StatusOK, successSpan)
}

func CommitOn(c *echo.Context, tx *sql.Tx, name string,
	fn func(context.Context, boil.ContextExecutor, int64, bool) error,
) error {
	const format = "commit on %s: %w"
	if err := nils.Check(c, tx, fn); err != nil {
		return badRequest(c, fmt.Errorf(format, "check", err))
	}

	key, err := KeyParam(c)
	if err != nil {
		return badRequest(c, fmt.Errorf(format, "key", err))
	}

	value := c.FormValue(name) == "on"

	ctx := c.Request().Context()
	if err := fn(ctx, tx, key, value); err != nil {
		return badRequest(c, fmt.Errorf(format, "fn", err))
	}

	return c.String(http.StatusOK, successSpan)
}

func CommitStr(c *echo.Context, tx *sql.Tx, name string,
	fn func(context.Context, boil.ContextExecutor, int64, string) error,
) error {
	const format = "commit str %s: %w"
	if err := nils.Check(c, tx, fn); err != nil {
		return badRequest(c, fmt.Errorf(format, "check", err))
	}

	key, err := Key(c)
	if err != nil {
		return badRequest(c, fmt.Errorf(format, "key", err))
	}

	value := c.FormValue(name)

	ctx := c.Request().Context()
	if err := fn(ctx, tx, key, value); err != nil {
		return badRequest(c, fmt.Errorf(format, "fn", err))
	}

	return StatusOK(c, name)
}

func CommitStrKey(c *echo.Context, tx *sql.Tx, name string,
	fn func(context.Context, boil.ContextExecutor, int64, string) error,
) error {
	const format = "commit str key %s: %w"
	if err := nils.Check(c, tx, fn); err != nil {
		return badRequest(c, fmt.Errorf(format, "check", err))
	}

	key, err := KeyParam(c)
	if err != nil {
		return badRequest(c, fmt.Errorf(format, "key", err))
	}

	value := c.FormValue(name)

	ctx := c.Request().Context()
	if err := fn(ctx, tx, key, value); err != nil {
		return badRequest(c, fmt.Errorf(format, "fn", err))
	}

	return StatusOK(c, name)
}

func StatusOK(c *echo.Context, name string) error {
	const code = http.StatusOK
	switch name {
	case
		"artifact-editor-credits-undo",
		"artifact-editor-comment-undo",
		"artifact-editor-date-undo",
		"artifact-editor-filename-undo",
		"artifact-editor-title-undo":
		return c.String(code, " "+checkMark)
	case
		"artifact-editor-date-update",
		"artifact-editor-releasers":
		return c.String(code, "Save "+checkMark)
	default:
		return c.String(code, successSpan)
	}
}

func Key(c *echo.Context) (int64, error) {
	sid := c.FormValue(EditorKey)
	key, err := strconv.Atoi(sid)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", EditorKey, err)
	}

	return int64(key), nil
}

// KeyParam returns the "id" value from the URL parameters or returns 0 and an error if it is invalid.
func KeyParam(c *echo.Context) (int64, error) {
	const format = "key param %s: %w"
	if err := nils.Check(c); err != nil {
		return 0, fmt.Errorf(format, "check", err)
	}

	id, err := echo.PathParam[int](c, "id")
	if err != nil {
		return 0, fmt.Errorf(format, "path", err)
	}

	return int64(id), nil
}

// badRequest returns an error response with a 400 status code,
// the server cannot or will not process the request due to something that is perceived to be a client error.
func badRequest(c *echo.Context, err error) error {
	const code = http.StatusBadRequest
	if err == nil {
		return c.String(code, "something went wrong")
	}
	return c.String(code, err.Error())
}

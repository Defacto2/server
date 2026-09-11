//nolint:dupl
package app

// Package file error.go contains the error handlers for the application.

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"syscall"

	"github.com/Defacto2/server/handler/releaser"
	"github.com/Defacto2/server/internal/nils"
	"github.com/labstack/echo/v5"
)

const ErrTmpl = "cannot render the html template for this page"

func logErr(sl *slog.Logger, msg, uri string, code int, err error) {
	sl.Error(msg, slog.Int("code", code),
		slog.String("uri", uri), slog.Any("error", err))
}

func logID(sl *slog.Logger, msg, uri string, id int64, err error) {
	sl.Error(msg, slog.Int64("file_id", id),
		slog.String("file_uri", uri), slog.Any("error", err))
}

// ArtifactErr renders the error page for the artifact links.
func ArtifactErr(sl *slog.Logger, c *echo.Context, id string) error {
	const msg = "artifact 404 context"
	if cErr := nils.Check(c, sl); cErr != nil {
		return fmt.Errorf("%s: %w", msg, cErr)
	}

	const code = http.StatusNotFound
	scode := strconv.Itoa(code)
	data := empty(c)
	data["title"] = scode + " error, artifact page not found"
	data["description"] = "HTTP status " + scode + " error"
	data["code"] = code
	data["logo"] = "Artifact not found"
	data["alert"] = "Artifact '" + strings.ToLower(id) + "' cannot be found"
	data["probl"] = "The artifact page does not exist, there is probably a typo with the URL."
	data["uriOkay"] = "f/"
	data["uriErr"] = id

	if rErr := c.Render(code, "status", data); rErr != nil {
		logErr(sl, msg, id, code, rErr)
		return InternalErr(sl, c, status, errorWithID(rErr, id, nil))
	}

	return nil
}

// ArtifactsErr renders the files error page for the Artifacts menu and categories.
// It provides different error messages to the standard error page.
func ArtifactsErr(sl *slog.Logger, c *echo.Context, uri string) error {
	const msg = "artifacts 404 context"
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	const code = http.StatusNotFound
	scode := strconv.Itoa(code)
	errs := fmt.Sprint("artifact page not found for,", uri)
	data := empty(c)
	data["title"] = scode + " error, files page not found"
	data["description"] = "HTTP status " + scode + " error"
	data["code"] = code
	data["logo"] = "Artifacts not found"
	data["alert"] = "Artifacts page cannot be found"
	data["probl"] = "The files category or menu option does not exist, there is probably a typo with the URL."
	data["uriOkay"] = "files/"
	data["uriErr"] = uri

	if rErr := c.Render(code, status, data); rErr != nil {
		logErr(sl, msg, uri, code, rErr)
		return InternalErr(sl, c, errs, rErr)
	}

	return nil
}

// BadRequestErr is the handler for handling Bad Request Errors,
// caused by invalid user input or a malformed client requests.
func BadRequestErr(sl *slog.Logger, c *echo.Context, uri string, err error) error {
	const msg = "bad request handler"
	if cErr := nils.Check(sl, c); cErr != nil {
		return fmt.Errorf("%s %q: %w", msg, uri, cErr)
	}

	const code = http.StatusBadRequest
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	data["title"] = "400 error, there is a complication"
	data["code"] = code
	data["logo"] = "Client error"
	data["alert"] = "Something went wrong, " + err.Error()
	data["probl"] = "It might be a settings or configuration problem or a legacy browser issue."
	data["uriErr"] = uri

	if err != nil {
		logErr(sl, msg, uri, code, err)
	}
	if rErr := c.Render(code, status, data); rErr != nil {
		logErr(sl, msg, uri, code, rErr)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}

	return nil
}

// DatabaseErr is the handler for database connection issues.
// A HTTP 503 Service Unavailable error is returned, to reflect the database
// connection issue but where the server is still running and usable for the client.
func DatabaseErr(sl *slog.Logger, c *echo.Context, uri string, err error) error {
	const msg = "database connection handler"
	if cErr := nils.Check(sl, c); cErr != nil {
		return fmt.Errorf("%s %q: %w", msg, uri, cErr)
	}

	const code = http.StatusServiceUnavailable
	scode := strconv.Itoa(code)
	data := empty(c)
	data["description"] = "HTTP status " + scode + " error"
	data["title"] = scode + " error, there is a complication"
	data["code"] = scode + " service unavailable"
	data["logo"] = "Database error"
	data["alert"] = "Cannot connect to the database!"
	data["uriErr"] = ""
	data["probl"] = "This is not your fault, but the server cannot communicate with the database to display this page."

	if err != nil {
		logErr(sl, msg+" cannot connect to the database", uri, code, err)
	}
	if rErr := c.Render(code, status, data); rErr != nil {
		logErr(sl, msg, uri, code, rErr)
		return echo.NewHTTPError(code, ErrTmpl)
	}

	return nil
}

// DownloadErr is the handler for missing download files and database ID errors.
func DownloadErr(sl *slog.Logger, c *echo.Context, uri string, err error) error {
	const msg = "download not found"
	if cErr := nils.Check(sl, c); cErr != nil {
		return fmt.Errorf("%s %q: %w", msg, uri, cErr)
	}

	const code = http.StatusNotFound
	id := c.Param("id")
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	data["title"] = "404 download error"
	data["code"] = code
	data["logo"] = "Download problem"
	data["alert"] = "Cannot send you this download"
	data["probl"] = "The download you are looking for might have been removed, " +
		"had its filename changed, or is temporarily unavailable. Is the URL correct?"
	data["uriErr"] = strings.Join([]string{uri, id}, "/")

	if err != nil {
		logErr(sl, msg+" for record "+id, uri, code, err)
	}
	if rErr := c.Render(code, status, data); rErr != nil {
		logErr(sl, msg, uri, code, rErr)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}

	return nil
}

// FileMissingErr is the handler for missing download files and database ID errors.
func FileMissingErr(sl *slog.Logger, c *echo.Context, uri string, err error) error {
	const msg = "file missing"
	if cErr := nils.Check(sl, c); cErr != nil {
		return fmt.Errorf("%s %q: %w", msg, uri, cErr)
	}

	const code = http.StatusServiceUnavailable
	id := c.Param("id")
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	data["title"] = "503 download unavailable"
	data["code"] = code
	data["logo"] = "Download unavailable"
	data["alert"] = "Cannot send you this download"
	data["probl"] = "The file download needs to be added to the server; " +
		"otherwise, there may be a problem with the server configuration, or the file may be lost."
	data["uriErr"] = strings.Join([]string{uri, id}, "/")

	if err != nil {
		logErr(sl, msg+" for record "+id, uri, code, err)
	}
	if rErr := c.Render(code, status, data); rErr != nil {
		logErr(sl, msg, uri, code, rErr)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}

	return nil
}

// ForbiddenErr is the handler for handling Forbidden Errors, caused by clients requesting
// pages that they do not have permission to access.
func ForbiddenErr(sl *slog.Logger, c *echo.Context, uri string, err error) error {
	const msg = "forbidden access"
	if cErr := nils.Check(sl, c); cErr != nil {
		return fmt.Errorf("%s %q: %w", msg, uri, cErr)
	}

	const code = http.StatusForbidden
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	data["title"] = "403, forbidden"
	data["code"] = code
	data["logo"] = "Forbidden"
	data["alert"] = "This page is locked"
	data["uriErr"] = uri

	if err != nil {
		data["probl"] = "This page is not intended for the general public, " + err.Error() + "."
		logErr(sl, msg, uri, code, err)
	}
	if rErr := c.Render(code, status, data); rErr != nil {
		logErr(sl, msg, uri, code, rErr)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}

	return nil
}

// InternalErr is the handler for handling Internal Server Errors, caused by programming bugs or crashes.
// The uri string is the part of the URL that caused the error.
// The optional error value is logged using the logger.
// If the echo context is nil then a user hostile, fallback error in raw text is returned.
func InternalErr(sl *slog.Logger, c *echo.Context, uri string, err error) error {
	const msg = "internal server error"
	if cErr := nils.Check(sl, c); cErr != nil {
		return fmt.Errorf("%s %q: %w", msg, uri, cErr)
	}

	const code = http.StatusInternalServerError
	switch {
	case errors.Is(err, syscall.EPIPE):
		// This is a common error when the client disconnects before the response is sent,
		// and commonly happens when using developer hot reloading.
		discard(err)
		return nil
	case errors.Is(err, echo.ErrRendererNotRegistered):
		logErr(sl, msg, uri, code, err)
		message := ""
		if err != nil {
			message = err.Error()
		}
		return echo.NewHTTPError(code, message)
	}

	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	data["title"] = "500 error, there is a complication"
	data["code"] = code
	data["logo"] = "Server error"
	data["alert"] = "Something crashed!"
	data["probl"] = "This is not your fault, " +
		"but the server encountered an internal error or misconfiguration and cannot display this page."
	data["uriErr"] = uri

	if err != nil {
		logErr(sl, msg, uri, code, err)
	}
	if rErr := c.Render(code, status, data); rErr != nil {
		logErr(sl, msg, uri, code, rErr)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}

	return nil
}

// PageErr renders the files page error page for the Artifacts menu and categories.
// It provides different error messages to the standard error page.
func PageErr(sl *slog.Logger, c *echo.Context, uri, page string) error {
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf("page 404 context: %w", err)
	}

	const code = http.StatusNotFound
	scode := strconv.Itoa(code)

	data := empty(c)
	data["title"] = scode + " error, files page not found"
	data["description"] = "HTTP status " + scode + " error"
	data["code"] = code
	data["logo"] = "Page not found"
	data["alert"] = "Artifacts " + uri + " page does not exist"
	data["probl"] = "The files page does not exist, there is probably a typo with the URL."
	data["uriOkay"] = "files/" + uri + "/"
	data["uriErr"] = page

	err := c.Render(code, status, data)
	if err != nil {
		return InternalErr(sl, c, "page not found for '"+page+"' at "+uri, err)
	}

	return nil
}

// ReleaserErr renders the files error page for the Groups menu and invalid releasers.
func ReleaserErr(sl *slog.Logger, c *echo.Context, invalidID string) error {
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf("releaser 404 context: %w", err)
	}

	const code = http.StatusNotFound
	scode := strconv.Itoa(code)

	data := empty(c)
	data["title"] = scode + " error, releaser page not found"
	data["description"] = "HTTP status " + scode + " error"
	data["code"] = code
	data["logo"] = "Releaser not found"
	data["alert"] = "Releaser '" + invalidID + "' cannot be found"
	data["probl"] = "The releaser page does not exist, there is probably a typo with the URL."
	data["uriOkay"] = "g/"
	data["uriErr"] = invalidID

	err := c.Render(code, status, data)
	if err != nil {
		return InternalErr(sl, c, "releaser page not found for, "+invalidID, err)
	}
	return nil
}

// ScenerErr renders the files error page for the People menu and invalid sceners.
func ScenerErr(sl *slog.Logger, c *echo.Context, id string) error {
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf("scene 404 context: %w", err)
	}

	code := http.StatusNotFound
	scode := strconv.Itoa(code)

	data := empty(c)
	data["title"] = scode + " error, scener page not found"
	data["description"] = "HTTP status " + scode + " error"
	data["code"] = code
	data["logo"] = "Scener not found"
	data["alert"] = "Scener '" + releaser.Humanize(id) + "' cannot be found"
	data["probl"] = "The scener page does not exist, there is probably a typo with the URL."
	data["uriOkay"] = "p/"
	data["uriErr"] = id

	err := c.Render(code, status, data)
	if err != nil {
		return InternalErr(sl, c, "scener page not found for, "+id, err)
	}
	return nil
}

// StatusErr is the handler for the HTTP status pages such as the 404 - not found.
// If the logger is nil then the error page is returned but no error is logged.
// If the echo context is nil then a user hostile, fallback error in raw text is returned.
func StatusErr(sl *slog.Logger, c *echo.Context, code int, uri string) error {
	const msg = "http status"
	if cErr := nils.Check(sl, c); cErr != nil {
		return fmt.Errorf("%s %q: %w", msg, uri, cErr)
	}

	scode := strconv.Itoa(code)
	data := empty(c)
	data["description"] = "HTTP status " + scode + " error"

	var title, alert, logo, probl string
	switch code {
	case http.StatusNotFound:
		title = "404 error, page not found"
		logo = "Page not found"
		alert = "The page cannot be found"
		probl = "The page you are looking for might have been removed, " +
			"had its name changed, or is temporarily unavailable."
	case http.StatusForbidden:
		title = "403 error, forbidden"
		logo = "Forbidden"
		alert = "The page is locked"
		probl = "You don't have permission to access this resource."
	case http.StatusInternalServerError:
		return InternalErr(sl, c, uri, nil)
	default:
		s := http.StatusText(code)
		if s == "" {
			logErr(sl, msg, uri, code, ErrStatus)
			return echo.NewHTTPError(http.StatusInternalServerError,
				scode+" status error for the URL, "+uri+": "+ErrStatus.Error())
		}

		const fmtcode = "%d error, %s"
		title = fmt.Sprintf(fmtcode, code, s)
		logo = s
		alert = s
		probl = fmt.Sprintf(fmtcode, code, s)
	}

	data["title"] = title
	data["code"] = code
	data["logo"] = logo
	data["alert"] = alert
	data["probl"] = probl
	data["uriErr"] = uri

	if rErr := c.Render(code, status, data); rErr != nil {
		logErr(sl, msg, uri, code, rErr)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}

	return nil
}

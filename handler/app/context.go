package app

import (
	"context"
	"crypto/sha512"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Defacto2/server/handler/download"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"google.golang.org/api/idtoken"
)

// AboutErr renders the about file error page for the About files links.
func AboutErr(z *zap.SugaredLogger, c echo.Context, id string) error {
	const name = "status"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	if c == nil {
		return InternalErr(z, c, name, ErrCxt)
	}
	data := empty(c)
	data["title"] = fmt.Sprintf("%d error, file about page not found", http.StatusNotFound)
	data["description"] = fmt.Sprintf("HTTP status %d error", http.StatusNotFound)
	data["code"] = http.StatusNotFound
	data["logo"] = "About file not found"
	data["alert"] = fmt.Sprintf("About file %q cannot be found", strings.ToLower(id))
	data["probl"] = "The about file page does not exist, there is probably a typo with the URL."
	data["uriOkay"] = "f/"
	data["uriErr"] = id
	err := c.Render(http.StatusNotFound, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// Artist is the handler for the Artist sceners page.
func Artist(z *zap.SugaredLogger, c echo.Context) error {
	data := empty(c)
	title := "Pixel artists and graphic designers"
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["description"] = demo
	return scener(z, c, postgres.Artist, data)
}

// BBS is the handler for the BBS page ordered by the most files.
func BBS(z *zap.SugaredLogger, c echo.Context) error {
	return bbsHandler(z, c, true)
}

// BBSAZ is the handler for the BBS page ordered alphabetically.
func BBSAZ(z *zap.SugaredLogger, c echo.Context) error {
	return bbsHandler(z, c, false)
}

// Checksum is the handler for the Checksum file record page.
func Checksum(z *zap.SugaredLogger, c echo.Context, id string) error {
	return download.Checksum(z, c, id)
}

// Code is the handler for the Coder sceners page.
func Coder(z *zap.SugaredLogger, c echo.Context) error {
	data := empty(c)
	title := "Coder and programmers"
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["description"] = demo
	return scener(z, c, postgres.Writer, data)
}

// DatabaseErr is the handler for handling database connection errors.
func DatabaseErr(z *zap.SugaredLogger, c echo.Context, uri string, err error) error {
	const code = http.StatusInternalServerError
	if z == nil {
		zapNil(err)
	} else if err != nil {
		z.Errorf("%d error for %q: %s", code, uri, err)
	}
	// render the fallback, text only error page
	if c == nil {
		if z == nil {
			zapNil(fmt.Errorf("%w: databaserr", ErrCxt))
		} else {
			z.Warnf("%s: %s", ErrTmpl, ErrCxt)
		}
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrCxt))
	}
	// render a user friendly error page
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	data["title"] = "500 error, there is a complication"
	data["code"] = code
	data["logo"] = "Database error"
	data["alert"] = "Cannot connect to the database!"
	data["uriErr"] = ""
	data["probl"] = "This is not your fault, but the server cannot communicate with the database to display this page."
	if err := c.Render(code, "status", data); err != nil {
		if z != nil {
			z.Errorf("%s: %s", ErrTmpl, err)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// Download is the handler for the Download file record page.
func Download(z *zap.SugaredLogger, c echo.Context, path string) error {
	d := download.Download{
		Inline: false,
		Path:   path,
	}
	err := d.HTTPSend(z, c)
	if err != nil {
		return DownloadErr(z, c, "d", err)
	}
	return nil
}

// DownloadErr is the handler for missing download files and database ID errors.
func DownloadErr(z *zap.SugaredLogger, c echo.Context, uri string, err error) error {
	const code = http.StatusNotFound
	id := c.Param("id")
	if z == nil {
		zapNil(err)
	} else if err != nil {
		z.Errorf("%d error for %q: %s", code, id, err)
	}
	// render the fallback, text only error page
	if c == nil {
		if z == nil {
			zapNil(fmt.Errorf("%w: downloaderr", ErrCxt))
		} else {
			z.Errorf("%s: %s", ErrTmpl, ErrCxt)
		}
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrCxt))
	}
	// render a user friendly error page
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	data["title"] = "404 download error"
	data["code"] = code
	data["logo"] = "Download problem"
	data["alert"] = "Cannot send you this download"
	data["probl"] = "The download you are looking for might have been removed, " +
		"had its filename changed, or is temporarily unavailable. " +
		"Is the URL correct?"
	data["uriErr"] = strings.Join([]string{uri, id}, "/")
	if err := c.Render(code, "status", data); err != nil {
		if z != nil {
			z.Errorf("%s: %s", ErrTmpl, err)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// FTP is the handler for the FTP page.
func FTP(z *zap.SugaredLogger, c echo.Context) error {
	const title, name = "FTP", "ftp"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := empty(c)
	const lead = "FTP sites are historical, internet-based file servers for uploading and downloading \"elite\" scene releases."
	const key = "releasers"
	data["title"] = title
	data["description"] = lead
	data["logo"] = "FTP sites"
	data["h1"] = title
	data["lead"] = lead
	// releaser.html specific data items
	data["itemName"] = name
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	defer db.Close()
	r := model.Releasers{}
	if err := r.FTP(ctx, db); err != nil {
		return DatabaseErr(z, c, name, err)
	}
	m := model.Summary{}
	if err := m.FTP(ctx, db); err != nil {
		return DatabaseErr(z, c, name, err)
	}
	data[key] = r
	data["stats"] = map[string]string{
		"pubs":   fmt.Sprintf("%d sites", len(r)),
		"issues": string(ByteFileS("file artifact", m.SumCount.Int64, m.SumBytes.Int64)),
		"years":  fmt.Sprintf("%d - %d", m.MinYear.Int16, m.MaxYear.Int16),
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// File is the handler for the file categories page.
func File(z *zap.SugaredLogger, c echo.Context, stats bool) error {
	const title, name = "File categories", "file"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := empty(c)
	data["title"] = title
	data["description"] = "A table of contents for the collection."
	data["logo"] = title
	data["h1"] = title
	data["lead"] = "This page shows the categories and platforms in the collection of file artifacts."
	data["stats"] = stats
	data["counter"] = Stats{}

	data, err := fileWStats(data, stats)
	if err != nil {
		z.Warn(err)
		data["dbError"] = true
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// Files is the handler for the list and preview of the files page.
// The uri is the category or collection of files to display.
// The page is the page number of the results to display.
func Files(z *zap.SugaredLogger, c echo.Context, uri, page string) error {
	if z == nil {
		return InternalErr(z, c, "files", ErrZap)
	}
	// check the uri is valid
	if !Valid(uri) {
		return FilesErr(z, c, uri)
	}
	// check the page is valid
	if page == "" {
		return files(z, c, uri, 1)
	}
	p, err := strconv.Atoi(page)
	if err != nil {
		return PageErr(z, c, uri, page)
	}
	return files(z, c, uri, p)
}

// FilesErr renders the files error page for the Files menu and categories.
// It provides different error messages to the standard error page.
func FilesErr(z *zap.SugaredLogger, c echo.Context, uri string) error {
	const name = "status"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	if c == nil {
		return InternalErr(z, c, name, ErrCxt)
	}
	data := empty(c)
	data["title"] = fmt.Sprintf("%d error, files page not found", http.StatusNotFound)
	data["description"] = fmt.Sprintf("HTTP status %d error", http.StatusNotFound)
	data["code"] = http.StatusNotFound
	data["logo"] = "Files not found"
	data["alert"] = "Files page cannot be found"
	data["probl"] = "The files category or menu option does not exist, there is probably a typo with the URL."
	data["uriOkay"] = "files/"
	data["uriErr"] = uri
	err := c.Render(http.StatusNotFound, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// ForbiddenErr is the handler for handling Forbidden Errors, caused by clients requesting
// pages that they do not have permission to access.
func ForbiddenErr(z *zap.SugaredLogger, c echo.Context, uri string, err error) error {
	const code = http.StatusForbidden
	if z == nil {
		zapNil(err)
	} else if err != nil {
		z.Errorf("%d error for %q: %s", code, uri, err)
	}
	// render the fallback, text only error page
	if c == nil {
		if z == nil {
			zapNil(fmt.Errorf("%w: internalerr", ErrCxt))
		} else {
			z.Errorf("%s: %s", ErrTmpl, ErrCxt)
		}
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrCxt))
	}
	// render a user friendly error page
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	data["title"] = "403, forbidden"
	data["code"] = code
	data["logo"] = "Forbidden"
	data["alert"] = "This page is locked"
	data["probl"] = fmt.Sprintf("This page is not intended for the general public, %s.", err.Error())
	data["uriErr"] = uri
	if err := c.Render(code, "status", data); err != nil {
		if z != nil {
			z.Errorf("%s: %s", ErrTmpl, err)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// GoogleCallback is the handler for the Google OAuth2 callback page to verify
// the [Google ID token].
//
// [Google ID token]: https://developers.google.com/identity/gsi/web/guides/verify-google-id-token
func GoogleCallback(z *zap.SugaredLogger, c echo.Context, clientID string, maxAge int, accounts ...[48]byte) error {
	const name = "google/callback"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}

	// Cross-Site Request Forgery cookie token
	const csrf = "g_csrf_token"
	cookie, err := c.Cookie(csrf)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return c.Redirect(http.StatusForbidden, "/signin")
		}
		return BadRequestErr(z, c, name, err)
	}
	token := cookie.Value

	// Cross-Site Request Forgery post token
	bodyToken := c.FormValue(csrf)
	if token != bodyToken {
		return BadRequestErr(z, c, name, ErrMisMatch)
	}

	// Create a new token verifier.
	// https://pkg.go.dev/google.golang.org/api/idtoken
	ctx := context.Background()
	validator, err := idtoken.NewValidator(ctx)
	if err != nil {
		return BadRequestErr(z, c, name, err)
	}

	// Verify the ID token and using the client ID from the Google API.
	credential := c.FormValue("credential")
	playload, err := validator.Validate(ctx, credential, clientID)
	if err != nil {
		return BadRequestErr(z, c, name, err)
	}

	// Verify the sub value against the list of allowed accounts.
	check := false
	if sub, ok := playload.Claims["sub"]; ok {
		for _, account := range accounts {
			if id, ok := sub.(string); ok && sha512.Sum384([]byte(id)) == account {
				check = true
				break
			}
		}
	}
	if !check {
		fullname := playload.Claims["name"]
		sub := playload.Claims["sub"]
		return ForbiddenErr(z, c, name,
			fmt.Errorf("%w %s. "+
				"If this is a mistake, contact Defacto2 admin and give them this Google account ID: %s",
				ErrUser, fullname, sub))
	}

	if err = sessionHandler(z, c, maxAge, playload.Claims); err != nil {
		return BadRequestErr(z, c, name, err)
	}
	return c.Redirect(http.StatusFound, "/")
}

// History is the handler for the History page.
func History(z *zap.SugaredLogger, c echo.Context) error {
	const name = "history"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	const lead = "In the past, alternative iterations of the name have included" +
		" De Facto, DF, DeFacto, Defacto II, Defacto 2, and the defacto2.com domain."
	const h1 = "The history of the brand"
	data := empty(c)
	data["carousel"] = "#carouselDf2Artpacks"
	data["description"] = lead
	data["logo"] = "The history of Defacto"
	data["h1"] = h1
	data["lead"] = lead
	data["title"] = h1
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// Index is the handler for the Home page.
func Index(z *zap.SugaredLogger, c echo.Context) error {
	const name = "index"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	const lead = "a website maintaining the historic PC cracking and warez scene subcultures." +
		" It covers digital objects including text files, demos, music, art, " +
		"magazines, and other projects."
	const desc = "Defacto2 is " + lead
	data := empty(c)
	data["title"] = "Home"
	data["description"] = desc
	data["h1"] = "Welcome,"
	data["lead"] = "You're at " + lead
	data["milestones"] = Collection()
	{
		// get the signed in given name
		sess, err := session.Get(SessionName, c)
		if err == nil {
			if name, ok := sess.Values["givenName"]; ok {
				if nameStr, ok := name.(string); ok && nameStr != "" {
					data["h1"] = "Welcome, " + nameStr
				}
			}
		}
	}
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// Inline is the handler for the Download file record page.
func Inline(z *zap.SugaredLogger, c echo.Context, path string) error {
	d := download.Download{
		Inline: true,
		Path:   path,
	}
	err := d.HTTPSend(z, c)
	if err != nil {
		return DownloadErr(z, c, "v", err)
	}
	return nil
}

// InternalErr is the handler for handling Internal Server Errors, caused by programming bugs or crashes.
// The uri string is the part of the URL that caused the error.
// The optional error value is logged using the zap sugared logger.
// If the zap logger is nil then the error page is returned but no error is logged.
// If the echo context is nil then a user hostile, fallback error in raw text is returned.
func InternalErr(z *zap.SugaredLogger, c echo.Context, uri string, err error) error {
	const code = http.StatusInternalServerError
	if z == nil {
		zapNil(err)
	} else if err != nil {
		z.Errorf("%d error for %q: %s", code, uri, err)
	}
	// render the fallback, text only error page
	if c == nil {
		if z == nil {
			zapNil(fmt.Errorf("%w: internalerr", ErrCxt))
		} else {
			z.Errorf("%s: %s", ErrTmpl, ErrCxt)
		}
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrCxt))
	}
	// render a user friendly error page
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	data["title"] = "500 error, there is a complication"
	data["code"] = code
	data["logo"] = "Server error"
	data["alert"] = "Something crashed!"
	data["probl"] = "This is not your fault," +
		" but the server encountered an internal error or misconfiguration and cannot display this page."
	data["uriErr"] = uri
	if err := c.Render(code, "status", data); err != nil {
		if z != nil {
			z.Errorf("%s: %s", ErrTmpl, err)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// Interview is the handler for the People Interviews page.
func Interview(z *zap.SugaredLogger, c echo.Context) error {
	const title, name = "Interviews with sceners", "interview"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := empty(c)
	data["title"] = title
	data["description"] = "Discussions with scene members."
	data["logo"] = title
	data["h1"] = title
	data["lead"] = "Here is a centralized page for the site's discussions and unedited" +
		" interviews with sceners, crackers, and demo makers. Currently, incomplete."
	data["interviews"] = Interviewees()
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// Magazine is the handler for the Magazine page.
func Magazine(z *zap.SugaredLogger, c echo.Context) error {
	return mag(z, c, true)
}

// MagazineAZ is the handler for the Magazine page ordered chronologically.
func MagazineAZ(z *zap.SugaredLogger, c echo.Context) error {
	return mag(z, c, false)
}

// mag is the handler for the magazine page.
func mag(z *zap.SugaredLogger, c echo.Context, chronological bool) error {
	const title, name = "Magazine", "magazine"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := empty(c)
	const lead = "The magazines are newsletters, reports, and publications about activities within The Scene subculture."
	const issue = "issue"
	const key = "releasers"
	data["title"] = title
	data["description"] = lead
	data["logo"] = title
	data["h1"] = title
	data["lead"] = lead
	// releaser.html specific data items
	data["itemName"] = issue
	data[key] = model.Releasers{}
	data["stats"] = map[string]string{}

	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	defer db.Close()
	r := model.Releasers{}
	if chronological {
		if err := r.Magazine(ctx, db); err != nil {
			return DatabaseErr(z, c, name, err)
		}
	} else {
		if err := r.MagazineAZ(ctx, db); err != nil {
			return DatabaseErr(z, c, name, err)
		}
	}
	m := model.Summary{}
	if err := m.Magazine(ctx, db); err != nil {
		return DatabaseErr(z, c, name, err)
	}
	data[key] = r
	data["stats"] = map[string]string{
		"pubs":   fmt.Sprintf("%d publications", len(r)),
		"issues": string(ByteFileS(issue, m.SumCount.Int64, m.SumBytes.Int64)),
		"years":  fmt.Sprintf("%d - %d", m.MinYear.Int16, m.MaxYear.Int16),
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// Musician is the handler for the Musiciansceners page.
func Musician(z *zap.SugaredLogger, c echo.Context) error {
	data := empty(c)
	title := "Musicians and composers"
	data["title"] = title
	data["logo"] = title
	data["h1"] = title
	data["description"] = demo
	return scener(z, c, postgres.Musician, data)
}

// Interview is the handler for the People Interviews page.
func Reader(z *zap.SugaredLogger, c echo.Context) error {
	const title, name = "Textfile reader", "reader"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := empty(c)
	data["title"] = title
	data["description"] = "Discussions with scene members."
	data["logo"] = title
	data["h1"] = title
	data["lead"] = "An incomplete list of discussions and unedited interviews with sceners," +
		" crackers and demo makers."
	data["interviews"] = Interviewees()
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// Thanks is the handler for the Thanks page.
func Thanks(z *zap.SugaredLogger, c echo.Context) error {
	const name = "thanks"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := empty(c)
	data["description"] = "Defacto2 thankyous."
	data["h1"] = "Thank you!"
	data["lead"] = "Thanks to the hundreds of people who have contributed to" +
		" Defacto2 over the decades with file submissions, " +
		"hard drive donations, interviews, corrections, artwork, and monetary contributions!"
	data["title"] = "Thanks!"
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// TheScene is the handler for the The Scene page.
func TheScene(z *zap.SugaredLogger, c echo.Context) error {
	const name = "thescene"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	const h1 = "The Scene?"
	const lead = "Collectively referred to as The Scene," +
		" it is a subculture of different computer activities where participants" +
		" actively share ideas and creations."
	data := empty(c)
	data["description"] = fmt.Sprint(h1, " ", lead)
	data["logo"] = "The underground"
	data["h1"] = h1
	data["lead"] = lead
	data["title"] = h1
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// Signin is the handler for the Sign in session page.
func Signin(z *zap.SugaredLogger, c echo.Context, clientID string) error {
	const name = "signin"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := empty(c)
	data["title"] = "Sign in"
	data["description"] = "Sign in to Defacto2."
	data["h1"] = "Sign in"
	data["lead"] = "This sign-in is not open to the general public, and no registration is available."
	data["callback"] = "/google/callback"
	data["clientID"] = clientID
	data["nonce"] = ""
	{ // get any existing session
		sess, err := session.Get(SessionName, c)
		if err != nil {
			return remove(z, c, name, data)
		}
		id, ok := sess.Values["sub"]
		if !ok {
			return remove(z, c, name, data)
		}
		idStr, ok := id.(string)
		if ok && idStr != "" {
			return SignOut(z, c)
		}
	}
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// BadRequestErr is the handler for handling Bad Request Errors, caused by invalid user input
// or a malformed client requests.
func BadRequestErr(z *zap.SugaredLogger, c echo.Context, uri string, err error) error {
	const code = http.StatusBadRequest
	if z == nil {
		zapNil(err)
	} else if err != nil {
		z.Errorf("%d error for %q: %s", code, uri, err)
	}
	// render the fallback, text only error page
	if c == nil {
		if z == nil {
			zapNil(fmt.Errorf("%w: internalerr", ErrCxt))
		} else {
			z.Errorf("%s: %s", ErrTmpl, ErrCxt)
		}
		return echo.NewHTTPError(http.StatusInternalServerError,
			fmt.Errorf("%w: handler app status", ErrCxt))
	}
	// render a user friendly error page
	data := empty(c)
	data["description"] = fmt.Sprintf("HTTP status %d error", code)
	data["title"] = "400 error, there is a complication"
	data["code"] = code
	data["logo"] = "Client error"
	data["alert"] = "Something went wrong, " + err.Error()
	data["probl"] = "It might be a settings or configuration problem or a legacy browser issue."
	data["uriErr"] = uri
	if err := c.Render(code, "status", data); err != nil {
		if z != nil {
			z.Errorf("%s: %s", ErrTmpl, err)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

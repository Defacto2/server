package app

// Package file render.go contains the handler functions for the app pages.
// The BBS, FTP, Magazine and Releaser handlers can be found in render_releaser.go.

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Defacto2/server/internal/cache"
	"github.com/Defacto2/server/internal/pouet"
	"github.com/Defacto2/server/internal/zoo"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const (
	sep  = ";"
	demo = "demo"
)

// empty is a map of default values for the app templates.
func empty(c echo.Context) map[string]interface{} {
	// the keys are listed in order of appearance in the templates.
	// * marked keys are required.
	// ! marked keys are suggested.
	return map[string]interface{}{
		"cacheFiles":  Caching.RecordCount, //   The number of records of files in the database.
		"canonical":   "",                  //   A canonical URL is the URL of the best representative page from a group of duplicate pages.
		"carousel":    "",                  //   The ID of the carousel to display.
		"counter":     Statistics(),        //   Empty database counts for files and categories.
		"dbError":     false,               //   If true, the database is not available.
		"description": "",                  // * A short description of the page that get inserted into the description meta element.
		"editor":      editor(c),           //   If true, the editor mode is enabled.
		"h1":          "",                  // ! The H1 heading of the page.
		"h1Sub":       "",                  //   The H1 sub-heading of the page.
		"jsdos":       false,               //   If true, the large, JS-DOS emulator files will be loaded.
		"lead":        "",                  // ! The enlarged, lead paragraph of the page.
		"logo":        "",                  // ! Text to insert into the monospaced, ASCII art logo.
		"readOnly":    true,                //   If true, the application is in read-only mode.
		"title":       "",                  // * The title of the page that get inserted into the title meta element.
	}
}

// editor returns true if the user is signed in and is an editor.
func editor(c echo.Context) bool {
	sess, err := session.Get(SessionName, c)
	if err != nil {
		return false
	}
	if id, ok := sess.Values["sub"]; ok && id != "" {
		// additional check could be sub against DB
		return true
	}
	return false
}

// emptyFiles is a map of default values specific to the files templates.
func emptyFiles(c echo.Context) map[string]interface{} {
	data := empty(c)
	data["demozoo"] = "0"
	data["sixteen"] = ""
	data["scener"] = ""
	data["website"] = ""
	data["unknownYears"] = true
	return data
}

// ProdPouet is the handler for the Pouet prod JSON page.
func ProdPouet(z *zap.SugaredLogger, c echo.Context, id string) error {
	const name = "pouet"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	p := pouet.Pouet{}
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}
	if err = p.Uploader(i); err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}
	if err = c.JSON(http.StatusOK, p); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return nil
}

// ProdZoo is the handler for the Demozoo production JSON page.
func ProdZoo(z *zap.SugaredLogger, c echo.Context, id string) error {
	const name = "demozoo"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := zoo.Demozoo{}
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}
	if err = data.Get(i); err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}
	if err = c.JSON(http.StatusOK, data); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return nil
}

// VotePouet is the handler for the Pouet production votes JSON page.
func VotePouet(z *zap.SugaredLogger, c echo.Context, id string) error {
	const title, name, sep = "Pouet", "pouet", ";"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	pv := pouet.Votes{}
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	cp := cache.Pouet
	if s, err := cp.Read(id); err == nil {
		if err := PouetCache(c, s); err == nil {
			z.Debugf("cache hit for pouet id %s", id)
			return nil
		}
	}
	z.Debugf("cache miss for pouet id %s", id)

	if err = pv.Votes(i); err != nil {
		return c.String(http.StatusNotFound, err.Error())
	}

	if err = c.JSON(http.StatusOK, pv); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	val := fmt.Sprintf("%.1f%s%d%s%d%s%d",
		pv.Stars, sep, pv.VotesDown, sep, pv.VotesUp, sep, pv.VotesMeh)
	if err := cp.Write(id, val, cache.ExpiredAt); err != nil {
		z.Errorf("failed to write pouet id %s to cache db: %s", id, err)
	}
	return nil
}

// PouetCache parses the cached data for the Pouet production votes.
// If the cache is valid it is returned as JSON response.
// If the cache is invalid or corrupt an error will be returned
// and a API request should be made to Pouet.
func PouetCache(c echo.Context, data string) error {
	if data == "" {
		return nil
	}
	pv := pouet.Votes{}
	x := strings.Split(data, sep)
	const expect = 4
	if l := len(x); l != expect {
		return fmt.Errorf("%w: %d, want %d", ErrData, l, expect)
	}
	stars, err := strconv.ParseFloat(x[0], 64)
	if err != nil {
		return fmt.Errorf("%w: %s", err, x[0])
	}
	vd, err := strconv.Atoi(x[1])
	if err != nil {
		return fmt.Errorf("%w: %s", err, x[1])
	}
	vu, err := strconv.Atoi(x[2])
	if err != nil {
		return fmt.Errorf("%w: %s", err, x[2])
	}
	vm, err := strconv.Atoi(x[3])
	if err != nil {
		return fmt.Errorf("%w: %s", err, x[3])
	}
	pv.Stars = stars
	pv.VotesDown = uint64(vd)
	pv.VotesUp = uint64(vu)
	pv.VotesMeh = uint64(vm)
	if err = c.JSON(http.StatusOK, pv); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return nil
}

// remove is a helper function to remove the session cookie by setting the MaxAge to -1.
func remove(z *zap.SugaredLogger, c echo.Context, name string, data map[string]interface{}) error {
	sess, err := session.Get(SessionName, c)
	if err != nil {
		const remove = -1
		sess.Options.MaxAge = remove
		_ = sess.Save(c.Request(), c.Response())
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// SignOut is the handler for the Sign out of Defacto2 page.
func SignOut(z *zap.SugaredLogger, c echo.Context) error {
	const name = "signout"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := empty(c)
	data["title"] = "Sign out"
	data["description"] = "Sign out of Defacto2."
	data["h1"] = "Sign out"
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// SignedOut is the handler to sign out and remove the current session.
func SignedOut(z *zap.SugaredLogger, c echo.Context) error {
	const name = "signedout"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	{ // get any existing session
		sess, err := session.Get(SessionName, c)
		if err != nil {
			return BadRequestErr(z, c, name, err)
		}
		id, ok := sess.Values["sub"]
		if !ok || id == "" {
			return ForbiddenErr(z, c, name, ErrSession)
		}
		const remove = -1
		sess.Options.MaxAge = remove
		err = sess.Save(c.Request(), c.Response())
		if err != nil {
			return InternalErr(z, c, name, err)
		}
	}
	return c.Redirect(http.StatusFound, "/")
}

// sessionHandler creates a [new session] and populates it with
// the claims data created by the [ID Tokens for Google HTTP APIs].
//
// [new session]: https://pkg.go.dev/github.com/gorilla/sessions
// [ID Tokens for Google HTTP APIs]: https://pkg.go.dev/google.golang.org/api/idtoken
func sessionHandler(
	z *zap.SugaredLogger,
	c echo.Context, maxAge int,
	claims map[string]interface{},
) error {
	const name = "google/callback"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}

	// get always returns a session, even if empty.
	session, err := session.Get(SessionName, c)
	if err != nil {
		return err
	}

	// session Options are cookie options and are all optional
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies
	const hour = 60 * 60
	session.Options = &sessions.Options{
		Path:     "/",                  // path that must exist in the requested URL to send the Cookie header
		Domain:   "",                   // which server can receive a cookie
		MaxAge:   hour * maxAge,        // maximum age for the cookie, in seconds
		Secure:   true,                 // cookie requires HTTPS except for localhost
		HttpOnly: true,                 // stops the cookie being read by JS
		SameSite: http.SameSiteLaxMode, // LaxMode (default) or StrictMode
	}

	// sub is the unique user id from google
	val, ok := claims["sub"]
	if !ok {
		return ErrClaims
	}
	session.Values["sub"] = val

	// optionals
	session.Values["givenName"] = claims["given_name"]
	session.Values["email"] = claims["email"]
	session.Values["emailVerified"] = claims["email_verified"]

	// save the session
	return session.Save(c.Request(), c.Response())
}

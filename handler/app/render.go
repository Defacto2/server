package app

// Package file render.go contains the handler functions for the app pages.
// The BBS, FTP, Magazine and Releaser handlers can be found in render_releaser.go.

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Defacto2/server/handler/download"
	"github.com/Defacto2/server/internal/cache"
	"github.com/Defacto2/server/internal/pouet"
	"github.com/Defacto2/server/internal/zoo"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"google.golang.org/api/idtoken"
)

var ErrData = fmt.Errorf("cache data is invalid or corrupt")

const (
	sep  = ";"
	demo = "demo"
)

// empty is a map of default values for the app templates.
func empty(c echo.Context) map[string]interface{} {
	editor := false
	sess, _ := session.Get("d2_op", c)
	if id, ok := sess.Values["sub"]; ok && id != "" {
		// additional check could be sub against DB
		editor = true
	}
	//fmt.Println("sess", sess, "valeues", sess.Values["sub"])
	// if sess.ID != "" {
	// 	editor = true
	// }

	// the keys are listed in order of appearance in the templates.
	// * marked keys are required.
	// ! marked keys are suggested.
	return map[string]interface{}{
		"title":       "", // * The title of the page that get inserted into the title meta element.
		"canonical":   "", //   A canonical URL is the URL of the best representative page from a group of duplicate pages.
		"description": "", // * A short description of the page that get inserted into the description meta element.

		"logo":     "",    // ! Text to insert into the monospaced, ASCII art logo.
		"h1":       "",    // ! The H1 heading of the page.
		"h1sub":    "",    //   The H1 sub-heading of the page.
		"lead":     "",    // ! The enlarged, lead paragraph of the page.
		"carousel": "",    //   The ID of the carousel to display.
		"jsdos":    false, //  	If true, the large, JS-DOS emulator files will be loaded.

		"counter":      Statistics(),        // Empty database counts for files and categories.
		"df2FileCount": Caching.RecordCount, // The number of records of files in the database.

		"dberror":  false,  // If true, the database is not available.
		"readonly": true,   // If true, the application is in read-only mode.
		"editor":   editor, // If true, the editor mode is enabled.
	}
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

// Checksum is the handler for the Checksum file record page.
func Checksum(z *zap.SugaredLogger, c echo.Context, id string) error {
	return download.Checksum(z, c, id)
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
	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
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

// Interview is the handler for the People Interviews page.
func Reader(z *zap.SugaredLogger, c echo.Context, id string) error {
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

func Signin(z *zap.SugaredLogger, c echo.Context) error {
	const name = "signin"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}

	data := empty(c)
	data["title"] = "Sign in"
	data["description"] = "Sign in to Defacto2."
	data["h1"] = "Sign in"
	data["lead"] = "Sign in to Defacto2."

	data["callback"] = "http://localhost:1323/google/callback"                                    // todo: passthrough data values?
	data["clientID"] = "885513036389-n4uee89egjaph948pbpg7qcesf00gi0g.apps.googleusercontent.com" // todo: pass through and validate data onload?

	err := c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

func Signout(z *zap.SugaredLogger, c echo.Context) error {
	const name = "signin"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	data := empty(c)
	data["title"] = "Sign out"
	data["description"] = "Sign out of Defacto2."
	data["h1"] = "Sign out"
	data["lead"] = "Sign out to Defacto2."

	sess, err := session.Get("d2_op", c)
	if err != nil {
		return echo.ErrNotFound
	}
	id, ok := sess.Values["sub"]
	if !ok || id == "" {
		return echo.ErrForbidden
	}

	const signout = -1
	sess.Options.MaxAge = signout
	err = sess.Save(c.Request(), c.Response())
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	return nil
}

// csrf_token_cookie = self.request.cookies.get('g_csrf_token')
// if not csrf_token_cookie:
//     webapp2.abort(400, 'No CSRF token in Cookie.')
// csrf_token_body = self.request.get('g_csrf_token')
// if not csrf_token_body:
//     webapp2.abort(400, 'No CSRF token in post body.')
// if csrf_token_cookie != csrf_token_body:
//     webapp2.abort(400, 'Failed to verify double submit cookie.')

func GoogleCallback(z *zap.SugaredLogger, c echo.Context, clientID string) error {
	// https://developers.google.com/identity/gsi/web/guides/verify-google-id-token
	// https://pkg.go.dev/google.golang.org/api/idtoken
	const name = "google/callback"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}

	const csrf = "g_csrf_token"
	cookie, err := c.Cookie(csrf)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			fmt.Println("no cookie, redirect to signin")
			return c.Redirect(http.StatusTemporaryRedirect, "/signin")
		}
		return BadRequestErr(z, c, name, err)
	}
	token := cookie.Value

	bodyToken := c.FormValue(csrf)
	if token != bodyToken {
		return BadRequestErr(z, c, name, fmt.Errorf("token mismatch"))
	}

	ctx := context.Background()
	validator, err := idtoken.NewValidator(ctx)
	if err != nil {
		return BadRequestErr(z, c, name, err)
	}

	credential := c.FormValue("credential")
	playload, err := validator.Validate(ctx, credential, clientID)
	if err != nil {
		return BadRequestErr(z, c, name, err)
	}

	if err = SessionHandler(z, c, token, playload.Claims); err != nil {
		return BadRequestErr(z, c, name, err)
	}

	// todo: redirect to the page they were on before signin
	err = c.String(http.StatusOK, "ok, confirm the terminal")
	if err != nil {
		return InternalErr(z, c, name, err)
	}

	return nil
}

func SessionHandler(z *zap.SugaredLogger, c echo.Context, token string, claims map[string]interface{}) error {
	// https://pkg.go.dev/github.com/gorilla/sessions
	// https://pkg.go.dev/google.golang.org/api/idtoken
	//
	// Features, simple API
	// Built-in backends to store sessions in cookies or the filesystem
	// Flash messages, values that last for exactly one request
	// Convenient way to switch session persistency (aka "remember me") and set other attributes
	// Mechanism to rotate authentication and encryption keys

	// get always returns a session, even if empty.
	session, err := session.Get("d2_op", c)
	if err != nil {
		return err
	}

	// session Options are cookie options and are all optional
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies
	const oneHour = 60 * 60
	session.Options = &sessions.Options{
		Path:     "/",                  // path that must exist in the requested URL to send the Cookie header
		Domain:   "",                   // which server can receive a cookie
		MaxAge:   oneHour * 6,          // maximum age for the cookie, in seconds
		Secure:   true,                 // cookie requires HTTPS except for localhost
		HttpOnly: true,                 // stops the cookie being read by JS
		SameSite: http.SameSiteLaxMode, // LaxMode (default) or StrictMode
	}

	// CSRF token from google
	if token == "" {
		return fmt.Errorf("no csrf token in claims")
	}
	session.Values["csrfToken"] = token

	// sub is the unique user id from google
	val, ok := claims["sub"]
	if !ok {
		return fmt.Errorf("no sub id in claims")
	}
	session.Values["sub"] = val

	// not required
	session.Values["givenName"] = claims["given_name"]
	session.Values["email"] = claims["email"]
	session.Values["emailVerified"] = claims["email_verified"]

	err = session.Save(c.Request(), c.Response())
	if err != nil {
		return err
	}

	return nil

	// csrf == cookie values
	// g_csrf_token=5322598e6a01d04   /// Cookie value: 5322598e6a01d04
	// iss https://accounts.google.com
	// sub 116860644014108518010
	// email bengarrett77@gmail.com
	// email_verified true
	// exp 1.705616751e+09
	// azp 885513036389-n4uee89egjaph948pbpg7qcesf00gi0g.apps.googleusercontent.com
	// aud 885513036389-n4uee89egjaph948pbpg7qcesf00gi0g.apps.googleusercontent.com
	// name Ben Garrett
	// family_name Garrett
	// iat 1.705613151e+09
	// nbf 1.705612851e+09
	// locale en-GB
	// given_name Ben
	// jti 6bc5ce2d2821bed7e143d7019a51c3fc958b91fa

	// body dump
	//c.Echo().Use(middleware.BodyDumpWithConfig(middleware.BodyDumpConfig{}))

	// data := empty()
	// data["title"] = "Google callback"
	// data["description"] = "Google callback."
	// data["h1"] = "Google callback"
	// data["lead"] = "Google callback."
	// err := c.Render(http.StatusOK, name, data)
	// if err != nil {
	// 	return InternalErr(z, c, name, err)
	// }
}

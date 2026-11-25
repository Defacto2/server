//nolint:wrapcheck
package app

// Package file context_auth.go contains handlers for operator sign-in, sign-out,
// Google OAuth2 ID-token verification, and cookie sessions.

import (
	"crypto/sha512"
	"crypto/subtle"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Defacto2/server/handler/sess"
	"github.com/Defacto2/server/internal/nils"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/v5/session"
	"github.com/labstack/echo/v5"
	"google.golang.org/api/idtoken"
)

const (
	callback  = "google_callback"
	signedout = "signedout"
	signin    = "signin"
	signout   = "signout"
)

// GoogleCallback is the handler for the Google OAuth2 callback page to verify
// the [Google ID token].
//
// [Google ID token]: https://developers.google.com/identity/gsi/web/guides/verify-google-id-token
func GoogleCallback(sl *slog.Logger, c *echo.Context, clientID string, maxAge int, accounts ...[48]byte,
) error {
	const format = "google callback context: %w"
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf(format, err)
	}

	// Cross-Site Request Forgery cookie token
	const csrf = "g_csrf_token"
	cookie, err := c.Cookie(csrf)
	if err != nil {
		return BadRequestErr(sl, c, callback, err)
	}
	token := cookie.Value

	// Cross-Site Request Forgery post token
	bodyToken := c.FormValue(csrf)
	if token == "" || subtle.ConstantTimeCompare([]byte(token), []byte(bodyToken)) != 1 {
		return BadRequestErr(sl, c, callback, ErrMisMatch)
	}

	// Create a new token verifier.
	// https://pkg.go.dev/google.golang.org/api/idtoken
	ctx := c.Request().Context()
	validator, err := idtoken.NewValidator(ctx)
	if err != nil {
		return BadRequestErr(sl, c, callback, err)
	}

	// Verify the ID token using the client ID from the Google API.
	credential := c.FormValue("credential")
	payload, err := validator.Validate(ctx, credential, clientID)
	if err != nil {
		return BadRequestErr(sl, c, callback, err)
	}

	// Verify the sub value against the list of allowed accounts.
	check := false
	if sub, subExists := payload.Claims["sub"]; subExists {
		for _, account := range accounts {
			if id, subString := sub.(string); subString && sha512.Sum384([]byte(id)) == account {
				check = true
				break
			}
		}
	}
	if !check {
		sub := payload.Claims["sub"]
		const msg = "%w. If this is a mistake, contact Defacto2 admin and give them this Google account ID: %s"
		return ForbiddenErr(sl, c, callback, fmt.Errorf(msg, ErrUser, sub))
	}

	if err = sessionHandler(c, maxAge, payload.Claims); err != nil {
		return BadRequestErr(sl, c, callback, err)
	}
	return c.Redirect(http.StatusFound, "/")
}

// sessionHandler creates a [new session] and populates it with
// the claims data created by the [ID Tokens for Google HTTP APIs].
//
// [new session]: https://pkg.go.dev/github.com/gorilla/sessions
// [ID Tokens for Google HTTP APIs]: https://pkg.go.dev/google.golang.org/api/idtoken
func sessionHandler(c *echo.Context, maxAge int, claims map[string]any,
) error {
	const format = "session handler get: %w"
	if err := nils.Check(c); err != nil {
		return fmt.Errorf(format, err)
	}

	s, err := session.Get(sess.Name, c)
	if err != nil {
		return fmt.Errorf(format, err)
	}

	// session Options are cookie options and are all optional
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies
	const hour = 60 * 60
	s.Options = &sessions.Options{
		Path:        "/",                  // path that must exist in the requested URL to send the Cookie header
		Domain:      "",                   // which server can receive a cookie
		MaxAge:      hour * maxAge,        // maximum age for the cookie, in seconds
		Secure:      true,                 // cookie requires HTTPS except for localhost
		HttpOnly:    true,                 // stops the cookie being read by JS
		SameSite:    http.SameSiteLaxMode, // LaxMode (default) or StrictMode
		Partitioned: false,
	}

	const uniqueGoogleID = "sub"
	val, valExists := claims[uniqueGoogleID]
	if !valExists {
		return ErrClaims
	}
	s.Values[uniqueGoogleID] = val
	s.Values["givenName"] = claims["given_name"]
	s.Values["email"] = claims["email"]
	s.Values["emailVerified"] = claims["email_verified"]

	// save the session
	return s.Save(c.Request(), c.Response())
}

// SignedOut is the handler to sign out and remove the current session.
func SignedOut(sl *slog.Logger, c *echo.Context) error {
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf("signed out context: %w", err)
	}
	{ // get any existing session
		s, err := session.Get(sess.Name, c)
		if err != nil {
			return BadRequestErr(sl, c, signedout, err)
		}

		id, ok := s.Values["sub"]
		if !ok || id == "" {
			return ForbiddenErr(sl, c, signedout, ErrSession)
		}

		const remove = -1
		s.Options.MaxAge = remove
		if err = s.Save(c.Request(), c.Response()); err != nil {
			return InternalErr(sl, c, signedout, err)
		}
	}

	return c.Redirect(http.StatusFound, "/")
}

// SignOut is the handler for the Sign out of Defacto2 page.
func SignOut(sl *slog.Logger, c *echo.Context) error {
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf("sign out context: %w", err)
	}

	const title = "Sign out"
	data := empty(c)
	data.SetNoIndex(true)
	data.SetHeader(title, title, "", "", "Sign out of Defacto2.")

	if err := c.Render(http.StatusOK, signout, data); err != nil {
		return InternalErr(sl, c, signout, err)
	}
	return nil
}

// Signin is the handler for the Sign in session page.
func Signin(sl *slog.Logger, c *echo.Context, clientID string, nonce []byte) error {
	if err := nils.Check(c, sl); err != nil {
		return fmt.Errorf("signin context: %w", err)
	}

	const title = "Sign in"
	data := empty(c)
	data.SetNoIndex(true)
	data.SetHeader(title, title, "This is not open to the general public.", "", "Sign in to Defacto2.")
	data["callback"] = "/google/callback"
	data["clientID"] = clientID
	data["nonce"] = string(nonce)
	{ // get any existing session
		s, err := session.Get(sess.Name, c)
		if err != nil {
			return expireCookie(sl, c, signin, data)
		}

		id, ok := s.Values["sub"]
		if !ok {
			return expireCookie(sl, c, signin, data)
		}
		val, find := id.(string)
		if find && val != "" {
			return SignOut(sl, c)
		}
	}

	if err := c.Render(http.StatusOK, signin, data); err != nil {
		return InternalErr(sl, c, signin, err)
	}
	return nil
}

// expireCookie removes the session cookie by setting MaxAge to -1 and renders the page.
func expireCookie(sl *slog.Logger, c *echo.Context, name string, data map[string]any) error {
	if err := nils.Check(sl, c); err != nil {
		return fmt.Errorf("context remove cookie: %w", err)
	}

	s, err := session.Get(sess.Name, c)
	if err == nil && s != nil {
		const remove = -1
		s.Options.MaxAge = remove
		_ = s.Save(c.Request(), c.Response())
	}

	if err = c.Render(http.StatusOK, name, data); err != nil {
		return InternalErr(sl, c, name, err)
	}
	return nil
}

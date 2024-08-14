// Package sess provides functions for handling session and cookies.
package sess

import (
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

// Name is the name given to the session cookie.
const Name = "d2_op"

// Editor returns true if the user is signed in and is an editor.
func Editor(c echo.Context) bool {
	sess, err := session.Get(Name, c)
	if err != nil {
		return false
	}
	if id, idExists := sess.Values["sub"]; idExists && id != "" {
		// an additional check could be added against a hard coded list of editor IDs.
		return true
	}
	return false
}

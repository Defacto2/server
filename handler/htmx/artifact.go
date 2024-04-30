package htmx

import (
	"net/http"
	"strconv"

	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
)

// RecordToggle handles the post submission for the File artifact is online and public toggle.
func RecordToggle(c echo.Context, state bool) error {
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, err)
	}
	if state {
		if err := model.UpdateOnline(c, int64(id)); err != nil {
			return badRequest(c, err)
		}
		return c.String(http.StatusOK, "online")
	}
	if err := model.UpdateOffline(c, int64(id)); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "offline")
}

// badRequest returns an error response with a 400 status code,
// the server cannot or will not process the request due to something that is perceived to be a client error.
func badRequest(c echo.Context, err error) error {
	return c.String(http.StatusBadRequest, err.Error())
}

package htmx

import (
	"net/http"
	"strconv"

	"github.com/Defacto2/server/internal/form"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// HumanizeAndCount handles the post submission for the File artifact classification,
// such as the platform, operating system, section or category tags.
func HumanizeAndCount(c echo.Context, logger *zap.SugaredLogger, name string) error {
	echo.FormFieldBinder(c) // todo replace with a struct, see: https://echo.labstack.com/docs/binding
	section := c.FormValue(name + "-categories")
	platform := c.FormValue(name + "-operatingsystem")
	s, err := form.HumanizeAndCount(section, platform)
	if err != nil {
		logger.Error(err)
		return badRequest(c, err)
	}
	return c.HTML(http.StatusOK, s)
}

// RecordClassification handles the post submission for the File artifact classification,
// such as the platform, operating system, section or category tags.
func RecordClassification(c echo.Context, logger *zap.SugaredLogger) error {
	section := c.FormValue("artifact-editor-categories")
	platform := c.FormValue("artifact-editor-operatingsystem")
	key := c.FormValue("artifact-editor-key")

	s, err := form.HumanizeAndCount(section, platform)
	if err != nil {
		logger.Error(err)
		return badRequest(c, err)
	}
	doNotUpdate := section == "" || platform == ""
	if doNotUpdate {
		return c.HTML(http.StatusOK, s)
	}

	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, err)
	}

	if err := model.UpdateClassification(int64(id), platform, section); err != nil {
		return badRequest(c, err)
	}

	return c.HTML(http.StatusOK, s)
}

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

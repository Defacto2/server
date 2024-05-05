package htmx

// Package file artifact.go provides functions for handling the HTMX requests for the artifact editor.

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Defacto2/server/internal/form"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// HumanizeAndCount handles the post submission for the File artifact classification,
// such as the platform, operating system, section or category tags.
// The return value is either the humanized and counted classification or an error.
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
// The return value is either the humanized and counted classification or an error.
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

func RecordDateIssued(c echo.Context) error {
	year := c.FormValue("artifact-editor-year")
	month := c.FormValue("artifact-editor-month")
	day := c.FormValue("artifact-editor-day")
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, err)
	}

	y, m, d := form.ValidDate(year, month, day)
	if !y || !m || !d {
		return c.NoContent(http.StatusNoContent)
	}
	if err := model.UpdateDateIssued(int64(id), year, month, day); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Save the date")
}

// RecordReleasers handles the post submission for the File artifact releaser.
// It will only update the releaser1 and the releaser2 values if they have changed.
// The return value is either "Updated" or "Update" depending on if the values have changed.
func RecordReleasers(c echo.Context) error {
	reset1 := c.FormValue("releaser1")
	reset2 := c.FormValue("releaser2")
	rel1 := c.FormValue("artifact-editor-releaser1")
	rel2 := c.FormValue("artifact-editor-releaser2")
	key := c.FormValue("artifact-editor-key")

	notModified := (rel1 == reset1 && rel2 == reset2)
	if notModified {
		return c.String(http.StatusNoContent, "")
	}
	if _, err := recordReleases(rel1, rel2, key); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Save the releasers")
}

// RecordReleasersReset handles the post submission for the File artifact releaser reset.
// It will always reset and save the releaser1 and the releaser2 values.
// The return value is always "Resetted" unless an error occurs.
func RecordReleasersReset(c echo.Context) error {
	reset1 := c.FormValue("releaser1")
	reset2 := c.FormValue("releaser2")
	rel1 := c.FormValue("artifact-editor-releaser1")
	rel2 := c.FormValue("artifact-editor-releaser2")
	key := c.FormValue("artifact-editor-key")

	notModified := (rel1 == reset1 && rel2 == reset2)
	if notModified {
		return c.String(http.StatusNoContent, "")
	}
	val, err := recordReleases(reset1, reset2, key)
	if err != nil {
		return badRequest(c, err)
	}
	html := ""
	s := strings.Split(val, "+")
	for i, x := range s {
		s[i] = "<q>" + x + "</q>"
	}
	html = strings.Join(s, " + ")
	return c.HTML(http.StatusOK, html)
}

func recordReleases(rel1, rel2, key string) (string, error) {
	id, err := strconv.Atoi(key)
	if err != nil {
		return "", fmt.Errorf("strconv.Atoi: %w", err)
	}
	val := rel1
	if rel2 != "" {
		val = rel1 + "+" + rel2
	}
	if err := model.UpdateReleasers(int64(id), val); err != nil {
		return "", fmt.Errorf("model.UpdateReleasers: %w", err)
	}
	return val, nil
}

func RecordFilename(c echo.Context) error {
	name := c.FormValue("artifact-editor-filename")
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, err)
	}
	name = form.SanitizeFilename(name)
	if err := model.UpdateFilename(int64(id), name); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

func RecordFilenameReset(c echo.Context) error {
	reset := c.FormValue("artifact-editor-filename-resetter")
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.UpdateFilename(int64(id), reset); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, reset)
}

func RecordTitle(c echo.Context) error {
	title := c.FormValue("artifact-editor-title")
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.UpdateTitle(int64(id), title); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

func RecordTitleReset(c echo.Context) error {
	reset := c.FormValue("artifact-editor-title-resetter")
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.UpdateTitle(int64(id), reset); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, reset)
}

// RecordToggle handles the post submission for the File artifact is online and public toggle.
// The return value is either "online" or "offline" depending on the state.
func RecordToggle(c echo.Context, state bool) error {
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, err)
	}
	if state {
		if err := model.UpdateOnline(int64(id)); err != nil {
			return badRequest(c, err)
		}
		return c.String(http.StatusOK, "online")
	}
	if err := model.UpdateOffline(int64(id)); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "offline")
}

func RecordVirusTotal(c echo.Context) error {
	link := c.FormValue("artifact-editor-virustotal")
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.UpdateVirusTotal(int64(id), link); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

// badRequest returns an error response with a 400 status code,
// the server cannot or will not process the request due to something that is perceived to be a client error.
func badRequest(c echo.Context, err error) error {
	return c.String(http.StatusBadRequest, err.Error())
}

package app

import (
	"net/http"

	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// PostIntro handles the POST request for the intro upload form.
func PostIntro(z *zap.SugaredLogger, c echo.Context) error {
	const name = "post intro"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	x, err := c.FormParams()
	if err != nil {
		return InternalErr(z, c, name, err)
	}
	c.JSONPretty(http.StatusOK, x, "  ")
	return nil
}

// EditorReadme handles the POST request for the editor readme forms.
func EditorReadme(z *zap.SugaredLogger, c echo.Context) error {
	const name = "editor readme"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}

	type Record struct {
		ID     int  `query:"id"`
		Readme bool `query:"readme"`
	}
	// in the handler for /users?id=<userID>
	var record Record
	err := c.Bind(&record)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request " + err.Error()})
	}
	err = model.UpdateNoReadme(z, c, int64(record.ID), record.Readme)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request " + err.Error()})
	}
	return c.JSON(http.StatusOK, record)
}

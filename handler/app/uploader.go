package app

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Defacto2/server/internal/command"
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

func (a AboutConf) EditorMeCP(z *zap.SugaredLogger, c echo.Context) error {
	const name = "editor copy readme"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}

	type Record struct {
		ID     int    `query:"id"`
		Target string `query:"readme"`
	}
	// in the handler for /users?id=<userID>
	var record Record
	err := c.Bind(&record)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request " + err.Error()})
	}

	f, err := model.Record(z, c, record.ID)
	if err != nil {
		return err
	}

	list := strings.Split(f.FileZipContent.String, "\n")
	target := ""
	for _, x := range list {
		s := strings.TrimSpace(x)
		if s == "" {
			continue
		}
		if strings.EqualFold(s, record.Target) {
			target = s
		}
	}
	if target == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request, target not found"})
	}

	fp := filepath.Join(a.DownloadDir, f.UUID.String)

	err = command.UnzipOne(fp, ".txt", target)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request, " + err.Error()})
	}
	return c.JSON(http.StatusOK, record)
}

func (a AboutConf) EdMeRM(z *zap.SugaredLogger, c echo.Context) error {
	const name = "editor remove readme"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}

	type Record struct {
		ID int `query:"id"`
	}
	// in the handler for /users?id=<userID>
	var record Record
	err := c.Bind(&record)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request " + err.Error()})
	}

	f, err := model.Record(z, c, record.ID)
	if err != nil {
		return err
	}

	err = command.RemoveMe(a.DownloadDir, f.UUID.String)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request, " + err.Error()})
	}
	return c.JSON(http.StatusOK, record)
}

func (a AboutConf) EdImgRM(z *zap.SugaredLogger, c echo.Context) error {
	const name = "editor remove images"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}

	type Record struct {
		ID int `query:"id"`
	}
	// in the handler for /users?id=<userID>
	var record Record
	err := c.Bind(&record)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request " + err.Error()})
	}
	fmt.Println(record)

	f, err := model.Record(z, c, record.ID)
	if err != nil {
		return err
	}

	err = command.RemoveImgs(a.ScreenshotDir, a.ThumbnailDir, f.UUID.String)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request, " + err.Error()})
	}
	return c.JSON(http.StatusOK, record)
}

// EditorMe handles the POST request for the editor readme forms.
func (a AboutConf) EditorMe(z *zap.SugaredLogger, c echo.Context) error {
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

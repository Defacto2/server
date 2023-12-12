package app

import (
	"errors"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var (
	ErrTarget = errors.New("target not found")
)

// badRequest returns a JSON response with a 400 status code,
// the server cannot or will not process the request due to something that is perceived to be a client error.
func badRequest(c echo.Context, err error) error {
	return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request " + err.Error()})
}

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
	c.JSON(http.StatusOK, x)
	return nil
}

// PostMeCP handles the POST request for the editor readme copy, input value.
func PostMeCP(z *zap.SugaredLogger, c echo.Context, downloadDir string) error {
	const name = "editor readme cp"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}

	type Form struct {
		ID     int    `query:"id"`
		Target string `query:"readme"`
	}
	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}

	r, err := model.Record(z, c, f.ID)
	if err != nil {
		return badRequest(c, err)
	}

	list := strings.Split(r.FileZipContent.String, "\n")
	target := ""
	for _, x := range list {
		s := strings.TrimSpace(x)
		if s == "" {
			continue
		}
		if strings.EqualFold(s, f.Target) {
			target = s
		}
	}
	if target == "" {
		return badRequest(c, ErrTarget)
	}

	src := filepath.Join(downloadDir, r.UUID.String)
	dst := filepath.Join(downloadDir, r.UUID.String+".txt")
	err = command.UnZipOne(z, src, dst, target)
	if err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

// PostMeHide handles the POST request for the editor readme, hide toggle.
func PostMeHide(z *zap.SugaredLogger, c echo.Context) error {
	const name = "editor readme hide"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}

	type Form struct {
		ID     int  `query:"id"`
		Readme bool `query:"readme"`
	}
	// in the handler for /users?id=<userID>
	var f Form
	err := c.Bind(&f)
	if err != nil {
		return badRequest(c, err)
	}

	if err = model.UpdateNoReadme(c, int64(f.ID), f.Readme); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, f)
}

// PostMeRm handles the POST request for the editor readme, remove button click.
func PostMeRm(z *zap.SugaredLogger, c echo.Context, downloadDir string) error {
	const name = "editor readme remove"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}

	type Form struct {
		ID int `query:"id"`
	}
	var f Form
	err := c.Bind(&f)
	if err != nil {
		return badRequest(c, err)
	}

	r, err := model.Record(z, c, f.ID)
	if err != nil {
		return err
	}

	if err = command.RemoveMe(downloadDir, r.UUID.String); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

func (dir Dirs) PostImgsCP(z *zap.SugaredLogger, c echo.Context) error {
	const name = "editor images copy"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}

	type Form struct {
		ID     int    `query:"id"`
		Target string `query:"readme"`
	}
	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}

	r, err := model.Record(z, c, f.ID)
	if err != nil {
		return badRequest(c, err)
	}

	list := strings.Split(r.FileZipContent.String, "\n")
	target := ""
	for _, x := range list {
		s := strings.TrimSpace(x)
		if s == "" {
			continue
		}
		if strings.EqualFold(s, f.Target) {
			target = s
		}
	}
	if target == "" {
		return badRequest(c, ErrTarget)
	}

	src := filepath.Join(dir.Download, r.UUID.String)
	cmd := command.Dirs{Download: dir.Download, Preview: dir.Preview, Thumbnail: dir.Thumbnail}
	err = cmd.UnZipImage(z, src, r.UUID.String, target)
	if err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

// PostMeRm handles the POST request for the editor complementary images, remove button click.
func (dir Dirs) PostImgsRm(z *zap.SugaredLogger, c echo.Context) error {
	const name = "editor images remove"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}

	type Form struct {
		ID int `query:"id"`
	}
	var f Form
	err := c.Bind(&f)
	if err != nil {
		return badRequest(c, err)
	}

	r, err := model.Record(z, c, f.ID)
	if err != nil {
		return err
	}

	if err = command.RemoveImgs(dir.Preview, dir.Thumbnail, r.UUID.String); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

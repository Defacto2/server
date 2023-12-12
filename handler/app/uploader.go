package app

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var (
	ErrExtract = errors.New("unknown extractor value")
	ErrTarget  = errors.New("target not found")
)

const (
	txt = ".txt" // txt file extension
)

// badRequest returns a JSON response with a 400 status code,
// the server cannot or will not process the request due to something that is perceived to be a client error.
func badRequest(c echo.Context, err error) error {
	return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad request " + err.Error()})
}

// Form is the form data for the editor.
type Form struct {
	ID     int    `query:"id"`     // ID is the auto incrementing database id of the record.
	Readme bool   `query:"readme"` // Readme hides the readme textfile from the about page.
	Target string `query:"target"` // Target is the name of the file to extract from the zip archive.
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

// ReadmeDel handles the post submission for the Delete readme asset button.
func ReadmeDel(z *zap.SugaredLogger, c echo.Context, downloadDir string) error {
	const name = "editor readme delete"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}

	var f Form
	if err := c.Bind(&f); err != nil {
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

// ReadmePost handles the post submission for the Readme in archive.
func ReadmePost(z *zap.SugaredLogger, c echo.Context, downloadDir string) error {
	const name = "editor readme"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
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
	dst := filepath.Join(downloadDir, r.UUID.String+txt)
	err = command.UnZipOne(z, src, dst, target)
	if err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

// ReadmeToggle handles the post submission for the Hide readme from view toggle.
func ReadmeToggle(z *zap.SugaredLogger, c echo.Context) error {
	const name = "editor readme toggle"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}

	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	if err := model.UpdateNoReadme(c, int64(f.ID), f.Readme); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, f)
}

// PreviewPost handles the post submission for the Preview from image in archive.
func (dir Dirs) PreviewPost(z *zap.SugaredLogger, c echo.Context) error {
	const name = "editor preview"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	return dir.extractor(z, c, imgs)
}

// PreviewDel handles the post submission for the Delete complementary images button.
func (dir Dirs) PreviewDel(z *zap.SugaredLogger, c echo.Context) error {
	const name = "editor preview remove"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}

	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	r, err := model.Record(z, c, f.ID)
	if err != nil {
		return badRequest(c, err)
	}
	if err = command.RemoveImgs(dir.Preview, dir.Thumbnail, r.UUID.String); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

// AnsiLovePost handles the post submission for the Preview from text in archive.
func (dir Dirs) AnsiLovePost(z *zap.SugaredLogger, c echo.Context) error {
	const name = "editor ansilove"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}
	return dir.extractor(z, c, ansis)
}

type extract int

const (
	imgs  extract = iota // extract image
	ansis                // extract ansilove compatible text
)

func (dir Dirs) extractor(z *zap.SugaredLogger, c echo.Context, p extract) error {
	if z == nil {
		return InternalErr(z, c, "extractor", ErrZap)
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
	switch p {
	case imgs:
		err = cmd.ExtractImage(z, src, r.UUID.String, target)
	case ansis:
		err = cmd.ExtractAnsiLove(z, src, r.UUID.String, target)
	default:
		return InternalErr(z, c, "extractor", fmt.Errorf("%w: %d", ErrExtract, p))
	}
	if err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

package app

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"
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
	Online bool   `query:"online"` // Online is the record online and public toggle.
	Readme bool   `query:"readme"` // Readme hides the readme textfile from the about page.
	Target string `query:"target"` // Target is the name of the file to extract from the zip archive.
	Value  string `query:"value"`  // Value is the value of the form input field to change.
	Year   int16  `query:"year"`   // Year is the year of the release.
	Month  int16  `query:"month"`  // Month is the month of the release.
	Day    int16  `query:"day"`    // Day is the day of the release.
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

// PlatformEdit handles the post submission for the Platform selection field.
func PlatformEdit(z *zap.SugaredLogger, c echo.Context) error {
	const name = "editor platform"
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
	if err = model.UpdatePlatform(c, int64(f.ID), f.Value); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

// TitleEdit handles the post submission for the Delete readme asset button.
func TitleEdit(z *zap.SugaredLogger, c echo.Context) error {
	const name = "editor title"
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
	if err = model.UpdateTitle(c, int64(f.ID), f.Value); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

func YMDEdit(z *zap.SugaredLogger, c echo.Context) error {
	const name = "editor ymd"
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
	y := ValidY(f.Year)
	m := ValidM(f.Month)
	d := ValidD(f.Day)
	if err = model.UpdateYMD(c, int64(f.ID), y, m, d); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

func ValidY(y int16) null.Int16 {
	if y < 1980 || y > int16(time.Now().Year()) {
		return null.Int16{Int16: 0, Valid: false}
	}
	return null.Int16{Int16: y, Valid: true}
}

func ValidM(m int16) null.Int16 {
	if m < 1 || m > 12 {
		return null.Int16{Int16: 0, Valid: false}
	}
	return null.Int16{Int16: m, Valid: true}
}

func ValidD(d int16) null.Int16 {
	if d < 1 || d > 31 {
		return null.Int16{Int16: 0, Valid: false}
	}
	return null.Int16{Int16: d, Valid: true}
}

// RecordToggle handles the post submission for the File artifact is online and public toggle.
func RecordToggle(z *zap.SugaredLogger, c echo.Context, state bool) error {
	const name = "editor record toggle"
	if z == nil {
		return InternalErr(z, c, name, ErrZap)
	}

	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	if state {
		if err := model.UpdateOnline(c, int64(f.ID)); err != nil {
			return badRequest(c, err)
		}
		return c.JSON(http.StatusOK, f)
	}
	if err := model.UpdateOffline(c, int64(f.ID)); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, f)
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
	ext := filepath.Ext(strings.ToLower(r.Filename.String))
	err = command.ExtractOne(z, src, dst, ext, target)
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

type extract int // extract target format for the file archive extractor

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
	ext := filepath.Ext(strings.ToLower(r.Filename.String))
	switch p {
	case imgs:
		err = cmd.ExtractImage(z, src, r.UUID.String, ext, target)
	case ansis:
		err = cmd.ExtractAnsiLove(z, src, r.UUID.String, ext, target)
	default:
		return InternalErr(z, c, "extractor", fmt.Errorf("%w: %d", ErrExtract, p))
	}
	if err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

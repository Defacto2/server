package app

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/postgres/models"
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

// TODO: move to render_about and make private

func readmeSuggest(r *models.File) string {
	filename := r.Filename.String
	group := r.GroupBrandFor.String
	if group == "" {
		group = r.GroupBrandBy.String
	}
	if x := strings.Split(group, " "); len(x) > 1 {
		group = x[0]
	}
	cont := strings.ReplaceAll(r.FileZipContent.String, "\r\n", "\n")
	content := strings.Split(cont, "\n")
	return ReadmeSug(filename, group, content...)
}

// FileZipContent, Filename, GroupName
func ReadmeSug(filename, group string, content ...string) string {
	// this is a port of variables.findTextfile in File.cfc
	finds := []string{}
	skip := []string{"scene.org", "scene.org.txt"}
	priority := []string{".nfo", ".txt", ".unp", ".doc"}
	candidate := []string{".diz", ".asc", ".1st", ".dox", ".me", ".cap", ".ans", ".pcb"}

	//content := strings.ReplaceAll(res.FileZipContent.String, "\r\n", "\n")
	//strings.Split(content, "\n")

	for _, name := range content {
		if name == "" {
			continue
		}
		s := strings.ToLower(name)
		if slices.Contains(skip, s) {
			continue
		}
		ext := filepath.Ext(s)
		if slices.Contains(priority, ext) {
			finds = append(finds, name)
			continue
		}
		if slices.Contains(candidate, ext) {
			finds = append(finds, name)
		}
	}
	if len(finds) == 1 {
		return finds[0]
	}

	finds = sortContent(finds)

	// match either the filename or the group name with a priority extension
	// e.g. .nfo, .txt, .unp, .doc
	base := filepath.Base(filename)
	for _, ext := range priority {
		for _, name := range finds {
			// match the filename + extension
			if strings.EqualFold(base+ext, name) {
				return name
			}
			// match the group name + extension
			if strings.EqualFold(group+ext, name) {
				return name
			}
		}
	}
	// match file_id.diz
	for _, name := range finds {
		if strings.EqualFold("file_id.diz", name) {
			return name
		}
	}
	// match either the filename or the group name with a candidate extension
	for _, ext := range candidate {
		for _, name := range finds {
			// match the filename + extension
			if strings.EqualFold(base+ext, name) {
				return name
			}
			// match the group name + extension
			if strings.EqualFold(group+ext, name) {
				return name
			}
		}
	}
	// match any finds that use a priority extension
	for _, name := range finds {
		s := strings.ToLower(name)
		ext := filepath.Ext(s)
		if slices.Contains(priority, ext) {
			return name
		}
	}
	// match the first file in the list
	for _, name := range finds {
		return name
	}
	return ""
}

func sortContent(content []string) []string {
	sort.Slice(content, func(i, j int) bool {
		// Fix any Windows path separators
		content[i] = strings.ReplaceAll(content[i], "\\", "/")
		content[j] = strings.ReplaceAll(content[j], "\\", "/")
		// Count the number of slashes in each string
		iCount := strings.Count(content[i], "/")
		jCount := strings.Count(content[j], "/")

		// Prioritize strings with fewer slashes (i.e., closer to the root)
		if iCount != jCount {
			return iCount < jCount
		}

		// If the number of slashes is the same, sort alphabetically
		return content[i] < content[j]
	})

	return content
}

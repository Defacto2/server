package htmx

// Package file artifact.go provides functions for handling the HTMX requests for the artifact editor.

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/internal/form"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// RecordToggle handles the post submission for the file artifact record toggle.
// The return value is either "online" or "offline" depending on the state.
func RecordToggle(c echo.Context, state bool) error {
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	if state {
		if err := model.UpdateOnline(int64(id)); err != nil {
			return badRequest(c, fmt.Errorf("model.UpdateOnline: %w", err))
		}
		return c.String(http.StatusOK, "online")
	}
	if err := model.UpdateOffline(int64(id)); err != nil {
		return badRequest(c, fmt.Errorf("model.UpdateOffline: %w", err))
	}
	return c.String(http.StatusOK, "offline")
}

// RecordClassification handles the post submission for the file artifact classifications,
// such as the platform, operating system, section or category tags.
// The return value is either the humanized and counted classification or an error.
func RecordClassification(c echo.Context, logger *zap.SugaredLogger) error {
	section := c.FormValue("artifact-editor-categories")
	platform := c.FormValue("artifact-editor-operatingsystem")
	key := c.FormValue("artifact-editor-key")
	html, err := form.HumanizeAndCount(section, platform)
	if err != nil {
		logger.Error(err)
		return badRequest(c, err)
	}
	invalid := section == "" || platform == ""
	if invalid {
		return c.HTML(http.StatusOK, string(html))
	}
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	if err := model.UpdateClassification(int64(id), platform, section); err != nil {
		return badRequest(c, err)
	}
	return c.HTML(http.StatusOK, string(html))
}

// RecordFilename handles the post submission for the file artifact filename.
func RecordFilename(c echo.Context) error {
	name := c.FormValue("artifact-editor-filename")
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	name = form.SanitizeFilename(name)
	if err := model.UpdateFilename(int64(id), name); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

// RecordFilenameReset handles the post submission for the file artifact filename reset.
func RecordFilenameReset(c echo.Context) error {
	val := c.FormValue("artifact-editor-filename-resetter")
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	if err := model.UpdateFilename(int64(id), val); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, val)
}

// RecordVirusTotal handles the post submission for the file artifact VirusTotal report link.
func RecordVirusTotal(c echo.Context) error {
	link := c.FormValue("artifact-editor-virustotal")
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	if !form.ValidVT(link) {
		return c.NoContent(http.StatusNoContent)
	}
	if err := model.UpdateVirusTotal(int64(id), link); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

// RecordTitle handles the post submission for the file artifact title.
func RecordTitle(c echo.Context) error {
	title := c.FormValue("artifact-editor-title")
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	if err := model.UpdateTitle(int64(id), title); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

// RecordTitleReset handles the post submission for the file artifact title reset.
func RecordTitleReset(c echo.Context) error {
	val := c.FormValue("artifact-editor-title-resetter")
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	if err := model.UpdateTitle(int64(id), val); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, val)
}

// RecordComment handles the post submission for the file artifact comment.
func RecordComment(c echo.Context) error {
	comment := c.FormValue("artifact-editor-comment")
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	if err := model.UpdateComment(int64(id), comment); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

// RecordCommentReset handles the post submission for the file artifact comment reset.
func RecordCommentReset(c echo.Context) error {
	val := c.FormValue("artifact-editor-comment-resetter")
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	if err := model.UpdateComment(int64(id), val); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Undo comment")
}

// RecordReleasers handles the post submission for the file artifact releasers.
// It will only update the releaser1 and the releaser2 values if they have changed.
// The return value is either "Updated" or "Update" depending on if the values have changed.
func RecordReleasers(c echo.Context) error {
	val1 := c.FormValue("releaser1")
	val2 := c.FormValue("releaser2")
	rel1 := c.FormValue("artifact-editor-releaser1")
	rel2 := c.FormValue("artifact-editor-releaser2")
	key := c.FormValue("artifact-editor-key")
	unchanged := (rel1 == val1 && rel2 == val2)
	if unchanged {
		return c.NoContent(http.StatusNoContent)
	}
	if _, err := recordReleases(rel1, rel2, key); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Save")
}

// RecordReleasersReset handles the post submission for the file artifact releasers reset.
// It will always reset and save the releaser1 and the releaser2 values.
// The return value is always "Resetted" unless an error occurs.
func RecordReleasersReset(c echo.Context) error {
	val1 := c.FormValue("releaser1")
	val2 := c.FormValue("releaser2")
	rel1 := c.FormValue("artifact-editor-releaser1")
	rel2 := c.FormValue("artifact-editor-releaser2")
	key := c.FormValue("artifact-editor-key")
	unchanged := (rel1 == val1 && rel2 == val2)
	if unchanged {
		return c.String(http.StatusNoContent, "")
	}
	val, err := recordReleases(val1, val2, key)
	if err != nil {
		return badRequest(c, err)
	}
	s := strings.Split(val, "+")
	for i, x := range s {
		s[i] = "<q>" + x + "</q>"
	}
	html := strings.Join(s, " + ")
	return c.HTML(http.StatusOK, html)
}

func recordReleases(rel1, rel2, key string) (string, error) {
	id, err := strconv.Atoi(key)
	if err != nil {
		return "", fmt.Errorf("%w: %w: %q", ErrKey, err, key)
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

// RecordDateIssued handles the post submission for the file artifact date of release.
func RecordDateIssued(c echo.Context) error {
	year := c.FormValue("artifact-editor-year")
	month := c.FormValue("artifact-editor-month")
	day := c.FormValue("artifact-editor-day")
	key := c.FormValue("artifact-editor-key")
	yearval := c.FormValue("artifact-editor-yearval")
	monthval := c.FormValue("artifact-editor-monthval")
	dayval := c.FormValue("artifact-editor-dayval")
	if year == yearval && month == monthval && day == dayval {
		return c.NoContent(http.StatusNoContent)
	}
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	y, m, d := form.ValidDate(year, month, day)
	if !y || !m || !d {
		return c.NoContent(http.StatusNoContent)
	}
	if err := model.UpdateDateIssued(int64(id), year, month, day); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Save")
}

// RecordDateIssuedReset handles the post submission for the file artifact date of release reset.
func RecordDateIssuedReset(c echo.Context, elmID string) error {
	reset := c.FormValue(elmID)
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	vals := strings.Split(reset, "-")
	const expected = 3
	if len(vals) != expected {
		return badRequest(c, fmt.Errorf("%w, requires YYYY-MM-DD", ErrDate))
	}
	year, month, day := vals[0], vals[1], vals[2]
	y, m, d := form.ValidDate(year, month, day)
	if !y || !m || !d {
		return badRequest(c, fmt.Errorf("%w, requires YYYY-MM-DD", ErrDate))
	}
	if err := model.UpdateDateIssued(int64(id), year, month, day); err != nil {
		return badRequest(c, err)
	}
	s := year
	if month != "0" {
		s += "-" + month
	}
	if day != "0" {
		s += "-" + day
	}
	return c.String(http.StatusOK, s)
}

// RecordCreatorText handles the post submission for the file artifact creator text.
func RecordCreatorText(c echo.Context) error {
	creator := c.FormValue("artifact-editor-credittext")
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	val := creatorFix(creator)
	if err := model.UpdateCreatorText(int64(id), val); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

// RecordCreatorIll handles the post submission for the file artifact creator illustrator.
func RecordCreatorIll(c echo.Context) error {
	creator := c.FormValue("artifact-editor-creditill")
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	val := creatorFix(creator)
	if err := model.UpdateCreatorIll(int64(id), val); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

// RecordCreatorProg handles the post submission for the file artifact creator programmer.
func RecordCreatorProg(c echo.Context) error {
	creator := c.FormValue("artifact-editor-creditprog")
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	val := creatorFix(creator)
	if err := model.UpdateCreatorProg(int64(id), val); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

// RecordCreatorAudio handles the post submission for the file artifact creator musician.
func RecordCreatorAudio(c echo.Context) error {
	creator := c.FormValue("artifact-editor-creditaudio")
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	val := creatorFix(creator)
	if err := model.UpdateCreatorAudio(int64(id), val); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

func creatorFix(s string) string {
	creators := strings.Split(s, ",")
	for i, c := range creators {
		creators[i] = releaser.Clean(c)
	}
	return strings.Join(creators, ",")
}

// RecordCreatorReset handles the post submission for the file artifact creators reset.
func RecordCreatorReset(c echo.Context) error {
	reset := c.FormValue("artifact-editor-credit-resetter")
	textval := c.FormValue("artifact-editor-credittext")
	illval := c.FormValue("artifact-editor-creditill")
	progval := c.FormValue("artifact-editor-creditprog")
	audioval := c.FormValue("artifact-editor-creditaudio")
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	vals := strings.Split(reset, ";")
	const expected = 4
	if len(vals) != expected {
		return badRequest(c, fmt.Errorf("%w, requires string;string;string;string", ErrCreators))
	}
	text := vals[0]
	ill := vals[1]
	prog := vals[2]
	audio := vals[3]
	if textval == text && illval == ill && progval == prog && audioval == audio {
		return c.NoContent(http.StatusNoContent)
	}
	if err := model.UpdateCreators(int64(id), text, ill, prog, audio); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Undo creators")
}

// RecordYouTube handles the post submission for the file artifact YouTube watch video link.
func RecordYouTube(c echo.Context) error {
	key := c.FormValue("artifact-editor-key")
	watch := c.FormValue("artifact-editor-youtube")
	val := c.FormValue("artifact-editor-youtubeval")
	watch = strings.TrimSpace(watch)
	val = strings.TrimSpace(val)
	if watch == val {
		return c.NoContent(http.StatusNoContent)
	}
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	const requirement = 11
	if len(watch) > 0 && len(watch) < requirement {
		return c.NoContent(http.StatusNoContent)
	}
	if err := model.UpdateYouTube(int64(id), watch); err != nil {
		return badRequest(c, err)
	}
	return RecordLinks(c)
}

// RecordDemozoo handles the post submission for the file artifact Demozoo production link.
func RecordDemozoo(c echo.Context) error {
	key := c.FormValue("artifact-editor-key")
	prod := c.FormValue("artifact-editor-demozoo")
	val := c.FormValue("artifact-editor-demozooval")
	if prod == val {
		return c.NoContent(http.StatusNoContent)
	}
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	if err := model.UpdateDemozoo(int64(id), prod); err != nil {
		return badRequest(c, err)
	}
	return RecordLinks(c)
}

// RecordPouet handles the post submission for the file artifact Pouet production link.
func RecordPouet(c echo.Context) error {
	key := c.FormValue("artifact-editor-key")
	pouet := c.FormValue("artifact-editor-pouet")
	val := c.FormValue("artifact-editor-pouetval")
	if pouet == val {
		return c.NoContent(http.StatusNoContent)
	}
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	if err := model.UpdatePouet(int64(id), pouet); err != nil {
		return badRequest(c, err)
	}
	return RecordLinks(c)
}

// Record16Colors handles the post submission for the file artifact 16 Colors link.
func Record16Colors(c echo.Context) error {
	key := c.FormValue("artifact-editor-key")
	colors := c.FormValue("artifact-editor-16colors")
	val := c.FormValue("artifact-editor-16colorsval")
	if colors == val {
		return c.NoContent(http.StatusNoContent)
	}
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	link := form.SanitizeURLPath(colors)
	if err := model.Update16Colors(int64(id), link); err != nil {
		return badRequest(c, err)
	}
	return RecordLinks(c)
}

// RecordGitHub handles the post submission for the file artifact GitHub repository link.
func RecordGitHub(c echo.Context) error {
	key := c.FormValue("artifact-editor-key")
	github := c.FormValue("artifact-editor-github")
	val := c.FormValue("artifact-editor-githubval")
	if github == val {
		return c.NoContent(http.StatusNoContent)
	}
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	link := form.SanitizeGitHub(github)
	if err := model.UpdateGitHub(int64(id), link); err != nil {
		return badRequest(c, err)
	}
	return RecordLinks(c)
}

// RecordRelations handles the post submission for the file artifact releaser relationships.
func RecordRelations(c echo.Context) error {
	key := c.FormValue("artifact-editor-key")
	rels := c.FormValue("artifact-editor-relations")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	if err := model.UpdateRelations(int64(id), rels); err != nil {
		return badRequest(c, err)
	}
	return RecordLinks(c)
}

// RecordSites handles the post submission for the file artifact website links.
func RecordSites(c echo.Context) error {
	key := c.FormValue("artifact-editor-key")
	rels := c.FormValue("artifact-editor-websites")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, fmt.Errorf("%w: %w: %q", ErrKey, err, key))
	}
	if err := model.UpdateSites(int64(id), rels); err != nil {
		return badRequest(c, err)
	}
	return RecordLinks(c)
}

// RecordLinks handles the post submission for a form submission to provide the
// HTML formatted links for the "Links" section of the artifact editor.
func RecordLinks(c echo.Context) error {
	youtube := c.FormValue("artifact-editor-youtube")
	demozoo := c.FormValue("artifact-editor-demozoo")
	pouet := c.FormValue("artifact-editor-pouet")
	colors16 := c.FormValue("artifact-editor-16colors")
	github := c.FormValue("artifact-editor-github")
	rels := c.FormValue("artifact-editor-relations")
	sites := c.FormValue("artifact-editor-websites")
	s := app.LinkSamples(youtube, demozoo, pouet, colors16, github, rels, sites)
	return c.HTML(http.StatusOK, s)
}

// badRequest returns an error response with a 400 status code,
// the server cannot or will not process the request due to something that is perceived to be a client error.
func badRequest(c echo.Context, err error) error {
	return c.String(http.StatusBadRequest, err.Error())
}

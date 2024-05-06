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
	invalid := section == "" || platform == ""
	if invalid {
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

	// todo: confirm date has changed before updating

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

func RecordDateIssuedReset(c echo.Context, elmId string) error {
	reset := c.FormValue(elmId)
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, err)
	}

	vals := strings.Split(reset, "-")
	if len(vals) != 3 {
		return badRequest(c, fmt.Errorf("invalid reset date format, requires YYYY-MM-DD"))
	}

	year, month, day := vals[0], vals[1], vals[2]
	y, m, d := form.ValidDate(year, month, day)
	if !y || !m || !d {
		return badRequest(c, fmt.Errorf("invalid reset date format, requires YYYY-MM-DD"))
	}
	if err := model.UpdateDateIssued(int64(id), year, month, day); err != nil {
		return badRequest(c, err)
	}

	s := string(year)
	if month != "0" {
		s += "-" + month
	}
	if day != "0" {
		s += "-" + day
	}
	return c.String(http.StatusOK, s)
}

func RecordCreatorText(c echo.Context) error {
	creator := c.FormValue("artifact-editor-credittext")
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, err)
	}
	// todo validate creator to be a valid uri
	if err := model.UpdateCreatorText(int64(id), creator); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

func RecordCreatorIll(c echo.Context) error {
	creator := c.FormValue("artifact-editor-creditill")
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.UpdateCreatorIll(int64(id), creator); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

func RecordCreatorProg(c echo.Context) error {
	creator := c.FormValue("artifact-editor-creditprog")
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.UpdateCreatorProg(int64(id), creator); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

func RecordCreatorAudio(c echo.Context) error {
	creator := c.FormValue("artifact-editor-creditaudio")
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.UpdateCreatorAudio(int64(id), creator); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

func RecordCreatorReset(c echo.Context) error {
	reset := c.FormValue("artifact-editor-credit-resetter")
	resetText := c.FormValue("artifact-editor-credittext")
	resetIll := c.FormValue("artifact-editor-creditill")
	resetProg := c.FormValue("artifact-editor-creditprog")
	resetAudio := c.FormValue("artifact-editor-creditaudio")
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, err)
	}
	vals := strings.Split(reset, ";")
	if len(vals) != 4 {
		return badRequest(c, fmt.Errorf("invalid reset creators format, requires string;string;string;string"))
	}
	text := vals[0]
	ill := vals[1]
	prog := vals[2]
	audio := vals[3]

	fmt.Printf("text %q %q\n", text, resetText)

	if resetText == text && resetIll == ill && resetProg == prog && resetAudio == audio {
		return c.NoContent(http.StatusNoContent)
	}

	if err := model.UpdateCreators(int64(id), text, ill, prog, audio); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Undo creators")

}

func RecordComment(c echo.Context) error {
	comment := c.FormValue("artifact-editor-comment")
	key := c.FormValue("artifact-editor-key")
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.UpdateComment(int64(id), comment); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

func RecordCommentReset(c echo.Context) error {
	reset := c.FormValue("artifact-editor-comment-resetter")
	key := c.FormValue("artifact-editor-key")

	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.UpdateComment(int64(id), reset); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Undo comment")
}

func RecordYouTube(c echo.Context) error {
	key := c.FormValue("artifact-editor-key")
	watch := c.FormValue("artifact-editor-youtube")
	val := c.FormValue("artifact-editor-youtubeval")
	if watch == val {
		return c.NoContent(http.StatusNoContent)
	}
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.UpdateYouTube(int64(id), watch); err != nil {
		return badRequest(c, err)
	}
	return RecordLinks(c)
}

func RecordDemozoo(c echo.Context) error {
	key := c.FormValue("artifact-editor-key")
	prod := c.FormValue("artifact-editor-demozoo")
	val := c.FormValue("artifact-editor-demozooval")
	if prod == val {
		return c.NoContent(http.StatusNoContent)
	}
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.UpdateDemozoo(int64(id), prod); err != nil {
		return badRequest(c, err)
	}
	return RecordLinks(c)
}

func RecordPouet(c echo.Context) error {
	key := c.FormValue("artifact-editor-key")
	pouet := c.FormValue("artifact-editor-pouet")
	val := c.FormValue("artifact-editor-pouetval")
	if pouet == val {
		return c.NoContent(http.StatusNoContent)
	}
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.UpdatePouet(int64(id), pouet); err != nil {
		return badRequest(c, err)
	}
	return RecordLinks(c)
}

func Record16Colors(c echo.Context) error {
	key := c.FormValue("artifact-editor-key")
	colors := c.FormValue("artifact-editor-16colors")
	val := c.FormValue("artifact-editor-16colorsval")
	if colors == val {
		return c.NoContent(http.StatusNoContent)
	}
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.Update16Colors(int64(id), colors); err != nil {
		return badRequest(c, err)
	}
	return RecordLinks(c)
}

func RecordGitHub(c echo.Context) error {
	key := c.FormValue("artifact-editor-key")
	github := c.FormValue("artifact-editor-github")
	val := c.FormValue("artifact-editor-githubval")
	if github == val {
		return c.NoContent(http.StatusNoContent)
	}
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.UpdateGitHub(int64(id), github); err != nil {
		return badRequest(c, err)
	}
	return RecordLinks(c)
}

func RecordRelations(c echo.Context) error {
	key := c.FormValue("artifact-editor-key")
	rels := c.FormValue("artifact-editor-relations")
	val := c.FormValue("artifact-editor-relationsval")
	if rels == val {
		return c.NoContent(http.StatusNoContent)
	}
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.UpdateRelations(int64(id), rels); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

func RecordSites(c echo.Context) error {
	key := c.FormValue("artifact-editor-key")
	rels := c.FormValue("artifact-editor-websites")
	val := c.FormValue("artifact-editor-websitesval")
	if rels == val {
		return c.NoContent(http.StatusNoContent)
	}
	id, err := strconv.Atoi(key)
	if err != nil {
		return badRequest(c, err)
	}
	if err := model.UpdateSites(int64(id), rels); err != nil {
		return badRequest(c, err)
	}
	return c.String(http.StatusOK, "Updated")
}

func RecordLinks(c echo.Context) error {
	links := []string{}
	youtube := c.FormValue("artifact-editor-youtube")
	if youtube != "" {
		links = append(links, recordlinksRel("youtube.com/watch?v="+youtube))
	}
	demozoo := c.FormValue("artifact-editor-demozoo")
	if demozoo != "" {
		links = append(links, recordlinksRel("demozoo.org/productions/"+demozoo))
	}
	pouet := c.FormValue("artifact-editor-pouet")
	if pouet != "" {
		links = append(links, recordlinksRel("pouet.net/prod.php?which="+pouet))
	}
	colors16 := c.FormValue("artifact-editor-16colors")
	if colors16 != "" {
		links = append(links, recordlinksRel("16colo.rs/"+colors16))
	}
	github := c.FormValue("artifact-editor-github")
	if github != "" {
		links = append(links, recordlinksRel("github.com/"+github))
	}
	rels := c.FormValue("artifact-editor-link-releasers")
	if rels != "" {
		links = append(links, recordlinksRels(rels))
	}
	sites := c.FormValue("artifact-editor-link-websites")
	if sites != "" {
		links = append(links, recordlinksSites(sites))
	}
	s := strings.Join(links, "<br>")
	return c.HTML(http.StatusOK, s)
}

func recordlinksRel(url string) string {
	return `<a href="https://` + url + `">` + url + `</a>`
}

func recordlinksSites(rels string) string {
	links := strings.Split(rels, "|")
	hrefs := []string{}
	for _, link := range links {
		s := strings.Split(link, ";")
		if len(s) != 2 {
			continue
		}
		name := s[0]
		id := s[1]
		ref := `<a href="https://` + id + `">` + name + `</a>`
		hrefs = append(hrefs, ref)
	}
	return strings.Join(hrefs, " + ")
}

func recordlinksRels(sites string) string {
	//  "NFO;9f1c2|Intro;a92116e". Split by | and then by ;
	links := strings.Split(sites, "|")
	hrefs := []string{}
	for _, link := range links {
		// 0 = NFO, 1 = 9f1c2
		// 0 = Intro, 1 = a92116e
		s := strings.Split(link, ";")
		if len(s) != 2 {
			continue
		}
		name := s[0]
		id := s[1]
		ref := `<a href="/f/` + id + `">` + name + `</a>`
		hrefs = append(hrefs, ref)
	}
	return strings.Join(hrefs, " + ")
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

	unchanged := (rel1 == reset1 && rel2 == reset2)
	if unchanged {
		return c.NoContent(http.StatusNoContent)
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

	unchanged := (rel1 == reset1 && rel2 == reset2)
	if unchanged {
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

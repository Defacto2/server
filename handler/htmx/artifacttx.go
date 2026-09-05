package htmx

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/demozoo"
	"github.com/Defacto2/server/handler/form"
	"github.com/Defacto2/server/handler/jsdos"
	"github.com/Defacto2/server/handler/pouet"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v5"
)

// Package file artifacttx.go has all the funcs that use
// transactions and commits to change data kept in the database.

// TxReadmeOff handles the htmx request to disable the in
// page display of both the text files readme and file_id.diz for the file artifact.
func TxReadmeOff(c *echo.Context, tx *sql.Tx) error {
	return CommitNotOn(c, tx, "readme-is-off", model.ReadmeDisable.Update)
}

// TxFilename handles the post submission for the file artifact filename.
func TxFilename(c *echo.Context, tx *sql.Tx) error {
	return CommitSanitize(c, tx, "artifact-editor-filename", form.SanitizeFilename, model.Filename.Update)
}

// TxFilenameUndo handles the post submission for the file artifact filename reset.
func TxFilenameUndo(c *echo.Context, tx *sql.Tx) error {
	return CommitStr(c, tx, "artifact-editor-filename-undo", model.Filename.Update)
}

// TxVirusTotal handles the post submission for the file artifact VirusTotal report link.
func TxVirusTotal(c *echo.Context, tx *sql.Tx) error {
	return CommitStr(c, tx, "artifact-editor-virustotal", model.VirusTotal.Update)
}

// TxTitle handles the post submission for the file artifact title.
func TxTitle(c *echo.Context, tx *sql.Tx) error {
	return CommitStr(c, tx, "artifact-editor-title", model.Title.Update)
}

// TxTitleUndo handles the post submission for the file artifact title reset.
func TxTitleUndo(c *echo.Context, tx *sql.Tx) error {
	return CommitStr(c, tx, "artifact-editor-title-undo", model.Title.Update)
}

// TxComment handles the post submission for the file artifact comment.
func TxComment(c *echo.Context, tx *sql.Tx) error {
	return CommitStr(c, tx, "artifact-editor-comment", model.Comment.Update)
}

// TxCommentUndo handles the post submission for the file artifact comment reset.
func TxCommentUndo(c *echo.Context, tx *sql.Tx) error {
	return CommitStr(c, tx, "artifact-editor-comment-undo", model.Comment.Update)
}

// TxCreditText handles the post submission for the file artifact creator text.
func TxCreditText(c *echo.Context, tx *sql.Tx) error {
	return CommitSanitize(c, tx, "artifact-editor-credittext",
		form.SanitizeCredit, model.CreatorText.Update)
}

// TxCreditIll handles the post submission for the file artifact creator illustrator.
func TxCreditIll(c *echo.Context, tx *sql.Tx) error {
	return CommitSanitize(c, tx, "artifact-editor-creditill",
		form.SanitizeCredit, model.CreatorIll.Update)
}

// TxCreditProg handles the post submission for the file artifact creator programmer.
func TxCreditProg(c *echo.Context, tx *sql.Tx) error {
	return CommitSanitize(c, tx, "artifact-editor-creditprog",
		form.SanitizeCredit, model.CreatorProg.Update)
}

// TxCreditAudio handles the post submission for the file artifact creator musician.
func TxCreditAudio(c *echo.Context, tx *sql.Tx) error {
	return CommitSanitize(c, tx, "artifact-editor-creditaudio",
		form.SanitizeCredit, model.CreatorAudio.Update)
}

// TxDemozoo handles the post submission for the file artifact Demozoo production link.
func TxDemozoo(c *echo.Context, tx *sql.Tx) error {
	err := CommitStr(c, tx, "artifact-editor-demozoo", model.Demozoo.Update)
	return HTMLLinkTo(c, err)
}

// TxPouet handles the post submission for the file artifact Pouet production link.
func TxPouet(c *echo.Context, tx *sql.Tx) error {
	err := CommitStr(c, tx, "artifact-editor-pouet", model.Pouet.Update)
	return HTMLLinkTo(c, err)
}

// Tx16Colors handles the post submission for the file artifact 16 Colors link.
func Tx16Colors(c *echo.Context, tx *sql.Tx) error {
	err := CommitSanitize(c, tx, "artifact-editor-16colors",
		form.SanitizeURLPath, model.Colors16.Update)
	return HTMLLinkTo(c, err)
}

// TxGitHub handles the post submission for the file artifact GitHub repository link.
func TxGitHub(c *echo.Context, tx *sql.Tx) error {
	err := CommitSanitize(c, tx, "artifact-editor-github",
		form.SanitizeGitHub, model.GitHub.Update)
	return HTMLLinkTo(c, err)
}

// TxRelations handles the post submission for the file artifact releaser relationships.
func TxRelations(c *echo.Context, tx *sql.Tx) error {
	err := CommitStr(c, tx, "artifact-editor-relations", model.Relations.Update)
	return HTMLLinkTo(c, err)
}

// TxWebsites handles the post submission for the file artifact website links.
func TxWebsites(c *echo.Context, tx *sql.Tx) error {
	err := CommitStr(c, tx, "artifact-editor-websites", model.Sites.Update)
	return HTMLLinkTo(c, err)
}

func TxEmulateUMB(c *echo.Context, tx *sql.Tx) error {
	return CommitOn(c, tx, "emulate-ram-umb", model.EmulateUMB.Update)
}

func TxEmulateEMS(c *echo.Context, tx *sql.Tx) error {
	return CommitOn(c, tx, "emulate-ram-ems", model.EmulateEMS.Update)
}

func TxEmulateXMS(c *echo.Context, tx *sql.Tx) error {
	return CommitOn(c, tx, "emulate-ram-xms", model.EmulateXMS.Update)
}

// TxEmulateBroken handles the patch submission for the broken emulation for a file artifact.
func TxEmulateBroken(c *echo.Context, tx *sql.Tx) error {
	return CommitNotOn(c, tx, "emulate-is-broken", model.EmulateBroken.Update)
}

// TxEmulateMachine handles the patch submission for the machine and graphic emulation for a file artifact.
func TxEmulateMachine(c *echo.Context, tx *sql.Tx) error {
	return CommitStrKey(c, tx, "emulate-machine", model.UpdateEmulateMachine)
}

// TxEmulateCPU handles the patch submission for the CPU emulation for a file artifact.
func TxEmulateCPU(c *echo.Context, tx *sql.Tx) error {
	return CommitStrKey(c, tx, "emulate-cpu", model.UpdateEmulateCPU)
}

// TxEmulateSFX handles the patch submission for the audio emulation for a file artifact.
func TxEmulateSFX(c *echo.Context, tx *sql.Tx) error {
	return CommitStrKey(c, tx, "emulate-sfx", model.UpdateEmulateSfx)
}

// TxEmulateRunProg handles the patch submission for the run program emulation.
func TxEmulateRunProg(c *echo.Context, tx *sql.Tx) error {
	const name = "emulate-run-program"
	const format = "tx emulate run prog: %w"
	if err := nils.Check(c, tx); err != nil {
		return fmt.Errorf(format, err)
	}

	key, err := KeyParam(c)
	if err != nil {
		return badRequest(c, err)
	}

	div := func(ok bool, s string) string {
		class := "text-success"
		if !ok {
			class = `d-block invalid-feedback`
		}
		return `<div id="emulate-run-program-feedback" class="` + class + `">` + s + `</div>`
	}

	val := strings.ToUpper(c.FormValue(name))
	if !jsdos.Valid(val) {
		s := `The command, or name contains invalid characters; or is too long`
		return c.String(http.StatusOK, div(false, s))
	}

	ctx := c.Request().Context()
	if err = model.UpdateEmulateRunProgram(ctx, tx, key, val); err != nil {
		return badRequest(c, err)
	}

	const custom = "✓ Custom command(s) "
	if val == "" {
		return c.String(http.StatusOK, div(true, custom+"removed"))
	}
	return c.String(http.StatusOK, div(true, custom+"saved"))
}

// HTMLLinkTo handles the post submission for a form submission to provide the
// HTML formatted links for the "Links" section of the artifact editor.
func HTMLLinkTo(c *echo.Context, err error) error {
	if err != nil {
		return badRequest(c, err)
	}

	youtube := c.FormValue("artifact-editor-youtube")
	demozoo := c.FormValue("artifact-editor-demozoo")
	pouet := c.FormValue("artifact-editor-pouet")
	colors16 := c.FormValue("artifact-editor-16colors")
	github := c.FormValue("artifact-editor-github")
	rels := c.FormValue("artifact-editor-relations")
	sites := c.FormValue("artifact-editor-websites")

	links := app.LinkPreviews(youtube, demozoo, pouet, colors16, github, rels, sites)
	for i, link := range links {
		links[i] = `<small><strong>Link to</strong></small> &nbsp; ` + link
	}
	return c.HTML(http.StatusOK, strings.Join(links, "<br>"))
}

// TxLinksUndo handles the post submission for the file artifact links reset.
func TxLinksUndo(c *echo.Context, tx *sql.Tx) error {
	const format = "tx links undo %s: %w"
	if err := nils.Check(c, tx); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	key, err := Key(c)
	if err != nil {
		return badRequest(c, fmt.Errorf(format, "key", err))
	}

	youtube := c.FormValue("artifact-editor-youtubeval") // FIX: replace tailname with undo
	if ok := form.ValidYouTube(youtube); !ok {
		return badRequest(c, fmt.Errorf(format, youtube, ErrYouTube))
	}

	colors16 := c.FormValue("artifact-editor-16colorstval")
	colors16 = form.SanitizeURLPath(colors16)

	github := c.FormValue("artifact-editor-githubval")
	github = form.SanitizeGitHub(github)

	demozooVal := c.FormValue("artifact-editor-demozooval")
	var demozooID int64
	if demozooVal != "" {
		demozooID, err = strconv.ParseInt(demozooVal, 10, 64)
		if err != nil {
			return badRequest(c, fmt.Errorf(format, "demozoo id "+demozooVal, err))
		}
		if demozooID > demozoo.Sanity {
			return badRequest(c, fmt.Errorf(format, "demozoo id does not exist "+demozooVal, err))
		}
	}

	pouetVal := c.FormValue("artifact-editor-pouetval")
	var pouetID int64
	if pouetVal != "" {
		pouetID, err = strconv.ParseInt(pouetVal, 10, 64)
		if err != nil {
			return badRequest(c, fmt.Errorf(format, "pouet id "+pouetVal, err))
		}
		if pouetID > pouet.Sanity {
			return badRequest(c, fmt.Errorf(format, "pouet id does not exist "+pouetVal, err))
		}
	}

	rels := c.FormValue("artifact-editor-relationsval")
	sites := c.FormValue("artifact-editor-websitesval")
	lnks := model.Links{
		ID:        key,
		Demozoo:   demozooID,
		Pouet:     pouetID,
		YouTube:   youtube,
		Colors16:  colors16,
		GitHub:    github,
		Relations: rels,
		Sites:     sites,
	}
	ctx := c.Request().Context()
	if err := lnks.Update(ctx, tx); err != nil {
		return badRequest(c, err)
	}

	links := app.LinkPreviews(youtube, demozooVal, pouetVal, colors16, github, rels, sites)
	for i, link := range links {
		links[i] = `<small><strong>Link to</strong></small> &nbsp; ` + link
	}
	return c.HTML(http.StatusOK, strings.Join(links, "<br>"))
}

// TxCreditUndo handles the post submission for the file artifact creators reset.
func TxCreditUndo(c *echo.Context, tx *sql.Tx) error {
	const format = "tx credit undo: %w"
	if err := nils.Check(c, tx); err != nil {
		return fmt.Errorf(format, err)
	}

	key, err := Key(c)
	if err != nil {
		return badRequest(c, fmt.Errorf(format, err))
	}

	reset := c.FormValue("artifact-editor-credits-undo")
	vals := strings.Split(reset, ";")
	const expected = 4
	if len(vals) != expected {
		const format = `%w, record creator reset requires string;string;string;string`
		return badRequest(c, fmt.Errorf(format, ErrYMD))
	}

	text := vals[0]
	ill := vals[1]
	prog := vals[2]
	audio := vals[3]
	creators := model.Creators{
		ID:    key,
		Text:  text,
		Ill:   ill,
		Prog:  prog,
		Audio: audio,
	}

	// INFO: form values match the "name" attribute of html elements
	textval := c.FormValue("artifact-editor-credittext")
	illval := c.FormValue("artifact-editor-creditill")
	progval := c.FormValue("artifact-editor-creditprog")
	audioval := c.FormValue("artifact-editor-creditaudio")
	if textval == text && illval == ill && progval == prog && audioval == audio {
		return c.NoContent(http.StatusNoContent) //nolint:wrapcheck
	}

	ctx := c.Request().Context()
	if err := creators.Update(ctx, tx); err != nil {
		return badRequest(c, err)
	}

	return StatusOK(c, "artifact-editor-credits-undo")
}

// TxTags handles the post submission for the file artifact classifications,
// such as the platform, operating system, section or category tags.
// The return value is either the humanized and counted classification or an error.
func TxTags(sl *slog.Logger, c *echo.Context, tx *sql.Tx) error {
	const format = "tx tags: %w"
	if err := nils.Check(sl, c, tx); err != nil {
		return fmt.Errorf(format, err)
	}

	section := c.FormValue("artifact-editor-categories")
	platform := c.FormValue("artifact-editor-operatingsystem")
	ctx := c.Request().Context()
	if invalid := section == "" || platform == ""; invalid {
		html, err := form.HumanizeCount(ctx, tx, section, platform)
		if err != nil {
			sl.Error("record classification", slog.Any("error", err))
			return badRequest(c, err)
		}
		return c.HTML(http.StatusOK, string(html)+" did not update")
	}

	key, err := Key(c)
	if err != nil {
		return badRequest(c, fmt.Errorf(format, err))
	}

	classification := model.Classification{
		ID: key, Platform: platform, Tag: section,
	}
	if err := classification.Update(ctx, tx); err != nil {
		return badRequest(c, err)
	}

	html, err := form.HumanizeCount(ctx, tx, section, platform)
	if err != nil {
		sl.Error("record classification", slog.Any("error", err))
		return badRequest(c, err)
	}

	return c.HTML(http.StatusOK, string(html))
}

// TxPublic handles the post submission for the file artifact record toggle.
// The return value is either "online" or "offline" depending on the state.
func TxPublic(c *echo.Context, tx *sql.Tx, state bool) error {
	const format = "tx public %s: %w"
	if err := nils.Check(c, tx); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	key, err := Key(c)
	if err != nil {
		return badRequest(c, fmt.Errorf(format, "key", err))
	}

	ctx := c.Request().Context()
	if err := model.UpdateOnline(ctx, tx, state, key); err != nil {
		return fmt.Errorf(format, "update", err)
	}

	if state {
		return c.String(http.StatusOK, "online")
	}
	return c.String(http.StatusOK, "offline")
}

// TxPublicByKey handles the post submission for the file artifact record toggle.
// The key string is converted into an integer and used as the artifact id.
// The return value is either "online" or "offline" depending on the state.
func TxPublicByKey(c *echo.Context, tx *sql.Tx, key string, state bool) error {
	const format = "tx public by key %d, %s: %w"
	if err := nils.Check(c, tx); err != nil {
		return fmt.Errorf(format, 0, "check", err)
	}

	id, err := strconv.ParseInt(key, 10, 64)
	if err != nil {
		return badRequest(c, fmt.Errorf(format, 0, key, err))
	}

	ctx := c.Request().Context()
	if err := model.UpdateOnline(ctx, tx, state, id); err != nil {
		return fmt.Errorf(format, id, "update", err)
	}

	if state {
		return c.String(http.StatusOK,
			"Record is visible to the public.")
	}
	const prohibited = "🚫"
	return c.String(http.StatusOK,
		prohibited+" Record is disabled and hidden from public access. "+prohibited)
}

// TxYouTube handles the post submission for the file artifact YouTube watch video link.
func TxYouTube(c *echo.Context, tx *sql.Tx) error {
	const format = "tx youtube: %w"
	if err := nils.Check(c, tx); err != nil {
		return fmt.Errorf(format, err)
	}

	key, err := Key(c)
	if err != nil {
		return badRequest(c, fmt.Errorf(format, err))
	}

	val := strings.TrimSpace(c.FormValue("artifact-editor-youtube"))
	if ok := form.ValidYouTube(val); !ok {
		return c.NoContent(http.StatusNoContent) //nolint:wrapcheck
	}

	ctx := c.Request().Context()
	if err := model.YouTube.Update(ctx, tx, key, val); err != nil {
		return badRequest(c, err)
	}

	return HTMLLinkTo(c, err)
}

// TxReleasers handles the post submission for the file artifact releasers.
// It will only update the releaser1 and the releaser2 values if they have changed.
// The return value is either "Updated" or "Update" depending on if the values have changed.
func TxReleasers(c *echo.Context, tx *sql.Tx) error {
	return txReleasers(c, tx, false)
}

// TxReleasersUndo handles the post submission for the file artifact releasers reset.
// It will always reset and save the releaser1 and the releaser2 values.
// The return value is always "Resetted" unless an error occurs.
func TxReleasersUndo(c *echo.Context, tx *sql.Tx) error {
	return txReleasers(c, tx, true)
}

func txReleasers(c *echo.Context, tx *sql.Tx, undo bool) error {
	const format = "tx releasers %s: %w"
	if err := nils.Check(c, tx); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	key, err := Key(c)
	if err != nil {
		return fmt.Errorf(format, "key", err)
	}

	undo1 := c.FormValue("releaser1")
	undo2 := c.FormValue("releaser2")

	relr1 := c.FormValue("artifact-editor-releaser1")
	relr2 := c.FormValue("artifact-editor-releaser2")

	unchanged := (relr1 == undo1 && relr2 == undo2)
	if unchanged {
		return c.String(http.StatusNoContent, "")
	}

	relFor := relr1
	relBy := relr2
	if undo {
		relFor = undo1
		relBy = undo2
	}

	ctx := c.Request().Context()
	if err := model.UpdateReleasers(ctx, tx, key, relFor, relBy); err != nil {
		return badRequest(c, err)
	}

	return StatusOK(c, "artifact-editor-releasers")
}

// YMD are the date of release form input options
type YMD int

const (
	DateForm YMD = iota // DateForm uses the form inputs
	DateUndo            // DateUndo uses the revert toggle
	DateLast            // DateLast uses the date lastmodification toggle
)

// TxYMD is the post submission for the date of release.
func TxYMD(c *echo.Context, tx *sql.Tx) error {
	return DateForm.commit(c, tx)
}

// TxYMDUndo is the post submission to undo the date of release.
func TxYMDUndo(c *echo.Context, tx *sql.Tx) error {
	return DateUndo.commit(c, tx)
}

// TxLastMod is the post submission to set the date of release
// to be the last modification date of the file download or file archive.
func TxLastMod(c *echo.Context, tx *sql.Tx) error {
	return DateLast.commit(c, tx)
}

func (ymd YMD) commit(c *echo.Context, tx *sql.Tx) error {
	const format = "tx ymd commit %s: %w"
	if err := nils.Check(c, tx); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	key, err := Key(c)
	if err != nil {
		return fmt.Errorf(format, "key", err)
	}

	y, m, d := "", "", ""

	undoY := c.FormValue("artifact-editor-year-store")
	undoM := c.FormValue("artifact-editor-month-store")
	undoD := c.FormValue("artifact-editor-day-store")

	year := c.FormValue("artifact-editor-year")
	month := c.FormValue("artifact-editor-month")
	day := c.FormValue("artifact-editor-day")

	if ymd == DateForm {
		unchanged := year == undoY && month == undoM && day == undoD
		switch {
		case unchanged:
			return c.String(http.StatusNoContent, "")
		case ymd == DateForm:
			y = year
			m = month
			d = day
		}
	}

	const fmtymd = `%w, record date issued reset requires YYYY-MM-DD`

	name := ""
	if ymd == DateLast {
		name = "artifact-editor-date-lastmods"
	}
	if ymd == DateUndo {
		name = "artifact-editor-date-undos"
	}
	if ymd == DateLast || ymd == DateUndo {
		s := c.FormValue(name)
		values := strings.Split(s, "-")
		const req = 3
		if len(values) != req {
			return badRequest(c, fmt.Errorf(fmtymd, ErrYMD))
		}
		y = values[0]
		m = values[1]
		d = values[2]
		unchanged := year == y && month == m && day == d
		if unchanged {
			return c.String(http.StatusNoContent, "")
		}
	}

	vy, vm, vd := form.ValidDate(y, m, d)
	if ok := vy && vm && vd; !ok {
		return badRequest(c, fmt.Errorf(fmtymd, ErrYMD))
	}

	ctx := c.Request().Context()
	if err := model.UpdateYMDS(ctx, tx, key, y, m, d); err != nil {
		return badRequest(c, err)
	}

	switch ymd {
	case DateUndo:
		return StatusOK(c, "artifact-editor-date-undo")
	default:
		return StatusOK(c, "artifact-editor-date-update")
	}
}

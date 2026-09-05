package app

// Package file dirs.go contains the artifact page directories and handlers.
//
// Public funcs are intended for the handler/router.
// Private funcs are internal to the app package.

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Defacto2/archive/rezip"
	"github.com/Defacto2/helper"
	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/handler/app/internal/filerecord"
	"github.com/Defacto2/server/handler/app/internal/simple"
	"github.com/Defacto2/server/handler/readme"
	"github.com/Defacto2/server/handler/sess"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/bengarrett/sauce"
	"github.com/dustin/go-humanize"
	"github.com/labstack/echo/v5"
	_ "golang.org/x/image/webp" // webp format decoder
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

const (
	PcEpoch  = model.EpochYear // pc epoch is the default year for PC/MS-DOS files without a timestamp
	CgaEpoch = 1988            // cga epoch is the final year to use the IBM PC CGA font for viewing textfiles

	// NOTE: limit to 200 items for display, "view content" + "Download content",
	// high limits take longer to render.
	maxArchiveItems = 200
	// NOTE: skip the rendering of readme text files that are larger than 1MB.
	sizeLimitBytes = 1_000_000
	// NOTE: skip the render of the "view content" button for 1MB > zip content textdata,
	// limit is ignored by Editors.
	maxZipContent = 1_000_000
)

// Dirs contains the directories used by the artifact pages.
type Dirs struct {
	Download  dir.Directory // path to the artifact download directory
	Preview   dir.Directory // path to the preview and screenshot directory
	Thumbnail dir.Directory // path to the file thumbnail directory
	Extra     dir.Directory // path to the extra files directory
	URI       string        // the URI of the File record
	ID        int64         // the database key of the File record
	UUID      string        // the database uuid key of the File record
	Platform  string        // the platform of the File record
	Section   string        // the section of the File record
	Magic     string        // the magic btyes of the File record
	Maximum   struct {
		Items  int   // maximum archive items to display
		Readme int64 // maximum size of text files to display
	}
	ReadOnly bool // disable editing features
}

// Artifact is the app handler for the file record that is used by the /f/[key] route,
// and is rendered by the app/artifact.tmpl and app/artifactedit.tmpl views.
func (ds *Dirs) Artifact(sl *slog.Logger, c *echo.Context, db *sql.DB) error { //nolint:funlen
	const format = "dirs artifact context %s: %w"
	const uri = "artifact"
	if err := nils.Check(sl, c, db); err != nil {
		return fmt.Errorf(format, "arguments", err)
	}

	ctx := c.Request().Context()
	art, err := ds.oneByKey(ctx, sl, c, db)
	if art == nil {
		return fmt.Errorf(format, "", ErrArtifact)
	}
	if err != nil {
		return fmt.Errorf(format, "one by key", err)
	}

	ds.ID = art.ID
	ds.UUID = filerecord.UnID(art)
	ds.Platform = filerecord.TagProgram(art)
	ds.Section = filerecord.TagCategory(art)
	ds.Magic = filerecord.Magic(art)
	ds.Maximum.Items = maxArchiveItems
	ds.Maximum.Readme = sizeLimitBytes

	// must always init an empty data map
	data := empty(c)

	if edit := !ds.ReadOnly && sess.Editor(c); edit {
		data = ds.addEditor(ctx, sl, c, art, data)
	}
	data = ds.fixANSI(ctx, sl, db, data)

	// webpage template metadata
	downloadURI := filerecord.DownloadID(art)
	h1 := filerecord.FirstHeader(art)
	data["canonical"] = strings.Join([]string{"f", downloadURI}, "/")
	data["unid"] = ds.UUID
	data["download"] = downloadURI
	data["title"] = filerecord.Basename(art)
	data["description"] = filerecord.Description(art)
	data["h1"] = h1
	data["ogtitle"] = h1
	data["lead"] = firstLead(art)
	data["comment"] = string(helper.MaskTerm([]byte(filerecord.Comment(art))...))

	data = ds.addDownload(art, data)
	if fix := !ds.ReadOnly && sess.Editor(c); fix {
		data = ds.fixMagic(ctx, sl, db, data)
	}

	data = ds.addAttribution(art, data)
	data = ds.addLinks(art, data)
	data = ds.addEmu(sl, art, data)

	content := art.FileZipContent.String
	if ok := len(content) <= maxZipContent; !ok && !sess.Editor(c) {
		data["contentDesc"] = "contains many files"
	} else {
		data = ds.addZip(content, data)
	}

	data["linkpreview"] = filerecord.LinkPreview(art)
	data["linkpreviewTip"] = filerecord.LinkPreviewTip(art)
	data["filentry"] = filerecord.FileEntry(art)

	if ok := ds.screenshot(); !ok {
		data["noScreenshot"] = true
	}

	if text := !filerecord.UnsupportedFile(art); text {
		data, err = ds.addReadme(sl, art, data)
		if err != nil {
			defer clear(data)
			ds.logErr(sl, "dirs artifact add readme", err)
		}
	} else {
		data = ds.addSAUCE(art, data)
	}

	err = c.Render(http.StatusOK, uri, data)
	defer clear(data)
	if err != nil {
		return InternalErr(sl, c, uri, errorWithID(err, ds.URI, art.ID))
	}

	return nil
}

// logErr logs the err as an error while also providing id and uri context from Dirs.
func (ds *Dirs) logErr(sl *slog.Logger, msg string, err error) {
	if sl == nil {
		return
	}
	sl.Error(msg,
		slog.String("uri", ds.URI),
		slog.Int64("id", ds.ID),
		slog.String("uuid", ds.UUID),
		// slog.String("name", dir.Filename),
		slog.Any("error", err))
}

// addAttribution returns the authors and attributions of the artifact.
func (ds *Dirs) addAttribution(art *models.File, data map[string]any) map[string]any {
	if art == nil {
		return data
	}

	data["writers"] = filerecord.AttrWriter(art)
	data["artists"] = filerecord.AttrArtist(art)
	data["programmers"] = filerecord.AttrProg(art)
	data["musicians"] = filerecord.AttrMusic(art)

	return data
}

// addDownload returns the metadata for the artifact download file.
func (ds *Dirs) addDownload(art *models.File, data map[string]any) map[string]any {
	if art == nil {
		return data
	}

	data["filename"] = filerecord.Basename(art)
	data["filesize"] = simple.BytesHuman(art.Filesize.Int64)
	data["filebyte"] = art.Filesize.Int64
	data["lastmodified"] = filerecord.LastModification(art)
	data["lastmodifiedAgo"] = filerecord.LastModificationAgo(art)
	data["checksum"] = filerecord.Checksum(art)
	data["magic"] = filerecord.Magic(art)
	data["releasers"] = releasersHrefs(art)
	data["published"] = filerecord.Date(art)
	data["section"] = filerecord.TagCategory(art)
	data["platform"] = filerecord.TagProgram(art)
	data["extraZip"] = filerecord.ExtraZip(art, ds.Extra)

	alert := filerecord.AlertURL(art)
	data["alertURL"] = alert
	if strings.TrimSpace(alert) != "" {
		data["noindex"] = true
	}

	return data
}

// addEditor adds editoring data for the artifact.
//
// These are the editable fields only visible to website editors after they have logged in.
func (ds *Dirs) addEditor(
	ctx context.Context, sl *slog.Logger, c *echo.Context, art *models.File, data map[string]any,
) map[string]any {
	if nils.Slog("dirs add editor", ctx, sl, c, art) {
		return data
	}

	paths := command.Dirs{
		Download:  ds.Download,
		Preview:   ds.Preview,
		Thumbnail: ds.Thumbnail,
		Extra:     ds.Extra,
	}
	unid := filerecord.UnID(art)
	path := filepath.Join(ds.Download.Path(), unid)

	data["epochYear"] = PcEpoch
	data["readonlymode"] = false
	data["modID"] = art.ID
	data["modTitle"] = filerecord.Title(art)
	data["modOnline"] = filerecord.RecordOnline(art)
	data["modReleasers"] = RecordRels(art.GroupBrandBy, art.GroupBrandFor)
	data["modReleaser1"], data["modReleaser2"] = filerecord.ReleaserPair(art)
	data["modYear"], data["modMonth"], data["modDay"] = filerecord.Dates(art)
	data["modLMYear"], data["modLMMonth"], data["modLMDay"] = filerecord.LastModifications(art)
	data["modAbsDownload"] = path
	data["modMagicMime"] = simple.MIME(sl, path)
	data["modMagicNumber"] = simple.MagicAsTitle(sl, path)
	data["modDBModify"] = filerecord.LastModificationDate(art)
	data["modStatModify"], data["modStatSizeB"], data["modStatSizeF"] = simple.StatHumanize(path)
	data["modDecompress"] = filerecord.ListContent(ctx, sl, ds.Maximum.Items, art, paths, path)
	if sess.Editor(c) {
		data["modDecompressLoc"] = simple.MkdirStale(sl, path)
	}
	// WARN: These operations must be done using os.Stat and not os.ReadDir or filepath.WalkDir.
	// Previous attempts to use a shared function with WalkDir caused a memory leakages when
	// the site was under heavy load.
	data["modAssetPreview"] = ds.pathsPreview(sl)
	data["modAssetThumbnail"] = ds.pathsThumb(sl)
	data["modAssetExtra"] = ds.pathsExtras()
	data["missingAssets"] = ds.pathsSuggest()
	data["modReadmeSuggest"] = filerecord.Readme(art)
	data["disableReadme"] = filerecord.DisableReadme(art)
	data["modZipContent"] = filerecord.ZipContent(art)
	data["modRelations"] = filerecord.RelationsStr(art)
	data["modWebsites"] = filerecord.WebsitesStr(art)
	data["modOS"] = filerecord.TagProgram(art)
	data["modTag"] = filerecord.TagCategory(art)
	data["alertURL"] = filerecord.AlertURL(art)
	data["forApproval"] = filerecord.RecordIsNew(art)
	data["disableApproval"] = filerecord.RecordProblems(art)
	data["disableRecord"] = filerecord.RecordOffline(art)
	data["modEmulateXMS"], data["modEmulateEMS"], data["modEmulateUMB"] = filerecord.JsdosMemory(art)
	data["modEmulateBroken"] = filerecord.JsdosBroken(art)
	data["modEmulateRun"] = filerecord.JsdosRun(art)
	data["modEmulateCPU"] = filerecord.JsdosCPU(art)
	data["modEmulateMachine"] = filerecord.JsdosMachine(art)
	data["modEmulateAudio"] = filerecord.JsdosSound(art)

	return data
}

// addEmu returns the js-dos emulator data for the file record of the artifact.
func (ds *Dirs) addEmu(sl *slog.Logger, art *models.File, data map[string]any,
) map[string]any {
	const msg = "jsdos emulator"
	if nils.Slog(msg, sl, art) {
		return data
	}

	data["jsdos6"] = false
	data["jsdos6Run"] = ""
	data["jsdos6RunGuess"] = ""
	data["jsdos6Config"] = ""
	data["jsdos6Zip"] = false
	data["jsdos6Utilities"] = false

	if emulate := filerecord.JsdosUse(art); !emulate {
		return data
	}

	runCmd, err := model.JsDosCommand(art)
	if err != nil {
		ds.logErr(sl, msg+" command", err)
		return data
	}
	data["jsdos6Run"] = runCmd

	runGuess, err := model.JsDosBinary(art)
	if err != nil {
		ds.logErr(sl, msg+" binary", err)
		return data
	}
	data["jsdos6RunGuess"] = runGuess

	config, err := model.JsDosConfig(art)
	if err != nil {
		ds.logErr(sl, msg+" config", err)
		return data
	}
	data["jsdos6"] = true
	data["jsdos6Config"] = string(config)
	data["jsdos6Zip"] = filerecord.JsdosArchive(art)
	data["jsdos6Utilities"] = filerecord.JsdosUtilities(art)

	return data
}

// addLinks returns the other relations and external links for the file record of the artifact.
func (ds *Dirs) addLinks(art *models.File, data map[string]any) map[string]any {
	if art == nil {
		return data
	}

	data["demozoo"] = filerecord.IdenficationDZ(art)
	data["github"] = filerecord.IdenficationGitHub(art)
	data["pouet"] = filerecord.IdenficationPouet(art)
	data["relations"] = filerecord.Relations(art)
	data["sixteenColors"] = filerecord.Idenfication16C(art)
	data["youtube"] = filerecord.IdenficationYT(art)
	data[websites] = filerecord.Websites(art)

	return data
}

// addReadme can append either a plain textfile, ansi encoded text file, or binary text file to the data map.
// Also handled is any embedded SAUCE metadata, that will be shown as "embed" information on the artifact page.
func (ds *Dirs) addReadme(sl *slog.Logger, art *models.File, data map[string]any) (map[string]any, error) {
	// these are required by the template and must always be set
	data["contentBinary"] = ""
	data["readmeSAUCE"] = false
	data["sauceTitle"] = ""
	data["sauceAuthor"] = ""
	data["sauceGroup"] = ""
	data["sauceDate"] = ""

	const format = "dirs textfiles: %s: %w"
	if err := nils.Check(sl, art); err != nil {
		return data, fmt.Errorf(format, "arguments", err)
	}

	text := readme.Text{ //nolint:exhaustruct_v5
		Download: ds.Download,
		Extra:    ds.Extra,
		UUID:     art.UUID.String,
		Filename: art.Filename.String,
		Platform: art.Platform.String,
		Section:  art.Section.String,
		Year:     art.DateIssuedYear.Int16,
		MaxSize:  ds.Maximum.Readme,
	}
	textBuf, runeBuf, err := text.Buffers(sl)
	if err != nil {
		if errors.Is(err, ErrDownload) {
			data["noDownload"] = true
			return data, nil
		}
		return data, fmt.Errorf(format, "read", err)
	}

	rec := text.Record
	if add := rec.ID == "SAUCE"; add {
		data["readmeSAUCE"] = true
		data["sauceTitle"] = rec.Title
		data["sauceAuthor"] = rec.Author
		data["sauceGroup"] = rec.Group
		data["sauceDate"] = rec.Date.Time.Format("2006 Jan 02")
	}

	// handle edge cases where the textBuf is empty
	if ok := textBuf.Len() > 0; !ok {
		return data, nil
	}

	classElems := [...]string{"reader-invert", "border", "border-black", "rounded-1", "p-1"}
	data["preElementClass"] = strings.Join(classElems[:], " ")

	// ansi and binary text files are handled by a different template
	if notRAW := runeBuf == nil; notRAW {
		data = ds.addTextBinary(art, textBuf, classElems[:], data)
		return data, nil
	}

	data = ds.addText8bit(sl, art, textBuf, data)
	data = ds.addTextUTF8(textBuf, runeBuf, data)

	return data, nil
}

// addSAUCE is used to read the SAUCE metadata embed into data files such
// as pictures and image files.
func (ds *Dirs) addSAUCE(art *models.File, data map[string]any) map[string]any {
	if art == nil {
		return data
	}

	name := filepath.Join(ds.Download.Path(), art.UUID.String)
	file, err := os.Open(name)
	if err != nil {
		return data
	}
	defer file.Close()

	rec, err := sauce.Read(file)
	if err != nil {
		return data
	}
	if ok := rec.ID == "SAUCE"; !ok {
		return data
	}
	data["readmeSAUCE"] = true
	data["sauceTitle"] = rec.Title
	data["sauceAuthor"] = rec.Author
	data["sauceGroup"] = rec.Group
	data["sauceDate"] = rec.Date.Time.Format("2006 Jan 02")

	return data
}

// addText8bit appends the text content for files that are encoded using the
// legacy character map encodings IBM CodePage 437 or ISO-8859-1 (Latin-1). In the case of
// CP437, the encoding goes through a conversion to modern Unicode (UTF-8). Allowing the text to be
// served correctly by the web server and viewable in a standard browser.
//
// All text content, either CP437, ISO, or UTF-8, also goes through a normalization process,
// to replace any "special" characters, such as non-breaking-spaces with standard spaces.
func (ds *Dirs) addText8bit(sl *slog.Logger, art *models.File, //nolint:funlen
	textBuf *bytes.Buffer, data map[string]any,
) map[string]any {
	if nils.Slog("dirs add text 8bit", sl, art, textBuf) {
		return data
	}
	if disable := art.RetrotxtNoReadme.Int16 != 0; disable {
		return data
	}
	const maxWidth = 80

	// strip any RTF formatting
	b := textBuf.Bytes()
	if simple.RTF(b) {
		b = simple.StripRTF(b)
	}
	if len(b) == 0 {
		return data
	}

	year, _, _ := filerecord.Dates(art)
	fontname := "font-dos font-large"
	if year != 0 && year <= CgaEpoch {
		fontname = "font-dos-mda"
	}

	data["topazCheck"] = ""
	data["vgaCheck"] = ""
	data["preClassLatin1"] = "d-none "
	data["preClassCP437"] = "d-none " + fontname

	platform, _ := data["platform"].(string)
	topazFont := platform == textamiga || platform == "console"
	if topazFont {
		data["topazCheck"] = "checked"
		data["preClassLatin1"] = ""
	} else {
		data["vgaCheck"] = "checked"
		data["preClassCP437"] = fontname
	}
	data["forceSimpleText"] = filerecord.ForceSimpleText(art)

	if LockIn80Columns(year, b...) {
		b = lockWidth(maxWidth, b)
	}

	byteEnc := ds.encoding(bytes.NewReader(b))
	switch byteEnc {
	case charmap.ISO8859_1:
		b = bytes.ReplaceAll(b, byteNBSP, byteSpace)
		b = bytes.ReplaceAll(b, byteSHY, byteHyphen)
	case charmap.CodePage437, unicode.UTF8:
		b = bytes.ReplaceAll(b, byteNBSP437, byteSpace)
	}

	switch byteEnc {
	case unicode.UTF8:
		body, err := decode(bytes.NewReader(b))
		if err != nil {
			ds.logErr(sl, "unicode utf-8", err)
			return data
		}
		data["contentLatin1"] = body
		data["contentCP437"] = body
		data["contentLines"] = strings.Count(body, "\n")
		data["contentRows"] = helper.MaxLineLength(body)

		return data
	default:
		// latin-1 iso8859-1 text
		isomap := charmap.ISO8859_1.NewDecoder().Reader(bytes.NewReader(b))
		isobody, err := decode(isomap)
		if err != nil {
			ds.logErr(sl, "iso8859-1", err)
			return data
		}
		data["contentLatin1"] = isobody

		// codepage-437 text
		cpmap := charmap.CodePage437.NewDecoder().Reader(bytes.NewReader(b))
		cpbody, err := decode(cpmap)
		ds.logErr(sl, "cp-437", err)
		if err != nil {
			return data
		}
		data["contentLines"] = strings.Count(cpbody, "\n")

		// stats
		data["contentRows"] = helper.MaxLineLength(cpbody)
		data["contentCP437"] = cpbody
	}

	return data
}

// addTextBinary prepares the viewer for both binary encoded text and ansi escape encoded texts.
//
// The elems entries are for individual CSS class names that will be used with "preElementClss".
func (ds *Dirs) addTextBinary(art *models.File, textBuf *bytes.Buffer,
	elems []string, data map[string]any,
) map[string]any {
	if nils.Slog("dirs add text binary", art, textBuf) {
		return data
	}

	data["contentBinary"] = template.HTML(textBuf.String())
	data["contentAmigaAnsi"] = ""
	data["contentBinarySwappers"] = true
	data["contentBinarySwapper"] = 0 // placeholder, future use for a possible VGA50 font option as default
	fontname := "font-dos"
	fontlarge := "font-large"

	// use metadata to determine the font CSS class to use
	year, _, _ := filerecord.Dates(art)
	switch {
	case data["platform"] == textamiga:
		fontname = "font-amiga"
	case year != 0 && year <= CgaEpoch:
		data["contentBinarySwappers"] = false
		fontname = "font-dos-cga"
		fontlarge = ""
	}

	// ansi encoded texts
	if data["magic"] != "Binary data or binary text" {
		class := append([]string{fontname, fontlarge, "render"}, elems...)
		data["preElementClass"] = strings.Join(class, " ")
		return data
	}

	// everything else is treated as binary text
	class := elems
	class = append(class, "text-bg-dark", "text-center")
	if year > CgaEpoch {
		// use high-res squared font
		fontname = "font-squared"
		class = append([]string{fontname}, class...)
		i, j := 1, 3
		class = slices.Replace(class, i, j, "reader-hires")
	} else {
		class = append([]string{fontname, "reader"}, class...)
	}
	data["preElementClass"] = strings.Join(class, " ")
	data["contentBinarySwappers"] = false

	return data
}

// addTextUTF8 adds the appropriate buffer for the "Web style" text.
// The use of the buf Buffer is for legacy "ISO-8859-1" encoded text that is backwards compatible with UTF-8.
// The use of the ruf Buffer is for multi-byte, Unicode runes.
func (ds *Dirs) addTextUTF8(textBuf, runeBuf *bytes.Buffer, data map[string]any) map[string]any {
	if nils.Slog("dirs add text utf8", textBuf, runeBuf) {
		return data
	}

	data["contentUTF8"] = ""
	if runeBuf.Len() > 0 {
		data["contentUTF8"] = runeBuf.String()
		return data
	}

	data["contentUTF8"] = textBuf.String()

	return data
}

// addZip returns the archive content for the file download of the artifact.
// NOTE: this can cause a performance hit for archives with 10,000+ items.
func (ds *Dirs) addZip(content string, data map[string]any) map[string]any {
	if content == "" {
		return data
	}

	items := SortContent(content)
	items = slices.DeleteFunc(items, func(s string) bool {
		return strings.TrimSpace(s) == ""
	})

	count := len(items)
	// Cap the number of items to avoid generating HTML output that's too large
	maxItems := ds.Maximum.Items
	if count > maxItems {
		items = items[:maxItems]
	}

	paths := slices.Compact(items)
	data["content"] = paths  // This is displayed as "#	Filename or path"
	data["contentDesc"] = "" // This is used by "Download info"

	switch len(paths) {
	case 0:
		return data
	case 1:
		data["contentDesc"] = "contains one file"
	default:
		data["contentDesc"] = fmt.Sprintf("contains %d files", count)
	}

	return data
}

// encoding returns the encoding for the model file entry.
// Based on the platform and section.
// Otherwise it will attempt to determine the encoding from the file byte content.
func (ds *Dirs) encoding(r io.Reader) encoding.Encoding { //nolint:ireturn
	platform := strings.TrimSpace(ds.Platform)
	section := strings.TrimSpace(ds.Section)

	if strings.EqualFold(platform, textamiga) ||
		strings.EqualFold(section, "appleii") ||
		strings.EqualFold(section, "atarist") {
		return charmap.ISO8859_1
	}

	if strings.Contains(strings.ToLower(ds.Magic), "utf-8") {
		return unicode.UTF8
	}

	if r != nil {
		if guess := helper.Determine(r); guess != nil {
			return guess
		}
	}

	return charmap.CodePage437
}

// fixANSI uses the magicnumber data to match ansi texts and update the artifact platform classification.
func (ds *Dirs) fixANSI(ctx context.Context, sl *slog.Logger, db *sql.DB, data map[string]any) map[string]any {
	if nils.Slog("dirs fix ansi", ctx, sl, db) {
		return data
	}

	platform, ok := data["modOS"].(string)
	if !ok {
		return data
	}

	magic, ok := data["modMagicNumber"].(string)
	if !ok {
		return data
	}

	textfile := strings.EqualFold(platform, tags.Text.String())
	if !textfile || magic != magicnumber.ANSIEscapeText.Title() {
		return data
	}

	id := ds.ID
	val := tags.ANSI.String()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		ds.logErr(sl, "dirs fix ansi update begin tx", err)
	}
	if err := model.Platform.Update(ctx, tx, id, val); err != nil {
		ds.logErr(sl, "dirs fix ansi update platform", err)
	}
	if err = tx.Commit(); err != nil {
		ds.logErr(sl, "dirs fix ansi update tx commit", err)
	}

	data["platform"] = val

	return data
}

// fixMagic updates the magic number for the file record of the artifact.
// It must be called after both the dir.filemetadata and dir.Editor functions.
//
// Due to potential performance issues with extra large filedownloads, this update
// should only be used by logged-in editors.
//
// NOTE: this can be a performance issue on large files.
func (ds *Dirs) fixMagic(ctx context.Context, sl *slog.Logger, db *sql.DB, data map[string]any) map[string]any {
	const msg = "dirs fix magic"
	if nils.Slog(msg, ctx, sl, db) {
		return data
	}

	recMagic := data["magic"]
	modMagic := data["modMagicNumber"]
	if recMagic != modMagic {
		data["magic"] = modMagic
		magic, valid := modMagic.(string)
		if !valid {
			ds.logErr(sl, msg+" mod magic is a string", ErrType)
			return data
		}
		if err := model.UpdateMagic(ctx, db, ds.ID, magic); err != nil {
			ds.logErr(sl, msg, err)
		}
	}

	if findRepack, ok := data["extraZip"].(bool); !ok {
		ds.logErr(sl, msg+" extra zip is a boolean", ErrType)
		return data
	} else if findRepack {
		return data
	}

	root, ok := data["modDecompressLoc"].(string)
	if !ok {
		ds.logErr(sl, msg+" mod decompress loc is a string", ErrType)
		return data
	}

	if st, err := os.Stat(root); err != nil || !st.IsDir() {
		ds.logErr(sl, msg, err)
		return data
	}

	return ds.makeAssets(ctx, sl, root, modMagic, data)
}

// makeAssets will create repackaged archives and images of textfiles if they're required.
func (ds *Dirs) makeAssets(ctx context.Context, sl *slog.Logger, root string, modMagic any, data map[string]any,
) map[string]any {
	const msg = "dirs make assets"
	if nils.Slog(msg, ctx, sl) {
		return data
	}

	name := filepath.Join(ds.Download.Path(), ds.UUID)
	zipArchiving := modMagic == magicnumber.PKWAREZip.Title()
	switch {
	case legacyArchiving(modMagic):
		// requires repacking
	case zipArchiving:
		if !requireReplacementZip(name) {
			// does not require repacking
			return data
		}
	case plainText(modMagic):
		return ds.makeTextImages(ctx, sl, data)
	default:
		return data
	}

	i, err := ds.makeZipfile(root)
	if err != nil {
		ds.logErr(sl, msg, err)
		return data
	}
	slog.Info(msg+" success", slog.String("uuid", ds.UUID), slog.Int64("bytes extracted", i))
	data["extraZip"] = true

	return data
}

func (ds *Dirs) makeTextImages(ctx context.Context, sl *slog.Logger, data map[string]any) map[string]any {
	const msg = "dirs make text images"
	if nils.Slog(msg, ctx, sl) {
		return data
	}

	platform := ds.Platform
	name := filepath.Join(ds.Download.Path(), ds.UUID)
	dirs := command.Dirs{
		Download:  ds.Download,
		Preview:   ds.Preview,
		Thumbnail: ds.Thumbnail,
		Extra:     ds.Extra,
	}
	if helper.File(filepath.Join(dirs.Thumbnail.Path(), ds.UUID+".png")) ||
		helper.File(filepath.Join(dirs.Thumbnail.Path(), ds.UUID+".webp")) {
		return data
	}

	amigaFont := strings.EqualFold(platform, tags.TextAmiga.String()) ||
		strings.EqualFold(platform, tags.Console.String())
	if err := dirs.TextImager(ctx, sl, name, ds.UUID, amigaFont); err != nil {
		ds.logErr(sl, msg, err)
	}

	data["missingAssets"] = ""

	return data
}

func (ds *Dirs) makeZipfile(rootpath string) (size int64, err error) { //nolint:nonamedreturns
	const format = "dirs make zipfile: %w"

	path, err := dir.MkdirTemp("makezip")
	if err != nil {
		return 0, fmt.Errorf(format, err)
	}
	base := ds.UUID + ".zip"
	src := filepath.Join(path, base)

	dest := filepath.Join(ds.Extra.Path(), base)
	_ = os.Remove(dest)

	_, err = rezip.CompressDir(rootpath, src)
	if err != nil {
		return 0, fmt.Errorf(format, err)
	}

	if rErr := helper.RenameCrossDevice(src, dest); rErr != nil {
		err = errors.Join(err, fmt.Errorf(format, rErr))
		defer func() {
			err = errors.Join(err, dir.RemoveAll(src))
		}()
		return 0, err
	}

	st, err := os.Stat(dest)
	if err != nil {
		return 0, fmt.Errorf(format, err)
	}

	return st.Size(), nil
}

// Previews returns a map of preview assets for the file record of the artifact.
// Up to four preview assets are returned, JPEG, PNG, WebP and AVIF.
func (ds *Dirs) pathsPreview(sl *slog.Logger) map[string][2]string {
	if sl == nil {
		sl = logs.Discard()
	}

	unid := strings.ToLower(ds.UUID)

	avif := filepath.Join(ds.Preview.Path(), unid+".avif")
	jpg := filepath.Join(ds.Preview.Path(), unid+".jpg")
	png := filepath.Join(ds.Preview.Path(), unid+".png")
	webp := filepath.Join(ds.Preview.Path(), unid+".webp")

	const size = 4
	matches := make(map[string][2]string, size)

	if st, err := os.Stat(avif); err == nil {
		matches["AVIF"] = [2]string{humanize.Comma(st.Size()), ""}
	}
	matches["JPG"] = simple.ImageXY(sl, jpg)
	matches["PNG"] = simple.ImageXY(sl, png)
	matches["WebP"] = simple.ImageXY(sl, webp)

	return matches
}

// Thumbnails returns a map of thumbnail assets for the file record of the artifact.
// Two thumbnail assets are returned, PNG and WebP.
func (ds *Dirs) pathsThumb(sl *slog.Logger) map[string][2]string {
	if sl == nil {
		sl = logs.Discard()
	}

	unid := strings.ToLower(ds.UUID)

	png := filepath.Join(ds.Thumbnail.Path(), unid+".png")
	webp := filepath.Join(ds.Thumbnail.Path(), unid+".webp")

	const size = 2
	matches := make(map[string][2]string, size)

	matches["PNG"] = simple.ImageXY(sl, png)
	matches["WebP"] = simple.ImageXY(sl, webp)

	return matches
}

// Extras returns a map of extra assets for the file record of the artifact.
// Up to three extra assets are returned, FILE_ID, README and Repacked ZIP.
func (ds *Dirs) pathsExtras() map[string][2]string {
	unid := strings.ToLower(ds.UUID)

	const size = 3
	matches := make(map[string][2]string, size)

	diz := filepath.Join(ds.Extra.Path(), unid+".diz")
	if st, err := os.Stat(diz); err == nil {
		i, _ := helper.Lines(diz)
		matches["FILE_ID"] = [2]string{humanize.Comma(st.Size()), fmt.Sprintf("%d lines", i)}
	}

	hlp := filepath.Join(ds.Extra.Path(), unid+".hlp")
	if st, err := os.Stat(hlp); err == nil {
		i, _ := helper.Lines(hlp)
		matches["HELPER"] = [2]string{humanize.Comma(st.Size()), fmt.Sprintf("%d lines", i)}
	}

	txt := filepath.Join(ds.Extra.Path(), unid+".txt")
	if st, err := os.Stat(txt); err == nil {
		i, _ := helper.Lines(txt)
		matches["README"] = [2]string{humanize.Comma(st.Size()), fmt.Sprintf("%d lines", i)}
	}

	zip := filepath.Join(ds.Extra.Path(), unid+".zip")
	if st, err := os.Stat(zip); err == nil {
		matches["Repacked ZIP"] = [2]string{humanize.Comma(st.Size()), "Deflate compression"}
	}

	return matches
}

// pathsSuggest returns missing assets suggestions for the artifact.
//
// example:
//
//	"create a preview image + create a thumbnail image"
func (ds *Dirs) pathsSuggest() string {
	uid := ds.UUID

	download := helper.File(filepath.Join(ds.Download.Path(), uid))

	preview := helper.File(filepath.Join(ds.Preview.Path(), uid+".png")) ||
		helper.File(filepath.Join(ds.Preview.Path(), uid+".webp")) ||
		helper.File(filepath.Join(ds.Preview.Path(), uid+".jpg"))

	thumbnail := helper.File(filepath.Join(ds.Thumbnail.Path(), uid+".png")) ||
		helper.File(filepath.Join(ds.Thumbnail.Path(), uid+".webp"))

	if download && preview && thumbnail {
		return ""
	}

	suggests := []string{}

	if !download {
		suggests = append(suggests, "offer a file for download")
	}

	platform := ds.Platform
	if platform == tags.Audio.String() {
		return strings.Join(suggests, " + ")
	}

	textfiles := platform == tags.Text.String() || platform == tags.TextAmiga.String()
	if !preview && !textfiles {
		suggests = append(suggests, "create a preview image")
	}

	if !thumbnail {
		suggests = append(suggests, "create a thumbnail image")
	}

	return strings.Join(suggests, " + ")
}

// oneByKey retrieves a single artifact record using [dir.URI] as the key.
func (ds *Dirs) oneByKey(ctx context.Context, sl *slog.Logger, c *echo.Context, db *sql.DB) (*models.File, error) {
	if err := nils.Check(ctx, sl, c, db); err != nil {
		return nil, fmt.Errorf("dirs one by key: %w", err)
	}

	var art *models.File
	var err error
	if sess.Editor(c) {
		art, err = model.OneEditByKey(ctx, db, ds.URI)
	} else {
		art, err = model.OneFileByKey(ctx, db, ds.URI)
	}
	if err != nil {
		if errors.Is(err, model.ErrBadID) {
			return nil, artifact404(sl, c, ds.URI)
		}
		return nil, DatabaseErr(sl, c, "f/"+ds.URI, err)
	}

	return art, nil
}

// screenshot returns true if a suitable screenshot is found on the host server.
//
// If the platform is "text" or "textamiga", false is always returned.
func (ds *Dirs) screenshot() bool {
	if ds.UUID == "" {
		return false
	}

	switch strings.ToLower(strings.TrimSpace(ds.Platform)) {
	case textamiga, text:
		return false
	}

	// check common image extensions efficiently without intermediate slice allocations
	exts := [...]string{".webp", ".png", ".jpg"}
	for _, ext := range exts {
		p := filepath.Join(ds.Preview.Path(), ds.UUID+ext)
		if helper.Stat(p) {
			return true // found a valid screenshot
		}
	}

	return false
}

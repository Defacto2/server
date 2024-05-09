package app

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"image"
	_ "image/gif"  // gif format decoder
	_ "image/jpeg" // jpeg format decoder
	_ "image/png"  // png format decoder
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/handler/sess"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/magic"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/render"
	"github.com/Defacto2/server/model"
	"github.com/dustin/go-humanize"
	"github.com/h2non/filetype"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	_ "golang.org/x/image/webp" // webp format decoder
	"golang.org/x/text/encoding/charmap"
)

// Dirs contains the directories used by the artifact pages.
type Dirs struct {
	Download  string // path to the artifact download directory
	Preview   string // path to the preview and screenshot directory
	Thumbnail string // path to the file thumbnail directory
	URI       string // the URI of the file record
}

type extract int // extract target format for the file archive extractor

const (
	imgs  extract = iota // extract image
	ansis                // extract ansilove compatible text
)

const (
	epoch = model.EpochYear // epoch is the default year for MS-DOS files without a timestamp
)

// Artifact404 renders the error page for the artifact links.
func Artifact404(c echo.Context, id string) error {
	const name = "status"
	if c == nil {
		return InternalErr(c, name, ErrCxt)
	}
	data := empty(c)
	data["title"] = fmt.Sprintf("%d error, artifact page not found", http.StatusNotFound)
	data["description"] = fmt.Sprintf("HTTP status %d error", http.StatusNotFound)
	data["code"] = http.StatusNotFound
	data["logo"] = "Artifact not found"
	data["alert"] = fmt.Sprintf("Artifact %q cannot be found", strings.ToLower(id))
	data["probl"] = "The artifact page does not exist, there is probably a typo with the URL."
	data["uriOkay"] = "f/"
	data["uriErr"] = id
	err := c.Render(http.StatusNotFound, name, data)
	if err != nil {
		return InternalErr(c, name, err)
	}
	return nil
}

// Artifact is the handler for the of the file record.
func (dir Dirs) Artifact(c echo.Context, logger *zap.SugaredLogger, readonly bool) error { //nolint:funlen
	const name = "artifact"
	if logger == nil {
		return InternalErr(c, name, ErrZap)
	}
	var art *models.File
	var err error
	if sess.Editor(c) {
		art, err = model.EditObf(dir.URI)
	} else {
		art, err = model.FindObf(dir.URI)
	}
	if err != nil {
		if errors.Is(err, model.ErrID) {
			return Artifact404(c, dir.URI)
		}
		return DatabaseErr(c, "f/"+dir.URI, err)
	}
	fname := art.Filename.String
	uuid := art.UUID.String
	data := empty(c)
	data = dir.artifactEditor(art, data, readonly)
	// page metadata
	data["uuid"] = uuid
	data["download"] = helper.ObfuscateID(art.ID)
	data["title"] = fname
	data["description"] = artifactDesc(art)
	data["h1"] = artifactIssue(art)
	data["lead"] = artifactLead(art)
	data["comment"] = art.Comment.String
	// file metadata
	data["filename"] = fname
	data["filesize"] = artifactByteCount(art.Filesize)
	data["filebyte"] = art.Filesize
	data["lastmodified"] = artifactLM(art)
	data["lastmodifiedAgo"] = artifactModAgo(art)
	data["checksum"] = strings.TrimSpace(art.FileIntegrityStrong.String)
	data["magic"] = art.FileMagicType.String
	data["releasers"] = string(LinkRels(art.GroupBrandBy, art.GroupBrandFor))
	data["published"] = model.PublishedFmt(art)
	data["section"] = strings.TrimSpace(art.Section.String)
	data["platform"] = strings.TrimSpace(art.Platform.String)
	data["alertURL"] = art.FileSecurityAlertURL.String
	// attributions and credits
	data["writers"] = art.CreditText.String
	data["artists"] = art.CreditIllustration.String
	data["programmers"] = art.CreditProgram.String
	data["musicians"] = art.CreditAudio.String
	// links to other records and sites
	data["relations"] = artifactRelations(art)
	data["websites"] = artifactWebsites(art)
	data["demozoo"] = artifactID(art.WebIDDemozoo.Int64)
	data["pouet"] = artifactID(art.WebIDPouet.Int64)
	data["sixteenColors"] = art.WebID16colors.String
	data["youtube"] = strings.TrimSpace(art.WebIDYoutube.String)
	data["github"] = art.WebIDGithub.String
	// js-dos emulator
	data = jsdos(logger, art, data, fname)
	// archive file content
	data = content(art, data)
	// record metadata
	data["linkpreview"] = LinkPreviewHref(art.ID, art.Filename.String, art.Platform.String)
	data["linkpreviewTip"] = LinkPreviewTip(art.Filename.String, art.Platform.String)
	data = filentry(art, data)
	d, err := dir.artifactReadme(art)
	if err != nil {
		return InternalErr(c, name, err)
	}
	maps.Copy(data, d)
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, name, err)
	}
	return nil
}

func (dir Dirs) artifactEditor(art *models.File, data map[string]interface{}, readonly bool) map[string]interface{} {
	if readonly || art == nil {
		return data
	}
	uuid := art.UUID.String
	abs := filepath.Join(dir.Download, uuid)
	data["readOnly"] = false
	data["modID"] = art.ID
	data["modTitle"] = art.RecordTitle.String
	data["modOnline"] = art.Deletedat.Time.IsZero()
	data["modReleasers"] = RecordRels(art.GroupBrandBy, art.GroupBrandFor)
	rr := RecordReleasers(art.GroupBrandFor, art.GroupBrandBy)
	data["modReleaser1"] = rr[0]
	data["modReleaser2"] = rr[1]
	data["modYear"] = art.DateIssuedYear.Int16
	data["modMonth"] = art.DateIssuedMonth.Int16
	data["modDay"] = art.DateIssuedDay.Int16
	data["modLastMod"] = !art.FileLastModified.IsZero()
	data["modLMYear"] = art.FileLastModified.Time.Year()
	data["modLMMonth"] = int(art.FileLastModified.Time.Month())
	data["modLMDay"] = art.FileLastModified.Time.Day()
	data["modAbsDownload"] = abs
	data["modKind"] = artifactMagic(abs)
	data["modStatModify"] = artifactStat(abs)[0]
	data["modStatSize"] = artifactStat(abs)[1]
	data["modAssets"] = dir.artifactAssets(uuid)
	data["modNoReadme"] = art.RetrotxtNoReadme.Int16 != 0
	data["modReadmeList"] = OptionsReadme(art.FileZipContent.String)
	data["modPreviewList"] = OptionsPreview(art.FileZipContent.String)
	data["modAnsiLoveList"] = OptionsAnsiLove(art.FileZipContent.String)
	data["modReadmeSuggest"] = readmeSuggest(art)
	data["modZipContent"] = strings.TrimSpace(art.FileZipContent.String)
	data["modRelations"] = art.ListRelations.String
	data["modWebsites"] = art.ListLinks.String
	data["modOS"] = strings.ToLower(strings.TrimSpace(art.Platform.String))
	data["modTag"] = strings.ToLower(strings.TrimSpace(art.Section.String))
	data["virusTotal"] = strings.TrimSpace(art.FileSecurityAlertURL.String)
	data["forApproval"] = !art.Deletedat.IsZero() && art.Deletedby.IsZero()
	data["disableRecord"] = !art.Deletedat.IsZero() && !art.Deletedby.IsZero()
	return data
}

func content(art *models.File, data map[string]interface{}) map[string]interface{} {
	if art == nil {
		return data
	}
	items := strings.Split(art.FileZipContent.String, "\n")
	items = slices.DeleteFunc(items, func(s string) bool {
		return strings.TrimSpace(s) == ""
	})
	paths := slices.Compact(items)
	data["content"] = paths
	data["contentDesc"] = ""

	l := len(paths)
	switch l {
	case 0:
		return data
	case 1:
		data["contentDesc"] = "contains one file"
	default:
		data["contentDesc"] = fmt.Sprintf("contains %d files", l)
	}
	return data
}

func jsdos(logger *zap.SugaredLogger,
	art *models.File,
	data map[string]interface{},
	fname string,
) map[string]interface{} {
	if logger == nil || art == nil {
		return data
	}
	data["jsdos6"] = false
	data["jsdos6Run"] = ""
	data["jsdos6Config"] = ""
	data["jsdos6Zip"] = false
	if emulate := artifactJSDos(art); emulate {
		data["jsdos6"] = emulate
		run, err := model.JsDosBinary(art)
		if err != nil {
			logger.Error(err)
			return data
		}
		data["jsdos6Run"] = run
		cfg, err := model.JsDosConfig(art)
		if err != nil {
			logger.Error(err)
			return data
		}
		data["jsdos6Config"] = cfg
		data["jsdos6Zip"] = filepath.Ext(strings.ToLower(fname)) == ".zip"
	}
	return data
}

func filentry(art *models.File, data map[string]interface{}) map[string]interface{} {
	if art == nil {
		return data
	}
	data["filentry"] = ""
	switch {
	case art.Createdat.Valid && art.Updatedat.Valid:
		c := Updated(art.Createdat.Time, "")
		u := Updated(art.Updatedat.Time, "")
		if c != u {
			c = Updated(art.Createdat.Time, "Created")
			u = Updated(art.Updatedat.Time, "Updated")
			data["filentry"] = c + br + u
			return data
		}
		c = Updated(art.Createdat.Time, "Created")
		data["filentry"] = c
	case art.Createdat.Valid:
		c := Updated(art.Createdat.Time, "Created")
		data["filentry"] = c
	case art.Updatedat.Valid:
		u := Updated(art.Updatedat.Time, "Updated")
		data["filentry"] = u
	}
	return data
}

// AnsiLovePost handles the post submission for the Preview from text in archive.
func (dir Dirs) AnsiLovePost(c echo.Context, logger *zap.SugaredLogger) error {
	return dir.extractor(c, logger, ansis)
}

// PreviewDel handles the post submission for the Delete complementary images button.
func (dir Dirs) PreviewDel(c echo.Context) error {
	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	r, err := model.EditFind(f.ID)
	if err != nil {
		return badRequest(c, err)
	}
	if err = command.RemoveImgs(r.UUID.String, dir.Preview, dir.Thumbnail); err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

// PreviewPost handles the post submission for the Preview from image in archive.
func (dir Dirs) PreviewPost(c echo.Context, logger *zap.SugaredLogger) error {
	return dir.extractor(c, logger, imgs)
}

// artifactReadme returns the readme data for the file record.
func (dir Dirs) artifactReadme(art *models.File) (map[string]interface{}, error) { //nolint:funlen
	data := map[string]interface{}{}
	if art == nil {
		return data, nil
	}
	if art.RetrotxtNoReadme.Int16 != 0 {
		return data, nil
	}
	platform := strings.TrimSpace(art.Platform.String)
	switch platform {
	case "markup", "pdf":
		return data, nil
	}
	if render.NoScreenshot(art, dir.Preview) {
		data["noScreenshot"] = true
	}
	// the bbs era, remote images protcol is not supported
	// example: /f/b02392f
	const unsupported = ".rip"
	if filepath.Ext(strings.ToLower(art.Filename.String)) == unsupported {
		return data, nil
	}
	b, err := render.Read(art, dir.Download)
	r := bufio.NewReader(bytes.NewReader(b))
	switch {
	case errors.Is(err, render.ErrDownload):
		data["noDownload"] = true
		return data, nil
	case errors.Is(err, render.ErrFilename), r == nil:
		return data, nil
	case err != nil:
		return data, fmt.Errorf("render.Read: %w", err)
	case b == nil, render.IsUTF16(r), isZip(b):
		return data, nil
	}
	// Remove control codes and metadata from byte array
	const (
		reAnsi    = `\x1b\[[0-9;]*[a-zA-Z]` // ANSI escape codes
		reAmiga   = `\x1b\[[0-9;]*[ ]p`     // unknown control code found in Amiga texts
		reSauce   = `SAUCE00.*`             // SAUCE metadata that is appended to some files
		nlWindows = "\r\n"                  // Windows line endings
		nlUnix    = "\n"                    // Unix line endings
	)
	re := regexp.MustCompile(reAnsi + `|` + reAmiga + `|` + reSauce)
	b = re.ReplaceAll(b, []byte{})
	b = bytes.ReplaceAll(b, []byte(nlWindows), []byte(nlUnix))
	if len(b) == 0 {
		return data, nil
	}
	const (
		sp      = 0x20 // space
		hyphen  = 0x2d // hyphen-minus
		shy     = 0xad // soft hyphen for ISO8859-1
		nbsp    = 0xa0 // non-breaking space for ISO8859-1
		nbsp437 = 0xff // non-breaking space for CP437
		space   = " "  // intentional space
	)
	switch render.Encoder(art, bytes.NewReader(b)) {
	case charmap.ISO8859_1:
		data["readmeLatin1Cls"] = ""
		data["readmeCP437Cls"] = "d-none" + space
		data["topazCheck"] = "checked"
		b = bytes.ReplaceAll(b, []byte{nbsp}, []byte{sp})
		b = bytes.ReplaceAll(b, []byte{shy}, []byte{hyphen})
	case charmap.CodePage437:
		data["readmeLatin1Cls"] = "d-none" + space
		data["readmeCP437Cls"] = ""
		data["vgaCheck"] = "checked"
		b = bytes.ReplaceAll(b, []byte{nbsp437}, []byte{sp})
	}
	d := charmap.ISO8859_1.NewDecoder().Reader(bytes.NewReader(b))
	readme, err := decode(d)
	if err != nil {
		return data, fmt.Errorf("decode: %w", err)
	}
	data["readmeLatin1"] = readme
	d = charmap.CodePage437.NewDecoder().Reader(bytes.NewReader(b))
	readme, err = decode(d)
	if err != nil {
		return data, fmt.Errorf("decode: %w", err)
	}
	data["readmeCP437"] = readme
	data["readmeLines"] = strings.Count(readme, "\n")
	data["readmeRows"] = helper.MaxLineLength(readme)
	return data, nil
}

// isZip checks if b is a known zip archive.
// when b is unknown, "application/octet-stream" is returned
// which can be a false positive with other legacy text files.
func isZip(b []byte) bool {
	switch http.DetectContentType(b) {
	case "archive/zip", "application/zip":
		return true
	}
	return false
}

func decode(r io.Reader) (string, error) {
	out := strings.Builder{}
	if _, err := io.Copy(&out, r); err != nil {
		return "", fmt.Errorf("io.Copy: %w", err)
	}
	if !strings.HasSuffix(out.String(), "\n\n") {
		out.WriteString("\n")
	}
	return out.String(), nil
}

// extractor is a helper function for the PreviewPost and AnsiLovePost handlers.
func (dir Dirs) extractor(c echo.Context, logger *zap.SugaredLogger, p extract) error {
	var f Form
	if err := c.Bind(&f); err != nil {
		return badRequest(c, err)
	}
	r, err := model.EditFind(f.ID)
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
		err = cmd.ExtractImage(logger, src, ext, r.UUID.String, target)
	case ansis:
		err = cmd.ExtractAnsiLove(logger, src, ext, r.UUID.String, target)
	default:
		return InternalErr(c, "extractor", fmt.Errorf("%w: %d", ErrExtract, p))
	}
	if err != nil {
		return badRequest(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

// artifactByteCount returns the file size for the file record.
func artifactByteCount(i int64) string {
	if i == 0 {
		return "(n/a)"
	}
	return humanize.Bytes(uint64(i))
}

// artifactDesc returns the description for the file record.
func artifactDesc(art *models.File) string {
	s := art.Filename.String
	if art.RecordTitle.String != "" {
		s = artifactIssue(art)
	}
	r1 := releaser.Clean(strings.ToLower(art.GroupBrandBy.String))
	r2 := releaser.Clean(strings.ToLower(art.GroupBrandFor.String))
	r := ""
	switch {
	case r1 != "" && r2 != "":
		r = fmt.Sprintf("%s + %s", r1, r2)
	case r1 != "":
		r = r1
	case r2 != "":
		r = r2
	}
	s = fmt.Sprintf("%s released by %s", s, r)
	y := art.DateIssuedYear.Int16
	if y > 0 {
		s = fmt.Sprintf("%s in %d", s, y)
	}
	return s
}

// artifactIssue returns the title of the file,
// unless the file is a magazine issue, in which case it returns the issue number.
func artifactIssue(art *models.File) string {
	sect := strings.TrimSpace(strings.ToLower(art.Section.String))
	if sect != "magazine" {
		return art.RecordTitle.String
	}
	s := art.RecordTitle.String
	if i, err := strconv.Atoi(s); err == nil {
		return fmt.Sprintf("Issue %d", i)
	}
	return s
}

// artifactLead returns the lead for the file record which is the filename and releasers.
func artifactLead(art *models.File) string {
	fname := art.Filename.String
	span := fmt.Sprintf("<span class=\"font-monospace fs-6 fw-light\">%s</span> ", fname)
	rels := string(LinkRels(art.GroupBrandBy, art.GroupBrandFor))
	return fmt.Sprintf("%s<br>%s", rels, span)
}

// artifactLM returns the last modified date for the file record.
func artifactLM(art *models.File) string {
	const none = "no timestamp"
	if !art.FileLastModified.Valid {
		return none
	}
	year, _ := strconv.Atoi(art.FileLastModified.Time.Format("2006"))
	if year <= epoch {
		return none
	}
	lm := art.FileLastModified.Time.Format("2006 Jan 2, 15:04")
	if lm == "0001 Jan 1, 00:00" {
		return none
	}
	return lm
}

// artifactMagic returns the MIME type for the file record.
func artifactMagic(name string) string {
	file, err := os.Open(name)
	if err != nil {
		return err.Error()
	}
	defer file.Close()

	const sample = 512
	head := make([]byte, sample)
	_, err = file.Read(head)
	if err != nil {
		return err.Error()
	}
	filetype.AddMatcher(magic.ANSIType(), magic.ANSIMatcher)
	filetype.AddMatcher(magic.ArcSeaType(), magic.ArcSeaMatcher)
	filetype.AddMatcher(magic.ARJType(), magic.ARJMatcher)
	filetype.AddMatcher(magic.DOSComType(), magic.DOSComMatcher)
	filetype.AddMatcher(magic.InterchangeType(), magic.InterchangeMatcher)
	filetype.AddMatcher(magic.PCXType(), magic.PCXMatcher)
	kind, err := filetype.Match(head)
	if err != nil {
		return err.Error()
	}
	if kind != filetype.Unknown {
		return kind.MIME.Value
	}

	return http.DetectContentType(head)
}

// artifactModAgo returns the last modified date in a human readable format.
func artifactModAgo(art *models.File) string {
	const none = "No recorded timestamp"
	if !art.FileLastModified.Valid {
		return none
	}
	year, _ := strconv.Atoi(art.FileLastModified.Time.Format("2006"))
	if year <= epoch {
		return none
	}
	return Updated(art.FileLastModified.Time, "Modified")
}

// artifactRelations returns the list of relationships for the file record.
func artifactRelations(art *models.File) template.HTML {
	s := art.ListRelations.String
	if s == "" {
		return ""
	}
	links := strings.Split(s, "|")
	if len(links) == 0 {
		return ""
	}
	rows := ""
	const expected = 2
	for _, link := range links {
		x := strings.Split(link, ";")
		if len(x) != expected {
			continue
		}
		name, href := x[0], x[1]
		rows += fmt.Sprintf("<tr><th scope=\"row\"><small>Link to</small></th>"+
			"<td><small><a class=\"text-truncate\" href=\"%s\">%s</a></small></td></tr>", href, name)
	}
	return template.HTML(rows)
}

// artifactWebsites returns the list of links for the file record.
func artifactWebsites(art *models.File) template.HTML {
	s := art.ListLinks.String
	if s == "" {
		return ""
	}
	links := strings.Split(s, "|")
	if len(links) == 0 {
		return ""
	}
	rows := ""
	const expected = 2
	for _, link := range links {
		x := strings.Split(link, ";")
		if len(x) != expected {
			continue
		}
		name, href := x[0], x[1]
		rows += fmt.Sprintf("<tr><th scope=\"row\"><small>Link to</small></th>"+
			"<td><small><a class=\"link-offset-3 icon-link icon-link-hover\" href=\"https://%s\">%s %s</a></small></td></tr>", href, name, LinkSVG())
	}
	return template.HTML(rows)
}

// artifactJSDos returns true if the file record is a known, MS-DOS executable.
// The supported file types are .zip archives and .exe, .com. binaries.
// Script files such as .bat and .cmd are not supported.
func artifactJSDos(art *models.File) bool {
	if strings.TrimSpace(strings.ToLower(art.Platform.String)) != "dos" {
		return false
	}
	ext := filepath.Ext(strings.ToLower(art.Filename.String))
	switch ext {
	case ".zip":
		return true
	case ".exe", ".com":
		return true
	case ".bat", ".cmd":
		return false
	default:
		return false
	}
}

// artifactID returns the record ID as a string.
func artifactID(id int64) string {
	if id == 0 {
		return ""
	}
	return strconv.FormatInt(id, 10)
}

// artifactStat returns the file last modified date and formatted file size.
func artifactStat(name string) [2]string {
	stat, err := os.Stat(name)
	if err != nil {
		return [2]string{err.Error(), err.Error()}
	}
	return [2]string{
		stat.ModTime().Format("2006-Jan-02"),
		fmt.Sprintf("%s bytes - %s - %s",
			humanize.Comma(stat.Size()),
			humanize.Bytes(uint64(stat.Size())),
			humanize.IBytes(uint64(stat.Size()))),
	}
}

// artifactAssets returns a list of downloads and image assets belonging to the file record.
// any errors are appended to the list.
func (dir Dirs) artifactAssets(uuid string) map[string]string {
	matches := map[string]string{}

	downloads, err := os.ReadDir(dir.Download)
	if err != nil {
		matches[err.Error()] = ""
	}
	images, err := os.ReadDir(dir.Preview)
	if err != nil {
		matches[err.Error()] = ""
	}
	thumbs, err := os.ReadDir(dir.Thumbnail)
	if err != nil {
		matches[err.Error()] = ""
	}

	for _, file := range downloads {
		if strings.HasPrefix(file.Name(), uuid) {
			if filepath.Ext(file.Name()) == "" {
				continue
			}
			s := strings.ToUpper(filepath.Ext(file.Name()))
			st, err := file.Info()
			if err != nil {
				matches[err.Error()] = err.Error()
			}
			switch s {
			case ".TXT":
				s = ".TXT readme"
				i, _ := helper.Lines(filepath.Join(dir.Download, file.Name()))
				matches[s] = fmt.Sprintf("%s bytes - %d lines", humanize.Comma(st.Size()), i)
			case ".ZIP":
				s = ".ZIP for emulator"
				matches[s] = humanize.Comma(st.Size()) + " bytes"
			}
		}
	}
	for _, file := range images {
		if strings.HasPrefix(file.Name(), uuid) {
			s := strings.ToUpper(filepath.Ext(file.Name()))
			if s == ".WEBP" {
				s = ".WebP"
			}
			matches[s+" preview "] = artifactImgInfo(filepath.Join(dir.Preview, file.Name()))
		}
	}
	for _, file := range thumbs {
		if strings.HasPrefix(file.Name(), uuid) {
			s := strings.ToUpper(filepath.Ext(file.Name()))
			if s == ".WEBP" {
				s = ".WebP"
			}
			matches[s+" thumb"] = artifactImgInfo(filepath.Join(dir.Thumbnail, file.Name()))
		}
	}

	return matches
}

// artifactImgInfo returns the image file size and dimensions.
func artifactImgInfo(name string) string {
	switch filepath.Ext(strings.ToLower(name)) {
	case ".jpg", ".jpeg", ".gif", ".png", ".webp":
	default:
		st, err := os.Stat(name)
		if err != nil {
			return err.Error()
		}
		return humanize.Comma(st.Size()) + " bytes"
	}
	reader, err := os.Open(name)
	if err != nil {
		return err.Error()
	}
	defer reader.Close()
	st, err := reader.Stat()
	if err != nil {
		return err.Error()
	}
	config, _, err := image.DecodeConfig(reader)
	if err != nil {
		return err.Error()
	}
	return fmt.Sprintf("%s bytes - %d x %d pixels", humanize.Comma(st.Size()), config.Width, config.Height)
}

// readmeSuggest returns a suggested readme file name for the record.
func readmeSuggest(r *models.File) string {
	if r == nil {
		return ""
	}
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
	return ReadmeSuggest(filename, group, content...)
}

// readmeFinds returns a list of readme text files found in the file archive.
func readmeFinds(content ...string) []string {
	finds := []string{}
	skip := []string{"scene.org", "scene.org.txt"}
	for _, name := range content {
		if name == "" {
			continue
		}
		s := strings.ToLower(name)
		if slices.Contains(skip, s) {
			continue
		}
		ext := filepath.Ext(s)
		if slices.Contains(priority(), ext) {
			finds = append(finds, name)
			continue
		}
		if slices.Contains(candidate(), ext) {
			finds = append(finds, name)
		}
	}
	return finds
}

// priority returns a list of readme text file extensions in priority order.
func priority() []string {
	return []string{".nfo", ".txt", ".unp", ".doc"}
}

// candidate returns a list of other, common text file extensions in priority order.
func candidate() []string {
	return []string{".diz", ".asc", ".1st", ".dox", ".me", ".cap", ".ans", ".pcb"}
}

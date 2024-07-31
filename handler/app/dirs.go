package app

// Package file dirs.go contains the artifact page directories and handlers.

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	_ "image/gif"  // gif format decoder
	_ "image/jpeg" // jpeg format decoder
	_ "image/png"  // png format decoder
	"io"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Defacto2/server/handler/app/internal/mf"
	"github.com/Defacto2/server/handler/app/internal/readme"
	"github.com/Defacto2/server/handler/app/internal/str"
	"github.com/Defacto2/server/handler/sess"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/render"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/dustin/go-humanize"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	_ "golang.org/x/image/webp" // webp format decoder
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

const epoch = model.EpochYear // epoch is the default year for MS-DOS files without a timestamp

// Artifact404 renders the error page for the artifact links.
func Artifact404(c echo.Context, id string) error {
	const name = "status"
	if c == nil {
		return InternalErr(c, name, errorWithID(ErrCxt, id, nil))
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
		return InternalErr(c, name, errorWithID(err, id, nil))
	}
	return nil
}

// Dirs contains the directories used by the artifact pages.
type Dirs struct {
	Download  string // path to the artifact download directory
	Preview   string // path to the preview and screenshot directory
	Thumbnail string // path to the file thumbnail directory
	Extra     string // path to the extra files directory
	URI       string // the URI of the file record
}

// Artifact is the handler for the of the file record.
func (dir Dirs) Artifact(c echo.Context, logger *zap.SugaredLogger, readonly bool) error {
	const name = "artifact"
	if logger == nil {
		return InternalErr(c, name, ErrZap)
	}
	art, err := dir.modelsFile(c)
	if err != nil {
		return err
	}
	data := empty(c)
	if !readonly {
		data = dir.Editor(art, data)
	}
	// page metadata
	data["unid"] = mf.UnID(art)
	data["download"] = mf.DownloadID(art)
	data["title"] = mf.Basename(art)
	data["description"] = mf.Description(art)
	data["h1"] = mf.FirstHeader(art)
	data["lead"] = firstLead(art)
	data["comment"] = mf.Comment(art)
	data = dir.filemetadata(art, data)
	data = dir.attributions(art, data)
	data = dir.otherRelations(art, data)
	data = jsdos(art, data, logger)
	data = content(art, data)
	data["linkpreview"] = mf.LinkPreview(art)
	data["linkpreviewTip"] = mf.LinkPreviewTip(art)
	data["filentry"] = mf.FileEntry(art)
	if skip := render.NoScreenshot(art, dir.Preview); skip {
		data["noScreenshot"] = true
	}
	if mf.EmbedReadme(art) {
		data, err = dir.embed(art, data)
		if err != nil {
			return InternalErr(c, name, errorWithID(err, dir.URI, art.ID))
		}
	}
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, name, errorWithID(err, dir.URI, art.ID))
	}
	return nil
}

func (dir Dirs) embed(art *models.File, data map[string]interface{}) (map[string]interface{}, error) {
	p, err := readme.Read(art, dir.Download, dir.Extra)
	if err != nil {
		if errors.Is(err, render.ErrDownload) {
			data["noDownload"] = true
			return data, nil
		}
		return nil, fmt.Errorf("dirs.embed read: %w", err)
	}
	d, err := embedText(art, data, p...)
	if err != nil {
		return nil, fmt.Errorf("dirs.embed text: %w", err)
	}
	maps.Copy(data, d)
	return d, nil
}

// Editor returns the editor data for the file record of the artifact.
// These are the editable fields for the file record that are only visible to the editor
// after they have logged in.
func (dir Dirs) Editor(art *models.File, data map[string]interface{}) map[string]interface{} {
	if art == nil {
		return data
	}
	unid := mf.UnID(art)
	abs := filepath.Join(dir.Download, unid)
	data["epochYear"] = epoch
	data["readonlymode"] = false
	data["modID"] = art.ID
	data["modTitle"] = mf.Title(art)
	data["modOnline"] = mf.RecordOnline(art)
	data["modReleasers"] = RecordRels(art.GroupBrandBy, art.GroupBrandFor)
	data["modReleaser1"], data["modReleaser2"] = mf.ReleaserPair(art)
	data["modYear"], data["modMonth"], data["modDay"] = mf.Dates(art)
	data["modLMYear"], data["modLMMonth"], data["modLMDay"] = mf.LastModifications(art)
	data["modAbsDownload"] = abs
	data["modMagicMime"] = str.MIME(abs)
	data["modMagicNumber"] = str.MagicAsTitle(abs)
	data["modDBModify"] = mf.LastModificationDate(art)
	data["modStatModify"], data["modStatSizeB"], data["modStatSizeF"] = str.StatHumanize(abs)
	data["modDecompress"] = mf.ListContent(art, abs)
	data["modDecompressLoc"] = str.MkContent(abs)
	data["modAssetPreview"] = dir.assets(dir.Preview, unid)
	data["modAssetThumbnail"] = dir.assets(dir.Thumbnail, unid)
	data["modAssetExtra"] = dir.assets(dir.Extra, unid)
	data["modNoReadme"] = mf.ReadmeNone(art)
	// data["modReadmeList"] = OptionsReadme(art.FileZipContent.String) // Check if this is needed
	// data["modPreviewList"] = OptionsPreview(art.FileZipContent.String)
	// data["modAnsiLoveList"] = OptionsAnsiLove(art.FileZipContent.String)
	data["modReadmeSuggest"] = mf.Readme(art)
	data["modZipContent"] = mf.ZipContent(art)
	data["modRelations"] = mf.RelationsStr(art)
	data["modWebsites"] = mf.WebsitesStr(art)
	data["modOS"] = mf.TagProgram(art)
	data["modTag"] = mf.TagCategory(art)
	data["virusTotal"] = mf.AlertURL(art) // FIXME, virusTotal is a dupe of ["alertURL"] ?
	data["forApproval"] = mf.RecordIsNew(art)
	data["disableApproval"] = mf.RecordProblems(art)
	data["disableRecord"] = mf.RecordOffline(art)
	data["missingAssets"] = dir.missingAssets(art)
	data["modEmulateXMS"], data["modEmulateEMS"], data["modEmulateUMB"] = mf.JsdosMemory(art)
	data["modEmulateBroken"] = mf.JsdosBroken(art)
	data["modEmulateRun"] = mf.JsdosRun(art)
	data["modEmulateCPU"] = mf.JsdosCPU(art)
	data["modEmulateMachine"] = mf.JsdosMachine(art)
	data["modEmulateAudio"] = mf.JsdosSound(art)
	return data
}

// modelsFile returns the URI artifact record from the file table.
func (dir Dirs) modelsFile(c echo.Context) (*models.File, error) {
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return nil, DatabaseErr(c, "f/"+dir.URI, err)
	}
	defer db.Close()
	var art *models.File
	if sess.Editor(c) {
		art, err = model.OneEditByKey(ctx, db, dir.URI)
	} else {
		art, err = model.OneFileByKey(ctx, db, dir.URI)
	}
	if err != nil {
		if errors.Is(err, model.ErrID) {
			return nil, Artifact404(c, dir.URI)
		}
		return nil, DatabaseErr(c, "f/"+dir.URI, err)
	}
	return art, nil
}

// Assets returns a list of downloads and images belonging to the file record.
// Any errors are appended to the list.
// The returned map contains a short description of the asset, the file size and extra information,
// such as image dimensions or the number of lines in a text file.
func (dir Dirs) assets(nameDir, unid string) map[string][2]string {
	matches := map[string][2]string{}
	files, err := os.ReadDir(nameDir)
	if err != nil {
		matches["error"] = [2]string{err.Error(), ""}
	}
	// Provide a string path and use that instead of dir Dirs.
	const assetDownload = ""
	for _, file := range files {
		if strings.HasPrefix(file.Name(), unid) {
			if filepath.Ext(file.Name()) == assetDownload {
				continue
			}
			ext := strings.ToUpper(filepath.Ext(file.Name()))
			st, err := file.Info()
			if err != nil {
				matches["error"] = [2]string{err.Error(), ""}
			}
			switch ext {
			case ".AVIF":
				s := "AVIF"
				matches[s] = [2]string{humanize.Comma(st.Size()), ""}
			case ".JPG":
				s := "Jpeg"
				matches[s] = str.ImageXY(filepath.Join(nameDir, file.Name()))
			case ".PNG":
				s := "PNG"
				matches[s] = str.ImageXY(filepath.Join(nameDir, file.Name()))
			case ".TXT":
				s := "README"
				i, _ := helper.Lines(filepath.Join(dir.Extra, file.Name()))
				matches[s] = [2]string{humanize.Comma(st.Size()), fmt.Sprintf("%d lines", i)}
			case ".WEBP":
				s := "WebP"
				matches[s] = str.ImageXY(filepath.Join(nameDir, file.Name()))
			case ".ZIP":
				s := "Repacked ZIP"
				matches[s] = [2]string{humanize.Comma(st.Size()), "Deflate compression"}
			}
		}
	}
	return matches
}

// missingAssets returns a string of missing assets for the file record of the artifact.
func (dir Dirs) missingAssets(art *models.File) string {
	uid := art.UUID.String
	missing := []string{}
	d := helper.File(filepath.Join(dir.Download, uid))
	p := helper.File(filepath.Join(dir.Preview, uid+".png"))
	t := helper.File(filepath.Join(dir.Thumbnail, uid+".png"))
	if d && p && t {
		return ""
	}
	if !d {
		missing = append(missing, "offer a file for download")
	}
	if art.Platform.String == tags.Audio.String() {
		return strings.Join(missing, " + ")
	}
	if !p {
		missing = append(missing, "create a preview image")
	}
	if !t {
		missing = append(missing, "create a thumbnail image")
	}
	return strings.Join(missing, " + ")
}

// attributions returns the author attributions for the file record of the artifact.
func (dir Dirs) attributions(art *models.File, data map[string]interface{}) map[string]interface{} {
	data["writers"] = mf.AttrWriter(art)
	data["artists"] = mf.AttrArtist(art)
	data["programmers"] = mf.AttrProg(art)
	data["musicians"] = mf.AttrMusic(art)
	return data
}

// filemetadata returns the file metadata for the file record of the artifact.
func (dir Dirs) filemetadata(art *models.File, data map[string]interface{}) map[string]interface{} {
	data["filename"] = mf.Basename(art)
	data["filesize"] = str.BytesHuman(art.Filesize.Int64)
	data["filebyte"] = art.Filesize
	data["lastmodified"] = mf.LastModification(art)
	data["lastmodifiedAgo"] = mf.LastModificationAgo(art)
	data["checksum"] = mf.Checksum(art)
	data["magic"] = mf.Magic(art)
	data["releasers"] = releasersHrefs(art)
	data["published"] = mf.Date(art)
	data["section"] = mf.TagCategory(art)
	data["platform"] = mf.TagProgram(art)
	data["alertURL"] = mf.AlertURL(art)
	data["extraZip"] = mf.ExtraZip(art, dir.Extra)
	return data
}

// otherRelations returns the other relations and external links for the file record of the artifact.
func (dir Dirs) otherRelations(art *models.File, data map[string]interface{}) map[string]interface{} {
	data["relations"] = mf.Relations(art)
	data["websites"] = mf.Websites(art)
	data["demozoo"] = mf.IdenficationDZ(art)
	data["pouet"] = mf.IdenficationPouet(art)
	data["sixteenColors"] = mf.Idenfication16C(art)
	data["youtube"] = mf.IdenficationYT(art)
	data["github"] = mf.IdenficationGitHub(art)
	return data
}

// jsdos returns the js-dos emulator data for the file record of the artifact.
func jsdos(art *models.File, data map[string]interface{}, logger *zap.SugaredLogger,
) map[string]interface{} {
	if logger == nil || art == nil {
		return data
	}
	data["jsdos6"] = false
	data["jsdos6Run"] = ""
	data["jsdos6RunGuess"] = ""
	data["jsdos6Config"] = ""
	data["jsdos6Zip"] = false
	data["jsdos6Utilities"] = false
	if emulate := mf.JsdosUse(art); !emulate {
		return data
	}
	data["jsdos6"] = true
	cmd, err := model.JsDosCommand(art)
	if err != nil {
		logger.Error(errorWithID(err, "js-dos command", art.ID))
		return data
	}
	data["jsdos6Run"] = cmd
	guess, err := model.JsDosBinary(art)
	if err != nil {
		logger.Error(errorWithID(err, "js-dos binary", art.ID))
		return data
	}
	data["jsdos6RunGuess"] = guess
	cfg, err := model.JsDosConfig(art)
	if err != nil {
		logger.Error(errorWithID(err, "js-dos config", art.ID))
		return data
	}
	data["jsdos6Config"] = cfg
	data["jsdos6Zip"] = mf.JsdosArchive(art)
	data["jsdos6Utilities"] = mf.JsdosUtilities(art)
	return data
}

// content returns the archive content for the file download of the artifact.
func content(art *models.File, data map[string]interface{}) map[string]interface{} {
	if art == nil {
		return data
	}
	data["content"] = ""
	data["contentDesc"] = ""
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

// errorWithID returns an error with the artifact ID appended to the error message.
// The key string is expected any will always be displayed in the error message.
// The id can be an integer or string value and should be the database numeric ID.
func errorWithID(err error, key string, id any) error {
	if err == nil {
		return nil
	}
	key = strings.TrimSpace(key)
	const cause = "caused by artifact"
	switch id.(type) {
	case int, int64:
		return fmt.Errorf("%w: %s %s (%d)", err, cause, key, id)
	case string:
		return fmt.Errorf("%w: %s %s (%s)", err, cause, key, id)
	default:
		return fmt.Errorf("%w: %s %s", err, cause, key)
	}
}

// embedText embeds the readme or file download text content for the file record of the artifact.
func embedText(art *models.File, data map[string]interface{}, b ...byte) (map[string]interface{}, error) {
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
		chk     = "checked"
	)
	textEncoding := render.Encoder(art, bytes.NewReader(b))
	data["topazCheck"] = ""
	data["vgaCheck"] = ""
	switch textEncoding {
	case charmap.ISO8859_1:
		data["readmeLatin1Cls"] = ""
		data["readmeCP437Cls"] = "d-none" + space
		data["topazCheck"] = chk
		b = bytes.ReplaceAll(b, []byte{nbsp}, []byte{sp})
		b = bytes.ReplaceAll(b, []byte{shy}, []byte{hyphen})
	case charmap.CodePage437:
		data["readmeLatin1Cls"] = "d-none" + space
		data["readmeCP437Cls"] = ""
		data["vgaCheck"] = chk
		b = bytes.ReplaceAll(b, []byte{nbsp437}, []byte{sp})
	case unicode.UTF8:
		// use Cad font as default
		data["readmeLatin1Cls"] = "d-none" + space
		data["readmeCP437Cls"] = ""
		data["vgaCheck"] = chk
	}
	var readme string
	var err error
	switch textEncoding {
	case unicode.UTF8:
		// unicode should apply to both latin1 and cp437
		readme, err = decode(bytes.NewReader(b))
		if err != nil {
			return data, fmt.Errorf("unicode utf8 decode: %w", err)
		}
		data["readmeLatin1"] = readme
		data["readmeCP437"] = readme
	default:
		d := charmap.ISO8859_1.NewDecoder().Reader(bytes.NewReader(b))
		readme, err = decode(d)
		if err != nil {
			return data, fmt.Errorf("iso8859_1 decode: %w", err)
		}
		data["readmeLatin1"] = readme
		d = charmap.CodePage437.NewDecoder().Reader(bytes.NewReader(b))
		readme, err = decode(d)
		if err != nil {
			return data, fmt.Errorf("codepage437 decode: %w", err)
		}
		data["readmeCP437"] = readme
	}
	data["readmeLines"] = strings.Count(readme, "\n")
	data["readmeRows"] = helper.MaxLineLength(readme)
	return data, nil
}

// decode decodes the text content from the reader.
func decode(src io.Reader) (string, error) {
	out := strings.Builder{}
	if _, err := io.Copy(&out, src); err != nil {
		return "", fmt.Errorf("io.Copy: %w", err)
	}
	if !strings.HasSuffix(out.String(), "\n\n") {
		out.WriteString("\n")
	}
	return out.String(), nil
}

// firstLead returns the lead for the file record which is the filename and releasers.
func firstLead(art *models.File) string {
	fname := art.Filename.String
	span := fmt.Sprintf("<span class=\"font-monospace fs-6 fw-light\">%s</span> ", fname)
	rels := string(LinkRels(art.GroupBrandBy, art.GroupBrandFor))
	return fmt.Sprintf("%s<br>%s", rels, span)
}

// releasersHrefs returns the releasers for the file record as a string of HTML links.
func releasersHrefs(art *models.File) string {
	if art == nil {
		return ""
	}
	return string(LinkRels(art.GroupBrandBy, art.GroupBrandFor))
}

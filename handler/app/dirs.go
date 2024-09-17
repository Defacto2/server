package app

// Package file dirs.go contains the artifact page directories and handlers.

import (
	"bytes"
	"context"
	"database/sql"
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

	"github.com/Defacto2/archive/pkzip"
	"github.com/Defacto2/archive/rezip"
	"github.com/Defacto2/helper"
	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/handler/app/internal/filerecord"
	"github.com/Defacto2/server/handler/app/internal/simple"
	"github.com/Defacto2/server/handler/readme"
	"github.com/Defacto2/server/handler/render"
	"github.com/Defacto2/server/handler/sess"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/postgres/models"
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
func (dir Dirs) Artifact(c echo.Context, db *sql.DB, logger *zap.SugaredLogger, readonly bool) error {
	const name = "artifact"
	art, err := dir.modelsFile(c, db)
	if art404 := art == nil || err != nil; art404 {
		return err
	}
	data := empty(c)
	if !readonly {
		data = dir.Editor(art, data)
	}
	// page metadata
	data["unid"] = filerecord.UnID(art)
	data["download"] = filerecord.DownloadID(art)
	data["title"] = filerecord.Basename(art)
	data["description"] = filerecord.Description(art)
	data["h1"] = filerecord.FirstHeader(art)
	data["lead"] = firstLead(art)
	data["comment"] = filerecord.Comment(art)
	data = dir.filemetadata(art, data)
	if !readonly {
		data = dir.updateMagics(db, logger, art.ID, art.UUID.String, data)
	}
	data = dir.attributions(art, data)
	data = dir.otherRelations(art, data)
	data = jsdos(art, data, logger)
	data = content(art, data)
	data["linkpreview"] = filerecord.LinkPreview(art)
	data["linkpreviewTip"] = filerecord.LinkPreviewTip(art)
	data["filentry"] = filerecord.FileEntry(art)
	if skip := render.NoScreenshot(art, dir.Preview); skip {
		data["noScreenshot"] = true
	}
	if filerecord.EmbedReadme(art) {
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

func repackZIP(name string) bool {
	x, err := pkzip.Methods(name)
	if err != nil {
		return false
	}
	for _, method := range x {
		if !method.Zip() {
			return true
		}
	}
	return false
}

func (dir Dirs) compressZIP(root, uid string) (int64, error) {
	basename := uid + ".zip"
	tmpArc := filepath.Join(helper.TmpDir(), basename)
	finalArc := filepath.Join(dir.Extra, basename)
	os.Remove(finalArc)

	fmt.Println("modDecompressLoc", root, "tmpArc", tmpArc)
	i, err := rezip.CompressDir(root, tmpArc)
	if err != nil {
		return 0, err
	}
	fmt.Println("rezipped extra files written:", i)

	if err = helper.RenameCrossDevice(tmpArc, finalArc); err != nil {
		defer os.RemoveAll(tmpArc)
		return 0, err
	}
	st, err := os.Stat(finalArc)
	if err != nil {
		return 0, err
	}
	return st.Size(), nil
}

// updateMagics updates the magic number for the file record of the artifact.
// It must be called after both the dir.filemetadata and dir.Editor functions.
func (dir Dirs) updateMagics(db *sql.DB, logger *zap.SugaredLogger,
	id int64, uid string, data map[string]interface{}) map[string]interface{} {
	if db == nil {
		return data
	}
	recMagic, modMagic := data["magic"], data["modMagicNumber"]
	if recMagic != modMagic {
		data["magic"] = modMagic
		magic := modMagic.(string)
		ctx := context.Background()
		if err := model.UpdateMagic(ctx, db, id, magic); err != nil && logger != nil {
			logger.Error(errorWithID(err, "update artifact editor magic", id))
		}
	}
	if findRepack := data["extraZip"].(bool); findRepack {
		return data
	}
	name := filepath.Join(dir.Download, uid)
	decompDir := data["modDecompressLoc"].(string)
	if st, err := os.Stat(decompDir); err != nil || !st.IsDir() {
		if logger != nil {
			logger.Error(errorWithID(err, "decompress directory", uid))
		}
		return data
	}
	switch {
	case redundantArchive(modMagic):
	case modMagic == magicnumber.PKWAREZip.Title():
		if !repackZIP(name) {
			return data
		}
	case plainText(modMagic):
		dirs := command.Dirs{
			Download:  dir.Download,
			Preview:   dir.Preview,
			Thumbnail: dir.Thumbnail,
		}
		if helper.File(filepath.Join(dirs.Thumbnail, uid+".png")) ||
			helper.File(filepath.Join(dirs.Thumbnail, uid+".webp")) {
			return data
		}
		if err := dirs.TextImager(logger, name, uid); err != nil {
			logger.Error(errorWithID(err, "text imager", uid))
		}
		data["missingAssets"] = ""
		return data
	default:
		return data
	}
	i, err := dir.compressZIP(decompDir, uid)
	if err != nil {
		if logger != nil {
			logger.Error(errorWithID(err, "compress directory", uid))
		}
		return data
	}
	if logger != nil {
		logger.Infof("Extra deflated zipfile created %d bytes: %s", i, uid)
	}
	data["extraZip"] = true
	return data
}

func redundantArchive(modMagic interface{}) bool {
	switch modMagic.(type) {
	case string:
	default:
		return false
	}
	switch modMagic.(string) {
	case
		magicnumber.ARChiveSEA.Title(),
		magicnumber.YoshiLHA.Title(),
		magicnumber.ArchiveRobertJung.Title(),
		magicnumber.PKWAREZipImplode.Title(),
		magicnumber.PKWAREZipReduce.Title(),
		magicnumber.PKWAREZipShrink.Title():
		return true
	default:
		return false
	}
}

func plainText(modMagic interface{}) bool {
	switch modMagic.(type) {
	case string:
	default:
		return false
	}
	switch modMagic.(string) {
	case
		magicnumber.UTF8Text.Title(),
		magicnumber.ANSIEscapeText.Title(),
		magicnumber.PlainText.Title():
		return true
	default:
		return false
	}
}

func (dir Dirs) embed(art *models.File, data map[string]interface{}) (map[string]interface{}, error) {
	if art == nil {
		return data, nil
	}
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
	unid := filerecord.UnID(art)
	abs := filepath.Join(dir.Download, unid)
	data["epochYear"] = epoch
	data["readonlymode"] = false
	data["modID"] = art.ID
	data["modTitle"] = filerecord.Title(art)
	data["modOnline"] = filerecord.RecordOnline(art)
	data["modReleasers"] = RecordRels(art.GroupBrandBy, art.GroupBrandFor)
	data["modReleaser1"], data["modReleaser2"] = filerecord.ReleaserPair(art)
	data["modYear"], data["modMonth"], data["modDay"] = filerecord.Dates(art)
	data["modLMYear"], data["modLMMonth"], data["modLMDay"] = filerecord.LastModifications(art)
	data["modAbsDownload"] = abs
	data["modMagicMime"] = simple.MIME(abs)
	data["modMagicNumber"] = simple.MagicAsTitle(abs)
	data["modDBModify"] = filerecord.LastModificationDate(art)
	data["modStatModify"], data["modStatSizeB"], data["modStatSizeF"] = simple.StatHumanize(abs)
	data["modDecompress"] = filerecord.ListContent(art, abs)
	data["modDecompressLoc"] = simple.MkContent(abs)
	data["modAssetPreview"] = dir.assets(dir.Preview, unid)
	data["modAssetThumbnail"] = dir.assets(dir.Thumbnail, unid)
	data["modAssetExtra"] = dir.assets(dir.Extra, unid)
	data["modNoReadme"] = filerecord.ReadmeNone(art)
	data["modReadmeSuggest"] = filerecord.Readme(art)
	data["modZipContent"] = filerecord.ZipContent(art)
	data["modRelations"] = filerecord.RelationsStr(art)
	data["modWebsites"] = filerecord.WebsitesStr(art)
	data["modOS"] = filerecord.TagProgram(art)
	data["modTag"] = filerecord.TagCategory(art)
	data["alertURL"] = filerecord.AlertURL(art)
	data["forApproval"] = filerecord.RecordIsNew(art)
	data["disableApproval"] = filerecord.RecordProblems(art)
	data["disableRecord"] = filerecord.RecordOffline(art)
	data["missingAssets"] = dir.missingAssets(art)
	data["modEmulateXMS"], data["modEmulateEMS"], data["modEmulateUMB"] = filerecord.JsdosMemory(art)
	data["modEmulateBroken"] = filerecord.JsdosBroken(art)
	data["modEmulateRun"] = filerecord.JsdosRun(art)
	data["modEmulateCPU"] = filerecord.JsdosCPU(art)
	data["modEmulateMachine"] = filerecord.JsdosMachine(art)
	data["modEmulateAudio"] = filerecord.JsdosSound(art)
	return data
}

// modelsFile returns the URI artifact record from the file table.
func (dir Dirs) modelsFile(c echo.Context, db *sql.DB) (*models.File, error) {
	ctx := context.Background()
	var art *models.File
	var err error
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
				matches[s] = simple.ImageXY(filepath.Join(nameDir, file.Name()))
			case ".PNG":
				s := "PNG"
				matches[s] = simple.ImageXY(filepath.Join(nameDir, file.Name()))
			case ".TXT":
				s := "README"
				i, _ := helper.Lines(filepath.Join(dir.Extra, file.Name()))
				matches[s] = [2]string{humanize.Comma(st.Size()), fmt.Sprintf("%d lines", i)}
			case ".WEBP":
				s := "WebP"
				matches[s] = simple.ImageXY(filepath.Join(nameDir, file.Name()))
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
	if art == nil {
		return ""
	}
	uid := art.UUID.String
	missing := []string{}
	dl := helper.File(filepath.Join(dir.Download, uid))
	pv := helper.File(filepath.Join(dir.Preview, uid+".png")) ||
		helper.File(filepath.Join(dir.Preview, uid+".webp"))
	th := helper.File(filepath.Join(dir.Thumbnail, uid+".png")) ||
		helper.File(filepath.Join(dir.Thumbnail, uid+".webp"))
	if dl && pv && th {
		return ""
	}
	if !dl {
		missing = append(missing, "offer a file for download")
	}
	platform := strings.TrimSpace(art.Platform.String)
	if platform == tags.Audio.String() {
		return strings.Join(missing, " + ")
	}
	textfiles := platform == tags.Text.String() || platform == tags.TextAmiga.String()
	if !pv && !textfiles {
		missing = append(missing, "create a preview image")
	}
	if !th {
		missing = append(missing, "create a thumbnail image")
	}
	return strings.Join(missing, " + ")
}

// attributions returns the author attributions for the file record of the artifact.
func (dir Dirs) attributions(art *models.File, data map[string]interface{}) map[string]interface{} {
	if art == nil {
		return data
	}
	data["writers"] = filerecord.AttrWriter(art)
	data["artists"] = filerecord.AttrArtist(art)
	data["programmers"] = filerecord.AttrProg(art)
	data["musicians"] = filerecord.AttrMusic(art)
	return data
}

// filemetadata returns the file metadata for the file record of the artifact.
func (dir Dirs) filemetadata(art *models.File, data map[string]interface{}) map[string]interface{} {
	if art == nil {
		return data
	}
	data["filename"] = filerecord.Basename(art)
	data["filesize"] = simple.BytesHuman(art.Filesize.Int64)
	data["filebyte"] = art.Filesize
	data["lastmodified"] = filerecord.LastModification(art)
	data["lastmodifiedAgo"] = filerecord.LastModificationAgo(art)
	data["checksum"] = filerecord.Checksum(art)
	data["magic"] = filerecord.Magic(art)
	data["releasers"] = releasersHrefs(art)
	data["published"] = filerecord.Date(art)
	data["section"] = filerecord.TagCategory(art)
	data["platform"] = filerecord.TagProgram(art)
	data["alertURL"] = filerecord.AlertURL(art)
	data["extraZip"] = filerecord.ExtraZip(art, dir.Extra)
	return data
}

// otherRelations returns the other relations and external links for the file record of the artifact.
func (dir Dirs) otherRelations(art *models.File, data map[string]interface{}) map[string]interface{} {
	if art == nil {
		return data
	}
	data["relations"] = filerecord.Relations(art)
	data["websites"] = filerecord.Websites(art)
	data["demozoo"] = filerecord.IdenficationDZ(art)
	data["pouet"] = filerecord.IdenficationPouet(art)
	data["sixteenColors"] = filerecord.Idenfication16C(art)
	data["youtube"] = filerecord.IdenficationYT(art)
	data["github"] = filerecord.IdenficationGitHub(art)
	return data
}

// jsdos returns the js-dos emulator data for the file record of the artifact.
func jsdos(art *models.File, data map[string]interface{}, logger *zap.SugaredLogger,
) map[string]interface{} {
	if art == nil {
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
	data["jsdos6"] = true
	cmd, err := model.JsDosCommand(art)
	if err != nil {
		if logger != nil {
			logger.Error(errorWithID(err, "js-dos command", art.ID))
		}
		return data
	}
	data["jsdos6Run"] = cmd
	guess, err := model.JsDosBinary(art)
	if err != nil {
		if logger != nil {
			logger.Error(errorWithID(err, "js-dos binary", art.ID))
		}
		return data
	}
	data["jsdos6RunGuess"] = guess
	cfg, err := model.JsDosConfig(art)
	if err != nil {
		if logger != nil {
			logger.Error(errorWithID(err, "js-dos config", art.ID))
		}
		return data
	}
	data["jsdos6Config"] = cfg
	data["jsdos6Zip"] = filerecord.JsdosArchive(art)
	data["jsdos6Utilities"] = filerecord.JsdosUtilities(art)
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
	if len(b) == 0 || art == nil {
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
	if art == nil {
		return ""
	}
	fname := art.Filename.String
	span := fmt.Sprintf("<span class=\"font-monospace fs-6 fw-light\">%s</span> ", fname)
	return fmt.Sprintf("%s<br>%s", releasersHrefs(art), span)
}

// releasersHrefs returns the releasers for the file record as a string of HTML links.
func releasersHrefs(art *models.File) string {
	if art == nil {
		return ""
	}
	magazine := strings.TrimSpace(art.Section.String) == tags.Mag.String()
	return string(LinkRels(magazine, art.GroupBrandBy, art.GroupBrandFor))
}

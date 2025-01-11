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
		data = detectANSI(db, logger, art.ID, data)
	}
	// page metadata
	uri := filerecord.DownloadID(art)
	data["canonical"] = strings.Join([]string{"f", uri}, "/")
	data["unid"] = filerecord.UnID(art)
	data["download"] = uri
	data["title"] = filerecord.Basename(art)
	data["description"] = filerecord.Description(art)
	data["h1"] = filerecord.FirstHeader(art)
	data["lead"] = firstLead(art)
	data["comment"] = filerecord.Comment(art)
	data = dir.filemetadata(art, data)
	if !readonly {
		platform := filerecord.TagProgram(art)
		data = dir.updateMagics(db, logger, art.ID, art.UUID.String, platform, data)
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
			defer clear(data)
			logger.Error(errorWithID(err, dir.URI, art.ID))
		}
	}
	err = c.Render(http.StatusOK, name, data)
	defer clear(data)
	if err != nil {
		return InternalErr(c, name, errorWithID(err, dir.URI, art.ID))
	}
	return nil
}

func detectANSI(db *sql.DB, logger *zap.SugaredLogger, id int64, data map[string]interface{}) map[string]interface{} {
	if db == nil {
		return data
	}
	mos, valid := data["modOS"].(string)
	if !valid {
		return data
	}
	numb, valid := data["modMagicNumber"].(string)
	if !valid {
		return data
	}
	textfile := strings.EqualFold(mos, tags.Text.String())
	if textfile && numb == magicnumber.ANSIEscapeText.Title() {
		if err := model.UpdatePlatform(db, id, tags.ANSI.String()); err != nil && logger != nil {
			logger.Error(errorWithID(err, "update artifact editor platform", id))
		}
		data["platform"] = tags.ANSI.String()
	}
	return data
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
	src := filepath.Join(helper.TmpDir(), basename)
	dest := filepath.Join(dir.Extra, basename)
	os.Remove(dest)
	_, err := rezip.CompressDir(root, src)
	if err != nil {
		return 0, fmt.Errorf("dirs compress zip: %w", err)
	}
	if err = helper.RenameCrossDevice(src, dest); err != nil {
		defer os.RemoveAll(src)
		return 0, fmt.Errorf("dirs compress zip: %w", err)
	}
	st, err := os.Stat(dest)
	if err != nil {
		return 0, fmt.Errorf("dirs compress zip: %w", err)
	}
	return st.Size(), nil
}

// updateMagics updates the magic number for the file record of the artifact.
// It must be called after both the dir.filemetadata and dir.Editor functions.
func (dir Dirs) updateMagics(db *sql.DB, logger *zap.SugaredLogger,
	id int64, uid, platform string, data map[string]interface{},
) map[string]interface{} {
	if db == nil {
		return data
	}
	recMagic, modMagic := data["magic"], data["modMagicNumber"]
	if recMagic != modMagic {
		data["magic"] = modMagic
		magic, valid := modMagic.(string)
		if !valid {
			if logger != nil {
				logger.Error(errorWithID(ErrType, "modMagicNumber is string", uid))
			}
			return data
		}
		ctx := context.Background()
		if err := model.UpdateMagic(ctx, db, id, magic); err != nil && logger != nil {
			logger.Error(errorWithID(err, "update artifact editor magic", id))
		}
	}
	findRepack, valid := data["extraZip"].(bool)
	if !valid {
		if logger != nil {
			logger.Error(errorWithID(ErrType, "extraZip is bool", uid))
		}
		return data
	}
	if findRepack {
		return data
	}
	decompDir, valid := data["modDecompressLoc"].(string)
	if !valid {
		if logger != nil {
			logger.Error(errorWithID(ErrType, "modDecompressLoc is string", uid))
		}
		return data
	}
	if st, err := os.Stat(decompDir); err != nil || !st.IsDir() {
		if logger != nil {
			logger.Error(errorWithID(err, "decompress directory", uid))
		}
		return data
	}
	return dir.checkMagics(logger, uid, decompDir, platform, modMagic, data)
}

func (dir Dirs) checkMagics(logger *zap.SugaredLogger,
	uid, decompDir, platform string,
	modMagic interface{},
	data map[string]interface{},
) map[string]interface{} {
	name := filepath.Join(dir.Download, uid)
	switch {
	case redundantArchive(modMagic):
	case modMagic == magicnumber.PKWAREZip.Title():
		if !repackZIP(name) {
			return data
		}
	case plainText(modMagic):
		return dir.plainTexts(logger, uid, platform, data)
	default:
		return data
	}
	if i, err := dir.compressZIP(decompDir, uid); err != nil {
		if logger != nil {
			logger.Error(errorWithID(err, "compress directory", uid))
		}
		return data
	} else if logger != nil {
		logger.Infof("Extra deflated zipfile created %d bytes: %s", i, uid)
	}
	data["extraZip"] = true
	return data
}

func (dir Dirs) plainTexts(logger *zap.SugaredLogger,
	uid, platform string, data map[string]interface{},
) map[string]interface{} {
	name := filepath.Join(dir.Download, uid)
	dirs := command.Dirs{
		Download:  dir.Download,
		Preview:   dir.Preview,
		Thumbnail: dir.Thumbnail,
	}
	if helper.File(filepath.Join(dirs.Thumbnail, uid+".png")) ||
		helper.File(filepath.Join(dirs.Thumbnail, uid+".webp")) {
		return data
	}
	amigaFont := strings.EqualFold(platform, tags.TextAmiga.String())
	if err := dirs.TextImager(logger, name, uid, amigaFont); err != nil {
		logger.Error(errorWithID(err, "text imager", uid))
	}
	data["missingAssets"] = ""
	return data
}

func redundantArchive(modMagic interface{}) bool {
	switch modMagic.(type) {
	case string:
	default:
		return false
	}
	val, valid := modMagic.(string)
	if !valid {
		return false
	}
	switch val {
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
	val, valid := modMagic.(string)
	if !valid {
		return false
	}
	switch val {
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
		return data, fmt.Errorf("dirs.embed read: %w", err)
	}
	d, err := embedText(art, data, p...)
	if err != nil {
		return data, fmt.Errorf("dirs.embed text: %w", err)
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
	d := command.Dirs{
		Download:  dir.Download,
		Preview:   dir.Preview,
		Thumbnail: dir.Thumbnail,
		Extra:     dir.Extra,
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
	data["modDecompress"] = filerecord.ListContent(art, d, abs)
	data["modDecompressLoc"] = simple.MkContent(abs)
	// These operations must be done using os.Stat and not os.ReadDir or filepath.WalkDir.
	// Previous attempts to use a shared function with WalkDir caused a memory leakages when
	// the site was under heavy load.
	data["modAssetPreview"] = dir.previews(unid)
	data["modAssetThumbnail"] = dir.thumbnails(unid)
	data["modAssetExtra"] = dir.extras(unid)
	data["missingAssets"] = dir.missingAssets(art)
	//
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

// Previews returns a map of preview assets for the file record of the artifact.
// Up to four preview assets are returned, JPEG, PNG, WebP and AVIF.
func (dir Dirs) previews(unid string) map[string][2]string {
	unid = strings.ToLower(unid)
	avif := filepath.Join(dir.Preview, unid+".avif")
	jpg := filepath.Join(dir.Preview, unid+".jpg")
	png := filepath.Join(dir.Preview, unid+".png")
	webp := filepath.Join(dir.Preview, unid+".webp")
	matches := make(map[string][2]string, 4)
	matches["Jpeg"] = simple.ImageXY(jpg)
	matches["PNG"] = simple.ImageXY(png)
	matches["WebP"] = simple.ImageXY(webp)
	if s, err := os.Stat(avif); err == nil {
		matches["AVIF"] = [2]string{humanize.Comma(s.Size()), ""}
	}
	return matches
}

// Thumbnails returns a map of thumbnail assets for the file record of the artifact.
// Two thumbnail assets are returned, PNG and WebP.
func (dir Dirs) thumbnails(unid string) map[string][2]string {
	unid = strings.ToLower(unid)
	png := filepath.Join(dir.Thumbnail, unid+".png")
	webp := filepath.Join(dir.Thumbnail, unid+".webp")
	matches := make(map[string][2]string, 2)
	matches["PNG"] = simple.ImageXY(png)
	matches["WebP"] = simple.ImageXY(webp)
	return matches
}

// Extras returns a map of extra assets for the file record of the artifact.
// Up to three extra assets are returned, FILE_ID, README and Repacked ZIP.
func (dir Dirs) extras(unid string) map[string][2]string {
	unid = strings.ToLower(unid)
	matches := make(map[string][2]string, 3)
	diz := filepath.Join(dir.Extra, unid+".diz")
	if s, err := os.Stat(diz); err == nil {
		i, _ := helper.Lines(diz)
		matches["FILE_ID"] = [2]string{humanize.Comma(s.Size()), fmt.Sprintf("%d lines", i)}
	}
	txt := filepath.Join(dir.Extra, unid+".txt")
	if s, err := os.Stat(txt); err == nil {
		i, _ := helper.Lines(txt)
		matches["README"] = [2]string{humanize.Comma(s.Size()), fmt.Sprintf("%d lines", i)}
	}
	zip := filepath.Join(dir.Extra, unid+".zip")
	if s, err := os.Stat(zip); err == nil {
		matches["Repacked ZIP"] = [2]string{humanize.Comma(s.Size()), "Deflate compression"}
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
	if len(b) == 0 || art == nil || art.RetrotxtNoReadme.Int16 != 0 {
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
	return string(LinkRelrs(magazine, art.GroupBrandBy, art.GroupBrandFor))
}

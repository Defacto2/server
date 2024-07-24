package app

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strings"
	uni "unicode"

	"github.com/Defacto2/server/handler/sess"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/magicnumber"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/render"
	"github.com/Defacto2/server/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

const epoch = model.EpochYear // epoch is the default year for MS-DOS files without a timestamp

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
	// artifact editor
	data = dir.artifactEditor(art, data, readonly)
	// page metadata
	data["unid"] = unid(art)
	data["download"] = downloadID(art)
	data["title"] = basename(art)
	data["description"] = description(art)
	data["h1"] = firstHeader(art)
	data["lead"] = firstLead(art)
	data["comment"] = comment(art)
	// file metadata
	data = dir.filemetadata(art, data)
	// attributions and credits
	data = dir.attributions(art, data)
	// links to other records and sites
	data = dir.otherRelations(art, data)
	// js-dos emulator
	data = jsdos(art, data, logger)
	// archive file content
	data = content(art, data)
	// record metadata
	data = recordmetadata(art, data)
	// readme text
	d, err := dir.artifactReadme(art)
	if err != nil {
		return InternalErr(c, name, errorWithID(err, dir.URI, art.ID))
	}
	maps.Copy(data, d)
	err = c.Render(http.StatusOK, name, data)
	if err != nil {
		return InternalErr(c, name, errorWithID(err, dir.URI, art.ID))
	}
	return nil
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

func (dir Dirs) attributions(art *models.File, data map[string]interface{}) map[string]interface{} {
	data["writers"] = attrWriter(art)
	data["artists"] = attrArtist(art)
	data["programmers"] = attrProg(art)
	data["musicians"] = attrMusic(art)
	return data
}

func (dir Dirs) filemetadata(art *models.File, data map[string]interface{}) map[string]interface{} {
	data["filename"] = basename(art)
	data["filesize"] = dirsBytes(art.Filesize.Int64)
	data["filebyte"] = art.Filesize
	data["lastmodified"] = lastModification(art)
	data["lastmodifiedAgo"] = lastModificationAgo(art)
	data["checksum"] = checksum(art)
	data["magic"] = magic(art)
	data["releasers"] = groupReleasers(art)
	data["published"] = dateIssued(art)
	data["section"] = section(art)
	data["platform"] = platform(art)
	data["alertURL"] = alertURL(art)
	data["extraZip"] = dir.extraZip(art)
	return data
}

func (dir Dirs) otherRelations(art *models.File, data map[string]interface{}) map[string]interface{} {
	data["relations"] = relations(art)
	data["websites"] = websites(art)
	data["demozoo"] = idenficationDZ(art)
	data["pouet"] = idenficationPouet(art)
	data["sixteenColors"] = idenfication16C(art)
	data["youtube"] = idenficationYT(art)
	data["github"] = idenficationGitHub(art)
	return data
}

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
	if emulate := jsdosUse(art); !emulate {
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
	data["jsdos6Zip"] = jsdosArchive(art)
	data["jsdos6Utilities"] = jsdosUtilities(art)
	return data
}

func recordmetadata(art *models.File, data map[string]interface{}) map[string]interface{} {
	if art == nil {
		return data
	}
	data["linkpreview"] = linkPreview(art)
	data["linkpreviewTip"] = linkPreviewTip(art)
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

// artifactReadme returns the readme data for the file record.
func (dir Dirs) artifactReadme(art *models.File) (map[string]interface{}, error) {
	data := map[string]interface{}{}
	if art == nil || art.RetrotxtNoReadme.Int16 != 0 {
		return data, nil
	}
	if unsupportedText(art) {
		return data, nil
	}
	if skip := render.NoScreenshot(art, dir.Download, dir.Preview); skip {
		data["noScreenshot"] = true
	}
	b, err := render.Read(art, dir.Download, dir.Extra)
	if err != nil {
		if errors.Is(err, render.ErrDownload) {
			data["noDownload"] = true
			return data, nil
		}
		if errors.Is(err, render.ErrFilename) {
			return data, nil
		}
		return data, fmt.Errorf("render.Read: %w", err)
	}
	if b == nil {
		return data, nil
	}
	r := bufio.NewReader(bytes.NewReader(b))
	// check the bytes are plain text but not utf16 or utf32
	if sign, err := magicnumber.Text(r); err != nil {
		return data, fmt.Errorf("magicnumber.Text: %w", err)
	} else if sign == magicnumber.Unknown ||
		sign == magicnumber.UTF16Text ||
		sign == magicnumber.UTF32Text {
		return data, nil
	}
	// trim trailing whitespace and MS-DOS era EOF marker
	b = bytes.TrimRightFunc(b, uni.IsSpace)
	const endOfFile = 0x1a // Ctrl+Z
	if bytes.HasSuffix(b, []byte{endOfFile}) {
		b = bytes.TrimSuffix(b, []byte{endOfFile})
	}
	if incompatible, err := incompatibleANSI(r); err != nil {
		return data, fmt.Errorf("incompatibleANSI: %w", err)
	} else if incompatible {
		return data, nil
	}
	b = removeControlCodes(b)
	return readmeEncoding(art, data, b...)
}

func readmeEncoding(art *models.File, data map[string]interface{}, b ...byte) (map[string]interface{}, error) {
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

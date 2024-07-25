package app

import (
	"fmt"
	"html/template"
	"image"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Defacto2/server/handler/app/internal/mf"
	"github.com/Defacto2/server/internal/archive"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/magicnumber"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/dustin/go-humanize"
	"github.com/h2non/filetype"
)

func (dir Dirs) Editor(art *models.File, data map[string]interface{}) map[string]interface{} {
	if art == nil {
		return data
	}
	unid := mf.UnID(art)
	abs := filepath.Join(dir.Download, unid)
	data["epochYear"] = epoch
	data["readOnly"] = false
	data["modID"] = art.ID
	data["modTitle"] = title(art)
	data["modOnline"] = public(art)
	data["modReleasers"] = RecordRels(art.GroupBrandBy, art.GroupBrandFor)
	data["modReleaser1"], data["modReleaser2"] = releaser1_2(art)
	data["modYear"], data["modMonth"], data["modDay"] = date(art)
	data["modLMYear"], data["modLMMonth"], data["modLMDay"] = dateLastMod(art)
	data["modAbsDownload"] = abs
	data["modMagicMime"] = artifactMIME(abs)
	data["modMagicNumber"] = magicNumber(abs)
	data["modDBModify"] = mf.LastModification(art)
	data["modStatModify"], data["modStatSizeB"], data["modStatSizeF"] = statHumanize(abs)
	data["modArchiveContent"] = artifactContent(abs, art.Platform.String) // FIXME
	data["modArchiveContentDst"], _ = artifactContentDst(abs)             // FIXME, error handling and return empty string
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
	data["modOS"] = mf.TagOS(art)
	data["modTag"] = mf.TagCategory(art)
	data["virusTotal"] = mf.AlertURL(art) // FIXME, virusTotal is a dupe of ["alertURL"] ?
	data["forApproval"] = recordIsNew(art)
	data["disableApproval"] = disableApproval(art)
	data["disableRecord"] = recordOffline(art)
	data["missingAssets"] = missingAssets(art, dir)
	// jsdos emulator
	data["modEmulateXMS"], data["modEmulateEMS"], data["modEmulateUMB"] = jsdosMemory(art)
	data["modEmulateBroken"] = jsdosBroken(art)
	data["modEmulateRun"] = jsdosRun(art)
	data["modEmulateCPU"] = jsdosCPU(art)
	data["modEmulateMachine"] = jsdosMachine(art)
	data["modEmulateAudio"] = jsdosSound(art)
	return data
}

func date(art *models.File) (int16, int16, int16) {
	if art == nil {
		return 0, 0, 0
	}
	y, m, d := int16(0), int16(0), int16(0)
	if art.DateIssuedYear.Valid {
		y = art.DateIssuedYear.Int16
	}
	if art.DateIssuedMonth.Valid {
		m = art.DateIssuedMonth.Int16
	}
	if art.DateIssuedDay.Valid {
		d = art.DateIssuedDay.Int16
	}
	return y, m, d
}

func dateLastMod(art *models.File) (int, int, int) {
	if art == nil {
		return 0, 0, 0
	}
	if !art.FileLastModified.Valid || art.FileLastModified.IsZero() {
		return 0, 0, 0
	}
	y := art.FileLastModified.Time.Year()
	m := int(art.FileLastModified.Time.Month())
	d := art.FileLastModified.Time.Day()
	return y, m, d
}

func public(art *models.File) bool {
	return art.Deletedat.Time.IsZero()
}

func releaser1_2(art *models.File) (string, string) {
	if art == nil {
		return "", ""
	}
	rr := RecordReleasers(art.GroupBrandFor, art.GroupBrandBy)
	return rr[0], rr[1]

}

func title(art *models.File) string {
	if art == nil {
		return ""
	}
	return art.RecordTitle.String
}

///

func magicNumber(name string) string {
	r, err := os.Open(name)
	if err != nil {
		return err.Error()
	}
	defer r.Close()
	sign, err := magicnumber.Find(r)
	if err != nil {
		return err.Error()
	}
	return sign.Title()
}

func disableApproval(art *models.File) string {
	validate := model.Validate(art)
	if validate == nil {
		return ""
	}
	x := strings.Split(validate.Error(), ",")
	s := make([]string, 0, len(x))
	for _, v := range x {
		if strings.TrimSpace(v) == "" {
			continue
		}
		s = append(s, v)
	}
	s = slices.Clip(s)
	return strings.Join(s, " + ")
}

func missingAssets(art *models.File, dir Dirs) string {
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

// artifactContentDst returns the destination directory for the extracted archive content.
// The directory is created if it does not exist. The directory is named after the source file.
func artifactContentDst(src string) (string, error) {
	name := strings.TrimSpace(strings.ToLower(filepath.Base(src)))
	dir := filepath.Join(os.TempDir(), "defacto2-server")

	pattern := "artifact-content-" + name
	dst := filepath.Join(dir, pattern)
	if st, err := os.Stat(dst); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dst, os.ModePerm); err != nil {
				return "", err
			}
			return dst, nil
		}
		return dst, nil
	} else if !st.IsDir() {
		return "", fmt.Errorf("error, not a directory: %s", dir)
	}
	return dst, nil
}

func artifactContent(src, platform string) template.HTML {
	const mb150 = 150 * 1024 * 1024
	if st, err := os.Stat(src); err != nil {
		return template.HTML(err.Error())
	} else if st.IsDir() {
		return "error, directory"
	} else if st.Size() > mb150 {
		return "will not decompress this archive as it is very large"
	}
	dst, err := artifactContentDst(src)
	if err != nil {
		return template.HTML(err.Error())
	}

	if entries, _ := os.ReadDir(dst); len(entries) == 0 {
		if err := archive.ExtractAll(src, dst); err != nil {
			defer os.RemoveAll(dst)
			return template.HTML(err.Error())
		}
	}

	files := 0
	var walkerCount = func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		files++
		return nil
	}
	if err := filepath.WalkDir(dst, walkerCount); err != nil {
		return template.HTML(err.Error())
	}

	var b strings.Builder
	items, zeroByteFiles := 0, 0
	var walkerFunc = func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, err := filepath.Rel(dst, path)
		if err != nil {
			debug := fmt.Sprintf(`<div class="border-bottom row mb-1">... %v more files</div>`, err)
			b.WriteString(debug)
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		bytes := info.Size()
		if bytes == 0 {
			zeroByteFiles++
			return nil
		}
		size := humanize.Bytes(uint64(info.Size()))
		image := false
		texts := false
		program := false
		r, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer r.Close()
		sign, err := magicnumber.Find512B(r)
		if err != nil {
			return nil
		}
		for _, v := range magicnumber.Images() {
			if v == sign {
				image = true
				break
			}
		}
		for _, v := range magicnumber.Texts() {
			if v == sign {
				texts = true
				break
			}
		}
		for _, v := range magicnumber.Programs() {
			if strings.EqualFold(platform, tags.DOS.String()) {
				break
			}
			if v == sign {
				program = true
				break
			}
		}
		items++
		htm := fmt.Sprintf(`<div class="col d-inline-block text-truncate" data-bs-toggle="tooltip" data-bs-title="%s">%s</div>`,
			rel, rel)
		if image || texts {
			htm += `<div class="col col-1 text-end"><svg width="16" height="16" fill="currentColor" aria-hidden="true">` +
				`<use xlink:href="/svg/bootstrap-icons.svg#images"></use></svg></div>`
		} else {
			htm += `<div class="col col-1"></div>`
		}
		if texts {
			htm += `<div class="col col-1 text-end"><svg width="16" height="16" fill="currentColor" aria-hidden="true">` +
				`<use xlink:href="/svg/bootstrap-icons.svg#file-text"></use></svg></div>`
		} else if program {
			htm += `<div class="col col-1 text-end"><svg width="16" height="16" fill="currentColor" aria-hidden="true">` +
				`<use xlink:href="/svg/bootstrap-icons.svg#terminal-plus"></use></svg></div>`
		} else {
			htm += `<div class="col col-1"></div>`
		}
		htm += fmt.Sprintf(`<div><small data-bs-toggle="tooltip" data-bs-title="%d bytes">%s</small>`, bytes, size)
		htm += fmt.Sprintf(` <small class="">%s</small></div>`, sign)
		htm = fmt.Sprintf(`<div class="border-bottom row mb-1">%s</div>`, htm)
		b.WriteString(htm)
		if items > 200 {
			more := fmt.Sprintf(`<div class="border-bottom row mb-1">... %d more files</div>`, files-items)
			b.WriteString(more)
			return filepath.SkipAll
		}
		return nil
	}
	err = filepath.WalkDir(dst, walkerFunc)
	if err != nil {
		return template.HTML(err.Error())
	}
	if zeroByteFiles > 0 {
		zero := fmt.Sprintf(`<div class="border-bottom row mb-1">... skipped %d empty (0 B) files</div>`, zeroByteFiles)
		b.WriteString(zero)
	}
	return template.HTML(b.String())
}

// artifactMIME returns the MIME type for the file record.
func artifactMIME(name string) string {
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

	kind, err := filetype.Match(head)
	if err != nil {
		return err.Error()
	}
	if kind != filetype.Unknown {
		return kind.MIME.Value
	}

	return http.DetectContentType(head)
}

func recordIsNew(art *models.File) bool {
	if art == nil {
		return false
	}
	return !art.Deletedat.IsZero() && art.Deletedby.IsZero()
}

func recordOffline(art *models.File) bool {
	if art == nil {
		return false
	}
	return !art.Deletedat.IsZero() && !art.Deletedby.IsZero()
}

// artifactAssets returns a list of downloads and image assets belonging to the file record.
// any errors are appended to the list.
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
			s := ""
			switch ext {
			case ".AVIF":
				s = "AVIF"
				matches[s] = [2]string{humanize.Comma(st.Size()), ""}
			case ".JPG":
				s = "Jpeg"
				matches[s] = artifactImgInfo(filepath.Join(nameDir, file.Name()))
			case ".PNG":
				s = "PNG"
				matches[s] = artifactImgInfo(filepath.Join(nameDir, file.Name()))
			case ".TXT":
				s = "README"
				i, _ := helper.Lines(filepath.Join(dir.Extra, file.Name()))
				matches[s] = [2]string{humanize.Comma(st.Size()), fmt.Sprintf("%d lines", i)}
			case ".WEBP":
				s = "WebP"
				matches[s] = artifactImgInfo(filepath.Join(nameDir, file.Name()))
			case ".ZIP":
				s = "Repacked ZIP"
				matches[s] = [2]string{humanize.Comma(st.Size()), "Deflate compression"}
			}
		}
	}
	return matches
}

// artifactImgInfo returns the image file size and dimensions.
func artifactImgInfo(name string) [2]string {
	switch filepath.Ext(strings.ToLower(name)) {
	case ".jpg", ".jpeg", ".gif", ".png", ".webp":
	default:
		st, err := os.Stat(name)
		if err != nil {
			return [2]string{err.Error(), ""}
		}
		return [2]string{humanize.Comma(st.Size()), ""}
	}
	reader, err := os.Open(name)
	if err != nil {
		return [2]string{err.Error(), ""}
	}
	defer reader.Close()
	st, err := reader.Stat()
	if err != nil {
		return [2]string{err.Error(), ""}
	}
	config, _, err := image.DecodeConfig(reader)
	if err != nil {
		return [2]string{err.Error(), ""}
	}
	return [2]string{humanize.Comma(st.Size()), fmt.Sprintf("%dx%d", config.Width, config.Height)}
}

func jsdosBroken(art *models.File) bool {
	if art == nil {
		return false
	}
	if art.DoseeIncompatible.Valid {
		return art.DoseeIncompatible.Int16 != 0
	}
	return false
}

func jsdosCPU(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.DoseeHardwareCPU.Valid {
		return art.DoseeHardwareCPU.String
	}
	return ""
}

func jsdosMachine(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.DoseeHardwareGraphic.Valid {
		return art.DoseeHardwareGraphic.String
	}
	return ""
}

func jsdosMemory(art *models.File) (bool, bool, bool) {
	if art == nil {
		return false, false, false
	}
	x, e, u := false, false, false
	if art.DoseeNoXMS.Valid {
		x = art.DoseeNoXMS.Int16 == 0
	}
	if art.DoseeNoEms.Valid {
		e = art.DoseeNoEms.Int16 == 0
	}
	if art.DoseeNoUmb.Valid {
		u = art.DoseeNoUmb.Int16 == 0
	}
	return x, e, u
}

func jsdosRun(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.DoseeRunProgram.Valid {
		return art.DoseeRunProgram.String
	}
	return ""
}

func jsdosSound(art *models.File) string {
	if art == nil {
		return ""
	}
	if art.DoseeHardwareAudio.Valid {
		return art.DoseeHardwareAudio.String
	}
	return ""
}

// statHumanize returns the last modified date, size in bytes and size formatted
// of the named file.
func statHumanize(name string) (string, string, string) {
	stat, err := os.Stat(name)
	if err != nil {
		return "", "", err.Error()
	}
	return stat.ModTime().Format("2006-Jan-02"),
		humanize.Comma(stat.Size()),
		humanize.Bytes(uint64(stat.Size()))
}

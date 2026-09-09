//nolint:exhaustruct_v5
package filerecord

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/Defacto2/archive"
	"github.com/Defacto2/helper"
	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/handler/readme"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
)

type DownloadContent struct {
	Source   string       // Source path of the file archive.
	MaxItems int          // MaxItems are the number of file archive entries to list.
	Dirs     command.Dirs // Dirs pointing to the server asset paths.

	entries       int
	files         int
	zeroByteFiles int
	index         int
	unid          string
	platform      string
	section       string
	tempRoot      string
	name          string
	elems         []string
	results       []string
	names         []string
	html          strings.Builder
}

// List the files found in a file archive such as a packaged ZIP file.
// This is used to generate the HTML for the "Download content" section of the File editor.
//
// This should only ever be used by the admin editor mode.
// It extracts the file archve to a temporary directory,
// to allow its extracted content can be parsed for usability
// using magicfile techniques and other metadata.
func (dlc *DownloadContent) List(ctx context.Context, sl *slog.Logger, art *models.File) template.HTML {
	if nils.Slog("filerecord list context", ctx, sl, art) {
		return ""
	}

	if html := dlc.init(ctx, sl, art); html != "" {
		return html
	}

	// Create root-scoped filesystem for secure operations
	root, err := os.OpenRoot(dlc.tempRoot)
	if err != nil {
		return template.HTML("failed to open root directory: " + err.Error())
	}
	defer root.Close()

	// Use root-scoped WalkerChmod to prevent path traversal attacks
	chmodFn := func(path string, d fs.DirEntry, err error) error {
		return WalkerChmod(root, path, d, err)
	}
	if err := filepath.WalkDir(dlc.tempRoot, chmodFn); err != nil {
		return template.HTML(err.Error())
	}

	if err := filepath.WalkDir(dlc.tempRoot, dlc.countFn); err != nil {
		return template.HTML(err.Error())
	}

	dlc.elems = make([]string, dlc.files)
	dlc.index = -1
	if err = filepath.WalkDir(dlc.tempRoot, dlc.listFn); err != nil {
		return template.HTML(err.Error())
	}
	if len(dlc.elems) > dlc.MaxItems {
		dlc.elems = dlc.elems[:dlc.MaxItems]
	}
	dlc.results = readme.SortList(false,
		strings.Join(dlc.elems[:dlc.index+1], "\n"))

	if dlc.files > dlc.MaxItems {
		dlc.html.WriteString(`<div class="border-bottom row mb-1">skipped` +
			strconv.Itoa(dlc.files-dlc.MaxItems) + ` other files</div>`)
	}

	if err = filepath.WalkDir(dlc.tempRoot, dlc.htmlFn); err != nil {
		dlc.html.Reset()
		return template.HTML(err.Error())
	}

	dlc.names = helper.SortNames("/", dlc.names)

	return dlc.renderHTML(ctx, sl)
}

// init the func, or otherwise returns an error string for use in the HTML.
func (dlc *DownloadContent) init(
	ctx context.Context, sl *slog.Logger, art *models.File,
) template.HTML {
	var err error

	dlc.unid = art.UUID.String
	if !art.UUID.Valid {
		return "error, no UUID"
	}

	dlc.platform = strings.TrimSpace(strings.ToLower(art.Platform.String))
	if !tags.IsPlatform(dlc.platform) {
		return "error, invalid platform"
	}
	dlc.section = strings.TrimSpace(strings.ToLower(art.Section.String))

	dlc.tempRoot, err = archive.ExtractTemp(ctx, dlc.Source)
	if err != nil {
		return dlc.extractErr(sl, err)
	}

	return ""
}

func (dlc *DownloadContent) extractErr(sl *slog.Logger, err error) template.HTML {
	if sl == nil {
		sl = logs.Discard()
	}

	const msg = "list content of archive extraction"
	if !errors.Is(err, archive.ErrNotArchive) && !errors.Is(err, archive.ErrNotImplemented) {
		sl.Info(msg+" caused an error", slog.String("src", dlc.Source), slog.Any("error", err))
		return template.HTML(err.Error())
	}

	e := Entry{
		module:  "",
		size:    "",
		format:  "",
		exec:    magicnumber.Windows{},
		sign:    0,
		zeros:   dlc.zeroByteFiles,
		bytes:   0,
		image:   false,
		text:    false,
		bintext: false,
		program: false,
	}
	if e.SkipFile(dlc.Source, dlc.platform) {
		return "error, empty byte file"
	}

	le := listErr(e)
	var b strings.Builder
	_, err = b.WriteString(le.HTML(e.bytes, dlc.platform, dlc.section))
	if err != nil {
		sl.Info(msg+" caused a write string error", slog.Any("error", err))
	}

	return template.HTML(b.String())
}

// countFn quickly sums the number of found files to [DownloadContent.files].
func (dlc *DownloadContent) countFn(_ string, d fs.DirEntry, err error) error {
	if err != nil {
		return fs.SkipDir
	}
	if !d.IsDir() {
		dlc.files++
	}
	return nil
}

// listFn fills the [DownloadContent.elems] slice to list the relative file paths of [DownloadContent.tempRoot].
func (dlc *DownloadContent) listFn(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return filepath.SkipDir
	}
	var skipEntry error
	if d.IsDir() {
		return skipEntry
	}

	rel, err := filepath.Rel(dlc.tempRoot, path)
	if err != nil {
		return skipEntry //nolint:nilerr
	}

	rel = strings.TrimSpace(rel)
	if rel == "" {
		return skipEntry
	}

	dlc.index++
	dlc.elems[dlc.index] = rel

	return nil
}

// htmlFn builds the html needed to display a single file entry.
func (dlc *DownloadContent) htmlFn(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return filepath.SkipDir
	}

	var skipEntry error
	rel, err := filepath.Rel(dlc.tempRoot, path)
	if err != nil {
		return skipEntry //nolint:nilerr
	}

	if usefile := slices.Contains(dlc.results, rel); !usefile {
		return skipEntry
	}

	e := Entry{
		module:  "",
		size:    "",
		format:  "",
		exec:    magicnumber.Windows{},
		sign:    0,
		zeros:   dlc.zeroByteFiles,
		bytes:   0,
		image:   false,
		text:    false,
		bintext: false,
		program: false,
	}
	if e.SkipEntry(path, d, dlc.platform) {
		dlc.zeroByteFiles = e.zeros
		return skipEntry
	}
	dlc.entries++
	if e.text {
		dlc.name = d.Name()
		dlc.names = append(dlc.names, dlc.name)
	}

	le := listEntry(e, rel, dlc.unid)
	dlc.html.WriteString(le.HTML(e.bytes, dlc.platform, dlc.section))
	if dlc.entries > dlc.MaxItems {
		return filepath.SkipAll
	}

	return nil
}

func (dlc *DownloadContent) renderHTML(ctx context.Context, sl *slog.Logger) template.HTML {
	count := len(dlc.names)
	if count == 0 {
		dlc.html.WriteString(skippedEmpty(dlc.zeroByteFiles))
		return template.HTML(dlc.html.String())
	}

	// always render a file_id.idx if it is found
	idx := dlc.indexDiz()
	if useDiz := idx >= 0; useDiz {
		elms := ""
		if len(dlc.names) > idx {
			elms = dlc.names[idx]
		}

		srcDIZ := filepath.Join(dlc.tempRoot, elms)
		if err := dlc.Dirs.DizDeferred(sl, srcDIZ, dlc.unid); err != nil {
			dlc.html.Reset()
			return template.HTML(err.Error())
		}

		if count == 1 {
			dlc.html.WriteString(skippedEmpty(dlc.zeroByteFiles))
			return template.HTML(dlc.html.String())
		}
	}

	// render an NFO or text if only a single text is found,
	// excluding any file_id.diz files
	srcNFO := ""
	if onlyNFO := idx == -1 && count == 1; onlyNFO {
		elms := ""
		if len(dlc.names) > 0 {
			elms = dlc.names[0]
		}
		srcNFO = filepath.Join(dlc.tempRoot, elms)
	}

	const maxItems = 2
	if textPair := idx != -1 && count == maxItems; textPair {
		invert := 1 - idx
		elms := ""
		if len(dlc.names) > invert {
			elms = dlc.names[invert]
		}
		srcNFO = filepath.Join(dlc.tempRoot, elms)
	}

	if srcNFO != "" {
		if err := dlc.Dirs.TextDeferred(ctx, sl, srcNFO, dlc.unid); err != nil {
			dlc.html.Reset()
			return template.HTML(err.Error())
		}
	}

	dlc.html.WriteString(skippedEmpty(dlc.zeroByteFiles))
	return template.HTML(dlc.html.String())
}

func (dlc *DownloadContent) indexDiz() int {
	for i, name := range dlc.names {
		s := strings.TrimSpace(name)
		if strings.EqualFold(s, "file_id.diz") { // FIX: create a shared universal file_id. ref?
			return i
		}
	}
	return -1
}

// WalkerChmod changes the file permissions for the extracted files.
// There are odd cases where the extracted files from DOS era ZIP files have no permissions.
// Uses os.Root to prevent path traversal attacks (fixes G122 security issue).
func WalkerChmod(root *os.Root, path string, d fs.DirEntry, err error) error {
	if err != nil {
		return fs.SkipDir
	}

	const format = "walker chmod failure to %s: %s: %w"
	const dirRW, fileRW = 0o755, 0o644

	// Get relative path for root-scoped operations
	relPath, err := filepath.Rel(root.Name(), path)
	if err != nil {
		return fmt.Errorf(format, "get relative path", path, err)
	}

	if d.IsDir() {
		if err := root.Chmod(relPath, dirRW); err != nil {
			return fmt.Errorf(format, "chmod directory", relPath, err)
		}
		return nil
	}

	if err := root.Chmod(relPath, fileRW); err != nil {
		return fmt.Errorf(format, "chmod file", relPath, err)
	}
	return nil
}

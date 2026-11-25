package config

// Package file repair.go contains the repair functions for assets and downloads.

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"
	"uuid"

	"github.com/Defacto2/archive/rezip"
	"github.com/Defacto2/helper"
	"github.com/Defacto2/magicnumber"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/command/option"
	"github.com/Defacto2/server/internal/config/fixarc"
	"github.com/Defacto2/server/internal/config/fixarj"
	"github.com/Defacto2/server/internal/config/fixlha"
	"github.com/Defacto2/server/internal/config/fixzip"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/extensions"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/bengarrett/ansimate"
)

const (
	Timeout = 1 * time.Minute
)

const (
	cfid      = "00000000-0000-0000-0000000000000000" // coldfusion uuid example
	syncthing = ".stfolder"                           // syncthing directory name
)

// RepairArchives checks the download directory for any legacy and obsolete archives.
// Obsolete archives are those that use a legacy compression method
// that is not supported by Go or JS libraries used by the website.
func (c *Config) RepairArchives(ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor) error {
	return c.repairArchives(ctx, sl, exec, false)
}

func (c *Config) repairArchives(ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor, dryRun bool) error {
	const format = "config archives repair: %w"
	if err := nils.Check(ctx, sl, exec); err != nil {
		return fmt.Errorf(format, err)
	}

	start := time.Now()
	download := dir.Directory(c.AbsDownload.String())

	s := repairs()
	for repair := range slices.Values(s[:]) {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf(format, err)
		}

		if err := repair.lookups(); err != nil {
			sl.Error("archives "+repair.String(), slog.Any("error", err))
			continue
		}

		artifacts, err := repair.artifacts(ctx, sl, exec)
		if err != nil {
			sl.Error("archives "+repair.String(), slog.Any("error", err))
			continue
		}

		if err := c.repairArtifacts(ctx, sl, repair, artifacts, dryRun); err != nil {
			sl.Error("Archives repair",
				slog.Any("error", err),
				slog.String("path", download.Path()),
				slog.String("format", repair.String()))
		}
	}

	sl.Info("Archives check", slog.String("task", "Time taken"),
		slog.Duration("time", time.Since(start).Round(time.Millisecond)))

	return nil
}

// repairArtifacts inspects each candidate artifact directly in the download directory
// and repacks legacy archives into the extra directory.
func (c *Config) repairArtifacts(
	ctx context.Context, sl *slog.Logger, repair Repair, artifacts []string, dryrun bool,
) error {
	root := c.AbsDownload.String()
	extra := dir.Directory(c.AbsExtra)
	if root == "" {
		return nil
	}

	for _, uid := range artifacts {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf("repair artifacts ctx error: %w", ctxErr)
		}

		// Skip if already repacked in extras.
		if st, err := os.Stat(extra.Join(uid + ".zip")); err == nil && !st.IsDir() {
			continue
		}

		// Locate the download file directly (<download>/<uid>).
		path := filepath.Join(root, uid)
		st, err := os.Stat(path)
		if err != nil || st.IsDir() {
			continue
		}

		var invalid bool
		switch repair {
		case Zip:
			if !fixzip.Legacy(sl, path) {
				continue
			}
			invalid = fixzip.Invalid(ctx, sl, path)
		case LHA:
			invalid = fixlha.Invalid(ctx, sl, path)
		case Arc:
			invalid = fixarc.Invalid(ctx, sl, path)
		case Arj:
			invalid = fixarj.Invalid(ctx, sl, path)
		}
		if invalid {
			continue
		}

		if dryrun {
			sl.Info("[dry-run] would repack archive",
				slog.String("format", repair.String()),
				slog.String("uuid", uid),
				slog.String("source", path),
				slog.String("destination", extra.Join(uid+".zip")))
			continue
		}

		ra := Repack{Source: path, UID: uid, Destination: extra}
		if err := repair.NewArchive(ctx, sl, ra); err != nil {
			sl.Error("repair and re-archive failed",
				slog.String("format", repair.String()),
				slog.String("uuid", uid),
				slog.Any("error", err))
		}
	}

	return nil
}

// Repair is a type of archive for the re-archive and recompress methods.
type Repair int

const (
	Zip Repair = iota // ZIP and PKZip archives
	LHA               // LHA and LZH archives
	Arc               // ARC archives
	Arj               // ARJ archives
)

func repairs() [4]Repair {
	return [4]Repair{Zip, LHA, Arc, Arj}
}

func (r Repair) String() string {
	return [...]string{"zip", "lha", "arc", "arj"}[r]
}

// Repack are the source and destination arguments required by the ReArchive Repair method.
type Repack struct {
	Source      string        // Source is the file extracted to a temporary directory and re-compressed.
	UID         string        // UID is the destination filename using a universal unique ID naming syntax.
	Destination dir.Directory // Destination is the directory to save the re-compressed file.
}

// NewArchive extracts the source archive file and then repackages (rearchives) it to a DEFLATE zip archive.
// The original source file is always kept.
func (r Repair) NewArchive(ctx context.Context, sl *slog.Logger, ra Repack) error { //nolint:funlen
	const format = "config repair and repack %s: %w"
	if err := nils.Check(ctx, sl); err != nil {
		return fmt.Errorf(format, "check", err)
	}
	if ra.Source == "" || ra.UID == "" {
		return fmt.Errorf(format, "source or uid are missing", ErrNoPath)
	}
	if err := ra.Destination.IsDir(); err != nil {
		return fmt.Errorf(format, "destination is not a directory", err)
	}

	// resolve extraction command per format
	extractCmd, extractArg, err := r.extractConfig()
	if err != nil {
		return fmt.Errorf(format, "extract config", err)
	}

	// prepare temporary directory
	root, err := dir.MkdirTemp("newarc")
	if err != nil {
		return fmt.Errorf(format, "make temp dir", err)
	}
	defer func() {
		if rmErr := dir.RemoveAll(root); rmErr != nil {
			sl.Error("repack failed to remove temp dir", slog.Any("error", rmErr))
		}
	}()

	// extract source archive
	timeout, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	cr := command.Runner{
		Timeout:    1 * Timeout,
		Log:        sl,
		WorkingDir: root,
	}
	arg := option.Join(extractArg, ra.Source)
	out, err := cr.Run(timeout, extractCmd, arg...)
	if err != nil {
		return fmt.Errorf(format, "run command error: "+string(out), err)
	}

	// inspect extracted files
	files, err := helper.Count(root)
	if err != nil {
		return fmt.Errorf(format, "tmp count", err)
	}

	const msg = "Repacked archive"
	sl.Info(msg, slog.String("entry", ra.UID),
		slog.Int("file(s)", files), slog.String("tmp", root))

	// compress extracted directory into standard ZIP
	name := ra.UID + ".zip"
	dest := filepath.Join(root, name)

	written, err := rezip.CompressDir(root, dest)
	if err != nil {
		return fmt.Errorf(format, "compress dir", err)
	}
	if written == 0 {
		return nil
	}

	// move temporary archive to destination
	newpath := ra.Destination.Join(name)
	if err := helper.RenameCrossDevice(dest, newpath); err != nil {
		return fmt.Errorf(format, "rename archive", err)
	}

	st, err := os.Stat(newpath)
	if err != nil {
		return fmt.Errorf(format, "stat archive", err)
	}

	sl.Info(msg, slog.String("Re-archive", "Contemporary 'deflate' zip archive created"),
		slog.Int64("bytes", st.Size()), slog.String("path", newpath))

	return nil
}

// extractConfig maps the Repair type to its extraction command and argument.
func (r Repair) extractConfig() (cmd, arg string, err error) { //nolint:nonamedreturns
	switch r {
	case Zip:
		return command.HWZip, "extract", nil
	case LHA:
		return command.Lha, "xf", nil
	case Arc:
		return command.Arc, "x", nil
	case Arj:
		return command.Zip7, "x", nil
	default:
		return "", "", fmt.Errorf("repair format %v: %w", r, ErrFormat)
	}
}

func (r Repair) lookups() error {
	const format = "cannot find %s exec: %w"
	switch r {
	case Zip:
		if _, err := command.Lookup(command.HWZip); err != nil {
			return fmt.Errorf(format, "hwzip", err)
		}
	case LHA:
		if _, err := command.Lookup(command.Lha); err != nil {
			return fmt.Errorf(format, "lha", err)
		}
	case Arc:
		if _, err := command.Lookup(command.Arc); err != nil {
			return fmt.Errorf(format, "arc", err)
		}
	case Arj:
		if _, err := command.Lookup(command.Zip7); err != nil {
			return fmt.Errorf(format, "7zz", err)
		}
	default:
	}
	return nil
}

func (r Repair) artifacts(ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor) ([]string, error) {
	const format = "repair artifacts %s: %w"
	if err := nils.Check(ctx, sl, exec); err != nil {
		return nil, fmt.Errorf(format, "check", err)
	}

	var files models.FileSlice
	var err error

	switch r {
	case Zip:
		files, err = fixzip.Files(ctx, exec)
	case LHA:
		files, err = fixlha.Files(ctx, exec)
	case Arc:
		files, err = fixarc.Files(ctx, exec)
	case Arj:
		files, err = fixarj.Files(ctx, exec)
	default:
		return nil, fmt.Errorf(format, "unsupported repair format "+r.String(), ErrFormat)
	}
	if err != nil {
		return nil, fmt.Errorf(format, r.String(), err)
	}

	sl.Info("Repair artifact", slog.String("era", "MS-DOS"),
		slog.String("format", r.String()), slog.Int("count", len(files)))

	artifacts := make([]string, 0, len(files))
	for _, f := range files {
		if f.UUID.Valid && f.UUID.String != "" {
			artifacts = append(artifacts, f.UUID.String)
		}
	}

	slices.Sort(artifacts)
	return slices.Clip(artifacts), nil
}

// Assets checks the file system directories for any invalid or unknown files.
// These specifically match the base filename against the UUID column in the database.
// When there is no matching UUID, the file is considered orphaned and these are moved
// to the orphaned directory without warning.
//
// There are no checks on the 3 directories that get scanned.
func (c *Config) Assets(ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor) error {
	return c.assetsScan(ctx, sl, exec, false)
}

func (c *Config) assetsScan(
	ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor, dryRun bool,
) error {
	const format = "repair assets %s: %w"
	if err := nils.Check(ctx, sl, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	start := time.Now()
	artifacts, count, err := allArtifactUUIDs(ctx, exec)
	if err != nil {
		return fmt.Errorf(format, "select all uuids", err)
	}

	const msg = "Repair assets"
	sl.Info(msg, slog.String("task", "Check UUID count"), slog.Int("result", count))

	dirs := [...]string{c.AbsDownload.String(), c.AbsPreview.String(), c.AbsThumbnail.String()}
	orphaned := dir.Directory(c.AbsOrphaned)

	var (
		wg          sync.WaitGroup
		totalChecks atomic.Int64
	)
	for _, targetDir := range dirs {
		if targetDir == "" {
			continue
		}
		wg.Go(func() {
			totalChecks.Add(scanAssetDir(ctx, sl, targetDir, artifacts, orphaned, dryRun))
		})
	}
	wg.Wait()

	sl.Info(msg,
		slog.String("task", "Time taken"), slog.Int64("checks", totalChecks.Load()),
		slog.Int("uuids", count), slog.Duration("time", time.Since(start).Round(time.Millisecond)),
	)
	return nil
}

func allArtifactUUIDs(ctx context.Context, exec boil.ContextExecutor) ([]string, int, error) {
	mods := []qm.QueryMod{qm.Select("uuid"), qm.WithDeleted()}
	files, err := models.Files(mods...).All(ctx, exec)
	if err != nil {
		return nil, 0, err
	}

	uuids := make([]string, 0, len(files))
	for _, f := range files {
		if f.UUID.Valid && f.UUID.String != "" {
			uuids = append(uuids, f.UUID.String)
		}
	}
	slices.Sort(uuids)
	return slices.Clip(uuids), len(files), nil
}

func scanAssetDir(
	ctx context.Context, sl *slog.Logger, root string, artifacts []string, orphaned dir.Directory, dryRun bool,
) int64 {
	var localCount int64
	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("repair assets walk path %s: %w", path, walkErr)
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf("walk path ctx error: %w", ctxErr)
		}
		if d.IsDir() {
			return nil
		}

		localCount++
		uid := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
		if _, found := slices.BinarySearch(artifacts, uid); !found {
			unknownAsset(sl, path, d.Name(), uid, orphaned, dryRun)
		}
		return nil
	})
	if walkErr != nil {
		sl.Error("Repair assets", slog.String("walk_directory", root), slog.Any("error", walkErr))
	}
	return localCount
}

// unknownAsset logs a warning message for an unknown asset file and moves it to the orphaned directory.
func unknownAsset(sl *slog.Logger, oldpath, name, uid string, orphaned dir.Directory, dryRun bool) {
	const msg = "unknown file"

	if sl == nil {
		sl = slog.Default()
	}

	if dryRun {
		sl.Warn("[dry-run] "+msg, slog.String("issue", "no matching artifact in the database; would move to orphaned"),
			slog.String("uuid", uid), slog.String("filename", name),
		)
		return
	}

	sl.Warn(msg, slog.String("issue", "no matching artifact in the database for the found file"),
		slog.String("uuid", uid), slog.String("filename", name),
	)

	// Format timestamp with sub-second precision to avoid collisions across concurrent goroutines
	now := time.Now().Format("2006-01-02_15-04-05.000000")
	dest := orphaned.Join(fmt.Sprintf("%s_%s_%s", now, uid, name))

	if err := helper.RenameCrossDevice(oldpath, dest); err != nil {
		sl.Error(msg,
			slog.String("issue", "could not move the file to the orphaned directory"),
			slog.String("source_path", oldpath),
			slog.String("destination_path", dest),
			slog.Any("error", err),
		)
	}
}

// assets checks the file system directories for any invalid or unknown files.
// If any are found, they are removed without warning (or only reported when dryRun is true).
func (c *Config) assets(ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor, dryRun bool) error {
	const format = "config assets repair %s: %w"
	if err := nils.Check(ctx, sl, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	backup := dir.Directory(c.AbsOrphaned)
	if dryRun {
		if err := backup.IsDir(); err != nil {
			return fmt.Errorf(format, "backup directory", err)
		}
	} else if err := backup.Check(sl); err != nil {
		return fmt.Errorf(format, "backup directory", err)
	}

	steps := [...]struct {
		name string
		fn   func() error
	}{
		{"images directories", func() error { return c.imageDirs(sl, dryRun) }},
		{"download directory", func() error {
			return downloadDir(sl, dir.Directory(c.AbsDownload), backup, dir.Directory(c.AbsExtra), dryRun)
		}},
		{"assets", func() error { return c.assetsScan(ctx, sl, exec, dryRun) }},
		{"archives", func() error { return c.repairArchives(ctx, sl, exec, dryRun) }},
		{"previews", func() error { return c.previews(ctx, sl, exec, dryRun) }},
		{"magic numbers", func() error { return c.magicNumbers(ctx, sl, exec, dryRun) }},
		{"textfiles", func() error { return c.textFiles(ctx, sl, exec, dryRun) }},
	}

	for _, step := range steps {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf(format, "execute", err)
		}

		if err := step.fn(); err != nil {
			return fmt.Errorf(format, step.name, err)
		}
	}

	return nil
}

// TextFiles checks the extra directory for any readme text files that are duplicates of the diz text files.
func (c *Config) TextFiles(ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor) error {
	return c.textFiles(ctx, sl, exec, false)
}

func (c *Config) textFiles(ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor, dryRun bool) error {
	const format = "repair text files %s: %w"
	if err := nils.Check(ctx, sl, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	uuids, err := model.UUID(ctx, exec)
	if err != nil {
		return fmt.Errorf(format, "model uuid", err)
	}

	extraDir := c.AbsExtra.String()
	dupes := 0

	const msg = "Fix textfile"
	for val := range slices.Values(uuids) {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf("%s: %w", msg, ctxErr)
		}
		if !val.UUID.Valid || val.UUID.String == "" {
			continue
		}

		basePath := filepath.Join(extraDir, val.UUID.String)
		diz := basePath + ".diz"
		txt := basePath + ".txt"
		if !isDuplicateTextPair(diz, txt) {
			continue
		}

		dupes++
		handleDuplicateText(sl, diz, txt, dryRun)
	}

	if dupes > 0 {
		sl.Info(msg, slog.String("duplicates", "Discovered text duplicates"), slog.Int("finds", dupes))
	}

	return nil
}

func isDuplicateTextPair(diz, txt string) bool {
	dizF, err := os.Stat(diz)
	if err != nil || dizF.Size() == 0 {
		return false
	}
	txtF, err := os.Stat(txt)
	if err != nil || txtF.Size() == 0 {
		return false
	}
	// match exact file sizes before performing hash checks
	if dizF.Size() != txtF.Size() {
		return false
	}
	// cryptographic hashes to confirm duplication
	dizSI, err := helper.StrongIntegrity(diz)
	if err != nil {
		return false
	}
	txtSI, err := helper.StrongIntegrity(txt)
	if err != nil {
		return false
	}
	return dizSI == txtSI
}

func handleDuplicateText(sl *slog.Logger, diz, txt string, dryRun bool) {
	const msg = "Fix textfile"
	if dryRun {
		target := filepath.Base(diz)
		if descriptor, err := isFileID(diz); err == nil && descriptor {
			target = filepath.Base(txt)
		}
		sl.Info("[dry-run] "+msg, slog.String("action", "would remove duplicate text: fileid == readme"),
			slog.String("filename", target),
		)
		return
	}

	dupe, err := Remove(diz, txt)
	if err != nil {
		sl.Error(msg,
			slog.String("problem", "Cannot remove file duplicates"),
			slog.String(extensions.FileID, diz), slog.String("readme_text", txt),
			slog.Any("error", err),
		)
		return
	}

	sl.Info(msg, slog.String("success", "Removed duplicate text: fileid == readme"),
		slog.String("filename", dupe),
	)
}

// Remove deletes either the named diz or txt file from an identical pair.
// The file deleted depends on whether the pair is identified as a FILE_ID.DIZ or a longer text file.
// If successful, the base name of the removed file is returned.
func Remove(diz, txt string) (string, error) {
	const format = "remove diz text %s: %w"
	descriptor, err := isFileID(diz)
	if err != nil {
		return "", fmt.Errorf(format, "inspect", err)
	}

	if !descriptor {
		if err := os.Remove(diz); err != nil {
			return "", fmt.Errorf(format, "remove diz", err)
		}
		return filepath.Base(diz), nil
	}

	if err := os.Remove(txt); err != nil {
		return "", fmt.Errorf(format, "remove readme", err)
	}

	return filepath.Base(txt), nil
}

func isFileID(path string) (bool, error) {
	r, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("is file id: %w", err)
	}
	defer func() { _ = r.Close() }()

	return FileID(r), nil
}

// FileID will return true if there are less than 10 lines of text
// and the maximum width of each line is no more than 45 characters.
// This is not a guarantee of a [FILE_ID.DIZ] but it is true for many situations.
//
// [FILE_ID.DIZ]: http://www.textfiles.com/computers/fileid.txt
func FileID(r io.Reader) bool {
	if r == nil {
		return false
	}

	scanner := bufio.NewScanner(r)

	const (
		maximumLines = 10
		maximumWidth = 45
	)

	lines := 0
	for scanner.Scan() {
		lines++
		if lines > maximumLines {
			return false
		}

		if utf8.RuneCountInString(scanner.Text()) > maximumWidth {
			return false
		}
	}

	if err := scanner.Err(); err != nil {
		return false
	}
	return lines > 0
}

// MagicNumbers checks the magic numbers of the artifacts and replaces any missing or
// legacy values with the current method of detection. Previous detection methods were
// done using the `file` command line utility, which is a bit too verbose for our needs.
func (c *Config) MagicNumbers(ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor) error {
	return c.magicNumbers(ctx, sl, exec, false)
}

func (c *Config) magicNumbers(ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor, dryRun bool) error {
	const format = "config magic numbers %s: %w"
	if err := nils.Check(ctx, sl, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	tick := time.Now()

	magics, err := model.OnlyMagicErrs(ctx, exec, false)
	if err != nil {
		return fmt.Errorf(format, "models file slice", err)
	}

	const msg = "magic numbers"
	const large = 1000
	if len(magics) > large && sl != nil {
		sl.Warn(msg+", there are a large number of artifacts to check, it could take a while",
			slog.Int("task_count", len(magics)))
	}

	downloadDir := c.AbsDownload.String()
	count := 0

	for val := range slices.Values(magics) {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf(format, "context", ctxErr)
		}

		if !val.UUID.Valid || val.UUID.String == "" {
			continue
		}

		name := filepath.Join(downloadDir, val.UUID.String)
		magicTitle, err := inspectMagic(name)
		if err != nil {
			continue
		}
		if dryRun {
			sl.Info("[dry-run] would update magic number",
				slog.Int64("id", val.ID), slog.String("magic", magicTitle))
			count++
			continue
		}
		if err := model.UpdateMagic(ctx, exec, val.ID, magicTitle); err != nil {
			if sl != nil {
				sl.Error(msg+" failed to update magic number", slog.Int64("id", val.ID), slog.Any("error", err))
			}
			continue
		}

		count++
	}

	if count > 0 && sl != nil {
		label := msg
		if dryRun {
			label = "[dry-run] " + msg
		}
		sl.Info(label, slog.Int("values_update", count),
			slog.Duration("time", time.Since(tick).Round(time.Millisecond)))
	}

	return nil
}

// inspectMagic safely opens a file, extracts its magic number title, and guarantees handle closure.
func inspectMagic(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("inspect magic: %w", err)
	}
	defer file.Close()

	magic := magicnumber.Find(file)
	title := magic.Title()
	if magic != magicnumber.ANSIEscapeText {
		return title, nil
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return title, nil
	}
	if ok, err := ansimate.Detect(file); err == nil && ok {
		return "ANSImation", nil
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return title, nil
	}
	if ok, err := ansimate.DetectBlink(file); err == nil && ok {
		return title + " with blinking", nil
	}
	return title, nil
}

// Previews checks the preview directory for any unnecessary preview images such as textfile artifacts.
func (c *Config) Previews(ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor) error {
	return c.previews(ctx, sl, exec, false)
}

func (c *Config) previews(ctx context.Context, sl *slog.Logger, exec boil.ContextExecutor, dryRun bool) error {
	const format = "config previews %s: %w"
	if err := nils.Check(ctx, sl, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	artifacts, err := model.OnlyTexts(ctx, exec)
	if err != nil {
		return fmt.Errorf(format, "models file slice", err)
	}

	var count, totals int64
	previewDir := c.AbsPreview.String()
	exts := [...]string{".png", ".webp"}

	for val := range slices.Values(artifacts) {
		// 1. Respect cancellation during disk cleanup
		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf(format, "context", ctxErr)
		}

		if !val.UUID.Valid || val.UUID.String == "" {
			continue
		}

		basePath := filepath.Join(previewDir, val.UUID.String)
		for _, ext := range exts {
			file := basePath + ext
			st, err := os.Stat(file)
			if err != nil {
				continue
			}

			if !dryRun {
				if err := os.Remove(file); err != nil {
					if sl != nil {
						sl.Error("previews could not remove unwanted file",
							slog.String("path", file), slog.Any("error", err))
					}
					continue
				}
			}

			count++
			totals += st.Size()
		}
	}

	if count == 0 {
		return nil
	}

	if sl != nil {
		msg := "Erased textfile previews"
		if dryRun {
			msg = "[dry-run] Would erase textfile previews"
		}
		sl.Info(msg, slog.Int64("count", count), slog.String("sum", helper.ByteCountFloat(totals)))
	}

	return nil
}

// ImageDirs checks the image directories for any invalid or unknown files.
func (c *Config) ImageDirs(sl *slog.Logger) error {
	return c.imageDirs(sl, false)
}

func (c *Config) imageDirs(sl *slog.Logger, dryRun bool) error {
	const msg = "image directories"
	if err := nils.Check(sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	backup := dir.Directory(c.AbsOrphaned.String())
	targets := [...]struct{ label, path string }{
		{"preview", c.AbsPreview.String()},
		{"thumb", c.AbsThumbnail.String()},
	}
	backupChecked := false
	for _, target := range targets {
		if _, err := os.Stat(target.path); err != nil {
			continue
		}
		if !backupChecked {
			if err := checkBackupDir(sl, backup, dryRun); err != nil {
				return fmt.Errorf("%s: %w: %s", msg, err, backup)
			}
			backupChecked = true
		}
		count, err := walkImageDir(sl, target.path, backup, dryRun)
		if err != nil {
			return fmt.Errorf("%s walk directory %w: %s", msg, err, target.path)
		}
		containsInfo(sl, target.label, count)
	}
	return nil
}

func checkBackupDir(sl *slog.Logger, backup dir.Directory, dryRun bool) error {
	const format = "check backup dir: %w"
	if dryRun {
		err := backup.IsDir()
		if err != nil {
			return fmt.Errorf(format, err)
		}
		return nil
	}
	err := backup.Check(sl)
	if err != nil {
		return fmt.Errorf(format, err)
	}
	return nil
}

func walkImageDir(sl *slog.Logger, root string, backup dir.Directory, dryRun bool) (int, error) {
	const format = "walk image dir: %s: %w"
	count := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf(format, path, err)
		}
		name := d.Name()
		if d.IsDir() {
			return removeDir(sl, name, path, root, dryRun)
		}
		if ext := filepath.Ext(name); ext == ".png" || ext == ".webp" {
			count++
		}
		return removeImage(sl, name, path, backup, dryRun)
	})
	if err != nil {
		return count, fmt.Errorf(format, "file path", err)
	}
	return count, nil
}

// containsInfo logs the number of files found in the directory.
func containsInfo(sl *slog.Logger, name string, count int) {
	const msg = "contains info"
	if err := nils.Check(sl); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	var s string
	switch strings.ToLower(name) {
	case "thumb":
		s = " thumbnails"
	case "preview":
		s = " previews"
	case "downloads":
		s = " artifact downloads"
	default:
		s = name
	}
	if MinimumFiles > count {
		sl.Warn("File"+s,
			slog.String("issue", "The directory contains too few files"),
			slog.Int("count", count), slog.Int("minimum", MinimumFiles))
		return
	}
	sl.Info("File"+s,
		slog.Int("count", count))
}

// DownloadDir checks the download directory for any invalid or unknown files.
func DownloadDir(sl *slog.Logger, src, dest, extra dir.Directory) error {
	return downloadDir(sl, src, dest, extra, false)
}

func downloadDir(sl *slog.Logger, src, dest, extra dir.Directory, dryRun bool) error {
	const msg = "download directory"
	if err := nils.Check(sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	checkDir := func(d dir.Directory) error {
		if dryRun {
			return d.IsDir()
		}
		return d.Check(sl)
	}
	if err := checkDir(src); err != nil {
		return fmt.Errorf("%s %w: %s", msg, err, src)
	}
	if err := checkDir(dest); err != nil {
		return fmt.Errorf("%s %w: %s", msg, err, dest)
	}
	if err := checkDir(extra); err != nil {
		return fmt.Errorf("%s %w: %s", msg, err, extra)
	}
	count := 0
	err := filepath.WalkDir(src.Path(), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%s walk path %w: %s", msg, err, path)
		}
		name := d.Name()
		if d.IsDir() {
			return removeDir(sl, name, path, src.Path(), dryRun)
		}
		if err = removeDownload(sl, name, path, dest, extra, dryRun); err != nil {
			return fmt.Errorf("%s remove download: %w", msg, err)
		}
		if filepath.Ext(name) == "" {
			count++
		}
		return renameDownload(sl, name, path, dryRun)
	})
	if err != nil {
		return fmt.Errorf("%s walk directory %w: %s", msg, err, src.Path())
	}
	containsInfo(sl, "downloads", count)
	return nil
}

// RenameDownload rename the download file if the basename uses an invalid coldfusion uuid.
func RenameDownload(sl *slog.Logger, basename, absPath string) error {
	return renameDownload(sl, basename, absPath, false)
}

func renameDownload(sl *slog.Logger, basename, absPath string, dryRun bool) error {
	const msg = "rename download"
	if err := nils.Check(sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	if basename == "" || absPath == "" {
		return fmt.Errorf("%s %w: %s %s", msg, ErrNoPath, basename, absPath)
	}
	ext := filepath.Ext(basename)
	rawname, found := strings.CutSuffix(basename, ext)
	if !found {
		return nil
	}
	const cflen = len(cfid) // coldfusion uuid length
	if len(rawname) != cflen {
		return nil
	}
	newname, _ := helper.CfUUID(rawname)
	if _, err := uuid.Parse(newname); err != nil {
		return fmt.Errorf("%s uuid validate %q: %w", msg, newname, err)
	}
	dir := filepath.Dir(absPath)
	oldpath := filepath.Join(dir, basename)
	newpath := filepath.Join(dir, newname+ext)
	renameFile(sl, oldpath, "renamed invalid cfid", newpath, dryRun)
	return nil
}

// RemoveDir check the directory for invalid names.
// If any are found, they are printed to stderr.
// Any directory that matches the name ".stfolder" is removed.
func RemoveDir(sl *slog.Logger, name, path, root string) error {
	return removeDir(sl, name, path, root, false)
}

func removeDir(sl *slog.Logger, name, path, root string, dryRun bool) error {
	const msg = "repair remove directory"
	if err := nils.Check(sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	if name == "" || path == "" || root == "" {
		return fmt.Errorf("%s: %w: %s %s %s", msg, ErrNoPath, name, path, root)
	}
	rootDir := filepath.Base(root)
	switch name {
	case rootDir:
		return nil
	case syncthing:
		if dryRun {
			sl.Info("[dry-run] "+msg, slog.String("would_remove", path))
			return filepath.SkipDir
		}
		if err := os.RemoveAll(path); err != nil {
			sl.Error(msg, slog.Any("error", err))
		} else {
			sl.Info(msg, slog.String("removed", path))
		}
		return filepath.SkipDir
	default:
		sl.Error(msg, slog.String("unknown_path", path))
		return nil
	}
}

// RemoveDownload checks the download files for invalid names and extensions.
// If any are found, they are removed without warning.
// Basename must be the name of the file with a valid file extension.
//
// Valid file extensions are none, .chiptune, .txt, and .zip.
func RemoveDownload(sl *slog.Logger, basename, path string, backup, extra dir.Directory) error {
	return removeDownload(sl, basename, path, backup, extra, false)
}

func removeDownload(sl *slog.Logger, basename, path string, backup, extra dir.Directory, dryRun bool) error {
	const msg = "remove download"
	if err := nils.Check(sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	if basename == "" || path == "" {
		return fmt.Errorf("%s %w: %s %s", msg, ErrNoPath, basename, path)
	}
	const filedownload = ""
	ext := filepath.Ext(basename)
	switch ext {
	case filedownload:
		return nil
	case ".txt", ".zip", ".chiptune":
		renameFile(sl, path, "rename valid ext", extra.Join(basename), dryRun)
	default:
		removeFile(sl, basename, "remove invalid ext", path, backup, dryRun)
	}
	return nil
}

// RemoveImage checks the image files for invalid names and extensions.
// If any are found, they are moved to the destDir without warning.
// Basename must be the name of the file with a valid file extension.
//
// Valid file extensions are .png and .webp, and basename must be a
// valid uuid or cfid with the correct length.
func RemoveImage(sl *slog.Logger, basename, path string, backup dir.Directory) error {
	return removeImage(sl, basename, path, backup, false)
}

func removeImage(sl *slog.Logger, basename, path string, backup dir.Directory, dryRun bool) error {
	const msg = "remove image"
	if err := nils.Check(sl); err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	if basename == "" || path == "" || backup == "" {
		return fmt.Errorf("%s %w: %s %s %s", msg, ErrNoPath, basename, path, backup)
	}
	const (
		png   = ".png"    // png file extension
		webp  = ".webp"   // webp file extension
		cflen = len(cfid) // coldfusion uuid length
	)
	ext := filepath.Ext(basename)
	if filename, found := strings.CutSuffix(basename, ext); found {
		if len(filename) == cflen {
			filename, _ = helper.CfUUID(filename)
			newpath := filepath.Dir(path)
			switch ext {
			case png, webp:
				renameFile(sl, path, "rename cfid "+ext, filepath.Join(newpath, filename+ext), dryRun)
				return nil
			}
		}
		if _, err := uuid.Parse(filename); err != nil {
			removeFile(sl, basename, "remove invalid uuid image", path, backup, dryRun)
			return nil //nolint:nilerr
		}
	}
	switch ext {
	case png, webp:
		return nil
	default:
		removeFile(sl, basename, "remove invalid uuid ext", path, backup, dryRun)
	}
	return nil
}

// removeFile moves the file to backup without warning, or only logs when dryRun is true.
func removeFile(sl *slog.Logger, name, info, path string, backup dir.Directory, dryRun bool) {
	const msg = "Remove file"
	if sl == nil {
		sl = logs.Discard()
	}
	if dryRun {
		sl.Info("[dry-run] "+msg, slog.String("name", name), slog.String("detail", info))
		return
	}
	sl.Info(msg, slog.String("name", name), slog.String("detail", info))
	defer func() {
		now := time.Now().Format("2006-01-02_15-04-05.000000")
		newpath := backup.Join(fmt.Sprintf("%s_%s", now, name))
		err := helper.RenameCrossDevice(path, newpath)
		if err != nil {
			sl.Error(msg, slog.String("name", name), slog.String("detail", info), slog.Any("error", err))
		}
	}()
}

// renameFile moves or renames the file without warning, or only logs when dryRun is true.
func renameFile(sl *slog.Logger, oldpath, info, newpath string, dryRun bool) {
	const msg = "Rename or move file"
	if sl == nil {
		sl = logs.Discard()
	}
	if dryRun {
		sl.Info("[dry-run] "+msg, slog.String("original_path", oldpath),
			slog.String("new_path", newpath), slog.String("detail", info))
		return
	}
	sl.Info(msg, slog.String("original_path", oldpath), slog.String("new_path", newpath), slog.String("detail", info))
	defer func() {
		if err := helper.RenameCrossDevice(oldpath, newpath); err != nil {
			sl.Error(msg, slog.String("original_path", oldpath), slog.String("new_path", newpath),
				slog.String("detail", info), slog.Any("error", err))
		}
	}()
}

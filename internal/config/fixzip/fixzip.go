package fixzip

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Defacto2/server/internal/archive/pkzip"
	"github.com/Defacto2/server/internal/archive/rezip"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Check returns the UUID of the zipped file if it requires re-archiving because it uses a
// legacy compression method that is not supported by Go or JS libraries.
//
// Check UUID named files are moved to the extra directory and are given a .zip extension.
func Check(ctx context.Context, path, extra string, d fs.DirEntry, artifacts ...string) string {
	if d.IsDir() {
		return ""
	}
	if ext := filepath.Ext(strings.ToLower(d.Name())); ext != ".zip" && ext != "" {
		return ""
	}
	uid := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
	if _, found := slices.BinarySearch(artifacts, uid); !found {
		return ""
	}
	logger := helper.Logger(ctx)
	extraZip := filepath.Join(extra, uid+".zip")
	if f, err := os.Stat(extraZip); err == nil && !f.IsDir() {
		return ""
	}
	methods, err := pkzip.Methods(path)
	if err != nil {
		logger.Errorf("%s: %s", err, path)
		return ""
	}
	for _, method := range methods {
		if !method.Zip() {
			return uid
		}
	}
	return ""
}

// Compress uses the [hwzip application] by Hans Wennborg to extract zip archives
// using antiquated compression methods that are not supported by Go, JS or other
// Linux utilities. The extracted files are then re-archived using Go and moved
// to the extra directory with a .zip extension.
//
// [hwzip application]: https://www.hanshq.net/zip.html
func Compress(ctx context.Context, path, extra, uid string) error {
	logger := helper.Logger(ctx)
	tmp, err := os.MkdirTemp(os.TempDir(), "defacto2-fixzip-")
	if err != nil {
		return fmt.Errorf("fixzip compress mkdir temp %w: %s", err, path)
	}
	defer os.RemoveAll(tmp)

	const extractArg = "extract"
	cmd := exec.Command(command.HWZip, extractArg, path)
	cmd.Dir = tmp
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("fixzip compress run %w: %s", err, path)
	}

	c, err := helper.Count(tmp)
	if err != nil {
		return fmt.Errorf("fixzip compress tmp count %w: %s", err, tmp)
	}
	logger.Infof("Rezipped %d files for %s found in: %s", c, uid, tmp)
	_, err = os.Stat(tmp)
	if err != nil {
		return fmt.Errorf("fixzip compress tmp stat %w: %s", err, tmp)
	}

	basename := uid + ".zip"
	tmpZip := filepath.Join(os.TempDir(), basename)
	if written, err := rezip.CompressDir(tmp, tmpZip); err != nil {
		return fmt.Errorf("fixzip compress dir %w: %s", err, tmp)
	} else if written == 0 {
		return nil
	}

	finalZip := filepath.Join(extra, basename)
	if err = helper.RenameCrossDevice(tmpZip, finalZip); err != nil {
		defer os.RemoveAll(tmpZip)
		return fmt.Errorf("fixzip compress rename %w: %s", err, tmpZip)
	}

	st, err := os.Stat(finalZip)
	if err != nil {
		return fmt.Errorf("fixzip compress zip stat %w: %s", err, finalZip)
	}
	logger.Infof("Extra deflated zipfile created %d bytes: %s", st.Size(), finalZip)
	return nil
}

// Files returns all the DOS platform artifacts using a .zip extension filename.
func Files(ctx context.Context, ce boil.ContextExecutor) (models.FileSlice, error) {
	mods := []qm.QueryMod{}
	mods = append(mods, qm.Select("uuid"))
	mods = append(mods, qm.Where("platform = ?", tags.DOS.String()))
	mods = append(mods, qm.Where("filename ILIKE ?", "%.zip"))
	mods = append(mods, qm.WithDeleted())
	files, err := models.Files(mods...).All(ctx, ce)
	if err != nil {
		return nil, fmt.Errorf("fixzip models files: %w", err)
	}
	return files, nil
}

// Invalid returns true if the zip file fails the hwzip list command.
// The path is the path to the zip file.
func Invalid(ctx context.Context, path string) bool {
	logger := helper.Logger(ctx)
	z, err := exec.Command(command.HWZip, "list", path).Output()
	if err != nil {
		logger.Errorf("fixzip invalid %s: %s", err, path)
		return true
	}
	if !strings.Contains(string(z), "Failed to parse ") {
		return true
	}
	return false
}

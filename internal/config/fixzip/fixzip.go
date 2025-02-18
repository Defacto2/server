package fixzip

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Defacto2/archive/pkzip"
	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

var ErrNoExecutor = errors.New("no context executor")

// Check returns the UUID of the name zipped file if it requires re-archiving because it uses a
// legacy compression method that is not supported by Go or JS libraries.
//
// Check UUID named files are moved to the extra directory and are given a .zip extension.
func Check(ctx context.Context, name string, extra dir.Directory, d fs.DirEntry, artifacts ...string) string {
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
	extraZip := extra.Join(uid + ".zip")
	if f, err := os.Stat(extraZip); err == nil && !f.IsDir() {
		return ""
	}
	methods, err := pkzip.Methods(name)
	if err != nil {
		logger.Errorf("%s: %s", err, name)
		return ""
	}
	for method := range slices.Values(methods) {
		if !method.Zip() {
			return uid
		}
	}
	return ""
}

// Files returns all the DOS platform artifacts using a .zip extension filename.
func Files(ctx context.Context, exec boil.ContextExecutor) (models.FileSlice, error) {
	if exec == nil {
		return nil, fmt.Errorf("config fixzip files %w", ErrNoExecutor)
	}
	mods := []qm.QueryMod{}
	mods = append(mods, qm.Select("uuid"))
	mods = append(mods, qm.Where("platform = ?", tags.DOS.String()))
	mods = append(mods, qm.Where("filename ILIKE ?", "%.zip"))
	mods = append(mods, qm.WithDeleted())
	files, err := models.Files(mods...).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("fixzip models files: %w", err)
	}
	return files, nil
}

// Invalid returns true if the zip file fails the hwzip list command.
// The path is the path to the zip file.
func Invalid(ctx context.Context, path string) bool {
	logger := helper.Logger(ctx)
	const name = command.HWZip
	cmd := exec.Command(name, "list", path)
	b, err := cmd.Output()
	if err != nil {
		logger.Errorf("fixzip invalid %s: %s", err, path)
		return true
	}
	if !strings.Contains(string(b), "Failed to parse ") {
		return true
	}
	return false
}

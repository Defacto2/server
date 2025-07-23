// Package fixarc checks for redundant SAE ARC files that require re-archiving.
package fixarc

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
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

var ErrNoExecutor = errors.New("no context executor")

// Check returns the UUID of the named zipped file if it requires re-archiving because it uses a
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

// Files returns all the DOS platform artifacts using a .arc extension filename.
func Files(ctx context.Context, exec boil.ContextExecutor) (models.FileSlice, error) {
	if exec == nil {
		return nil, fmt.Errorf("config fixarc files %w", ErrNoExecutor)
	}
	mods := []qm.QueryMod{}
	mods = append(mods, qm.Select("uuid"))
	mods = append(mods, qm.Where("platform = ?", tags.DOS.String()))
	mods = append(mods, qm.Where("filename ILIKE ?", "%.arc"))
	mods = append(mods, qm.WithDeleted())
	files, err := models.Files(mods...).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("fixarc models files: %w", err)
	}
	return files, nil
}

// Invalid returns true if the arc file fails the arc test command.
// The path is the path to the arc archive file.
func Invalid(ctx context.Context, path string) bool {
	logger := helper.Logger(ctx)
	const name = command.Arc
	cmd := exec.Command(name, "t", path)
	b, err := cmd.Output()
	if err != nil {
		logger.Errorf("fixarc invalid %s: %s", err, path)
		return true
	}
	return strings.Contains(string(b), "is not an archive")
}

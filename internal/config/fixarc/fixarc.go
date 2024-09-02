package fixarc

import (
	"context"
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

// Files returns all the DOS platform artifacts using a .arc extension filename.
func Files(ctx context.Context, ce boil.ContextExecutor) (models.FileSlice, error) {
	mods := []qm.QueryMod{}
	mods = append(mods, qm.Select("uuid"))
	mods = append(mods, qm.Where("platform = ?", tags.DOS.String()))
	mods = append(mods, qm.Where("filename ILIKE ?", "%.arc"))
	mods = append(mods, qm.WithDeleted())
	files, err := models.Files(mods...).All(ctx, ce)
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

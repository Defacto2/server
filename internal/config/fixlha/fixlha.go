// Package fixlha checks for legacy LHA files that require re-archiving.
package fixlha

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/panics"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

// Check returns the UUID of the zipped file if it requires re-archiving because it uses a
// legacy compression method that is not supported by Go or JS libraries.
//
// Check UUID named files are moved to the extra directory and are given a .zip extension.
func Check(extra dir.Directory, d fs.DirEntry, artifacts ...string) string {
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
	extraZip := extra.Join(uid + ".zip")
	if f, err := os.Stat(extraZip); err == nil && !f.IsDir() {
		return ""
	}
	return uid
}

// Files returns all the DOS platform artifacts using a .zip extension filename.
func Files(ctx context.Context, exec boil.ContextExecutor) (models.FileSlice, error) {
	const msg = "fix lha files"
	if err := panics.ContextB(ctx, exec); err != nil {
		return nil, fmt.Errorf("%s: %w", msg, err)
	}
	const size = 5
	mods := make([]qm.QueryMod, 0, size)
	mods = append(mods, qm.Select("uuid"))
	mods = append(mods, qm.Where("platform = ?", tags.DOS.String()))
	mods = append(mods, qm.Where("filename ILIKE ?", "%.lha"))
	mods = append(mods, qm.Or("filename ILIKE ?", "%.lzh"))
	mods = append(mods, qm.WithDeleted())
	files, err := models.Files(mods...).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("%s models: %w", msg, err)
	}
	return files, nil
}

// Invalid returns true if the lha file fails the lha test command.
// The path is the path to the lha archive file.
func Invalid(sl *slog.Logger, path string) bool {
	const msg = "lha fixer is invalid"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}
	const name = command.Lha
	cmd := exec.Command(name, "t", path)
	b, err := cmd.Output()
	if err != nil {
		sl.Error(msg,
			slog.String("lha file path", path),
			slog.Any("error", err))
		return true
	}
	return len(bytes.TrimSpace(b)) == 0
}

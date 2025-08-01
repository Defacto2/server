// Package fixarc checks for redundant SAE ARC files that require re-archiving.
package fixarc

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Defacto2/archive/pkzip"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/dir"
	"github.com/Defacto2/server/internal/panics"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

// Check returns the UUID of the named zipped file if it requires re-archiving because it uses a
// legacy compression method that is not supported by Go or JS libraries.
//
// Check UUID named files are moved to the extra directory and are given a .zip extension.
func Check(sl *slog.Logger, name string, extra dir.Directory, d fs.DirEntry, artifacts ...string) string {
	const msg = "fix arc check"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}
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
	methods, err := pkzip.Methods(name)
	if err != nil {
		sl.Error(msg, slog.String("named file", name), slog.Any("error", err))
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
	const msg = "fix arc files"
	if err := panics.ContextB(ctx, exec); err != nil {
		return nil, fmt.Errorf("%s: %w", msg, err)
	}
	mods := []qm.QueryMod{}
	mods = append(mods, qm.Select("uuid"))
	mods = append(mods, qm.Where("platform = ?", tags.DOS.String()))
	mods = append(mods, qm.Where("filename ILIKE ?", "%.arc"))
	mods = append(mods, qm.WithDeleted())
	files, err := models.Files(mods...).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("%s models: %w", msg, err)
	}
	return files, nil
}

// Invalid returns true if the arc file fails the arc test command.
// The path is the path to the arc archive file.
func Invalid(sl *slog.Logger, path string) bool {
	const msg = "arc fixer is invalid"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}
	const name = command.Arc
	cmd := exec.Command(name, "t", path)
	b, err := cmd.Output()
	if err != nil {
		sl.Error(msg,
			slog.String("arc file path", path),
			slog.Any("error", err))
		return true
	}
	return strings.Contains(string(b), "is not an archive")
}

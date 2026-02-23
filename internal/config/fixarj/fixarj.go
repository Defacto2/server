// Package fixarj checks for legacy ARJ files that require re-archiving.
package fixarj

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
	"time"

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
	ext := filepath.Ext(d.Name())
	if strings.ToLower(ext) != ".zip" && ext != "" {
		return ""
	}
	uid := strings.TrimSuffix(d.Name(), ext)
	if _, found := slices.BinarySearch(artifacts, uid); !found {
		return ""
	}
	extraZip := extra.Join(uid + ".zip")
	if _, err := os.Stat(extraZip); err == nil {
		return ""
	}
	return uid
}

// Files returns all the DOS platform artifacts using a .zip extension filename.
func Files(ctx context.Context, exec boil.ContextExecutor) (models.FileSlice, error) {
	const msg = "fix arj files"
	if err := panics.ContextB(ctx, exec); err != nil {
		return nil, fmt.Errorf("%s: %w", msg, err)
	}
	const size = 4
	mods := make([]qm.QueryMod, 0, size)
	mods = append(mods,
		qm.Select("uuid"),
		qm.Where("platform = ?", tags.DOS.String()),
		qm.Where("filename ILIKE ?", "%.arj"),
		qm.WithDeleted(),
	)
	files, err := models.Files(mods...).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("%s models: %w", msg, err)
	}
	return files, nil
}

// Invalid returns true if the arj file fails the 7zz list command.
// The path is the path to the arj archive file.
func Invalid(sl *slog.Logger, path string) bool {
	const msg = "arj fixer is invalid"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}
	const arjTimeout = 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), arjTimeout)
	defer cancel()

	// use 7-ZIP to test and extract the .arj file.
	cmd := exec.CommandContext(ctx, command.Zip7, "t", path)
	b, err := cmd.CombinedOutput()
	if err != nil {
		sl.Error(msg,
			slog.String("command in use", command.Zip7),
			slog.String("arj file path", path),
			slog.Any("error", err))
		return true
	}
	return !bytes.Contains(b, []byte("Everything is Ok"))
}

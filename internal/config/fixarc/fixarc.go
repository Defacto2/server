// Package fixarc checks for redundant SEA ARC files that require re-archiving.
package fixarc

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"

	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

// Files returns all the DOS platform artifacts using a .arc extension filename.
func Files(ctx context.Context, exec boil.ContextExecutor) (models.FileSlice, error) {
	const format = "fix arc files %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return models.FileSlice{}, fmt.Errorf(format, "check", err)
	}

	const size = 4
	mods := make([]qm.QueryMod, 0, size)
	mods = append(mods, qm.Select("uuid"))
	mods = append(mods, qm.Where("platform = ?", tags.DOS.String()))
	mods = append(mods, qm.Where("filename ILIKE ?", "%.arc"))
	mods = append(mods, qm.WithDeleted())

	files, err := models.Files(mods...).All(ctx, exec)
	if err != nil {
		return models.FileSlice{}, fmt.Errorf(format, "all", err)
	}

	return files, nil
}

// Invalid returns true if the arc file fails the arc test command.
// The path is the path to the arc archive file.
func Invalid(ctx context.Context, sl *slog.Logger, path string) bool {
	const msg = "arc fixer is invalid"
	if err := nils.Check(ctx, sl); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}

	const arcTimeout = 10 * time.Second
	subCtx, cancel := context.WithTimeout(ctx, arcTimeout)
	defer cancel()

	cmd := exec.CommandContext(subCtx, command.Arc, "t", path) //nolint:gosec // path is okay
	b, err := cmd.CombinedOutput()
	if err != nil {
		sl.Error(msg,
			slog.String("command", command.Arc), slog.String("path", path), slog.Any("error", err))
		return true
	}

	return strings.Contains(string(b), "is not an archive")
}

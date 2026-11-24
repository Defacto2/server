// Package fixzip checks for legacy ZIP files with out-of-date compression that require re-archiving.
package fixzip

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"slices"
	"strings"

	"github.com/Defacto2/archive/pkzip"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

// Legacy returns true if the zip file at path uses a legacy compression method
// that is not supported by standard Go or JS DEFLATE zip readers.
func Legacy(sl *slog.Logger, path string) bool {
	const msg = "fix hwzip check"
	if err := nils.Check(sl); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	methods, err := pkzip.Methods(path)
	if err != nil {
		sl.Error(msg,
			slog.String("filename", path), slog.Any("error", err))
		return false
	}
	for method := range slices.Values(methods) {
		if !method.Zip() {
			return true
		}
	}
	return false
}

// Files returns all the DOS platform artifacts using a .zip extension filename.
func Files(ctx context.Context, exec boil.ContextExecutor) (models.FileSlice, error) {
	const msg = "fix hwzip files"
	if err := nils.Check(ctx, exec); err != nil {
		return nil, fmt.Errorf("%s: %w", msg, err)
	}
	const size = 4
	mods := make([]qm.QueryMod, 0, size)
	mods = append(
		mods,
		qm.Select("uuid"),
		qm.Where("platform = ?", tags.DOS.String()),
		qm.Where("filename ILIKE ?", "%.zip"),
		qm.WithDeleted(),
	)
	files, err := models.Files(mods...).All(ctx, exec)
	if err != nil {
		return nil, fmt.Errorf("%s models: %w", msg, err)
	}
	return files, nil
}

// Invalid returns true if the zip file fails the hwzip list command.
// The path is the path to the zip file.
func Invalid(ctx context.Context, sl *slog.Logger, path string) bool {
	const msg = "hwzip fixer is invalid"
	if err := nils.Check(ctx, sl); err != nil {
		panic(fmt.Errorf("%s: %w", msg, err))
	}
	const name = command.HWZip
	cmd := exec.CommandContext(ctx, name, "list", path)
	b, err := cmd.CombinedOutput()
	if err != nil {
		sl.Error(msg,
			slog.String("command", name), slog.String("path", path), slog.Any("error", err))
		return true
	}
	if strings.Contains(string(b), "Failed to parse ") {
		return true
	}
	return false
}

package model

import (
	"context"
	"fmt"

	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

// DeleteOne permanently removes a single file record from the database by its ID,
// including records that have been soft-deleted.
func DeleteOne(ctx context.Context, exec boil.ContextExecutor, key int64) error {
	const format = "delete one record %d: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, key, err)
	}

	if key < 1 {
		return fmt.Errorf(format, key, ErrBadIDInt)
	}

	mods := models.FileWhere.ID.EQ(key)
	_, err := models.Files(mods, qm.WithDeleted()).DeleteAll(ctx, exec, true)
	if err != nil {
		return fmt.Errorf(format, key, err)
	}

	return nil
}

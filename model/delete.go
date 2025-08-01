package model

import (
	"context"
	"fmt"

	"github.com/Defacto2/server/internal/panics"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

// DeleteOne retrieves a single file record from the database using the record key.
// This function can return records that have been marked as deleted.
func DeleteOne(ctx context.Context, exec boil.ContextExecutor, key int64) error {
	const msg = "delete one"
	if panics.BoilExec(exec) {
		return fmt.Errorf("%s: %w", msg, panics.ErrNoBoil)
	}
	if key < 1 {
		return fmt.Errorf("%s key value %d: %w", msg, key, ErrKey)
	}
	mods := models.FileWhere.ID.EQ(key)
	_, err := models.Files(mods, qm.WithDeleted()).DeleteAll(ctx, exec, true)
	if err != nil {
		return fmt.Errorf("%s record %d: %w", msg, key, err)
	}
	return nil
}

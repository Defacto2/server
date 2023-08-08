package model

// This file is the custom software category for the HTML3 template.

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/model/modext"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Softs contain statistics for releases that could be considered software.
type Softs struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of releases that could be considered as digital or pixel art.
func (s *Softs) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if s.Bytes > 0 && s.Count > 0 {
		return nil
	}
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		SoftwareExpr(),
		qm.From(From)).Bind(ctx, db, s)
}

// SoftwareExpr is a the query mod expression for software files.
func SoftwareExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(modext.PJava()),
		qm.Or2(models.FileWhere.Platform.EQ(modext.PLinux())),
		qm.Or2(models.FileWhere.Platform.EQ(modext.PDos())),
		qm.Or2(models.FileWhere.Platform.EQ(modext.PScript())),
		qm.Or2(models.FileWhere.Platform.EQ(modext.PWindows())),
	)
}

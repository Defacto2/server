package model

// Package html3_software.go contains the database queries the HTML3 software category.

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/model/expr"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Softs contain statistics for releases that could be considered software.
type Softs struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

func (s *Softs) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if s.Bytes > 0 && s.Count > 0 {
		return nil
	}
	return models.NewQuery(
		qm.Select(postgres.Stat()...),
		SoftwareExpr(),
		qm.From(From)).Bind(ctx, db, s)
}

func SoftwareExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(expr.PJava()),
		qm.Or2(models.FileWhere.Platform.EQ(expr.PLinux())),
		qm.Or2(models.FileWhere.Platform.EQ(expr.PDos())),
		qm.Or2(models.FileWhere.Platform.EQ(expr.PScript())),
		qm.Or2(models.FileWhere.Platform.EQ(expr.PWindows())),
	)
}

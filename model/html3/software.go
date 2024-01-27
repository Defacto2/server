package html3

// Package html3_software.go contains the database queries the HTML3 software category.

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model/expr"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Softwares contain statistics for releases that could be considered software.
type Softwares struct {
	Bytes int `boil:"size_total"`
	Count int `boil:"count_total"`
}

// Stat returns the total bytes and count of releases that could be considered software.
func (s *Softwares) Stat(ctx context.Context, db *sql.DB) error {
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

// SoftwareExpr returns a query modifier for the software category.
func SoftwareExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(expr.PJava()),
		qm.Or2(models.FileWhere.Platform.EQ(expr.PLinux())),
		qm.Or2(models.FileWhere.Platform.EQ(expr.PDos())),
		qm.Or2(models.FileWhere.Platform.EQ(expr.PScript())),
		qm.Or2(models.FileWhere.Platform.EQ(expr.PWindows())),
	)
}

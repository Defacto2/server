package model

// This file is the custom art category for the HTML3 template.

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/model/expr"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Package html3_art.go contains the database queries the HTML3 digital or pixel art category.

// Arts contain statistics for releases that could be considered as digital or pixel art.
type Arts struct {
	Bytes int `boil:"size_total"`
	Count int `boil:"count_total"`
}

// Stat returns the total bytes and count of releases that could be considered as digital or pixel art.
func (a *Arts) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if a.Bytes > 0 && a.Count > 0 {
		return nil
	}
	return models.NewQuery(
		qm.Select(postgres.Stat()...),
		ArtExpr(),
		qm.From(From)).Bind(ctx, db, a)
}

// ArtExpr returns a query modifier for the digital or pixel art category.
func ArtExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.NEQ(expr.SBbs()),
		models.FileWhere.Platform.EQ(expr.PImage()),
	)
}

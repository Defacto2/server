package model

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// nfo is the database model for the nfo table.
// proof
// nfo-tool

type Nfo struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
	Year0 int `boil:"min_year"`
	YearX int `boil:"max_year"`
}

func (n *Nfo) Stat(ctx context.Context, db *sql.DB) error {
	// if n.Bytes > 0 && n.Count > 0 {
	// 	return nil
	// }
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter, postgres.MinYear, postgres.MaxYear),
		qm.Expr(
			models.FileWhere.Section.EQ(nfo()),
		),
		qm.From(From)).Bind(ctx, db, n)
}

type NfoTool struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

func (n *NfoTool) Stat(ctx context.Context, db *sql.DB) error {
	// if n.Bytes > 0 && n.Count > 0 {
	// 	return nil
	// }
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		qm.Expr(
			models.FileWhere.Section.EQ(nfoTool()),
		),
		qm.From(From)).Bind(ctx, db, n)
}

type Proof struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

func (p *Proof) Stat(ctx context.Context, db *sql.DB) error {
	// if p.Bytes > 0 && p.Count > 0 {
	// 	return nil
	// }
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		qm.Expr(
			models.FileWhere.Section.EQ(proof()),
		),
		qm.From(From)).Bind(ctx, db, p)
}

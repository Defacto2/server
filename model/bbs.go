package model

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/model/modext"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type BBS struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
	Year0 int `boil:"min_year"`
	YearX int `boil:"max_year"`
}

func (b *BBS) Stat(ctx context.Context, db *sql.DB) error {
	// if a.Bytes > 0 && a.Count > 0 {
	// 	return nil
	// }
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter, postgres.MinYear, postgres.MaxYear),
		qm.Expr(
			models.FileWhere.Section.EQ(modext.SBbs()),
		),
		qm.From(From)).Bind(ctx, db, b)
}

type BBStro struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

func (b *BBStro) Stat(ctx context.Context, db *sql.DB) error {
	// if a.Bytes > 0 && a.Count > 0 {
	// 	return nil
	// }
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		qm.Expr(
			models.FileWhere.Section.EQ(modext.SBbs()),
			models.FileWhere.Platform.EQ(modext.PDos()),
		),
		qm.From(From)).Bind(ctx, db, b)
}

type BBSText struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

func (b *BBSText) Stat(ctx context.Context, db *sql.DB) error {
	// if a.Bytes > 0 && a.Count > 0 {
	// 	return nil
	// }
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		qm.Expr(
			models.FileWhere.Section.EQ(modext.SBbs()),
			models.FileWhere.Platform.EQ(modext.PText()),
		),
		qm.From(From)).Bind(ctx, db, b)
}

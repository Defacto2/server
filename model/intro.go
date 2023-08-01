package model

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type Demo struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

func (d *Demo) Stat(ctx context.Context, db *sql.DB) error {
	if d.Bytes > 0 && d.Count > 0 {
		return nil
	}
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		qm.Expr(
			models.FileWhere.Section.EQ(demo()),
		),
		qm.From(From)).Bind(ctx, db, d)
}

// Intro contain statistics for releases that could be considered intros or cracktros.
type Intro struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
	Year0 int `boil:"min_year"`
	YearX int `boil:"max_year"`
}

// Stat counts the total number and total byte size of releases that could be considered intros or cracktros.
func (i *Intro) Stat(ctx context.Context, db *sql.DB) error {
	// if i.Bytes > 0 && i.Count > 0 {
	// 	return nil
	// }
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter, postgres.MinYear, postgres.MaxYear),
		qm.Expr(
			models.FileWhere.Section.EQ(intro()),
		),
		qm.From(From)).Bind(ctx, db, i)
}

type IntroDOS struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

func (i *IntroDOS) Stat(ctx context.Context, db *sql.DB) error {
	// if i.Bytes > 0 && i.Count > 0 {
	// 	return nil
	// }
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		qm.Expr(
			models.FileWhere.Section.EQ(intro()),
			models.FileWhere.Platform.EQ(dos()),
		),
		qm.From(From)).Bind(ctx, db, i)
}

type IntroWindows struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of releases that could be considered Windows intros or cracktros.
func (i *IntroWindows) Stat(ctx context.Context, db *sql.DB) error {
	// if i.Bytes > 0 && i.Count > 0 {
	// 	return nil
	// }
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		qm.Expr(
			models.FileWhere.Section.EQ(intro()),
			models.FileWhere.Platform.EQ(windows()),
		),
		qm.From(From)).Bind(ctx, db, i)
}

type Installer struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of releases that could be considered installers.
func (i *Installer) Stat(ctx context.Context, db *sql.DB) error {
	// if i.Bytes > 0 && i.Count > 0 {
	// 	return nil
	// }
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		qm.Expr(
			models.FileWhere.Section.EQ(install()),
		),
		qm.From(From)).Bind(ctx, db, i)
}

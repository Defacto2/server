package model

// This package contains sqlboiler models for the intros, installers and demoscene releases.

import (
	"context"
	"database/sql"
	"time"

	"github.com/Defacto2/server/model/modext"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type Demo struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total bytes of demoscene releases.
func (d *Demo) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.DemoExpr(),
		qm.From(From)).Bind(ctx, db, d)
}

// List returns a list of demoscene releases.
func (d *Demo) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.DemoExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
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
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter, postgres.MinYear, postgres.MaxYear),
		modext.IntroExpr(),
		qm.From(From)).Bind(ctx, db, i)
}

// List returns a list of releases that could be considered intros or cracktros.
func (i *Intro) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.IntroExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

type IntroDOS struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of releases that could be considered DOS intros or cracktros.
func (i *IntroDOS) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.IntroDOSExpr(),
		qm.From(From)).Bind(ctx, db, i)
}

// List returns a list of releases that could be considered DOS intros or cracktros.
func (i *IntroDOS) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.IntroDOSExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

type IntroWindows struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
	Cache time.Time
}

// Stat counts the total number and total byte size of releases that could be considered Windows intros or cracktros.
func (i *IntroWindows) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.IntroWindowsExpr(),
		qm.From(From)).Bind(ctx, db, i)
}

// List returns a list of releases that could be considered Windows intros or cracktros.
func (i *IntroWindows) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.IntroWindowsExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

type Installer struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of releases that could be considered installers.
func (i *Installer) Stat(ctx context.Context, db *sql.DB) error {
	return models.NewQuery(
		qm.Select(postgres.SumSize, postgres.Counter),
		modext.InstallExpr(),
		qm.From(From)).Bind(ctx, db, i)
}

// List returns a list of releases that could be considered installers.
func (i *Installer) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	return models.Files(modext.InstallExpr(),
		qm.Offset(calc(offset, limit)), qm.Limit(limit)).All(ctx, db)
}

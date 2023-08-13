package model

// Package file file_cracktros_intros.go contains sqlboiler models for the intros, installers and demoscene releases.

import (
	"context"
	"database/sql"
	"time"

	"github.com/Defacto2/server/model/expr"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Demo is a the model for the demoscene releases.
type Demo struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

func (d *Demo) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Stat()...),
		expr.DemoExpr(),
		qm.From(From)).Bind(ctx, db, d)
}

func (d *Demo) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.DemoExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// Intro contain statistics for releases that could be considered intros or cracktros.
type Intro struct {
	Bytes   int `boil:"size_sum"`
	Count   int `boil:"counter"`
	YearMin int `boil:"min_year"`
	YearMax int `boil:"max_year"`
}

func (i *Intro) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.IntroExpr(),
		qm.From(From)).Bind(ctx, db, i)
}

func (i *Intro) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.IntroExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// IntroDOS contain statistics for releases that could be considered DOS intros or cracktros.
type IntroDOS struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

func (i *IntroDOS) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Stat()...),
		expr.IntroDOSExpr(),
		qm.From(From)).Bind(ctx, db, i)
}

func (i *IntroDOS) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.IntroDOSExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// IntroWindows contain statistics for releases that could be considered Windows intros or cracktros.
type IntroWindows struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
	Cache time.Time
}

func (i *IntroWindows) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Stat()...),
		expr.IntroWindowsExpr(),
		qm.From(From)).Bind(ctx, db, i)
}

func (i *IntroWindows) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.IntroWindowsExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// Installer contain statistics for releases that could be considered installers.
type Installer struct {
	Bytes int `boil:"size_sum"`
	Count int `boil:"counter"`
}

func (i *Installer) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Stat()...),
		expr.InstallExpr(),
		qm.From(From)).Bind(ctx, db, i)
}

func (i *Installer) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.InstallExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

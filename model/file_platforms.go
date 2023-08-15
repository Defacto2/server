package model

// Package file_platoforms.go contains the database queries for operating systems.

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/model/expr"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// DOS is a the model for the MS-DOS operating system.
type DOS struct {
	Bytes   int `boil:"size_sum"`
	Count   int `boil:"counter"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (d *DOS) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.DOSExpr(),
		qm.From(From)).Bind(ctx, db, d)
}

func (d *DOS) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.DOSExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// Java is a the model for the Java operating system.
type Java struct {
	Bytes   int `boil:"size_sum"`
	Count   int `boil:"counter"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (j *Java) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.JavaExpr(),
		qm.From(From)).Bind(ctx, db, j)
}

func (j *Java) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.JavaExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// Linux is a the model for the Linux operating system.
type Linux struct {
	Bytes   int `boil:"size_sum"`
	Count   int `boil:"counter"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (l *Linux) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.LinuxExpr(),
		qm.From(From)).Bind(ctx, db, l)
}

func (l *Linux) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.LinuxExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// Mac is a the model for the Macintosh operating system.
type Mac struct {
	Bytes   int `boil:"size_sum"`
	Count   int `boil:"counter"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (m *Mac) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.MacExpr(),
		qm.From(From)).Bind(ctx, db, m)
}

func (m *Mac) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.MacExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// Script is a the model for the script and interpreted languages.
type Script struct {
	Bytes   int `boil:"size_sum"`
	Count   int `boil:"counter"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (s *Script) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.ScriptExpr(),
		qm.From(From)).Bind(ctx, db, s)
}

func (s *Script) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.ScriptExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

// Windows is a the model for the Windows operating system.
type Windows struct {
	Bytes   int `boil:"size_sum"`
	Count   int `boil:"counter"`
	MinYear int `boil:"min_year"`
	MaxYear int `boil:"max_year"`
}

func (w *Windows) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Columns()...),
		expr.WindowsExpr(),
		qm.From(From)).Bind(ctx, db, w)
}

func (w *Windows) List(ctx context.Context, db *sql.DB, offset, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	return models.Files(
		expr.WindowsExpr(),
		qm.Offset(calc(offset, limit)),
		qm.Limit(limit)).All(ctx, db)
}

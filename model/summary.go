package model

import (
	"context"
	"database/sql"
	"strings"

	"github.com/Defacto2/sceners/pkg/rename"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Package file summary.go contains the database queries for the statistics of files.

// Summary counts the total number files, file sizes and the earliest and latest years.
type Summary struct {
	SumBytes int `boil:"size_total"`  // Sum total of the file sizes.
	SumCount int `boil:"count_total"` // Sum total count of the files.
	MinYear  int `boil:"min_year"`    // Minimum or earliest year of the files.
	MaxYear  int `boil:"max_year"`    // Maximum or latest year of the files.
}

func (s *Summary) All(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	return models.NewQuery(
		qm.Select(postgres.Statistics()...),
		qm.From(From)).Bind(ctx, db, s)
}

func (r *Summary) BBS(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if err := queries.Raw(string(postgres.SumBBS())).Bind(ctx, db, r); err != nil {
		return err
	}
	return nil
}

func (r *Summary) FTP(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if err := queries.Raw(string(postgres.SumFTP())).Bind(ctx, db, r); err != nil {
		return err
	}
	return nil
}

func (r *Summary) Magazine(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if err := queries.Raw(string(postgres.SumMag())).Bind(ctx, db, r); err != nil {
		return err
	}
	return nil
}

func (s *Summary) Releaser(ctx context.Context, db *sql.DB, name string) error {
	if db == nil {
		return ErrDB
	}
	n := strings.ToUpper(rename.DeObfuscateURL(name))
	x := null.StringFrom(n)
	return models.NewQuery(
		qm.Select(postgres.Statistics()...),
		qm.Where("upper(group_brand_for) = ?", x),
		qm.From(From)).Bind(ctx, db, s)
}

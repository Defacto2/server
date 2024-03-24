package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Scener is a collective, group or individual, that releases files.
type Scener string

// Sceners is a collection of sceners.
type Sceners []*struct {
	Name Scener `boil:"scener"`
}

// List gets the unique scener names.
func (s *Scener) List(ctx context.Context, db *sql.DB, name string) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	boil.DebugMode = true
	return models.Files(
		qm.Where(ScenerSQL(name)),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, db)
}

// ScenerSQL is the SQL query for getting sceners.
func ScenerSQL(name string) string {
	n := strings.ToUpper(releaser.Humanize(name))
	exact := fmt.Sprintf("(upper(credit_text) = '%s')"+
		" OR (upper(credit_program) = '%s')"+
		" OR (upper(credit_illustration) = '%s')"+
		" OR (upper(credit_audio) = '%s')", n, n, n, n)
	first := fmt.Sprintf("(upper(credit_text) LIKE '%s,%%')"+
		" OR (upper(credit_program) LIKE '%s,%%')"+
		" OR (upper(credit_illustration) LIKE '%s,%%')"+
		" OR (upper(credit_audio) LIKE '%s,%%')", n, n, n, n)
	middle := fmt.Sprintf("(upper(credit_text) LIKE '%%,%s,%%')"+
		" OR (upper(credit_program) LIKE '%%,%s,%%')"+
		" OR (upper(credit_illustration) LIKE '%%,%s,%%')"+
		" OR (upper(credit_audio) LIKE '%%,%s,%%')", n, n, n, n)
	last := fmt.Sprintf("(upper(credit_text) LIKE '%%,%s')"+
		" OR (upper(credit_program) LIKE '%%,%s')"+
		" OR (upper(credit_illustration) LIKE '%%,%s')"+
		" OR (upper(credit_audio) LIKE '%%,%s')", n, n, n, n)
	return fmt.Sprintf("(%s) OR (%s) OR (%s) OR (%s)", exact, first, middle, last)
}

// All gets a list of all sceners.
func (s *Sceners) All(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if len(*s) > 0 {
		return nil
	}
	query := string(postgres.Sceners())
	return queries.Raw(query).Bind(ctx, db, s)
}

// Writer gets a list of sceners who have been credited for text.
func (s *Sceners) Writer(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if len(*s) > 0 {
		return nil
	}
	query := string(postgres.Writers())
	return queries.Raw(query).Bind(ctx, db, s)
}

// Artist gets a list of sceners who have been credited for graphics or art.
func (s *Sceners) Artist(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if len(*s) > 0 {
		return nil
	}
	query := string(postgres.Artists())
	return queries.Raw(query).Bind(ctx, db, s)
}

// Coder gets a list of sceners who have been credited for programming.
func (s *Sceners) Coder(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if len(*s) > 0 {
		return nil
	}
	query := string(postgres.Coders())
	return queries.Raw(query).Bind(ctx, db, s)
}

// Musician gets a list of sceners who have been credited for music or audio.
func (s *Sceners) Musician(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if len(*s) > 0 {
		return nil
	}
	query := string(postgres.Musicians())
	return queries.Raw(query).Bind(ctx, db, s)
}

// Sort gets a sorted slice of unique sceners.
func (s Sceners) Sort() []string {
	var sceners []string
	for _, scener := range s {
		sceners = append(sceners, strings.Split(string(scener.Name), ",")...)
	}
	return helper.DeleteDupe(sceners...)
}

package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Defacto2/sceners/pkg/rename"
	"github.com/Defacto2/server/pkg/helper"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// SELECT *
// FROM "files"
// WHERE "credit_text" ILIKE '%Absolute Zero%' OR "credit_program" ILIKE '%Absolute Zero%' OR "credit_illustration" ILIKE '%Absolute Zero%' OR "credit_audio" ILIKE '%Absolute Zero%'
// LIMIT 50

// SELECT *
// FROM "files"
// WHERE "credit_text" ILIKE '%ab%,' OR
// "credit_program" ILIKE '%ab%,' OR
// "credit_illustration" ILIKE '%ab%,' OR
// "credit_audio" ILIKE '%ab%,'
// LIMIT 50

// TODO matchText matchProg etc (name string) string
// match ben ben, ,ben, ,ben

type Scener string

type Sceners []*struct {
	Name Scener `boil:"scener"`
}

func (s *Scener) List(ctx context.Context, db *sql.DB, name string) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	//n := strings.ToUpper(rename.DeObfuscateURL(name))
	//x := null.StringFrom(n)
	boil.DebugMode = true
	return models.Files(
		qm.Where(ScenerSQL(name)),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, db)
}

func ScenerSQL(name string) string {
	n := strings.ToUpper(rename.DeObfuscateURL(name))
	return fmt.Sprintf("(upper(credit_text) ILIKE '%%%s%%')"+
		" OR (upper(credit_program) ILIKE '%%%s%%')"+
		" OR (upper(credit_illustration) ILIKE '%%%s%%')"+
		" OR (upper(credit_audio) ILIKE '%%%s%%')", n, n, n, n)
}

// All gets a list of all sceners.
func (s *Sceners) All(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if len(*s) > 0 {
		return nil
	}
	query := string(postgres.SelectSceners())
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
	query := string(postgres.SelectWriter())
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
	query := string(postgres.SelectArtist())
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
	query := string(postgres.SelectCoder())
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
	query := string(postgres.SelectMusician())
	return queries.Raw(query).Bind(ctx, db, s)
}

// Sort gets a sorted slice of unique sceners.
func (s Sceners) Sort() []string {
	var sceners []string
	for _, scener := range s {
		sceners = append(sceners, strings.Split(string(scener.Name), ",")...)
	}
	return helper.DeleteDupe(sceners)
}

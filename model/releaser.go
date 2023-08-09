package model

// Package releaser.go contains the database queries the releasers and groups.

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/pkg/helpers"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/volatiletech/sqlboiler/v4/queries"
)

// Summary counts the total number files, file sizes and the earliest and latest years.
type Summary struct {
	SumBytes int `boil:"size_total"`  // Sum total of the file sizes.
	SumCount int `boil:"count_total"` // Sum total count of the files.
	MinYear  int `boil:"min_year"`    // Minimum or earliest year of the files.
	MaxYear  int `boil:"max_year"`    // Maximum or latest year of the files.
}

// Releaser is a collective, group or individual, that releases files.
type Releaser struct {
	Name  string `boil:"releaser"`  // Name of the releaser.
	URI   string ``                 // URI slug for the releaser with no boiler bind.
	Bytes int    `boil:"size_sum"`  // Bytes are the total size of all the files under this releaser.
	Count int    `boil:"count_sum"` // Count is the total number of files under this releaser.
}

// Releasers is a collection of releasers.
type Releasers []*struct {
	Unique Releaser `boil:",bind"` // Unique is the releaser.
}

// Stat counts the total number of files and file sizes for all the releasers.
// func (r *Rels) Stat(ctx context.Context, db *sql.DB) error {
// 	if db == nil {
// 		return ErrDB
// 	}
// 	mods := qm.SQL(string(postgres.StatRelr()))
// 	f, err := models.Files(mods).All(ctx, db)
// 	if err != nil {
// 		return err
// 	}
// 	r.Count = len(f)
// 	return nil
// }

// All gets the unique releaser names and their total file count and file sizes.
func (r *Releasers) All(ctx context.Context, db *sql.DB, offset, limit int, o Order) error {
	if db == nil {
		return ErrDB
	}
	if len(*r) > 0 {
		return nil
	}
	if err := queries.Raw(string(postgres.SelectRelr())).Bind(ctx, db, r); err != nil {
		return err
	}
	r.Slugs()
	return nil
}

func (s *Summary) All(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if err := queries.Raw(string(postgres.SumAll())).Bind(ctx, db, s); err != nil {
		return err
	}
	return nil
}

// Magazine gets the unique magazine titles and their total issue count and file sizes.
func (r *Releasers) Magazine(ctx context.Context, db *sql.DB, offset, limit int, o Order) error {
	if db == nil {
		return ErrDB
	}
	if len(*r) > 0 {
		return nil
	}
	if err := queries.Raw(string(postgres.SelectMag())).Bind(ctx, db, r); err != nil {
		return err
	}
	r.Slugs()
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

// BBS gets the unique BBS site names and their total file count and file sizes.
func (r *Releasers) BBS(ctx context.Context, db *sql.DB, offset, limit int, o Order) error {
	if db == nil {
		return ErrDB
	}
	if len(*r) > 0 {
		return nil
	}
	if err := queries.Raw(string(postgres.SelectBBS())).Bind(ctx, db, r); err != nil {
		return err
	}
	r.Slugs()
	return nil
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

// FTP gets the unique FTP site names and their total file count and file sizes.
func (r *Releasers) FTP(ctx context.Context, db *sql.DB, offset, limit int, o Order) error {
	if db == nil {
		return ErrDB
	}
	if len(*r) > 0 {
		return nil
	}
	if err := queries.Raw(string(postgres.SelectFTP())).Bind(ctx, db, r); err != nil {
		return err
	}
	r.Slugs()
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

// Slugs saves URL friendly strings to the Group names.
func (r *Releasers) Slugs() {
	for _, releaser := range *r {
		releaser.Unique.URI = helpers.Slug(releaser.Unique.Name)
	}
}

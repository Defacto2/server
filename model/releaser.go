package model

// Package releaser.go contains the database queries the releasers and groups.

import (
	"context"
	"database/sql"
	"strings"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Releaser is a collective, group or individual, that releases files.
type Releaser struct {
	// Name of the releaser.
	Name string `boil:"releaser"`
	// URI slug for the releaser with no boiler bind.
	URI string ``
	// Bytes are the total size of all the files under this releaser.
	Bytes int `boil:"size_total"`
	// Count is the total number of files under this releaser.
	Count int `boil:"count_sum"`
	// Year is used for optional sorting and is the earliest year the releaser was active.
	Year null.Int `boil:"min_year"`
}

// Releasers is a collection of releasers.
type Releasers []*struct {
	Unique Releaser `boil:",bind"` // Unique is the releaser.
}

type ReleaserList struct {
	Name string `boil:"releaser"`
}

// ReleaserStr is a distinct data list of releasers.
type ReleaserStr []ReleaserList

func (r *ReleaserStr) List(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	query := "SELECT DISTINCT releaser " +
		"FROM files " +
		"CROSS JOIN LATERAL (values(group_brand_for),(group_brand_by)) AS T(releaser) " +
		"WHERE NULLIF(releaser, '') IS NOT NULL " +
		"GROUP BY releaser " +
		"ORDER BY releaser ASC"
	return queries.Raw(query).Bind(ctx, db, r)
}

func (r *Releasers) List(ctx context.Context, db *sql.DB, name string) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	n := strings.ToUpper(releaser.Humanize(name))
	x := null.StringFrom(n)
	return models.Files(
		qm.Where("upper(group_brand_for) = ? OR upper(group_brand_by) = ?", x, x),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, db)
}

// All gets the unique releaser names and their total file count and file sizes.
// When reorder is true the results are ordered by the total file counts.
func (r *Releasers) All(ctx context.Context, db *sql.DB, offset, limit int, reorder bool) error {
	if db == nil {
		return ErrDB
	}
	if len(*r) > 0 {
		return nil
	}
	query := string(postgres.DistReleaser())
	if reorder {
		query = string(postgres.DistReleaserSummed())
	}
	if err := queries.Raw(query).Bind(ctx, db, r); err != nil {
		return err
	}
	r.Slugs()
	return nil
}

// Magazine gets the unique magazine titles and their total issue count and file sizes.
func (r *Releasers) MagazineAZ(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if len(*r) > 0 {
		return nil
	}
	if err := queries.Raw(string(postgres.DistMagazine())).Bind(ctx, db, r); err != nil {
		return err
	}
	r.Slugs()
	return nil
}

// Magazine gets the unique magazine titles and their total issue count and file sizes.
func (r *Releasers) Magazine(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if len(*r) > 0 {
		return nil
	}
	if err := queries.Raw(string(postgres.DistMagazineByYear())).Bind(ctx, db, r); err != nil {
		return err
	}
	r.Slugs()
	return nil
}

// BBS gets the unique BBS site names and their total file count and file sizes.
func (r *Releasers) BBS(ctx context.Context, db *sql.DB, offset, limit int, reorder bool) error {
	if db == nil {
		return ErrDB
	}
	if len(*r) > 0 {
		return nil
	}
	query := string(postgres.DistBBS())
	if reorder {
		query = string(postgres.DistBBSSummed())
	}
	if err := queries.Raw(query).Bind(ctx, db, r); err != nil {
		return err
	}
	r.Slugs()
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
	if err := queries.Raw(string(postgres.DistFTP())).Bind(ctx, db, r); err != nil {
		return err
	}
	r.Slugs()
	return nil
}

// Slugs saves URL friendly strings to the Group names.
func (r *Releasers) Slugs() {
	for _, releaser := range *r {
		releaser.Unique.URI = helper.Slug(releaser.Unique.Name)
	}
}

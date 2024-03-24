package model

// Package releaser.go contains the database queries the releasers and groups.

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	namer "github.com/Defacto2/releaser/name"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Releaser is a collective, group or individual, that releases files.
type Releaser struct {
	Name  string   `boil:"releaser"`   // Name of the releaser.
	URI   string   ``                  // URI slug for the releaser, with no boiler bind.
	Bytes int      `boil:"size_total"` // Bytes are the total size of all the files under this releaser.
	Count int      `boil:"count_sum"`  // Count is the total number of files under this releaser.
	Year  null.Int `boil:"min_year"`   // Year is used for optional sorting
	// and is the earliest year the releaser was active.
}

// Releasers is a collection of releasers.
type Releasers []*struct {
	Unique Releaser `boil:",bind"` // Unique releaser.
}

// ReleaserName is a releaser name.
type ReleaserName struct {
	Name string `boil:"releaser"`
}

// ReleaserNames is a distinct data list of releasers.
type ReleaserNames []ReleaserName

// OrderBy is the sorting order for ALL the releasers.
type OrderBy uint

const (
	Prolific     OrderBy = iota // Prolific orders by the total artifact count.
	Alphabetical                // Alphabetical orders by the releaser name.
	Oldest                      // Oldest orders by the year of the first artifact.
)

// List gets the unique releaser names.
func (r *ReleaserNames) List(ctx context.Context, db *sql.DB) error {
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

// List gets the unique releaser names.
func (r *Releasers) List(ctx context.Context, db *sql.DB, name string) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	s, err := namer.Humanize(namer.Path(name))
	if err != nil {
		return nil, err
	}
	n := strings.ToUpper(s)
	x := null.StringFrom(n)
	return models.Files(
		qm.Where("upper(group_brand_for) = ? OR upper(group_brand_by) = ?", x, x),
		qm.OrderBy(ClauseOldDate),
	).All(ctx, db)
}

// All gets the unique releaser names and their total file count and file sizes.
// When reorder is true the results are ordered by the total file counts.
func (r *Releasers) All(ctx context.Context, db *sql.DB, order OrderBy, limit, page int) error {
	if db == nil {
		return ErrDB
	}
	if r != nil && len(*r) > 0 {
		return nil
	}
	var query string
	switch order {
	case Prolific:
		query = string(postgres.DistReleaserSummed())
	case Alphabetical:
		query = string(postgres.DistReleaser())
	case Oldest:
		query = string(postgres.DistReleaserByYear())
	default:
		return ErrOrderBy
	}
	if limit > 0 {
		if page < 1 {
			page = 1
		}
		limit, offset := calculateLimitAndOffset(page, limit)
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
	}
	if err := queries.Raw(query).Bind(ctx, db, r); err != nil {
		return err
	}
	r.Slugs()
	return nil
}

// Find the unique releaser names that are similar to the named strings.
// The results are ordered by the total file counts.
// The required limit is the maximum number of results to return or defaults to 10.
func (r *Releasers) Similar(ctx context.Context, db *sql.DB, limit uint, names ...string) error {
	if len(names) == 0 {
		return nil
	}
	if db == nil {
		return ErrDB
	}
	if r != nil && len(*r) > 0 {
		return nil
	}

	like := names
	for i, name := range names {
		x, err := namer.Humanize(namer.Path(name))
		if err != nil {
			return err
		}
		like[i] = strings.ToUpper(x)
	}
	query := string(postgres.ReleaserSimilarTo(like...))
	{
		const page, max = 1, 10
		size := int(limit) | max
		val, offset := calculateLimitAndOffset(page, size)
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", val, offset)
	}
	if err := queries.Raw(query).Bind(ctx, db, r); err != nil {
		return err
	}
	r.Slugs()
	return nil
}

func calculateLimitAndOffset(pageNumber int, pageSize int) (int, int) {
	limit := pageSize
	offset := (pageNumber - 1) * pageSize
	return limit, offset
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
func (r *Releasers) BBS(ctx context.Context, db *sql.DB, reorder bool) error {
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
func (r *Releasers) FTP(ctx context.Context, db *sql.DB) error {
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

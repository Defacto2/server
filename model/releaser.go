package model

// Package releaser.go contains the database queries the releasers and groups.

import (
	"context"
	"database/sql"
	"strings"

	"github.com/Defacto2/sceners/pkg/rename"
	"github.com/Defacto2/server/pkg/helper"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

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

func (r *Releasers) List(ctx context.Context, db *sql.DB, name string) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	n := strings.ToUpper(rename.DeObfuscateURL(name))
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
	query := string(postgres.SelectRels())
	if reorder {
		query = string(postgres.SelectRelPros())
	}
	if err := queries.Raw(query).Bind(ctx, db, r); err != nil {
		return err
	}
	r.Slugs()
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

// BBS gets the unique BBS site names and their total file count and file sizes.
func (r *Releasers) BBS(ctx context.Context, db *sql.DB, offset, limit int, reorder bool) error {
	if db == nil {
		return ErrDB
	}
	if len(*r) > 0 {
		return nil
	}
	query := string(postgres.SelectBBS())
	if reorder {
		query = string(postgres.SelectBBSPros())
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
	if err := queries.Raw(string(postgres.SelectFTP())).Bind(ctx, db, r); err != nil {
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

package model

import (
	"context"
	"database/sql"

	"github.com/Defacto2/server/pkg/helpers"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Rels counts the total number of unique groups, products and releasers.
type Rels struct {
	Count int `boil:"counter"` // Count of unique groups.
}

// Releaser is a collective, group or individual, that releases files.
type Releaser struct {
	Name  string `boil:"group_brand"` // Name of the releaser.
	URI   string ``                   // URI slug for the scener.
	Bytes int    `boil:"size_sum"`    // Bytes are the total size of all the files under this releaser.
	Count int    `boil:"count"`       // Count is the total number of files under this releaser.
}

// Releasers is a collection of releasers.
type Releasers []*struct {
	Unique Releaser `boil:",bind"` // Unique is the releaser.
}

// Stat counts the total number of unique groups.
func (r *Rels) Stat(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	f, err := models.Files(qm.SQL(postgres.SQLGroupStat())).All(ctx, db)
	if err != nil {
		return err
	}
	r.Count = len(f)
	return nil
}

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
		releaser.Unique.URI = helpers.Slug(releaser.Unique.Name)
	}
}

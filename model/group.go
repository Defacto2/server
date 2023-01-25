package model

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Groups contain statistics for releases that could be considered as digital or pixel art.
type Groups struct {
	Bytes int `boil:"size_sum"` // unused
	Count int `boil:"counter"`
}

// Stat counts the total number and total byte size of releases that could be considered as digital or pixel art.
func (g *Groups) Stat(ctx context.Context, db *sql.DB) error {
	if g.Count > 0 {
		return nil
	}
	var err error
	if g.Count, err = GroupCount(ctx, db); err != nil {
		return err
	}
	return nil
}

type Group struct {
	Name  string `boil:"group_brand"` // todo rename to group_brand
	URI   string // URI slug for the scener.
	Bytes int    `boil:"size_sum"`
	Count int    `boil:"count"`
}

type GroupCol []*struct {
	Group Group `boil:",bind"`
}

var Collection GroupCol

// GroupList returns the names and statistics of the unique groups.
func (g *GroupCol) GroupList(ctx context.Context, db *sql.DB) error {
	// return models.Files(
	// 	qm.Select(models.FileColumns.GroupBrandFor),
	// 	qm.Distinct(models.FileColumns.GroupBrandFor),
	// 	qm.Load("", qm.Select(SumSize, Counter), qm.From(From)),
	// ).Bind(ctx, db, g)
	// "SELECT COUNT(DISTINCT(LOWER(TRIM(files.group_brand_for)))) FROM files"
	if len(*g) > 0 {
		return nil
	}
	err := models.Files(
		qm.SQL("SELECT DISTINCT group_brand, "+
			"COUNT(group_brand) AS count, "+
			"SUM(files.filesize) AS size_sum "+
			"FROM files "+
			"CROSS JOIN LATERAL (values(group_brand_for),(group_brand_by)) AS T(group_brand) "+
			"WHERE NULLIF(group_brand, '') IS NOT NULL "+ // handle empty and null values
			"GROUP BY group_brand "+
			"ORDER BY group_brand"),
	).Bind(ctx, db, g)
	if err != nil {
		return err
	}
	g.Slugs()
	return nil
}

// Slug returns a URL friendly string of the group name.
func (g *GroupCol) Slugs() {
	for _, group := range *g {
		group.Group.URI = Slug(group.Group.Name)
	}
}

// Count the number of records associated with the group.
func Count(g string, ctx context.Context, db *sql.DB) (int, error) {
	// TODO: in postgresql, when comparing lowercase in queries, any column indexes are void
	x := null.String{String: string(g), Valid: true}
	c, err := models.Files(
		qm.Select(models.FileColumns.GroupBrandFor), models.FileWhere.GroupBrandFor.EQ(x),
	).Count(ctx, db)
	if err != nil {
		return -1, err
	}
	return int(c), nil
}

// Slug returns a URL friendly string of the named group.
func Slug(name string) string {
	s := name
	// hyphen to underscore
	re := regexp.MustCompile(`\-`)
	s = re.ReplaceAllString(s, "_")
	// multiple groups get separated with asterisk
	re = regexp.MustCompile(`\, `)
	s = re.ReplaceAllString(s, "*")
	// any & characters need replacement due to HTML escaping
	re = regexp.MustCompile(` \& `)
	s = re.ReplaceAllString(s, " ampersand ")
	// numbers receive a leading hyphen
	re = regexp.MustCompile(` ([0-9])`)
	s = re.ReplaceAllString(s, "-$1")
	// delete all other characters
	const deleteAllExcept = `[^A-Za-z0-9 \-\+\.\_\*]`
	re = regexp.MustCompile(deleteAllExcept)
	s = re.ReplaceAllString(s, "")
	// trim whitespace and replace any space separators with hyphens
	s = strings.TrimSpace(strings.ToLower(s))
	re = regexp.MustCompile(` `)
	s = re.ReplaceAllString(s, "-")
	return s
}

func Tester(ctx context.Context, db *sql.DB) (int, error) {
	c, err := models.Files(
		qm.Select(models.FileColumns.GroupBrandFor),
		qm.Distinct(models.FileColumns.GroupBrandFor),
	).Count(ctx, db)
	if err != nil {
		return -1, err
	}
	return int(c), nil
}

// GroupCount returns to total number of unique groups.
func GroupCount(ctx context.Context, db *sql.DB) (int, error) {
	c, err := models.Files(qm.SQL("SELECT COUNT(DISTINCT(LOWER(TRIM(files.group_brand_for)))) FROM files")).Count(ctx, db)
	if err != nil {
		fmt.Println(err)
		return -1, err
	}
	return int(c), nil
}

// GroupList returns a collection of the unique groups.
func GroupList(ctx context.Context, db *sql.DB) (models.FileSlice, error) {
	return models.Files(
		qm.Select(models.FileColumns.GroupBrandFor),
		qm.Distinct(models.FileColumns.GroupBrandFor),
	).All(ctx, db)
}

func latency() *time.Time {
	start := time.Now()
	r := new(big.Int)
	const n, k = 1000, 10
	r.Binomial(n, k)
	return &start
}

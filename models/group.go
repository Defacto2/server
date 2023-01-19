package models

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Defacto2/server/helpers"
	"github.com/Defacto2/server/postgres"
	"github.com/Defacto2/server/postgres/models"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// Groups is a cached collection of important, expensive group data.
// The Mu mutex must always be locked when writing to the Groups map.
type G struct {
	Mu   sync.RWMutex
	List map[string]Scener
}

// Group is a distinct scener group or organisation associated with the file record.
type Group string

// Grps is a cached collection of important, expensive group data.
// The Update method uses a background Go routine, so the Mu mutex must
// be locked before using this varable.
var Groups G // TODO: move to main? it may require its own package?

// Update or build the group collection with any missing group data.
// TODO: run this in production on startup?
func (g *G) Update() error {
	start := latency()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db, err := postgres.ConnectDB()
	if err != nil {
		return err
	}
	defer db.Close()
	results, err := GroupList(ctx, db)
	if err != nil {
		return err
	}

	go func() {
		if g.List == nil {
			g.List = make(map[string]Scener, len(results))
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		db, err := postgres.ConnectDB()
		if err != nil {
			return
		}
		defer db.Close()
		for i, r := range results {
			if r == nil {
				continue
			}
			name := strings.TrimSpace(r.GroupBrandFor.String)
			if len(name) == 0 {
				continue
			}
			key := Group(name).Slug()
			g.Mu.RLock()
			cached := g.List[key]
			g.Mu.RUnlock()
			if cached.Count > 0 {
				continue
			}
			sum, err := Group(name).Count(ctx, db)
			if err != nil {
				fmt.Println(err)
				continue
			}
			cached.Name = name
			cached.URI = key
			cached.Count = sum
			g.Mu.Lock()
			g.List[key] = cached
			g.Mu.Unlock()
			fmt.Printf("%s\r%d. Cached the group %q with %d records  ", helpers.Eraseline, i, name, sum)
		}
		fmt.Printf("\nCache builder, time taken, %s.\n", time.Since(*start))
	}()
	return nil
}

// Count the number of records associated with the group.
func (g Group) Count(ctx context.Context, db *sql.DB) (int, error) {
	// TODO: in postgresql, when comparing lowercase in queries, any column indexes are void
	c, err := models.Files(
		Select(models.FileColumns.GroupBrandFor),
		Where(groupFor, g),
	).Count(ctx, db)
	if err != nil {
		return -1, err
	}
	return int(c), nil
}

// Slug returns a URL friendly string of the group name.
func (g Group) Slug() string {
	s := string(g)
	re := regexp.MustCompile(`\-`)
	s = re.ReplaceAllString(s, "_")
	re = regexp.MustCompile(`\, `)
	s = re.ReplaceAllString(s, "*")
	re = regexp.MustCompile(` \& `)
	s = re.ReplaceAllString(s, " ampersand ")
	re = regexp.MustCompile(` ([0-9])`)
	s = re.ReplaceAllString(s, "-$1")
	const deleteAllExcept = `[^A-Za-z0-9 \-\+\.\_\*]`
	re = regexp.MustCompile(deleteAllExcept)
	s = re.ReplaceAllString(s, "")
	s = strings.ToLower(s)
	re = regexp.MustCompile(` `)
	s = re.ReplaceAllString(s, "-")
	return s
}

func Tester(ctx context.Context, db *sql.DB) (int, error) {
	c, err := models.Files(
		Select(models.FileColumns.GroupBrandFor),
		Distinct(models.FileColumns.GroupBrandFor),
	).Count(ctx, db)
	if err != nil {
		return -1, err
	}
	return int(c), nil
}

// GroupCount returns to total number of unique groups.
func GroupCount(ctx context.Context, db *sql.DB) (int, error) {
	c, err := models.Files(SQL("SELECT COUNT(DISTINCT(LOWER(TRIM(files.group_brand_for)))) FROM files")).Count(ctx, db)
	if err != nil {
		fmt.Println(err)
		return -1, err
	}
	return int(c), nil
}

// GroupList returns a collection of the unique groups.
func GroupList(ctx context.Context, db *sql.DB) (models.FileSlice, error) {
	return models.Files(
		Select(models.FileColumns.GroupBrandFor),
		Distinct(models.FileColumns.GroupBrandFor),
	).All(ctx, db)
}

func latency() *time.Time {
	start := time.Now()
	r := new(big.Int)
	const n, k = 1000, 10
	r.Binomial(n, k)
	return &start
}

package models

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/Defacto2/server/helpers"
	"github.com/Defacto2/server/postgres"
	"github.com/Defacto2/server/postgres/models"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// GroupCache is a cached collection of important, expensive group data.
// The Mu mutex must always be locked before writing this varable.
var GroupCache GroupCol

// GroupCol is a cached collection of important, expensive group data.
// The Mu mutex must always be locked when writing to the Groups map.
type GroupCol struct {
	Mu     sync.Mutex
	Groups map[string]Scener
}

// Update or build the group collection with any missing group data.
func (g *GroupCol) Update() error {
	// TODO: create libs using base: https://github.com/Defacto2/df2/blob/937cf38cddda8a38091258ae62f2db31e1b672cf/pkg/groups/internal/rename/rename.go#L83
	// TODO: run this in production on startup?
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db, err := postgres.ConnectDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// results are a distinct list of groups
	// TODO: use the specific All() func for this.
	results, err := models.Files(
		Select(models.FileColumns.GroupBrandFor),
		Distinct(models.FileColumns.GroupBrandFor),
	).All(ctx, db)
	if err != nil {
		return err
	}

	g.Mu.Lock()
	defer g.Mu.Unlock()

	if g.Groups == nil {
		g.Groups = make(map[string]Scener, len(results))
	}

	for i, r := range results {
		name := strings.TrimSpace(r.GroupBrandFor.String)
		if len(name) == 0 {
			continue
		}
		key := GroupForURL(name)
		cached := g.Groups[key]
		if cached.Count > 0 {
			continue
		}
		if reflect.DeepEqual(r, Scener{}) {
			continue
		}
		sum, err := CountGroup(name, ctx, db)
		if err != nil {
			fmt.Println(err)
			continue
		}
		g.Groups[key] = Scener{
			Name:  name,
			URI:   key,
			Count: sum,
		}
		fmt.Printf("%s\r%d. Cached the group %q with %d records", helpers.Eraseline, i, name, sum)
	}
	fmt.Println()
	return nil
}

// CountGroup returns the number of records associated with the named group.
func CountGroup(name string, ctx context.Context, db *sql.DB) (int, error) {
	// TODO: in postgresql, when comparing lowercase in queries, any column indexes are void
	c, err := models.Files(
		Select(models.FileColumns.GroupBrandFor),
		Where(groupFor, name),
	).Count(ctx, db)
	if err != nil {
		return -1, err
	}
	return int(c), nil
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

// GroupsTotalCount returns to total number of unique groups.
func GroupsTotalCount(ctx context.Context, db *sql.DB) (int, error) {
	c, err := models.Files(SQL("select count(distinct(LOWER(TRIM(files.group_brand_for)))) FROM files")).Count(ctx, db)
	if err != nil {
		fmt.Println(err)
		return -1, err
	}
	return int(c), nil
}

// TODO: move to Defacto2/sceners
func GroupForURL(g string) string {
	re := regexp.MustCompile(`\-`)
	s := re.ReplaceAllString(g, "_")
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

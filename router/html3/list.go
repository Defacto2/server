package html3

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/sceners"
	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/pkg/helpers"
	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/tags"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"go.uber.org/zap"

	pgm "github.com/Defacto2/server/pkg/postgres/models"
)

// Navigate handles offset and record limit pagination.
type Navigate struct {
	Current  string // Current name of the current record query.
	Limit    int    // Limit the number of records to return per query.
	Link1    int    // Link1 of the dynamic pagination.
	Link2    int    // Link2 of the dynamic pagination.
	Link3    int    // Link3 of the dynamic pagination.
	Page     int    // Page number of the current record query.
	PagePrev int    // PagePrev is the page number to the previous record query.
	PageNext int    // PageNext is the page number to the next record query.
	PageMax  int    // PageMax is the maximum and last page number of the record query.
	QueryStr string // QueryStr to append to all pagination links.
}

// Sugared logger passthrough.
type sugared struct {
	log *zap.SugaredLogger
}

// Category lists the file records associated with the category tag that is provided by the ID param in the URL.
func (s *sugared) Category(c echo.Context) error {
	return s.List(BySection, c)
}

// Platform lists the file records associated with the platform tag that is provided by the ID param in the URL.
func (s *sugared) Platform(c echo.Context) error {
	return s.List(ByPlatform, c)
}

// Group lists the file records associated with the group that is provided by the ID param in the URL.
func (s *sugared) Group(c echo.Context) error {
	return s.List(ByGroup, c)
}

// Group lists the file records described as art are digital + pixel art files.
func (s *sugared) Art(c echo.Context) error {
	return s.List(AsArt, c)
}

// Group lists the file records described as document + text art files.
func (s *sugared) Documents(c echo.Context) error {
	return s.List(AsDocuments, c)
}

// Group lists the file records described as software files.
func (s *sugared) Software(c echo.Context) error {
	return s.List(AsSoftware, c)
}

// List all the records associated with the RecordsBy grouping.
func (s *sugared) List(tt RecordsBy, c echo.Context) error {
	start := latency()
	id := c.Param("id")

	count, limit, page := 0, 0, 1
	offset := strings.TrimPrefix(c.Param("offset"), "/")
	if offset != "" {
		// this permits blank offsets param but returns 404 for a /0 value
		page, _ = strconv.Atoi(offset)
		if page < 1 {
			return echo.NewHTTPError(http.StatusNotFound,
				fmt.Sprintf("Page %d of %s doesn't exist", page, tt))
		}
	}

	name := sceners.CleanURL(id)
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		s.log.Warnf("%s: %s", errConn, err)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errConn)
	}
	defer db.Close()

	var records pgm.FileSlice
	order := Clauses(c.QueryString())
	switch tt {
	case BySection:
		limit = 2500
		records, err = order.FilesByCategory(id, page, limit, ctx, db)
		x, _ := model.CountByCategory(id, ctx, db)
		count = int(x)
	case ByPlatform:
		limit = 2500
		records, err = order.FilesByPlatform(id, ctx, db)
		x, _ := model.CountByPlatform(id, ctx, db)
		count = int(x)
	case ByGroup:
		// ByGroups do not need a pagination limit.
		records, err = order.FilesByGroup(name, ctx, db)
		count = len(records)
	case AsArt:
		limit = 1000
		records, err = order.ArtFiles(page, limit, ctx, db)
		count, _ = model.ArtCount(ctx, db)
	case AsDocuments:
		limit = 1000
		records, err = order.DocumentFiles(page, limit, ctx, db)
		count, _ = model.DocumentCount(ctx, db)
	case AsSoftware:
		limit = 1000
		records, err = order.SoftwareFiles(page, limit, ctx, db)
		count, _ = model.SoftwareCount(ctx, db)
	default:
		s.log.Warnf("%s: %s", errTag, tt)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errTag)
	}
	if err != nil {
		s.log.Warnf("%s: %s", errConn, err)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errConn)
	}
	if limit > 0 && count == 0 {
		return echo.NewHTTPError(http.StatusNotFound,
			fmt.Sprintf("The %s %q doesn't exist", tt, id))
	}

	var byteSum int64
	switch tt {
	case BySection:
		byteSum, err = model.ByteCountByCategory(id, ctx, db)
	case ByPlatform:
		byteSum, err = model.ByteCountByPlatform(id, ctx, db)
	case ByGroup:
		byteSum, err = model.ByteCountByGroup(name, ctx, db)
	case AsArt:
		byteSum, err = model.ArtByteCount(ctx, db)
	case AsDocuments:
		byteSum, err = model.DocumentByteCount(ctx, db)
	case AsSoftware:
		byteSum, err = model.SoftwareByteCount(ctx, db)
	default:
		s.log.Warnf("%s: %s", errTag, tt)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errTag)
	}
	if err != nil {
		s.log.Warnf("%s %s", errConn, err)
		return echo.NewHTTPError(http.StatusServiceUnavailable, errConn)
	}
	stat := fmt.Sprintf("%d files, %s", count, helpers.ByteCountFloat(byteSum))

	maxPage := uint(0)
	if limit > 0 {
		maxPage = helpers.PageCount(count, limit)
		if page > int(maxPage) {
			return echo.NewHTTPError(http.StatusNotFound,
				fmt.Sprintf("Page %d of %d for %s doesn't exist", page, maxPage, tt))
		}
	}

	current, desc := "", ""
	switch tt {
	case BySection, ByPlatform:
		key := tags.TagByURI(id)
		info := tags.Infos[key]
		name := tags.Names[key]
		desc = fmt.Sprintf("%s - %s.", name, info)
		s, err := url.JoinPath(tt.String(), key.String())
		if err != nil {
			log.Warnf("Could not create a URL string from %q and %q.", tt.String(), key.String())
		}
		current = s
	case AsArt:
		desc = fmt.Sprintf("%s, %s.", "Digital + pixel art", textArt)
		current = tt.String()
	case AsDocuments:
		desc = fmt.Sprintf("%s, %s.", "Document + text art", textDoc)
		current = tt.String()
	case AsSoftware:
		desc = fmt.Sprintf("%s, %s.", "Software", textSof)
		current = tt.String()
	default:
		current = tt.String()
	}

	navi := Navigate{
		Current:  current,
		Limit:    limit,
		Page:     page,
		PagePrev: previous(page),
		PageNext: next(page, maxPage),
		PageMax:  int(maxPage),
		QueryStr: qs(c.QueryString()),
	}
	navi.Link1, navi.Link2, navi.Link3 = Pagi(page, maxPage)
	err = c.Render(http.StatusOK, tt.String(), map[string]interface{}{
		"title":       fmt.Sprintf("%s%s%s", title, fmt.Sprintf("/%s/", tt), id),
		"home":        "",
		"description": desc,
		"parent":      tt.Parent(),
		"stats":       stat,
		"sort":        sorter(c.QueryString()),
		"records":     records,
		"latency":     fmt.Sprintf("%s.", time.Since(*start)),
		"navigate":    navi,
	})
	if err != nil {
		s.log.Errorf("%s: %s %d", errTmpl, err, tt)
		return echo.NewHTTPError(http.StatusInternalServerError, errTmpl)
	}
	return nil
}

func qs(s string) string {
	if s == "" {
		return ""
	}
	return fmt.Sprintf("?%s", s)
}

func previous(page int) int {
	if page == 1 {
		return 1
	}
	return page - 1
}

func next(page int, maxPage uint) int {
	max := int(maxPage)
	if page >= max {
		return max
	}
	return page + 1
}

// Pagi returns up to three page numbers for pagination links.
// The absolute numbers will always be in sequence except for returned
// values of zero, which should be skipped.
func Pagi(page int, maxPage uint) (int, int, int) {
	max := int(maxPage)
	switch max {
	case 0, 1, 2:
		return 0, 0, 0
	case 3:
		return 2, 0, 0
	case 4:
		return 2, 3, 0
	}
	a := page + -1
	b := page + 0
	c := page + 1
	if c > max {
		diff := c - max
		c = max - diff
		b = max - diff - 1
		a = max - diff - 2
		return a, b, c
	}
	if c == max {
		diff := c - max + 1
		c = max - diff
		b = max - diff - 1
		a = max - diff - 2
		return a, b, c
	}
	if a <= 1 {
		a = 2
		b = 3
		c = 4
	}
	return a, b, c
}

// Limit increases the limit value to stop unnecessary pagination of records,
// where the second page contains significantly fewer records than the first.
// Instead, all records are shown on a single page.
func Limit(count, limit int) int {
	if count > limit && count < limit+(limit/2) {
		return limit + (limit / 2)
	}
	return limit
}

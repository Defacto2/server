package html3

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/model/html3"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Navigate handles offset and record limit pagination.
type Navigate struct {
	Current  string // Current name of the active record query.
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
	zlog *zap.SugaredLogger
}

// All method lists every release.
func (s *sugared) All(c echo.Context) error {
	return s.List(c, Everything)
}

// Category lists the file records associated with the category tag that is provided by the ID param in the URL.
func (s *sugared) Category(c echo.Context) error {
	return s.List(c, BySection)
}

// Platform lists the file records associated with the platform tag that is provided by the ID param in the URL.
func (s *sugared) Platform(c echo.Context) error {
	return s.List(c, ByPlatform)
}

// Group lists the file records associated with the group that is provided by the ID param in the URL.
func (s *sugared) Group(c echo.Context) error {
	return s.List(c, ByGroup)
}

// Art lists the file records described as art are digital + pixel art files.
func (s *sugared) Art(c echo.Context) error {
	return s.List(c, AsArt)
}

// Documents lists the file records described as document + text art files.
func (s *sugared) Documents(c echo.Context) error {
	return s.List(c, AsDocument)
}

// Software lists the file records described as software files.
func (s *sugared) Software(c echo.Context) error {
	return s.List(c, AsSoftware)
}

// List all the records associated with the RecordsBy grouping.
func (s *sugared) List(c echo.Context, tt RecordsBy) error { //nolint:funlen
	start := helper.Latency()
	var id string
	switch tt {
	case BySection, ByPlatform:
		id = ID(c)
	default:
		id = c.Param("id")
	}
	// pagination offset and page number
	page := 1
	offset := strings.TrimPrefix(c.Param("offset"), "/")
	if offset != "" {
		// this permits blank offsets param but returns 404 for a /0 value
		page, _ = strconv.Atoi(offset)
		if page < 1 {
			return echo.NewHTTPError(http.StatusNotFound,
				fmt.Sprintf("Page %d of %s doesn't exist", page, tt))
		}
	}
	// query database to return records and statistics
	limit, count, byteSum, records, err := Query(c, tt, page)
	if err != nil {
		s.zlog.Warnf("%s query error: %s", tt, err)
		return echo.NewHTTPError(http.StatusServiceUnavailable, ErrConn)
	}
	if limit > 0 && count == 0 {
		return echo.NewHTTPError(http.StatusNotFound,
			fmt.Sprintf("The %s %q doesn't exist", tt, id))
	}
	// pagination maximum page number
	maxPage := uint(0)
	if limit > 0 {
		maxPage = helper.PageCount(count, limit)
		if page > int(maxPage) {
			return echo.NewHTTPError(http.StatusNotFound,
				fmt.Sprintf("Page %d of %d for %s doesn't exist", page, maxPage, tt))
		}
	}
	// pagination values
	current := strings.TrimPrefix(tt.String(), "html3_")
	switch tt {
	case BySection:
		current = fmt.Sprintf("category/%s", id)
	case ByPlatform:
		current = fmt.Sprintf("platform/%s", id)
	}
	navi := Navi(limit, page, maxPage, current, qs(c.QueryString()))
	navi.Link1, navi.Link2, navi.Link3 = Pagi(page, maxPage)
	// string based values for use in templates
	stat := fmt.Sprintf("%d files, %s", count, helper.ByteCountFloat(byteSum))
	title, desc := ListInfo(tt, current, id)
	err = c.Render(http.StatusOK, tt.String(), map[string]interface{}{
		"title":       title,
		"home":        "",
		"description": desc,
		"parent":      tt.Parent(),
		"stats":       stat,
		"sort":        Sorter(c.QueryString()),
		"records":     records,
		"latency":     fmt.Sprintf("%s.", time.Since(*start)),
		"navigate":    navi,
	})
	if err != nil {
		s.zlog.Errorf("%s: %s %d", ErrTmpl, err, tt)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

// ListInfo returns the title and description for the RecordsBy grouping.
func ListInfo(tt RecordsBy, current, id string) (string, string) {
	var desc string
	switch tt {
	case BySection, ByPlatform:
		key := tags.TagByURI(id)
		info := tags.Infos()[key]
		name := tags.Names()[key]
		desc = fmt.Sprintf("%s - %s.", name, info)
	case AsArt:
		desc = fmt.Sprintf("%s, %s.", "Digital + pixel art", textArt)
	case AsDocument:
		desc = fmt.Sprintf("%s, %s.", "Document + text art", textDoc)
	case AsSoftware:
		desc = fmt.Sprintf("%s, %s.", "Software", textSof)
	}
	title := fmt.Sprintf("%s/%s", title, current)
	if tt == ByGroup && id != "" {
		title = fmt.Sprintf("%s/%s", title, id)
	}
	return title, desc
}

// Query returns a slice of records based on the RecordsBy grouping.
// The three integers returned are the limit, the total count of records and the file sizes summed.
func Query(c echo.Context, tt RecordsBy, offset int) (int, int, int64, models.FileSlice, error) {
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		return queryErr("query connect db", err)
	}
	clause := c.QueryString()
	defer db.Close()
	switch tt {
	case Everything:
		return QueryEverything(ctx, db, clause, offset)
	case BySection:
		return QueryBySection(ctx, db, c, offset)
	case ByPlatform:
		return QueryByPlatform(ctx, db, c, offset)
	case ByGroup:
		return QueryByGroup(ctx, db, c)
	case AsArt:
		return QueryAsArt(ctx, db, clause, offset) // TODO: no pagination display
	case AsDocument:
		return QueryAsDocument(ctx, db, clause, offset)
	case AsSoftware:
		return QueryAsSoftware(ctx, db, clause, offset)
	}
	return 0, 0, 0, nil, nil // TODO error
}

func queryErr(info string, err error) (int, int, int64, models.FileSlice, error) {
	return 0, 0, 0, nil, fmt.Errorf("query %s: %w", info, err)
}

func statErr(info string, err error) (int, int, int64, models.FileSlice, error) {
	return 0, 0, 0, nil, fmt.Errorf("stat %s: %w", info, err)
}

func dbErr() (int, int, int64, models.FileSlice, error) {
	return 0, 0, 0, nil, ErrDB
}

// QueryEverything returns a slice of all the records, "Everything".
func QueryEverything(ctx context.Context, db *sql.DB, clause string, offset int) (int, int, int64, models.FileSlice, error) {
	if db == nil {
		return dbErr()
	}
	const limit = model.Maximum
	order := Clauses(clause)
	records, err := order.Everything(ctx, db, offset, limit)
	if err != nil {
		return queryErr("all releases:", err)
	}
	var stat model.Files
	if err = stat.Stat(ctx, db); err != nil {
		return statErr("all releases:", err)
	}
	total := stat.Count
	byteSum := int64(stat.Bytes)
	return limit, total, byteSum, records, nil
}

// QueryBySection returns a slice of all the records filtered by the section id, "by Category".
func QueryBySection(ctx context.Context, db *sql.DB, c echo.Context, offset int) (int, int, int64, models.FileSlice, error) {
	if db == nil {
		return dbErr()
	}
	const limit = model.Maximum
	order := Clauses(c.QueryString())
	id := ID(c)
	records, err := order.ByCategory(ctx, db, offset, limit, id)
	if err != nil {
		return queryErr("by category:", err)
	}
	total, err := model.CountByCategory(ctx, db, id)
	if err != nil {
		return statErr("total by category:", err)
	}
	byteSum, err := model.ByteCountByCategory(ctx, db, id)
	if err != nil {
		fmt.Printf("ID is using %q\n", id)
		fmt.Println(c.ParamNames(), c.ParamValues())
		return statErr("byte by category:", err)
	}
	return limit, int(total), byteSum, records, nil
}

// QueryByPlatform returns a slice of all the records filtered by the platform id, "by Platform and media".
func QueryByPlatform(ctx context.Context, db *sql.DB, c echo.Context, offset int) (int, int, int64, models.FileSlice, error) {
	if db == nil {
		return dbErr()
	}
	const limit = model.Maximum
	order := Clauses(c.QueryString())
	id := ID(c)
	records, err := order.ByPlatform(ctx, db, offset, limit, id)
	if err != nil {
		return queryErr("by platform:", err)
	}
	total, err := model.CountByPlatform(ctx, db, id)
	if err != nil {
		return statErr("total by platform:", err)
	}
	byteSum, err := model.ByteCountByPlatform(ctx, db, id)
	if err != nil {
		return statErr("bytes by platform:", err)
	}
	return limit, int(total), byteSum, records, nil
}

// QueryByGroup returns a slice of all the records filtered by the group id, "by Group".
// The group records do not use pagination limits or offsets.
func QueryByGroup(ctx context.Context, db *sql.DB, c echo.Context) (int, int, int64, models.FileSlice, error) {
	if db == nil {
		return dbErr()
	}
	order := Clauses(c.QueryString())
	id := c.Param("id")
	records, err := order.ByGroup(ctx, db, id)
	if err != nil {
		return queryErr("by group:", err)
	}
	total := len(records)
	byteSum, err := model.ByteCountByReleaser(ctx, db, id)
	// name := releaser.Clean(id)
	if err != nil {
		return statErr("bytes by group:", err)
	}
	return 0, total, byteSum, records, nil
}

// QueryAsArt returns a slice of all the records filtered by "Digital + pixel art".
func QueryAsArt(ctx context.Context, db *sql.DB, clause string, offset int) (int, int, int64, models.FileSlice, error) {
	if db == nil {
		return dbErr()
	}
	const limit = model.Maximum
	order := Clauses(clause)
	records, err := order.Art(ctx, db, offset, limit)
	if err != nil {
		return queryErr("as art:", err)
	}
	var stat html3.Arts
	if err := stat.Stat(ctx, db); err != nil {
		return statErr("as art:", err)
	}
	total := stat.Count
	byteSum := int64(stat.Bytes)
	return limit, total, byteSum, records, nil
}

// QueryAsDocument returns a slice of all the records filtered by "Document + text art".
func QueryAsDocument(ctx context.Context, db *sql.DB, clause string, offset int) (int, int, int64, models.FileSlice, error) {
	if db == nil {
		return dbErr()
	}
	const limit = model.Maximum
	order := Clauses(clause)
	records, err := order.Document(ctx, db, offset, limit)
	if err != nil {
		return queryErr("as document:", err)
	}
	var stat html3.Documents
	if err := stat.Stat(ctx, db); err != nil {
		return statErr("as document:", err)
	}
	total := stat.Count
	byteSum := int64(stat.Bytes)
	return limit, total, byteSum, records, nil
}

// QueryAsSoftware returns a slice of all the records filtered by "Software".
func QueryAsSoftware(ctx context.Context, db *sql.DB, clause string, offset int) (int, int, int64, models.FileSlice, error) {
	if db == nil {
		return dbErr()
	}
	const limit = model.Maximum
	order := Clauses(clause)
	records, err := order.Software(ctx, db, offset, limit)
	if err != nil {
		return queryErr("as software:", err)
	}
	var stat html3.Softwares
	if err := stat.Stat(ctx, db); err != nil {
		return statErr("as software:", err)
	}
	total := stat.Count
	byteSum := int64(stat.Bytes)
	return limit, total, byteSum, records, nil
}

// Navi returns a populated Navigate struct for pagination.
func Navi(limit, page int, maxPage uint, current, qs string) Navigate {
	return Navigate{
		Current:  current,
		Limit:    limit,
		Page:     page,
		PagePrev: previous(page),
		PageNext: next(page, maxPage),
		PageMax:  int(maxPage),
		QueryStr: qs,
	}
}

// qs returns a query string with a leading question mark.
func qs(s string) string {
	if s == "" {
		return ""
	}
	return fmt.Sprintf("?%s", s)
}

// previous returns the previous page number.
func previous(page int) int {
	if page == 1 {
		return 1
	}
	return page - 1
}

// next returns the next page number.
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
	const page1, page2, page3, page4 = 1, 2, 3, 4
	max := int(maxPage)
	switch max {
	case 0, page1, page2:
		return 0, 0, 0
	case page3:
		return page2, 0, 0
	case page4:
		return page2, page3, 0
	}
	a := page + -1
	b := page + 0
	c := page + 1
	if c > max {
		diff := c - max
		c = max - diff
		b = max - diff - page1
		a = max - diff - page2
		return a, b, c
	}
	if c == max {
		diff := c - max + page1
		c = max - diff
		b = max - diff - page1
		a = max - diff - page2
		return a, b, c
	}
	if a <= 1 {
		a = page2
		b = page3
		c = page4
	}
	return a, b, c
}

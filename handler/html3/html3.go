// Package html3 renders the html3 sub-route of the website.
// This generates pages for the website for browsing of the file database using HTML3 styled tables.
package html3

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/model/html3"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"
	"go.uber.org/zap"
)

// Sort and order the records by the column name.
const (
	NameAsc      = "C=N&O=A"     // Name ascending order.
	NameDes      = "C=N&O=D"     // Name descending order.
	PublAsc      = "C=D&O=A"     // Date published ascending order.
	PublDes      = "C=D&O=D"     // Date published descending order.
	PostAsc      = "C=P&O=A"     // Posted ascending order.
	PostDes      = "C=P&O=D"     // Posted descending order.
	SizeAsc      = "C=S&O=A"     // Size ascending order.
	SizeDes      = "C=S&O=D"     // Size descending order.
	DescAsc      = "C=I&O=A"     // Description ascending order.
	DescDes      = "C=I&O=D"     // Description descending order.
	Name    Sort = "Name"        // Sort records by the filename.
	Publish Sort = "Publish"     // Sort records by the published year, month and day.
	Posted  Sort = "Posted"      // Sort records by the record creation dated.
	Size    Sort = "Size"        // Sort records by the file size in byte units.
	Desc    Sort = "Description" // Sort the records by the title.
)

// Prefix is the root path of the HTML3 router group.
const Prefix = "/html3"

const (
	asc  = "A" // asc is order by ascending.
	desc = "D" // desc is order by descending.
)

var (
	ErrConn   = errors.New("the server cannot connect to the database")
	ErrDB     = errors.New("database value is nil")
	ErrPage   = errors.New("unknown records by type")
	ErrRoutes = errors.New("echo instance is nil")
	ErrSQL    = errors.New("database connection problem or a SQL error")
	ErrTag    = errors.New("no database query was for the tag")
	ErrTmpl   = errors.New("the server could not render the HTML template for this page")
	ErrZap    = errors.New("zap logger is nil")
)

// Clauses for ordering file record queries.
func Clauses(query string) html3.Order {
	switch strings.ToUpper(query) {
	case NameAsc: // Name ascending order should match the case.
		return html3.NameAsc
	case NameDes:
		return html3.NameDes
	case PublAsc:
		return html3.PublAsc
	case PublDes:
		return html3.PublDes
	case PostAsc:
		return html3.PostAsc
	case PostDes:
		return html3.PostDes
	case SizeAsc:
		return html3.SizeAsc
	case SizeDes:
		return html3.SizeDes
	case DescAsc:
		return html3.DescAsc
	case DescDes:
		return html3.DescDes
	default:
		return html3.NameAsc
	}
}

// Description returns a HTML3 friendly file description.
func Description(section, platform, brand, title null.String) string {
	return File{
		Section:  section.String,
		Platform: platform.String,
		GroupBy:  brand.String,
		Title:    title.String,
	}.Description()
}

// Error renders a custom HTTP error page for the HTML3 sub-group.
func Error(c echo.Context, err error) error {
	start := helper.Latency()
	code := http.StatusInternalServerError
	msg := "This is a server problem"
	var httpError *echo.HTTPError
	if errors.As(err, &httpError) {
		code = httpError.Code
		msg = fmt.Sprint(httpError.Message)
	}
	return c.Render(code, "html3_error", map[string]interface{}{
		"title":       fmt.Sprintf("%d error, there is a complication", code),
		"description": msg + ".",
		"latency":     time.Since(*start).String() + ".",
	})
}

// FileHref creates a URL to link to the file download of the ID.
func FileHref(logr *zap.SugaredLogger, id int64) string {
	if logr == nil {
		return ErrZap.Error()
	}
	href, err := url.JoinPath("/", "html3", "d",
		helper.ObfuscateID(id))
	if err != nil {
		logr.Error("FileHref ID %d could not be made into a valid URL: %s", err)
		return ""
	}
	return href
}

// FileLinkPad adds whitespace padding after the hyperlinked filename.
func FileLinkPad(width int, name null.String) string {
	if !name.Valid {
		return Leading(width)
	}
	return File{Filename: name.String}.FileLinkPad(width)
}

// Filename returns a truncated filename with to the w maximum width.
func Filename(width int, name null.String) string {
	return helper.TruncFilename(width, name.String)
}

// GlobTo returns the path to the template file.
func GlobTo(name string) string {
	const pathSeparator = "/"
	return strings.Join([]string{"view", "html3", name}, pathSeparator)
}

// ID returns the ID from the URL path.
// This is only used for the category and platform routes.
func ID(c echo.Context) string {
	x := strings.TrimSuffix(c.Path(), ":offset")
	s := strings.Split(x, "/")
	const expected = 4
	if len(s) != expected {
		return ""
	}
	return s[3]
}

// LeadFS formats the file size to the fixed-width length w value.
func LeadFS(width int, size int64) string {
	return File{Size: size}.LeadFS(width)
}

// LeadFSInt formats the file size to the fixed-width length w value.
func LeadFSInt(width, size int) string {
	return File{Size: int64(size)}.LeadFS(width)
}

// Leading repeats the number of space characters.
func Leading(count int) string {
	if count < 1 {
		return ""
	}
	return strings.Repeat(padding, count)
}

// LeadInt takes an int and returns it as a string, w characters wide with whitespace padding.
func LeadInt(width, i int) string {
	s := noValue
	if i > 0 {
		s = strconv.Itoa(i)
	}
	l := utf8.RuneCountInString(s)
	if l >= width {
		return s
	}
	count := width - l
	if count > maxPad {
		count = maxPad
	}
	return strings.Repeat(padding, count) + s
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
		return QueryAsArt(ctx, db, clause, offset)
	case AsDocument:
		return QueryAsDocument(ctx, db, clause, offset)
	case AsSoftware:
		return QueryAsSoftware(ctx, db, clause, offset)
	}
	return 0, 0, 0, nil, fmt.Errorf("%w: %d", ErrPage, tt)
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

// Sorter creates the query string for the sortable columns.
// Replacing the O key value with the opposite value, either A or D.
func Sorter(query string) map[string]string {
	s := Sortings()
	switch strings.ToUpper(query) {
	case NameAsc:
		s[Name] = desc
	case NameDes:
		s[Name] = asc
	case PublAsc:
		s[Publish] = desc
	case PublDes:
		s[Publish] = asc
	case PostAsc:
		s[Posted] = desc
	case PostDes:
		s[Posted] = asc
	case SizeAsc:
		s[Size] = desc
	case SizeDes:
		s[Size] = asc
	case DescAsc:
		s[Desc] = desc
	case DescDes:
		s[Desc] = asc
	default:
		// When no query is provided, it is assumed the records have been
		// ordered with Name ASC. So set DESC for the clickable Name link.
		s[Name] = desc
	}
	// to be usable in the template, convert the map keys into strings
	fix := make(map[string]string, len(s))
	for key, value := range s {
		fix[string(key)] = value
	}
	return fix
}

// Sortings are the name and order of columns that the records can be ordered by.
func Sortings() map[Sort]string {
	return map[Sort]string{
		Name:    asc,
		Publish: asc,
		Posted:  asc,
		Size:    asc,
		Desc:    asc,
	}
}

// Templates returns a map of the templates used by the HTML3 sub-group route.
func Templates(logr *zap.SugaredLogger, fs embed.FS) map[string]*template.Template {
	t := make(map[string]*template.Template)
	t["html3_index"] = index(logr, fs)
	t["html3_all"] = list(logr, fs)
	t["html3_art"] = list(logr, fs)
	t["html3_documents"] = list(logr, fs)
	t["html3_software"] = list(logr, fs)
	t["html3_groups"] = listGroups(logr, fs)
	t["html3_group"] = list(logr, fs)
	t[string(tag)] = listTags(logr, fs)
	t["html3_platform"] = list(logr, fs)
	t["html3_category"] = list(logr, fs)
	t["html3_error"] = httpErr(logr, fs)
	return t
}

// TemplateFuncMap are a collection of mapped functions that can be used in a template.
func TemplateFuncMap(logr *zap.SugaredLogger) template.FuncMap {
	return template.FuncMap{
		"byteInt":  LeadFSInt,
		"descript": Description,
		"fmtByte":  LeadFS,
		"fmtURI":   releaser.Link,
		"icon":     html3.Icon,
		"leading":  Leading,
		"leadInt":  LeadInt,
		"leadStr":  html3.LeadStr,
		"linkPad":  FileLinkPad,
		"linkFile": Filename,
		"publish":  html3.PublishedFW,
		"posted":   html3.Created,
		"linkHref": func(id int64) string {
			return FileHref(logr, id)
		},
		"metaByName": func(s string) tags.TagData {
			t, err := tagByName(s)
			if err != nil {
				logr.Errorw("tag", "error", err)
				return tags.TagData{}
			}
			return t
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s) //nolint:gosec
		},
	}
}

// Sort is the display name of column that can be used to sort and order the records.
type Sort string

func queryErr(info string, err error) (int, int, int64, models.FileSlice, error) {
	return 0, 0, 0, nil, fmt.Errorf("query %s: %w", info, err)
}

func statErr(info string, err error) (int, int, int64, models.FileSlice, error) {
	return 0, 0, 0, nil, fmt.Errorf("stat %s: %w", info, err)
}

func dbErr() (int, int, int64, models.FileSlice, error) {
	return 0, 0, 0, nil, ErrDB
}

// Package models contain the custom queries for the database that are not available using the ORM,
// as well as methods to interact with the query data.
package models

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/Defacto2/sceners"
	"github.com/Defacto2/server/postgres/models"
	"github.com/volatiletech/null/v8"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// https://github.com/volatiletech/sqlboiler#constants

type Count int // Count is the number of found files.

const (
	Art  int = iota // Art are digital + pixel art files.
	Doc             // Doc are document + text art files.
	Soft            // Soft are software files.
)

const (
	// Name
	NameAsc = "C=N&O=A"
	NameDes = "C=N&O=D"
	// Date published
	PublAsc = "C=D&O=A"
	PublDes = "C=D&O=D"
	// Posted
	PostAsc = "C=P&O=A"
	PostDes = "C=P&O=D"
	// Size
	SizeAsc = "C=S&O=A"
	SizeDes = "C=S&O=D"
	// Description
	DescAsc = "C=I&O=A"
	DescDes = "C=I&O=D"
)

// Counts caches the number of found files fetched from SQL queries.
var Counts = map[int]Count{
	Art:  0,
	Doc:  0,
	Soft: 0,
}

// ArtImagesCount counts the number of files that could be classified as digital or pixel art.
func ArtImagesCount(ctx context.Context, db *sql.DB) (int, error) {
	if c := Counts[Art]; c > 0 {
		return int(c), nil
	}
	c, err := models.Files(
		Where("platform = ?", "image"),
		Where("section != ?", "bbs")).Count(ctx, db)
	if err != nil {
		return -1, err
	}
	Counts[Art] = Count(c)
	return int(c), nil
}

// ByteCountByCategory sums the byte filesizes for all the files that match a category.
func ByteCountByCategory(s string, ctx context.Context, db *sql.DB) (int64, error) {
	i, err := models.Files(
		SQL("SELECT sum(filesize) FROM files WHERE section = $1",
			null.StringFrom(s)),
	).Count(ctx, db)
	if err != nil {
		return 0, err
	}
	return i, err
}

// ByteCountByPlatform sums the byte filesizes for all the files that match a category.
func ByteCountByPlatform(s string, ctx context.Context, db *sql.DB) (int64, error) {
	i, err := models.Files(
		SQL("SELECT sum(filesize) FROM files WHERE platform = $1",
			null.StringFrom(s)),
	).Count(ctx, db)
	if err != nil {
		return 0, err
	}
	return i, err
}

// DocumentCount counts the number of files that could be classified as document or text art.
func DocumentCount(ctx context.Context, db *sql.DB) (int, error) {
	if c := Counts[Doc]; c > 0 {
		return int(c), nil
	}
	c, err := models.Files(
		Where("platform = ?", "ansi"),
		Or("platform = ?", "text"),
		Or("platform = ?", "textamiga"),
		Or("platform = ?", "pdf")).Count(ctx, db)
	if err != nil {
		return -1, err
	}
	Counts[Doc] = Count(c)
	return int(c), nil
}

func Download(id int, ctx context.Context, db *sql.DB) (*models.File, error) {
	file, err := models.Files(models.FileWhere.ID.EQ(int64(id))).One(ctx, db)
	if err != nil {
		return &models.File{}, err
	}
	return file, err
}

// FilesByCategory returns all the files that match a category tag.
func FilesByCategory(s, query string, ctx context.Context, db *sql.DB) (models.FileSlice, error) {
	x := null.StringFrom(s)
	return models.Files(Where("section = ?", x), OrderBy(Clauses(query))).All(ctx, db)
}

// FilesByPlatform returns all the files that match a platform tag.
func FilesByPlatform(s, query string, ctx context.Context, db *sql.DB) (models.FileSlice, error) {
	x := null.StringFrom(s)
	return models.Files(Where("platform = ?", x), OrderBy(Clauses(query))).All(ctx, db)
}

// FilesByGroup returns all the files that match an exact group name.
func FilesByGroup(s, query string, ctx context.Context, db *sql.DB) (models.FileSlice, error) {
	x := null.StringFrom(s)
	return models.Files(Where("group_brand_for = ?", x), OrderBy(Clauses(query))).All(ctx, db)
}

func Clauses(query string) string {
	const a, d = "asc", "desc"
	ca := models.FileColumns.Createdat
	dy := models.FileColumns.DateIssuedYear
	dm := models.FileColumns.DateIssuedMonth
	dd := models.FileColumns.DateIssuedDay
	fn := models.FileColumns.Filename
	fs := models.FileColumns.Filesize
	rt := models.FileColumns.RecordTitle
	switch strings.ToUpper(query) {
	case NameAsc:
		return fmt.Sprintf("%s %s", fn, a)
	case NameDes:
		return fmt.Sprintf("%s %s", fn, d)
	case PublAsc:
		return fmt.Sprintf("%s %s, %s %s, %s %s", dy, a, dm, a, dd, a)
	case PublDes:
		return fmt.Sprintf("%s %s, %s %s, %s %s", dy, d, dm, d, dd, d)
	case PostAsc:
		return fmt.Sprintf("%s %s", ca, a)
	case PostDes:
		return fmt.Sprintf("%s %s", ca, d)
	case SizeAsc:
		return fmt.Sprintf("%s %s", fs, a)
	case SizeDes:
		return fmt.Sprintf("%s %s", fs, d)
	case DescAsc:
		return fmt.Sprintf("%s %s", rt, a)
	case DescDes:
		return fmt.Sprintf("%s %s", rt, d)
	default:
		return fmt.Sprintf("%s %s", fn, a)
	}
}

// Format the group name for printing.
func Grouper(s string) string {
	l := strings.TrimSpace(s)
	l = strings.ToLower(l)

	re := regexp.MustCompile(`-ampersand-`)
	l = re.ReplaceAllString(l, ` & `)

	re = regexp.MustCompile(`-`)
	l = re.ReplaceAllString(l, ` `)

	re = regexp.MustCompile(`_`)
	l = re.ReplaceAllString(l, `-`)

	sentence := []string{}

	const spaceSubstitute = ","
	words := strings.Split(l, spaceSubstitute)

	c := cases.Title(language.English)

	for _, word := range words {
		re = regexp.MustCompile(`iso`)
		word = re.ReplaceAllString(word, `ISO`)
		re = regexp.MustCompile(`xxx`)
		word = re.ReplaceAllString(word, `XXX`)
		re = regexp.MustCompile(`\*`)
		word = re.ReplaceAllString(word, `, `)
		switch word {
		case "in", "of", "or":
			sentence = append(sentence, strings.ToLower(word))
		default:
			sentence = append(sentence, c.String(word))
		}
	}
	return strings.Join(sentence, " ")
}

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

type Scener struct {
	URI   string
	Name  string
	Count int
}

type Sceners map[string]Scener

func Groups(ctx context.Context, db *sql.DB) (Sceners, error) {
	// TODO: create libs using base: https://github.com/Defacto2/df2/blob/937cf38cddda8a38091258ae62f2db31e1b672cf/pkg/groups/internal/rename/rename.go#L83
	//q, err := models.Files(SQL("select distinct(LOWER(TRIM(files.group_brand_for))) FROM files")).Exec(ctx, db)
	q, err := models.Files(
		Select(models.FileColumns.GroupBrandFor),
		Distinct(models.FileColumns.GroupBrandFor),
	).All(ctx, db)
	if err != nil {
		return nil, err
	}
	// count := 1
	// for _, g := range q {
	// 	if s := strings.TrimSpace(g.GroupBrandFor.String); len(s) > 0 {
	// 		count++
	// 	}
	// }
	//s := make([]string, count)
	m := make(map[string]Scener)
	for i, g := range q {
		if x := strings.TrimSpace(g.GroupBrandFor.String); len(x) > 0 {

			uri := GroupForURL(x)

			fmt.Println("--->", sceners.Cleaner(x))

			n := Scener{Name: x, URI: uri}
			if i < 250 {
				// TODO: defer and save to a global?
				n.Count, err = CountGroup(x, ctx, db)
				if err != nil {
					fmt.Println(err)
				}
			}
			m[uri] = n
			//s = append(s, n)
		}
	}
	return m, nil
}

// 9425
// 9155
func GroupsTotalCount(ctx context.Context, db *sql.DB) (int, error) {
	c, err := models.Files(SQL("select count(distinct(LOWER(TRIM(files.group_brand_for)))) FROM files")).Count(ctx, db)
	if err != nil {
		fmt.Println(err)
		return -1, err
	}
	return int(c), nil
}

func CountGroup(s string, ctx context.Context, db *sql.DB) (int, error) {
	// TODO: in postgresql, when comparing lowercase in queries, any column indexes are void
	c, err := models.Files(
		Select(models.FileColumns.GroupBrandFor),
		Where("group_brand_for = ?", s),
	).Count(ctx, db)
	if err != nil {
		return -1, err
	}
	return int(c), nil
}

// SoftwareCount counts the number of files that could be classified as software.
func SoftwareCount(ctx context.Context, db *sql.DB) (int, error) {
	if c := Counts[Soft]; c > 0 {
		return int(c), nil
	}
	c, err := models.Files(
		Where("platform = ?", "java"),
		Or("platform = ?", "linux"),
		Or("platform = ?", "dos"),
		Or("platform = ?", "php"),
		Or("platform = ?", "windows")).Count(ctx, db)
	if err != nil {
		return -1, err
	}
	Counts[Soft] = Count(c)
	return int(c), nil
}

package html3_test

import (
	"context"
	"database/sql"
	"embed"
	"net/http"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/html3"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/null/v8"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

// checked in Sep 26, test coverage was great at 75%+

func TestFilename(t *testing.T) {
	t.Parallel()

	name := null.StringFrom("filename.txt")

	s := html3.Filename(2, name)
	be.Equal(t, s, ".txt")

	s = html3.Filename(6, name)
	be.Equal(t, s, "f..txt")

	s = html3.Filename(10, name)
	be.Equal(t, s, "filen..txt")
}

func TestRoutes(t *testing.T) {
	t.Parallel()

	e := echo.New()
	sl := logs.Discard()

	var db sql.DB
	got := html3.Route(sl, e, &db)
	be.True(t, got != nil)
}

func TestGlobTo(t *testing.T) {
	t.Parallel()

	got := html3.GlobTo("file")
	be.Equal(t, got, "view/html3/file")
}

func TestTemplatesPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected html3 templates to panic")
		}
	}()

	_ = html3.Templates(t.Context(), nil, nil, embed.FS{})
}

func TestTemplates(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	db := testutil.DB(t)
	fsys := testutil.ProjectFS(t)

	_ = html3.Templates(t.Context(), sl, db, fsys)
}

func TestError(t *testing.T) {
	t.Parallel()

	c := testutil.NewContext(t, "")
	got := html3.Error(c, nil)
	be.Err(t, got)

	got = html3.Error(c, html3.ErrPage)
	be.Err(t, got)

	got = html3.Error(c, echo.ErrNotFound)
	be.Err(t, got)
}

func TestID(t *testing.T) {
	t.Parallel()

	c := testutil.NewContext(t, "")
	got := html3.ID(c)
	be.Equal(t, got, "")
}

func TestLeadFS(t *testing.T) {
	t.Parallel()

	got := html3.LeadFS(0, null.Int64From(0))
	be.Equal(t, got, "0B")

	got = html3.LeadFS(10, null.Int64From(3))
	wants := strings.Repeat(" ", 8) + "3B"
	be.Equal(t, got, wants)
}

func TestLeadInt(t *testing.T) {
	t.Parallel()

	x := html3.LeadInt(0, 0)
	be.Equal(t, x, "-")

	x = html3.LeadInt(10, 3)
	wants := strings.Repeat(" ", 9) + "3"
	be.Equal(t, x, wants)
}

func TestQuery(t *testing.T) {
	t.Parallel()

	c := testutil.NewContext(t, "")
	ctx := t.Context()

	x, y, z, fs, err := html3.Everything.Query(ctx, c, nil, -1)
	be.True(t, x == 0)
	be.True(t, y == 0)
	be.True(t, z == 0)
	be.Equal(t, len(fs), 0)
	be.Err(t, err)

	db := testutil.DB(t)

	pathValues := echo.PathValues{
		{Name: "id", Value: "defacto2"},
		{Name: "limit", Value: "2"},
	}
	c = testutil.NewPath(t, "/html3", pathValues)
	queries := []func(context.Context, *echo.Context, *sql.DB, int) (int, int, int64, models.FileSlice, error){
		html3.Everything.Query,
		html3.AsArt.Query,
		html3.AsDocument.Query,
		html3.AsSoftware.Query,
		// html3.BySection.Query,
		// html3.ByPlatform.Query,
		html3.ByGroup.Query,
	}
	for _, fn := range queries {
		x, y, z, fs, err = fn(ctx, c, db, -1)
		be.True(t, x == 0 || x == model.Maximum)
		be.True(t, y > 0)
		be.True(t, z > 0)
		be.True(t, len(fs) > 0)
		be.Err(t, err, nil)
	}

	//c = testutil.NewContext(t, "/platform/ansi")
	// x, y, z, fs, err = html3.BySection.Query(ctx, c, db, -1)
	// be.True(t, x == 0 || x == model.Maximum)
	// be.True(t, y > 0)
	// be.True(t, z > 0)
	// be.True(t, len(fs) > 0)
	// be.Err(t, err, nil)
}

func TestListInfo(t *testing.T) {
	t.Parallel()

	a, b := html3.Everything.ListInfo("", "")
	be.Equal(t, a, "Index of /html3/")
	be.Equal(t, b, "")

	a, b = html3.ByGroup.ListInfo("aaa", "bbb")
	be.Equal(t, a, "Index of /html3/bbb")
	be.Equal(t, b, "")

	a, b = html3.BySection.ListInfo("abc", "xyz")
	be.Equal(t, a, "Index of /html3/abc")
	be.Equal(t, b, "n/a - n/a.")

	a, b = html3.AsSoftware.ListInfo("abc", "xyz")
	be.Equal(t, a, "Index of /html3/abc")
	be.Equal(t, b, "Software, applications and programs for any platform.")
}

func TestRecordsBy(t *testing.T) {
	t.Parallel()

	by := html3.Everything
	be.Equal(t, by.String(), "html3_all")
	be.Equal(t, by.Parent(), "")

	by = html3.AsSoftware
	be.Equal(t, by.String(), "html3_software")
	be.Equal(t, by.Parent(), "")

	by = html3.BySection
	be.Equal(t, by.String(), "html3_category")
	be.Equal(t, by.Parent(), "categories")
}

func TestClauses(t *testing.T) {
	t.Parallel()

	tests := [...]string{
		html3.NameAsc,
		html3.NameDes,
		html3.PublAsc,
		html3.PublDes,
		html3.PostAsc,
		html3.PostDes,
		html3.SizeAsc,
		html3.SizeDes,
		html3.DescAsc,
		html3.DescDes,
	}

	for i, s := range tests {
		be.Equal(t, int(html3.Clauses(s)), i)
	}

	be.Equal(t, int(html3.Clauses("")),
		int(html3.Clauses(html3.NameAsc)))
}

func TestSorter(t *testing.T) {
	t.Parallel()

	tests := [...]string{
		html3.NameAsc,
		html3.NameDes,
		html3.PublAsc,
		html3.PublDes,
		html3.PostAsc,
		html3.PostDes,
		html3.SizeAsc,
		html3.SizeDes,
		html3.DescAsc,
		html3.DescDes,
	}

	const a, d = "A", "D"
	for _, s := range tests {
		switch s {
		case html3.NameAsc:
			be.Equal(t, html3.Sorter(s)[string(html3.Name)], d)
		case html3.NameDes:
			be.Equal(t, html3.Sorter(s)[string(html3.Name)], a)
		case html3.PublAsc:
			be.Equal(t, html3.Sorter(s)[string(html3.Publish)], d)
		case html3.PublDes:
			be.Equal(t, html3.Sorter(s)[string(html3.Publish)], a)
		case html3.PostAsc:
			be.Equal(t, html3.Sorter(s)[string(html3.Posted)], d)
		case html3.PostDes:
			be.Equal(t, html3.Sorter(s)[string(html3.Posted)], a)
		case html3.SizeAsc:
			be.Equal(t, html3.Sorter(s)[string(html3.Size)], d)
		case html3.SizeDes:
			be.Equal(t, html3.Sorter(s)[string(html3.Size)], a)
		case html3.DescAsc:
			be.Equal(t, html3.Sorter(s)[string(html3.Desc)], d)
		case html3.DescDes:
			be.Equal(t, html3.Sorter(s)[string(html3.Desc)], a)
		}
	}
}

func TestFile_Description(t *testing.T) {
	t.Parallel()

	type fields struct {
		Title    string
		GroupBy  string
		Section  string
		Platform string
	}

	const (
		x = "Hello world"
		g = "Defacto2"
		s = "intro"
		p = "dos"
		m = "magazine"
	)

	tests := []struct {
		name   string
		fields fields
		expect string
	}{
		{"empty", fields{Title: "", GroupBy: "", Section: "", Platform: ""}, ""},
		{"only title", fields{Title: x, GroupBy: "", Section: "", Platform: ""}, ""},
		{"req group", fields{Title: x, GroupBy: "", Section: "", Platform: p}, ""},
		{"default", fields{Title: x, GroupBy: g, Section: "", Platform: ""}, "Hello world from Defacto2."},
		{"with platform", fields{Title: x, GroupBy: g, Section: "", Platform: p}, "Hello world from Defacto2 for Dos."},
		{"no title", fields{Title: "", GroupBy: g, Section: "", Platform: p}, "A release from Defacto2 for Dos."},
		{"magazine", fields{Title: "1", GroupBy: g, Section: m, Platform: p}, "Defacto2 issue 1 for Dos."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f := html3.File{
				Filename: "",
				Title:    tt.fields.Title,
				GroupBy:  tt.fields.GroupBy,
				Section:  tt.fields.Section,
				Platform: tt.fields.Platform,
				Size:     0,
			}
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				be.Equal(t, f.Description(), tt.expect)
			})
		})
	}
}

func TestDescription(t *testing.T) {
	t.Parallel()

	empty := null.String{String: "", Valid: false}
	s := html3.Description(empty, empty, empty, empty)
	be.Equal(t, s, "")

	s = html3.Description(
		null.StringFrom("intro"),
		null.StringFrom("dos"),
		null.StringFrom("Defacto2"),
		null.StringFrom("Hello world"),
	)
	be.Equal(t, "Hello world from Defacto2 for Dos.", s)
}

func TestFileHref(t *testing.T) {
	t.Parallel()

	got := html3.FileHref(nil, 0)
	want := "argument 0: slog logger pointer is nil"
	be.Equal(t, got, want)

	sl := logs.Discard()
	got = html3.FileHref(sl, 0)
	be.Equal(t, got, "/html3/d/0")
}

func TestFileLinkPad(t *testing.T) {
	t.Parallel()

	n := null.String{String: "", Valid: false}
	s := html3.FileLinkPad(0, n)
	be.Equal(t, s, "")

	s = html3.FileLinkPad(20, null.StringFrom("file"))
	be.Equal(t, "                ", s)
}

func TestFormattings(t *testing.T) {
	t.Parallel()

	be.Equal(t,
		html3.File{Filename: "", Title: "", GroupBy: "", Section: "", Platform: "", Size: 0}.FileLinkPad(0), "", "empty")
	be.Equal(t,
		html3.File{Filename: "", Title: "", GroupBy: "", Section: "", Platform: "", Size: 0}.FileLinkPad(4), "    ", "4 pads")
	be.Equal(t,
		html3.File{Filename: "file", Title: "", GroupBy: "", Section: "", Platform: "", Size: 0}.FileLinkPad(6), "  ", "2 pads")
	be.Equal(t,
		html3.File{Filename: "file.txt", Title: "", GroupBy: "", Section: "", Platform: "", Size: 0}.FileLinkPad(6), "", "too big")
	be.Equal(t,
		html3.File{Filename: "", Title: "", GroupBy: "", Section: "", Platform: "", Size: 0}.LeadFS(0), "0B", "0 size")
	be.Equal(t,
		html3.File{Filename: "", Title: "", GroupBy: "", Section: "", Platform: "", Size: 1}.LeadFS(3), " 1B", "1 size")
	be.Equal(t,
		html3.File{Filename: "", Title: "", GroupBy: "", Section: "", Platform: "", Size: 1000}.LeadFS(0), "1000B", "1000 size")
	be.Equal(t,
		html3.File{Filename: "", Title: "", GroupBy: "", Section: "", Platform: "", Size: 1024}.LeadFS(0), "1k", "1k size")
	be.Equal(t,
		html3.File{Filename: "", Title: "", GroupBy: "", Section: "", Platform: "", Size: int64(1024 * 1024)}.LeadFS(0), "1M", "1MB size")
	be.Equal(t,
		html3.File{Filename: "", Title: "", GroupBy: "", Section: "", Platform: "", Size: int64(1024 * 1024 * 1024)}.LeadFS(0), "1G", "1GB size")
	be.Equal(t,
		html3.File{Filename: "", Title: "", GroupBy: "", Section: "", Platform: "", Size: int64(1024 * 1024 * 1024 * 1024)}.LeadFS(0), "1T", "1TB size")
	be.Equal(t,
		html3.LeadInt(0, 1), "1")
	be.Equal(t,
		html3.LeadInt(1, 1), "1")
	be.Equal(t,
		html3.LeadInt(3, 1), "  1")
	be.True(t,
		html3.File{Filename: "", Title: "", GroupBy: "", Section: "", Platform: "java", Size: 0}.IsOS())
	be.Equal(t,
		html3.File{Filename: "", Title: "", GroupBy: "", Section: "", Platform: "java", Size: 0}.OS(), " for Java")
}

func TestPagi(t *testing.T) {
	t.Parallel()

	type args struct {
		page    int
		maxPage int
	}
	tests := []struct {
		name  string
		args  args
		want  int
		want1 int
		want2 int
	}{
		{"empty", args{page: 0, maxPage: 0}, 0, 0, 0},
		{"1 page", args{page: 1, maxPage: 1}, 0, 0, 0},
		{"2 pages", args{page: 1, maxPage: 2}, 0, 0, 0},
		{"3 pages", args{page: 1, maxPage: 3}, 2, 0, 0},
		{"4 pages", args{page: 1, maxPage: 4}, 2, 3, 0},
		{"start of many pages", args{page: 2, maxPage: 10}, 2, 3, 4},
		{"middle of many pages", args{page: 5, maxPage: 10}, 4, 5, 6},
		{"near end of many pages", args{9, 10}, 7, 8, 9},
		{"last of many pages", args{10, 10}, 7, 8, 9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, got1, got2 := html3.Pagi(tt.args.page, tt.args.maxPage)
			be.Equal(t, got, tt.want)
			be.Equal(t, got1, tt.want1)
			be.Equal(t, got2, tt.want2)
		})
	}
}

func TestNavi(t *testing.T) {
	t.Parallel()

	limit := 10
	page := 2
	maxPage := 5
	current := "current"
	qs := "query"

	expected := html3.Navigate{
		Current:  current,
		QueryStr: qs,
		Limit:    limit,
		Link1:    0,
		Link2:    0,
		Link3:    0,
		Page:     page,
		PagePrev: 1,
		PageNext: 3,
		PageMax:  5,
	}

	result := html3.Navi(limit, page, maxPage, current, qs)

	if result != expected {
		t.Errorf("Navi(%d, %d, %d, %s, %s) = %v; want %v", limit, page, maxPage, current, qs, result, expected)
	}
}

func testTmpl(t *testing.T, got error) {
	t.Helper()

	be.Err(t, got)
	httpErr, ok := got.(*echo.HTTPError)
	be.True(t, ok)
	be.Equal(t, httpErr.Code, http.StatusInternalServerError)
	be.Equal(t, httpErr.Message, "html3: cannot render the template") // this is expected
}

func TestAll(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	sl := logs.Discard()
	c := testutil.NewContext(t, "")
	got := html3.All(sl, c, db)
	testTmpl(t, got)
}

func TestCategories(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	c := testutil.NewContext(t, "")
	got := html3.Categories(sl, c)
	testTmpl(t, got)
}

func TestGroups(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	sl := logs.Discard()
	c := testutil.NewContext(t, "")
	got := html3.Groups(sl, c, db)
	testTmpl(t, got)
}

func TestIndex(t *testing.T) {
	t.Parallel()

	db := testutil.DB(t)
	sl := logs.Discard()
	c := testutil.NewContext(t, "")
	got := html3.Index(sl, c, db)
	testTmpl(t, got)
}

func TestPlatforms(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	c := testutil.NewContext(t, "")
	got := html3.Platforms(sl, c)
	testTmpl(t, got)
}

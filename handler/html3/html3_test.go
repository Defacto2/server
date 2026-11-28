package html3_test

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/html3"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/null/v8"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

type mapRenderer struct {
	tmpls map[string]*template.Template
}

func (r mapRenderer) Render(_ *echo.Context, w io.Writer, name string, data any) error {
	t, ok := r.tmpls[name]
	if !ok || t == nil {
		return html3.ErrTmpl
	}
	return t.ExecuteTemplate(w, "layout", data)
}

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

	// Moved-permanently redirects.
	redirects := []struct {
		from string
		to   string
	}{
		{"/html3/index", "/html3"},
		{"/html3/categories/index", "/html3/categories"},
		{"/html3/platforms/index", "/html3/platforms"},
	}
	for _, tc := range redirects {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.from, nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		be.Equal(t, rec.Code, http.StatusMovedPermanently)
		be.Equal(t, rec.Header().Get(echo.HeaderLocation), tc.to)
	}

	// Unknown /html3/:uri hits custom404.
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/html3/nonexistent", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	be.Equal(t, rec.Code, http.StatusNotFound)
}

func TestGlobTo(t *testing.T) {
	t.Parallel()

	got := html3.GlobTo("file")
	be.Equal(t, got, "view/html3/file")
}

func TestTemplates(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	db := testutil.DB(t)
	fsys := testutil.ProjectFS(t)

	_, err := html3.Templates(t.Context(), nil, nil, embed.FS{})
	be.Err(t, err)

	_, err = html3.Templates(t.Context(), sl, db, fsys)
	be.Err(t, err, nil)
}

func TestError(t *testing.T) {
	t.Parallel()

	be.Err(t, html3.Error(nil, nil))

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

	be.Equal(t, html3.ID(nil), "")

	c := testutil.NewContext(t, "")
	be.Equal(t, html3.ID(c), "")

	c.SetPath("/html3/category/bbs")
	be.Equal(t, html3.ID(c), "bbs")

	c.SetPath("/html3/platform/dos:offset")
	be.Equal(t, html3.ID(c), "dos")
}

func TestLeadFS(t *testing.T) {
	t.Parallel()

	got := html3.LeadFS(0, null.Int64From(0))
	be.Equal(t, got, "0B")

	got = html3.LeadFS(10, null.Int64From(3))
	wants := strings.Repeat(" ", 8) + "3B"
	be.Equal(t, got, wants)

	be.Equal(t, html3.LeadFSInt(0, 0), "0B")
	be.Equal(t, html3.LeadFSInt(6, 1024), "    1k")
}

func TestLeadInt(t *testing.T) {
	t.Parallel()

	x := html3.LeadInt(0, 0)
	be.Equal(t, x, "-")

	x = html3.LeadInt(10, 3)
	wants := strings.Repeat(" ", 9) + "3"
	be.Equal(t, x, wants)
}

func TestQueryOfflineGuards(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	c := testutil.NewContext(t, "/html3")

	_, _, _, _, err := html3.RecordsBy(99).Query(ctx, c, nil, 1)
	be.Err(t, err, html3.ErrPage)
	_, _, _, _, err = html3.Everything.Query(ctx, nil, nil, 1)
	be.Err(t, err)

	_, _, _, _, err = html3.QueryAsArt(ctx, nil, "", 1)
	be.Err(t, err)
	_, _, _, _, err = html3.QueryAsDocument(ctx, nil, "", 1)
	be.Err(t, err)
	_, _, _, _, err = html3.QueryAsSoftware(ctx, nil, "", 1)
	be.Err(t, err)
	_, _, _, _, err = html3.QueryByGroup(ctx, nil, c)
	be.Err(t, err)
	_, _, _, _, err = html3.QueryBySection(ctx, nil, c, 1)
	be.Err(t, err)
	_, _, _, _, err = html3.QueryBySection(ctx, nil, nil, 1)
	be.Err(t, err)
	_, _, _, _, err = html3.QueryByPlatform(ctx, nil, c, 1)
	be.Err(t, err)
	_, _, _, _, err = html3.QueryByPlatform(ctx, nil, nil, 1)
	be.Err(t, err)
	_, _, _, _, err = html3.QueryEverything(ctx, nil, "", 1)
	be.Err(t, err)
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
	queries := []struct {
		name string
		fn   func(context.Context, *echo.Context, *sql.DB, int) (int, int, int64, models.FileSlice, error)
	}{
		{"everything", html3.Everything.Query},
		{"as art", html3.AsArt.Query},
		{"as document", html3.AsDocument.Query},
		{"as software", html3.AsSoftware.Query},
		{"by group", html3.ByGroup.Query},
	}
	for _, q := range queries {
		t.Run(q.name, func(t *testing.T) {
			t.Parallel()
			x, y, z, fs, err := q.fn(ctx, c, db, -1)
			be.True(t, x == 0 || x == model.Maximum)
			be.True(t, y > 0)
			be.True(t, z > 0)
			be.True(t, len(fs) > 0)
			be.Err(t, err, nil)
		})
	}
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

	a, b = html3.ByPlatform.ListInfo("platform/dos", "dos")
	be.Equal(t, a, "Index of /html3/platform/dos")
	be.True(t, strings.Contains(b, "Dos"))

	a, b = html3.AsArt.ListInfo("art", "")
	be.Equal(t, a, "Index of /html3/art")
	be.True(t, strings.HasPrefix(b, "Digital + pixel art, "))

	a, b = html3.AsDocument.ListInfo("documents", "")
	be.Equal(t, a, "Index of /html3/documents")
	be.True(t, strings.HasPrefix(b, "Document + text art, "))

	a, b = html3.AsSoftware.ListInfo("abc", "xyz")
	be.Equal(t, a, "Index of /html3/abc")
	be.Equal(t, b, "Software, applications and programs for any platform.")
}

func TestRecordsBy(t *testing.T) {
	t.Parallel()

	cases := []struct {
		by     html3.RecordsBy
		str    string
		parent string
	}{
		{html3.Everything, "html3_all", ""},
		{html3.BySection, "html3_category", "categories"},
		{html3.ByPlatform, "html3_platform", "platforms"},
		{html3.ByGroup, "html3_group", "groups"},
		{html3.AsArt, "html3_art", ""},
		{html3.AsDocument, "html3_documents", ""},
		{html3.AsSoftware, "html3_software", ""},
		{html3.RecordsBy(-1), "", ""},
		{html3.RecordsBy(99), "", ""},
	}
	for _, tc := range cases {
		be.Equal(t, tc.by.String(), tc.str)
		be.Equal(t, tc.by.Parent(), tc.parent)
	}
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
	be.Equal(t, html3.Sorter("")[string(html3.Name)], d)

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
			be.Equal(t, f.Description(), tt.expect)
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
	be.Equal(t, s, "Hello world from Defacto2 for Dos.")
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
	be.Equal(t, s, "                ")
}

func TestFormattings(t *testing.T) {
	t.Parallel()

	be.Equal(t, html3.File{}.FileLinkPad(0), "")
	be.Equal(t, html3.File{}.FileLinkPad(4), "    ")
	be.Equal(t, html3.File{Filename: "file"}.FileLinkPad(6), "  ")
	be.Equal(t, html3.File{Filename: "file.txt"}.FileLinkPad(6), "")
	be.Equal(t, html3.File{}.LeadFS(0), "0B")
	be.Equal(t, html3.File{Size: 1}.LeadFS(3), " 1B")
	be.Equal(t, html3.File{Size: 1000}.LeadFS(0), "1000B")
	be.Equal(t, html3.File{Size: 1024}.LeadFS(0), "1k")
	be.Equal(t, html3.File{Size: int64(1024 * 1024)}.LeadFS(0), "1M")
	be.Equal(t, html3.File{Size: int64(1024 * 1024 * 1024)}.LeadFS(0), "1G")
	be.Equal(t, html3.File{Size: int64(1024 * 1024 * 1024 * 1024)}.LeadFS(0), "1T")
	be.Equal(t, html3.LeadInt(0, 1), "1")
	be.Equal(t, html3.LeadInt(1, 1), "1")
	be.Equal(t, html3.LeadInt(3, 1), "  1")
	be.True(t, html3.File{Platform: "java"}.IsOS())
	be.True(t, !html3.File{Platform: "unknown"}.IsOS())
	be.Equal(t, html3.File{Platform: "java"}.OS(), " for Java")
	be.Equal(t, html3.File{Platform: "unknown"}.OS(), "")
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
		{"beyond last page", args{20, 10}, 7, 8, 9},
		{"negative maxPage", args{1, -5}, 0, 0, 0},
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

	expected := html3.Navigate{
		Current:  "current",
		QueryStr: "query",
		Limit:    10,
		Link1:    0,
		Link2:    0,
		Link3:    0,
		Page:     2,
		PagePrev: 1,
		PageNext: 3,
		PageMax:  5,
	}
	be.Equal(t, html3.Navi(10, 2, 5, "current", "query"), expected)

	first := html3.Navi(10, 1, 5, "all", "")
	be.Equal(t, first.PagePrev, 1)
	be.Equal(t, first.PageNext, 2)

	last := html3.Navi(10, 5, 5, "all", "")
	be.Equal(t, last.PagePrev, 4)
	be.Equal(t, last.PageNext, 5)
}

func testTmpl(t *testing.T, got error) {
	t.Helper()

	be.Err(t, got)
	httpErr, ok := errors.AsType[*echo.HTTPError](got)
	be.True(t, ok)
	be.Equal(t, httpErr.Code, http.StatusInternalServerError)
	be.Equal(t, httpErr.Message, "html3: cannot render the template")
}

func TestSugaredOffline(t *testing.T) {
	t.Parallel()

	sl := logs.Discard()
	dummyDB := &sql.DB{}

	// Nil-guard paths across all sugared handlers.
	be.Err(t, html3.Categories(nil, nil))
	be.Err(t, html3.Platforms(nil, nil))
	be.Err(t, html3.Groups(nil, nil, nil))
	be.Err(t, html3.Groups(sl, testutil.NewContext(t, "/html3/groups"), nil))
	be.Err(t, html3.Index(nil, nil, nil))
	be.Err(t, html3.List(nil, nil, nil, html3.Everything))
	be.Err(t, html3.All(nil, nil, nil))
	be.Err(t, html3.Art(nil, nil, nil))
	be.Err(t, html3.Category(nil, nil, nil))
	be.Err(t, html3.Documents(nil, nil, nil))
	be.Err(t, html3.Group(nil, nil, nil))
	be.Err(t, html3.Platform(nil, nil, nil))
	be.Err(t, html3.Software(nil, nil, nil))

	// Invalid page offset returns 404 before touching the DB.
	cZero := testutil.NewPath(t, "/html3/groups/0", echo.PathValues{{Name: "offset", Value: "/0"}})
	err := html3.Groups(sl, cZero, dummyDB)
	httpErr, ok := errors.AsType[*echo.HTTPError](err)
	be.True(t, ok)
	be.Equal(t, httpErr.Code, http.StatusNotFound)

	cListZero := testutil.NewPath(t, "/html3/all/0", echo.PathValues{{Name: "offset", Value: "/0"}})
	err = html3.List(sl, cListZero, dummyDB, html3.Everything)
	httpErr, ok = errors.AsType[*echo.HTTPError](err)
	be.True(t, ok)
	be.Equal(t, httpErr.Code, http.StatusNotFound)

	// Categories and Platforms render successfully with real templates offline.
	db, err := sql.Open(postgres.DriverName, "")
	be.Err(t, err, nil)
	defer func() { _ = db.Close() }()

	tmpls, err := html3.Templates(t.Context(), sl, db, testutil.ProjectFS(t))
	be.Err(t, err, nil)

	e := echo.New()
	e.Renderer = mapRenderer{tmpls: tmpls}

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/html3/categories", nil)
	rec := httptest.NewRecorder()
	be.Err(t, html3.Categories(sl, e.NewContext(req, rec)), nil)
	be.Equal(t, rec.Code, http.StatusOK)
	be.True(t, strings.Contains(rec.Body.String(), "Index of /html3/categories"))

	req = httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/html3/platforms", nil)
	rec = httptest.NewRecorder()
	be.Err(t, html3.Platforms(sl, e.NewContext(req, rec)), nil)
	be.Equal(t, rec.Code, http.StatusOK)
	be.True(t, strings.Contains(rec.Body.String(), "Index of /html3/platforms"))
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

func TestTemplatesOfflineDB(t *testing.T) {
	t.Parallel()

	db, err := sql.Open(postgres.DriverName, "")
	be.Err(t, err, nil)
	defer func() { _ = db.Close() }()

	sl := logs.Discard()

	// Empty FS returns map with nil templates.
	emptyTmpls, err := html3.Templates(t.Context(), sl, db, embed.FS{})
	be.Err(t, err, nil)
	be.Equal(t, len(emptyTmpls), 11)
	for _, tmpl := range emptyTmpls {
		be.True(t, tmpl == nil)
	}

	fsys := testutil.ProjectFS(t)
	tmpls, err := html3.Templates(t.Context(), sl, db, fsys)
	be.Err(t, err, nil)
	be.True(t, tmpls["html3_index"] != nil)

	// Invoke TemplateFuncMap closures.
	fm := html3.TemplateFuncMap(t.Context(), sl, db)
	linkHref, ok := fm["linkHref"].(func(int64) string)
	be.True(t, ok)
	be.Equal(t, linkHref(1), "/html3/d/9b1c6")

	safeHTML, ok := fm["safeHTML"].(func(string) template.HTML)
	be.True(t, ok)
	be.Equal(t, safeHTML("<b>hi</b>"), template.HTML("<b>hi</b>"))

	metaByName, ok := fm["metaByName"].(func(string) tags.TagData)
	be.True(t, ok)
	be.Equal(t, metaByName("missing-tag"), tags.TagData{})
}

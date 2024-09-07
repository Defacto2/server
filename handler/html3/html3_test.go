package html3_test

import (
	"embed"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/html3"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volatiletech/null/v8"
	"go.uber.org/zap"
)

func logr() *zap.SugaredLogger {
	return zap.NewExample().Sugar()
}

func newContext() echo.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func TestSugared(t *testing.T) {
	sug := html3.Sugared{}
	err := sug.Category(newContext(), nil)
	require.Error(t, err)
	err = sug.Documents(newContext(), nil)
	require.Error(t, err)
	err = sug.Group(newContext(), nil)
	require.Error(t, err)
	err = sug.Platform(newContext(), nil)
	require.Error(t, err)
	err = sug.Platforms(newContext())
	require.Error(t, err)
	err = sug.Software(newContext(), nil)
	require.Error(t, err)
}

func TestGroups(t *testing.T) {
	sug := html3.Sugared{}
	err := sug.Groups(newContext(), nil)
	require.Error(t, err)
}

func TestIndex(t *testing.T) {
	sug := html3.Sugared{}
	err := sug.Index(newContext(), nil)
	require.Error(t, err)
}

func TestRoutes(t *testing.T) {
	t.Parallel()
	e := echo.New()
	g := html3.Routes(e, nil, nil)
	assert.NotNil(t, g)
}

func TestSugaredAll(t *testing.T) {
	sug := html3.Sugared{}
	err := sug.All(newContext(), nil)
	require.Error(t, err)
}

func TestSugaredArt(t *testing.T) {
	sug := html3.Sugared{}
	err := sug.Art(newContext(), nil)
	require.Error(t, err)
}

func TestIsCategories(t *testing.T) {
	sug := html3.Sugared{}
	err := sug.Categories(newContext())
	require.Error(t, err)
}

func TestGlobTo(t *testing.T) {
	s := html3.GlobTo("file")
	assert.Equal(t, "view/html3/file", s)
}

func TestTemplates(t *testing.T) {
	t.Parallel()
	x := html3.Templates(nil, nil, embed.FS{})
	assert.NotEmpty(t, x)
}

func TestError(t *testing.T) {
	err := html3.Error(newContext(), nil)
	require.Error(t, err)
}

func TestID(t *testing.T) {
	s := html3.ID(newContext())
	assert.Equal(t, "", s)
}

func TestLeadFS(t *testing.T) {
	x := html3.LeadFS(0, null.Int64From(0))
	assert.Equal(t, "0B", x)
	x = html3.LeadFS(10, null.Int64From(3))
	assert.Equal(t, strings.Repeat(" ", 8)+"3B", x)
}

func TestLeadInt(t *testing.T) {
	x := html3.LeadInt(0, 0)
	assert.Equal(t, "-", x)
	x = html3.LeadInt(10, 3)
	assert.Equal(t, strings.Repeat(" ", 9)+"3", x)
}

func TestQuery(t *testing.T) {
	a, b, c, fs, err := html3.Query(newContext(), nil, -1, -1)
	assert.Empty(t, a)
	assert.Empty(t, b)
	assert.Empty(t, c)
	assert.Empty(t, fs)
	assert.Error(t, err)
	a, b, c, fs, err = html3.Query(newContext(), nil, html3.Everything, -1)
	assert.Empty(t, a)
	assert.Empty(t, b)
	assert.Empty(t, c)
	assert.Empty(t, fs)
	assert.Error(t, err)
	_, _, _, _, err = html3.Query(newContext(), nil, html3.BySection, -1)
	assert.Error(t, err)
	_, _, _, _, err = html3.Query(newContext(), nil, html3.ByPlatform, -1)
	assert.Error(t, err)
	_, _, _, _, err = html3.Query(newContext(), nil, html3.ByGroup, -1)
	assert.Error(t, err)
	_, _, _, _, err = html3.Query(newContext(), nil, html3.AsArt, -1)
	assert.Error(t, err)
	_, _, _, _, err = html3.Query(newContext(), nil, html3.AsDocument, -1)
	assert.Error(t, err)
	_, _, _, _, err = html3.Query(newContext(), nil, html3.AsSoftware, -1)
	assert.Error(t, err)
}

func TestListInfo(t *testing.T) {
	a, b := html3.ListInfo(0, "", "")
	assert.Equal(t, "Index of /html3/", a)
	assert.Equal(t, "", b)
	a, b = html3.ListInfo(10, "aaa", "bbb")
	assert.Equal(t, "Index of /html3/aaa", a)
	assert.Equal(t, "", b)
}

func TestRecordsBy(t *testing.T) {
	by := html3.Everything
	assert.Equal(t, "html3_all", by.String())
	assert.Equal(t, "", by.Parent())
	by = html3.AsSoftware
	assert.Equal(t, "html3_software", by.String())
	assert.Equal(t, "", by.Parent())
	by = html3.BySection
	assert.Equal(t, "html3_category", by.String())
	assert.Equal(t, "categories", by.Parent())
}

func TestClauses(t *testing.T) {
	tests := []string{
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
		assert.Equal(t, int(html3.Clauses(s)), i)
	}
	assert.Equal(t, int(html3.Clauses("")),
		int(html3.Clauses(html3.NameAsc)), "default should be name asc")
}

func TestSorter(t *testing.T) {
	tests := []string{
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
			assert.Equal(t, d, html3.Sorter(s)[string(html3.Name)])
		case html3.NameDes:
			assert.Equal(t, a, html3.Sorter(s)[string(html3.Name)])
		case html3.PublAsc:
			assert.Equal(t, d, html3.Sorter(s)[string(html3.Publish)])
		case html3.PublDes:
			assert.Equal(t, a, html3.Sorter(s)[string(html3.Publish)])
		case html3.PostAsc:
			assert.Equal(t, d, html3.Sorter(s)[string(html3.Posted)])
		case html3.PostDes:
			assert.Equal(t, a, html3.Sorter(s)[string(html3.Posted)])
		case html3.SizeAsc:
			assert.Equal(t, d, html3.Sorter(s)[string(html3.Size)])
		case html3.SizeDes:
			assert.Equal(t, a, html3.Sorter(s)[string(html3.Size)])
		case html3.DescAsc:
			assert.Equal(t, d, html3.Sorter(s)[string(html3.Desc)])
		case html3.DescDes:
			assert.Equal(t, a, html3.Sorter(s)[string(html3.Desc)])
		}
	}
}

func TestFile_Description(t *testing.T) {
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
		name      string
		fields    fields
		expect    string
		assertion assert.ComparisonAssertionFunc
	}{
		{"empty", fields{}, "", assert.Exactly},
		{"only title", fields{Title: x}, "", assert.Exactly},
		{"req group", fields{Title: x, Platform: p}, "", assert.Exactly},
		{"default", fields{x, g, "", ""}, "Hello world from Defacto2.", assert.Exactly},
		{"with platform", fields{x, g, "", p}, "Hello world from Defacto2 for Dos.", assert.Exactly},
		{"no title", fields{"", g, "", p}, "A release from Defacto2 for Dos.", assert.Exactly},
		{"magazine", fields{"1", g, m, p}, "Defacto2 issue 1 for Dos.", assert.Exactly},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := html3.File{
				Title:    tt.fields.Title,
				GroupBy:  tt.fields.GroupBy,
				Section:  tt.fields.Section,
				Platform: tt.fields.Platform,
			}
			t.Run(tt.name, func(t *testing.T) {
				tt.assertion(t, tt.expect, f.Description())
			})
		})
	}
}

func TestDescription(t *testing.T) {
	empty := null.String{}
	s := html3.Description(empty, empty, empty, empty)
	assert.Empty(t, s)
	s = html3.Description(
		null.StringFrom("intro"),
		null.StringFrom("dos"),
		null.StringFrom("Defacto2"),
		null.StringFrom("Hello world"))
	assert.Equal(t, "Hello world from Defacto2 for Dos.", s)
}

func TestFileHref(t *testing.T) {
	s := html3.FileHref(nil, 0)
	assert.Equal(t, "zap logger is nil", s)
	s = html3.FileHref(logr(), 0)
	assert.Equal(t, "/html3/d/0", s)
}

func TestFileLinkPad(t *testing.T) {
	n := null.String{}
	s := html3.FileLinkPad(0, n)
	assert.Equal(t, "", s)
	s = html3.FileLinkPad(20, null.StringFrom("file"))
	assert.Equal(t, "                ", s)
}

//nolint:testifylint
func TestFormattings(t *testing.T) {
	assert.Equal(t, html3.File{Filename: ""}.FileLinkPad(0), "", "empty")
	assert.Equal(t, html3.File{Filename: ""}.FileLinkPad(4), "    ", "4 pads")
	assert.Equal(t, html3.File{Filename: "file"}.FileLinkPad(6), "  ", "2 pads")
	assert.Equal(t, html3.File{Filename: "file.txt"}.FileLinkPad(6), "", "too big")
	assert.Equal(t, html3.File{Size: 0}.LeadFS(0), "0B", "0 size")
	assert.Equal(t, html3.File{Size: 1}.LeadFS(3), " 1B", "1 size")
	assert.Equal(t, html3.File{Size: 1000}.LeadFS(0), "1000B", "1000 size")
	assert.Equal(t, html3.File{Size: 1024}.LeadFS(0), "1k", "1k size")
	assert.Equal(t, html3.File{Size: int64(math.Pow(1024, 2))}.LeadFS(0), "1M", "1MB size")
	assert.Equal(t, html3.File{Size: int64(math.Pow(1024, 3))}.LeadFS(0), "1G", "1GB size")
	assert.Equal(t, html3.File{Size: int64(math.Pow(1024, 4))}.LeadFS(0), "1T", "1TB size")
	assert.Equal(t, html3.LeadInt(0, 1), "1")
	assert.Equal(t, html3.LeadInt(1, 1), "1")
	assert.Equal(t, html3.LeadInt(3, 1), "  1")
	assert.True(t, html3.File{Platform: "java"}.IsOS())
	assert.Equal(t, html3.File{Platform: "java"}.OS(), " for Java")
}

func TestPagi(t *testing.T) {
	type args struct {
		page    int
		maxPage uint
	}
	tests := []struct {
		name  string
		args  args
		want  int
		want1 int
		want2 int
	}{
		{"empty", args{}, 0, 0, 0},
		{"1 page", args{1, 1}, 0, 0, 0},
		{"2 pages", args{1, 2}, 0, 0, 0},
		{"3 pages", args{1, 3}, 2, 0, 0},
		{"4 pages", args{1, 4}, 2, 3, 0},
		{"start of many pages", args{2, 10}, 2, 3, 4},
		{"middle of many pages", args{5, 10}, 4, 5, 6},
		{"near end of many pages", args{9, 10}, 7, 8, 9},
		{"last of many pages", args{10, 10}, 7, 8, 9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := html3.Pagi(tt.args.page, tt.args.maxPage)
			assert.Equal(t, tt.want, got, "value a")
			assert.Equal(t, tt.want1, got1, "value b")
			assert.Equal(t, tt.want2, got2, "value c")
		})
	}
}

func TestNavi(t *testing.T) {
	limit := 10
	page := 2
	maxPage := uint(5)
	current := "current"
	qs := "query"

	expected := html3.Navigate{
		Current:  current,
		Limit:    limit,
		Page:     page,
		PagePrev: 1,
		PageNext: 3,
		PageMax:  5,
		QueryStr: qs,
	}

	result := html3.Navi(limit, page, maxPage, current, qs)

	if result != expected {
		t.Errorf("Navi(%d, %d, %d, %s, %s) = %v; want %v", limit, page, maxPage, current, qs, result, expected)
	}
}

func TestTemplateFuncMap(t *testing.T) {
	t.Parallel()
	fm := html3.TemplateFuncMap(nil, nil)
	assert.Nil(t, fm)
}

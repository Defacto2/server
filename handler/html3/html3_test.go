package html3_test

import (
	"database/sql"
	"embed"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/html3"
	"github.com/Defacto2/server/internal/logs"
	"github.com/aarondl/null/v8"
	"github.com/labstack/echo/v4"
	"github.com/nalgeon/be"
)

func newContext() echo.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func TestRoutes(t *testing.T) {
	t.Parallel()
	e := echo.New()
	sl := logs.Discard()
	var db sql.DB
	g := html3.Routes(e, &db, sl)
	be.True(t, g != nil)
}

func TestGlobTo(t *testing.T) {
	s := html3.GlobTo("file")
	be.Equal(t, "view/html3/file", s)
}

func TestTemplates(t *testing.T) {
	t.Parallel()
	x := html3.Templates(nil, nil, embed.FS{})
	be.Equal(t, len(x), 11)
}

func TestError(t *testing.T) {
	err := html3.Error(newContext(), nil)
	be.Err(t, err)
}

func TestID(t *testing.T) {
	s := html3.ID(newContext())
	be.Equal(t, s, "")
}

func TestLeadFS(t *testing.T) {
	x := html3.LeadFS(0, null.Int64From(0))
	be.Equal(t, "0B", x)
	x = html3.LeadFS(10, null.Int64From(3))
	be.Equal(t, strings.Repeat(" ", 8)+"3B", x)
}

func TestLeadInt(t *testing.T) {
	x := html3.LeadInt(0, 0)
	be.Equal(t, "-", x)
	x = html3.LeadInt(10, 3)
	be.Equal(t, strings.Repeat(" ", 9)+"3", x)
}

func TestQuery(t *testing.T) {
	a, b, c, fs, err := html3.Query(newContext(), nil, -1, -1)
	be.True(t, a == 0)
	be.True(t, b == 0)
	be.True(t, c == 0)
	be.Equal(t, len(fs), 0)
	be.Err(t, err)
	a, b, c, fs, err = html3.Query(newContext(), nil, html3.Everything, -1)
	be.True(t, a == 0)
	be.True(t, b == 0)
	be.True(t, c == 0)
	be.Equal(t, len(fs), 0)
	be.Err(t, err)
	_, _, _, _, err = html3.Query(newContext(), nil, html3.BySection, -1)
	be.Err(t, err)
	_, _, _, _, err = html3.Query(newContext(), nil, html3.ByPlatform, -1)
	be.Err(t, err)
	_, _, _, _, err = html3.Query(newContext(), nil, html3.ByGroup, -1)
	be.Err(t, err)
	_, _, _, _, err = html3.Query(newContext(), nil, html3.AsArt, -1)
	be.Err(t, err)
	_, _, _, _, err = html3.Query(newContext(), nil, html3.AsDocument, -1)
	be.Err(t, err)
	_, _, _, _, err = html3.Query(newContext(), nil, html3.AsSoftware, -1)
	be.Err(t, err)
}

func TestListInfo(t *testing.T) {
	a, b := html3.ListInfo(0, "", "")
	be.Equal(t, "Index of /html3/", a)
	be.Equal(t, b, "")
	a, b = html3.ListInfo(10, "aaa", "bbb")
	be.Equal(t, "Index of /html3/aaa", a)
	be.Equal(t, b, "")
}

func TestRecordsBy(t *testing.T) {
	by := html3.Everything
	be.Equal(t, "html3_all", by.String())
	be.Equal(t, by.Parent(), "")
	by = html3.AsSoftware
	be.Equal(t, "html3_software", by.String())
	be.Equal(t, by.Parent(), "")
	by = html3.BySection
	be.Equal(t, "html3_category", by.String())
	be.Equal(t, "categories", by.Parent())
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
		be.Equal(t, int(html3.Clauses(s)), i)
	}
	be.Equal(t, int(html3.Clauses("")),
		int(html3.Clauses(html3.NameAsc)))
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
			be.Equal(t, d, html3.Sorter(s)[string(html3.Name)])
		case html3.NameDes:
			be.Equal(t, a, html3.Sorter(s)[string(html3.Name)])
		case html3.PublAsc:
			be.Equal(t, d, html3.Sorter(s)[string(html3.Publish)])
		case html3.PublDes:
			be.Equal(t, a, html3.Sorter(s)[string(html3.Publish)])
		case html3.PostAsc:
			be.Equal(t, d, html3.Sorter(s)[string(html3.Posted)])
		case html3.PostDes:
			be.Equal(t, a, html3.Sorter(s)[string(html3.Posted)])
		case html3.SizeAsc:
			be.Equal(t, d, html3.Sorter(s)[string(html3.Size)])
		case html3.SizeDes:
			be.Equal(t, a, html3.Sorter(s)[string(html3.Size)])
		case html3.DescAsc:
			be.Equal(t, d, html3.Sorter(s)[string(html3.Desc)])
		case html3.DescDes:
			be.Equal(t, a, html3.Sorter(s)[string(html3.Desc)])
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
		name   string
		fields fields
		expect string
	}{
		{"empty", fields{}, ""},
		{"only title", fields{Title: x}, ""},
		{"req group", fields{Title: x, Platform: p}, ""},
		{"default", fields{x, g, "", ""}, "Hello world from Defacto2."},
		{"with platform", fields{x, g, "", p}, "Hello world from Defacto2 for Dos."},
		{"no title", fields{"", g, "", p}, "A release from Defacto2 for Dos."},
		{"magazine", fields{"1", g, m, p}, "Defacto2 issue 1 for Dos."},
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
				be.Equal(t, f.Description(), tt.expect)
			})
		})
	}
}

func TestDescription(t *testing.T) {
	empty := null.String{}
	s := html3.Description(empty, empty, empty, empty)
	be.Equal(t, s, "")
	s = html3.Description(
		null.StringFrom("intro"),
		null.StringFrom("dos"),
		null.StringFrom("Defacto2"),
		null.StringFrom("Hello world"))
	be.Equal(t, "Hello world from Defacto2 for Dos.", s)
}

func TestFileHref(t *testing.T) {
	s := html3.FileHref(nil, 0)
	be.Equal(t, s, "sl slog logger pointer is nil")
	sl := logs.Discard()
	s = html3.FileHref(sl, 0)
	be.Equal(t, "/html3/d/0", s)
}

func TestFileLinkPad(t *testing.T) {
	n := null.String{}
	s := html3.FileLinkPad(0, n)
	be.Equal(t, s, "")
	s = html3.FileLinkPad(20, null.StringFrom("file"))
	be.Equal(t, "                ", s)
}

func TestFormattings(t *testing.T) {
	be.Equal(t, html3.File{Filename: ""}.FileLinkPad(0), "", "empty")
	be.Equal(t, html3.File{Filename: ""}.FileLinkPad(4), "    ", "4 pads")
	be.Equal(t, html3.File{Filename: "file"}.FileLinkPad(6), "  ", "2 pads")
	be.Equal(t, html3.File{Filename: "file.txt"}.FileLinkPad(6), "", "too big")
	be.Equal(t, html3.File{Size: 0}.LeadFS(0), "0B", "0 size")
	be.Equal(t, html3.File{Size: 1}.LeadFS(3), " 1B", "1 size")
	be.Equal(t, html3.File{Size: 1000}.LeadFS(0), "1000B", "1000 size")
	be.Equal(t, html3.File{Size: 1024}.LeadFS(0), "1k", "1k size")
	be.Equal(t, html3.File{Size: int64(1024 * 1024)}.LeadFS(0), "1M", "1MB size")
	be.Equal(t, html3.File{Size: int64(1024 * 1024 * 1024)}.LeadFS(0), "1G", "1GB size")
	be.Equal(t, html3.File{Size: int64(1024 * 1024 * 1024 * 1024)}.LeadFS(0), "1T", "1TB size")
	be.Equal(t, html3.LeadInt(0, 1), "1")
	be.Equal(t, html3.LeadInt(1, 1), "1")
	be.Equal(t, html3.LeadInt(3, 1), "  1")
	be.True(t, html3.File{Platform: "java"}.IsOS())
	be.Equal(t, html3.File{Platform: "java"}.OS(), " for Java")
}

func TestPagi(t *testing.T) {
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
			be.Equal(t, tt.want, got)
			be.Equal(t, tt.want1, got1)
			be.Equal(t, tt.want2, got2)
		})
	}
}

func TestNavi(t *testing.T) {
	limit := 10
	page := 2
	maxPage := 5
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

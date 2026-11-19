package app_test

import (
	"bytes"
	"database/sql"
	"html/template"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/areacode"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

func TestTemplates(t *testing.T) {
	t.Parallel()

	w := app.WebApp{}
	_, err := w.Templates(t.Context(), nil)
	be.Err(t, err)
}

func TestFuncMap(t *testing.T) {
	t.Parallel()

	w := app.WebApp{}
	db := &sql.DB{}
	m := w.FuncMap(t.Context(), db)
	keys := slices.Sorted(maps.Keys(m))
	be.True(t, slices.Contains(keys, "add"))
	be.True(t, slices.Contains(keys, "version"))
	be.True(t, slices.Contains(keys, "az"))
	be.True(t, slices.Contains(keys, "msdos"))
}

func TestLinkSamples(t *testing.T) {
	t.Parallel()

	x := app.LinkPreviews("1", "2", "3", "4", "5", "6", "7")
	be.True(t, len(x) == 7)
	be.True(t, strings.Contains(x[0], "youtube.com/watch?v=1"))
	be.True(t, strings.Contains(x[1], "demozoo.org/productions/2"))
}

func TestLinkRelsPerf(t *testing.T) {
	t.Parallel()

	s := app.LinkRelsPerf("", "")
	be.Equal(t, s, "")

	s = app.LinkRelsPerf("Group 1", "Group 2")
	be.True(t, strings.Contains(string(s), "Group 1"))
	be.True(t, strings.Contains(string(s), "Group 2"))
	be.True(t, strings.Contains(string(s), `href="/g/group-1"`))
	be.True(t, strings.Contains(string(s), `href="/g/group-2"`))
}

func TestTemplTemplates(t *testing.T) {
	t.Parallel()

	tpl := app.WebApp{}
	x, err := tpl.Templates(t.Context(), nil)
	be.Err(t, err)
	be.True(t, x == nil)
}

func TestFuncClosures(t *testing.T) {
	t.Parallel()

	// tpl := app.WebApp{}
	// x := tpl.FuncClosure()
	// be.True(t, x == template.FuncMap{})
}

func TestLinkRelrs(t *testing.T) {
	t.Parallel()

	x := string(app.LinkRelrs(false, nil, nil))
	be.True(t, x == "")

	x = string(app.LinkRelsPerf(nil, nil))
	be.True(t, x == "")

	x = string(app.LinkReleasers(false, false, nil, nil))
	be.True(t, x == "")
}

func TestTempls(t *testing.T) {
	t.Parallel()

	x := app.WebApp{}
	pages := x.Pages()

	p := filepath.Join("../", "../", "view", "app")
	view, err := filepath.Abs(p)
	be.Err(t, err, nil)

	for _, page := range pages {
		be.True(t, page != "")

		ext := filepath.Ext(string(page))
		be.Equal(t, ".tmpl", ext)

		stat, err := os.Stat(filepath.Join(view, string(page)))
		be.Err(t, err, nil)
		be.True(t, stat.Size() > 0)
	}
}

func TestScenerTemplate(t *testing.T) {
	t.Parallel()

	wa := app.WebApp{}
	db := &sql.DB{}
	tmpl, err := template.New("scener.tmpl").Funcs(wa.FuncMap(t.Context(), db)).
		ParseFiles(filepath.Join("../", "../", "view", "app", "scener.tmpl"))
	be.Err(t, err, nil)

	names := []string{"007", "ACE"}
	groups, letters := app.BuildScenerIndex(names)
	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "content", map[string]any{
		"sceners":       names,
		"scenerRole":    "coder",
		"scenerGroups":  groups,
		"scenerLetters": letters,
	})
	be.Err(t, err, nil)

	out := buf.String()
	be.True(t, strings.Contains(out, `href="/coder" rel="nofollow" aria-current="page"`))
	be.True(t, strings.Contains(out, `id="scener-sec-num"`))
	be.True(t, strings.Contains(out, `id="scener-sec-A"`))
	be.True(t, strings.Contains(out, `id="scener-filter-clear" class="btn btn-outline-secondary"`))
	be.True(t, strings.Contains(out, `2 sceners`))
}

func TestSearchPostTemplate(t *testing.T) {
	t.Parallel()

	wa := app.WebApp{}
	db := &sql.DB{}
	tmpl, err := template.New("searchpost.tmpl").Funcs(wa.FuncMap(t.Context(), db)).
		ParseFiles(filepath.Join("../", "../", "view", "app", "searchpost.tmpl"))
	be.Err(t, err, nil)

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "content", map[string]any{
		"title":      "Game or app titles search",
		"info":       "search artifact titles, descriptions, and comments",
		"searchMode": "desc",
		"editor":     true,
	})
	be.Err(t, err, nil)

	out := buf.String()
	be.True(t, strings.Contains(out, `href="/search/desc" rel="nofollow" aria-current="page"`))
	be.True(t, strings.Contains(out, `href="/editor/search/id"`))
	be.True(t, strings.Contains(out, `name="search-term-query"`))
	be.True(t, strings.Contains(out, `ultima 7, ultima 8`))
}

func TestSearchHtmxTemplate(t *testing.T) {
	t.Parallel()

	wa := app.WebApp{}
	db := &sql.DB{}
	tmpl, err := template.New("searchhtmx.tmpl").Funcs(wa.FuncMap(t.Context(), db)).
		ParseFiles(filepath.Join("../", "../", "view", "app", "searchhtmx.tmpl"))
	be.Err(t, err, nil)

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "content", map[string]any{
		"title":            "Lookup releasers",
		"info":             "find groups, names, magazines, bbs boards, ftp sites",
		"helpText":         "acronyms only match exact finds: 'rc' and 'rcn' are treated different",
		"hxPost":           "/search/releaser",
		"searchMode":       "releaser",
		"inputPlaceholder": "Type to search for a releaser…",
	})
	be.Err(t, err, nil)

	out := buf.String()
	be.True(t, strings.Contains(out, `href="/search/releaser" rel="nofollow" aria-current="page"`))
	be.True(t, strings.Contains(out, `id="search-htmx-clear" class="btn btn-outline-secondary" disabled`))
	be.True(t, strings.Contains(out, `hx-post="/search/releaser"`))
}

func TestAreacodesTemplate(t *testing.T) {
	t.Parallel()

	wa := app.WebApp{}
	db := &sql.DB{}
	tmpl, err := template.New("areacodes.tmpl").Funcs(wa.FuncMap(t.Context(), db)).
		ParseFiles(filepath.Join("../", "../", "view", "app", "areacodes.tmpl"))
	be.Err(t, err, nil)

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "content", map[string]any{
		"telephonecodes":  areacode.AreaCodes(),
		"territories":     areacode.Regions(),
		"abbreviations":   areacode.Abbreviations(),
		"areaCodeEntries": app.BuildAreaCodeEntries(),
		"regionEntries":   app.BuildAreaRegionEntries(),
		"abbrEntries":     app.BuildAreaAbbrEntries(),
	})
	be.Err(t, err, nil)

	out := buf.String()
	be.True(t, strings.Contains(out, `hx-trigger="input changed delay:150ms, search"`))
	be.True(t, strings.Contains(out, `id="areacodes-clear" disabled`))
	be.True(t, strings.Contains(out, `href="#sec-codes"`))
	be.True(t, strings.Contains(out, `>212</span>`))
	be.True(t, strings.Contains(out, `>NY</span>`))
	be.True(t, strings.Contains(out, `New York (NY)`))
}

func TestWebsitesTemplate(t *testing.T) {
	t.Parallel()

	wa := app.WebApp{}
	db := &sql.DB{}
	tmpl, err := template.New("websites.tmpl").Funcs(wa.FuncMap(t.Context(), db)).
		ParseFiles(
			filepath.Join("../", "../", "view", "app", "websites.tmpl"),
			filepath.Join("../", "../", "view", "app", "website.tmpl"),
		)
	be.Err(t, err, nil)

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "content", map[string]any{
		"accordion": app.List(),
	})
	be.Err(t, err, nil)

	out := buf.String()
	be.True(t, strings.Contains(out, `id="websites-expand-all"`))
	be.True(t, strings.Contains(out, `row-cols-xl-3`))
	be.True(t, strings.Contains(out, `<h3 class="h5 card-title mb-0">`))
	be.True(t, strings.Contains(out, `>2022</span>`))
}

func TestArtifactTemplatesRender(t *testing.T) {
	t.Parallel()

	wa := app.WebApp{}
	db := &sql.DB{}
	files := []string{
		"layout.tmpl", "modal.tmpl", "option_os.tmpl", "option_tag.tmpl",
		"artifact.tmpl", "pagination.tmpl", "opengraph.tmpl",
		"layoutlock.tmpl", "layoutjs.tmpl", "layoutup.tmpl", "layoutjsup.tmpl",
		"uploader.tmpl", "uploader_modal.tmpl",
		"artifactinfo.tmpl", "artifactjsdos.tmpl", "artifactfile.tmpl",
		"artifactedit.tmpl", "artifacteditjsdos.tmpl", "artifactfooter.tmpl", "artifactlock.tmpl",
	}
	paths := make([]string, len(files))
	for i, f := range files {
		paths[i] = filepath.Join("..", "..", "view", "app", f)
	}
	funcs := wa.FuncMap(t.Context(), db)
	funcs["classification"] = func(s, p string) string { return "MS-DOS" }
	funcs["classificationStr"] = func(s, p string) string { return "MS-DOS" }
	tmpl, err := template.New("").Funcs(funcs).ParseFiles(paths...)
	be.Err(t, err, nil)

	c := testutil.NewContext(t, "/f/b32729e")
	data := app.EmptyTester(c)
	data["editor"] = true
	data["readonlymode"] = false
	data["modID"] = int64(59506)
	data["unid"] = "c8cd0b4c-2f54-11e0-8827-cc1607e15609"
	data["download"] = "b32729e"
	data["modOnline"] = true
	data["filename"] = "test.zip"
	data["modTitle"] = "Test Artifact"
	data["modReleaser1"] = "Razor 1911"
	data["modReleaser2"] = ""
	data["modYear"] = int16(1995)
	data["modMonth"] = int16(1)
	data["modDay"] = int16(2)
	data["modLMYear"] = int16(1995)
	data["modLMMonth"] = int16(1)
	data["modLMDay"] = int16(2)
	data["epochYear"] = int16(1980)
	data["modOS"] = "dos"
	data["modTag"] = "releaseadvert"
	data["modRelations"] = ""
	data["modWebsites"] = ""
	data["writers"] = ""
	data["artists"] = ""
	data["programmers"] = ""
	data["musicians"] = ""
	data["youtube"] = ""
	data["sixteenColors"] = ""
	data["demozoo"] = ""
	data["pouet"] = ""
	data["github"] = ""
	data["comment"] = ""
	data["alertURL"] = ""
	data["magic"] = ""
	data["modAbsDownload"] = ""
	data["modMagicMime"] = ""
	data["modMagicNumber"] = ""
	data["modDBModify"] = ""
	data["modStatModify"] = ""
	data["modStatSizeB"] = ""
	data["modStatSizeF"] = ""
	data["modAssetPreview"] = false
	data["modAssetThumbnail"] = false
	data["modAssetExtra"] = []string{}
	data["missingAssets"] = ""
	data["modReadmeSuggest"] = ""
	data["disableReadme"] = false
	data["modZipContent"] = ""
	data["modReleasers"] = ""
	data["forApproval"] = false
	data["disableApproval"] = ""
	data["disableRecord"] = ""
	data["modEmulateBroken"] = false
	data["modEmulateRun"] = ""
	data["modEmulateCPU"] = ""
	data["modEmulateMachine"] = ""
	data["modEmulateAudio"] = ""

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "layout", data)
	be.Err(t, err, nil)
	be.True(t, strings.Contains(buf.String(), `name="artifact-editor-pouetundo"`))
}

func TestInterviewTemplate(t *testing.T) {
	t.Parallel()

	wa := app.WebApp{}
	db := &sql.DB{}
	tmpl, err := template.New("interview.tmpl").Funcs(wa.FuncMap(t.Context(), db)).
		ParseFiles(filepath.Join("..", "..", "view", "app", "interview.tmpl"))
	be.Err(t, err, nil)

	items := app.Interviewees()
	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "content", map[string]any{
		"interviews":     items,
		"interviewCount": app.InterviewCount(items),
	})
	be.Err(t, err, nil)

	out := buf.String()
	be.True(t, strings.Contains(out, `id="interview-filter-clear" class="btn btn-outline-secondary"`))
	be.True(t, strings.Contains(out, `<h2 class="h5 card-title mb-0">EMPRESS</h2>`))
	be.True(t, !strings.Contains(out, `media savy`))
}

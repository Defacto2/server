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
	"testing/fstest"
	"time"

	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/handler/areacode"
	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/config"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/null/v8"
	"github.com/nalgeon/be"
)

type testStringer string

func (s testStringer) String() string { return string(s) }

func TestTemplates(t *testing.T) {
	t.Parallel()

	w := app.WebApp{}
	got, err := w.Templates(t.Context(), nil)
	be.Err(t, err)
	be.True(t, got == nil)

	// Nil View FS returns error instead of panicking.
	db := &sql.DB{}
	fsys := testutil.ProjectFS(t)
	nilView := app.WebApp{Public: fsys}
	got, err = nilView.Templates(t.Context(), db)
	be.Err(t, err)
	be.True(t, got == nil)

	// Missing Public FS fails SRI verification.
	noPub := app.WebApp{View: fsys}
	got, err = noPub.Templates(t.Context(), db)
	be.Err(t, err)
	be.True(t, got == nil)

	// Empty View FS returns parse error instead of panicking.
	emptyView := app.WebApp{Public: fsys, View: fstest.MapFS{}}
	got, err = emptyView.Templates(t.Context(), db)
	be.Err(t, err)
	be.True(t, got == nil)

	for _, readOnly := range []bool{false, true} {
		wa := app.WebApp{
			Public:      fsys,
			View:        fsys,
			Environment: config.Config{ReadOnly: config.Toggle(readOnly)},
		}
		tmpls, err := wa.Templates(t.Context(), db)
		be.Err(t, err, nil)
		be.Equal(t, len(tmpls), len(wa.Pages()))
		for key, tmpl := range tmpls {
			be.True(t, key != "")
			be.True(t, tmpl != nil)
		}
	}
}

func TestFuncMap(t *testing.T) {
	t.Parallel()

	w := app.WebApp{}
	be.True(t, w.FuncMap(t.Context(), nil) == nil)
	be.Equal(t, len(w.FuncDB(t.Context(), nil)), 0)

	db := &sql.DB{}
	m := w.FuncMap(t.Context(), db)
	keys := slices.Sorted(maps.Keys(m))
	be.True(t, slices.Contains(keys, "add"))
	be.True(t, slices.Contains(keys, "version"))
	be.True(t, slices.Contains(keys, "az"))
	be.True(t, slices.Contains(keys, "msdos"))

	// Exercise static and element template funcs.
	subFn, ok := m["sub"].(func(int, int) int)
	be.True(t, ok)
	be.Equal(t, subFn(1984, 1987), 3)

	fmtName, ok := m["fmtName"].(func(string) string)
	be.True(t, ok)
	be.Equal(t, fmtName("HELLO"), "Hello")

	fmtRangeURI, ok := m["fmtRangeURI"].(func(string) string)
	be.True(t, ok)
	be.True(t, fmtRangeURI("razor-1911") != "")

	initFn, ok := m["initialisms"].(func(string) string)
	be.True(t, ok)
	be.True(t, initFn("razor-1911") != "")

	for _, name := range []string{"az", "year", "msdos", "mergeIcon"} {
		fn, ok := m[name].(func() template.HTML)
		be.True(t, ok)
		be.True(t, fn() != "")
	}
}

func TestLinkSamples(t *testing.T) {
	t.Parallel()

	be.True(t, app.LinkPreviews("", "", "", "", "", "", "") == nil)

	x := app.LinkPreviews("1", "2", "3", "4", "5", "6", "7")
	be.Equal(t, len(x), 7)
	be.True(t, strings.Contains(x[0], "youtube.com/watch?v=1"))
	be.True(t, strings.Contains(x[1], "demozoo.org/productions/2"))

	// Zero IDs for demozoo and pouet are skipped.
	skipped := app.LinkPreviews("abc", "0", "0", "", "", "", "")
	be.Equal(t, len(skipped), 1)
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

func TestFuncClosures(t *testing.T) {
	t.Parallel()

	prevDir := t.TempDir()
	thumbDir := t.TempDir()
	wa := app.WebApp{
		Version: "1.2.3",
		Brand:   []byte("DF2"),
		Subresource: app.SRI{
			Bootstrap5:      "sha-bs5",
			Bootstrap5JS:    "sha-bs5js",
			BootstrapIcons:  "sha-icons",
			CanvasAnsi:      "sha-ansi",
			CanvasReadme:    "sha-readme",
			ChiptunePlayer:  "sha-chip",
			EditArtifact:    "sha-edart",
			EditAssets:      "sha-edass",
			EditForApproval: "sha-edapp",
			Htmx:            "sha-htmx",
			HtmxMultipart:   "sha-htmxm",
			IndexJS:         "sha-idx",
			Jsdos6JS:        "sha-jsdos",
			DosboxJS:        "sha-dosbox",
			Layout:          "sha-lay",
			LayoutJS:        "sha-layjs",
			Pouet:           "sha-pouet",
			Uploader:        "sha-up",
		},
		Environment: config.Config{
			AbsPreview:   config.Absprev(prevDir),
			AbsThumbnail: config.Absthumb(thumbDir),
		},
	}

	m := wa.FuncClosure()
	logo, ok := m["logo"].(func() string)
	be.True(t, ok)
	be.Equal(t, logo(), "DF2")
	version, ok := m["version"].(func() string)
	be.True(t, ok)
	be.Equal(t, version(), "1.2.3")
	for _, key := range []string{
		"sri_bootstrap5", "sri_bootstrap5JS", "sri_bootstrapIcons",
		"sri_canvasAnsi", "sri_canvasReadme", "sri_chiptunePlayer",
		"sri_editArtifact", "sri_editAssets", "sri_editForApproval",
		"sri_htmx", "sri_htmxMultipart", "sri_indexJS",
		"sri_jsdos6JS", "sri_dosboxJS", "sri_layout",
		"sri_layoutJS", "sri_pouet", "sri_uploader",
	} {
		fn, ok := m[key].(func() string)
		be.True(t, ok)
		be.True(t, strings.HasPrefix(fn(), "sha-"))
	}

	const unid = "00000000-0000-0000-0000-000000000001"
	recordImgSampleStat, ok := m["recordImgSampleStat"].(func(string) bool)
	be.True(t, ok)
	be.Equal(t, recordImgSampleStat(unid), false)
	recordImgSample, ok := m["recordImgSample"].(func(string) template.HTML)
	be.True(t, ok)
	be.True(t, strings.Contains(string(recordImgSample(unid)), "No preview"))
	recordThumbSample, ok := m["recordThumbSample"].(func(string) template.HTML)
	be.True(t, ok)
	be.True(t, strings.Contains(string(recordThumbSample(unid)), "No thumbnail"))
	screenshot, ok := m["screenshot"].(func(string, string) template.HTML)
	be.True(t, ok)
	be.True(t, screenshot(unid, "desc") == "")
	animPreview, ok := m["animPreview"].(func(string) string)
	be.True(t, ok)
	be.Equal(t, animPreview(unid), "")
	thumb, ok := m["thumb"].(func(string, string, bool) template.HTML)
	be.True(t, ok)
	be.True(t, strings.Contains(string(thumb(unid, "desc", false)), "no thumbnail"))
	recordPreviewSrc, ok := m["recordPreviewSrc"].(func(string, string) string)
	be.True(t, ok)
	be.True(t, strings.Contains(recordPreviewSrc(unid, ".png"), "no such file"))
	recordThumbnailSrc, ok := m["recordThumbnailSrc"].(func(string, string) string)
	be.True(t, ok)
	be.True(t, strings.Contains(recordThumbnailSrc(unid, ".png"), "no such file"))

	ogFn, ok := m["og_image"].(func(any) string)
	be.True(t, ok)
	be.True(t, strings.Contains(ogFn(nil), "defacto2-ascii.png"))
	be.True(t, strings.Contains(ogFn(123), "defacto2-ascii.png"))
	be.True(t, strings.Contains(ogFn(""), "defacto2-ascii.png"))
	be.True(t, ogFn(unid) != "")
}

func TestLinkRelrs(t *testing.T) {
	t.Parallel()

	be.Equal(t, string(app.LinkRelrs(false, nil, nil)), "")
	be.Equal(t, string(app.LinkRels(nil, nil)), "")
	be.Equal(t, string(app.LinkRelsPerf(nil, nil)), "")
	be.Equal(t, string(app.LinkReleasers(false, false, nil, nil)), "")
	be.Equal(t, string(app.NameRels(nil, nil)), "")
	be.Equal(t, string(app.NameRels("", "")), "")

	// LinkRels/LinkRelrs/LinkRelsPerf with valid groups (including single non-nil arg)
	rels := string(app.LinkRels("Group 1", "Group 2"))
	be.True(t, strings.Contains(rels, `href="/g/group-1"`))
	be.True(t, strings.Contains(rels, `href="/g/group-2"`))
	be.True(t, strings.Contains(string(app.LinkRels("Group 1", nil)), `href="/g/group-1"`))
	be.True(t, strings.Contains(string(app.LinkRelrs(false, nil, "Group 2")), `href="/g/group-2"`))
	be.True(t, strings.Contains(string(app.LinkRelsPerf("Group 1", nil)), `href="/g/group-1"`))

	// Empty first argument swaps with second, and exercises null.String, *string, fmt.Stringer.
	swapped := string(app.LinkReleasers(true, false, "", "Solo Group"))
	be.True(t, strings.Contains(swapped, `href="/g/solo-group"`))

	strVal := "Ptr Group"
	var nilPtr *string
	byPtr := string(app.LinkReleasers(true, false, &strVal, nilPtr))
	be.True(t, strings.Contains(byPtr, `href="/g/ptr-group"`))

	byNull := string(app.LinkReleasers(true, false, null.StringFrom("Null Group"), null.String{}))
	be.True(t, strings.Contains(byNull, `href="/g/null-group"`))

	byStringer := string(app.LinkReleasers(true, false, testStringer("Str Group"), 42))
	be.True(t, strings.Contains(byStringer, `href="/g/str-group"`))

	// NameRels with second-only arg
	named := string(app.NameRels("", "Razor 1911"))
	be.True(t, strings.Contains(named, "Razor 1911"))
}

func TestReleaseDecades(t *testing.T) {
	t.Parallel()

	wa := app.WebApp{}
	tmpl, err := template.New("decades").Funcs(wa.FuncMap(t.Context(), &sql.DB{})).
		Parse(`{{- range releaseDecades .}}{{.Label}}={{.Year}};{{end -}}`)
	be.Err(t, err, nil)

	var buf bytes.Buffer
	be.Err(t, tmpl.Execute(&buf, model.Releasers(nil)), nil)
	be.Equal(t, buf.String(), "")

	buf.Reset()
	input := model.Releasers{
		nil,
		{Unique: model.Releaser{Year: null.IntFrom(0)}},
		{Unique: model.Releaser{Year: null.IntFrom(1985)}},
		{Unique: model.Releaser{Year: null.IntFrom(1992)}},
	}
	be.Err(t, tmpl.Execute(&buf, input), nil)
	be.Equal(t, buf.String(), "1980s=1981;1990s=1990;")
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
		be.Equal(t, ext, ".tmpl")

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
	be.True(t, !strings.Contains(out, `id="scener-filter-clear"`))
	be.True(t, strings.Contains(out, `2 sceners`))

	names = []string{"007", "ACE", "ZERO"}
	groups, letters = app.BuildScenerIndex(names)
	buf.Reset()
	err = tmpl.ExecuteTemplate(&buf, "content", map[string]any{
		"sceners":       names,
		"scenerRole":    "coder",
		"scenerGroups":  groups,
		"scenerLetters": letters,
	})
	be.Err(t, err, nil)
	be.True(t, strings.Contains(buf.String(), `id="scener-filter-clear" class="btn btn-outline-secondary"`))
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
		"title":      "Search titles & comments",
		"info":       "Find games, apps, descriptions, and artifact comments.",
		"searchMode": "desc",
		"editor":     true,
	})
	be.Err(t, err, nil)

	out := buf.String()
	be.True(t, strings.Contains(out, `href="/search/desc" rel="nofollow" aria-current="page"`))
	be.True(t, strings.Contains(out, `href="/editor/search/id"`))
	be.True(t, strings.Contains(out, `method="get" action="/search/desc"`))
	be.True(t, strings.Contains(out, `name="query"`))
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
		"title":            "Search groups and names",
		"info":             "Find groups, magazines, BBS boards, and FTP sites.",
		"helpText":         "Acronyms match exact initialisms: 'rc' and 'rcn' return different results.",
		"hxPost":           "/search/releaser",
		"searchMode":       "releaser",
		"inputPlaceholder": "Type a group, magazine, BBS, or FTP name...",
	})
	be.Err(t, err, nil)

	out := buf.String()
	be.True(t, strings.Contains(out, `href="/search/releaser" rel="nofollow" aria-current="page"`))
	be.True(t, strings.Contains(out, `id="search-htmx-clear" class="btn btn-outline-secondary" disabled`))
	be.True(t, strings.Contains(out, `hx-post="/search/releaser"`))
	be.True(t, !strings.Contains(out, `class="input-group-text"`))
}

func TestDependenciesTemplate(t *testing.T) {
	t.Parallel()

	wa := app.WebApp{}
	db := &sql.DB{}
	tmpl, err := template.New("dependencies.tmpl").Funcs(wa.FuncMap(t.Context(), db)).
		ParseFiles(filepath.Join("../", "../", "view", "app", "dependencies.tmpl"))
	be.Err(t, err, nil)

	arcTools := []command.ToolInfo{
		{
			Prog:     command.HWZip,
			Expected: "HWZip ver 2+",
			Abs:      "/usr/local/sbin/hwzip",
			Found:    true,
			Regular:  true,
			Perm:     0o755,
			Size:     102400,
			Output:   "HWZIP 2.4",
		},
		{
			Prog:     command.Unrar,
			Expected: "UNRAR freeware (c) Alexander Roshal",
			Abs:      "/usr/bin/unrar",
			Found:    true,
			Warning:  true,
			Regular:  true,
			Symlink:  true,
			Perm:     0o755,
			Size:     204800,
			Output:   "unrar-free 0.1.1",
			Err:      "not Alexander Roshal freeware (possible unrar-free)",
		},
		{
			Prog:     command.Cab,
			Expected: "gcab ver 1.6+",
			Found:    false,
			Err:      `exec: "gcab": executable file not found in $PATH`,
		},
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "content", map[string]any{
		"depOK":    1,
		"depTotal": 3,
		"depSections": []map[string]any{
			{
				"Title":   "Archive & Compression Tools",
				"Caption": "File and directory archive, compression, and package tools.",
				"Tools":   arcTools,
				"OK":      1,
				"Total":   3,
			},
		},
	})
	be.Err(t, err, nil)

	out := buf.String()
	be.True(t, strings.Contains(out, "1 / 3 available"))
	be.True(t, strings.Contains(out, `<span class="badge text-bg-success">OK</span>`))
	be.True(t, strings.Contains(out, `<span class="badge text-bg-warning">Warning</span>`))
	be.True(t, strings.Contains(out, `<span class="badge text-bg-danger">Missing</span>`))
	be.True(t, strings.Contains(out, `<span class="badge text-bg-secondary">symlink</span>`))
	be.True(t, strings.Contains(out, "102 kB"))
}

func TestRoutesTemplate(t *testing.T) {
	t.Parallel()

	wa := app.WebApp{}
	db := &sql.DB{}
	tmpl, err := template.New("routes.tmpl").Funcs(wa.FuncMap(t.Context(), db)).
		ParseFiles(filepath.Join("../", "../", "view", "app", "routes.tmpl"))
	be.Err(t, err, nil)

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "content", map[string]any{
		"routesCount": 3,
		"routeMethods": []map[string]any{
			{"Method": "GET", "Badge": "bg-success-subtle text-success-emphasis", "Count": 2},
			{"Method": "DELETE", "Badge": "bg-danger-subtle text-danger-emphasis", "Count": 1},
		},
		"routesList": []map[string]any{
			{"Method": "GET", "Path": "/editor/routes", "Badge": "bg-success-subtle text-success-emphasis", "Linkable": true},
			{"Method": "GET", "Path": "/f/:id", "Badge": "bg-success-subtle text-success-emphasis", "Linkable": false},
			{"Method": "DELETE", "Path": "/editor/delete/forever/:key", "Badge": "bg-danger-subtle text-danger-emphasis", "Linkable": false},
		},
	})
	be.Err(t, err, nil)

	out := buf.String()
	be.True(t, strings.Contains(out, "3 routes"))
	be.True(t, strings.Contains(out, `<a href="/editor/routes"`))
	be.True(t, strings.Contains(out, `<code class="small text-break">/f/:id</code>`))
	be.True(t, strings.Contains(out, `<code class="small text-break">/editor/delete/forever/:key</code>`))
	be.True(t, !strings.Contains(out, "https://defacto2.net"))
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
	funcs["classification"] = func(_, _ string) string { return "MS-DOS" }
	funcs["classificationStr"] = func(_, _ string) string { return "MS-DOS" }
	tmpl, err := template.New("").Funcs(funcs).ParseFiles(paths...)
	be.Err(t, err, nil)

	c := testutil.NewContext(t, "/f/b32729e")
	data := app.EmptyTester(c)
	data["canonical"] = "f/b32729e"
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

	data["h1"] = "Beyond Akira BBS"
	data["lead"] = "BBS artifacts"

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "layout", data)
	be.Err(t, err, nil)
	out := buf.String()
	be.True(t, strings.Contains(out, `<meta property="og:url" content="https://defacto2.net/f/b32729e">`))
	be.True(t, strings.Contains(out, `<meta property="og:image" content="https://defacto2.net/`))
	be.True(t, strings.Contains(out, `<meta name="twitter:card" content="summary_large_image">`))
	be.True(t, strings.Contains(out, `<meta name="twitter:site" content="@Defacto2">`))
	be.True(t, strings.Contains(out, `<meta name="twitter:image" content="https://defacto2.net/`))
	be.True(t, strings.Contains(out, `name="artifact-editor-pouetundo"`))
	be.True(t, strings.Contains(out, `<h1 class="h2 fw-semibold mb-1 container-xxl">Beyond Akira BBS</h1>`))
	be.True(t, strings.Contains(out, `<p class="text-body-secondary mb-2 container-xxl">`))
	be.True(t, strings.Contains(out, `<div class="container-xxl">`))

	data["contentBinary"] = "<div><span>XB</span></div>"
	data["customFont"] = true
	buf.Reset()
	err = tmpl.ExecuteTemplate(&buf, "layout", data)
	be.Err(t, err, nil)
	out = buf.String()
	be.True(t, strings.Contains(out, `XBIN or ICEDraw text <small class="text-body-tertiary ms-1">This file includes custom fonts that cannot be shown in browser.</small>`))
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

func TestReleaserTemplate(t *testing.T) {
	t.Parallel()

	wa := app.WebApp{}
	db := &sql.DB{}
	tmpl, err := template.New("releaser.tmpl").Funcs(wa.FuncMap(t.Context(), db)).
		ParseFiles(filepath.Join("..", "..", "view", "app", "releaser.tmpl"))
	be.Err(t, err, nil)

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "content", map[string]any{
		"h1":       "Releasers",
		"itemName": "file",
		"editor":   true,
		"stats":    map[string]string{"pubs": "1 releaser", "orderBy": "by file artifact count"},
		"releasers": model.Releasers{
			{Unique: model.Releaser{URI: "razor-1911", Name: "Razor 1911", Count: 5, Bytes: 1024, Year: null.IntFrom(1992)}},
		},
	})
	be.Err(t, err, nil)

	out := buf.String()
	be.True(t, strings.Contains(out, `row-cols-lg-3`))
	be.True(t, strings.Contains(out, `position-relative p-3 border`))
	be.True(t, strings.Contains(out, `<h2 class="h5 mb-0">`))
	be.True(t, !strings.Contains(out, `id="releaser-filter"`))
	be.True(t, strings.Contains(out, `releaser-item`))
	be.True(t, strings.Contains(out, `data-count="5"`))
	be.True(t, !strings.Contains(out, `badge rounded-pill text-bg-warning`))

	buf.Reset()
	err = tmpl.ExecuteTemplate(&buf, "content", map[string]any{
		"h1":       "Releasers",
		"itemName": "file",
		"editor":   false,
		"stats":    map[string]string{"pubs": "3 releasers", "orderBy": "by file artifact count"},
		"releasers": model.Releasers{
			{Unique: model.Releaser{URI: "a", Name: "A", Count: 5, Bytes: 1024}},
			{Unique: model.Releaser{URI: "b", Name: "B", Count: 3, Bytes: 1024}},
			{Unique: model.Releaser{URI: "c", Name: "C", Count: 2, Bytes: 1024}},
		},
	})
	be.Err(t, err, nil)
	threeRelOut := buf.String()
	be.True(t, strings.Contains(threeRelOut, `id="releaser-filter"`))
	be.True(t, strings.Contains(threeRelOut, `id="releaser-filter-clear"`))
	be.True(t, strings.Contains(threeRelOut, `id="releaser-count"`))
	be.True(t, strings.Contains(threeRelOut, `id="releaser-empty"`))

	buf.Reset()
	err = tmpl.ExecuteTemplate(&buf, "content", map[string]any{
		"h1":       "BBS",
		"itemName": "bbs",
		"editor":   false,
		"stats":    map[string]string{"pubs": "1 board", "orderBy": "by file artifact count"},
		"releasers": model.Releasers{
			{Unique: model.Releaser{URI: "cbbs", Name: "CBBS", Count: 1, Bytes: 512}},
		},
	})
	be.Err(t, err, nil)
	bbsOut := buf.String()
	be.True(t, strings.Contains(bbsOut, `Historic BBS platforms (1984&ndash;1994)`))
	be.True(t, strings.Contains(bbsOut, `href="/g/nochange-bbs"`))
	be.True(t, strings.Contains(bbsOut, `href="/g/evolution-software-bbs"`))
}

func TestReleaserYearTemplate(t *testing.T) {
	t.Parallel()

	wa := app.WebApp{}
	db := &sql.DB{}
	tmpl, err := template.New("releaseryear.tmpl").Funcs(wa.FuncMap(t.Context(), db)).
		ParseFiles(
			filepath.Join("..", "..", "view", "app", "releaseryear.tmpl"),
			filepath.Join("..", "..", "view", "app", "releasertimeline.tmpl"),
		)
	be.Err(t, err, nil)

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "content", map[string]any{
		"h1":       "Releasers",
		"itemName": "file",
		"editor":   true,
		"stats":    map[string]string{"pubs": "1 releaser", "orderBy": "by year"},
		"releasers": model.Releasers{
			{Unique: model.Releaser{URI: "razor-1911", Name: "Razor 1911", Count: 5, Bytes: 1024, Year: null.IntFrom(1992)}},
		},
	})
	be.Err(t, err, nil)

	out := buf.String()
	be.True(t, strings.Contains(out, `<h2 id="year-1992" class="h3 font-monospace fw-semibold mb-0" style="scroll-margin-top: 5rem;">1992</h2>`))
	be.True(t, strings.Contains(out, `aria-label="Jump to decade"`))
	be.True(t, strings.Contains(out, `href="#year-1981">1980s</a>`))
	be.True(t, strings.Contains(out, `href="#year-1990">1990s</a>`))
	be.True(t, strings.Contains(out, `position-relative p-3 border`))
	be.True(t, strings.Contains(out, `<h3 class="h5 mb-0">`))
	be.True(t, !strings.Contains(out, `badge rounded-pill text-bg-warning`))

	buf.Reset()
	err = tmpl.ExecuteTemplate(&buf, "content", map[string]any{
		"h1":       "Releasers",
		"itemName": "file",
		"editor":   false,
		"stats":    map[string]string{"pubs": "1 releaser", "orderBy": "by year"},
		"releasers": model.Releasers{
			{Unique: model.Releaser{URI: "cbbs", Name: "CBBS", Count: 1, Bytes: 512, Year: null.IntFrom(1980)}},
		},
	})
	be.Err(t, err, nil)
	be.True(t, strings.Contains(buf.String(), `<h2 id="year-1980" class="h3 font-monospace fw-semibold mb-0" style="scroll-margin-top: 5rem;">1980</h2>`))
	be.True(t, !strings.Contains(buf.String(), `aria-label="Jump to decade"`))

	buf.Reset()
	err = tmpl.ExecuteTemplate(&buf, "content", map[string]any{
		"h1":       "Releasers",
		"itemName": "file",
		"editor":   false,
		"stats":    map[string]string{"pubs": "3 releasers", "orderBy": "by year"},
		"releasers": model.Releasers{
			{Unique: model.Releaser{URI: "grp-1998", Name: "Grp 1998", Count: 1, Bytes: 512, Year: null.IntFrom(1998)}},
			{Unique: model.Releaser{URI: "grp-2000", Name: "Grp 2000", Count: 1, Bytes: 512, Year: null.IntFrom(2000)}},
			{Unique: model.Releaser{URI: "grp-2004", Name: "Grp 2004", Count: 1, Bytes: 512, Year: null.IntFrom(2004)}},
			{Unique: model.Releaser{URI: "grp-2006", Name: "Grp 2006", Count: 1, Bytes: 512, Year: null.IntFrom(2006)}},
		},
	})
	be.Err(t, err, nil)
	gapOut := buf.String()
	be.True(t, strings.Contains(gapOut, `<h2 id="year-2000" class="h3 font-monospace fw-semibold mb-0" style="scroll-margin-top: 5rem;">2000</h2>`))
	be.True(t, strings.Contains(gapOut, `href="#year-2000">2000s</a>`))
	be.True(t, !strings.Contains(gapOut, `>1999</h2>`))
	be.True(t, !strings.Contains(gapOut, `>2005</h2>`))
}

func TestYearRange(t *testing.T) {
	t.Parallel()

	wa := app.WebApp{}
	fn, ok := wa.FuncMap(t.Context(), &sql.DB{})["yearRange"].(func(int, int) []int)
	be.True(t, ok)

	now := time.Now().Year()
	tests := []struct {
		name  string
		start int
		end   int
		want  []int
	}{
		{"gap", 1984, 1987, []int{1985, 1986}},
		{"initial zero from 1984", 0, 1984, []int{1981, 1982, 1983}},
		{"initial zero at 1981", 0, 1981, []int{}},
		{"epoch year 1980 does not panic", 0, 1980, []int{}},
		{"pre-epoch 1978 does not panic", 0, 1978, []int{}},
		{"equal start and end", 1990, 1990, []int{}},
		{"reversed range", 1992, 1990, []int{}},
		{"future end clamped to now", now, now + 5, []int{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := fn(tc.start, tc.end)
			be.True(t, slices.Equal(got, tc.want))
		})
	}
}

func TestThanksTemplate(t *testing.T) {
	t.Parallel()

	wa := app.WebApp{}
	db := &sql.DB{}
	tmpl, err := template.New("thanks.tmpl").Funcs(wa.FuncMap(t.Context(), db)).
		ParseFiles(filepath.Join("..", "..", "view", "app", "thanks.tmpl"))
	be.Err(t, err, nil)

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "content", map[string]any{})
	be.Err(t, err, nil)

	out := buf.String()
	be.True(t, strings.Contains(out, `<div class="container-xxl">`))
	be.True(t, !strings.Contains(out, `<h1`))
	be.True(t, strings.Contains(out, `<h2 class="h3 mb-2">Support Defacto2</h2>`))
	be.True(t, strings.Contains(out, `https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&amp;hosted_button_id=PBV5H45GG5L8G`))
	be.True(t, strings.Contains(out, `https://ko-fi.com/codebyben`))
	be.True(t, strings.Contains(out, `https://www.buymeacoffee.com/4rtEGvUIY`))
	be.True(t, strings.Contains(out, `https://m.do.co/c/a9270bdb9e74`))
	be.True(t, strings.Contains(out, `aria-hidden="true"`))
	be.True(t, strings.Contains(out, `PayPal<sup class="text-danger">*</sup>`))
}

func TestAPIInfoTemplate(t *testing.T) {
	t.Parallel()

	wa := app.WebApp{}
	db := &sql.DB{}
	tmpl, err := template.New("apiinfo.tmpl").Funcs(wa.FuncMap(t.Context(), db)).
		ParseFiles(filepath.Join("..", "..", "view", "app", "apiinfo.tmpl"))
	be.Err(t, err, nil)

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "content", map[string]any{})
	be.Err(t, err, nil)

	out := buf.String()
	be.True(t, strings.Contains(out, `<div class="container-xxl">`))
	be.True(t, strings.Contains(out, `aria-label="API sections"`))
	be.True(t, strings.Contains(out, `id="sec-endpoints"`))
	be.True(t, strings.Contains(out, `Example endpoint links`))
	be.True(t, !strings.Contains(out, `bg-primary text-white`))
	be.True(t, strings.Contains(out, `text-bg-success`))
	be.True(t, !strings.Contains(out, `class="col-lg-10"`))
	be.True(t, strings.Contains(out, `GET /api/v1/areacode/212`))
	be.True(t, !strings.Contains(out, `GET /api/v1/areacodes/212`))
	be.True(t, strings.Contains(out, `areacodes/search/602`))
	be.True(t, !strings.Contains(out, `areacode/search/602`))
	be.True(t, !strings.Contains(out, `maybe required`))
	be.True(t, !strings.Contains(out, `Unique Universally Unique Identifier`))
	be.True(t, strings.Contains(out, `navigator.clipboard`))
}

func TestCategoriesTemplate(t *testing.T) {
	t.Parallel()

	wa := app.WebApp{}
	db := &sql.DB{}
	tmpl, err := template.New("categories.tmpl").Funcs(wa.FuncMap(t.Context(), db)).
		ParseFiles(
			filepath.Join("..", "..", "view", "app", "categories.tmpl"),
			filepath.Join("..", "..", "view", "app", "categoriesmore.tmpl"),
		)
	be.Err(t, err, nil)

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "content", map[string]any{
		"stats": true,
		"statSummary": map[string]any{
			"CountFmt":  "42",
			"BytesFmt":  "2.0 kB",
			"MinYear":   1981,
			"MaxYear":   2024,
			"SpanYears": 44,
		},
		"statRows": []map[string]any{
			{
				"Name":     "ANSI art",
				"URI":      "ansi",
				"Count":    10,
				"CountFmt": "10",
				"CountPct": 100,
				"Bytes":    1024,
				"BytesFmt": "1.0 kB",
				"BytesPct": 100,
				"MinYear":  1984,
				"MaxYear":  1999,
			},
		},
	})
	be.Err(t, err, nil)

	out := buf.String()
	be.True(t, strings.Contains(out, `<div class="container-xxl">`))
	be.True(t, !strings.Contains(out, `text-left`))
	be.True(t, strings.Contains(out, `id="category-stats-table"`))
	be.True(t, strings.Contains(out, `44 years of releases`))
	be.True(t, strings.Contains(out, `href="/files/ansi"`))
	be.True(t, strings.Contains(out, `href="/file"`))
	be.True(t, !strings.Contains(out, `id="category-filter"`))
	be.True(t, !strings.Contains(out, `/files/text-amiga`))

	buf.Reset()
	err = tmpl.ExecuteTemplate(&buf, "content", map[string]any{"stats": false})
	be.Err(t, err, nil)
	catOut := buf.String()
	be.True(t, !strings.Contains(catOut, `id="category-stats-table"`))
	be.True(t, strings.Contains(catOut, `id="category-filter"`))
	be.True(t, strings.Contains(catOut, `data-category-card`))
	be.True(t, strings.Contains(catOut, `aria-controls="windows-tab-pane" aria-selected="false"`))
	be.True(t, strings.Contains(catOut, `/files/text-amiga`))
	be.True(t, !strings.Contains(catOut, `class="badge bg-`))
	be.True(t, !strings.Contains(catOut, `lead card-link`))
	be.True(t, !strings.Contains(catOut, `card-subtitle lead`))
}

func TestArtifactsTemplate(t *testing.T) {
	t.Parallel()

	wa := app.WebApp{}
	db := &sql.DB{}
	tmpl, err := template.New("artifacts.tmpl").Funcs(wa.FuncMap(t.Context(), db)).
		ParseFiles(
			filepath.Join("..", "..", "view", "app", "artifacts.tmpl"),
			filepath.Join("..", "..", "view", "app", "artifactsedit.tmpl"),
			filepath.Join("..", "..", "view", "app", "pagination.tmpl"),
		)
	be.Err(t, err, nil)

	c := testutil.NewContext(t, "/files/bbs")
	data := app.EmptyTester(c)
	data["title"] = "Bulletin Board Systems artifacts"
	data["editor"] = true
	data["demozoo"] = "0"
	data["sixteen"] = ""
	data["csdb"] = 0
	data["janeway"] = 0
	data["bbs"] = false
	data["scener"] = ""
	data["tidbits"] = template.HTML("")
	data["website"] = []any{}
	data["unknownYears"] = false
	data["stats"] = map[string]string{"files": "1 file", "years": "1990 - 1990"}
	data["records"] = models.FileSlice{
		{
			ID:             1,
			Filename:       null.StringFrom("VERYLONG_BBS_ADVERT_FILE_1990.ZIP"),
			GroupBrandFor:  null.StringFrom("The BBS"),
			RecordTitle:    null.StringFrom("BBS Ad"),
			Section:        null.StringFrom("bbs"),
			Platform:       null.StringFrom("text"),
			DateIssuedYear: null.Int16From(1990),
			FileMagicType:  null.StringFrom("PKLITE compressed executable of a TheDraw COM screen"),
			WebIDDemozoo:   null.Int64From(42),
			UUID:           null.StringFrom("00000000-0000-0000-0000-000000000000"),
		},
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "content", data)
	be.Err(t, err, nil)

	out := buf.String()
	be.True(t, strings.Contains(out, `minmax(min(100%, 320px), 400px)`))
	be.True(t, strings.Contains(out, `<code class="h6 mb-0 text-break">`))
	be.True(t, strings.Contains(out, `<h2 class="card-title text-truncate fs-6">`))
	be.True(t, !strings.Contains(out, `><!-- no thumbnail found --></a>`))
	be.True(t, strings.Contains(out, `<!-- no thumbnail found -->`))
	be.True(t, strings.Contains(out, `text-info-emphasis`))
	be.True(t, strings.Contains(out, `<div class="card-text"><small class="text-body-secondary">PKLITE compressed executable of a TheDraw COM screen</small></div>`))
	be.True(t, !strings.Contains(out, `text-body-info`))
	be.True(t, strings.Contains(out, `href="/f/`))
	be.True(t, strings.Contains(out, `<button type="button" id="fluid-button" class="btn btn-sm btn-outline-secondary">Unlock columns</button>`))
	be.True(t, !strings.Contains(out, `id="artifact-filter"`))
	be.True(t, strings.Contains(out, `artifact-item`))
	be.True(t, strings.Contains(out, `<div class="d-flex flex-wrap align-items-baseline gap-3">`))
	be.True(t, !strings.Contains(out, `&nbsp; <a class="card-link"`))
	be.True(t, !strings.Contains(out, `id="batch-open-menu"`))
	be.True(t, strings.Contains(out, `data-artifact-url="/f/`))

	recs, ok := data["records"].(models.FileSlice)
	be.True(t, ok)
	for i := 2; i <= 5; i++ {
		rec := recs[0]
		rec.ID = int64(i)
		recs = append(recs, rec)
	}
	data["records"] = recs
	buf.Reset()
	err = tmpl.ExecuteTemplate(&buf, "content", data)
	be.Err(t, err, nil)
	fiveOut := buf.String()
	be.True(t, strings.Contains(fiveOut, `id="batch-open-menu"`))
	be.True(t, strings.Contains(fiveOut, `id="artifact-filter"`))
	be.True(t, strings.Contains(fiveOut, `id="artifact-filter-clear"`))
	be.True(t, strings.Contains(fiveOut, `id="artifact-count"`))
	be.True(t, strings.Contains(fiveOut, `id="artifact-empty"`))

	data["editor"] = false
	buf.Reset()
	err = tmpl.ExecuteTemplate(&buf, "content", data)
	be.Err(t, err, nil)
	noEd := buf.String()
	be.True(t, strings.Contains(noEd, `named-group-page-1`))

	data["noGroupLinks"] = true
	buf.Reset()
	err = tmpl.ExecuteTemplate(&buf, "content", data)
	be.Err(t, err, nil)
	noLinks := buf.String()
	be.True(t, !strings.Contains(noLinks, `named-group-page-1`))
	be.True(t, strings.Contains(noLinks, `BBS`))
	be.True(t, !strings.Contains(noLinks, `id="batch-open-menu"`))
	be.True(t, !strings.Contains(noLinks, `data-artifact-url=`))
}

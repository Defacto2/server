package app_test

import (
	"io/fs"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

func TestMilestone(t *testing.T) {
	t.Parallel()

	be.Equal(t, app.Milestones(nil).Count(), 0)

	ms := app.Collection()

	const expectedMileStones = 100
	be.True(t, ms.Count() > expectedMileStones)

	one := ms[0]
	const expectedYear = 1971
	be.Equal(t, one.Year, expectedYear)
	be.Equal(t, one.Title, "Secrets of the Little Blue Box")

	prevYear := 0
	for i, record := range ms {
		be.True(t, record.Year >= expectedYear)
		be.True(t, record.Year >= prevYear)
		prevYear = record.Year

		be.True(t, record.Month >= 0 && record.Month <= 12)
		be.True(t, record.Day >= 0 && record.Day <= 31)
		if record.Day > 0 {
			be.True(t, record.Month > 0)
			date := time.Date(record.Year, time.Month(record.Month), record.Day, 0, 0, 0, 0, time.UTC)
			be.Equal(t, date.Year(), record.Year)
			be.Equal(t, int(date.Month()), record.Month)
			be.Equal(t, date.Day(), record.Day)
		}

		if (record.Link != "") != (record.LinkTitle != "") {
			t.Errorf("milestone[%d] (%s) Link=%q and LinkTitle=%q must both be set or empty",
				i, record.Title, record.Link, record.LinkTitle)
		}
		if record.Link != "" && !validMilestoneLink(record.Link) {
			t.Errorf("milestone[%d] (%s) invalid Link %q", i, record.Title, record.Link)
		}

		for j, item := range record.List {
			if item.LinkTitle == "" || !validMilestoneLink(item.Link) {
				t.Errorf("milestone[%d] (%s) List[%d] invalid entry: Title=%q Link=%q",
					i, record.Title, j, item.LinkTitle, item.Link)
			}
		}
	}
}

func TestMilestoneMedia(t *testing.T) {
	t.Parallel()

	fsys := testutil.ProjectFS(t)

	for i, m := range app.Collection() {
		p := m.Picture
		media := []struct {
			field string
			file  string
			ext   string
		}{
			{"Jpg", p.Jpg, ".jpg"},
			{"Png", p.Png, ".png"},
			{"Webp", p.Webp, ".webp"},
			{"Avif", p.Avif, ".avif"},
			{"Webm", p.Webm, ".webm"},
		}
		hasMedia := false
		for _, item := range media {
			if item.file == "" {
				continue
			}
			hasMedia = true
			gotExt := strings.ToLower(filepath.Ext(item.file))
			if gotExt != item.ext {
				t.Errorf("milestone[%d] (%s) Picture.%s = %q: want extension %s, got %s",
					i, m.Title, item.field, item.file, item.ext, gotExt)
			}
			assetPath := "public/image/milestone/" + item.file
			if _, err := fs.Stat(fsys, assetPath); err != nil {
				t.Errorf("milestone[%d] (%s) Picture.%s missing file %s: %v",
					i, m.Title, item.field, assetPath, err)
			}
		}
		if hasMedia && strings.TrimSpace(p.Alt) == "" {
			t.Errorf("milestone[%d] (%s) has media but empty Picture.Alt", i, m.Title)
		}
		if p.Webm == "" && (p.Avif != "" || p.Webp != "") && p.Jpg == "" && p.Png == "" {
			t.Errorf("milestone[%d] (%s) has Avif/Webp without Jpg/Png fallback", i, m.Title)
		}
		if p.LicenseLink != "" && !validMilestoneLink(p.LicenseLink) {
			t.Errorf("milestone[%d] (%s) invalid Picture.LicenseLink %q", i, m.Title, p.LicenseLink)
		}
	}
}

func TestMilestoneAnchors(t *testing.T) {
	t.Parallel()

	fsys := testutil.ProjectFS(t)
	ms := app.Collection()

	reID := regexp.MustCompile(`id="([^"]+)"`)
	reHref := regexp.MustCompile(`href="([^"]*)"`)
	reFootnote := regexp.MustCompile(`href="#([^"]*-fn(\d+))">\[(\d+)\]</a>`)

	balancedTags := []string{"p", "a", "span", "li", "ul", "div", "section", "em", "strong", "q", "sup", "small", "abbr", "pre"}
	tagChecks := make([]struct {
		name  string
		open  *regexp.Regexp
		close *regexp.Regexp
	}, len(balancedTags))
	for i, tag := range balancedTags {
		tagChecks[i].name = tag
		tagChecks[i].open = regexp.MustCompile(`<` + tag + `[\s>]`)
		tagChecks[i].close = regexp.MustCompile(`</` + tag + `>`)
	}

	validIDs := make(map[string]bool)
	seenSlugs := make(map[string]bool)
	for i, m := range ms {
		if m.Title != "" {
			slug := helper.Slug(m.Title)
			if slug == "the-first-copy-protections" {
				slug += "-" + strconv.Itoa(m.Year)
			}
			if seenSlugs[slug] {
				t.Errorf("milestone[%d] (%s) duplicate slug %q", i, m.Title, slug)
			}
			seenSlugs[slug] = true
			validIDs[slug] = true
		}
		for _, match := range reID.FindAllStringSubmatch(m.Content, -1) {
			validIDs[match[1]] = true
		}
		for _, chk := range tagChecks {
			openCount := len(chk.open.FindAllStringIndex(m.Content, -1))
			closeCount := len(chk.close.FindAllStringIndex(m.Content, -1))
			if openCount != closeCount {
				t.Errorf("milestone[%d] (%s) unbalanced <%s> tags: %d open vs %d close",
					i, m.Title, chk.name, openCount, closeCount)
			}
		}
		for _, match := range reFootnote.FindAllStringSubmatch(m.Content, -1) {
			if match[2] != match[3] {
				t.Errorf("milestone[%d] (%s) footnote #%s has mismatched label [%s]",
					i, m.Title, match[1], match[3])
			}
		}
	}

	for i, m := range ms {
		for _, match := range reHref.FindAllStringSubmatch(m.Content, -1) {
			href := match[1]
			if href == "" || strings.TrimSpace(href) != href {
				t.Errorf("milestone[%d] (%s) empty or padded href %q", i, m.Title, href)
				continue
			}
			if target, ok := strings.CutPrefix(href, "#"); ok {
				if !validIDs[target] {
					t.Errorf("milestone[%d] (%s) broken internal anchor #%s", i, m.Title, target)
				}
				continue
			}
			if !validMilestoneLink(href) {
				t.Errorf("milestone[%d] (%s) invalid content href %q", i, m.Title, href)
			}
		}
	}

	tmplBytes, err := fs.ReadFile(fsys, "view/app/index.tmpl")
	be.Err(t, err, nil)

	reQuickJump := regexp.MustCompile(`data-milestone-id="([^"]+)"`)
	matches := reQuickJump.FindAllStringSubmatch(string(tmplBytes), -1)
	be.True(t, len(matches) > 0)
	for _, match := range matches {
		if !validIDs[match[1]] {
			t.Errorf("index.tmpl quick-jump data-milestone-id=%q does not match any milestone anchor", match[1])
		}
	}
}

func validMilestoneLink(raw string) bool {
	if raw == "" || strings.TrimSpace(raw) != raw {
		return false
	}
	if strings.HasPrefix(raw, "/") {
		return !strings.ContainsAny(raw, " \t\n")
	}
	parsed, err := url.ParseRequestURI(raw)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}

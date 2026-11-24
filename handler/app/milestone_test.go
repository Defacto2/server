package app_test

import (
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/handler/app"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

func TestMilestone(t *testing.T) {
	t.Parallel()

	ms := app.Collection()

	const expectedMileStones = 100
	be.True(t, ms.Count() > expectedMileStones)

	one := ms[0]
	const expectedYear = 1971
	be.Equal(t, expectedYear, one.Year)
	be.Equal(t, "Secrets of the Little Blue Box", one.Title)

	for _, record := range ms {
		be.True(t, record.Year != 0)
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
	}
}

func TestMilestoneAnchors(t *testing.T) {
	t.Parallel()

	fsys := testutil.ProjectFS(t)

	reID := regexp.MustCompile(`id="([^"]+)"`)
	reHashHref := regexp.MustCompile(`href="#([^"]+)"`)
	reOpenP := regexp.MustCompile(`<p[\s>]`)
	reCloseP := regexp.MustCompile(`</p>`)

	validIDs := make(map[string]bool)
	for i, m := range app.Collection() {
		if m.Title != "" {
			slug := helper.Slug(m.Title)
			if slug == "the-first-copy-protections" {
				slug += "-" + strconv.Itoa(m.Year)
			}
			validIDs[slug] = true
		}
		for _, match := range reID.FindAllStringSubmatch(m.Content, -1) {
			validIDs[match[1]] = true
		}
		openP := len(reOpenP.FindAllStringIndex(m.Content, -1))
		closeP := len(reCloseP.FindAllStringIndex(m.Content, -1))
		if openP != closeP {
			t.Errorf("milestone[%d] (%s) unbalanced <p> tags: %d open vs %d close",
				i, m.Title, openP, closeP)
		}
	}

	for i, m := range app.Collection() {
		for _, match := range reHashHref.FindAllStringSubmatch(m.Content, -1) {
			if !validIDs[match[1]] {
				t.Errorf("milestone[%d] (%s) broken internal anchor #%s", i, m.Title, match[1])
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

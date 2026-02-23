package tags_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/tags"
	"github.com/nalgeon/be"
)

const (
	firstCategory = "announcements"
	lastCategory  = "releaseinstall"
	firstPlatform = "ansi"
	lastPlatform  = "windows"
	noname        = "non-existing-name"
)

func TestByName(t *testing.T) {
	t.Parallel()
	tt := tags.T{}
	x, err := tt.ByName("")
	be.Err(t, err)
	be.Equal(t, x, tags.TagData{})
}

func TestTBuild(t *testing.T) {
	t.Parallel()
	tt := tags.T{}
	err := tt.Build(t.Context(), nil)
	be.Err(t, err)
}

func TestNameByURI(t *testing.T) {
	t.Parallel()
	uri := "programmingtool"
	expected := "computer tool"
	name := tags.NameByURI(uri)
	be.Equal(t, expected, name)

	errs := "error: unknown slug"
	name = tags.NameByURI(noname)
	be.True(t, strings.Contains(name, errs))
}

func TestDescription(t *testing.T) {
	t.Parallel()
	tag := "announcements"
	expected := "public announcements by Scene groups and organizations"

	desc, err := tags.Description(tag)
	be.Err(t, err, nil)
	be.Equal(t, expected, desc)

	desc, err = tags.Description(noname)
	be.True(t, errors.Is(err, tags.ErrTag))
	be.Equal(t, desc, "")
}

func TestPlatform(t *testing.T) {
	t.Parallel()
	tag := "announcements"
	platform := "ansi"
	expected := "an ansi announcement"

	desc, err := tags.Platform(platform, tag)
	be.Err(t, err, nil)
	be.Equal(t, expected, desc)

	desc, err = tags.Platform(platform, noname)
	be.True(t, errors.Is(err, tags.ErrTag))
	be.Equal(t, desc, "")

	desc, err = tags.Platform(noname, platform)
	be.True(t, errors.Is(err, tags.ErrPlatform))
	be.Equal(t, desc, "")

	desc, err = tags.Platform(noname, noname)
	be.True(t, errors.Is(err, tags.ErrPlatform))
	be.Equal(t, desc, "")
}

func TestHumanize(t *testing.T) {
	t.Parallel()
	s := tags.Humanize(-1, -1)
	be.True(t, strings.Contains(s, "unknown"))
	s = tags.Humanize(tags.ANSI, -1)
	be.True(t, strings.Contains(s, "unknown"))
	s = tags.Humanize(tags.ANSI, tags.News)
	be.Equal(t, "an ansi from a news outlet", s)

	s = tags.Humanize(tags.ANSI, tags.Restrict)
	be.Equal(t, "an insider ansi textfile", s)

	s = tags.Humanize(tags.Video, tags.Intro)
	be.Equal(t, "a bumper video", s)

	s = tags.Humanize(tags.ANSI, tags.Interview)
	be.Equal(t, "an ansi interview", s)

	s = tags.Humanize(tags.Audio, tags.Intro)
	be.Equal(t, "a chiptune or intro music", s)

	s = tags.Humanize(tags.DataB, tags.Nfo)
	be.Equal(t, "a database of releases", s)

	s = tags.Humanize(tags.DOS, tags.Demo)
	be.Equal(t, "a demo on MS Dos", s)

	s = tags.Humanize(tags.Markup, tags.Nfo)
	be.Equal(t, "a nfo file or scene release in HTML", s)

	s = tags.Humanize(tags.Image, tags.Nfo)
	be.Equal(t, "an image nfo file or scene release", s)

	s = tags.Humanize(tags.PDF, tags.Proof)
	be.Equal(t, "a PDF document about release proof", s)

	s = tags.Humanize(tags.Text, tags.Nfo)
	be.Equal(t, "an nfo textfile", s)

	s = tags.Humanize(tags.TextAmiga, tags.Nfo)
	be.Equal(t, "an Amiga nfo textfile", s)

	s = tags.Humanize(tags.Video, tags.Guide)
	be.Equal(t, "a guide or how-to video", s)

	s = tags.Humanize(tags.Windows, tags.Demo)
	be.Equal(t, "a demo on Windows", s)

	s = tags.Humanize(tags.Linux, tags.Install)
	be.Equal(t, "a Linux scene software installer", s)

	s = tags.Humanize(tags.ANSI, tags.Logo)
	be.Equal(t, "an ansi logo", s)

	s = tags.Humanize(tags.Image, tags.Proof)
	be.Equal(t, "a proof of release photo", s)

	s = tags.Humanize(tags.Image, tags.News)
	be.Equal(t, "a screenshot of an article from a news outlet", s)
}

func TestHumanizes(t *testing.T) {
	t.Parallel()
	s := tags.Humanizes(-1, -1)
	be.Equal(t, "all files", s)
	s = tags.Humanizes(tags.ANSI, -1)
	be.Equal(t, "ansi format textfiles", s)
	s = tags.Humanizes(-1, tags.News)
	be.Equal(t, "articles from mainstream news outlets", s)
	const aft = "ansi format textfiles"
	s = tags.Humanizes(tags.ANSI, tags.News)
	be.Equal(t, aft, s)
	s = tags.Humanizes(tags.ANSI, tags.Restrict)
	be.Equal(t, aft, s)
	s = tags.Humanizes(tags.Video, tags.Intro)
	be.Equal(t, "videos and animations", s)
	s = tags.Humanizes(tags.ANSI, tags.Interview)
	be.Equal(t, aft, s)
	s = tags.Humanizes(tags.Audio, tags.Intro)
	be.Equal(t, "music, chiptunes and audio samples", s)
	s = tags.Humanizes(tags.DataB, tags.Nfo)
	be.Equal(t, "databases of releases", s)
	s = tags.Humanizes(tags.DOS, tags.Demo)
	be.Equal(t, "demos on MS Dos", s)
	s = tags.Humanizes(tags.Markup, tags.Nfo)
	be.Equal(t, "nfo file or scene release as HTML files", s)
	s = tags.Humanizes(tags.Image, tags.Nfo)
	be.Equal(t, "images, pictures and photos", s)
	s = tags.Humanizes(tags.PDF, tags.Proof)
	be.Equal(t, "release proof as PDF documents", s)
	s = tags.Humanizes(tags.Text, tags.Nfo)
	be.Equal(t, "nfo textfiles", s)
	s = tags.Humanizes(tags.TextAmiga, tags.Nfo)
	be.Equal(t, "Amiga nfo textfiles", s)
	s = tags.Humanizes(tags.Video, tags.Guide)
	be.Equal(t, "videos and animations", s)
	s = tags.Humanizes(tags.Windows, tags.Demo)
	be.Equal(t, "demos on Windows", s)
	s = tags.Humanizes(tags.Linux, tags.Install)
	be.Equal(t, "scene software installer programs for Linux and BSD", s)
	s = tags.Humanizes(tags.ANSI, tags.Logo)
	be.Equal(t, "ansi format logos", s)
	s = tags.Humanizes(tags.Image, tags.Proof)
	be.Equal(t, "proof of release photos", s)
	s = tags.Humanizes(tags.Image, tags.News)
	be.Equal(t, "images, pictures and photos", s)
}

func TestIsCategory(t *testing.T) {
	t.Parallel()
	t.Run("Existing Category", func(t *testing.T) {
		t.Parallel()
		result := tags.IsCategory(firstCategory)
		be.True(t, result)
	})
	t.Run("Last Category", func(t *testing.T) {
		t.Parallel()
		result := tags.IsCategory(lastCategory)
		be.True(t, result)
	})
	t.Run("Existing Platform", func(t *testing.T) {
		t.Parallel()
		result := tags.IsCategory("ansi")
		be.True(t, !result)
	})
	t.Run("Non-existing Category", func(t *testing.T) {
		t.Parallel()
		result := tags.IsCategory(noname)
		be.True(t, !result)
	})
}

func TestIsPlatform(t *testing.T) {
	t.Parallel()
	t.Run("Existing Platform", func(t *testing.T) {
		t.Parallel()
		result := tags.IsPlatform(firstPlatform)
		be.True(t, result)
	})
	t.Run("Last Platform", func(t *testing.T) {
		t.Parallel()
		result := tags.IsPlatform(lastPlatform)
		be.True(t, result)
	})
	t.Run("Existing Category", func(t *testing.T) {
		t.Parallel()
		result := tags.IsPlatform(lastCategory)
		be.True(t, !result)
	})
	t.Run("Non-existing Platform", func(t *testing.T) {
		t.Parallel()
		result := tags.IsPlatform("non-existing-platform")
		be.True(t, !result)
	})
}

func TestIsTag(t *testing.T) {
	t.Parallel()
	t.Run("Existing Category", func(t *testing.T) {
		t.Parallel()
		result := tags.IsTag(firstCategory)
		be.True(t, result)
	})
	t.Run("Existing Platform", func(t *testing.T) {
		t.Parallel()
		result := tags.IsTag(lastPlatform)
		be.True(t, result)
	})
	t.Run("Non-existing Tag", func(t *testing.T) {
		t.Parallel()
		result := tags.IsTag("non-existing-tag")
		be.True(t, !result)
	})
}

func TestOSTags(t *testing.T) {
	t.Parallel()
	oses := tags.OSTags()
	be.Equal(t, "dos", oses[0])
	be.Equal(t, "mac10", oses[4])
}

func TestCategoryCounts(t *testing.T) {
	t.Parallel()
	// Verify that all categories from FirstCategory to LastCategory fit in CategoryCount
	categoryCount := tags.CategoryCount
	list := tags.List()
	categories := 0
	for _, tag := range list {
		if tags.IsCategory(tag.String()) {
			categories++
		}
	}
	be.Equal(t, categories, categoryCount)
}

func TestPlatformCounts(t *testing.T) {
	t.Parallel()
	// Verify that all platforms from FirstPlatform to LastPlatform fit in PlatformCount
	platformCount := tags.PlatformCount
	list := tags.List()
	platforms := 0
	for _, tag := range list {
		if tags.IsPlatform(tag.String()) {
			platforms++
		}
	}
	be.Equal(t, platforms, platformCount)
}

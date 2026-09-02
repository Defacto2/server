package tags_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

// checked in Aug 26, test coverage was great at 90%+

const (
	firstCategory = "announcements"
	lastCategory  = "releaseinstall"
	firstPlatform = "ansi"
	lastPlatform  = "windows"
	noname        = "non-existing-name"
	expectedCount = 43
)

func TestBuild(t *testing.T) {
	t.Parallel()

	var obj tags.T
	got := obj.Build(t.Context(), nil)
	be.Err(t, got)

	db := testutil.DB(t)
	got = obj.Build(t.Context(), db)
	be.Err(t, got, nil)
	be.True(t, len(obj.List) > 1)

	for _, tag := range tags.List {
		listed, err := obj.ByName(tag.String())
		be.Err(t, err, nil)
		if listed.Count == 0 {
			continue
		}
		be.True(t, listed.URI != "")
		be.True(t, listed.Name != "")
		be.True(t, listed.Info != "")
		be.True(t, listed.Count > 1)
	}
}

func TestInvalidExec(t *testing.T) {
	t.Parallel()

	be.True(t, tags.InvalidExec(nil))

	db := testutil.DB(t)
	be.True(t, !tags.InvalidExec(db))
}

func TestIsText(t *testing.T) {
	t.Parallel()

	be.True(t, tags.IsText(tags.Text.String()))
}

func TestTagStrings(t *testing.T) {
	t.Parallel()

	uris := tags.URIs()
	names := tags.Names()
	determiner := tags.Determiner()
	infos := tags.Infos()

	be.True(t, len(uris) == expectedCount)
	be.True(t, len(names) == expectedCount)
	be.True(t, len(determiner) == expectedCount)
	be.True(t, len(infos) == expectedCount)

	for i := range expectedCount {
		if i == 0 {
			continue
		}
		x := tags.Tag(i)
		be.True(t, uris[x] != "")
		be.True(t, names[x] != "")
		be.True(t, determiner[x] != "")
		be.True(t, infos[x] != "")
	}
}

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
	be.Equal(t, s, "an ansi from a news outlet")

	s = tags.Humanize(tags.ANSI, tags.Restrict)
	be.Equal(t, s, "an insider ansi textfile")

	s = tags.Humanize(tags.Video, tags.Intro)
	be.Equal(t, s, "a bumper video")

	s = tags.Humanize(tags.ANSI, tags.Interview)
	be.Equal(t, s, "an ansi interview")

	s = tags.Humanize(tags.Audio, tags.Intro)
	be.Equal(t, s, "chiptune or scene music")

	s = tags.Humanize(tags.DataB, tags.Nfo)
	be.Equal(t, s, "a database of releases")

	s = tags.Humanize(tags.DOS, tags.Demo)
	be.Equal(t, s, "a demo on MS Dos")

	s = tags.Humanize(tags.Markup, tags.Nfo)
	be.Equal(t, s, "a nfo file or scene release in HTML")

	s = tags.Humanize(tags.Image, tags.Nfo)
	be.Equal(t, s, "an image nfo file or scene release")

	s = tags.Humanize(tags.PDF, tags.Proof)
	be.Equal(t, s, "a PDF document about release proof")

	s = tags.Humanize(tags.Text, tags.Nfo)
	be.Equal(t, s, "a release textfile")

	s = tags.Humanize(tags.TextAmiga, tags.Nfo)
	be.Equal(t, s, "an amiga or console release textfile")

	s = tags.Humanize(tags.Video, tags.Guide)
	be.Equal(t, s, "a guide or how-to video")

	s = tags.Humanize(tags.Windows, tags.Demo)
	be.Equal(t, s, "a demo on Windows")

	s = tags.Humanize(tags.Linux, tags.Install)
	be.Equal(t, s, "a Linux scene software installer")

	s = tags.Humanize(tags.ANSI, tags.Logo)
	be.Equal(t, s, "an ansi logo")

	s = tags.Humanize(tags.Image, tags.Proof)
	be.Equal(t, s, "a proof of release photo")

	s = tags.Humanize(tags.Image, tags.News)
	be.Equal(t, s, "a screenshot of an article from a news outlet")
}

func TestHumanizes(t *testing.T) {
	t.Parallel()

	none := tags.Tag(-1)
	const at = "texts in an ansi format"
	s := none.Humanizes(none)
	be.Equal(t, s, "all files")

	s = tags.ANSI.Humanizes(none)
	be.Equal(t, s, at)

	s = none.Humanizes(tags.News)
	be.Equal(t, s, "reprinted articles from media outlets")

	s = tags.ANSI.Humanizes(tags.News)
	be.Equal(t, s, at)

	s = tags.ANSI.Humanizes(tags.Restrict)
	be.Equal(t, s, at)

	s = tags.Video.Humanizes(tags.Intro)
	be.Equal(t, s, "videos and animations")

	s = tags.ANSI.Humanizes(tags.Interview)
	be.Equal(t, s, at)

	s = tags.Audio.Humanizes(tags.Intro)
	be.Equal(t, s, "music, chiptunes, and audio")

	s = tags.DataB.Humanizes(tags.Nfo)
	be.Equal(t, s, "databases of releases")

	s = tags.DOS.Humanizes(tags.Demo)
	be.Equal(t, s, "demos on MS Dos")

	s = tags.Markup.Humanizes(tags.Nfo)
	be.Equal(t, s, "nfo file or scene release as HTML files")

	s = tags.Image.Humanizes(tags.Nfo)
	be.Equal(t, s, "images, pictures, and photos")

	s = tags.PDF.Humanizes(tags.Proof)
	be.Equal(t, s, "release proof as PDF documents")

	s = tags.Text.Humanizes(tags.Nfo)
	be.Equal(t, s, "release textfiles")

	s = tags.TextAmiga.Humanizes(tags.Nfo)
	be.Equal(t, s, "amiga/console text infos")

	s = tags.Video.Humanizes(tags.Guide)
	be.Equal(t, s, "videos and animations")

	s = tags.Windows.Humanizes(tags.Demo)
	be.Equal(t, s, "demos on Windows")

	s = tags.Linux.Humanizes(tags.Install)
	be.Equal(t, s, "scene software installer programs on Linux and Unix")

	s = tags.ANSI.Humanizes(tags.Logo)
	be.Equal(t, s, "logos in an ansi format")

	s = tags.Image.Humanizes(tags.Proof)
	be.Equal(t, s, "photos used to prove a release")

	s = tags.Image.Humanizes(tags.News)
	be.Equal(t, s, "images, pictures, and photos")
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
	categories := 0
	for _, tag := range tags.List {
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
	platforms := 0
	for _, tag := range tags.List {
		if tags.IsPlatform(tag.String()) {
			platforms++
		}
	}
	be.Equal(t, platforms, platformCount)
}

func TestHumanizeRange(t *testing.T) {
	t.Parallel()

	for _, platform := range tags.List {
		for _, section := range tags.List {
			got := tags.Humanize(platform, section)
			be.True(t, got != "")
		}
	}
}

func TestHumanizesRange(t *testing.T) {
	t.Parallel()

	for _, platform := range tags.List {
		for _, section := range tags.List {
			got := platform.Humanizes(section)
			be.True(t, got != "")
		}
	}
}

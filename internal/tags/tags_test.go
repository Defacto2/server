package tags_test

import (
	"testing"

	"github.com/Defacto2/server/internal/tags"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	firstCategory = "announcements"
	lastCategory  = "releaseinstall"
	firstPlatform = "ansi"
	lastPlatform  = "windows"
	noname        = "non-existing-name"
)

func TestNameByURI(t *testing.T) {
	uri := "programmingtool"
	expected := "computer tool"
	name := tags.NameByURI(uri)
	assert.Equal(t, expected, name)

	errs := "error: unknown slug"
	name = tags.NameByURI(noname)
	assert.Contains(t, name, errs)
}

func TestDescription(t *testing.T) {
	tag := "announcements"
	expected := "public announcements by Scene groups and organizations"

	desc, err := tags.Description(tag)
	require.NoError(t, err)
	assert.Equal(t, expected, desc)

	desc, err = tags.Description(noname)
	require.ErrorIs(t, err, tags.ErrTag)
	assert.Empty(t, desc)
}

func TestPlatform(t *testing.T) {
	tag := "announcements"
	platform := "ansi"
	expected := "an ansi announcement"

	desc, err := tags.Platform(platform, tag)
	require.NoError(t, err)
	assert.Equal(t, expected, desc)

	desc, err = tags.Platform(platform, noname)
	require.ErrorIs(t, err, tags.ErrTag)
	assert.Empty(t, desc)

	desc, err = tags.Platform(noname, platform)
	require.ErrorIs(t, err, tags.ErrPlatform)
	assert.Empty(t, desc)

	desc, err = tags.Platform(noname, noname)
	require.ErrorIs(t, err, tags.ErrPlatform)
	assert.Empty(t, desc)
}

func TestHumanize(t *testing.T) {
	t.Parallel()
	s := tags.Humanize(-1, -1)
	assert.Contains(t, s, "unknown")
	s = tags.Humanize(tags.ANSI, -1)
	assert.Contains(t, s, "unknown")
	s = tags.Humanize(tags.ANSI, tags.News)
	assert.Equal(t, "an ansi from a news outlet", s)

	s = tags.Humanize(tags.ANSI, tags.Restrict)
	assert.Equal(t, "an insider ansi textfile", s)

	s = tags.Humanize(tags.Video, tags.Intro)
	assert.Equal(t, "a bumper video", s)

	s = tags.Humanize(tags.ANSI, tags.Interview)
	assert.Equal(t, "an ansi interview", s)

	s = tags.Humanize(tags.Audio, tags.Intro)
	assert.Equal(t, "a chiptune or intro music", s)

	s = tags.Humanize(tags.DataB, tags.Nfo)
	assert.Equal(t, "a database of releases", s)

	s = tags.Humanize(tags.DOS, tags.Demo)
	assert.Equal(t, "a demo on MS Dos", s)

	s = tags.Humanize(tags.Markup, tags.Nfo)
	assert.Equal(t, "a nfo file or scene release in HTML", s)

	s = tags.Humanize(tags.Image, tags.Nfo)
	assert.Equal(t, "an image nfo file or scene release", s)

	s = tags.Humanize(tags.PDF, tags.Proof)
	assert.Equal(t, "a release proof as a PDF document", s)

	s = tags.Humanize(tags.Text, tags.Nfo)
	assert.Equal(t, "an nfo textfile", s)

	s = tags.Humanize(tags.TextAmiga, tags.Nfo)
	assert.Equal(t, "an Amiga nfo textfile", s)

	s = tags.Humanize(tags.Video, tags.Guide)
	assert.Equal(t, "a guide or how-to video", s)

	s = tags.Humanize(tags.Windows, tags.Demo)
	assert.Equal(t, "a demo on Windows", s)

	s = tags.Humanize(tags.Linux, tags.Install)
	assert.Equal(t, "a Linux scene software installer", s)

	s = tags.Humanize(tags.ANSI, tags.Logo)
	assert.Equal(t, "an ansi logo", s)

	s = tags.Humanize(tags.Image, tags.Proof)
	assert.Equal(t, "a proof of release photo", s)

	s = tags.Humanize(tags.Image, tags.News)
	assert.Equal(t, "a screenshot of an article from a news outlet", s)
}

func TestHumanizes(t *testing.T) {
	t.Parallel()
	s := tags.Humanizes(-1, -1)
	assert.Equal(t, "all files", s)
	s = tags.Humanizes(tags.ANSI, -1)
	assert.Equal(t, "ansi format textfiles", s)
	s = tags.Humanizes(-1, tags.News)
	assert.Equal(t, "articles from mainstream news outlets", s)
	const aft = "ansi format textfiles"
	s = tags.Humanizes(tags.ANSI, tags.News)
	assert.Equal(t, aft, s)
	s = tags.Humanizes(tags.ANSI, tags.Restrict)
	assert.Equal(t, aft, s)
	s = tags.Humanizes(tags.Video, tags.Intro)
	assert.Equal(t, "videos and animations", s)
	s = tags.Humanizes(tags.ANSI, tags.Interview)
	assert.Equal(t, aft, s)
	s = tags.Humanizes(tags.Audio, tags.Intro)
	assert.Equal(t, "music, chiptunes and audio samples", s)
	s = tags.Humanizes(tags.DataB, tags.Nfo)
	assert.Equal(t, "databases of releases", s)
	s = tags.Humanizes(tags.DOS, tags.Demo)
	assert.Equal(t, "demos on MS Dos", s)
	s = tags.Humanizes(tags.Markup, tags.Nfo)
	assert.Equal(t, "nfo file or scene release as HTML files", s)
	s = tags.Humanizes(tags.Image, tags.Nfo)
	assert.Equal(t, "images, pictures and photos", s)
	s = tags.Humanizes(tags.PDF, tags.Proof)
	assert.Equal(t, "release proof as PDF documents", s)
	s = tags.Humanizes(tags.Text, tags.Nfo)
	assert.Equal(t, "nfo textfiles", s)
	s = tags.Humanizes(tags.TextAmiga, tags.Nfo)
	assert.Equal(t, "Amiga nfo textfiles", s)
	s = tags.Humanizes(tags.Video, tags.Guide)
	assert.Equal(t, "videos and animations", s)
	s = tags.Humanizes(tags.Windows, tags.Demo)
	assert.Equal(t, "demos on Windows", s)
	s = tags.Humanizes(tags.Linux, tags.Install)
	assert.Equal(t, "scene software installer programs for Linux and BSD", s)
	s = tags.Humanizes(tags.ANSI, tags.Logo)
	assert.Equal(t, "ansi format logos", s)
	s = tags.Humanizes(tags.Image, tags.Proof)
	assert.Equal(t, "proof of release photos", s)
	s = tags.Humanizes(tags.Image, tags.News)
	assert.Equal(t, "images, pictures and photos", s)
}

func TestTags(t *testing.T) {
	count := 0
	for i, uri := range tags.URIs() {
		count++
		// confirm tag uri strings
		assert.NotEqual(t, "", uri)
		// confirm tag names
		assert.NotEqual(t, "", tags.Names()[i])
		// confirm tag descriptions
		assert.NotEqual(t, "", tags.Infos()[i])
	}
	u := tags.TagByURI
	assert.Equal(t, int(u("")), -1)
	assert.Equal(t, 0, int(u(firstCategory)))
	assert.Equal(t, firstCategory, tags.Announcement.String())
	assert.Equal(t, tags.CategoryCount, 26)
	assert.Equal(t, tags.PlatformCount, 16)
	assert.Equal(t, tags.CategoryCount+tags.PlatformCount, count)
}

func TestIsCategory(t *testing.T) {
	t.Run("Existing Category", func(t *testing.T) {
		result := tags.IsCategory(firstCategory)
		assert.True(t, result, "Expected true for existing category")
	})
	t.Run("Last Category", func(t *testing.T) {
		result := tags.IsCategory(lastCategory)
		assert.True(t, result, "Expected true for last category")
	})
	t.Run("Existing Platform", func(t *testing.T) {
		result := tags.IsCategory("ansi")
		assert.False(t, result, "Expected false for existing platform")
	})
	t.Run("Non-existing Category", func(t *testing.T) {
		result := tags.IsCategory(noname)
		assert.False(t, result, "Expected false for non-existing category")
	})
}

func TestIsPlatform(t *testing.T) {
	t.Run("Existing Platform", func(t *testing.T) {
		result := tags.IsPlatform(firstPlatform)
		assert.True(t, result, "Expected true for existing platform")
	})
	t.Run("Last Platform", func(t *testing.T) {
		result := tags.IsPlatform(lastPlatform)
		assert.True(t, result, "Expected true for last platform")
	})
	t.Run("Existing Category", func(t *testing.T) {
		result := tags.IsPlatform(lastCategory)
		assert.False(t, result, "Expected false for existing category")
	})
	t.Run("Non-existing Platform", func(t *testing.T) {
		result := tags.IsPlatform("non-existing-platform")
		assert.False(t, result, "Expected false for non-existing platform")
	})
}

func TestIsTag(t *testing.T) {
	t.Run("Existing Category", func(t *testing.T) {
		result := tags.IsTag(firstCategory)
		assert.True(t, result, "Expected true for existing category")
	})
	t.Run("Existing Platform", func(t *testing.T) {
		result := tags.IsTag(lastPlatform)
		assert.True(t, result, "Expected true for existing platform")
	})
	t.Run("Non-existing Tag", func(t *testing.T) {
		result := tags.IsTag("non-existing-tag")
		assert.False(t, result, "Expected false for non-existing tag")
	})
}

func TestOSTags(t *testing.T) {
	t.Parallel()
	oses := tags.OSTags()
	assert.Equal(t, "dos", oses[0])
	assert.Equal(t, "mac10", oses[4])
}

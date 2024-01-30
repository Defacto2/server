package tags_test

import (
	"testing"

	"github.com/Defacto2/server/internal/tags"
	"github.com/stretchr/testify/assert"
)

const (
	firstCategory = "announcements"
	lastCategory  = "releaseinstall"
	firstPlatform = "ansi"
	lastPlatform  = "windows"
	noname        = "non-existing-name"
)

func TestByNames(t *testing.T) {
	t.Parallel()
	tee := tags.T{}
	data, err := tee.ByName("")
	assert.Error(t, err)
	assert.Empty(t, data)

	data, err = tee.ByName("non-existing-name")
	assert.NoError(t, err)
	assert.Empty(t, data)

	data, err = tee.ByName("bbs")
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestHumanize(t *testing.T) {
	t.Parallel()
	s := tags.Humanize(-1, -1)
	assert.Contains(t, s, "unknown")
	s = tags.Humanize(tags.ANSI, -1)
	assert.Contains(t, s, "unknown")

	s = tags.Humanize(tags.ANSI, tags.News)
	assert.Equal(t, s, "an ansi from a news outlet")

	s = tags.Humanize(tags.ANSI, tags.Restrict)
	assert.Equal(t, s, "an insider ansi textfile")

	s = tags.Humanize(tags.Video, tags.Intro)
	assert.Equal(t, s, "a bumper video")

	s = tags.Humanize(tags.ANSI, tags.Interview)
	assert.Equal(t, s, "an ansi interview")

	s = tags.Humanize(tags.Audio, tags.Intro)
	assert.Equal(t, s, "a chiptune or intro music")

	s = tags.Humanize(tags.DataB, tags.Nfo)
	assert.Equal(t, s, "a database of releases")

	s = tags.Humanize(tags.DOS, tags.Demo)
	assert.Equal(t, s, "a demo on MS Dos")

	s = tags.Humanize(tags.Markup, tags.Nfo)
	assert.Equal(t, s, "a nfo file or scene release in HTML")

	s = tags.Humanize(tags.Image, tags.Nfo)
	assert.Equal(t, s, "an image nfo file or scene release")

	s = tags.Humanize(tags.PDF, tags.Proof)
	assert.Equal(t, s, "a release proof as a PDF document")

	s = tags.Humanize(tags.Text, tags.Nfo)
	assert.Equal(t, s, "an nfo textfile")

	s = tags.Humanize(tags.TextAmiga, tags.Nfo)
	assert.Equal(t, s, "an Amiga nfo textfile")

	s = tags.Humanize(tags.Video, tags.Guide)
	assert.Equal(t, s, "a guide or how-to video")

	s = tags.Humanize(tags.Windows, tags.Demo)
	assert.Equal(t, s, "a demo on Windows")

	s = tags.Humanize(tags.Linux, tags.Install)
	assert.Equal(t, s, "a Linux scene software installer")

	s = tags.Humanize(tags.ANSI, tags.Logo)
	assert.Equal(t, s, "an ansi logo")

	s = tags.Humanize(tags.Image, tags.Proof)
	assert.Equal(t, s, "a proof of release photo")

	s = tags.Humanize(tags.Image, tags.News)
	assert.Equal(t, s, "a screenshot of an article from a news outlet")
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
	assert.Equal(t, s, aft)

	s = tags.Humanizes(tags.ANSI, tags.Restrict)
	assert.Equal(t, s, aft)

	s = tags.Humanizes(tags.Video, tags.Intro)
	assert.Equal(t, s, "videos and animations")

	s = tags.Humanizes(tags.ANSI, tags.Interview)
	assert.Equal(t, s, aft)

	s = tags.Humanizes(tags.Audio, tags.Intro)
	assert.Equal(t, s, "music, chiptunes and audio samples")

	s = tags.Humanizes(tags.DataB, tags.Nfo)
	assert.Equal(t, s, "databases of releases")

	s = tags.Humanizes(tags.DOS, tags.Demo)
	assert.Equal(t, s, "demos on MS Dos")

	s = tags.Humanizes(tags.Markup, tags.Nfo)
	assert.Equal(t, s, "nfo file or scene release as HTML files")

	s = tags.Humanizes(tags.Image, tags.Nfo)
	assert.Equal(t, s, "images, pictures and photos")

	s = tags.Humanizes(tags.PDF, tags.Proof)
	assert.Equal(t, s, "release proof as PDF documents")

	s = tags.Humanizes(tags.Text, tags.Nfo)
	assert.Equal(t, s, "nfo textfiles")

	s = tags.Humanizes(tags.TextAmiga, tags.Nfo)
	assert.Equal(t, s, "Amiga nfo textfiles")

	s = tags.Humanizes(tags.Video, tags.Guide)
	assert.Equal(t, s, "videos and animations")

	s = tags.Humanizes(tags.Windows, tags.Demo)
	assert.Equal(t, s, "demos on Windows")

	s = tags.Humanizes(tags.Linux, tags.Install)
	assert.Equal(t, s, "scene software installer programs for Linux and BSD")

	s = tags.Humanizes(tags.ANSI, tags.Logo)
	assert.Equal(t, s, "ansi format logos")

	s = tags.Humanizes(tags.Image, tags.Proof)
	assert.Equal(t, s, "proof of release photos")

	s = tags.Humanizes(tags.Image, tags.News)
	assert.Equal(t, s, "images, pictures and photos")

}

func TestTags(t *testing.T) {
	count := 0
	for i, uri := range tags.URIs() {
		count++
		// confirm tag uri strings
		assert.NotEqual(t, uri, "")
		// confirm tag names
		assert.NotEqual(t, tags.Names()[i], "")
		// confirm tag descriptions
		assert.NotEqual(t, tags.Infos()[i], "")
	}
	u := tags.TagByURI
	assert.Equal(t, int(u("")), -1)
	assert.Equal(t, int(u(firstCategory)), 0)
	assert.Equal(t, tags.Announcement.String(), firstCategory)
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
	assert.Equal(t, 5, len(oses))
	assert.Equal(t, "dos", oses[0])
	assert.Equal(t, "mac10", oses[4])
}

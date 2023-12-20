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

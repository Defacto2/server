package tags_test

import (
	"testing"

	"github.com/Defacto2/server/pkg/tags"
	"github.com/stretchr/testify/assert"
)

func TestTags(t *testing.T) {
	count := 0
	for i, uri := range tags.URIs {
		count++
		// confirm tag uri strings
		assert.NotEqual(t, uri, "")
		// confirm tag names
		assert.NotEqual(t, tags.Names[i], "")
		// confirm tag descriptions
		assert.NotEqual(t, tags.Infos[i], "")
	}
	u := tags.TagByURI
	assert.Equal(t, int(u("")), -1)
	assert.Equal(t, int(u("announcements")), 0)
	assert.Equal(t, tags.Announcement.String(), "announcements")
	assert.Equal(t, tags.CategoryCount, 26)
	assert.Equal(t, tags.PlatformCount, 16)
	assert.Equal(t, tags.CategoryCount+tags.PlatformCount, count)
}

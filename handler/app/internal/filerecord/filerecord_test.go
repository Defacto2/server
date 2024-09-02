package filerecord_test

import (
	"testing"

	"github.com/Defacto2/server/handler/app/internal/filerecord"
	"github.com/stretchr/testify/assert"
)

func TestLinkPreviewHref(t *testing.T) {
	t.Parallel()
	s := filerecord.LinkPreviewHref(nil, "", "")
	assert.Empty(t, s)
	s = filerecord.LinkPreviewHref(1, "filename.xxx", "invalid")
	assert.Empty(t, s)
	s = filerecord.LinkPreviewHref(1, "filename.txt", "text")
	assert.Equal(t, "/v/9b1c6", s)
}

func TestLegacyString(t *testing.T) {
	t.Parallel()
	s := filerecord.LegacyString("")
	assert.Empty(t, s)
	s = filerecord.LegacyString("Hello world 123.")
	assert.Equal(t, "Hello world 123.", s)
	s = filerecord.LegacyString("£100")
	assert.Equal(t, "£100", s)
	s = filerecord.LegacyString("\xa3100")
	assert.Equal(t, "£100", s)
	s = filerecord.LegacyString("€100")
	assert.Equal(t, "€100", s)
	s = filerecord.LegacyString("\x80100")
	assert.Equal(t, "€100", s)
}

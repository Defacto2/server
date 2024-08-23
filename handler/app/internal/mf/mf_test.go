package mf_test

import (
	"testing"

	"github.com/Defacto2/server/handler/app/internal/mf"
	"github.com/stretchr/testify/assert"
)

func TestLinkPreviewHref(t *testing.T) {
	t.Parallel()
	s := mf.LinkPreviewHref(nil, "", "")
	assert.Empty(t, s)
	s = mf.LinkPreviewHref(1, "filename.xxx", "invalid")
	assert.Empty(t, s)
	s = mf.LinkPreviewHref(1, "filename.txt", "text")
	assert.Equal(t, "/v/9b1c6", s)
}

func TestLegacyString(t *testing.T) {
	t.Parallel()
	s := mf.LegacyString("")
	assert.Empty(t, s)
	s = mf.LegacyString("Hello world 123.")
	assert.Equal(t, "Hello world 123.", s)
	s = mf.LegacyString("£100")
	assert.Equal(t, "£100", s)
	s = mf.LegacyString("\xa3100")
	assert.Equal(t, "£100", s)
	s = mf.LegacyString("€100")
	assert.Equal(t, "€100", s)
	s = mf.LegacyString("\x80100")
	assert.Equal(t, "€100", s)
}

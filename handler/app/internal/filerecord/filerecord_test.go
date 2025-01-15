package filerecord_test

import (
	"testing"

	"github.com/Defacto2/server/handler/app/internal/filerecord"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/stretchr/testify/assert"
	"github.com/volatiletech/null/v8"
)

// TODO
// func TestListContent(t *testing.T) {
// 	x := models.File{}
// 	dirs := command.Dirs{}
// 	s := filerecord.ListContent(&x, dirs, "")
// 	assert.Contains(t, s, "no UUID")

// 	r0 := "00000000-0000-0000-0000-000000000000"
// 	x.UUID = null.StringFrom(r0)
// 	s = filerecord.ListContent(&x, dirs, "")
// 	assert.Contains(t, s, "invalid platform")

// 	x.Platform = null.StringFrom("dos")
// 	s = filerecord.ListContent(&x, dirs, "")
// 	assert.Contains(t, s, "invalid platform")
// }

func TestAlertURL(t *testing.T) {
	x := models.File{}
	s := filerecord.AlertURL(&x)
	assert.Empty(t, s)
	x.FileSecurityAlertURL = null.StringFrom("invalid")
	s = filerecord.AlertURL(&x)
	assert.Empty(t, s)
	x.FileSecurityAlertURL = null.StringFrom("https://example.com")
	s = filerecord.AlertURL(&x)
	assert.Equal(t, "https://example.com", s)

}

func TestListEntry(t *testing.T) {
	t.Parallel()
	le := filerecord.ListEntry{}
	assert.Empty(t, le)
	s := le.HTML(1, "x", "y")
	assert.Contains(t, s, "1 bytes")
}

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

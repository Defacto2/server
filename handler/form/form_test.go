package form_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/Defacto2/server/handler/form"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHumanizeCount(t *testing.T) {
	t.Parallel()
	html, err := form.HumanizeCount(nil, "", "")
	require.NoError(t, err)
	assert.Contains(t, html, `0 existing artifacts`)
	htm := form.HumanizeCountStr(nil, "", "")
	require.NoError(t, err)
	assert.Contains(t, htm, `0 existing artifacts`)
}

func TestSanitizeFilename(t *testing.T) {
	t.Parallel()
	s := form.SanitizeFilename("")
	assert.Empty(t, s)
	s = form.SanitizeFilename(`c:\Windows\System32\cmd.exe`)
	assert.Equal(t, "c:-Windows-System32-cmd.exe", s)
	s = form.SanitizeFilename(`../tmp/somefile.txt`)
	assert.Equal(t, "tmp-somefile.txt", s)
}

func TestSanitizePath(t *testing.T) {
	t.Parallel()
	s := form.SanitizeSeparators("")
	assert.Empty(t, s)
	s = form.SanitizeSeparators(`///some//messy/path////`)
	assert.Equal(t, "some/messy/path", s)
}

func TestSanitizeURLPath(t *testing.T) {
	t.Parallel()
	s := form.SanitizeURLPath("")
	assert.Empty(t, s)
	s = form.SanitizeURLPath("https://example.com/some/messy/path")
	assert.Empty(t, s)

	s = form.SanitizeURLPath(`///some//messy/path////`)
	assert.Equal(t, "some/messy/path", s)

	s = form.SanitizeURLPath(`///some/!@#$#@%^@&(+/very_messy/path//*^&()//`)
	assert.Equal(t, "some/very_messy/path", s)

	s = form.SanitizeGitHub("//refs/heads/\\/<main>///")
	assert.Equal(t, "heads/main", s)
}

func TestValidDate(t *testing.T) {
	t.Parallel()
	x := time.Now().Year()
	year := strconv.Itoa(x)
	next := strconv.Itoa(x + 1)
	y, m, d := form.ValidDate("", "", "")
	assert.False(t, y)
	assert.False(t, m)
	assert.False(t, d)
	y, m, d = form.ValidDate(year, "", "")
	assert.True(t, y)
	assert.False(t, m)
	assert.False(t, d)
	y, m, d = form.ValidDate(next, "", "")
	assert.False(t, y)
	assert.False(t, m)
	assert.False(t, d)
	y, m, d = form.ValidDate(year, "-10", "")
	assert.True(t, y)
	assert.False(t, m)
	assert.False(t, d)
	y, m, d = form.ValidDate(year, "1", "")
	assert.True(t, y)
	assert.True(t, m)
	assert.False(t, d)
	y, m, d = form.ValidDate("", "1", "")
	assert.False(t, y)
	assert.True(t, m)
	assert.False(t, d)
	y, m, d = form.ValidDate(year, "30", "")
	assert.True(t, y)
	assert.False(t, m)
	assert.False(t, d)
	y, m, d = form.ValidDate("", "1", "1")
	assert.False(t, y)
	assert.True(t, m)
	assert.True(t, d)
	y, m, d = form.ValidDate(next, "13", "32")
	assert.False(t, y)
	assert.False(t, m)
	assert.False(t, d)
	y, m, d = form.ValidDate("abc", "efg", "hij")
	assert.False(t, y)
	assert.False(t, m)
	assert.False(t, d)
}

func TestValidVT(t *testing.T) {
	t.Parallel()
	assert.False(t, form.ValidVT("https://example.com"))
	assert.False(t, form.ValidVT("https://virustotal.com"))
	assert.True(t, form.ValidVT("https://www.virustotal.com/gui/file/"+
		"50c69b4e65380a0ada587656225ef260ffb9f352e1c1adb3f2222588eadf836d"))
}

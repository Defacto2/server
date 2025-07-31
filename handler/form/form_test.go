package form_test

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/server/handler/form"
	"github.com/nalgeon/be"
)

func TestHumanizeCount(t *testing.T) {
	t.Parallel()
	html, err := form.HumanizeCount(nil, "", "")
	be.Err(t, err)
	found := strings.Contains(string(html), `0 existing artifacts`)
	be.True(t, !found)
	htm := form.HumanizeCountStr(nil, "", "")
	be.Err(t, err)
	found = strings.Contains(htm, `0 existing artifacts`)
	be.True(t, found)
}

func TestSanitizeFilename(t *testing.T) {
	t.Parallel()
	s := form.SanitizeFilename("")
	be.Equal(t, s, "")
	s = form.SanitizeFilename(`c:\Windows\System32\cmd.exe`)
	be.Equal(t, "c:-Windows-System32-cmd.exe", s)
	s = form.SanitizeFilename(`../tmp/somefile.txt`)
	be.Equal(t, "tmp-somefile.txt", s)
}

func TestSanitizePath(t *testing.T) {
	t.Parallel()
	s := form.SanitizeSeparators("")
	be.Equal(t, s, "")
	s = form.SanitizeSeparators(`///some//messy/path////`)
	be.Equal(t, "some/messy/path", s)
}

func TestSanitizeURLPath(t *testing.T) {
	t.Parallel()
	s := form.SanitizeURLPath("")
	be.Equal(t, s, "")
	s = form.SanitizeURLPath("https://example.com/some/messy/path")
	be.Equal(t, s, "")

	s = form.SanitizeURLPath(`///some//messy/path////`)
	be.Equal(t, "some/messy/path", s)

	s = form.SanitizeURLPath(`///some/!@#$#@%^@&(+/very_messy/path//*^&()//`)
	be.Equal(t, "some/very_messy/path", s)

	s = form.SanitizeGitHub("//refs/heads/\\/<main>///")
	be.Equal(t, "heads/main", s)
}

func TestValidDate(t *testing.T) {
	t.Parallel()
	x := time.Now().Year()
	year := strconv.Itoa(x)
	next := strconv.Itoa(x + 1)
	y, m, d := form.ValidDate("", "", "")
	be.True(t, !y)
	be.True(t, !m)
	be.True(t, !d)
	y, m, d = form.ValidDate(year, "", "")
	be.True(t, y)
	be.True(t, !m)
	be.True(t, !d)
	y, m, d = form.ValidDate(next, "", "")
	be.True(t, !y)
	be.True(t, !m)
	be.True(t, !d)
	y, m, d = form.ValidDate(year, "-10", "")
	be.True(t, y)
	be.True(t, !m)
	be.True(t, !d)
	y, m, d = form.ValidDate(year, "1", "")
	be.True(t, y)
	be.True(t, m)
	be.True(t, !d)
	y, m, d = form.ValidDate("", "1", "")
	be.True(t, !y)
	be.True(t, m)
	be.True(t, !d)
	y, m, d = form.ValidDate(year, "30", "")
	be.True(t, y)
	be.True(t, !m)
	be.True(t, !d)
	y, m, d = form.ValidDate("", "1", "1")
	be.True(t, !y)
	be.True(t, m)
	be.True(t, d)
	y, m, d = form.ValidDate(next, "13", "32")
	be.True(t, !y)
	be.True(t, !m)
	be.True(t, !d)
	y, m, d = form.ValidDate("abc", "efg", "hij")
	be.True(t, !y)
	be.True(t, !m)
	be.True(t, !d)
}

func TestValidVT(t *testing.T) {
	t.Parallel()
	be.True(t, !form.ValidVT("https://example.com"))
	be.True(t, !form.ValidVT("https://virustotal.com"))
	be.True(t, form.ValidVT("https://www.virustotal.com/gui/file/"+
		"50c69b4e65380a0ada587656225ef260ffb9f352e1c1adb3f2222588eadf836d"))
}

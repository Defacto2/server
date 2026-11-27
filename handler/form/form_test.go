package form_test

import (
	"context"
	"database/sql"
	"html/template"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/server/handler/form"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/nalgeon/be"
)

func TestCheckname(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"local file", "archive.zip", nil},
		{"nested local", "sub/archive.zip", nil},
		{"empty", "", form.ErrFilename},
		{"traversal", "../etc/passwd", form.ErrFilename},
		{"absolute", "/etc/passwd", form.ErrFilename},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			be.Err(t, form.Checkname(tt.input), tt.wantErr)
		})
	}
}

func TestHumanizeCount(t *testing.T) {
	t.Parallel()

	t.Run("nil exec", func(t *testing.T) {
		t.Parallel()

		html, err := form.HumanizeCount(t.Context(), nil, "", "")
		be.Err(t, err)
		be.Equal(t, html, template.HTML(""))
	})

	t.Run("missing classifications", func(t *testing.T) {
		t.Parallel()

		db := new(sql.DB)
		tests := []struct {
			name     string
			section  string
			platform string
			wantText string
		}{
			{"both missing", "", "", "please choose both classifications, 0 existing artifacts"},
			{"section missing", "", "dos", "please choose a tag as category, 0 existing artifacts"},
			{"platform missing", "demo", "", "please choose an operating system, 0 existing artifacts"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				html, err := form.HumanizeCount(t.Context(), db, tt.section, tt.platform)
				be.Err(t, err, nil)
				wantHTML := template.HTML(`<span class="text-danger-emphasis">` + tt.wantText + `</span>`)
				be.Equal(t, html, wantHTML)

				str := form.HumanizeCountStr(t.Context(), nil, tt.section, tt.platform)
				be.Equal(t, str, tt.wantText)
			})
		}
	})

	t.Run("db error", func(t *testing.T) {
		t.Parallel()

		db, err := sql.Open("pgx", "")
		be.Err(t, err, nil)
		defer db.Close()

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		html, err := form.HumanizeCount(ctx, db, "demo", "dos")
		be.Err(t, err)
		be.Equal(t, html, template.HTML(""))

		str := form.HumanizeCountStr(ctx, db, "demo", "dos")
		be.True(t, strings.Contains(str, "form humanize and count classification"))
	})
}

func TestSanitizeCredit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"single", "  the x bbs  ", "X BBS"},
		{"multiple", " the x bbs , razor 1911 ", "X BBS,Razor 1911"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			be.Equal(t, form.SanitizeCredit(tt.input), tt.want)
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	t.Parallel()

	s := form.SanitizeFilename("")
	be.Equal(t, s, "")

	be.Equal(t, form.SanitizeFilename("."), "")
	be.Equal(t, form.SanitizeFilename(".."), "")

	s = form.SanitizeFilename(`c:\Windows\System32\cmd.exe`)
	be.Equal(t, s, "c:-Windows-System32-cmd.exe")

	s = form.SanitizeFilename(`../tmp/somefile.txt`)
	be.Equal(t, s, "tmp-somefile.txt")
}

func TestSanitizePath(t *testing.T) {
	t.Parallel()

	s := form.SanitizeSeparators("")
	be.Equal(t, s, "")

	s = form.SanitizeSeparators(`///some//messy/path////`)
	be.Equal(t, s, "some/messy/path")

	s = form.SanitizeSeparators(`some///very////messy/path`)
	be.Equal(t, s, "some/very/messy/path")

	s = form.SanitizeSeparators("example.com/hello/")
	be.Equal(t, s, "example.com/hello")
}

func TestSanitizeURLPath(t *testing.T) {
	t.Parallel()
	s := form.SanitizeURLPath("")
	be.Equal(t, s, "")
	s = form.SanitizeURLPath("https://example.com/some/messy/path")
	be.Equal(t, s, "")

	s = form.SanitizeURLPath(`///some//messy/path////`)
	be.Equal(t, s, "some/messy/path")

	s = form.SanitizeURLPath(`///some/!@#$#@%^@&(+/very_messy/path//*^&()//`)
	be.Equal(t, s, "some/very_messy/path")

	s = form.SanitizeGitHub("//refs/heads/\\/<main>///")
	be.Equal(t, s, "heads/main")
}

func TestValidDate(t *testing.T) {
	t.Parallel()

	x := time.Now().Year()
	year := strconv.Itoa(x)
	next := strconv.Itoa(x + 1)

	tests := []struct {
		name                string
		y, m, d             string
		wantY, wantM, wantD bool
	}{
		{"all empty", "", "", "", false, false, false},
		{"all zero", "0", "0", "0", true, true, true},
		{"year only", year, "", "", true, false, false},
		{"year zero rest", year, "0", "0", true, true, true},
		{"pre-epoch year", "1979", "0", "0", false, true, true},
		{"future year", next, "", "", false, false, false},
		{"negative month", year, "-10", "", true, false, false},
		{"year and month", year, "1", "", true, true, false},
		{"month without year", "", "1", "", false, true, false},
		{"invalid month", year, "30", "", true, false, false},
		{"day without month", year, "", "1", true, false, true},
		{"month and day without year", "", "1", "1", false, true, true},
		{"all out of range", next, "13", "32", false, false, false},
		{"non-numeric", "abc", "efg", "hij", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			y, m, d := form.ValidDate(tt.y, tt.m, tt.d)
			be.Equal(t, y, tt.wantY)
			be.Equal(t, m, tt.wantM)
			be.Equal(t, d, tt.wantD)
		})
	}
}

func TestValidYouTube(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"empty", "", true},
		{"valid 11 chars", "dQw4w9WgXcQ", true},
		{"too short", "short", false},
		{"too long", "123456789012", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			be.Equal(t, form.ValidYouTube(tt.input), tt.want)
		})
	}
}

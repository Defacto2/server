package download

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

func TestTextContentType(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{"utf-16le", []byte{0xff, 0xfe, 'A', 0x00}, "text/plain; charset=utf-16le"},
		{"utf-16be", []byte{0xfe, 0xff, 0x00, 'A'}, "text/plain; charset=utf-16be"},
		{"latin1", []byte{0x8a, 'e', '\n'}, "text/plain; charset=iso-8859-1"},
		{"utf-8", []byte("hello world\n"), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := filepath.Join(tmp, tt.name+".txt")
			err := os.WriteFile(p, tt.data, 0o600)
			be.Err(t, err, nil)

			got, err := textContentType(p)
			be.Err(t, err, nil)
			be.Equal(t, got, tt.want)
		})
	}

	t.Run("empty file", func(t *testing.T) {
		t.Parallel()

		p := filepath.Join(tmp, "empty.txt")
		be.Err(t, os.WriteFile(p, nil, 0o600), nil)
		got, err := textContentType(p)
		be.Err(t, err)
		be.Equal(t, got, "")
	})

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()

		got, err := textContentType(filepath.Join(tmp, "does-not-exist.txt"))
		be.Err(t, err, os.ErrNotExist)
		be.Equal(t, got, "")
	})
}

func TestInline(t *testing.T) {
	t.Parallel()

	t.Run("nil context", func(t *testing.T) {
		t.Parallel()

		be.Err(t, inline(nil, true, "", "", ""))
	})

	tmp := t.TempDir()
	newCtx := func() (*echo.Context, *httptest.ResponseRecorder) {
		e := echo.New()
		e.Filesystem = os.DirFS("/")
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		return e.NewContext(req, rec), rec
	}

	tests := []struct {
		name     string
		file     string
		data     []byte
		text     bool
		ext      string
		wantCT   string
		wantCode int
	}{
		{"utf-16le text", "u16.txt", []byte{0xff, 0xfe, 'H', 0x00}, true, ".txt", "text/plain; charset=utf-16le", http.StatusOK},
		{"latin1 text", "lat.nfo", []byte{0x8a, 'e', '\n'}, true, ".nfo", "text/plain; charset=iso-8859-1", http.StatusOK},
		{"utf-8 text", "u8.txt", []byte("hello\n"), true, ".txt", "text/plain; charset=utf-8", http.StatusOK},
		{"image overrides text", "pic.png", []byte{0x89, 'P', 'N', 'G'}, true, ".png", "image/png", http.StatusOK},
		{"uppercase image overrides text", "PIC.PNG", []byte{0x89, 'P', 'N', 'G'}, true, ".PNG", "image/png", http.StatusOK},
		{"media overrides text", "song.mp3", []byte("ID3\x03\x00"), true, ".mp3", "audio/mpeg", http.StatusOK},
		{"binary non-text", "app.bin", []byte{0x00, 0x01, 0x02}, false, ".bin", "application/octet-stream", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := filepath.Join(tmp, tt.file)
			be.Err(t, os.WriteFile(p, tt.data, 0o600), nil)

			c, rec := newCtx()
			err := inline(c, tt.text, p, tt.file, tt.ext)
			be.Err(t, err, nil)
			be.Equal(t, rec.Code, tt.wantCode)
			be.Equal(t, rec.Body.Bytes(), tt.data)
			be.True(t, strings.HasPrefix(rec.Header().Get(echo.HeaderContentDisposition), "inline"))
			be.Equal(t, rec.Header().Get(echo.HeaderContentType), tt.wantCT)
		})
	}

	t.Run("missing text file", func(t *testing.T) {
		t.Parallel()

		c, _ := newCtx()
		be.Err(t, inline(c, true, filepath.Join(tmp, "missing.txt"), "missing.txt", ".txt"), os.ErrNotExist)
	})

	t.Run("missing binary file", func(t *testing.T) {
		t.Parallel()

		c, _ := newCtx()
		be.Err(t, inline(c, false, filepath.Join(tmp, "missing.bin"), "missing.bin", ".bin"))
	})
}

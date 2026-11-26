package command_test

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/dir"
	"github.com/google/uuid"
	"github.com/nalgeon/be"
)

func renderBBS(t *testing.T, name, content string) (image.Image, string, string, string) {
	t.Helper()

	prevdir := t.TempDir()
	thumbdir := t.TempDir()
	dirs := command.Dirs{
		Preview:   dir.Directory(prevdir),
		Thumbnail: dir.Directory(thumbdir),
	}

	unid := uuid.New().String()
	src := filepath.Join(t.TempDir(), name)
	be.Err(t, os.WriteFile(src, []byte(content), 0o600), nil)
	be.Err(t, dirs.TextImager(t.Context(), testLog, src, unid, false), nil)

	f, err := os.Open(filepath.Join(prevdir, unid+".png"))
	be.Err(t, err, nil)
	t.Cleanup(func() { _ = f.Close() })

	img, err := png.Decode(f)
	be.Err(t, err, nil)
	return img, prevdir, thumbdir, unid
}

func TestPCBoard(t *testing.T) {
	t.Parallel()

	pcb := "@CLS@@X1E\xdb\xdb\xdb PCBOARD AD \xdb\xdb\xdb\r\n@X0FWhite on black\x00line\n@X4CRed@PAUSE@"
	img, prevdir, thumbdir, unid := renderBBS(t, "sample.pcb", pcb)

	be.Equal(t, img.Bounds().Dx(), 640)
	be.Equal(t, img.Bounds().Dy(), 48) // 3 visible lines * 16px

	// Row 0 starts with @X1E (Yellow fg = 255,255,85 on Blue bg) and CP437 0xDB solid block.
	r, g, b, _ := img.At(0, 0).RGBA()
	be.Equal(t, r>>8, uint32(0xff))
	be.Equal(t, g>>8, uint32(0xff))
	be.Equal(t, b>>8, uint32(0x55))

	// Row 2 starts with @X4C (CGA Red bg = 170,0,0) on top pixel row above 'R'.
	r, g, b, _ = img.At(0, 32).RGBA()
	be.Equal(t, r>>8, uint32(0xaa))
	be.Equal(t, g>>8, uint32(0x00))
	be.Equal(t, b>>8, uint32(0x00))

	wst, err := os.Stat(filepath.Join(prevdir, unid+".webp"))
	be.Err(t, err, nil)
	be.True(t, wst.Size() > 0)

	tst, err := os.Stat(filepath.Join(thumbdir, unid+".webp"))
	be.Err(t, err, nil)
	be.True(t, tst.Size() > 0)
}

func TestPCBoardTall(t *testing.T) {
	t.Parallel()

	var sb strings.Builder
	sb.WriteString("@X1E")
	sb.WriteString(strings.Repeat("\xdb", 80))
	sb.WriteByte('\n')
	sb.WriteString("@X0F")
	line80 := strings.Repeat("X", 80)
	for range 59 {
		sb.WriteString(line80)
		sb.WriteByte('\n')
	}

	img, _, _, _ := renderBBS(t, "tall.pcb", sb.String())
	be.Equal(t, img.Bounds().Dx(), 640)
	be.Equal(t, img.Bounds().Dy(), 50*16) // capped at 50 rows * 16px

	// Verify row 0 (yellow solid block) was not scrolled off by 80-column wrapping on row 50.
	r, g, b, _ := img.At(0, 0).RGBA()
	be.Equal(t, r>>8, uint32(0xff))
	be.Equal(t, g>>8, uint32(0xff))
	be.Equal(t, b>>8, uint32(0x55))
}

func TestPCBoardLarge(t *testing.T) {
	t.Parallel()

	// Build a 48-row PCBoard file > 4,096 bytes (with lots of @X color codes per line)
	var sb strings.Builder
	for range 48 {
		for range 20 {
			sb.WriteString("@X1E\xdb@X4C\xdb@X0A\xdb@X0F\xdb")
		}
		sb.WriteString("\r\n")
	}
	be.True(t, sb.Len() > 4096)

	img, _, _, _ := renderBBS(t, "large.pcb", sb.String())
	be.Equal(t, img.Bounds().Dx(), 640)
	be.Equal(t, img.Bounds().Dy(), 48*16)
}

func TestBBSFormats(t *testing.T) {
	t.Parallel()

	t.Run("utf8", func(t *testing.T) {
		t.Parallel()
		// Valid UTF-8 block chars + trailing DOS EOF (0x1A) exercise prepareBBSBytes transcoding.
		img, _, _, _ := renderBBS(t, "utf8.pcb", "@X1E███ UTF8\x1a")
		be.Equal(t, img.Bounds().Dx(), 640)
		be.Equal(t, img.Bounds().Dy(), 16)

		r, g, b, _ := img.At(0, 0).RGBA()
		be.Equal(t, r>>8, uint32(0xff))
		be.Equal(t, g>>8, uint32(0xff))
		be.Equal(t, b>>8, uint32(0x55))
	})

	t.Run("renegade", func(t *testing.T) {
		t.Parallel()
		img, _, _, _ := renderBBS(t, "pipe.txt", "|14\xdb\xdb\xdb Renegade|07")
		be.Equal(t, img.Bounds().Dx(), 640)
		be.Equal(t, img.Bounds().Dy(), 16)

		r, g, b, _ := img.At(0, 0).RGBA()
		be.Equal(t, r>>8, uint32(0xff))
		be.Equal(t, g>>8, uint32(0xff))
		be.Equal(t, b>>8, uint32(0x55))
	})
}

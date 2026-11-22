package command_test

import (
	"image/png"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/dir"
	"github.com/google/uuid"
	"github.com/nalgeon/be"
)

func TestTextImagerPCBoard(t *testing.T) {
	t.Parallel()

	prevdir := t.TempDir()
	thumbdir := t.TempDir()
	dirs := command.Dirs{
		Preview:   dir.Directory(prevdir),
		Thumbnail: dir.Directory(thumbdir),
	}

	sl := slog.Default()
	unid := uuid.New().String()
	src := filepath.Join(t.TempDir(), "sample.pcb")
	pcb := "@CLS@@X1E\xdb\xdb\xdb PCBOARD AD \xdb\xdb\xdb\r\n@X0FWhite on black\x00line\n@X4CRed@PAUSE@"
	be.Err(t, os.WriteFile(src, []byte(pcb), 0o600), nil)

	got := dirs.TextImager(t.Context(), sl, src, unid, false)
	be.Err(t, got, nil)

	pngPath := filepath.Join(prevdir, unid+".png")
	f, err := os.Open(pngPath)
	be.Err(t, err, nil)
	t.Cleanup(func() { _ = f.Close() })

	img, err := png.Decode(f)
	be.Err(t, err, nil)
	be.Equal(t, img.Bounds().Dx(), 640)
	be.Equal(t, img.Bounds().Dy(), 48) // 3 visible lines * 16px

	// Row 0 starts with @X1E (Yellow fg = 255,255,85 on Blue bg) and CP437 0xDB solid block.
	r, g, b, _ := img.At(0, 0).RGBA()
	be.Equal(t, uint8(r>>8), uint8(0xff))
	be.Equal(t, uint8(g>>8), uint8(0xff))
	be.Equal(t, uint8(b>>8), uint8(0x55))

	// Row 2 starts with @X4C (CGA Red bg = 170,0,0) on top pixel row above 'R'.
	r, g, b, _ = img.At(0, 32).RGBA()
	be.Equal(t, uint8(r>>8), uint8(0xaa))
	be.Equal(t, uint8(g>>8), uint8(0x00))
	be.Equal(t, uint8(b>>8), uint8(0x00))

	webpPath := filepath.Join(prevdir, unid+".webp")
	wst, err := os.Stat(webpPath)
	be.Err(t, err, nil)
	be.True(t, wst.Size() > 0)

	thumbPath := filepath.Join(thumbdir, unid+".webp")
	tst, err := os.Stat(thumbPath)
	be.Err(t, err, nil)
	be.True(t, tst.Size() > 0)
}

func TestTextImagerPCBoardTallCropAndFullWidth(t *testing.T) {
	t.Parallel()

	prevdir := t.TempDir()
	thumbdir := t.TempDir()
	dirs := command.Dirs{
		Preview:   dir.Directory(prevdir),
		Thumbnail: dir.Directory(thumbdir),
	}

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

	unid := uuid.New().String()
	src := filepath.Join(t.TempDir(), "tall.pcb")
	be.Err(t, os.WriteFile(src, []byte(sb.String()), 0o600), nil)

	got := dirs.TextImager(t.Context(), slog.Default(), src, unid, false)
	be.Err(t, got, nil)

	f, err := os.Open(filepath.Join(prevdir, unid+".png"))
	be.Err(t, err, nil)
	t.Cleanup(func() { _ = f.Close() })

	img, err := png.Decode(f)
	be.Err(t, err, nil)
	be.Equal(t, img.Bounds().Dx(), 640)
	be.Equal(t, img.Bounds().Dy(), 50*16) // capped at 50 rows * 16px

	// Verify row 0 (yellow solid block) was not scrolled off by 80-column wrapping on row 50.
	r, g, b, _ := img.At(0, 0).RGBA()
	be.Equal(t, uint8(r>>8), uint8(0xff))
	be.Equal(t, uint8(g>>8), uint8(0xff))
	be.Equal(t, uint8(b>>8), uint8(0x55))
}

func TestTextImagerPCBoardLargeOver4KB(t *testing.T) {
	t.Parallel()

	prevdir := t.TempDir()
	thumbdir := t.TempDir()
	dirs := command.Dirs{
		Preview:   dir.Directory(prevdir),
		Thumbnail: dir.Directory(thumbdir),
	}

	// Build a 48-row PCBoard file > 4,096 bytes (with lots of @X color codes per line)
	var sb strings.Builder
	for range 48 {
		for range 20 {
			sb.WriteString("@X1E\xdb@X4C\xdb@X0A\xdb@X0F\xdb")
		}
		sb.WriteString("\r\n")
	}

	unid := uuid.New().String()
	src := filepath.Join(t.TempDir(), "large.pcb")
	be.Err(t, os.WriteFile(src, []byte(sb.String()), 0o600), nil)

	got := dirs.TextImager(t.Context(), slog.Default(), src, unid, false)
	be.Err(t, got, nil)

	f, err := os.Open(filepath.Join(prevdir, unid+".png"))
	be.Err(t, err, nil)
	t.Cleanup(func() { _ = f.Close() })

	img, err := png.Decode(f)
	be.Err(t, err, nil)
	be.Equal(t, img.Bounds().Dx(), 640)
	be.Equal(t, img.Bounds().Dy(), 48*16)
}

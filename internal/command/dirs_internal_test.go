package command

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	imgpng "image/png"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

var testLog = logs.Discard()

func TestTheDraw(t *testing.T) {
	t.Parallel()

	t.Run("missing", func(t *testing.T) {
		t.Parallel()

		missing := filepath.Join(t.TempDir(), "missing.com")
		outPath, cleanup, ok, err := thedrawCOM(nil, missing)
		be.Err(t, err, nil)
		be.Equal(t, outPath, "")
		be.True(t, !ok)
		be.True(t, cleanup == nil)
	})

	t.Run("plain", func(t *testing.T) {
		t.Parallel()

		plain := filepath.Join(t.TempDir(), "plain.ans")
		be.Err(t, os.WriteFile(plain, []byte("\x1b[32mhi\x1b[0m\r\n"), 0o600), nil)

		outPath, cleanup, ok, err := thedrawCOM(nil, plain)
		be.Err(t, err, nil)
		be.Equal(t, outPath, "")
		be.True(t, !ok)
		be.True(t, cleanup == nil)
	})

	t.Run("com", func(t *testing.T) {
		t.Parallel()

		const stubLen = 0xF0
		payload := [...]byte{0x0E, 'Y', 'O'}
		td := make([]byte, stubLen+len(payload))
		copy(td[0:3], []byte{0xEB, 0x3D, 0x90})
		binary.LittleEndian.PutUint16(td[4:6], uint16(len(payload)))
		copy(td[9:], []byte("TheDraw COM file Screen Save\x1A"))
		copy(td[stubLen:], payload[:])

		tdPath := filepath.Join(t.TempDir(), "TD.COM")
		be.Err(t, os.WriteFile(tdPath, td, 0o600), nil)

		var logBuf bytes.Buffer
		sl := slog.New(slog.NewTextHandler(&logBuf, nil))

		outPath, cleanup, ok, err := thedrawCOM(sl, tdPath)
		be.Err(t, err, nil)
		be.True(t, ok)
		if cleanup == nil {
			t.Fatal("expected non-nil cleanup")
		}
		t.Cleanup(cleanup)

		be.True(t, strings.HasSuffix(outPath, ".ans"))
		decoded, err := os.ReadFile(outPath)
		be.Err(t, err, nil)
		be.True(t, strings.Contains(string(decoded), "YO"))

		cleanup()
		_, err = os.Stat(outPath)
		be.True(t, errors.Is(err, os.ErrNotExist))
		be.Equal(t, logBuf.String(), "")
	})
}

func TestBinText(t *testing.T) {
	t.Parallel()

	const (
		dumpSize    = 4096
		screenBytes = 4000
		cellBytes   = 2
		attrDefault = 0x07
		attrCyan    = 0x0B
		bsaveHeader = 7
		bsaveSeg    = 0xB800
		expectedArg = 7
	)

	makeRawDump := func() []byte {
		raw := make([]byte, dumpSize)
		for i := 0; i+1 < len(raw); i += cellBytes {
			raw[i] = ' '
			raw[i+1] = attrDefault
		}
		raw[0], raw[1] = 'A', attrCyan
		return raw
	}

	t.Run("missing", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		missing := filepath.Join(tmpDir, "missing.bin")
		dst := filepath.Join(tmpDir, "out.png")

		args, cleanup, err := binTextArgs(nil, missing, dst)
		be.Err(t, err, nil)
		be.True(t, cleanup == nil)
		be.True(t, slices.Equal(args, []string{missing, "-o", dst}))
	})

	t.Run("plain", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		dst := filepath.Join(tmpDir, "out.png")
		plain := filepath.Join(tmpDir, "plain.bin")
		be.Err(t, os.WriteFile(plain, []byte("hello"), 0o600), nil)

		args, cleanup, err := binTextArgs(nil, plain, dst)
		be.Err(t, err, nil)
		be.True(t, cleanup == nil)
		be.True(t, slices.Equal(args, []string{plain, "-o", dst}))
	})

	t.Run("raw cga", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		dst := filepath.Join(tmpDir, "out.png")
		tpg := filepath.Join(tmpDir, "SPYHUNT.TPG")
		be.Err(t, os.WriteFile(tpg, makeRawDump(), 0o600), nil)

		args, cleanup, err := binTextArgs(nil, tpg, dst)
		be.Err(t, err, nil)
		be.True(t, cleanup == nil)
		be.True(t, slices.Equal(args, []string{"-t", "bin", "-c", "80", tpg, "-o", dst}))
	})

	t.Run("bsave", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		dst := filepath.Join(tmpDir, "out.png")
		raw := makeRawDump()
		bsave := make([]byte, bsaveHeader+screenBytes)
		bsave[0] = 0xFD
		binary.LittleEndian.PutUint16(bsave[1:3], bsaveSeg)
		binary.LittleEndian.PutUint16(bsave[5:7], screenBytes)
		copy(bsave[bsaveHeader:], raw[:screenBytes])

		bsavePath := filepath.Join(tmpDir, "BSAVE.BIN")
		be.Err(t, os.WriteFile(bsavePath, bsave, 0o600), nil)

		var logBuf bytes.Buffer
		sl := slog.New(slog.NewTextHandler(&logBuf, nil))

		args, cleanup, err := binTextArgs(sl, bsavePath, dst)
		be.Err(t, err, nil)
		if cleanup == nil {
			t.Fatal("expected non-nil cleanup")
		}
		t.Cleanup(cleanup)

		if len(args) != expectedArg {
			t.Fatalf("got %d args (%v), want %d", len(args), args, expectedArg)
		}
		be.True(t, slices.Equal(args[:4], []string{"-t", "bin", "-c", "80"}))
		be.True(t, args[4] != bsavePath)
		be.True(t, slices.Equal(args[5:], []string{"-o", dst}))

		stripped, err := os.ReadFile(args[4])
		be.Err(t, err, nil)
		be.True(t, bytes.Equal(stripped, raw[:screenBytes]))

		cleanup()
		_, err = os.Stat(args[4])
		be.True(t, errors.Is(err, os.ErrNotExist))
		be.Equal(t, logBuf.String(), "")
	})
}

func TestBSAVE(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data []byte
	}{
		{"zero length", []byte{0xFD, 0x00, 0xB8, 0x00, 0x00, 0x00, 0x00, 'X'}},
		{"truncated", []byte{0xFD, 0x00, 0xB8, 0x00, 0x00, 0xFF, 0xFF, 'X'}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			src := filepath.Join(t.TempDir(), "bad.bin")
			be.Err(t, os.WriteFile(src, tt.data, 0o600), nil)

			got, cleanup, err := stripBSAVEHeader(testLog, src)
			be.Err(t, err, nil)
			be.Equal(t, got, src)
			be.True(t, cleanup == nil)
		})
	}
}

func TestFindName(t *testing.T) {
	t.Parallel()

	sl := testLog

	t.Run("nil logger", func(t *testing.T) {
		t.Parallel()
		be.Equal(t, findName(nil, t.TempDir(), "FILE_ID.DIZ"), "")
	})

	t.Run("missing dir", func(t *testing.T) {
		t.Parallel()
		missing := filepath.Join(t.TempDir(), "no-such-dir")
		be.Equal(t, findName(sl, missing, "FILE_ID.DIZ"), "")
	})

	t.Run("exact", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		lower := filepath.Join(root, "file_id.diz")
		exact := filepath.Join(root, "FILE_ID.DIZ")
		be.Err(t, os.WriteFile(lower, []byte("lower"), 0o600), nil)
		be.Err(t, os.WriteFile(exact, []byte("exact"), 0o600), nil)

		be.Equal(t, findName(sl, root, "FILE_ID.DIZ"), exact)
	})

	t.Run("fold", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		lower := filepath.Join(root, "file_id.diz")
		be.Err(t, os.WriteFile(lower, []byte("lower"), 0o600), nil)

		be.Equal(t, findName(sl, root, "FILE_ID.DIZ"), lower)
	})

	t.Run("skip dir", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		be.Err(t, os.Mkdir(filepath.Join(root, "FILE_ID.DIZ"), 0o750), nil)

		be.Equal(t, findName(sl, root, "FILE_ID.DIZ"), "")
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		be.Err(t, os.WriteFile(filepath.Join(root, "other.txt"), []byte("x"), 0o600), nil)

		be.Equal(t, findName(sl, root, "FILE_ID.DIZ"), "")
	})
}

func TestDuplicateRetry(t *testing.T) {
	t.Parallel()

	sl := testLog

	t.Run("nil logger", func(t *testing.T) {
		t.Parallel()

		err := duplicateSecondTry(nil, os.ErrNotExist, "src", "dst", "diz")
		be.True(t, errors.Is(err, nils.ErrSlogLogger))
	})

	t.Run("other error", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		dst := filepath.Join(tmpDir, "out.diz")
		err := duplicateSecondTry(sl, os.ErrPermission, filepath.Join(tmpDir, "FILE_ID.DIZ"), dst, "diz")
		be.True(t, errors.Is(err, os.ErrPermission))
		_, statErr := os.Stat(dst)
		be.True(t, errors.Is(statErr, os.ErrNotExist))
	})

	t.Run("fallback copy", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		actual := filepath.Join(tmpDir, "file_id.diz")
		want := []byte("hello diz\n")
		be.Err(t, os.WriteFile(actual, want, 0o600), nil)

		missingCase := filepath.Join(tmpDir, "FILE_ID.DIZ")
		dst := filepath.Join(tmpDir, "out.diz")

		err := duplicateSecondTry(sl, os.ErrNotExist, missingCase, dst, "diz")
		be.Err(t, err, nil)

		got, err := os.ReadFile(dst)
		be.Err(t, err, nil)
		be.True(t, bytes.Equal(got, want))
	})

	t.Run("missing", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		src := filepath.Join(tmpDir, "FILE_ID.DIZ")
		dst := filepath.Join(tmpDir, "out.diz")

		err := duplicateSecondTry(sl, os.ErrNotExist, src, dst, "diz")
		be.True(t, errors.Is(err, os.ErrNotExist))
	})
}

func TestIsAnim(t *testing.T) {
	t.Parallel()

	t.Run("missing", func(t *testing.T) {
		t.Parallel()

		ok, err := isANSImation(filepath.Join(t.TempDir(), "missing.ans"))
		be.Err(t, err)
		be.True(t, !ok)
	})

	t.Run("plain", func(t *testing.T) {
		t.Parallel()

		src := filepath.Join(t.TempDir(), "plain.txt")
		be.Err(t, os.WriteFile(src, []byte("plain text\r\n"), 0o600), nil)

		ok, err := isANSImation(src)
		be.Err(t, err, nil)
		be.True(t, !ok)
	})

	t.Run("animated", func(t *testing.T) {
		t.Parallel()

		ok, err := isANSImation(testutil.FileData("maximum3ans").Abs())
		be.Err(t, err, nil)
		be.True(t, ok)
	})

	t.Run("blink", func(t *testing.T) {
		t.Parallel()

		src := filepath.Join(t.TempDir(), "blink.ans")
		be.Err(t, os.WriteFile(src, []byte("\x1b[5;31mHello World\x1b[0m\r\n"), 0o600), nil)

		ok, err := isANSImation(src)
		be.Err(t, err, nil)
		be.True(t, ok)
	})
}

func TestFramePNG(t *testing.T) {
	t.Parallel()

	frame := image.NewRGBA(image.Rect(0, 0, 4, 4))
	want := color.RGBA{R: 0xCC, G: 0x33, B: 0x11, A: 0xFF}
	frame.SetRGBA(1, 2, want)

	tmp, err := writeFramePNG(testLog, frame)
	be.Err(t, err, nil)
	t.Cleanup(func() { _ = os.Remove(tmp) })
	be.True(t, strings.HasSuffix(tmp, ".png"))

	f, err := os.Open(tmp)
	be.Err(t, err, nil)
	defer f.Close()

	decoded, err := imgpng.Decode(f)
	be.Err(t, err, nil)
	be.Equal(t, decoded.Bounds(), image.Rect(0, 0, 4, 4))
	be.Equal(t, color.RGBAModel.Convert(decoded.At(1, 2)), color.Color(want))
}

//nolint:gosec
package fixlha_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/binary"
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Defacto2/server/internal/command"
	"github.com/Defacto2/server/internal/config/fixlha"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/testutil"
	"github.com/nalgeon/be"
)

var testLog = logs.Discard()

func requireLha(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath(command.Lha); err != nil {
		t.Skipf("%s not in PATH: %v", command.Lha, err)
	}
}

func mustPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	fn()
}

// makeLHA builds a minimal valid level-0 stored (-lh0-) LHA archive.
func makeLHA(name string, payload []byte) []byte {
	var crc uint16
	for _, b := range payload {
		crc ^= uint16(b)
		for range 8 {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}

	var body bytes.Buffer
	body.WriteString("-lh0-")
	size := uint32(len(payload)) //nolint:gosec // small test payload
	_ = binary.Write(&body, binary.LittleEndian, size)
	_ = binary.Write(&body, binary.LittleEndian, size)
	_ = binary.Write(&body, binary.LittleEndian, uint32(0)) // MS-DOS mtime
	body.WriteByte(0x20)                                    // MS-DOS attr
	body.WriteByte(0x00)                                    // header level 0
	body.WriteByte(byte(len(name)))
	body.WriteString(name)
	_ = binary.Write(&body, binary.LittleEndian, crc)

	hdr := body.Bytes()
	var sum byte
	for _, b := range hdr {
		sum += b
	}

	var out bytes.Buffer
	out.WriteByte(byte(len(hdr)))
	out.WriteByte(sum)
	out.Write(hdr)
	out.Write(payload)
	out.WriteByte(0x00) // end-of-archive marker
	return out.Bytes()
}

func TestFiles(t *testing.T) {
	t.Parallel()

	t.Run("nil ctx", func(t *testing.T) {
		t.Parallel()
		_, err := fixlha.Files(nil, new(sql.DB)) //nolint:staticcheck // testing nil context
		be.True(t, errors.Is(err, nils.ErrArgument))
	})

	t.Run("nil exec", func(t *testing.T) {
		t.Parallel()
		_, err := fixlha.Files(t.Context(), nil)
		be.True(t, errors.Is(err, nils.ErrArgument))
	})

	t.Run("db", func(t *testing.T) {
		t.Parallel()
		db := testutil.DB(t)
		fs, err := fixlha.Files(t.Context(), db)
		be.Err(t, err, nil)
		be.True(t, len(fs) > 0)
	})
}

func TestInvalid(t *testing.T) {
	t.Parallel()

	t.Run("nil ctx", func(t *testing.T) {
		t.Parallel()
		mustPanic(t, func() {
			fixlha.Invalid(nil, testLog, "file.lha") //nolint:staticcheck // testing nil context
		})
	})

	t.Run("nil logger", func(t *testing.T) {
		t.Parallel()
		mustPanic(t, func() {
			fixlha.Invalid(t.Context(), nil, "file.lha")
		})
	})

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		requireLha(t)

		lhaPath := filepath.Join(t.TempDir(), "valid.lha")
		be.Err(t, os.WriteFile(lhaPath, makeLHA("hi.txt", []byte("hello lha\n")), 0o600), nil)

		be.True(t, !fixlha.Invalid(t.Context(), testLog, lhaPath))
	})

	t.Run("not archive", func(t *testing.T) {
		t.Parallel()
		requireLha(t)

		txt := filepath.Join(t.TempDir(), "plain.txt")
		be.Err(t, os.WriteFile(txt, []byte("not an archive\n"), 0o600), nil)

		be.True(t, fixlha.Invalid(t.Context(), testLog, txt))
	})

	t.Run("missing", func(t *testing.T) {
		t.Parallel()
		requireLha(t)

		var logBuf bytes.Buffer
		sl := slog.New(slog.NewTextHandler(&logBuf, nil))

		missing := filepath.Join(t.TempDir(), "missing.lha")
		be.True(t, fixlha.Invalid(t.Context(), sl, missing))
		be.True(t, logBuf.Len() > 0)
	})

	t.Run("canceled", func(t *testing.T) {
		t.Parallel()
		requireLha(t)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		missing := filepath.Join(t.TempDir(), "canceled.lha")
		be.True(t, fixlha.Invalid(ctx, testLog, missing))
	})
}

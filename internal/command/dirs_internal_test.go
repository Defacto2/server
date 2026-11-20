package command

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/nalgeon/be"
)

func TestTheDrawTempANS(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	plain := filepath.Join(tmpDir, "plain.ans")
	be.Err(t, os.WriteFile(plain, []byte("\x1b[32mhi\x1b[0m\r\n"), 0o600), nil)

	_, cleanup, ok, err := thedrawCOM(nil, plain)
	be.Err(t, err, nil)
	be.True(t, !ok)
	be.True(t, cleanup == nil)

	const stubLen = 0xF0
	payload := []byte{0x0E, 'Y', 'O'}
	td := make([]byte, stubLen+len(payload))
	copy(td[0:3], []byte{0xEB, 0x3D, 0x90})
	binary.LittleEndian.PutUint16(td[4:6], uint16(len(payload)))
	copy(td[9:], []byte("TheDraw COM file Screen Save\x1A"))
	copy(td[stubLen:], payload)

	tdPath := filepath.Join(tmpDir, "TD.COM")
	be.Err(t, os.WriteFile(tdPath, td, 0o600), nil)

	outPath, cleanup, ok, err := thedrawCOM(nil, tdPath)
	be.Err(t, err, nil)
	be.True(t, ok)
	be.True(t, cleanup != nil)

	decoded, err := os.ReadFile(outPath)
	be.Err(t, err, nil)
	be.True(t, len(decoded) > 0)
	cleanup()
	_, err = os.Stat(outPath)
	be.True(t, os.IsNotExist(err))
}

func TestBinTextArgs(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	dst := filepath.Join(tmpDir, "out.png")

	// Plain binary text (not CGAVideoDump) keeps default ansilove args.
	plain := filepath.Join(tmpDir, "plain.bin")
	err := os.WriteFile(plain, []byte("hello"), 0o600)
	be.Err(t, err, nil)

	args, cleanup, err := binTextArgs(nil, plain, dst)
	be.Err(t, err, nil)
	be.True(t, cleanup == nil)
	be.True(t, slices.Equal(args, []string{plain, "-o", dst}))

	// Raw 4096-byte CGA video RAM dump (e.g. .tpg) uses 80-col binary mode.
	raw := make([]byte, 4096)
	for i := 0; i < len(raw); i += 2 {
		raw[i] = ' '
		raw[i+1] = 0x07
	}
	raw[0], raw[1] = 'A', 0x0B
	tpg := filepath.Join(tmpDir, "SPYHUNT.TPG")
	err = os.WriteFile(tpg, raw, 0o600)
	be.Err(t, err, nil)

	args, cleanup, err = binTextArgs(nil, tpg, dst)
	be.Err(t, err, nil)
	be.True(t, cleanup == nil)
	be.True(t, slices.Equal(args, []string{"-t", "bin", "-c", "80", tpg, "-o", dst}))

	// BSAVE-wrapped CGA video RAM dump strips the 7-byte header into a temp file.
	bsave := make([]byte, 7+4000)
	bsave[0] = 0xFD
	binary.LittleEndian.PutUint16(bsave[1:3], 0xB800)
	binary.LittleEndian.PutUint16(bsave[5:7], 4000)
	copy(bsave[7:], raw[:4000])
	bsavePath := filepath.Join(tmpDir, "BSAVE.BIN")
	err = os.WriteFile(bsavePath, bsave, 0o600)
	be.Err(t, err, nil)

	args, cleanup, err = binTextArgs(nil, bsavePath, dst)
	be.Err(t, err, nil)
	be.True(t, cleanup != nil)
	be.Equal(t, len(args), 7)
	be.True(t, slices.Equal(args[:4], []string{"-t", "bin", "-c", "80"}))
	be.True(t, args[4] != bsavePath)
	be.True(t, slices.Equal(args[5:], []string{"-o", dst}))

	stripped, err := os.ReadFile(args[4])
	be.Err(t, err, nil)
	be.Equal(t, len(stripped), 4000)
	cleanup()
	_, err = os.Stat(args[4])
	be.True(t, os.IsNotExist(err))
}

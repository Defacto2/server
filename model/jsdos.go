package model

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Defacto2/server/handler/jsdos"
	"github.com/Defacto2/server/handler/jsdos/msdos"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/subpop/go-ini"
)

// JsDosCommand returns the program executable or commands to run in the js-dos emulator.
// If the dosee_run_program is set then it is the preferred executable.
// If the filename is a .com or .exe then it will return the filename.
// Otherwise, it will attempt to find the most likely executable in the archive.
func JsDosCommand(f *models.File) (string, error) {
	const format = "jsdos command: %w"
	if f == nil {
		return "", fmt.Errorf(format, ErrModel)
	}

	if f.DoseeRunProgram.Valid && f.DoseeRunProgram.String != "" {
		return f.DoseeRunProgram.String, nil
	}

	return JsDosBinary(f)
}

// JsDosBinary returns the program executable to run in the js-dos emulator.
// If the filename is a .com or .exe then it will return the filename.
// Otherwise, it will attempt to find the most likely executable in the archive.
func JsDosBinary(f *models.File) (string, error) {
	const format = "jsdos binary: %w"
	if f == nil {
		return "", fmt.Errorf(format, ErrModel)
	}

	if !f.Filename.Valid || f.Filename.IsZero() || f.Filename.String == "" {
		return "", nil
	}

	path := strings.ToLower(f.Filename.String)
	switch filepath.Ext(path) {
	case ".com", ".exe", ".bat":
		// okay
	default:
		if !f.FileZipContent.Valid || f.FileZipContent.IsZero() || f.FileZipContent.String == "" {
			return "", nil
		}
	}

	const separatorDOS = "\\"
	const separatorUnix = "/"
	findname := jsdos.FindBinary(f.Filename.String, f.FileZipContent.String)
	if !strings.Contains(findname, separatorDOS) &&
		!strings.Contains(findname, separatorUnix) {
		return msdos.Truncate(findname), nil
	}

	// Replace all path separators with DOS path separators.
	// Often the FileZipContent paths use non-DOS path separators
	// despite the zipfile being a DOS file.
	dir := filepath.Dir(findname)
	dir = strings.ReplaceAll(dir, separatorUnix, separatorDOS)
	base := msdos.Truncate(filepath.Base(findname))

	return strings.Join([]string{dir, base}, separatorDOS), nil
}

// JsDosConfig creates a js-dos .ini configuration for the emulator.
func JsDosConfig(f *models.File) ([]byte, error) {
	const format = "jsdos config: %w"
	if f == nil {
		return []byte{}, fmt.Errorf(format, ErrModel)
	}

	emulation := jsdos.Jsdos{} //nolint:exhaustruct
	cpu := f.DoseeHardwareCPU.String
	if f.DoseeHardwareCPU.Valid && cpu != "" {
		emulation.CPU(cpu)
	}

	gfx := f.DoseeHardwareGraphic.String
	if f.DoseeHardwareGraphic.Valid && gfx != "" {
		emulation.Machine(gfx)
	}

	sfx := f.DoseeHardwareAudio.String
	if f.DoseeHardwareAudio.Valid && sfx != "" {
		emulation.Sound(sfx)
	}

	mem := f.DoseeNoEms.Int16
	if f.DoseeNoEms.Valid && mem == 1 {
		emulation.NoEMS(true)
	}

	mem = f.DoseeNoXMS.Int16
	if f.DoseeNoXMS.Valid && mem == 1 {
		emulation.NoXMS(true)
	}

	mem = f.DoseeNoUmb.Int16
	if f.DoseeNoUmb.Valid && mem == 1 {
		emulation.NoUMB(true)
	}

	p, err := ini.Marshal(emulation)
	if err != nil {
		return []byte{}, fmt.Errorf(format, err)
	}

	return p, nil
}

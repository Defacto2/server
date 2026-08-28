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
	const msg = "jsdos command"
	if f == nil {
		return "", fmt.Errorf("%s: %w", msg, ErrModel)
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
	if f == nil {
		return "", ErrModel
	}
	if !f.Filename.Valid || f.Filename.IsZero() || f.Filename.String == "" {
		return "", nil
	}
	name := strings.ToLower(f.Filename.String)
	switch filepath.Ext(name) {
	case ".com", ".exe", ".bat":
		break
	default:
		if !f.FileZipContent.Valid || f.FileZipContent.IsZero() || f.FileZipContent.String == "" {
			return "", nil
		}
	}
	const dosPathSeparator, winPathSeparator = "\\", "/"
	findname := jsdos.FindBinary(f.Filename.String, f.FileZipContent.String)
	if !strings.Contains(findname, dosPathSeparator) && !strings.Contains(findname, winPathSeparator) {
		return msdos.Truncate(findname), nil
	}
	dir := filepath.Dir(findname)
	// replace all windows path separators with dos path separators,
	// as often the FileZipContent paths use non-dos path separators
	// despite the zipfile being a DOS file.
	dir = strings.ReplaceAll(dir, winPathSeparator, dosPathSeparator)
	base := msdos.Truncate(filepath.Base(findname))
	return strings.Join([]string{dir, base}, dosPathSeparator), nil
}

// JsDosConfig creates a js-dos .ini configuration for the emulator.
func JsDosConfig(f *models.File) (string, error) {
	const msg = "jsdos config"
	if f == nil {
		return "", fmt.Errorf("%s: %w", msg, ErrModel)
	}
	j := jsdos.Jsdos{} //nolint:exhaustruct // External library with many optional configuration fields
	cpu := f.DoseeHardwareCPU.String
	if f.DoseeHardwareCPU.Valid && cpu != "" {
		j.CPU(cpu)
	}
	hw := f.DoseeHardwareGraphic.String
	if f.DoseeHardwareGraphic.Valid && hw != "" {
		j.Machine(hw)
	}
	sfx := f.DoseeHardwareAudio.String
	if f.DoseeHardwareAudio.Valid && sfx != "" {
		j.Sound(sfx)
	}
	mem := f.DoseeNoEms.Int16
	if f.DoseeNoEms.Valid && mem == 1 {
		j.NoEMS(true)
	}
	mem = f.DoseeNoXMS.Int16
	if f.DoseeNoXMS.Valid && mem == 1 {
		j.NoXMS(true)
	}
	mem = f.DoseeNoUmb.Int16
	if f.DoseeNoUmb.Valid && mem == 1 {
		j.NoUMB(true)
	}
	b, err := ini.Marshal(j)
	if err != nil {
		return "", fmt.Errorf("%s ini marshal: %w", msg, err)
	}
	return string(b), nil
}

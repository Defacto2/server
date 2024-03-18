package model

// Package file helper.go contains helper functions for the model package.

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/postgres/models"
)

var ErrModel = errors.New("error, no file model")

type DosConf int

const (
	CGA      DosConf = iota // Force CGA (Color Graphics Adapter) graphics mode circa 1981.
	Covox                   // Enable Covox Speech Accelerator & and PC speaker & disable all other sound devices.
	EGA                     // Force EGA (Enhanced Graphics Adapter) graphics mode circa 1984.
	ET3000                  // Force Tseng Labs ET3000 graphics mode.
	ET4000                  // Force Tseng Labs ET4000 graphics mode.
	GUS                     // Enable Gravis Ultra Sound & disable all other sound devices.
	Hercules                // Force Hercules graphics mode circa 1982.
	i8086                   // This emulates the 8086 CPU in real mode.
	i386                    // This emulates the 80386 CPU in protected mode with additional checks that may 'slow' the emulation.
	iAuto                   // The recommended DOSBox CPU settings.
	iMax                    // Force emulation to run at maximum speed permitted by the browser.
	NoAudio                 // Disable all audio devices which may improve performance.
	NoEMS                   // Disable Expanded Memory Specification (EMS) which may conflict with some software.
	NoLFB                   // Force SuperVGA s3 Trio 64 chip with no-line frame buffer hack, it is sometimes faster than the default.
	NoUMB                   // Disable Upper Memory Blocks (UMB) which may conflict with some software.
	NoXMS                   // Disable Extended Memory Specification (XMS) which may conflict with some software.
	OldVBE                  // Force SuperVGA s3 Trio 64 chip with 64 MB of RAM (memory capacity) using VEGA 1.3 instead of VESA VBE 2.0.
	Paradise                // Force SuperVGA Paradise PVGA1A chip common in the late 1980s.
	SB1                     // Force Creative Labs Sound Blaster 1.0, AdLib & PC Speaker audio with driver support.
	SB16                    // Force Creative Labs Sound Blaster 16, AdLib Gold, AdLib & PC Speaker, this is the most software-compatible audio selection.
	SVGA                    // Force SuperVGA s3 Trio 64 chip.
	Tandy                   // Force Tandy 1000 emulation which uses the Tandy Graphics Adapter and the 3-channel Tandy speaker circa 1984.
	VGAOnly                 // Force VGA graphics mode for compatibility with software that fails with SuperVGA.
)

// todo struct machine, cpu, sound, memory

func DosConfig(cfg DosConf) []string {
	const beep, box, cpu, gus, midi, sb = "[speaker]", "[dosbox]", "[cpu]", "[gus]", "[midi]", "[sblaster]"
	mem640k := []string{box, "xms=false", "ems=false", "umb=false"}
	noSB := []string{sb, "sbtype=none", "oplmode=none"}
	noMidi := []string{midi, "mpu401=none", "mididevice=none"}
	switch cfg {
	case CGA:
		return []string{box, "machine=cga", "memsize=1"}
	case Covox:
		s := []string{beep, "pcspeaker=true", "pcrate=44100", "tandy=off", "disney=true"}
		s = append(s, noSB...)
		s = append(s, noMidi...)
		return s
	case EGA:
		return []string{box, "machine=ega", "memsize=4"}
	case ET3000:
		return []string{box, "machine=svga_et3000", "memsize=16"}
	case ET4000:
		return []string{box, "machine=svga_et4000", "memsize=16"}
	case GUS:
		s := []string{gus, "gus=true", "gusrate=44100", "gusbase=240", "gusirq=5", "gusdma=1", "ultradir=C:\\ULTRASND"}
		s = append(s, noSB...)
		s = append(s, noMidi...)
		s = append(s, beep, "pcspeaker=true", "pcrate=44100", "tandy=off", "disney=true")
		return s
	case Hercules:
		return []string{box, "machine=hercules", "memsize=1"}
	case i8086:
		s := []string{cpu, "core=normal", "cputype=8086", "cycles=fixed 500"}
		return append(s, mem640k...)
	case i386:
		s := []string{cpu, "core=normal", "cputype=386_slow"}
		return append(s, mem640k...)
	case iAuto:
		return []string{cpu, "core=auto", "cputype=auto", "cycles=auto"}
	case iMax:
		return []string{cpu, "core=dynamic", "cputype=auto", "cycles=max"} // TODO: dynamic core may not be supported, replace with auto
	case NoAudio:
		s := noMidi
		s = append(s, noSB...)
		s = append(s, gus, "gus=false")
		s = append(s, beep, "pcspeaker=false", "tandy=off", "disney=false")
		return s
	case NoEMS:
		return []string{box, "ems=false"}
	case NoLFB:
		return []string{box, "machine=vesa_nolfb", "memsize=16"}
	case NoUMB:
		return []string{box, "umb=false"}
	case NoXMS:
		return []string{box, "xms=false"}
	case OldVBE:
		return []string{box, "machine=vesa_oldvbe", "memsize=16"}
	case Paradise:
		return []string{box, "machine=svga_paradise"}
	case SB1:
		s := noMidi
		s = append(s, sb, "sbtype=sb1", "sbbase=220", "irq=7", "dma=1", "sbmixer=true", "oplmode=auto", "oplrate=44100", "oplemu=default")
		s = append(s, gus, "gus=false")
		s = append(s, beep, "pcspeaker=true", "pcrate=44100", "tandy=off", "disney=false")
		// todo join filepaths
		s = append(s, "[autoexec]", "PATH=S:\\DRIVERS\\CT-1320C\\;S:\\DRIVERS\\ADLIB\\;%PATH%", "SET SOUND=S:\\DRIVERS\\CT-1320C\\")
		return s
	case SB16:
		s := noMidi
		s = append(s, sb, "sbtype=sb16", "sbbase=220", "irq=7", "dma=1", "hdma=5", "sbmixer=true", "oplmode=auto", "oplrate=44100", "oplemu=default")
		s = append(s, gus, "gus=false")
		s = append(s, beep, "pcspeaker=true", "pcrate=44100", "tandy=off", "disney=false")
		return s
	case SVGA:
		return []string{box, "machine=svga_s3", "memsize=16"}
	case Tandy:
		return []string{box, "machine=tandy", "memsize=1", beep, "pcspeaker=true", "pcrate=44100", "tandy=auto", "tandyrate=44100"}
	case VGAOnly:
		return []string{box, "machine=vga", "memsize=16"}
	}
	return []string{}
}

func DosPaths(zipContent string) []string {
	if zipContent == "" {
		return []string{}
	}
	const delimiter = ":" // the colon is an illegal character as a DOS filename
	archive := zipContent
	archive = strings.ReplaceAll(archive, "\r\n", delimiter) // replace Microsoft-style CRLF with delimiter
	archive = strings.ReplaceAll(archive, "\n", delimiter)   // replace Unix LF with delimiter
	archive = strings.ReplaceAll(archive, "\r", delimiter)   // replace 8-bit microcomputer era CR with delimiter
	paths := strings.Split(archive, delimiter)
	// TODO convert into DOS 8.3 filename format?
	return paths
}

func DosBins(paths ...string) []string {
	if len(paths) == 0 {
		return []string{}
	}
	programs := []string{".bat", ".com", ".exe"}
	executables := []string{}
	for _, path := range paths {
		p := strings.ToLower(path)
		if slices.Contains(programs, filepath.Ext(p)) {
			executables = append(executables, path)
		}
	}
	return executables
}

func DosMatch(filename string, paths ...string) string {
	if filename == "" || len(paths) == 0 {
		return ""
	}
	// sort by the number of directories in the path, to prioritise binaries in the root of the archive
	sort.Slice(paths, func(i, j int) bool {
		return len(filepath.SplitList(paths[i])) < len(filepath.SplitList(paths[j]))
	})
	// prioritise the most likely executable that matches the archive name
	// e.g. if the archive is 'myapp.zip' then the most likely executable order are
	// 'myapp.exe', 'myapp.com', 'myapp.bat'
	root := paths
	sort.Slice(root, func(i, j int) bool {
		// only consider executables in the root of the archive
		return len(filepath.SplitList(root[i])) == 0
	})
	if len(root) == 0 {
		return ""
	}
	base := filepath.Base(filename)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	priority := []string{name + ".exe", name + ".com", name + ".bat"}
	for _, name := range priority {
		for _, path := range root {
			if strings.ToLower(path) == name {
				return path
			}
		}
	}
	return ""
}

func DosBin(paths ...string) string {
	if len(paths) == 0 {
		return ""
	}
	// sort by the number of directories in the path, to prioritise binaries in the root of the archive
	sort.Slice(paths, func(i, j int) bool {
		return len(filepath.SplitList(paths[i])) < len(filepath.SplitList(paths[j]))
	})
	// in the future we could limit the directory depth of the search

	for _, path := range paths {
		if strings.ToLower(filepath.Ext(path)) == ".bat" {
			return path
		}
	}
	for _, path := range paths {
		if strings.ToLower(filepath.Ext(path)) == ".com" {
			return path
		}
	}
	for _, path := range paths {
		if strings.ToLower(filepath.Ext(path)) == ".exe" {
			return path
		}
	}
	return ""
}

func DosBinary(filename, zipContent string) string {
	if filename == "" {
		return ""
	}
	if zipContent == "" {
		return filename
	}

	archives := []string{".zip"} // js-dos only supports ZIP archives
	ext := strings.ToLower(filepath.Ext(filename))
	if !slices.Contains(archives, ext) {
		return filename
	}
	paths := DosPaths(zipContent)
	bins := DosBins(paths...)
	switch len(bins) {
	case 0:
		return ""
	case 1:
		return bins[0]
	}
	if s := DosMatch(filename, paths...); s != "" {
		return s
	}
	return DosBin(paths...)
}

func DosFmt(filename string) string {
	const maxLength, extLength = 8, 4
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)

	if len(name) <= maxLength && len(ext) <= extLength {
		return filename
	}

	if len(name) > maxLength {
		return name[:maxLength-2] + "~1" + ext[:extLength]
	}
	return name + ext[:extLength]
}

func JsDosBinary(f *models.File) string {
	if f == nil {
		return ErrModel.Error()
	}
	if f.DoseeRunProgram.Valid && f.DoseeRunProgram.String != "" {
		return f.DoseeRunProgram.String
	}
	if !f.Filename.Valid || f.Filename.IsZero() || f.Filename.String == "" {
		return ""
	}
	name := strings.ToLower(f.Filename.String)
	switch filepath.Ext(name) {
	case ".com", ".exe":
		return DosFmt(f.Filename.String)
	}
	if !f.FileZipContent.Valid || f.FileZipContent.IsZero() || f.FileZipContent.String == "" {
		return ""
	}
	return DosFmt(DosBinary(f.Filename.String, f.FileZipContent.String))
}

func PublishedFmt(f *models.File) string {
	if f == nil {
		return ErrModel.Error()
	}
	ys, ms, ds := "", "", ""
	if f.DateIssuedYear.Valid {
		if i := int(f.DateIssuedYear.Int16); helper.IsYear(i) {
			ys = strconv.Itoa(i)
		}
	}
	if f.DateIssuedMonth.Valid {
		if s := time.Month(f.DateIssuedMonth.Int16); s.String() != "" {
			ms = s.String()
		}
	}
	if f.DateIssuedDay.Valid {
		if i := int(f.DateIssuedDay.Int16); helper.IsDay(i) {
			ds = strconv.Itoa(i)
		}
	}
	if isYearOnly := ys != "" && ms == "" && ds == ""; isYearOnly {
		return ys
	}
	if isInvalidDay := ys != "" && ms != "" && ds == ""; isInvalidDay {
		return ys + " " + ms
	}
	if isInvalid := ys == "" && ms == "" && ds == ""; isInvalid {
		return "unknown date"
	}
	return fmt.Sprintf("%s %s %s", ys, ms, ds)
}

func calc(o, l int) int {
	if o < 1 {
		o = 1
	}
	return (o - 1) * l
}

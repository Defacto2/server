package magicnumberr

// Package file executable.go contains the functions that parse Microsoft and IBM system executable files.

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

// Pklite matches the PKLITE archive format in the byte slice which is a
// compressed executable format for DOS and 16-bit Windows.
func Pklite(r io.ReaderAt) bool {
	const size = 6
	const offset = 30
	p := make([]byte, size)
	sr := io.NewSectionReader(r, offset, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{0x50, 0x4b, 0x4c, 0x49, 0x54, 0x45})
}

// Pksfx matches the PKSFX archive format in the byte slice which is a
// self-extracting archive format.
func Pksfx(r io.ReaderAt) bool {
	const size = 5
	const offset = 526
	p := make([]byte, size)
	sr := io.NewSectionReader(r, offset, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{0x50, 0x4b, 0x53, 0x70, 0x58})
}

// DosKWAJ returns true if the reader begins with the KWAJ compression signature,
// found in some DOS executables.
func DosKWAJ(r io.ReaderAt) bool {
	const size = 8
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{'K', 'W', 'A', 'J', 0x88, 0xf0, 0x27, 0xd1})
}

// DosSZDD returns true if the reader begins with the SZDD compression signature.
func DosSZDD(r io.ReaderAt) bool {
	const size = 8
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{'S', 'Z', 'D', 'D', 0x88, 0xf0, 0x27, 0x33})
}

// MSExe returns true if the reader begins with the Microsoft executable signature.
func MSExe(r io.ReaderAt) bool {
	const size = 2
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return p[0] == 'M' && p[1] == 'Z' || p[0] == 'Z' && p[1] == 'M'
}

// MSComp returns true if the reader contains the Microsoft Compound File signature.
func MSComp(r io.ReaderAt) bool {
	const size = 8
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if n, err := sr.Read(p); err != nil || n < size {
		return false
	}
	return bytes.Equal(p, []byte{0xd0, 0xcf, 0x11, 0xe0, 0xa1, 0xb1, 0x1a, 0xe1})
}

// Windows represents the Windows specific information in the executable header.
type Windows struct {
	TimeDateStamp time.Time          // The time the executable was compiled, only included in PE files
	Major         int                // Major minimum version, for example, Windows 3.0 would be 3
	Minor         int                // Minor minimum version, for example, Windows 3.0 would be 0
	NE            NewExecutable      // The New Executable, a legacy format replaced by the Portable Executable format
	PE            PortableExecutable // The Portable Executable CPU architecture
	PE64          bool               // True if the executable is a 64-bit Portable Executable (PE32+)
}

func (w Windows) String() string {
	const (
		Windows2x   = 2
		WindowsNTv3 = 3
		WindowsNT   = 4
	)
	switch {
	case w.NE == DOSv4Exe, w.NE == OS2Exe:
		return fmt.Sprintf("%s v%d.%d", w.NE, w.Major, w.Minor)
	case w.NE == UnknownNE:
		return "Unknown NE executable"
	}
	switch {
	case w.Major == Windows2x && w.NE == Windows286Exe:
		return fmt.Sprintf("Windows/286 v%d.%d", w.Major, w.Minor)
	case w.Major == Windows2x && w.NE == Windows386Exe:
		return fmt.Sprintf("Windows/386 v%d.%d", w.Major, w.Minor)
	case w.NE == Windows286Exe:
		return fmt.Sprintf("Windows v%d.%d for 286", w.Major, w.Minor)
	case w.NE == Windows386Exe:
		return fmt.Sprintf("Windows v%d.%d for 386+", w.Major, w.Minor)
	}
	switch {
	case w.PE == Intel386PE && w.Major < WindowsNTv3:
		// this is a guess, as Windows 95/98/ME are not part of the NT family
		return "Windows 95/98/ME"
	case w.PE == Intel386PE && w.Major <= WindowsNT:
		return fmt.Sprintf("Windows NT v%d.%d", w.Major, w.Minor)
	}
	os := fmt.Sprintf("Windows NT v%d.%d", w.Major, w.Minor)
	for name, ver := range WindowsNames() {
		if w.Major == ver[0] && w.Minor == ver[1] {
			os = name
			break
		}
	}
	return pe(w.PE, w.PE64, os)
}

func pe(pe PortableExecutable, pe64 bool, os string) string {
	switch {
	case pe == UnknownPE && pe64:
		return "Unknown PE+ executable"
	case pe == UnknownPE:
		return "Unknown PE executable"
	case pe == Intel386PE:
		return os + " 32-bit"
	case pe == AMD64PE:
		return os + " 64-bit"
	case pe == ARMPE:
		return os + " for ARM"
	case pe == ARM64PE:
		return os + " for ARM64"
	case pe == ItaniumPE:
		return os + " for Itanium"
	}
	return ""
}

// WindowsName represents the Windows version names and their minimum version numbers.
type WindowsName map[string][2]int

// WindowsNames returns the Windows version names and their minimum version numbers.
// The minimum version numbers are based on the minimum system version required by the executable,
// and not the libraries or system calls in use by the program.
//
// The minimum version numbers were discontinued by Microsoft in Windows 8.1 and
// may not be accurate for modern programs.
func WindowsNames() WindowsName {
	return WindowsName{
		"Windows 2000":                        {5, 0},
		"Windows XP":                          {5, 1},
		"Windows XP Professional x64 Edition": {5, 2},
		"Windows Vista":                       {6, 0},
		"Windows 7":                           {6, 1},
		"Windows 8":                           {6, 2},
		"Windows 8.1":                         {6, 3},
		"Windows 10":                          {10, 0},
	}
}

// NewExecutable represents the New Executable file type, a format used by Microsoft and IBM
// from the mid-1980s to improve on the limitations of the MS-DOS MZ executable format.
type NewExecutable int

const (
	NoneNE        NewExecutable = iota - 1 // Not a New Executable
	UnknownNE                              // Unknown New Executable
	OS2Exe                                 // Microsoft IBM OS/2 New Executable
	Windows286Exe                          // Windows requiring an Intel 286 CPU New Executable
	DOSv4Exe                               // MS-DOS v4 New Executable
	Windows386Exe                          // Windows requiring an Intel 386 CPU New Executable
)

func (ne NewExecutable) String() string {
	switch ne {
	case NoneNE:
		return "Not a New Executable"
	case UnknownNE:
		return "Unknown New Executable"
	case OS2Exe:
		return "OS/2 New Executable"
	case Windows286Exe:
		return "Windows for 286 New Executable"
	case DOSv4Exe:
		return "MS-DOS v4 New Executable"
	case Windows386Exe:
		return "Windows for 386+ New Executable"
	}
	return ""
}

// PortableExecutable represents the Portable Executable file type, a format used by Microsoft
// for executables, object code, DLLs, FON Font files, and others. In this implementation, only
// executables for desktop Windows are considered.
type PortableExecutable uint16

const (
	UnknownPE  PortableExecutable = 0x0    // Unknown Portable Executable
	Intel386PE PortableExecutable = 0x14c  // Intel 386 Portable Executable
	AMD64PE    PortableExecutable = 0x8664 // AMD64 Portable Executable
	ARMPE      PortableExecutable = 0x1c0  // ARM Portable Executable
	ARM64PE    PortableExecutable = 0xaa64 // ARM64 Portable Executable
	ItaniumPE  PortableExecutable = 0x200  // Itanium Portable Executable
)

// FindExecutable reads the first 1KB from the reader and returns the specific information contained
// within the executable headers. Both the New Executable and Portable Executable formats are supported,
// which are commonly used by IBM and Microsoft desktop operating systems from PC/MS-DOS to modern Windows.
func FindExecutable(r io.ReaderAt) (Windows, error) {
	win := Default()
	if r == nil {
		return win, fmt.Errorf("nil reader")
	}
	const size = 1024 * 3
	p := make([]byte, size)
	sr := io.NewSectionReader(r, 0, size)
	if _, err := sr.Read(p); err != nil {
		return win, fmt.Errorf("magic number find first %d bytes: %w", size, err)
	}
	win = NE(p)
	if win.NE == NoneNE {
		win = PE(p)
	}
	return win, nil
}

func Default() Windows {
	return Windows{
		Major:         0,
		Minor:         0,
		TimeDateStamp: time.Time{},
		PE64:          false,
		PE:            UnknownPE,
		NE:            NoneNE,
	}
}

// NE returns the New Executable file type from the byte slice.
//
// Windows programs that are New Executables are usually for the ancient Windows 2 or 3.x editions.
// Windows v2 came in two versions, Windows 2 (for the 286 CPU) and Windows/386,
// while Windows 3.0+ unified support for both CPUs.
// The New Executable format was replaced by the Portable Executable format in Windows 95/NT.
//
// If a Windows program is detected, the major and minor version numbers are returned,
// for example, a Windows 3.0 requirement would return 3 and 0.
func NE(p []byte) Windows {
	none := Default()
	const min = 64
	if len(p) < min {
		return none
	}
	if p[0] != 'M' || p[1] != 'Z' {
		return none
	}
	const segmentedHeaderIndex = 0x3c // the location of the segmented header
	const executableTypeIndex = 0x36  // the executable type aka the operating system
	const winMinorIndex = 0x3e        // the location of the Windows minor version
	const winMajorIndex = 0x3f        // the location of the Windows major version
	offset := binary.LittleEndian.Uint16(p[segmentedHeaderIndex:])
	if len(p) < int(offset)+int(winMajorIndex) {
		return none
	}
	segmentedHeader := [2]byte{
		p[offset+0],
		p[offset+1],
	}
	if segmentedHeader != [2]byte{'N', 'E'} {
		return none
	}
	minor := int(p[offset+winMinorIndex])
	major := int(p[offset+winMajorIndex])
	newType := NewExecutable(p[offset+executableTypeIndex])
	w := Windows{}
	switch newType {
	case Windows286Exe, Windows386Exe, OS2Exe, DOSv4Exe, UnknownNE:
		w.Major = major
		w.Minor = minor
		w.NE = newType
		return w
	}
	return none
}

// PE returns the Portable Executable file type from the byte slice.
//
// The [Portable Executable format] is used by Microsoft for executables, object code, DLLs, FON Font files, and others.
// In this implementation, only executables for desktop Windows are considered. The information returned is the
// CPU architecture, the Windows NT version, and the time the executable was compiled.
//
// The major and minor version numbers are not always accurate.
//
// [Portable Executable format]: https://learn.microsoft.com/en-us/windows/win32/debug/pe-format
func PE(p []byte) Windows {
	none := Default()
	const min = 64
	if len(p) < min {
		return none
	}
	if p[0] != 'M' || p[1] != 'Z' {
		return none
	}
	// the location of the portable executable header
	const peHeaderIndex = 0x3c
	offset := binary.LittleEndian.Uint16(p[peHeaderIndex:])
	if len(p) < int(offset) {
		return none
	}

	signature := [4]byte{p[offset+0], p[offset+1], p[offset+2], p[offset+3]}
	if signature != [4]byte{'P', 'E', 0, 0} {
		return none
	}
	// the location of the COFF (Common Object File Format) header
	coffHeaderIndex := offset + uint16(len(signature))
	const coffLen = 20
	if len(p) < int(coffHeaderIndex)+coffLen {
		return none
	}
	machine := [2]byte{
		p[coffHeaderIndex],
		p[coffHeaderIndex+1],
	}
	timeDateStamp := binary.LittleEndian.Uint32(p[coffHeaderIndex+4:])
	compiled := time.Unix(int64(timeDateStamp), 0)

	optionalHeaderIndex := coffHeaderIndex + coffLen
	magic := [2]byte{
		p[optionalHeaderIndex+0],
		p[optionalHeaderIndex+1],
	}

	const winMajorOffset = 40 // the location of the Windows major version
	const winMinorOffset = 42 // the location of the Windows minor version
	major := optionalHeaderIndex + winMajorOffset
	osMajorB := []byte{
		p[major+0],
		p[major+1],
	}
	minor := optionalHeaderIndex + winMinorOffset
	osMinorB := []byte{
		p[minor+0],
		p[minor+1],
	}
	osMajor := int(binary.LittleEndian.Uint16(osMajorB))
	osMinor := int(binary.LittleEndian.Uint16(osMinorB))
	w := Windows{
		Major:         osMajor,
		Minor:         osMinor,
		TimeDateStamp: compiled,
		PE64:          magic == [2]byte{0x0b, 0x02},
		NE:            NoneNE,
	}
	pem := binary.LittleEndian.Uint16(machine[:])
	w.PE = portexec(pem)
	return w
}

func portexec(pem uint16) PortableExecutable {
	switch PortableExecutable(pem) {
	case Intel386PE:
		return Intel386PE
	case AMD64PE:
		return AMD64PE
	case ARMPE:
		return ARMPE
	case ARM64PE:
		return ARM64PE
	case ItaniumPE:
		return ItaniumPE
	default:
		return UnknownPE
	}
}

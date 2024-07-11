// Package jsdos configures the js-dos v6.22 emulator.
package jsdos

import (
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const AudioRate = "44100" // AudioRate is the sample rate of the audio that is emulated.

const (
	auto = "auto"
	off  = "off"
	none = "none"
	no   = "false"
	yes  = "true"
)

// Jsdos is the dynamic generated configuration file for the js-dos v6.22 emulator.
type Jsdos struct {
	Dosbox    `ini:"dosbox"`
	Processor `ini:"cpu"`
	Midi      `ini:"midi"`
	SBlaster  `ini:"sblaster"`
	GUS       `ini:"gus"`
	Speaker   `ini:"speaker"`
	DOS       `ini:"dos"`
}

// Dosbox is the [dosbox] section of the configuration file.
type Dosbox struct {
	Machine Platform `ini:"machine"`
	MemSize RAM      `ini:"memsize"`
}

// Processor is the [cpu] section of the configuration file.
type Processor struct {
	Core   Core   `ini:"core"`
	Model  CPU    `ini:"cputype"`
	Cycles Cycles `ini:"cycles"`
}

// Midi is the [midi] section of the configuration file.
type Midi struct {
	MPU401 string `ini:"mpu401"`
	Device string `ini:"mididevice"`
}

// SBlaster is the [sblaster] section of the configuration file.
type SBlaster struct {
	Type    string `ini:"sbtype"`
	Base    string `ini:"sbbase"`
	IRQ     string `ini:"irq"`
	DMA     string `ini:"dma"`
	HDMA    string `ini:"hdma"`
	Mixer   string `ini:"sbmixer"`
	OplMode string `ini:"oplmode"`
	OplRate string `ini:"oplrate"`
	OplEmu  string `ini:"oplemu"`
}

// GUS is the [gus] section of the configuration file.
type GUS struct {
	Enable string `ini:"gus"`      // Enable is the Gravis UltraSound card.
	Rate   string `ini:"gusrate"`  // Rate is the sample rate of the Gravis UltraSound card.
	Base   string `ini:"gusbase"`  // Base is the I/O port address of the Gravis UltraSound card.
	IRQ    string `ini:"gusirq"`   // IRQ is the interrupt request line of the Gravis UltraSound card.
	DMA    string `ini:"gusdma"`   // DMA is the direct memory access channel of the Gravis UltraSound card.
	Dir    string `ini:"ultradir"` // Dir is the directory within the emulation where the driver patch files for GUS playback are located.
}

// Speaker is the [speaker] section of the configuration file.
type Speaker struct {
	Enable    string `ini:"pcspeaker"` // Enable is the PC speaker audio.
	Rate      string `ini:"pcrate"`    // Rate is the sample rate of the PC speaker.
	Tandy     string `ini:"tandy"`     // Tandy 3-channel sound chip emulation.
	TandyRate string `ini:"tandyrate"` // TandyRate is sample rate of the Tandy 3-channel sound chip.
	Disney    string `ini:"disney"`    // Disney Sound Source aka Covox Speech Thing emulation.
}

// DOS is the [dos] section of the configuration file.
type DOS struct {
	XMS string `ini:"xms"` // XMS is the Extended Memory used for programs that require more than 1 MB of memory.
	EMS string `ini:"ems"` // EMS is the Expanded Memory used for programs that require more than 1 MB of memory.
	UMB string `ini:"umb"` // UMB is the Upper Memory Blocks used for programs that require more than 640 KB of memory.
}

type Platform string // Platform is the machine dosbox tries to emulate.

const (
	Hercules         Platform = "hercules"      // Hercules graphics mode circa 1982.
	CGA              Platform = "cga"           // CGA (Color Graphics Adapter) graphics mode circa 1981.
	Tandy            Platform = "tandy"         // Tandy 1000 emulation which uses the Tandy Graphics Adapter and the 3-channel Tandy speaker circa 1984.
	PCjr             Platform = "pcjr"          // PCjr (Personal Computer Junior) emulation which uses the PCjr graphics and sound circa 1984.
	EGA              Platform = "ega"           // EGA (Enhanced Graphics Adapter) graphics mode circa 1984.
	VGAOnly          Platform = "vgaonly"       // VGAOnly is the VGA graphics mode for compatibility with software that fails with SuperVGA.
	SuperVgaS3       Platform = "svga_s3"       // SuperVgaS3 is SuperVGA using the s3 Trio 64 chip.
	SuperVgaET3000   Platform = "svga_et3000"   // SuperVgaET3000 is SuperVGA using the Tseng Labs ET3000 graphics mode.
	SuperVgaET4000   Platform = "svga_et4000"   // SuperVgaET4000 is SuperVGA using the Tseng Labs ET4000 graphics mode.
	SuperVgaParadise Platform = "svga_paradise" // SuperVgaParadise is SuperVGA using the Paradise PVGA1A chip, common in the late 1980s.
	VesaNoFrameBuff  Platform = "vesa_nolfb"    // VesaNoFrameBuff is the VESA graphics mode without the Linear Frame Buffer that is sometimes faster than SuperVgaS3.
	VesaV1           Platform = "vesa_oldvbe"   // VesaV1 is the VESA graphics mode using the old VBE 1.2 standard.
)

type RAM string // RAM is the amount of memory to emulate.

const (
	Mem286 RAM = "1"  // Mem286 is 1MB of RAM.
	Mem386 RAM = "4"  // Mem386 is 4MB of RAM.
	Mem486 RAM = "16" // Mem486 is 16MB of RAM.
)

type Core string // Core used in the CPU emulation.

const (
	AutoCore Core = auto     // Auto sets real-mode programs to use the normal core, for protected mode programs it switches to dynamic core.
	Dynamic  Core = auto     // [unsupported] "dynamic"  Dynamic is the optimal core for most games, except for programs that employ massive self-modifying code.
	Normal   Core = "normal" // Normal has the program interpreted instruction by instruction.
	Simple   Core = "simple" // Simple has the program interpreted instruction by instruction, optimized for (8088/8086/286) real-mode games.
)

type CPU string // CPU model and method used in the emulation.

const (
	IAuto    CPU = auto           // IAuto is the fastest CPU type.
	I386     CPU = "386"          // I386 is the optimal Intel 80386.
	I386Pre  CPU = "386_prefetch" // I386Pre is the optimal Intel 80386 with a normal core.
	I386Slow CPU = "386_slow"     // I386Slow is the slowest Intel 80386 with a normal core.
	I486Slow CPU = "486_slow"     // I486Slow is the Intel 80486.
	I586Slow CPU = "pentium_slow" // I586Slow is the Intel 80586 aka Pentium.
)

type Cycles string // Cycles is the amount of instructions to attempt to emulate each millisecond.

const (
	AutoCycles Cycles = auto          // CycleAuto, real-mode programs will run at 3000 cycles and protected mode games run with Max.
	Max        Cycles = "max"         // Max all programs run at the maximum speed the browser permits.
	Fix5Mhz    Cycles = "fixed 330"   // Fix5Mhz attempts to run at a fixed 5MHz 8086, 0.330 MIPS.
	Fix10Mhz   Cycles = "fixed 750"   // Fix10Mhz attempts to run at a fixed 10MHz 8088, 0.750 MIPS.
	Fix12Mhz   Cycles = "fixed 1280"  // Fix12Mhz attempts to run at a fixed 12MHz 80286, 1.280 MIPS.
	Fix16Mhz   Cycles = "fixed 2150"  // Fix16Mhz attempts to run at a fixed 16MHz 80386DX, 2.15 MIPS.
	Fix33Mhz   Cycles = "fixed 4300"  // Fix33Mhz attempts to run at a fixed 33MHz 80386DX, 4.3 MIPS.
	Fix25Mhz   Cycles = "fixed 8700"  // Fix25Mhz attempts to run at a fixed 25MHz 80486DX, 8.7 MIPS.
	Fix66Mhz   Cycles = "fixed 25600" // Fix25Mhz attempts to run at a fixed 66MHz 80486DX2, 25.6 MIPS.
)

// CPU sets the CPU model, core and speed used in the emulation.
// The value can be "8086", "386", "486", or "auto".
func (j *Jsdos) CPU(value string) {
	switch strings.ToLower(value) {
	case "8086":
		j.Processor.Model = I386Slow
		j.Processor.Core = Simple
		j.Processor.Cycles = Fix5Mhz
	case "386":
		j.Processor.Model = I386Slow
		j.Processor.Core = Normal
	case "486":
		j.Processor.Model = IAuto
		j.Processor.Core = Normal
		j.Processor.Cycles = Max
	case "auto":
		j.Processor.Model = IAuto
		j.Processor.Core = AutoCore
		j.Processor.Cycles = AutoCycles
	default:
		// use the js-dos defaults
	}
}

// Machine sets the machine to emulate, either the personal computer series or the graphics hardware.
// The value can be "vga", "tandy", "svga", "paradise", "oldvbe", "nolfb", "et4000", "et3000", "ega", or "cga".
// The "tandy" machine will also enable the Tandy 3-channel sound chip.
func (j *Jsdos) Machine(value string) {
	// todo: fix, DEFAULT (svga_s3)
	// Possible values: hercules, cga, tandy, pcjr, ega, vgaonly, svga_s3, svga_et3000, svga_et4000, svga_paradise, vesa_nolfb, vesa_oldvbe
	switch strings.ToLower(value) {
	case "vga":
		j.Dosbox.Machine = VGAOnly
		j.Dosbox.MemSize = Mem486
	case "tandy":
		j.Dosbox.Machine = Tandy
		j.Dosbox.MemSize = Mem286
		j.Tandy()
	case "svga":
		j.Dosbox.Machine = SuperVgaS3
		j.Dosbox.MemSize = Mem486
	case "paradise":
		j.Dosbox.Machine = SuperVgaParadise
		j.Dosbox.MemSize = Mem486
	case "oldvbe":
		j.Dosbox.Machine = VesaV1
		j.Dosbox.MemSize = Mem486
	case "nolfb":
		j.Dosbox.Machine = VesaNoFrameBuff
		j.Dosbox.MemSize = Mem486
	case "et4000":
		j.Dosbox.Machine = SuperVgaET4000
		j.Dosbox.MemSize = Mem486
	case "et3000":
		j.Dosbox.Machine = SuperVgaET3000
		j.Dosbox.MemSize = Mem486
	case "ega":
		j.Dosbox.Machine = EGA
		j.Dosbox.MemSize = Mem386
	case "cga":
		j.Dosbox.Machine = CGA
		j.Dosbox.MemSize = Mem286
	default:
		// use the js-dos defaults
	}
}

// Sound sets the sound card and audio output.
// The value can be "sb16", "sb1", "pcspeaker", "none", "gus", or "covox".
func (j *Jsdos) Sound(value string) {
	const SoundBlasterV1, SoundBlaster16 = "sb1", "sb16"
	switch strings.ToLower(value) {
	case "sb16":
		j.NoMIDI()
		j.FM()
		j.NoGUS()
		j.Beeper()
		j.SBlaster.Type = SoundBlaster16
		j.SBlaster.Base = "220"
		j.SBlaster.IRQ = "7"
		j.SBlaster.DMA = "1"
		j.SBlaster.HDMA = "5"
		j.SBlaster.Mixer = yes
	case "sb1":
		// s = append(s, "[autoexec]", "PATH=S:\\DRIVERS\\CT-1320C\\;S:\\DRIVERS\\ADLIB\\;%PATH%", "SET SOUND=S:\\DRIVERS\\CT-1320C\\")
		j.NoMIDI()
		j.FM()
		j.NoGUS()
		j.Beeper()
		j.SBlaster.Type = SoundBlasterV1
		j.SBlaster.Base = "220"
		j.SBlaster.IRQ = "7"
		j.SBlaster.DMA = "1"
		j.SBlaster.Mixer = yes
	case "pcspeaker":
		j.Beeper()
	case none:
		j.NoMIDI()
		j.NoBlaster()
		j.NoGUS()
		j.NoBeeper()
	case "gus":
		j.NoMIDI()
		j.NoBlaster()
		j.Covox()
		j.GUS.Enable = yes
		j.GUS.Rate = AudioRate
		j.GUS.Base = "240"
		j.GUS.IRQ = "5"
		j.GUS.DMA = "1"
		j.GUS.Dir = "" // "C:\\ULTRASND"
	case "covox":
		j.NoMIDI()
		j.NoBlaster()
		j.Covox()
	default:
		// use the js-dos defaults
	}
}

// Beeper enables the PC speaker audio.
func (j *Jsdos) Beeper() {
	j.Speaker.Enable = yes
	j.Speaker.Rate = AudioRate
	j.Speaker.Tandy = off
	j.Speaker.Disney = no
}

// Covox enables the Covox Speech Thing audio also sold as the Disney Sound Source.
func (j *Jsdos) Covox() {
	j.Speaker.Enable = yes
	j.Speaker.Rate = AudioRate
	j.Speaker.Tandy = off
	j.Speaker.Disney = yes
}

// FM enables FM synthesis music aka the AdLib range of cards.
func (j *Jsdos) FM() {
	j.OplMode = "auto"
	j.OplRate = AudioRate
	j.OplEmu = "default"
}

// Tandy enables the Tandy 1000 series, 3-channel sound chip.
func (j *Jsdos) Tandy() {
	j.Speaker.Enable = yes
	j.Speaker.Rate = AudioRate
	j.Speaker.Tandy = yes
	j.Speaker.TandyRate = AudioRate
	j.Speaker.Disney = no
}

// NoEMS disables the use of the EMS memory.
// This is useful for software that does not support the Expanded Memory Specification.
func (j *Jsdos) NoEMS(value bool) {
	switch value {
	case true:
		j.DOS.EMS = no
	case false:
		return
	}
}

// NoXMS disables the use of the XMS memory.
// This is useful for software that does not support the Extended Memory Specification.
func (j *Jsdos) NoXMS(value bool) {
	switch value {
	case true:
		j.DOS.XMS = no
	case false:
		return
	}
}

// NoUMB disables the use of the upper memory blocks.
// This is useful for software that conflicts with the UMB memory.
func (j *Jsdos) NoUMB(value bool) {
	switch value {
	case true:
		j.DOS.UMB = no
	case false:
		return
	}
}

// NoBeeper disables the use of the PC speaker.
func (j *Jsdos) NoBeeper() {
	j.Speaker.Enable = no
	j.Speaker.Tandy = off
	j.Speaker.Disney = no
}

// NoBlaster disables the use of the Sound Blaster card.
func (j *Jsdos) NoBlaster() {
	j.SBlaster.Type = none
	j.OplMode = none
}

// NoGUS disables the use of the Gravis UltraSound card.
func (j *Jsdos) NoGUS() {
	j.GUS.Enable = no
}

// NoMIDI disables the use of the MIDI devices for music.
func (j *Jsdos) NoMIDI() {
	j.Midi.MPU401 = none
	j.Midi.Device = none
}

// Binaries returns a list of the executable files from the archive paths.
func Binaries(paths ...string) []string {
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

// Binary returns a path to the most likely executable file from the archive paths.
// The search order is .bat, .com, .exe. with the root paths having priority.
// If no executable is found then an empty string is returned.
func Binary(paths ...string) string {
	if len(paths) == 0 {
		return ""
	}
	// sort by the number of directories in the path, to prioritize binaries in the root of the archive
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

// FindBinary returns the most likely executable file from the archive paths.
// Binaries that matches the filename take priority over other executables.
// If no executable is found then an empty string is returned.
func FindBinary(filename, zipContent string) string {
	if filename == "" {
		return ""
	}
	if zipContent == "" {
		return filename
	}

	archives := []string{".zip", ".lha", ".lzh", ".arc", ".arj"}
	ext := strings.ToLower(filepath.Ext(filename))
	if !slices.Contains(archives, ext) {
		return filename
	}
	paths := Paths(zipContent)
	bins := Binaries(paths...)
	switch len(bins) {
	case 0:
		return ""
	case 1:
		return bins[0]
	}
	if s := Finds(filename, paths...); s != "" {
		return s
	}
	return Binary(paths...)
}

// Finds the most likely executable in the archive paths.
// Binaries that matches the filename take priority over other executables.
// If no executable is found then an empty string is returned.
func Finds(filename string, paths ...string) string {
	if filename == "" || len(paths) == 0 {
		return ""
	}
	// sort by the number of directories in the path, to prioritize binaries in the root of the archive
	sort.Slice(paths, func(i, j int) bool {
		return len(filepath.SplitList(paths[i])) < len(filepath.SplitList(paths[j]))
	})
	// prioritize the most likely executable that matches the archive name
	// e.g. if the archive is 'myapp.zip' then the most likely executable order are
	// 'myapp.exe', 'myapp.com', 'myapp.bat'
	root := paths
	sort.Slice(root, func(i, _ int) bool { // unused j variable
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

// Fmt8dot3 returns a DOS 8.3 filename format, truncating the filename if necessary.
// For example, "my backup collection.7zip" would return "my bac~1.7zi".
func Fmt8dot3(filename string) string {
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

func special(r rune) bool {
	const (
		underscore  = '_'
		caret       = '^'
		dollar      = '$'
		tilde       = '~'
		exclamation = '!'
		number      = '#'
		percent     = '%'
		ampersand   = '&'
		hyphen      = '-'
		open        = '{'
		close       = '}'
		at          = '@'
		quote       = '`'
		apostrophe  = '\''
		openParen   = '('
		closeParen  = ')'
	)
	switch r {
	case underscore, caret, dollar, tilde, exclamation, number,
		percent, ampersand, hyphen, open, close, at, quote,
		apostrophe, openParen, closeParen:
		return true
	}
	return false
}

func Fat16Rename(filename string) string {
	if filename == "" {
		return ""
	}
	// A-Z 0-9
	name := strings.TrimSpace(strings.ToUpper(filename))
	// Remove diacritics and accents from the filename.
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	s, _, _ := transform.String(t, name)

	p := []rune(s)
	l := len(p)
	const extLength = 3
	for i, r := range p {
		if unicode.Is(unicode.Letter, r) || unicode.Is(unicode.Number, r) {
			continue
		}
		if special(r) {
			continue
		}
		if unicode.Is(unicode.Space, r) {
			p[i] = '_'
			continue
		}
		if r == '.' {
			continue
		}
		p[i] = 'X'
	}
	// handle single . for extension by matching the length of the extension vs the i value
	// ! # $ % & ' ( ) - @ ^ _ ` { } ~

	// none: ", *, +, ,, /, :, ;, <, =, >, ?, \, [, ], |

	// Can contain only the letters A through Z, the numbers 0 through 9, and the following special characters:
	// underscore (_), caret (^), dollar sign ($), tilde (~), exclamation point (!), number sign (#), percent sign (%), ampersand (&), hyphen (-), braces ({}), at sign (@), single quotation mark (`), apostrophe ('), and parentheses (). No other special characters are acceptable.
	// Cannot contain spaces, commas, backslashes, or periods (except the period that separates the name from the extension).
	// Cannot be identical to the name of another file or subdirectory in the same directory.

	return string(p)
}

// Paths returns a list of file and directory paths from the zip content.
func Paths(zipContent string) []string {
	if zipContent == "" {
		return []string{}
	}
	const delimiter = ":" // the colon is an illegal character as a DOS filename
	archive := zipContent
	archive = strings.ReplaceAll(archive, "\r\n", delimiter) // replace Microsoft-style CRLF with delimiter
	archive = strings.ReplaceAll(archive, "\n", delimiter)   // replace Unix LF with delimiter
	archive = strings.ReplaceAll(archive, "\r", delimiter)   // replace 8-bit microcomputer era CR with delimiter
	paths := strings.Split(archive, delimiter)
	// FOR LATER, convert into DOS 8.3 filename format?
	return paths
}

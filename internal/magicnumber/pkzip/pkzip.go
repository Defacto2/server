// Package pkzip provides constants and functions for working with PKZip files to determine the compression methods used.
// Most modern zip files use the Deflated method, which is supported by the Go standard library's archive/zip package and
// the Stored method, which is uncompressed.
//
// But many older zip files use other compression methods, such as Shrunk, Reduced, Imploded, and others. This package
// provides a way to determine the compression methods used in a zip file and whether the file should be handled by a
// third-party application installed on the host system.
package pkzip

import (
	"archive/zip"
	"slices"
	"strconv"
	"strings"
)

// Compression is the PKZip compression method used by a ZIP archive file.
type Compression uint16

const (
	Stored Compression = iota
	Shrunk
	ReducedFactor1
	ReducedFactor2
	ReducedFactor3
	ReducedFactor4
	Imploded
	Reserved
	Deflated
	EnhancedDeflated
	PKWareDataCompressionLibraryImplode
	BZIP2
	Reserved2
	LZMA
	Reserved3
	Reserved4
	IBMTERSE
	IBMLZ77z
)

const (
	PPMd1 Compression = iota + 98
)

func (c Compression) String() string {
	switch c {
	case Stored:
		return "Stored"
	case Shrunk:
		return "Shrunk"
	case ReducedFactor1:
		return "Reduced with factor 1"
	case ReducedFactor2:
		return "Reduced with factor 2"
	case ReducedFactor3:
		return "Reduced with factor 3"
	case ReducedFactor4:
		return "Reduced with factor 4"
	case Imploded:
		return "Imploded"
	case Reserved:
		return "Reserved"
	case Deflated:
		return "Deflated"
	case EnhancedDeflated:
		return "Enhanced Deflated"
	case PKWareDataCompressionLibraryImplode:
		return "PKWare Data Compression Library Imploded"
	case BZIP2:
		return "BZIP2"
	case Reserved2:
		return "Reserved"
	case LZMA:
		return "LZMA"
	case Reserved3:
		return "Reserved"
	case Reserved4:
		return "Reserved"
	case IBMTERSE:
		return "IBM TERSE"
	case IBMLZ77z:
		return "IBM LZ77z"
	case PPMd1:
		return "PPMd version I, Rev 1"
	default:
		return "Unknown"
	}
}

// Zip returns true if the compression method is Deflated or Stored.
func (c Compression) Zip() bool {
	switch c {
	case Stored, Deflated:
		return true
	}
	return false
}

// Diagnostic is a diagnostic code returned by the PKZip command-line utilities.
type Diagnostic uint16

const (
	Normal Diagnostic = iota
	Warning
	GenericError
	SevereError
	BufferError
	TTYError
	DiskError
	MemoryError
	Unused
	ZipNotFound
	OptionsError
	FilesNoFound
	ZipBomb
)

const (
	DiskFull          Diagnostic = 50
	PrematureExit     Diagnostic = 51
	UserAbort         Diagnostic = 80
	CompressionMethod Diagnostic = 81
	BadDecryption     Diagnostic = 82
)

func (d Diagnostic) String() string {
	switch d {
	case Normal:
		return "No errors or warnings"
	case Warning:
		return "One or more warning errors were encountered, but processing completed successfully anyway."
	case GenericError:
		return "A generic error in the zipfile format was detected. Processing may have completed successfully anyway; some broken zipfiles created by other archivers have simple work-arounds, but if the zipfile is created by PKZIP, please report the problem to PKWARE, Inc."
	case SevereError:
		return "A severe error in the zipfile format was detected. Processing probably failed immediately."
	case BufferError:
		return "Insufficient memory to perform operation"
	case TTYError:
		return "TTY user input error"
	case DiskError:
		return "Decompression to disk error"
	case MemoryError:
		return "Decompression in-memory error"
	case Unused:
		return "Unused"
	case ZipNotFound:
		return "Zip file not found"
	case OptionsError:
		return "Invalid command line options"
	case FilesNoFound:
		return "Zipfiles not found"
	case ZipBomb:
		return "Zip bomb detected"
	case DiskFull:
		return "Disk full"
	case PrematureExit:
		return "Unexpected premature exit"
	case UserAbort:
		return "User abort exit using control-C"
	case CompressionMethod:
		return "Unsupported ZIP compression method found in the archive"
	case BadDecryption:
		return "Bad decryption"
	default:
		return "Unknown"
	}
}

// ExitStatus is intended to be used with the exec.Command.Run method to determine
// the exit status of the PKZip command-line utilities.
//
//	err := exec.Command("unzip", "-T", "archive.zip").Run()
//	if err != nil {
//		diag := pkzip.ExitStatus(err)
//		switch diag {
//		case pkzip.Normal, pkzip.Warning:
//			// normal or warnings are fine
//			return nil
//		}
//		return fmt.Errorf("unzip test failed: %s", diag)
//	}
func ExitStatus(err error) Diagnostic {
	if err == nil {
		return Normal
	}
	const status = "exit status"
	if !strings.HasPrefix(err.Error(), status) {
		return Diagnostic(99)
	}
	s := strings.TrimSpace(strings.TrimPrefix(err.Error(), status))
	code, err := strconv.Atoi(s)
	if err != nil {
		return Diagnostic(99)
	}
	return Diagnostic(code)
}

// Methods returns the PKZip compression methods used in the named file.
func Methods(name string) ([]Compression, error) {
	r, err := zip.OpenReader(name)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	methods := []Compression{}
	for _, file := range r.File {
		fh := file.FileHeader
		methods = append(methods, Compression(fh.Method))
	}
	slices.Sort(methods)
	return slices.Compact(methods), nil
}

// Zip returns true if the named file is a PKZip file that exclusively
// uses the Deflated or Stored compression methods. These are the methods
// supported by the Go standard library's archive/zip package.
func Zip(name string) (bool, error) {
	methods, err := Methods(name)
	if err != nil {
		return false, err
	}
	for _, method := range methods {
		if !method.Zip() {
			return false, nil
		}
	}
	return true, nil
}

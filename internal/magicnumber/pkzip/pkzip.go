// Package pkzip provides constants and functions for working with
// PKZip files to determine the compression methods used.
// Most modern zip files use the Deflated method, which is supported
// by the Go standard library's archive/zip package and
// the Stored method, which is uncompressed.
//
// But many older zip files use other compression methods, such as
// Shrunk, Reduced, Imploded, and others. This package
// provides a way to determine the compression methods used in a zip
// file and whether the file should be handled by a
// third-party application installed on the host system.
package pkzip

import (
	"archive/zip"
	"fmt"
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

const reserved = "Reserved"

func (c Compression) String() string {
	compress := map[Compression]string{
		Stored:                              "Stored",
		Shrunk:                              "Shrunk",
		ReducedFactor1:                      "Reduced with factor 1",
		ReducedFactor2:                      "Reduced with factor 2",
		ReducedFactor3:                      "Reduced with factor 3",
		ReducedFactor4:                      "Reduced with factor 4",
		Imploded:                            "Imploded",
		Reserved:                            reserved,
		Deflated:                            "Deflated",
		EnhancedDeflated:                    "Enhanced Deflated",
		PKWareDataCompressionLibraryImplode: "PKWare Data Compression Library Imploded",
		BZIP2:                               "BZIP2",
		Reserved2:                           reserved,
		LZMA:                                "LZMA",
		Reserved3:                           reserved,
		Reserved4:                           reserved,
		IBMTERSE:                            "IBM TERSE",
		IBMLZ77z:                            "IBM LZ77z",
		PPMd1:                               "PPMd version I, Rev 1",
	}
	if name, known := compress[c]; known {
		return name
	}
	return reserved
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

const generic = "A generic error in the zipfile format was detected. " +
	"Processing may have completed successfully anyway; " +
	"some broken zipfiles created by other archivers have simple work-arounds, " +
	"but if the zipfile is created by PKZIP, please report the problem to PKWARE, Inc."

const server = "A severe error in the zipfile format was detected. " +
	"Processing probably failed immediately."

func (d Diagnostic) String() string {
	diag := map[Diagnostic]string{
		Normal:            "No errors or warnings",
		Warning:           "One or more warning errors were encountered, but processing completed successfully anyway.",
		GenericError:      generic,
		SevereError:       server,
		BufferError:       "Insufficient memory to perform operation",
		TTYError:          "TTY user input error",
		DiskError:         "Decompression to disk error",
		MemoryError:       "Decompression in-memory error",
		Unused:            "Unused",
		ZipNotFound:       "Zip file not found",
		OptionsError:      "Invalid command line options",
		FilesNoFound:      "Zipfiles not found",
		ZipBomb:           "Zip bomb detected",
		DiskFull:          "Disk full",
		PrematureExit:     "Unexpected premature exit",
		UserAbort:         "User abort exit using control-C",
		CompressionMethod: "Unsupported ZIP compression method found in the archive",
		BadDecryption:     "Bad decryption",
	}
	if problem, known := diag[d]; known {
		return problem
	}
	return "Unknown"
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
	const (
		status = "exit status"
		unused = 99
	)
	if !strings.HasPrefix(err.Error(), status) {
		return Diagnostic(unused)
	}
	s := strings.TrimSpace(strings.TrimPrefix(err.Error(), status))
	code, err := strconv.Atoi(s)
	if err != nil {
		return Diagnostic(unused)
	}
	return Diagnostic(code)
}

// Methods returns the PKZip compression methods used in the named file.
func Methods(name string) ([]Compression, error) {
	r, err := zip.OpenReader(name)
	if err != nil {
		return nil, fmt.Errorf("pkzip methods: %w", err)
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
		return false, fmt.Errorf("pkzip deflate or store check: %w", err)
	}
	for _, method := range methods {
		if !method.Zip() {
			return false, nil
		}
	}
	return true, nil
}

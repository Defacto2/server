// Package helpers are general, shared functions.
package helper

import (
	"crypto/sha512"
	"embed"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os"
)

const (
	// Eraseline is an ANSI escape control to erase the active line of the terminal.
	Eraseline = "\x1b[2K"
	// ReadWrite is the file mode for read and write access.
	ReadWrite = 0o666
	// byteUnits is a list of units used for formatting byte sizes.
	byteUnits = "kMGTPE"
)

var (
	ErrDiffLength = fmt.Errorf("files are of different lengths")
	ErrDirPath    = errors.New("directory path is a file")
	ErrFilePath   = errors.New("file path is a directory")
	ErrExistPath  = errors.New("path ready exists and will not overwrite")
	ErrOSFile     = errors.New("os file is nil")
	ErrRead       = fmt.Errorf("could not read files")
)

// GetLocalIPs returns a list of local IP addresses.
// credit: https://gosamples.dev/local-ip-address/
func GetLocalIPs() ([]net.IP, error) {
	var ips []net.IP
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addresses {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP)
			}
		}
	}
	return ips, nil
}

// GetLocalHosts returns a list of local hostnames.
func GetLocalHosts() ([]string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	hosts := []string{}
	hosts = append(hosts, hostname)
	// confirm localhost is resolvable
	if _, err = net.LookupHost("localhost"); err != nil {
		return nil, err
	}
	hosts = append(hosts, "localhost")
	return hosts, nil
}

// Integrity returns the sha384 hash of the named embed file.
// This is intended to be used for Subresource Integrity (SRI)
// verification with integrity attributes in HTML script and link tags.
func Integrity(name string, fs embed.FS) (string, error) {
	b, err := fs.ReadFile(name)
	if err != nil {
		return "", err
	}
	return IntegrityBytes(b), nil
}

// IntegrityFile returns the sha384 hash of the named file.
// This can be used as a link cache buster.
func IntegrityFile(name string) (string, error) {
	b, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}
	return IntegrityBytes(b), nil
}

// IntegrityBytes returns the sha384 hash of the given byte slice.
func IntegrityBytes(b []byte) string {
	sum := sha512.Sum384(b)
	b64 := base64.StdEncoding.EncodeToString(sum[:])
	return fmt.Sprintf("sha384-%s", b64)
}

// Touch creates a new, empty named file.
// If the file already exists, an error is returned.
func Touch(name string) error {
	file, err := os.OpenFile(name, os.O_CREATE|os.O_EXCL, ReadWrite)
	if err != nil {
		return err
	}
	return file.Close()
}

// TouchW creates a new named file with the given data.
// If the file already exists, an error is returned.
func TouchW(name string, data ...byte) (int, error) {
	file, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY, ReadWrite)
	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, file.Close()
	}
	i, err := file.Write(data)
	if err != nil {
		return 0, err
	}
	return i, file.Close()
}

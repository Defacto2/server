// Package helpers are general functions shared with all parts of the web application.
package helpers

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	// Eraseline is an ANSI escape control to erase the active line of the terminal.
	Eraseline = "\x1b[2K"
	// byteUnits is a list of units used for formatting byte sizes.
	byteUnits = "kMGTPE"
)

// ByteCount formats b as in a compact, human-readable unit of measure.
func ByteCount(b int64) string {
	// source: https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d%s", b, strings.Repeat("B", 1))
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.0f%c",
		float64(b)/float64(div), byteUnits[exp])
}

// ByteCountFloat formats b as in a human-readable unit of measure.
// Units measured in gigabytes or larger are returned with 1 decimal place.
func ByteCountFloat(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d bytes", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	const gigabyte = 2
	if exp < gigabyte {
		return fmt.Sprintf("%.0f %cB",
			float64(b)/float64(div), byteUnits[exp])
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), byteUnits[exp])
}

// LastChr returns the last character or rune of the string.
func LastChr(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	r, _ := utf8.DecodeLastRuneInString(s)
	return string(r)
}

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

// Sentence formats the first letter of a string to use a capital character.
func Sentence(s string) string {
	if s == "" {
		return ""
	}
	caser := cases.Title(language.English)
	if len(s) == 1 {
		return caser.String(s)
	}
	x := strings.Split(s, " ")
	return fmt.Sprintf("%s %s", caser.String(x[0]), strings.Join(x[1:], " "))
}

// Slug returns a URL friendly string of the named group.
func Slug(name string) string {
	s := name
	// hyphen to underscore
	re := regexp.MustCompile(`\-`)
	s = re.ReplaceAllString(s, "_")
	// multiple groups get separated with asterisk
	re = regexp.MustCompile(`\, `)
	s = re.ReplaceAllString(s, "*")
	// any & characters need replacement due to HTML escaping
	re = regexp.MustCompile(` \& `)
	s = re.ReplaceAllString(s, " ampersand ")
	// numbers receive a leading hyphen
	re = regexp.MustCompile(` ([0-9])`)
	s = re.ReplaceAllString(s, "-$1")
	// delete all other characters
	const deleteAllExcept = `[^A-Za-z0-9 \-\+\.\_\*]`
	re = regexp.MustCompile(deleteAllExcept)
	s = re.ReplaceAllString(s, "")
	// trim whitespace and replace any space separators with hyphens
	s = strings.TrimSpace(strings.ToLower(s))
	re = regexp.MustCompile(` `)
	s = re.ReplaceAllString(s, "-")
	return s
}

// ShortMonth takes a month integer and abbreviates it to a three letter English month.
func ShortMonth(month int) string {
	if month < 1 || month > 12 {
		return ""
	}
	const abbreviated = 3
	s := fmt.Sprint(time.Month(month))
	if len(s) >= abbreviated {
		return s[0:abbreviated]
	}
	return ""
}

// TruncFilename reduces a filename to the length of w characters.
// The file extension is always preserved with the truncation.
func TruncFilename(w int, name string) string {
	const trunc = "."
	if w == 0 {
		return ""
	}
	l := len(name)
	if w >= l {
		return name
	}
	ext := filepath.Ext(name)
	if w <= len(ext) {
		return ext
	}
	s := name[0 : w-len(ext)-len(trunc)]
	return fmt.Sprintf("%s%s%s", s, trunc, ext)
}

// TrimPunct removes any trailing, common punctuation characters from the string.
func TrimPunct(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	rs := []rune(s)
	for i := len(rs) - 1; i >= 0; i-- {
		r := rs[i]
		// https://www.compart.com/en/unicode/category/Po
		if !unicode.Is(unicode.Po, r) {
			punctless := string(rs[0 : i+1])
			return strings.TrimSpace(punctless)
		}
	}
	return s
}

// Package helpers are general, shared functions.
package helper

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

const (
	// Eraseline is an ANSI escape control to erase the active line of the terminal.
	Eraseline = "\x1b[2K"
	// Timeout is the HTTP client timeout.
	Timeout = 5 * time.Second
	// ReadWrite is the file mode for read and write access.
	ReadWrite = 0o666
	// User-Agent to send with the HTTP request.
	UserAgent = "Defacto2 2024 app under construction (thanks!)"
	// byteUnits is a list of units used for formatting byte sizes.
	byteUnits = "kMGTPE"
)

var (
	ErrDiffLength = errors.New("files are of different lengths")
	ErrDirPath    = errors.New("directory path is a file")
	ErrExistPath  = errors.New("path ready exists and will not overwrite")
	ErrFilePath   = errors.New("file path is a directory")
	ErrKey        = errors.New("could not generate a random session key")
	ErrOSFile     = errors.New("os file is nil")
	ErrRead       = errors.New("could not read files")
)

// Add1 returns the value of a + 1.
// The type of a must be an integer type or the result is 0.
func Add1(a any) int64 {
	switch val := a.(type) {
	case int, int8, int16, int32, int64:
		i := reflect.ValueOf(val).Int()
		return i + 1
	default:
		return 0
	}
}

// CookieStore generates a key for use with the sessions cookie store middleware.
// envKey is the value of an imported environment session key. But if it is empty,
// a 32-bit randomized value is generated that changes on every restart.
//
// The effect of using a randomized key will invalidate all existing sessions on every restart.
func CookieStore(envKey string) ([]byte, error) {
	if envKey != "" {
		key := []byte(envKey)
		return key, nil
	}
	const length = 32
	key := make([]byte, length)
	n, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrKey, err.Error())
	}
	if n != length {
		return nil, ErrKey
	}
	return key, nil
}

// Day returns true if the i value can be used as a day time value.
func Day(i int) bool {
	const maxDay = 31
	if i > 0 && i <= maxDay {
		return true
	}
	return false
}

// Determine returns the encoding of the plain text byte slice.
// If the byte slice contains Unicode multi-byte characters then nil is returned.
// Otherwise a charmap.ISO8859_1 or charmap.CodePage437 encoding is returned.
func Determine(reader io.Reader) encoding.Encoding {
	if reader == nil {
		return nil
	}
	const (
		controlStart   = 0x00  // ASCII control character start
		controlEnd     = 0x1f  // ASCII control character end
		undefinedStart = 0x7f  // Latin-1 undefined characters start
		undefinedEnd   = 0x9f  // Latin-1 undefined characters end
		escape         = 0x1b  // ASCII escape control character
		unknownRune    = 65533 // Unicode replacement character (�)
	)
	// The following characters are considered whitespace characters in C
	// with the isspace function:
	// https://en.cppreference.com/w/c/string/byte/isspace
	const (
		formFeed       = '\f'
		newline        = '\n'
		carriageReturn = '\r'
		tab            = '\t'
		verticalTab    = '\v'
	)

	// KCF key qualifiers on the Commodore Amiga, see: https://wiki.amigaos.net/wiki/Keymap_Library

	const (
		kcfAltEsc = 0x9b // the Amiga had Keymap Qualifier Bits, which could be a typo to generate an Alt-Esc sequence?
		bell      = 0x07 // ASCII bell character that is sometimes found in Amiga ANSI files
		house     = 0x7f // CP-437 house character that displays a unique glyph in the Amiga Topaz font
	)

	p, err := io.ReadAll(reader)
	if err != nil {
		return nil
	}

	for _, char := range p {
		r := rune(char)
		switch {
		case char == escape:
			// escape control character commonly used in ANSI escaped sequences
			continue
		case // oddball control characters that are sometimes found in Amiga ANSI files
			char == kcfAltEsc,
			char == bell,
			char == house:
			continue
		case // common whitespace control characters
			char == formFeed,
			char == newline,
			char == carriageReturn,
			char == tab,
			char == verticalTab:
			continue
		case r == unknownRune:
			// when an unknown extended-ASCII character (128-255) is encountered
			continue
		case char >= undefinedStart && char <= undefinedEnd:
			// unused ASCII, which we can probably assumed to be CP-437
			return charmap.CodePage437
		case char >= controlStart && char <= controlEnd:
			// ASCII control characters, which we can probably assumed to be CP-437 glyphs
			return charmap.CodePage437
		case r > unknownRune:
			// The maximum value of an 8-bit character is 255 (0xff),
			// so rune valud above that, 256+ (0x100) is a Unicode multi-byte character,
			// which we can assume to be UTF-8.
			return unicode.UTF8
		}
	}
	return sequences(p)
}

// sequences returns the encoding based on the presence of common CP-437 or ISO-8859-1 character sequences.
// Full block, medium shade, horizontal bars and half blocks are sequences of characters that are often
// unique to the CP-437 encoding.
func sequences(p []byte) encoding.Encoding {
	const (
		lowerHalfBlock = 0xdc
		upperHalfBlock = 0xdf
		doubleHorizBar = 0xcd
		singleHorizBar = 0xc4
		mediumShade    = 0xb1
		fullBlock      = 0xdb
		interpunct     = 0xfa
	)
	chars := []byte{
		lowerHalfBlock,
		upperHalfBlock,
		doubleHorizBar,
		singleHorizBar,
		mediumShade,
		fullBlock,
		interpunct,
	}
	for _, char := range chars {
		const count = 4
		subslice := bytes.Repeat([]byte{char}, count)
		if bytes.Contains(p, subslice) {
			return charmap.CodePage437
		}
	}
	guillemets := []byte{0xae, 0xaf} // «»
	if bytes.Contains(p, guillemets) {
		return charmap.CodePage437
	}
	return charmap.ISO8859_1
}

// FixSceneOrg returns a working URL if the provided rawURL is a known,
// broken link to a scene.org file. Otherwise it returns the original URL.
func FixSceneOrg(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	if u.Host == "scene.org" && u.Path == "/file.php" {
		return rawURL
	}
	if u.Host == "files.scene.org" {
		p := u.Path
		x := strings.Split(p, "/")
		if len(x) > 0 && x[1] == "view" {
			x[1] = "get"
			newURL := &url.URL{
				Scheme: "https",
				Host:   "files.scene.org",
				Path:   strings.Join(x, "/"),
			}
			return newURL.String()
		}
	}
	return rawURL
}

// DownloadResponse contains the details of a downloaded file.
type DownloadResponse struct {
	ContentLength string // ContentLength is the size of the file in bytes.
	ContentType   string // ContentType is the MIME type of the file.
	LastModified  string // LastModified is the last modified date of the file.
	Path          string // Path is the path to the downloaded file.
}

// GetFile downloads a file from a remote URL and saves it to the default temp directory.
// It returns the path to the downloaded file.
func GetFile(url string) (DownloadResponse, error) {
	url = FixSceneOrg(url)

	// Get the remote file
	client := http.Client{
		Timeout: Timeout,
	}
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return DownloadResponse{}, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	req.Header.Set("User-Agent", UserAgent)
	res, err := client.Do(req)
	if err != nil {
		return DownloadResponse{}, fmt.Errorf("client.Do: %w", err)
	}
	defer res.Body.Close()

	dlr := DownloadResponse{
		ContentLength: res.Header.Get("Content-Length"),
		ContentType:   res.Header.Get("Content-Type"),
		LastModified:  res.Header.Get("Last-Modified"),
	}
	// Create the file in the default temp directory
	tmpFile, err := os.CreateTemp("", "downloadfile-*")
	if err != nil {
		return DownloadResponse{}, fmt.Errorf("os.CreateTemp: %w", err)
	}
	defer tmpFile.Close()

	// Write the body to file
	if _, err := io.Copy(tmpFile, res.Body); err != nil {
		defer os.Remove(tmpFile.Name())
		return DownloadResponse{}, fmt.Errorf("io.Copy: %w", err)
	}
	dlr.Path = tmpFile.Name()
	return dlr, nil
}

// GetStat returns the content length of a remote URL.
// It returns an error if the URL is invalid or the request fails.
// The content length is -1 if it is unknown.
func GetStat(url string) (int64, error) {
	const unknown = -1
	url = FixSceneOrg(url)
	client := http.Client{
		Timeout: Timeout,
	}
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return unknown, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	req.Header.Set("User-Agent", UserAgent)
	res, err := client.Do(req)
	if err != nil {
		return unknown, fmt.Errorf("client.Do: %w", err)
	}
	defer res.Body.Close()
	return res.ContentLength, nil
}

// Latency returns the stored, current local time.
func Latency() *time.Time {
	start := time.Now()
	r := new(big.Int)
	const n, k = 1000, 10
	r.Binomial(n, k)
	return &start
}

// LocalIPs returns a list of local IP addresses.
// credit: https://gosamples.dev/local-ip-address/
func LocalIPs() ([]net.IP, error) {
	var ips []net.IP
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("net.InterfaceAddrs: %w", err)
	}

	for _, addr := range addresses {
		if ipnet, ipnetExists := addr.(*net.IPNet); ipnetExists && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP)
			}
		}
	}
	return ips, nil
}

// LocalHosts returns a list of local hostnames.
func LocalHosts() ([]string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("os.Hostname: %w", err)
	}
	hosts := []string{}
	hosts = append(hosts, hostname)
	// confirm localhost is resolvable
	if _, err = net.LookupHost("localhost"); err != nil {
		return nil, fmt.Errorf("net.LookupHost: %w", err)
	}
	hosts = append(hosts, "localhost")
	return hosts, nil
}

// Ping sends a HTTP GET request to the provided URI and returns the status code and size of the response.
func Ping(uri string) (int, int64, error) {
	client := http.Client{
		Timeout: Timeout,
	}
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return http.StatusInternalServerError, 0, fmt.Errorf("helper ping new request %w: %s", err, uri)
	}
	req.Header.Set("User-Agent", UserAgent)
	res, err := client.Do(req)
	if err != nil {
		return http.StatusInternalServerError, 0, fmt.Errorf("helper ping client do %w: %s", err, uri)
	}
	defer res.Body.Close()
	size, err := io.Copy(io.Discard, res.Body)
	if err != nil {
		return http.StatusInternalServerError, 0, fmt.Errorf("helper ping body copy %w: %s", err, uri)
	}
	return res.StatusCode, size, nil
}

// LocalHostPing sends a HTTP GET request to the provided URI on the localhost
// and returns the status code and size of the response.
func LocalHostPing(uri string, proto string, port int) (int, int64, error) {
	if _, err := net.LookupHost("localhost"); err != nil {
		return http.StatusInternalServerError, 0, fmt.Errorf("helper localhost ping lookup %w", err)
	}
	url := fmt.Sprintf("%s://localhost:%d%s", proto, port, uri)
	return Ping(url)
}

// TimeDistance describes the difference between two time values.
// The seconds parameter determines if the string should include seconds.
func TimeDistance(from, to time.Time, seconds bool) string {
	// This function is a port of a CFWheels framework function programmed in ColdFusion (CFML).
	// https://github.com/cfwheels/cfwheels/blob/cf8e6da4b9a216b642862e7205345dd5fca34b54/wheels/global/misc.cfm#L112

	delta := to.Sub(from)
	secs, mins, hrs := int(delta.Seconds()),
		int(delta.Minutes()),
		int(delta.Hours())

	const hours, days, months, year, years, twoyears = 1440, 43200, 525600, 657000, 919800, 1051200
	switch {
	case mins <= 1:
		if !seconds {
			return lessMin(secs)
		}
		return lessMinAsSec(secs)
	case mins < hours:
		return lessHours(mins, hrs)
	case mins < days:
		return lessDays(mins, hrs)
	case mins < months:
		return lessMonths(mins, hrs)
	case mins < year:
		return "about 1 year"
	case mins < years:
		return "over 1 year"
	case mins < twoyears:
		return "almost 2 years"
	default:
		y := mins / months

		return fmt.Sprintf("%d years", y)
	}
}

// lessMin returns a string describing the time difference in seconds or minutes.
func lessMin(secs int) string {
	const minute = 60
	switch {
	case secs < minute:
		return "less than a minute"
	default:
		return "1 minute"
	}
}

// lessMinAsSec returns a string describing the time difference in seconds.
func lessMinAsSec(secs int) string {
	const five, ten, twenty, forty = 5, 10, 20, 40
	switch {
	case secs < five:
		return "less than 5 seconds"
	case secs < ten:
		return "less than 10 seconds"
	case secs < twenty:
		return "less than 20 seconds"
	case secs < forty:
		return "half a minute"
	default:
		return "1 minute"
	}
}

// lessHours returns a string describing the time difference in hours.
func lessHours(mins, hrs int) string {
	const parthour, abouthour, hours = 45, 90, 1440

	switch {
	case mins < parthour:
		return fmt.Sprintf("%d minutes", mins)
	case mins < abouthour:
		return "about 1 hour"
	case mins < hours:
		return fmt.Sprintf("about %d hours", hrs)
	default:
		return ""
	}
}

// lessDays returns a string describing the time difference in days.
func lessDays(mins, hrs int) string {
	const day, days = 2880, 43200
	switch {
	case mins < day:
		return "1 day"
	case mins < days:
		const hoursinaday = 24
		d := hrs / hoursinaday
		return fmt.Sprintf("%d days", d)
	default:
		return ""
	}
}

// lessMonths returns a string describing the time difference in months.
func lessMonths(mins, hrs int) string {
	const month, months = 86400, 525600
	switch {
	case mins < month:
		return "about 1 month"
	case mins < months:
		const hoursinamonth = 730
		m := hrs / hoursinamonth
		return fmt.Sprintf("%d months", m)
	default:
		return ""
	}
}

// Year returns true if the i value is greater than 1969
// or equal to the current year.
func Year(i int) bool {
	const unix = 1970
	now := time.Now().Year()
	if i >= unix && i <= now {
		return true
	}
	return false
}

// Package str provides functions for handling string input data.
package str

import (
	"errors"
	"fmt"
	"html/template"
	"image"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/handler/app/internal/exts"
	"github.com/Defacto2/server/internal/helper"
	"github.com/Defacto2/server/internal/magicnumber"
	"github.com/dustin/go-humanize"
	"github.com/h2non/filetype"
	"github.com/volatiletech/null/v8"
)

var (
	ErrLinkType = errors.New("the id value is an invalid type")
	ErrNegative = errors.New("value cannot be a negative number")
)

const (
	textamiga = "textamiga"
	typeErr   = "error: received an invalid type to "
)

// ContentSRC returns the destination directory for the extracted archive content.
// The directory is created if it does not exist. The directory is named after the source file.
func ContentSRC(src string) (string, error) {
	name := strings.TrimSpace(strings.ToLower(filepath.Base(src)))
	dir := filepath.Join(os.TempDir(), "defacto2-server")

	pattern := "artifact-content-" + name
	dst := filepath.Join(dir, pattern)
	if st, err := os.Stat(dst); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dst, os.ModePerm); err != nil {
				return "", err
			}
			return dst, nil
		}
		return dst, nil
	} else if !st.IsDir() {
		return "", fmt.Errorf("error, not a directory: %s", dir)
	}
	return dst, nil
}

// ImageXY returns the image file size and dimensions.
func ImageXY(name string) [2]string {
	switch filepath.Ext(strings.ToLower(name)) {
	case ".jpg", ".jpeg", ".gif", ".png", ".webp":
	default:
		st, err := os.Stat(name)
		if err != nil {
			return [2]string{err.Error(), ""}
		}
		return [2]string{humanize.Comma(st.Size()), ""}
	}
	reader, err := os.Open(name)
	if err != nil {
		return [2]string{err.Error(), ""}
	}
	defer reader.Close()
	st, err := reader.Stat()
	if err != nil {
		return [2]string{err.Error(), ""}
	}
	config, _, err := image.DecodeConfig(reader)
	if err != nil {
		return [2]string{err.Error(), ""}
	}
	return [2]string{humanize.Comma(st.Size()), fmt.Sprintf("%dx%d", config.Width, config.Height)}
}

// LinkID creates a URL to link to the record.
// The id is obfuscated to prevent direct linking.
// The elem is the element to link to, such as 'f' for file or 'd' for download.
func LinkID(id any, elem string) (string, error) {
	var i int64
	switch val := id.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		i = reflect.ValueOf(val).Int()
		if i <= 0 {
			return "", fmt.Errorf("app link id %w: %d", ErrNegative, i)
		}
	default:
		return "", fmt.Errorf("app link id %w: %s", ErrLinkType, reflect.TypeOf(id).String())
	}
	href, err := url.JoinPath("/", elem, helper.ObfuscateID(i))
	if err != nil {
		return "", fmt.Errorf("app link id %d could not be made into a valid url: %w", i, err)
	}
	return href, nil
}

// LinkPreviewTip returns a tooltip to describe the preview link.
func LinkPreviewTip(name, platform string) string {
	if name == "" {
		return ""
	}
	platform = strings.TrimSpace(platform)
	ext := strings.ToLower(filepath.Ext(name))
	switch {
	case slices.Contains(exts.Archives(), ext):
		// this case must always be first
		return ""
	case platform == textamiga, platform == "text":
		return "Read this as text"
	case slices.Contains(exts.Documents(), ext):
		return "Read this as text"
	case slices.Contains(exts.Images(), ext):
		return "View this as an image or photo"
	case slices.Contains(exts.Media(), ext):
		return "Play this as media"
	}
	return ""
}

// LinkRelr returns a link to the named group page.
func LinkRelr(name string) (string, error) {
	href, err := url.JoinPath("/", "g", helper.Slug(name))
	if err != nil {
		return "", fmt.Errorf("name %q could not be made into a valid url: %w", name, err)
	}
	return href, nil
}

func MakeLink(name, class string, performant bool) (string, error) {
	ref, err := LinkRelr(name)
	if err != nil {
		return "", fmt.Errorf("app make link %w", err)
	}
	x := helper.Capitalize(strings.ToLower(name))
	title := x
	if !performant {
		title = releaser.Link(helper.Slug(name))
	}
	s := fmt.Sprintf(`<a class="%s" href="%s">%s</a>`, class, ref, title)
	if x != "" && title == "" {
		s = "error: could not link group"
	}
	return s, nil
}

func MagicAsTitle(name string) string {
	r, err := os.Open(name)
	if err != nil {
		return err.Error()
	}
	defer r.Close()
	sign, err := magicnumber.Find(r)
	if err != nil {
		return err.Error()
	}
	return sign.Title()
}

// MIME returns the MIME type for the file record.
func MIME(name string) string {
	file, err := os.Open(name)
	if err != nil {
		return err.Error()
	}
	defer file.Close()

	const sample = 512
	head := make([]byte, sample)
	_, err = file.Read(head)
	if err != nil {
		return err.Error()
	}

	kind, err := filetype.Match(head)
	if err != nil {
		return err.Error()
	}
	if kind != filetype.Unknown {
		return kind.MIME.Value
	}

	return http.DetectContentType(head)
}

// Releasers returns a HTML links for the primary and secondary group names.
func Releasers(prime, second string) template.HTML {
	var s string
	switch {
	case prime != "" && second != "":
		s = fmt.Sprintf("%s <strong>+</strong><br>%s", prime, second)
	case prime != "":
		s = prime
	case second != "":
		s = second
	default:
		return ""
	}
	return template.HTML(s)
}

func ReleaserPair(a, b any) [2]string {
	av, bv := "", ""
	switch val := a.(type) {
	case string:
		av = reflect.ValueOf(val).String()
	case null.String:
		if val.Valid {
			av = val.String
		}
	}
	switch val := b.(type) {
	case string:
		bv = reflect.ValueOf(val).String()
	case null.String:
		if val.Valid {
			bv = val.String
		}
	}
	av = strings.TrimSpace(av)
	bv = strings.TrimSpace(bv)
	switch {
	case av != "" && bv != "":
		return [2]string{av, bv}
	case bv != "":
		return [2]string{bv, ""}
	case av != "":
		return [2]string{av, ""}
	}
	return [2]string{}
}

// StatHumanize returns the last modified date, size in bytes and size formatted
// of the named file.
func StatHumanize(name string) (string, string, string) {
	stat, err := os.Stat(name)
	if err != nil {
		return "", "", err.Error()
	}
	return stat.ModTime().Format("2006-Jan-02"),
		humanize.Comma(stat.Size()),
		humanize.Bytes(uint64(stat.Size()))
}

// Updated returns a string of the time since the given time t.
// The time is formatted as "Last updated 1 hour ago".
// If the time is not valid, an empty string is returned.
func Updated(t any, s string) string {
	if t == nil {
		return ""
	}
	if s == "" {
		s = "Time"
	}
	switch val := t.(type) {
	case null.Time:
		if !val.Valid {
			return ""
		}
		return fmt.Sprintf("%s %s ago", s, helper.TimeDistance(val.Time, time.Now(), true))
	case time.Time:
		return fmt.Sprintf("%s %s ago", s, helper.TimeDistance(val, time.Now(), true))
	default:
		return fmt.Sprintf("%supdated: %s", typeErr, reflect.TypeOf(t).String())
	}
}

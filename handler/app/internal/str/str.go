package str

import (
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/handler/app/internal/exts"
	"github.com/Defacto2/server/internal/helper"
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

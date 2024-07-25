package app

// Package file dirs.go contains the artifact page directories and handlers.

import (
	"fmt"
	"html/template"
	_ "image/gif"  // gif format decoder
	_ "image/jpeg" // jpeg format decoder
	_ "image/png"  // png format decoder
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/Defacto2/server/handler/app/internal/str"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/dustin/go-humanize"
	"github.com/labstack/echo/v4"
	"github.com/volatiletech/null/v8"
	_ "golang.org/x/image/webp" // webp format decoder
)

type extract int // extract target format for the file archive extractor

const (
	picture  extract = iota // extract a picture or image
	ansitext                // extract ansilove compatible text
)

// Artifact404 renders the error page for the artifact links.
func Artifact404(c echo.Context, id string) error {
	const name = "status"
	if c == nil {
		return InternalErr(c, name, errorWithID(ErrCxt, id, nil))
	}
	data := empty(c)
	data["title"] = fmt.Sprintf("%d error, artifact page not found", http.StatusNotFound)
	data["description"] = fmt.Sprintf("HTTP status %d error", http.StatusNotFound)
	data["code"] = http.StatusNotFound
	data["logo"] = "Artifact not found"
	data["alert"] = fmt.Sprintf("Artifact %q cannot be found", strings.ToLower(id))
	data["probl"] = "The artifact page does not exist, there is probably a typo with the URL."
	data["uriOkay"] = "f/"
	data["uriErr"] = id
	err := c.Render(http.StatusNotFound, name, data)
	if err != nil {
		return InternalErr(c, name, errorWithID(err, id, nil))
	}
	return nil
}

// errorWithID returns an error with the artifact ID appended to the error message.
// The key string is expected any will always be displayed in the error message.
// The id can be an integer or string value and should be the database numeric ID.
func errorWithID(err error, key string, id any) error {
	if err == nil {
		return nil
	}
	key = strings.TrimSpace(key)
	const cause = "caused by artifact"
	switch id.(type) {
	case int, int64:
		return fmt.Errorf("%w: %s %s (%d)", err, cause, key, id)
	case string:
		return fmt.Errorf("%w: %s %s (%s)", err, cause, key, id)
	default:
		return fmt.Errorf("%w: %s %s", err, cause, key)
	}
}

// Dirs contains the directories used by the artifact pages.
type Dirs struct {
	Download  string // path to the artifact download directory
	Preview   string // path to the preview and screenshot directory
	Thumbnail string // path to the file thumbnail directory
	Extra     string // path to the extra files directory
	URI       string // the URI of the file record
}

func decode(src io.Reader) (string, error) {
	out := strings.Builder{}
	if _, err := io.Copy(&out, src); err != nil {
		return "", fmt.Errorf("io.Copy: %w", err)
	}
	if !strings.HasSuffix(out.String(), "\n\n") {
		out.WriteString("\n")
	}
	return out.String(), nil
}

// dirsBytes returns the file size for the file record.
func dirsBytes(i int64) string {
	if i == 0 {
		return "(n/a)"
	}
	return humanize.Bytes(uint64(i))
}

///===========================================================================
/// KEEEP BELOW as htmx pkg requires these functions

// FirstLead returns the lead for the file record which is the filename and releasers.
func FirstLead(art *models.File) string {
	fname := art.Filename.String
	span := fmt.Sprintf("<span class=\"font-monospace fs-6 fw-light\">%s</span> ", fname)
	rels := string(LinkRels(art.GroupBrandBy, art.GroupBrandFor))
	return fmt.Sprintf("%s<br>%s", rels, span)
}

func GroupReleasers(art *models.File) string {
	if art == nil {
		return ""
	}
	return string(LinkRels(art.GroupBrandBy, art.GroupBrandFor))
}

// LinkRelrs returns the groups associated with a release and a link to each group.
// The performant flag will use the group name instead of the much slower group slug formatter.
func LinkRelrs(performant bool, a, b any) template.HTML {
	const class = "text-nowrap link-offset-2 link-underline link-underline-opacity-25"
	var av, bv string
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

	av, bv = strings.TrimSpace(av), strings.TrimSpace(bv)
	if av == "" && bv != "" {
		av = bv
		bv = ""
	}

	var prime, second string
	var err error
	if av != "" {
		prime, err = str.MakeLink(av, class, performant)
		if err != nil {
			return template.HTML(fmt.Sprintf("error: %s", err))
		}
	}
	if bv != "" {
		second, err = str.MakeLink(bv, class, performant)
		if err != nil {
			return template.HTML(fmt.Sprintf("error: %s", err))
		}
	}
	return str.Releasers(prime, second)
}

// LinkRelrs returns the groups associated with a release and a link to each group.
func LinkRels(a, b any) template.HTML {
	if a == nil || b == nil {
		return ""
	}
	return LinkRelrs(false, a, b)
}

// LinkRelFast returns the groups associated with a release and a link to each group.
// It is a faster version of LinkRelrs and should be used with the templates that have large lists of group names.
func LinkRelFast(a, b any) template.HTML {
	if a == nil || b == nil {
		return ""
	}
	return LinkRelrs(true, a, b)
}

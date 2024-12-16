package tidbit

import (
	"html/template"
	"slices"
	"strings"

	"github.com/Defacto2/releaser"
)

// URI is a the URL slug of the releaser.
type URI string

// ID is the identifier of the tidbit.
type ID int

// String returns the tidbit description.
func (id ID) String() string {
	if s := tidbits[id]; s != "" {
		return s
	}
	return ""
}

// URI returns the URIs of the tidbit.
func (id ID) URI() []URI {
	if x := groups[id]; x != nil {
		return x
	}
	return nil
}

// URL returns the HTML links of the tidbit but the provided URI is excluded.
func (id ID) URL(uri string) template.HTML {
	if id == -1 {
		return template.HTML("")
	}
	urls := id.URI()
	slices.Sort(urls)
	html := []string{}
	for _, u := range urls {
		if u == URI(uri) {
			continue
		}
		s := string(u)
		html = append(html, `<a href="/g/`+s+`">`+releaser.Link(s)+`</a>`)
	}
	s := strings.Join(html, " &nbsp; ")
	return template.HTML(s)
}

// Tibits is a map of tidbits mapped to their URIs.
type Tibits map[ID][]URI

// Tidbit is a map of tidbits mapped to their descriptions.
type Tidbit map[ID]string

var groups = Tibits{
	1: []URI{"untouchables", "the-untouchables"},
	2: []URI{"five-o", "boys-from-company-c", "the-firm"},
}

var tidbits = Tidbit{
	1: "Untouchables were a famed US based game release group. The Untouchables were a 1990s scene group from Norway.",
	2: "Five-O and the BCC were a US based game release groups that merged in December 1988, the next month in January they changed their name to The Firm.",
}

// Find returns the tidbit ID for the given URI.
// If the URI is not found, -1 is returned.
//
// The ID returned can be used in a string conversion to get the description.
// The ID can also be used to get the URIs of the tidbit.
func Find(uri string) ID {
	for id, uris := range groups {
		for _, u := range uris {
			if u == URI(uri) {
				return id
			}
		}
	}
	return -1
}

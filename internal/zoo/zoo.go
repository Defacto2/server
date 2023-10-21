// Package zoo provides data about releasers and groups on the Demozoo website.
// https://demozoo.org
package zoo

// URI is a the URL slug of the releaser.
type URI string

// GroupID is the Demozoo ID of the group.
type GroupID uint

// Groups is a map of releasers URIs mapped to their Demozoo IDs.
type Groups map[URI]GroupID

var groups = Groups{
	"acid-productions":                  7647,
	"class":                             16508,
	"defacto2":                          10000,
	"fairlight":                         239,
	"international-network-of-crackers": 12175,
	"insane-creators-enterprise":        2169,
	"mirage":                            45887,
	"paradigm":                          26612,
	"razor-1911":                        519,
	"silicon-dream-artists":             25795,
	"superior-art-creations":            7050,
	"the-dream-team":                    20609,
	"the-humble-guys":                   7421,
	"the-silents":                       101,
	"tristar-ampersand-red-sector-inc":  69,
}

// Find returns the Demozoo group ID for the given uri.
// It returns 0 if the uri is not known.
func Find(uri string) GroupID {
	if _, ok := groups[URI(uri)]; ok {
		return groups[URI(uri)]
	}
	return 0
}

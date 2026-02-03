// Package janeway provides data about releasers and groups from the Janeway Amiga Scene [website].
// [website]: https://janeway.exotica.org.uk/
package janeway

// URI is the URL slug of the releaser.
type URI string

// GroupID is the group id.
type GroupID int

// Groups is a map of releasers URIs mapped to their janeway author/group id.
type Groups map[string]GroupID

var groups = Groups{
	"alpha-flight":                     2276,
	"dynamix":                          23611,
	"dynasty":                          20956,
	"dytec":                            13810,
	"fairlight":                        252,
	"faith":                            11638,
	"hoodlum":                          11127,
	"paradox":                          656,
	"prestige":                         9256,
	"prophecy":                         651,
	"quartex":                          48,
	"razor-1911":                       338,
	"rebels":                           479,
	"red-sector-inc":                   222,
	"scoopex":                          210,
	"skid-row":                         4277,
	"skillion":                         12175,
	"the-silents":                      39,
	"triad":                            1545,
	"tristar-ampersand-red-sector-inc": 968,
}

// Find returns the janeway author/group id for the given releaser uri.
// If the releaser is not found, an empty value is returned.
func Find(uri string) GroupID {
	return groups[uri]
}

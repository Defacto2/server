// Package janeway provides data about releasers and groups from the Janeway Amiga Scene [website].
// [website]: https://janeway.exotica.org.uk/
package janeway

// URI is the URL slug of the releaser.
type URI string

// GroupID is the group id.
type GroupID int

// Groups is a map of releasers URIs mapped to their janeway author/group id.
type Groups map[string]GroupID

//nolint:gochecknoglobals
var groups = Groups{
	"abandon":                          2327,
	"accumulators":                     5964,
	"alpha-flight":                     2276,
	"anthrox":                          564,
	"backlash":                         10245,
	"censor-design":                    1522,
	"classic":                          469,
	"cocaine":                          42905,
	"crack-inc":                        20587,
	"crystal":                          261,
	"d_tect":                           834,
	"desire":                           3098,
	"dynamix":                          23611,
	"dynasty":                          20956,
	"dytec":                            13810,
	"fairlight":                        252,
	"faith":                            11638,
	"fila":                             64273,
	"hoodlum":                          11127,
	"image-console":                    3460,
	"interpol":                         29691,
	"legend":                           358,
	"lightforce":                       3649,
	"mystic":                           9025,
	"nightfall":                        13152,
	"outlaws":                          3849,
	"paradox":                          656,
	"paranoimia":                       251,
	"prestige":                         9256,
	"prodigy":                          14491,
	"prophecy":                         651,
	"quartex":                          48,
	"razor-1911":                       338,
	"rebels":                           479,
	"red-sector-inc":                   222,
	"scoopex":                          210,
	"shining-8":                        4155,
	"skid-row":                         4277,
	"skillion":                         12175,
	"subzero":                          20921,
	"the-silents":                      39,
	"the-fast-guys":                    67939,
	"the-flame-arrows":                 11758,
	"the-organized-crime":              5441,
	"therapy":                          8850,
	"triad":                            1545,
	"tristar-ampersand-red-sector-inc": 968,
	"thunderloop":                      31805,
	"vision":                           308,
}

// Find returns the janeway author/group id for the given releaser uri.
// If the releaser is not found, an empty value is returned.
func Find(uri string) GroupID {
	return groups[uri]
}

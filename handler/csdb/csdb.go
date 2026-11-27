// Package csdb provides data about releasers and groups on The C-64 Scene Database [website].
// [website]: https://csdb.dk
//
//nolint:mnd,gochecknoglobals
package csdb

// URI is the URL slug of the releaser.
type URI string

// GroupID is the group id.
type GroupID int

// Groups is a map of releaser URIs mapped to their csdb group id.
type Groups map[URI]GroupID

var groups = Groups{
	"2000ad":                           261,
	"alpha-flight":                     215,
	"censor-design":                    2310,
	"desire":                           6614,
	"dynamix":                          751,
	"dytec":                            257,
	"fairlight":                        20,
	"fantastic-4-cracking-group":       148,
	"legend":                           176,
	"lightforce":                       3105,
	"motiv8":                           182,
	"nightfall":                        5814,
	"quartex":                          2396,
	"razor-1911":                       431,
	"rebels":                           10411,
	"red-sector-inc":                   602,
	"shining-8":                        761,
	"skid-row":                         2784,
	"the-silents":                      1099,
	"triad":                            132,
	"tristar-ampersand-red-sector-inc": 915,
	"unitrax":                          3084,
	"vision":                           510,
}

// Find returns the csdb group id for the given releaser uri.
// If the releaser is not found, 0 is returned.
func Find(uri string) GroupID {
	return groups[URI(uri)]
}

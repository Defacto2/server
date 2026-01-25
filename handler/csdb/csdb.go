// Package csdb provides data about releasers and groups on The C-64 Scene Database [website].
// [website]: https://csdb.dk
package csdb

// URI is the URL slug of the releaser.
type URI string

// GroupID is the group id.
type GroupID int

// Groups is a map of releasers URIs mapped to their csdb group id.
type Groups map[string]GroupID

func groups() Groups {
	return Groups{
		"2000ad":                           261,
		"alpha-flight":                     215,
		"dynamix":                          751,
		"dytec":                            257,
		"fairlight":                        20,
		"fantastic-4-cracking-group":       148,
		"motiv8":                           182,
		"quartex":                          2396,
		"razor-1911":                       431,
		"rebels":                           10411,
		"red-sector-inc":                   602,
		"skid-row":                         2784,
		"the-silents":                      1099,
		"triad":                            132,
		"tristar-ampersand-red-sector-inc": 915,
	}
}

// Find returns the csdb group id for the given releaser uri.
// If the releaser is not found, an empty value is returned.
func Find(uri string) GroupID {
	return groups()[uri]
}

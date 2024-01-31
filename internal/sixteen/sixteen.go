// Package sixteen provides data about releasers and groups on the 16colors website.
// [16colors]: https://16colo.rs
package sixteen

// URI is a the URL slug of the releaser.
type URI string

// Grouptag is the 16colors group tag.
type GroupTag string

// Groups is a map of releasers URIs mapped to their 16colors group tag.
type Groups map[GroupTag]GroupTag

func groups() Groups {
	return Groups{
		"acid-productions":                  "group/acid",
		"defacto2":                          "group/defacto 2",
		"international-network-of-crackers": "tags/content/inc",
		"insane-creators-enterprise":        "group/ice",
		"superior-art-creations":            "group/sac",
		"razor-1911":                        "tags/content/razor 1911",
	}
}

// Find returns the 16colors group tag for the given releaser uri.
// If the releaser is not found, an empty string is returned.
func Find(uri string) GroupTag {
	return groups()[GroupTag(uri)]
}

// Package sixteen provides data about releasers and groups on the 16colors website.
// [16colors]: https://16colo.rs
package sixteen

// URI is a the URL slug of the releaser.
type URI string

// GroupTag is the 16colors group tag.
type GroupTag string

// Groups is a map of releasers URIs mapped to their 16colors group tag.
type Groups map[GroupTag]GroupTag

func groups() Groups {
	return Groups{
		"hype":                              "pack/hype/",
		"international-network-of-crackers": "tags/content/inc",
		"razor-1911":                        "tags/content/razor 1911",
		"acid-productions":                  "group/acid",
		"defacto2":                          "group/defacto 2",
		"insane-creators-enterprise":        "group/ice",
		"superior-art-creations":            "group/sac",
		"aces-of-ansi-art":                  "group/aaa",
		"bitchin-ansi-designs":              "group/bad",
		"creators-of-intense-art":           "group/cia",
		"art-creation-enterprises":          "group/ace",
		"damn-excellent-ansi-design":        "group/dead",
		"katharsis":                         "/group/katharsis",
	}
}

// Find returns the 16colors group tag for the given releaser uri.
// If the releaser is not found, an empty string is returned.
func Find(uri string) GroupTag {
	return groups()[GroupTag(uri)]
}

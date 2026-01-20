// Package sixteen provides data about releasers and groups on the [16colors] website.
// [16colors]: https://16colo.rs
package sixteen

// URI is the URL slug of the releaser.
type URI string

// GroupTag is the 16colors group tag.
type GroupTag string

// Groups is a map of releasers URIs mapped to their 16colors group tag.
type Groups map[GroupTag]GroupTag

func groups() Groups {
	return Groups{
		"hype":                               "pack/hype/",
		"international-network-of-crackers":  "tags/content/inc",
		"razor-1911":                         "tags/content/razor 1911",
		"acid-productions":                   "group/acid",
		"defacto2":                           "group/defacto 2",
		"insane-creators-enterprise":         "group/ice",
		"superior-art-creations":             "group/sac",
		"aces-of-ansi-art":                   "group/aaa",
		"bitchin-ansi-designs":               "group/bad",
		"creators-of-intense-art":            "group/cia",
		"art-creation-enterprises":           "group/ace",
		"damn-excellent-ansi-design":         "group/dead",
		"katharsis":                          "group/katharsis",
		"artists-in-revolt":                  "group/artists%20in%20revolt",
		"hipe":                               "group/hipe",
		"licensed-to-draw":                   "group/ltd",
		"graphics-rendered-in-magnificence":  "group/grimoire",
		"nc_17":                              "group/nc-17",
		"silicon-dream-artists":              "group/sda",
		"mirage":                             "group/mirage",
		"tribe":                              "group/tribe",
		"art-creation-enterprise":            "group/ace",
		"ansi-factory":                       "group/afc",
		"relentless-pursuit-of-magnificence": "group/rpm",
	}
}

// Find returns the 16colors group tag for the given releaser uri.
// If the releaser is not found, an empty string is returned.
func Find(uri string) GroupTag {
	return groups()[GroupTag(uri)]
}

package web

// URI is the URL slug of the releaser.
type URI string

// URL is the historical URL of the releaser website.
type Website struct {
	URL        string // the website URL
	Name       string // the website name
	NotWorking bool   // the website is not working
}

// Groups is a map of releasers URIs mapped to their websites.
type Groups map[URI]Website

var groups = Groups{
	"acid-productions": Website{
		URL:  "https://www.acid.org",
		Name: "ACiD Productions",
	},
	"defacto2": Website{
		URL:  "https://defacto2.net",
		Name: "Defacto2",
	},
	"razor-1911": Website{
		URL:        "https://www.razor1911.com",
		Name:       "Razor 1911",
		NotWorking: true,
	},
	"razor-1911-demo": Website{
		URL:        "https://www.razor1911.com/demo",
		Name:       "Razor 1911",
		NotWorking: true,
	},
}

// Find returns the website for the given uri.
// It returns an empty string if the uri is not known.
func Find(uri string) Website {
	if _, ok := groups[URI(uri)]; ok {
		return groups[URI(uri)]
	}
	return Website{}
}

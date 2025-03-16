// Package tidbit offeres hyperlinked historical information about the Scene releasers and groups.
package tidbit

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Defacto2/releaser"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// URI is a the URL slug of the releaser.
type URI string

// ID is the identifier of the tidbit.
type ID int

const extensions = parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock

// Markdown returns the markdown content of the tidbit that is stored in the assets/md/tidbit directory.
// If the file does not exist or is empty then nil is returned.
func (id ID) Markdown() []byte {
	assets := filepath.Join("assets", "md", "tidbit")
	name := filepath.Join(assets, fmt.Sprintf("%d.md", id))
	if st, err := os.Stat(name); err != nil || st.IsDir() || st.Size() == 0 {
		return nil
	}
	b, err := os.ReadFile(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tidbit: %d.md read error: %v\n", id, err)
		return nil
	}
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(b)
	renderer := html.NewRenderer(html.RendererOptions{
		Flags: html.CommonFlags | html.HrefTargetBlank})
	return markdown.Render(doc, renderer)
}

// String returns the tidbit description.
func (id ID) String() string {
	if b := id.Markdown(); b != nil {
		return string(b)
	}
	return ""
}

// URI returns the URIs of the tidbit.
func (id ID) URI() []URI {
	if x := groups()[id]; x != nil {
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
	for val := range slices.Values(urls) {
		if val == URI(uri) {
			continue
		}
		s := string(val)
		html = append(html, `<a href="/g/`+s+`">`+releaser.Link(s)+`</a>`)
	}
	s := strings.Join(html, " &nbsp; ")
	return template.HTML(s)
}

// Tibits is a map of tidbits mapped to their URIs.
type Tibits map[ID][]URI

// Tidbit is a map of tidbits mapped to their descriptions.
type Tidbit map[ID]string

func groups() Tibits {
	return Tibits{
		1:   []URI{"untouchables", "the-untouchables"},
		2:   []URI{"five-o", "boys-from-company-c", "pirates-r-us", "the-firm"},
		3:   []URI{"fairlight", "united-software-association*fairlight"},
		400: []URI{"fairlight", "artists-in-revolt"},
		500: []URI{"fairlight", "fairlight-dox"},
		6:   []URI{"aces-of-ansi-art", "acid-productions"},
		7:   []URI{"the-duplicators"},
		8:   []URI{"pirates-club-inc"},
		9:   []URI{"against-software-protection"},
		10:  []URI{"software-pirates-inc"},
		11:  []URI{"the-illinois-pirates"},
		12:  []URI{"cracking-101", "national-elite-underground-alliance", "buck-naked-productions"},
		13:  []URI{"esp-pirates", "esp-headquarters-bbs"},
		14:  []URI{"silicon-valley-swappe-shoppe"},
		15:  []URI{"five-o", "toads"},
		16:  []URI{"c-ampersand-m", "boys-from-company-c"},
		17:  []URI{"canadian-pirates-inc", "ptl-club"},
		18:  []URI{"canadian-pirates-inc", "kgb", "ptl-club"},
		19:  []URI{"ptl-club", "sprint", "the-underground-council", "byte-bandits-bbs", "triad"},
		20:  []URI{"new-york-crackers", "miami-cracking-machine", "international-network-of-crackers"},
		21:  []URI{"public-domain"},
		22:  []URI{"bentley-sidwell-productions", "the-firm"},
		23:  []URI{"boys-from-company-c"},
		24:  []URI{"fairlight"},
		25:  []URI{"future-crew"},
		26:  []URI{"international-network-of-crackers"},
		28:  []URI{"the-firm", "mutual-assured-destruction", "public-enemy"},
		27:  []URI{"the-firm", "swat", "national-underground-application-alliance", "fairlight"},
		29:  []URI{"international-network-of-crackers", "triad"},
		30:  []URI{"cmen"},
		31:  []URI{"erkle"},
		32:  []URI{"extasy", "xerox", "fairlight"},
		33:  []URI{"norwegian-cracking-company", "international-network-of-crackers", "the-humble-guys"},
		34:  []URI{"scd_dox", "software-chronicles-digest"},
		35:  []URI{"software-chronicles-digest"},
		36:  []URI{"the-humble-guys"},
		37:  []URI{"netrunners", "minor-threat", "nexus"},
		38:  []URI{"mai-review", "sda-review", "silicon-dream-artists"},
		39:  []URI{"silicon-dream-artists"},
		40:  []URI{"hype"},
		41:  []URI{"alpha-flight", "outlaws", "storm-inc"},
		42:  []URI{"thhg"},
		43:  []URI{"tmh"},
		44:  []URI{"the-racketeers"},
		45:  []URI{"crackers-in-action"},
		46:  []URI{"legion-of-doom"},
		47:  []URI{"the-grand-council"},
		48:  []URI{"untouchables", "uniq", "xap", "pentagram"},
		49:  []URI{"italsoft"},
		50:  []URI{"future-brain-inc", "the-humble-guys"},
		51:  []URI{"pirate"},
		52:  []URI{"creators-of-intense-art", "art-creation-enterprise"},
		53:  []URI{"vla"},
		54:  []URI{"the-north-west-connection"},
		55:  []URI{"the-sysops-association-network"},
		56:  []URI{"american-pirate-industries"},
		57:  []URI{"pirates-sick-of-initials"},
		58:  []URI{"byte-bandits-bbs"},
		59:  []URI{"sorcerers"},
		60:  []URI{"katharsis"},
		61:  []URI{"national-elite-underground-alliance"},
		62:  []URI{"public-enemy", "pe*trsi*tdt", "north-american-society-of-anarchists", "red-sector-inc", "the-dream-team"},
		63:  []URI{"public-enemy"},
		64:  []URI{"razor-1911"},
		65:  []URI{"tristar-ampersand-red-sector-inc", "red-sector-inc"},
		66:  []URI{"tristar-ampersand-red-sector-inc", "pe*trsi*tdt", "the-dream-team", "skid-row", "coop"},
		67:  []URI{"tristar-ampersand-red-sector-inc"},
		68:  []URI{"the-dream-team"},
		69:  []URI{"rom-1911", "razor-1911"},
		70:  []URI{"high-society"},
		71:  []URI{"trinity-reviews", "lancelot-2"},
		72:  []URI{"real-pirates-guide"},
		73:  []URI{"the-amatuer-crackist-tutorial"},
		74:  []URI{"church-chat", "ptl-club"},
		75:  []URI{"corrupted-programming-international", "cpi-newsletter"},
		76:  []URI{"official-unprotection-scheme-library", "copycats-inc"},
		77:  []URI{"the-elementals-piratelist"},
		78:  []URI{"game-release-list"},
		79:  []URI{"gif-news"},
		80:  []URI{"hackers-unlimited", "mickey-mouse-club"},
		81:  []URI{"national-pirate-list"},
		82:  []URI{"phreakers-handbook"},
		83:  []URI{"spectrum"},
		84:  []URI{"the-pirate-world", "the-pirate-syndicate"},
		85:  []URI{"fairlight"},
	}
}

// Find returns the tidbit IDs for the given URI.
//
// The ID returned can be used in a string conversion to get the description.
// The ID can also be used to get the URIs of the tidbit.
func Find(uri string) []ID {
	ids := []ID{}
	for id, uris := range groups() {
		for val := range slices.Values(uris) {
			if val == URI(uri) {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

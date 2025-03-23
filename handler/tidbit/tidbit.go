// Package tidbit offeres hyperlinked historical information about the Scene releasers and groups.
package tidbit

import (
	"embed"
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

const extensions = parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock | parser.Footnotes

// Markdown returns the markdown content of the tidbit that is stored in the directory
// in the provided file system. If the file does not exist or is empty then nil is returned.
//
// Generally the String method should be used to get the description of the tidbit instead
// of this Markdown method.
func (id ID) Markdown(fs embed.FS, dir string) []byte {
	name := filepath.Join(dir, fmt.Sprintf("%d.md", id))
	b, err := fs.ReadFile(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tidbit: %d.md read error: %v\n", id, err)
		return nil
	}
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(b)
	renderer := html.NewRenderer(html.RendererOptions{
		Flags: html.CommonFlags | html.HrefTargetBlank,
	})
	return markdown.Render(doc, renderer)
}

// String returns the tidbit description that is stored as a markdown file in the provided file system.
func (id ID) String(fs embed.FS) string {
	if b := id.Markdown(fs, "public/md/tidbit"); b != nil {
		return string(b)
	}
	return ""
}

// URI returns the URIs of the tidbit.
func (id ID) URI() []URI {
	if x := Groups()[id]; x != nil {
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

// Groups returns the tidbit IDs and their matching URIs.
func Groups() Tibits {
	return Tibits{
		1:    []URI{"untouchables", "the-untouchables"},
		1111: []URI{"the-racketeers", "digital-gang", "strata_crackers", "usalliance", "byt"},
		111:  []URI{"the-firm"},
		2:    []URI{"five-o"},
		200:  []URI{"five-o", "boys-from-company-c", "pirates-r-us", "the-firm"},
		3:    []URI{"fairlight", "united-software-association*fairlight"},
		4:    []URI{"the-zapper"},
		5:    []URI{"ipl"},
		400:  []URI{"fairlight", "artists-in-revolt"},
		500:  []URI{"fairlight", "fairlight-dox"},
		6:    []URI{"aces-of-ansi-art", "acid-productions"},
		7:    []URI{"the-duplicators"},
		8:    []URI{"pirates-club-inc"},
		9:    []URI{"against-software-protection"},
		10:   []URI{"software-pirates-inc"},
		11:   []URI{"the-illinois-pirates"},
		12:   []URI{"cracking-101", "national-elite-underground-alliance", "buck-naked-productions"},
		13:   []URI{"esp-pirates", "esp-headquarters-bbs"},
		14:   []URI{"silicon-valley-swappe-shoppe"},
		15:   []URI{"toads"},
		515:  []URI{"five-o", "toads"},
		16:   []URI{"c-ampersand-m"},
		17:   []URI{"the-billionarre-boys-club"},
		18:   []URI{"canadian-pirates-inc", "kgb", "ptl-club"},
		19:   []URI{"the-underground-council"},
		199:  []URI{"ptl-club", "sprint", "the-underground-council", "byte-bandits-bbs", "triad"},
		20:   []URI{"new-york-crackers", "miami-cracking-machine", "international-network-of-crackers"},
		201:  []URI{"miami-cracking-machine"},
		202:  []URI{"new-york-crackers"},
		21:   []URI{"public-domain"},
		22:   []URI{"bentley-sidwell-productions"},
		23:   []URI{"boys-from-company-c"},
		24:   []URI{"fairlight"},
		25:   []URI{"future-crew"},
		26:   []URI{"international-network-of-crackers"},
		27:   []URI{"vortex-software"},
		28:   []URI{"opyright-infiltration-agency"},
		2700: []URI{"the-firm", "swat", "national-underground-application-alliance", "fairlight"},
		2800: []URI{"the-firm", "mutual-assured-destruction", "public-enemy"},
		29:   []URI{"big-brother"},
		30:   []URI{"cmen"},
		31:   []URI{"erkle"},
		32:   []URI{"extasy", "xerox", "fairlight"},
		333:  []URI{"norwegian-cracking-company", "international-network-of-crackers", "the-humble-guys"},
		34:   []URI{"scd_dox", "software-chronicles-digest"},
		35:   []URI{"software-chronicles-digest"},
		36:   []URI{"the-humble-guys"},
		37:   []URI{"netrunners", "minor-threat", "nexus"},
		38:   []URI{"mai-review", "sda-review", "silicon-dream-artists"},
		39:   []URI{"silicon-dream-artists"},
		40:   []URI{"hype"},
		41:   []URI{"alpha-flight", "outlaws", "storm-inc"},
		42:   []URI{"thhg"},
		43:   []URI{"tmh"},
		44:   []URI{"the-racketeers"},
		45:   []URI{"crackers-in-action"},
		46:   []URI{"legion-of-doom"},
		47:   []URI{"the-grand-council"},
		48:   []URI{"untouchables", "uniq", "xap", "pentagram"},
		49:   []URI{"italsoft"},
		50:   []URI{"future-brain-inc"},
		51:   []URI{"pirate"},
		52:   []URI{"creators-of-intense-art", "art-creation-enterprise"},
		53:   []URI{"vla"},
		54:   []URI{"the-north-west-connection"},
		55:   []URI{"the-sysops-association-network"},
		56:   []URI{"american-pirate-industries"},
		57:   []URI{"pirates-sick-of-initials"},
		58:   []URI{"byte-bandits-bbs"},
		59:   []URI{"sorcerers"},
		590:  []URI{"sorcerers", "future-brain-inc"},
		60:   []URI{"katharsis"},
		61:   []URI{"national-elite-underground-alliance"},
		62:   []URI{"public-enemy", "pe*trsi*tdt", "north-american-society-of-anarchists", "red-sector-inc", "the-dream-team"},
		63:   []URI{"public-enemy"},
		64:   []URI{"razor-1911"},
		65:   []URI{"tristar-ampersand-red-sector-inc", "red-sector-inc"},
		66:   []URI{"tristar-ampersand-red-sector-inc", "pe*trsi*tdt", "the-dream-team", "skid-row", "coop"},
		67:   []URI{"tristar-ampersand-red-sector-inc"},
		68:   []URI{"the-dream-team"},
		69:   []URI{"rom-1911", "razor-1911"},
		70:   []URI{"high-society"},
		71:   []URI{"trinity-reviews", "lancelot-2"},
		72:   []URI{"real-pirates-guide"},
		73:   []URI{"the-amatuer-crackist-tutorial"}, //nolint:misspell
		74:   []URI{"church-chat", "ptl-club"},
		75:   []URI{"corrupted-programming-international", "cpi-newsletter"},
		76:   []URI{"official-unprotection-scheme-library", "copycats-inc"},
		77:   []URI{"the-elementals-piratelist"},
		78:   []URI{"game-release-list"},
		79:   []URI{"gif-news"},
		80:   []URI{"hackers-unlimited", "mickey-mouse-club"},
		800:  []URI{"hackers-unlimited", "mickey-mouse-club", "crackers-in-action"},
		81:   []URI{"national-pirate-list"},
		82:   []URI{"phreakers-handbook"},
		83:   []URI{"spectrum"},
		84:   []URI{"the-pirate-world", "the-pirate-syndicate"},
		85:   []URI{"fairlight"},
		86:   []URI{"scb"},
		87:   []URI{"imperial-warlords"},
		88:   []URI{"public-brand-software"},
		89:   []URI{"stealth-pirates-corp"},
		90:   []URI{"occult-network"},
		91:   []URI{"microcomputer-assembly-software-hackers"},
		92:   []URI{"the-knights-of-the-round-table"},
		93:   []URI{"black-star-productions"},
		94:   []URI{"the-washington-state-network"},
		95:   []URI{"reno"},
		96:   []URI{"myth-inc"},
		97:   []URI{"crime-syndicate-net"},
		98:   []URI{"ecp"},
		99:   []URI{"west-coast-alliance"},
		100:  []URI{"warriors-against-copy-protection"},
		101:  []URI{"the-stealth-pirate-network"},
		102:  []URI{"state-of-the-art"},
		103:  []URI{"psycho-corporate-productions"},
		104:  []URI{"association-of-software-conspiracy"},
		105:  []URI{"north-american-pirate_phreak-association"},
		106:  []URI{"the-union"},
		107:  []URI{"the-sure-logic-syndicate"},
		108:  []URI{"the-alternative"},
		109:  []URI{"quartex"},
		110:  []URI{"gainseville-pirates-association"},
		112:  []URI{"ffa"},
		113:  []URI{"eagle-soft-incorporated"},
		114:  []URI{"digital-exchange-pirate-board-alliance"},
		115:  []URI{"bdp"},
		116:  []URI{"assembly-language-magazine"},
		117:  []URI{"acme"},
		118:  []URI{"acronym"},
		119:  []URI{"bad-ass-dudes", "bad-news"},
		120:  []URI{"bad-association"},
		121:  []URI{"classic-vocs"},
		122:  []URI{"club-elan"},
		123:  []URI{"damn-excellent-ansi-design"},
		124:  []URI{"digital-noise-alliance"},
		125:  []URI{"disassemblers-of-america"},
		126:  []URI{"dragon-clan"},
		127:  []URI{"dread"},
		128:  []URI{"dutch-computer-enterprise"},
		129:  []URI{"east-coast-connection"},
		130:  []URI{"electro-magnetic-crackers", "ultra-tech*electro-magnetic-crackers"},
		//131:  []URI{"ultra-tech*electro-magnetic-crackers"},
		132: []URI{"game_busters"},
		133: []URI{"gcl"},
		134: []URI{"heavenly-hackers-group"},
		135: []URI{"interceptor"},
		136: []URI{"los-angeles-sysops-alliance"},
		137: []URI{"mad"},
		138: []URI{"oblivion"},
		139: []URI{"pacific-brigade"},
		140: []URI{"paradox"},
		141: []URI{"pc_cracking-service"},
		142: []URI{"personal-cracking"},
		143: []URI{"petra"},
		144: []URI{"powr"},
		145: []URI{"psycho"},
		146: []URI{"roach"},
	}
}

// Find returns the tidbit IDs for the given URI.
//
// The ID returned can be used in a string conversion to get the description.
// The ID can also be used to get the URIs of the tidbit.
func Find(uri string) []ID {
	ids := []ID{}
	for id, uris := range Groups() {
		for val := range slices.Values(uris) {
			if val == URI(uri) {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

// Package tidbit offers hyperlinked historical information about the Scene releasers and groups.
package tidbit

import (
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Defacto2/releaser"
	"github.com/Defacto2/server/internal/panics"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// URI is the URL slug of the releaser.
type URI string

// ID is the identifier of the tidbit.
type ID int

const extensions = parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock | parser.Footnotes

// Markdown returns the markdown content of the tidbit that is stored in the directory
// in the provided file system. If the file does not exist or is empty then nil is returned.
//
// Generally the String method should be used to get the description of the tidbit instead
// of this Markdown method.
func (id ID) Markdown(sl *slog.Logger, fs embed.FS, dir string) []byte {
	const msg = "tidbit markdown"
	if sl == nil {
		panic(fmt.Errorf("%s: %w", msg, panics.ErrNoSlog))
	}
	name := filepath.Join(dir, fmt.Sprintf("%d.md", id))
	b, err := fs.ReadFile(name)
	if err != nil {
		name := fmt.Sprintf("%d.md", id)
		sl.Error(msg, slog.String("read error", name), slog.Any("error", err))
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
func (id ID) String(sl *slog.Logger, fs embed.FS) string {
	if b := id.Markdown(sl, fs, "public/md/tidbit"); b != nil {
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
func Groups() Tibits { //nolint:maintidx
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
		666:  []URI{"acid-productions"},
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
		// 131:  []URI{"ultra-tech*electro-magnetic-crackers"},
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
		147: []URI{"west-coast-cracking-production"},
		148: []URI{"warez"},
		149: []URI{"unit-173"},
		150: []URI{"union-of-crackers"},
		151: []URI{"tired-of-protection"},
		152: []URI{"the-warez-alliance"},
		153: []URI{"the-syndicate"},
		154: []URI{"strike"},
		155: []URI{"software-in-danger"},
		156: []URI{"roi-production", "warez-houze-bbs"},
		157: []URI{"the-ware-report"},
		158: []URI{"copyright-infiltration-agency"},
		159: []URI{"copyright-infiltration-agency", "crackers-in-action"},
		160: []URI{"god-network"},
		161: []URI{"corporation-for-public-cybercasting-2001"},
		162: []URI{"defjam"},
		163: []URI{"amateur-crackist-tutorial"},
		164: []URI{"phoenix"},
		165: []URI{"advanced-pirate-technology"},
		166: []URI{"censor", "rabid"},
		167: []URI{"cosmic-press"},
		168: []URI{"dune-newsletter", "deadly-underground-network-of-elites"},
		169: []URI{"fear-newsletter"},
		170: []URI{"galactic-review"},
		171: []URI{"illuminatus"},
		172: []URI{"victoria-independent-piracy"},
		173: []URI{"tsan-newsletter"},
		174: []URI{"toxic-shock"},
		175: []URI{"the-humble-review", "thg-fx"},
		176: []URI{"radiant"},
		177: []URI{"quasar-magazine"},
		178: []URI{"propaganda"},
		179: []URI{"prevues"},
		180: []URI{"nuke-infojournal"},
		181: []URI{"micropirates-inc"},
		182: []URI{"legion-of-dynamic-discord"},
		183: []URI{"lancelot"},
		184: []URI{"insanity"},
		185: []URI{"imphobia"},
		186: []URI{"bbslst"},
		187: []URI{"7415"},
		188: []URI{"ameriboards"},
		189: []URI{"copiers-newsletter"},
		190: []URI{"apex"},
		191: []URI{"dmacks-lost-classics"},
		192: []URI{"dmz-review"},
		193: []URI{"eliteslst"},
		194: []URI{"euro_bbs_list"},
		195: []URI{"excretion-anarchy"},
		196: []URI{"graphically-enhanced-magazine"},
		197: []URI{"hype-magazine"},
		198: []URI{"lancelot-2"},
		203: []URI{"mac_pirate-list"},
		204: []URI{"mr-bane-800-number-list"},
		205: []URI{"national-network-of-anarchists-and-nihilists"},
		206: []URI{"not-productions"},
		207: []URI{"pirates-analyze-warez"},
		208: []URI{"pirates-cove"},
		209: []URI{"swedish-real-top-list"},
		210: []URI{"the-cutting-edge"},
		211: []URI{"the-dark-wheel"},
		212: []URI{"the-hack-report"},
		213: []URI{"the-illustrator", "prognosis"},
		214: []URI{"the-product"},
		215: []URI{"the-software-review"},
		216: []URI{"uncut-ware-report"},
		217: []URI{"ware-report"},
		218: []URI{"wildsider"},
		219: []URI{"world-elite-bbs-list"},
		220: []URI{"adrenalin"},
		221: []URI{"brumus-bear-review"},
		222: []URI{"apex-reviewers"},
		223: []URI{"anti-warez-association"},
		224: []URI{"altered-reality"},
		225: []URI{"globelist-world-bbs-listing"},
		226: []URI{"elite-underground"},
		227: []URI{"elite-bbs-listing"},
		228: []URI{"demented-review-crew"},
		229: []URI{"cybercrime-international-network", "packet"},
		230: []URI{"criminals-of-radical-extremes"},
		231: []URI{"criminal-intent"},
		232: []URI{"corruption"},
		233: []URI{"iridium-magazine"},
		234: []URI{"hijack"},
		235: []URI{"ice-weekly-newsletter"},
		236: []URI{"infinity-crew"},
		237: []URI{"insane-reality"},
		238: []URI{"maniac-magazine"},
		239: []URI{"osmium"},
		240: []URI{"pandemonium"},
		241: []URI{"paradigm-press"},
		242: []URI{"pc-charts"},
		243: []URI{"phantasy-magazine"},
		244: []URI{"pirate-sites"},
		245: []URI{"review-inc", "wildstar"},
		246: []URI{"shallow-grounds"},
		247: []URI{"spetznas"},
		248: []URI{"spreadpoint"},
		249: []URI{"terratron"},
		250: []URI{"the-real-console-release-charts"},
		251: []URI{"the-review-crew"},
		252: []URI{"top-group-charts"},
		253: []URI{"traderslist", "fabulous-baker-boys"},
		254: []URI{"underground-experts-united"},
		255: []URI{"unreal-magazine"},
		256: []URI{"atomic-review"},
		257: []URI{"aryan-sekret-service"},
		259: []URI{"anti-debugging-tricks"},
		260: []URI{"lithium", "alt_1"},
		261: []URI{"the-alliance"},
		262: []URI{"brotherhood-of-warez"},
		263: []URI{"blur"},
		264: []URI{"blackadder-ftp-site-list"},
		265: []URI{"big-book-of-boxes"},
		266: []URI{"best-of-the-best-phreaking-man", "the-brotherhood"},
		267: []URI{"bizarre-types-of-wares"},
		268: []URI{"ralph-productions", "coming-soon"},
		269: []URI{"console-news"},
		270: []URI{"corpse", "corosion"}, //nolint:misspell
		271: []URI{"consolitation"},
		272: []URI{"doyadigm"},
		273: []URI{"digital-press", "console-gaming-informers"},
		274: []URI{"distorted"},
		275: []URI{"demo-reviews"},
		276: []URI{"datazine"},
		277: []URI{"dreadloc"},
		278: []URI{"frontier-console-magazine"},
		279: []URI{"evolution-magazine"},
		280: []URI{"german-modem-scene-report"},
		281: []URI{"future-game-listings"},
		282: []URI{"image-magazine"},
		283: []URI{"hard-core-hackers"},
		284: []URI{"insomnia-emag"},
		285: []URI{"banch-o-guyz", "lamer-of-the-world"},
		286: []URI{"nsdap"},
		287: []URI{"no-fear", "nofear-news"},
		288: []URI{"pirate-software-alliance"},
		289: []URI{"ntt"},
		290: []URI{"really-awful-music", "ram-newszine"},
		291: []URI{"quality-control-reviews"},
		292: []URI{"progeny"},
		293: []URI{"primal", "primag"},
		294: []URI{"rebels-of-telecommunications", "prophecy"},
		295: []URI{"ransom"},
		296: []URI{"scam-magazine"},
		297: []URI{"robin-nests-hack-report"},
		298: []URI{"software-runners-from-hell"},
		299: []URI{"software-pirates-alliance"},
		300: []URI{"strictly-pirates"},
		301: []URI{"stellar-7", "the-console-mag"},
		302: []URI{"the-brotherhood-of-gods-and-retards"},
		303: []URI{"the-demo-contact-list"},
		304: []URI{"the-pirates-manifesto"},
		305: []URI{"the-week-charts"},
		306: []URI{"the-worldwide-amiga-bbs-list"},
		307: []URI{"toxin"},
		308: []URI{"trip-2-hell"},
		309: []URI{"virus-laboratories-and-distribution"},
		310: []URI{"a_list"},
		311: []URI{"amiga-major-games-release-charts"},
		312: []URI{"bbs-and-users-digest"},
		313: []URI{"cyberspace-chart", "lisence"}, //nolint:misspell
		314: []URI{"dutch-trader-charts"},
		315: []URI{"fake-list"},
		316: []URI{"higher-mental-plane"},
		317: []URI{"hydra-magazine"},
		318: []URI{"infinity-e_mag"},
		319: []URI{"monthly-competition"},
		320: []URI{"psxdox"},
		321: []URI{"pure-console"},
		322: []URI{"swedish-elite-bbs-list"},
		323: []URI{"the-jargon-file"},
		324: []URI{"the-dreamcharts-release"},
		325: []URI{"sony-playstation-report"},
		326: []URI{"inquisition"},
		327: []URI{"the-week-in-warez"},
		328: []URI{"the-naked-truth-magazine"},
		329: []URI{"reality-check-network"},
		330: []URI{"the-game-review"},
		331: []URI{"the-reviewers-guild"},
		332: []URI{"trc-ware-report"},
		334: []URI{"the-warez-magazine"},
		335: []URI{"3rd-world-paki-report"},
		336: []URI{"affinity"},
		337: []URI{"cybermail"},
		338: []URI{"defacto"},
		339: []URI{"the-gamers-edge", "the-reviewers-guild"},
		340: []URI{"rebels"},
		341: []URI{"defacto2"},
		342: []URI{"the-council"},
		343: []URI{"core"},
		344: []URI{"chemical-reaction"},
		345: []URI{"email-compilation"},
		347: []URI{"aesthetic"},
		348: []URI{"core-reviews", "core"},
		349: []URI{"hybrid-christmas-e_mag"},
		350: []URI{"monthly-console-scene-charts-international"},
		351: []URI{"most-complete-psx-cheat-list-ever"},
		352: []URI{"pc-hacking-faq"},
		353: []URI{"console-release-dates"},
		354: []URI{"zenith-zine", "kirra"},
		355: []URI{"scooby-snack-magazine"},
		356: []URI{"sony-playstation-game-reviews"},
		357: []URI{"tfw"},
		358: []URI{"the-reservoir-warez-report", "the-reservoir-dogs"},
		359: []URI{"sneakers"},
		360: []URI{"wave"},
		361: []URI{"top-telnet-traders-weekly"},
		362: []URI{"weekly-courier-report"},
		363: []URI{"the-warez-loop"},
		364: []URI{"the-legendary-report"},
		365: []URI{"how-to-crack"},
		366: []URI{"madwizards"},
		367: []URI{"mp3"},
		368: []URI{"new-dtl"},
		369: []URI{"anemia"},
		370: []URI{"just-the-facts"},
		371: []URI{"the-flame-arrows"},
		372: []URI{"impact"},
		373: []URI{"emporio"},
		374: []URI{"vengeance"},
		375: []URI{"outlaws-exchange"},
		376: []URI{"spirit-of-illusion"},
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

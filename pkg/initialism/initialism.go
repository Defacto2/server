// Package initialism provides a list of initialisms for releasers of The Scene.
package initialism

import "strings"

// URI is a the URL slug of the releaser.
type URI string

// List is a map of initialisms to releasers.
type List map[URI][]string

// all initialisms should be in their stylized form
var initialisms = List{
	"2000ad":                                   {"2KAD", "2000 AD"},
	"aces-of-ansi-art":                         {"AAA"},
	"acid-productions":                         {"ACiD", "ANSi Creators in Demand"},
	"arab-team-4-reverse-engineering":          {"AT4RE"},
	"art-of-reverse-engineering":               {"AORE"},
	"backlash":                                 {"BLH"},
	"bentley-sidwell-productions":              {"BSP"},
	"blades-of-steel":                          {"BOS", "Blades"},
	"boys-from-company-c":                      {"BCC"},
	"canadian-pirates-inc":                     {"CPI"},
	"cd-images-for-the-elite":                  {"CiFE"},
	"celerity-utilities-division":              {"CUD"},
	"cheat-requests-for-the-underground-elite": {"CRUE"},
	"class":                                 {"CLS"},
	"classic":                               {"CLS"},
	"chemical-reaction":                     {"CRO"},
	"crack-in-morocco":                      {"CiM"},
	"darksiders":                            {"DS"},
	"damn-excellent-ansi-design":            {"DEAD"},
	"deviance":                              {"DEV", "DVN"},
	"divine":                                {"DVN"},
	"drunken-rom-group":                     {"DRG", "Drunken"},
	"drink-or-die":                          {"DOD"},
	"dynasty":                               {"DYN"},
	"eclipse":                               {"ECL"},
	"energy":                                {"NRG"},
	"fairlight":                             {"FLT"},
	"faith":                                 {"FTH"},
	"fighting-for-fun":                      {"fff"},
	"fusion":                                {"FSN"},
	"genesis":                               {"GNS"},
	"graphics-rendered-in-magnificence":     {"GRiM"},
	"highroad":                              {"HR"},
	"hoodlum":                               {"HLM"},
	"hybrid":                                {"HYB"},
	"independent":                           {"non-affiliated people", "IND"},
	"independent-crackers-union":            {"ICU"},
	"insane-creators-enterprise":            {"iCE"},
	"international-network-of-crackers":     {"INC"},
	"licensed-to-draw":                      {"LTD"},
	"lightforce":                            {"LFC", "LF"},
	"linezer0":                              {"Lz0"},
	"motiv8":                                {"M8"},
	"mirage":                                {"MIR"},
	"national-elite-underground-alliance":   {"NEUA", "North Eastern Underground Alliance"},
	"orion":                                 {"ORiON", "ORN"},
	"paradigm":                              {"PDM", "Zeus"},
	"paradox":                               {"PDX"},
	"pentagram":                             {"PTG"},
	"phrozen-crew":                          {"PC"},
	"pirates-gone-crazy":                    {"PGC"},
	"pirates-with-attitude":                 {"PWA"},
	"ptl-club":                              {"PTL"},
	"prestige":                              {"PSG", "PST"},
	"public-enemy":                          {"PE"},
	"razor-1911":                            {"RZR", "Razor"},
	"rebels":                                {"RBS"},
	"reloaded":                              {"RLD"},
	"resistance-is-futile":                  {"RiF"},
	"reverse-engineers-dream":               {"RED"},
	"rise-in-superior-couriering":           {"RiSC"},
	"seek-n-destroy":                        {"SND"},
	"skid-row":                              {"SR", "Skidrow"},
	"scoopex":                               {"SCX", "SPX"},
	"silicon-dream-artists":                 {"SDA"},
	"superior-art-creations":                {"SAC"},
	"the-crazed-asylum":                     {"TCA"},
	"the-console-division":                  {"TCD"},
	"the-dream-team":                        {"TDT"},
	"the-firm":                              {"FiRM", "FRM"},
	"the-force-team":                        {"TFT"},
	"the-humble-guys":                       {"THG", "Humble"},
	"the-millennium-group":                  {"TMG"},
	"the-nova-team":                         {"TNT"},
	"the-sabotage-rebellion-hackers":        {"TSRh"},
	"tristar-ampersand-red-sector-inc":      {"TRSi", "TRS", "Tristar"},
	"united-cracking-force":                 {"UCF"},
	"united-software-association*fairlight": {"USA-FLT"},
	"united-software-association":           {"USA"},
	"virility":                              {"VRL"},
	"zero-waiting-time":                     {"ZWT"},
}

// Join returns the initialisms for the URI as a comma separated string.
func Join(uri string) string {
	i := Initialism(uri)
	if len(i) == 0 {
		return ""
	}
	return strings.Join(i, ", ")
}

// Initialism returns the initialism for the URI.
// If the URI does not have an initialism then an empty string is returned.
func Initialism(uri string) []string {
	return initialisms[URI(uri)]
}

// Initialisms returns the list of initialisms.
func Initialisms() List {
	return initialisms
}

// IsInitialism returns true if the URI has an initialism.
func IsInitialism(uri string) bool {
	_, ok := initialisms[URI(uri)]
	return ok
}

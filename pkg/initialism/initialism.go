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
	"advanced-art-of-cracking-group":           {"AAOCG"},
	"advanced-pirate-technology":               {"APT"},
	"air":                                      {"Team AiR", "Addiction In Releasing"},
	"alpha-flight":                             {"AFL"},
	"amnesia":                                  {"AMN"},
	"anemia":                                   {"ANM"},
	"arab-team-4-reverse-engineering":          {"AT4RE"},
	"arrogant-couriers-with-essays":            {"ACE"},
	"art-of-reverse-engineering":               {"AORE"},
	"artists-in-revolt":                        {"AiR"},
	"backlash":                                 {"BLH"},
	"bad-ass-dudes":                            {"BAD"},
	"bentley-sidwell-productions":              {"BSP"},
	"blades-of-steel":                          {"BOS", "Blades"},
	"bitchin-ansi-design":                      {"BAD"},
	"boys-from-company-c":                      {"BCC"},
	"canadian-pirates-inc":                     {"CPI"},
	"cd-images-for-the-elite":                  {"CiFE"},
	"celerity-utilities-division":              {"CUD"},
	"cheat-requests-for-the-underground-elite": {"CRUE"},
	"class":                               {"CLS"},
	"classic":                             {"CLS"},
	"chemical-reaction":                   {"CRO"},
	"couriers-of-pirated-software":        {"COPS"},
	"crackers-in-action":                  {"CIA"},
	"crack-in-morocco":                    {"CiM"},
	"creators-of-intense-art":             {"CIA"},
	"cybercrime-international-network":    {"CCi", "CyberCrime Inc."},
	"darksiders":                          {"DS"},
	"da-breaker-crew":                     {"DBC"},
	"damn-excellent-ansi-design":          {"DeAD"},
	"damn-excellent-ansi-designers":       {"DeAD"}, // Correct
	"dead-on-arrival":                     {"DOA"},
	"dead-pirates-society":                {"DPS"},
	"defacto":                             {"DF"},
	"defacto2":                            {"DF2"},
	"deviance":                            {"DEV", "DVN"},
	"devotion":                            {"DEV", "devot"},
	"digerati":                            {"DGT"},
	"distinct":                            {"DTC", "DTN"},
	"divide-by-zero":                      {"DBZ"},
	"divine":                              {"DVN"},
	"drunken-rom-group":                   {"DRG", "Drunken"},
	"drink-or-die":                        {"DOD"},
	"dvt":                                 {"Devotion", "TeamDVT"},
	"dynasty":                             {"DYN"},
	"dynamix":                             {"DNX"},
	"dytec":                               {"DYT", "DTC"},
	"dvniso":                              {"Deviance"},
	"eagle-soft-incorporated":             {"ESI"},
	"ebola-virus-crew":                    {"EVC"},
	"eclipse":                             {"ECL"},
	"empire-of-darkness":                  {"EOD"},
	"energy":                              {"NRG"},
	"equinox":                             {"EQX"},
	"esp-pirates":                         {"ESP"},
	"fairlight":                           {"FLT"},
	"fairlight-dox":                       {"FDX", "FLTDOX"},
	"faith":                               {"FTH"},
	"fantastic-4-cracking-group":          {"F4CG"},
	"fast-action-trading-elite":           {"fATE"},
	"fighting-for-fun":                    {"fff"},
	"fight-only-for-freedom":              {"FOFF"},
	"file-rappers":                        {"FR"},
	"flying-horse-cracking-force":         {"FHCF"},
	"future-crew":                         {"FC"},
	"future-brain-inc":                    {"FBi"},
	"fusion":                              {"FSN"},
	"genesis":                             {"GNS"},
	"graphic-revolution-in-progress":      {"GRiP"},
	"elite-couriers-group":                {"ECG"},
	"epsilon":                             {"EPS"},
	"ghost-riders":                        {"GRS"},
	"graphics-rendered-in-magnificence":   {"GRiM"},
	"highroad":                            {"HR"},
	"hipe":                                {"HPE"},
	"hoodlum":                             {"HLM"},
	"humble-dox":                          {"The Humble Guys DOX"},
	"hybrid":                              {"HYB"},
	"kalisto":                             {"KAL"},
	"illusion":                            {"iLL"},
	"independent":                         {"IND"},
	"independent-crackers-union":          {"ICU"},
	"insane-creators-enterprise":          {"iCE"},
	"international-network-of-crackers":   {"INC"},
	"international-cracking-crew":         {"iCC"},
	"inc-documentation-division":          {"IDD"},
	"inc-utility-division":                {"IUD"},
	"legacy":                              {"LGC", "LGY"},
	"licensed-to-draw":                    {"LTD"},
	"light-speed-distributors":            {"LSD"},
	"lightforce":                          {"LFC", "LF"},
	"linezer0":                            {"Lz0", "Linezero"},
	"live-now-die-later":                  {"LnDL"},
	"lucid":                               {"LCD"},
	"malicious-art-denomination":          {"MAD"},
	"majic-12":                            {"M12"},
	"millenium":                           {"MnM"},
	"mutual-assured-destruction":          {"MAD"},
	"manifest":                            {"MFD", "Manifest Destiny"},
	"motiv8":                              {"M8"},
	"mirage":                              {"MIR"},
	"myth*deviance":                       {"MDVN"},
	"national-elite-underground-alliance": {"NEUA", "North Eastern Underground Alliance"},
	"national-underground-application-alliance": {"NUAA"},
	"napalm":                     {"NPM"},
	"netrunners":                 {"NR"},
	"new-york-crackers":          {"NYC"},
	"nexus":                      {"NXS", "NX"},
	"nokturnal-trading-alliance": {"NTA"},
	"north-american-pirate_phreak-association": {"NAPPA"},
	"oddity":                             {"ODT"},
	"old-warez-inc":                      {"OWI"},
	"orion":                              {"ORN"},
	"origin":                             {"OGN"},
	"originally-funny-guys":              {"OFG"},
	"paradigm":                           {"PDM", "Zeus"},
	"paradox":                            {"PDX"},
	"pentagram":                          {"PTG"},
	"phrozen-crew":                       {"PC"},
	"pirates-analyze-warez":              {"PAW"},
	"pirates-gone-crazy":                 {"PGC"},
	"pirates-sick-of-initials":           {"PSi"},
	"pirates-with-attitude":              {"PWA"},
	"ptl-club":                           {"PTL"},
	"prestige":                           {"PSG", "PST"},
	"public-enemy":                       {"PE"},
	"public-enemy*red-sector-inc":        {"PE", "PE/RSI"},
	"razor-1911":                         {"RZR", "Razor"},
	"razor-1911-cd-division":             {"RazorCD"},
	"reality-check-network":              {"RCN"},
	"rebels":                             {"RBS"},
	"red-sector-inc":                     {"RSI"},
	"release-on-rampage":                 {"RoR"},
	"reloaded":                           {"RLD"},
	"relentless-pursuit-of-magnificence": {"RPM"},
	"request-to-send":                    {"RTS"},
	"resistance-is-futile":               {"RiF"},
	"reverse-engineers-dream":            {"RED"},
	"reverse-engineering-in-software":    {"REiS"},
	"reverse-engineering-passion-team":   {"REPT"},
	"rise-in-superior-couriering":        {"RiSC"},
	"seek-n-destroy":                     {"SND", "Seek and Destroy"},
	"skid-row":                           {"SR", "Skidrow"},
	"scoopex":                            {"SCX", "SPX"},
	"scienide":                           {"SCi"},
	"silicon-dream-artists":              {"SDA"},
	"sodom":                              {"SDM"},
	"software-chronicles-digest":         {"SCD"},
	"software-pirates-inc":               {"SPI"},
	"superior-art-creations":             {"SAC"},
	"surprise-productions":               {"SP"},
	"the-crazed-asylum":                  {"TCA"},
	"the-console-division":               {"TCD"},
	"the-dream-team":                     {"TDT"},
	"the-dream-team*skid-row":            {"TDT/SR"},
	"the-dream-team*tristar-ampersand-red-sector-inc": {"TDT/TRSi"},
	"the-firm":                              {"FiRM", "FRM"},
	"the-force-team":                        {"TFT"},
	"the-grand-council":                     {"TGC"},
	"the-humble-guys":                       {"THG", "Humble"},
	"the-millennium-group":                  {"TMG"},
	"the-nova-team":                         {"TNT"},
	"the-one-and-only":                      {"TOAO"},
	"the-outlaws":                           {"TOL", "OL"},
	"the-players-club":                      {"TPC"},
	"the-reversers-ultimate-epidemic":       {"tRUE"},
	"the-reviewers-guild":                   {"TRG"},
	"the-sabotage-rebellion-hackers":        {"TSRh"},
	"the-software-innovation-network":       {"SIN"},
	"the-sysops-association-network":        {"TSAN"},
	"the-underground-council":               {"UGC"},
	"the-untouchables":                      {"UNT"},
	"thg-fx":                                {"The Humble Guys FX"},
	"tristar":                               {"TRS"},
	"tristar-ampersand-red-sector-inc":      {"TRSi", "TRS", "Tristar"},
	"ultra-tech":                            {"UT"},
	"union":                                 {"UNi"},
	"united-artist-association":             {"UAA"},
	"united-couriers":                       {"UC"},
	"united-cracking-force":                 {"UCF"},
	"united-group-international":            {"UGI"},
	"united-reverse-engineering-team":       {"URET"},
	"united-software-association*fairlight": {"USA/FLT"},
	"united-software-association":           {"USA"},
	"underpl":                               {"UPL"},
	"untouchables":                          {"UNT"},
	"vengeance":                             {"VGN", "VEN"},
	"virility":                              {"VRL"},
	"xtreeme":                               {"XT"},
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
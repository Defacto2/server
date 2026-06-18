// Package name provides functionality for handling the URL path of a releaser.
// It contains a type Path that represents the URL path of a releaser and methods
// to validate and retrieve the well-known styled name of the releaser.
// It also contains a map of releasers and their well-known styled names.
package name

import (
	"errors"
	"maps"
	"regexp"
	"slices"
	"strings"
)

var ErrInvalidPath = errors.New("the path contains invalid characters")

// A Path is the partial URL path of the releaser.
type Path string

// String returns the well-known styled name of the releaser if it exists in the
// names, lowercase or uppercase lists. Otherwise it returns an empty string.
//
// Example:
//
//	name.Path("acid-productions").String() = "ACiD Productions"
//	name.Path("razor-1911").String() = "" // unlisted
func (path Path) String() string {
	p := Path(strings.ToLower(string(path)))
	if _, match := specials[p]; match {
		return specials[p]
	}
	return ""
}

// Valid returns true if the URL path uses valid characters.
// Valid URL paths are all lowercase and contain only alphanumeric characters, dashes, underscores,
// ampersands and asterisks.
//
// Example:
//
//	name.Path("acid-productions").Valid() = true
//	name.Path("acid-productions!").Valid() = false
func (path Path) Valid() bool {
	re := regexp.MustCompile(`^[a-z0-9\&\-_\*]+$`)
	return re.MatchString(string(path))
}

// A List is a map of releasers and their well-known styled names.
type List map[Path]string

/*
The following list of styled names is used to test the Path type and its methods.

Stylized names should avoid using special characters that may get encoded in the URL
or converted due to their special uses within the name.
*/

// Names returns the list of well-known styled names.
func Names() *List { //nolint:funlen
	list := List{
		"mother-superior-ftp":                   "The Mother Superior FTP",
		"wolves-house-ftp":                      "The Wolves House FTP",
		"rock-ftp":                              "The Rock FTP",
		"dafat-ftp":                             "DaFat FTP",
		"swat":                                  "SWaT",
		"skill":                                 "SKiLL",
		"icepack":                               "iCEPACK",
		"drift":                                 "DRiFT",
		"god-network":                           "G.O.D. Network",
		"german-diskdoubler":                    "German DiskDoubler",
		"excel_xl":                              "EXCEL/XL!",
		"ob_gyn":                                "OB/GYN",
		"primag":                                "PRiMAG",
		"roi-production":                        "ROI Production",
		"dome":                                  "DoME",
		"image-productions-2":                   "iMAGE Productions (#2)",
		"image-nj":                              "iMAGE (NJ)",
		"ninja":                                 "NiNJA",
		"orgasming-gaming-magazine":             "orGAsMING Gaming Magazine",
		"gameboycolor-world-charts":             "GameBoyColor World Charts",
		"email-compilation":                     "e.mail compilation",
		"2000-ad":                               "2000AD",
		"79th-trac":                             "79th TRAC",
		"acid-productions":                      "ACiD Productions",
		"biased":                                "bIASED",
		"binpda":                                "BiNPDA",
		"coop":                                  "TDT / TRSi",
		"core":                                  "CoRE",
		"copycats-inc":                          "CopyCats Inc",
		"coreutil":                              "The Utility Division of CORE",
		"crackpl":                               "CrackPL",
		"cybermail":                             "CyberMail",
		"dbcdemo":                               "DBCDemo",
		"dmacks-lost-classics":                  "Dmack's Lost Classics",
		"dreadloc":                              "DREADLoC",
		"dumptruck":                             "dumpTruck",
		"defacto2net":                           "Defacto2 website",
		"drm-ftp":                               "dRM FTP",
		"dst-ftp":                               "dst FTP",
		"dvniso":                                "DVNiSO",
		"dvtiso":                                "DVTiSO",
		"epic":                                  "EPiC",
		"esp-pirates":                           "ESP Pirates",
		"extreme-net":                           "ExtremeNET",
		"excretion-anarchy":                     "eXCReTION",
		"fx2-graphics-group":                    "Fx/2 Graphics Group",
		"hashx":                                 "Hash X",
		"htbzine":                               "HTBZine",
		"linezer0":                              "LineZer0",
		"lucid":                                 "LuCiD",
		"ice-weekly-newsletter":                 "iCE Weekly Newsletter",
		"icon":                                  "iCON",
		"imars":                                 "iMARS",
		"jrp":                                   "Japanese Release Project",
		"oneup":                                 "OneUp",
		"orion":                                 "ORiON",
		"mmi":                                   "MMi",
		"mp2k":                                  "MP2K",
		"nc_17":                                 "NC-17",
		"nicjr":                                 "NicJr",
		"noclass":                               "NoClass",
		"nofx-bbs":                              "NoFX BBS",
		"nukethis":                              "NukeThis",
		"numbers":                               "NUMbers",
		"nrp":                                   "NoRePack",
		"paradox":                               "Paradox",
		"phoenixbbs":                            "Phoenix BBS",
		"pjs-tower-bbs":                         "PJs Tower BBS",
		"playme":                                "PlayMe",
		"pocketheaven":                          "PocketHeaven",
		"psico":                                 "PSiCO",
		"ptl-club":                              "PTL Club",
		"pouet":                                 "Pouët",
		"risciso":                               "RISCiSO",
		"sda-review":                            "SDA Review",
		"seek-n-destroy":                        "Seek n Destroy",
		"sma-posse":                             "SMA Posse",
		"shitonlygerman":                        "ShitOnlyGerman",
		"software-pirates-inc":                  "Software Pirates Inc",
		"surprise-productions":                  "Surprise! Productions",
		"syndicate":                             "SyNDiCaTE",
		"r2":                                    "Rebels + 2000AD",
		"razordox":                              "RazorDOX",
		"rhvid":                                 "RHViD",
		"rzsoft-ftp":                            "RZSoft FTP",
		"tkc*crackers-in-action":                "tKC/Crackers in Action",
		"tdu_jam":                               "TDU Jam!",
		"team-xtx":                              "Team XTX",
		"thg-fx":                                "THG-FX",
		"tft-team":                              "TFT Team",
		"tpinc":                                 "TPiNC",
		"trsi":                                  "TRSi",
		"tristar-ampersand-red-sector-inc":      "Tristar & Red Sector Inc",
		"the-dvdr-releasing-standards":          "The DVDR Releasing Standards",
		"the-firm":                              "The FiRM",
		"tsg-ftp":                               "tSG FTP",
		"tport":                                 "tPORt",
		"underpl":                               "UnderPL",
		"unreal-magazine":                       "UnReal Magazine",
		"united-software-association*fairlight": "United Software Association + Fairlight PC Division",
		"vdr-lake-ftp":                          "VDR Lake FTP",
		"well-release-anything":                 "We'll Release Anything",
		"uniq":                                  "UNiQ",
		"ypogeios":                              "YPOGEiOS",
		"xdb":                                   "X-db",
		"xquizit-ftp":                           "XquiziT FTP",
		"pnx":                                   "Cyber Angels Phoenix",
		"cpi-newsletter":                        "CPI Newsletter",
		"warez":                                 "WareZ",
		"mai-review":                            "MAi Review",
		"nuke-infojournal":                      "[NuKE] InfoJournal",
		"tsan-newsletter":                       "TSAN Newsletter",
		"vip-magazine":                          "ViP Magazine",
		"dmz-review":                            "DMZ Review",
		"mr-bane-800-number-list":               "Mr. Bane's 800 Number List",
		"ware-report":                           "WARE Report",
		"apex-reviewers":                        "APEX Reviewers",
		"globelist-world-bbs-listing":           "GlobeList World BBS Listing",
		"spetznas":                              "SpetzNas",
		"insomnia-emag":                         "iNSOMNiA E-Mag",
		"nofear-news":                           "NOFEAR News",
		"ram-newszine":                          "RAM Newszine",
		"scam-magazine":                         "SCAM! Magazine",
		"ntt":                                   "ENTiTY",
		"eliteslst":                             "ELITES.LST",
		"radiant":                               "RADiANT",
		"genesis-ppe":                           "Genesis PPE",
		"genesis-404":                           "Genesis (404)",
		"poison":                                "POiSON",
		"natosoft":                              "NATOsoft",
		"scd_dox":                               "SCD-Dox",
		"bs-enterprize":                         "BS Enterprize", //nolint:misspell
		"ralph-productions":                     "RalPh Productions",
		"unknown-couriers":                      "The Unknöwn Couriers",
		"wat-courier-crew":                      "WAT Courier Crew",
		"usalliance":                            "USAlliance",
		"acronym":                               "ACRONYMINIM",
		"powr":                                  "PoWR",
		"maim":                                  "MAiM",
		"relic":                                 "RELiC",
		"hipe":                                  "HiPE",
		"spectral":                              "Spec┼raL",
		"rpim":                                  "RPiM",
		"pri":                                   "PRi",
		"starjammers":                           "StarJammers",
		"mobius":                                "Möbius",
		"eclipse-ca":                            "Eclipse (CA)",
		"bom-squad":                             "BOM Squad",
		"wildsiderz":                            "WildSider",
		"motorsoft":                             "MotorSoft",
		"scorpion":                              "Scorpion ¥",
		"pmr-productions":                       "PMR Productions",
		"micropirates-inc":                      "MicroPirates Inc",
		"bad-association":                       "BAD Association",
		"the-underground-council":               "The UnderGround Council",
		"the-nameless-ones-1989":                "The Nameless Ones (1989)",
		"trc-ware-report":                       "TRC Ware Report",
		"backlash":                              "BackLash",
		"fogo":                                  "fOGO",
		"toss":                                  "ToSS",
		"xtc-systems-bbs":                       "XTC Systems BBS",
		"mci-escapes-bbs":                       "MCi Escapes BBS",
		"atlanta-pcug-bbs":                      "Atlanta PCUG BBS",
		"esp-headquarters-bbs":                  "ESP HeadQuarters BBS",
	}
	return &list
}

// Lowercase are a collection of styled names that use all lowercasing.
func Lowercase() []string {
	return []string{
		"intel",
		"mci-escapes",
		"scenet",
		"notwikipedia",
		"xpress",
	}
}

// Uppercase are a collection of styled names that use all uppercasing.
func Uppercase() []string { //nolint:funlen
	return []string{
		"lspd",
		"rise",
		"icch",
		"mash",
		"casa",
		"orpa",
		"arts",
		"acronym",
		"jake",
		"ytmar",
		"edge",
		"ameriboards",
		"nuke",
		"bbslst",
		"thhg",
		"2nd2none-bbs",
		"3wa-bbs",
		"acb-bbs",
		"anz-ftp",
		"beer",
		"bcp-bbs",
		"cusa",
		"ckc-bbs",
		"cnx-ftp",
		"core",
		"crsiso",
		"cwl-bbs",
		"dv8-bbs",
		"es-bbs",
		"dread",
		"fake",
		"fate",
		"fic-bbs",
		"hasp",
		"lkcc",
		"lms-bbs",
		"ls-bbs",
		"lsdiso",
		"lpc-bbs",
		"lta-bbs",
		"lube",
		"mor-ftp",
		"msv-ftp",
		"new-dtl",
		"nsdap",
		"nohk",
		"nos-ftp",
		"og-bbs",
		"okc-bbs",
		"pe*trsi*tdt",
		"petra",
		"pplk",
		"pmc-bbs",
		"pp-bbs",
		"ppps-bbs",
		"pox-ftp",
		"ps5b",
		"psi-bbs",
		"qed-bbs",
		"reno",
		"scum",
		"swag",
		"scf-ftp",
		"scsi-ftp",
		"shot",
		"tiw-bbs",
		"tbb-ftp",
		"tcsm-bbs",
		"tfz-2-bbs",
		"triad",
		"toads",
		"tog-ftp",
		"top-ftp",
		"tph-qqt",
		"tph-qqt-ftp",
		"trt-2001-bbs",
		"tsi-bbs",
		"tsc-bbs",
		"uct-bbs",
		"u4ea-ftp",
		"x_ess",
		"zoo-ftp",
		"phoenix",
		"sprint",
	}
}

const (
	spacedAmpersand = " & " // " & " is a special case
	spacedComma     = ", "  // ", " is a special case
)

// Specials is cache of the special styled names that is used by Path.String().
// The cache greatly improves benchmark performance.
var specials = *Special() //nolint:gochecknoglobals

// Special returns the list of styled names that use special mix or all lower or upper casing.
func Special() *List {
	list := make(List, len(*Names())+len(Lowercase())+len(Uppercase()))
	maps.Copy(list, *Names())
	maps.Copy(list, *Lower())
	maps.Copy(list, *Upper())
	return &list
}

// Lower returns the list of styled names that use all lowercasing.
func Lower() *List {
	list := make(List, len(Lowercase()))
	for value := range slices.Values(Lowercase()) {
		p := Path(value)
		s, _ := Humanize(p)
		list[p] = strings.ToLower(s)
	}
	return &list
}

// Upper returns the list of styled names that use all uppercasing.
func Upper() *List {
	list := make(List, len(Uppercase()))
	for value := range slices.Values(Uppercase()) {
		p := Path(value)
		x, _ := Humanize(p)
		list[p] = strings.ToUpper(x)
	}
	return &list
}

// Humanize deobfuscates the URL path and returns the formatted, human-readable group name.
// If the URL path contains invalid characters then an error is returned.
func Humanize(path Path) (string, error) {
	if !path.Valid() {
		return "", ErrInvalidPath
	}
	s := strings.ToLower(string(path))
	// the order of these expressions is critical
	// strings.replaceall is more performant than regex
	s = strings.ReplaceAll(s, "-ampersand-", spacedAmpersand)
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, "*", spacedComma)
	return s, nil
}

// Obfuscate formats the named string to be used as a URL path.
//
// Example:
//
//	string(Obfuscate("ACiD Productions")) = "acid-productions"
//	string(Obfuscate("Razor 1911 Demo & Skillion")) = "razor-1911-demo-ampersand-skillion"
//	string(Obfuscate("TDU-Jam!")) = "tdu_jam"
func Obfuscate(name string) Path {
	s := strings.TrimSpace(strings.ToLower(name))
	re := regexp.MustCompile(`[^a-z0-9\&\-\,\ ]`)
	s = re.ReplaceAllString(s, "")
	// the order of these expressions is critical
	// strings.replaceall is more performant than regex
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, spacedAmpersand, "-ampersand-")
	s = strings.ReplaceAll(s, spacedComma, "*")
	s = strings.ReplaceAll(s, " ", "-")
	return Path(s)
}

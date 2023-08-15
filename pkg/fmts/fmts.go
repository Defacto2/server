package fmts

import (
	"strings"

	"github.com/Defacto2/sceners/pkg/rename"
)

// URI is a the URL slug of the releaser.
type URI string

// Names is a map of releasers and their well-known styled names.
type Names map[URI]string

var names = Names{
	"acid-productions":                 "ACiD Productions",
	"core":                             "CoRE",
	"crackpl":                          "CrackPL",
	"dumptruck":                        "dumpTruck",
	"defacto2net":                      "defacto2.net",
	"dvniso":                           "DVNiSO",
	"esp-pirates":                      "ESP Pirates",
	"linezer0":                         "LineZer0",
	"icon":                             "iCON",
	"imars":                            "iMARS",
	"oneup":                            "OneUp",
	"orion":                            "ORiON",
	"mmi":                              "MMi",
	"mp2k":                             "MP2K",
	"nc_17":                            "NC-17",
	"risciso":                          "RISCiSO",
	"seek-n-destroy":                   "Seek 'n Destroy",
	"sma-posse":                        "SMA Posse",
	"software-pirates-inc":             "Software Pirates Inc.",
	"surprise-productions":             "Surprise! Productions",
	"razordox":                         "RazorDOX",
	"team-xtx":                         "Team XTX",
	"tft-team":                         "TFT Team",
	"tpinc":                            "TPiNC",
	"tristar-ampersand-red-sector-inc": "Tristar & Red Sector Inc.",
	"the-dvdr-releasing-standards":     "The DVDR Releasing Standards",
	"the-firm":                         "The FiRM",
	"thg-fx":                           "THG-FX",
	"tdu_jam":                          "TDU Jam!",
	"the-dream-team*tristar-ampersand-red-sector-inc": "The Dream Team + TRSi",
	"tport":   "tPORt",
	"underpl": "UnderPL",
}

// Name returns the well-known styled name of the releaser.
func Name(uri string) string {
	if _, ok := names[URI(uri)]; ok {
		return names[URI(uri)]
	}
	s := rename.DeObfuscateURL(uri)
	return strings.ReplaceAll(s, ", ", " + ")
}

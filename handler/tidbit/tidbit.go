package tidbit

import (
	"html/template"
	"slices"
	"strings"

	"github.com/Defacto2/releaser"
)

// URI is a the URL slug of the releaser.
type URI string

// ID is the identifier of the tidbit.
type ID int

// String returns the tidbit description.
func (id ID) String() string {
	if s := tidbits[id]; s != "" {
		return s
	}
	return ""
}

// URI returns the URIs of the tidbit.
func (id ID) URI() []URI {
	if x := groups[id]; x != nil {
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
	for _, u := range urls {
		if u == URI(uri) {
			continue
		}
		s := string(u)
		html = append(html, `<a href="/g/`+s+`">`+releaser.Link(s)+`</a>`)
	}
	s := strings.Join(html, " &nbsp; ")
	return template.HTML(s)
}

// Tibits is a map of tidbits mapped to their URIs.
type Tibits map[ID][]URI

// Tidbit is a map of tidbits mapped to their descriptions.
type Tidbit map[ID]string

var groups = Tibits{
	1:  []URI{"untouchables", "the-untouchables"},
	2:  []URI{"five-o", "boys-from-company-c", "the-firm"},
	3:  []URI{"fairlight", "united-software-association", "united-software-association*fairlight"},
	4:  []URI{"fairlight", "artists-in-revolt"},
	5:  []URI{"fairlight", "fairlight-dox"},
	6:  []URI{"aces-of-ansi-art", "acid-productions"},
	7:  []URI{"the-duplicators"},
	8:  []URI{"pirates-club-inc"},
	9:  []URI{"against-software-protection"},
	10: []URI{"software-pirates-inc"},
	11: []URI{"the-illinois-pirates"},
	12: []URI{"cracking-101", "national-elite-underground-alliance"},
	13: []URI{"esp-pirates", "esp-headquarters-bbs"},
	14: []URI{"silicon-valley-swappe-shoppe"},
	15: []URI{"five-o", "toads"},
	16: []URI{"c-ampersand-m", "boys-from-company-c"},
	17: []URI{"canadian-pirates-inc", "ptl-club"},
	18: []URI{"canadian-pirates-inc", "kgb", "ptl-club"},
}

var tidbits = Tidbit{
	1: "Untouchables were a famed US based game release group. The Untouchables were a 1990s scene group from Norway.",
	2: "Five-O and the BCC were a US based game release groups that merged in December 1988, " +
		"the next month in January they changed their name to The Firm, who were the first prolific PC group.",
	3: "Fairlight, founded on the Commodore 64 in 1987 is one of the oldest brands in the scene. " +
		"Fairlight PC first released in February 1991 but immediately collborated with USA to form the successful USA/FLT. " +
		"One of the first major busts in the USA on the PC scene forced USA to disband in February 1992.",
	4: "In 1992 Fairlight briefly ran the artgroup Artists In Revolt, sometimes referenced as the Fairlight Art Division. ",
	5: "Fairlight DOX (FLTDOX) was a sub-group of Fairlight that specialised in releasing documentation and trainers for games.",
	6: "Aces of ANSI Art is credited as one of the first art groups, that in mid-1990 reformed as the famous ANSi Creators in Demand, aka ACiD Productions.",
	7: "The Duplicators are the earliest game crackers on the PC that offer reliable dated releases.",
	8: "Pirates Club Inc from 1983 is the oldest known pirate group on the PC.",
	9: "ASP was a group that specialised in writing original and also resharing UNPROTECT text instructions for PC applications and later games. " +
		"These instructions were used to bypass disk copy protection and were shared on legitimate BBSes, Compuserve, and elite pirate boards.",
	10: "SPI is once of the oldest groups on the PC and one of the first enduring brands of the 1980s that created numerous custom utilities in addition to PC game releases.",
	11: "The Illinois Pirates release of their King's Quest walkthrough is the earliest known scene documentation for a PC exclusive game.",
	12: "Buckaroo Banzai was a prolific game hacker and trainer maker who was most famous for his 1980s cracking tutorials on the PC and Apple II." +
		" The series was republished and revised as the Ancient Art of Cracking for NUEA.",
	13: "ESP Pirates seems to be a label used by the cracker Mr Peace who was the sysop of the Phoenix based ESP Headquarters BBS." +
		" A number of the ESP releases were later repacked by Mr Peace to include advertisements for his BBS.",
	14: "Silicon Valley Swappe Shoppe looks to be a personal label used by Mr. Turbo who probably started on the Apple II and later moved onto the console scene.",
	15: "March 1986 saw both TOADS and Five-O release custom CGA loader screens for their game releases. " +
		"These screens were probably done in PC Paint from Mouse Systems and were the first known scene art for the PC. " +
		"It is unsure if the images were drawn by the credited crackers or by a separate artist.",
	16: "There is no information on C&M other than the credited cracker and CGA artist Zanna Martin would join BCC at the end of 1987.",
	17: "Canadian Pirates Inc were an early Ontario, Canada group that was active in the 1980s and often collaborated with PTL Club before eventually merging.",
	18: "According to the Church Chat Volume 4 text file, KGB was was an offshoot group formed by the merger of Canadian Pirates Inc and PTL Club but soon faded away.",
}

// Find returns the tidbit IDs for the given URI.
//
// The ID returned can be used in a string conversion to get the description.
// The ID can also be used to get the URIs of the tidbit.
func Find(uri string) []ID {
	ids := []ID{}
	for id, uris := range groups {
		for _, u := range uris {
			if u == URI(uri) {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

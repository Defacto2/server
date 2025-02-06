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
	if s := tidbits()[id]; s != "" {
		return s
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

func groups() Tibits {
	return Tibits{
		1:  []URI{"untouchables", "the-untouchables"},
		2:  []URI{"five-o", "boys-from-company-c", "pirates-r-us", "the-firm"},
		3:  []URI{"fairlight", "united-software-association", "united-software-association*fairlight"},
		4:  []URI{"fairlight", "artists-in-revolt"},
		5:  []URI{"fairlight", "fairlight-dox"},
		6:  []URI{"aces-of-ansi-art", "acid-productions"},
		7:  []URI{"the-duplicators"},
		8:  []URI{"pirates-club-inc"},
		9:  []URI{"against-software-protection"},
		10: []URI{"software-pirates-inc"},
		11: []URI{"the-illinois-pirates"},
		12: []URI{"cracking-101", "national-elite-underground-alliance", "buck-naked-productions"},
		13: []URI{"esp-pirates", "esp-headquarters-bbs"},
		14: []URI{"silicon-valley-swappe-shoppe"},
		15: []URI{"five-o", "toads"},
		16: []URI{"c-ampersand-m", "boys-from-company-c"},
		17: []URI{"canadian-pirates-inc", "ptl-club"},
		18: []URI{"canadian-pirates-inc", "kgb", "ptl-club"},
		19: []URI{"ptl-club", "sprint", "the-underground-council", "byte-bandits-bbs", "triad"},
		20: []URI{"new-york-crackers", "miami-cracking-machine", "international-network-of-crackers"},
		21: []URI{"public-domain"},
		22: []URI{"bentley-sidwell-productions", "the-firm"},
		23: []URI{"boys-from-company-c"},
		24: []URI{"fairlight"},
		25: []URI{"future-crew"},
		26: []URI{"international-network-of-crackers"},
		28: []URI{"the-firm", "mutual-assured-destruction", "public-enemy"},
		27: []URI{"the-firm", "swat", "national-underground-application-alliance", "fairlight"},
		29: []URI{"international-network-of-crackers", "triad"},
		30: []URI{"cmen"},
		31: []URI{"erkle"},
		32: []URI{"extasy", "xerox", "fairlight"},
		33: []URI{"norwegian-cracking-company", "international-network-of-crackers", "the-humble-guys"},
		34: []URI{"scd_dox", "software-chronicles-digest"},
		35: []URI{"software-chronicles-digest"},
		36: []URI{"the-humble-guys"},
		37: []URI{"netrunners", "minor-threat", "nexus"},
		38: []URI{"mai-review", "sda-review", "silicon-dream-artists"},
		39: []URI{"silicon-dream-artists"},
		40: []URI{"hype"},
		41: []URI{"alpha-flight", "outlaws", "storm-inc"},
		42: []URI{"thhg"},
		43: []URI{"tmh"},
		44: []URI{"the-racketeers"},
		45: []URI{"crackers-in-action"},
		46: []URI{"legion-of-doom"},
		47: []URI{"the-grand-council"},
		48: []URI{"untouchables", "uniq", "xap", "pentagram"},
		49: []URI{"italsoft"},
		50: []URI{"future-brain-inc", "the-humble-guys"},
		51: []URI{"pirate"},
		52: []URI{"creators-of-intense-art", "art-creation-enterprise"},
		53: []URI{"vla"},
	}
}

func tidbits() Tidbit {
	return Tidbit{
		1: `Confusingly, numerous groups used the name, "The Untouchables" or the initialism "UNT". ` +
			`"Untouchables" were a USA based release and trainer group. But there were 3 other "The Untouchables" or "UNT",<br>` +
			`1. A Dutch demo and trainer group from 1990+<br>` +
			`2. A UK based Atari ST group<br>` +
			`3. A Dutch PC release group from 1994-95`,
		2: "Five-O and BCC were a US based game release groups that merged in December 1988. " +
			"The next month they joined with Pirates R Us and changed their name to The Firm, who became the first prolific game release group on the PC.",
		3: "Fairlight, founded on the Commodore 64 in 1987 is one of the oldest brands in the scene. " +
			"Fairlight PC first released in February 1991 but immediately worked with USA to form the successful USA/FLT collboration. " +
			"Late January 1992 saw a major piracy bust in Detroit that forced USA to disband and Fairlight to go solo.",
		4: "In 1992 Fairlight briefly ran the artgroup Artists In Revolt, sometimes referenced as the Fairlight Art Division. ",
		5: "Fairlight DOX (FLTDOX) was a sub-group of Fairlight that specialised in releasing documentation and trainers for games.",
		6: "Aces of ANSI Art is credited as one of the first art groups. In mid-1990 they reformed as the ANSi Creators in Demand, and later known as ACiD Productions, the most prolific art group in the North American scene of the era.",
		7: "The Duplicators are the earliest game crackers on the PC that offer reliable dated releases.",
		8: "Pirates Club Inc from 1983 is the oldest known pirate group on the PC.",
		9: "ASP was a group that specialised in writing original and also resharing UNPROTECT text instructions for PC applications and later games. " +
			"These instructions were used to bypass disk copy protection and were shared on legitimate BBSes, Compuserve, and elite pirate boards.",
		10: "SPI is once of the oldest groups on the PC and one of the first enduring brands of the 1980s that created numerous custom utilities in addition to PC game releases." +
			"Their most celebrated tool was SnatchIT, which when combined with the commercial tool, " +
			`<a href="https://winworldpc.com/product/copy-ii-pc/2xx">Central Point Copy II PC</a> allowed (the then common) self-booting games to duplicated.`,
		11: "The Illinois Pirates release of their King's Quest walkthrough is the earliest known scene documentation for a PC exclusive game.",
		12: "Buckaroo Banzai was a prolific game hacker and trainer maker who was most famous for his 1980s cracking tutorials on the PC and Apple II." +
			" The series was republished and revised as the Ancient Art of Cracking for NUEA and under Buck Naked Productions.",
		13: "ESP Pirates seems to be a label used by the cracker Mr Peace who was the sysop of the Phoenix based ESP Headquarters BBS." +
			" A number of the ESP releases were later repacked by Mr Peace to include advertisements for his BBS.",
		14: "Silicon Valley Swappe Shoppe looks to be a personal label used by Mr. Turbo who probably started on the Apple II and later moved onto the console scene.",
		15: "March 1986 saw both TOADS and Five-O release custom CGA loader screens for their game releases. " +
			"These screens were probably done in PC Paint from Mouse Systems and were the first known scene art for the PC. " +
			"It is unsure if the images were drawn by the credited crackers or by a separate artist.",
		16: "There is no information on C&M other than the credited cracker and CGA artist Zanna Martin would join BCC at the end of 1987.",
		17: "Canadian Pirates Inc were an early Ontario, Canada group that was active in the 1980s and often collaborated with PTL Club before eventually merging.",
		18: "According to the Church Chat Volume 4 text file, KGB was an offshoot group formed by the merger of Canadian Pirates Inc and PTL Club but soon faded away.",
		19: "In November 1989, a number of groups including PTL Club, $print, The Underground Council, and the Byte Bandits merged and became Traid.",
		20: "INC was formed in September 1989 by the merger of the New York Crackers, ECA (currently unknown), and the Miami Cracking Machine. " +
			"NYC would leave a few months later after disagreements on the structure with the new group leaving MCM as the direct ancestor of INC.",
		21: "Fake public domain releases were a common tactic in the early-mid 1980s and used by pirates to distribute their warez on legitimate BBSes.",
		22: "BSP was a Texas-centric group, probably founded in 1988. The key members of BSP would join The FiRM in March 1989 and would occasionally " +
			`<a href="/f/ab2a1ce">be called</a> the <em>Ex-BSP division</em> in reference to this location. ` +
			"Some BSP releases from 1988 advertise the group as a division of Legions of Lucifer Inc. But this is not to be confused with the " +
			`<a href="https://textfiles.meulie.net/magazines/LOL/lol-20.phk">Legions of Lucifer</a> ` +
			"founded by Digitone Cypher in 1990 and became LoL-Phuck in 1991.",
		23: "BCC were founded in October 1987 as a game release group based in the state of Virgina.",
		24: "Fairlight PC only released games published on floppy disks. The group faded away the mid-1990s as the game industry moved exclusively to CD ROM, and piracy to CD RIPs. It wasn't until November 1998 that Fairlight returned as one of the earlier ISO groups releasing complete CD images of games. In the 2000s the Fairlight brand went in two unrelated directions, with a demoscene component seeing great success in that community that weirdly, was juxtaposed with a now criminal piracy group that was getting unwanted attention due to the rise of BitTorrent and sites like The Pirate Bay.",
		25: "Future Crew was the most famous demoscene group on the PC in the 1990s. The PC was primarily a business platform and games or multimedia were always secondary. But early demos by the Future Crew helped to change the mindset some, that the PC wasn't only for productivity and would become the platform of the future for general computing, gaming and multimedia." +
			"<br>The founding information on the Future Crew has been muddled over the years by incorrect and conflicting dates put out by the group themselves in different documents. Their first release, <strong>GR8</strong> came out in July 1989 and their second <strong>Yo!</strong> sometime in 1990.",
		26: "The first CD RIP was probably created by INC with their March 1992 release of " +
			`<a href="/f/aa209be">Battle Chess Multimedia</a>.` +
			" Though the packaging of the release was so jank that Fairlight felt the need to create their " +
			`<a href="/f/a91e0ae">own custom fix</a> to simplify the install process.`,
		27: "CyberChrist of SWAT briefly stole the FiRM brand in October 1993 for use as a game release group, while NUAA was to be used for productivity and utility software. " +
			"This was short-lived and a week later the game group became the USA based division of Fairlight PC.",
		28: "In August 1994, Public Enemy and MAD join under the unauthorized name of The FiRM, though this only lasted a few months.",
		29: `TRIAD went quiet in early 1990 but with some key members <a href="/f/a9229aa">turning up in INC</a>.`,
		30: "CMEN was a parody group that pretended to be Australian, but was run out of the Midnite Oil BBS in 214 (Dallas).",
		31: "According to BAD News #7, ERKle was a brief, pretend group created by The Pieman of The Humble Guys.",
		32: "Xerox was a German release group that rebranded as Extasy before merging with Fairlight.",
		33: "NCC would focus in Europe and collaborate with US groups for cracked software exchanges, first with INC, and eventually with The Humble Guys. ",
		34: "SCD-DoX was a shortlived documentation group created by the Software Chronicles Digest Magazine.",
		35: "For its first 5 issues, SCD stood for Southern Califoria Distribution and was part of a larger regional release group.",
		36: "According to a published retirement letter, The Humble Guys were founded on the 22 Jan 1990 and became the source of the first use of the _.nfo_ file extension. ",
		37: "Netrunners merged into the Minor Threat and Nexus collaboration at the end of October 1993. Minor Threat focused on applications and Nexus on game releases.",
		38: "SDA Review published by Silicon Dream Artists was DOS magazine that reviewed scene PC game releases. " +
			"The first 4 issues were under the name MAI Review, but the title was changed after the November 1991 merger of Masters of Abstractions and Illusions and MaD.",
		39: "Silicon Dream Artists was formed after the merging of Masters of Abstractions and Illusions and Maximized ANSi Designers in November 1991.",
		40: "HYPE from 1992 created elite BBS ads both in ANSI and as PC loaders. The brand was later reused by an unrelated warez release group in 1995.",
		41: "The German Alpha Flight started releasing on the Commodore 64 in at least 1986. The team joined the Amiga platform the following year. But on the PC, AFL didn’t release until many years later, with a team in Belgium waiting until Christmas 1992. The Belgium group changed brands to the Outlaws by mid-1993. And regular AFL releases didn’t occur until late 1993 under a new German lead team. But for reasons unknown, this group re-branded themselves as Storm Inc. However, AFL PC returned in 1994 under new, international membership being lead out of the USA and Australia.",
		42: "THHG may have been 2 separate groups led by Tom Tom from Germany, or maybe it is just the same group with numerous name changes. " +
			"But THHG has stood for The Hugo Husten Group and The Horrible Hackers from Germany.",
		43: "TMH was an early unprotect document writer, who may have been an individual or a group.",
		44: "The Racketeers were an Apple II pirate group from the early 1980s.",
		45: "Crackers in Action probably started as a personal brand for the cracker Live Wire based in Denver, but later became a national group.",
		46: "The Legion of Doom was a well known phreaking and hacker group (LOD/H) that was active in the 1980s and early 1990s, " +
			"with some of the earliest members having their start in the early Apple II piracy and phreaking community.",
		47: "The Grand Council a local 1980s release group from the <em>313</em> Detroit and Flint region of Michigan.",
		48: "Untouchables were founded on the 13th February 1993 with the joining of two groups, XAP and UNiQ. " +
			"The new release group lasted less than half a year before abandoning the name and reforming as Pentagram in July.",
		49: "Italsoft was an odd entity out of Argentina that would modify existing pirated games and then claim them as their own. " +
			"But oddly, they would often change the copyright notices in the documentation and software to fake the publisher and the release year.",
		50: "FBI were a Dutch group that created some of the first cracktros on the PC. In 1990 they would collaborate with THG to release European games in the USA.",
		51: "PIRATE was a USA text magazine that covered the Apple II and IBM PC scenes.",
		52: "CiA was an artgroup founded in July 1993 and in the following month, they doubled in size after absorbing ACE.",
		53: "VLA were an early PC programming and demogroup from the USA.",
	}
}

// Find returns the tidbit IDs for the given URI.
//
// The ID returned can be used in a string conversion to get the description.
// The ID can also be used to get the URIs of the tidbit.
func Find(uri string) []ID {
	ids := []ID{}
	for id, uris := range groups() {
		for _, u := range uris {
			if u == URI(uri) {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

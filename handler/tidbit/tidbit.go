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

func tidbits() Tidbit {
	return Tidbit{
		1: `Confusingly, numerous groups used the name, "The Untouchables" or the initialism "UNT". ` +
			`"Untouchables" were a USA based release and trainer group. But there were 3 other "The Untouchables" or "UNT",<br>` +
			`1. A Dutch demo and trainer group from 1990+<br>` +
			`2. A UK based Atari ST group<br>` +
			`3. A Dutch PC release group from 1994-95`,
		2: "Five-O and BCC were a US based game release groups that merged in December 1988. " +
			"The next month they joined with Pirates R Us and changed their name to The Firm, who became the first prolific game release group on the PC.",
		3: "<p>In an <a href=\"/f/a93540\">interview</a> with Insanity, Genesis discusses the founding of USA/Fairlight stating that it took just 12 hours from the initial idea to their first release which occurred on Thursday, 12th September 1991. The name <strong>United Software Association</strong>, he admits is a bit on then nose and rushed but they liked the catchy USA initialism.</p>" +
			"<p>The group management included The NotSoHumble Babe, Silencer and Genesis. Many of the initial members were formerly of <a href=\"/g/the-humble-guys\">The Humble Guys</a>, including The Humble Babe. However, when <a href=\"/f/a93ed4\">THG abruptly kicked out</a> a dozen or so people for no specific reason, a number of those and other members decided to leave for USA. And as a protest, The Humble Babe became The NotSoHumble Babe for the remainder of her time in the scene.</p>" +
			"<p>The very first and second USA release was a beta of the adventure game, <a href=\"/f/ac27bd8\">Spellcasting 201</a> and the title's documentation. Their third release was a freeware preview of <a href=\"/f/ad2a0e8\">Shadow Sorcerer</a>. These initial releases must have caused some bad feedback because in their following release, <a href=\"/f/b62d133\">Might and Magic 3</a>, a major CRPG series of the time, they state, <q>For those of you who though we were capable of only betas and demos, this release should convince you we're not some lame wannabe group - we're here to stay.</q></p>" +
			"<p>Here to stay they did, at least for the following number of months. USA rapidly became one of the most prolific PC cracking groups quickly overtaking their <a href=\"/f/b24832\">arch-rivals</a> The Humble Guys for the duration of their brief existence. " +
			"Yet this all came to a crashing halt at the end of January 1991 when The NotSoHumble Babe and The Grim Reaper (of <a href=\"/g/international-network-of-crackers\">INC</a>) were <a href=\"/f/aa20a3a\">arrested</a> for credit card fraud while meeting up in a Detroit car park. Back in 1991, software cracking and the sharing of commercial software wasn't criminal but using stolen or fake credit and telephone calling cards <a href=\"/f/a819b62\">definitely was</a> and usually involved The Secret Service or the FBI. These very public arrests <a href=\"/f/ab3a95\">caused panic</a> within the piracy scene in the United States and led to a number of people quitting and a week later the abrupt end to United Software Association.</p>" +
			"<p>The Fairlight connection of USA/Fairlight seems somewhat tenuous, as initially <strong>Fairlight PC</strong> had released <a href=\"/f/b32488a\">a single PC game</a> back in February of 1991. When they reemerged in September together as USA/Fairlight PC, the Fairlight side was to be the source of European published PC games. Yet it seems Fairlight PC had little impact, with only (Fairlight <a href=\"http://janeway.exotica.org.uk/target.php?idp=6375&idr=1940&tgt=1\">co-founder</a>) Strider generally being the sole member and the majority of releases of the cooperation being from the United States and not Europe.</p>" +
			"<p>Fairlight on the PC returned in March 1992 but without the “PC” suffix in its name. Much of the initial membership comprised of former members of USA, however with the new leadership comprising of Strider, Nemesis Enforcer and Riverndell sysop Trick Lord. Their <a href=\"/f/ac232a6\">return release</a> quietly mentions <q>please note that there is no more coop USA/FLT</q> " +
			" and boldly claims they're the world's oldest group, though that is not true.</p>",
		400: "In 1992 Fairlight briefly ran the artgroup Artists In Revolt, sometimes referenced as the Fairlight Art Division. ",
		500: "Fairlight DOX (FLTDOX) was a sub-group of Fairlight that specialised in releasing documentation and trainers for games.",
		6: "Aces of ANSI Art is credited as one of the first organized art groups in the elite BBS scene. " +
			"After becoming disorganized and demotivated, key members in mid-1990 reformed as the ANSi Creators in Demand, and later known as ACiD Productions, the most prolific art group of the era in the North American scene.",
		7: "The Duplicators are the earliest game crackers on the PC that offer reliable dated releases.",
		8: "Pirates Club Inc from 1983 is the oldest known pirate group on the PC.",
		9: "ASP was a group that specialised in writing original and also resharing UNPROTECT text instructions for PC applications and later games. " +
			"These instructions were used to bypass disk copy protection and were shared on legitimate BBSes, Compuserve, and elite pirate boards.",
		10: "SPI from Texas is once of the oldest groups on the PC and one of the first enduring brands of the 1980s that created numerous custom utilities in addition to PC game releases." +
			" Their most celebrated tool was SnatchIT, which when combined with the commercial tool, " +
			`<a href="https://winworldpc.com/product/copy-ii-pc/2xx">Central Point Copy II PC</a> allowed (the then common) self-booting games to duplicated.`,
		11: "The Illinois Pirates release of their King's Quest walkthrough is the earliest known scene documentation for a PC exclusive game.",
		12: "Buckaroo Banzai was a prolific game hacker and trainer maker who was most famous for his 1980s cracking tutorials on the PC and Apple II." +
			" The series was republished and revised as the Ancient Art of Cracking for NUEA and under Buck Naked Productions.",
		13: "ESP Pirates seems to be a cracking brand used by the cracker Mr Peace who was the sysop of the Phoenix based ESP Headquarters BBS." +
			" A number of the ESP releases were later repacked by Mr Peace to include advertisements for his pirate BBS that started in 1987.",
		14: "Silicon Valley Swappe Shoppe looks to be a personal brand used by Mr. Turbo who probably started on the Apple II and later moved onto the console scene.",
		15: "March 1986 saw both TOADS and Five-O release custom CGA loader screens for their game releases. " +
			"These screens were probably done in <a href=\"https://winworldpc.com/product/pc-paint/100a\">PC Paint</a> from Mouse Systems and were the first known scene art for the PC. " +
			"It is unsure if the images were drawn by the credited crackers or by a separate artist.",
		16: "There is no information on C&M other than the credited cracker and CGA artist Zanna Martin would join BCC at the end of 1987.",
		17: "Canadian Pirates Inc were an early Ontario, Canada group that was active in the 1980s and often collaborated with PTL Club from Illinois before eventually merging.",
		18: "According to the Church Chat Volume 4 text file, KGB was an offshoot group formed by the merger of Canadian Pirates Inc and PTL Club but soon faded away.",
		19: "In November 1989, a number of groups including PTL Club, $print, The Underground Council, and the Byte Bandits merged and became Traid.",
		20: "INC was formed in September 1989 by the merger of the New York Crackers, ECA (currently unknown but maybe <em>Elite Crackers Association</em>), and the Miami Cracking Machine. " +
			"NYC would leave a few months later after disagreements on the structure with the new group leaving MCM as the direct ancestor of INC.",
		21: "Fake public domain releases were a common tactic in the early-mid 1980s and used by pirates to distribute their warez on legitimate BBSes.",
		22: "BSP was a Texas-centric group, probably founded in 1988. The key members of BSP would join The FiRM in March 1989 and would occasionally " +
			`<a href="/f/ab2a1ce">be called</a> the <em>Ex-BSP division</em> in reference to this location. ` +
			"Some BSP releases from 1988 advertise the group as a division of Legions of Lucifer Inc. But this is not to be confused with the " +
			`<a href="https://textfiles.meulie.net/magazines/LOL/lol-20.phk">Legions of Lucifer</a> ` +
			"founded by Digitone Cypher in 1990 and became LoL-Phuck in 1991.",
		23: "BCC were founded in October 1987 as a game release group based in the state of Virgina.",
		24: "<p>1992 saw lots of success for <strong>Fairlight</strong> and the group <a href=\"/f/b42ec96\">ballooning</a> with a large membership including many former members of <a href=\"/g/international-network-of-crackers\">INC</a>, " +
			"and the US side of the group being run by Ford Perfect. Yet thanks to some <a href=\"/f/b528606\">immature drama</a> at his instigation, by years end the group <a href=\"/f/b0411d\">collapsed</a>, with many parting ways to form <a href=\"/g/sinister\">Sinister</a>. " +
			"Ford Perfect continued on with the name for a brief time before <a href=\"/f/a72d0b#:~:text=Ford Perfect just  may have left\">possibly quitting</a> the scene and finally ending the group.</p>" +
			"<p>All this must have been to the dismay of Strider who restored the group in March 1993 with <a href=\"/f/b047d2\">a release</a> of the sequel to one of the best reviewed microcomputer games of all time, Lemmings. This new Fairlight was tiny in comparison to the one from the previous year and only comprised of Swedish members. In this first release they state <q>Time to focus on Quality, and bring the honor to the name FairLight on PC again.</q></p>" +
			"<p>Fairlight on the PC was a cracking group that only released games published onto <a href=\"/f/b52d81d\">floppy disks</a>. While not unusual, this narrow scope caused the group faded away before 1996 as the game industry moved exclusively to CD ROM, and piracy to CD RIPs. It wasn't until November 1998 that <a href=\"/f/ac2be5\">Fairlight returned</a> with JBM and Holy Beast, as one of the earlier <strong>ISO groups</strong> releasing complete CD (and eventually DVD) images of games.</p>" +
			"<p>In the 2000s the Fairlight brand went in two unrelated directions, with the legitimate <strong>Demoscene</strong> <a href=\"/f/ab3caf\">component</a> seeing great success in that community. Which awkwardly, was <a href=\"/f/ac33f8\">juxtaposed</a> with a <a href=\"https://www.copyright.gov/docs/2265_stat.html\">now criminal</a> piracy group that was getting unwanted attention due to the rise of BitTorrent and sites like The Pirate Bay." +
			" March 2011 saw <a href=\"/f/ad4991\">1,000 ISO releases</a> under Fairlight and its 25th anniversary in 2012 with both unrelated Demoscene and piracy activities going strong.</p>",
		25: "Future Crew was the most famous demoscene group on the PC in the 1990s. The PC was primarily a business platform and games or multimedia were always secondary. But early demos by the Future Crew helped to change the mindset some, that the PC wasn't only for productivity and would become the platform of the future for general computing, gaming and multimedia." +
			"<br>The founding information on the Future Crew has been muddled over the years by incorrect and conflicting dates put out by the group themselves in different documents. Their first release, <strong>GR8</strong> came out in July 1989 and their second <strong>Yo!</strong> sometime in 1990.",
		26: "The first CD RIP was probably created by INC with their March 1992 release of " +
			`<a href="/f/aa209be">Battle Chess Multimedia</a>.` +
			" Though the packaging of the release was so jank that Fairlight felt the need to create their " +
			`<a href="/f/a91e0ae">own custom fix</a> to simplify the install process.`,
		27: "CyberChrist of SWAT briefly stole the FiRM brand in October 1993 for use as a game release group, while NUAA was to be used for productivity and utility software. " +
			"This was short-lived and a week later the game group became the United States side of Fairlight, but again, only for a very brief time.",
		28: "In August 1994, Public Enemy and MAD join under the unauthorized name of The FiRM, though this only lasted a few months.",
		29: `TRIAD went quiet in early 1990 but with some key members <a href="/f/a9229aa">turning up in INC</a>.`,
		30: "CMEN was a parody group that pretended to be Australian, but was run out of the Midnite Oil BBS in 214 (Dallas).",
		31: "According to BAD News #7, ERKle was a brief, pretend group created by The Pieman of The Humble Guys.",
		32: "Xerox was a German release group that rebranded as Extasy before merging with Fairlight.",
		33: "NCC would focus in Europe and collaborate with US groups for cracked software exchanges, first with INC, and eventually with The Humble Guys. ",
		34: "SCD-Dox was a shortlived documentation group created by Software Chronicles Digest Productions.",
		35: "SCD started out as <strong>Southern California Distribution</strong> and was part of a larger regional release group. As such, in November 1990 they published a newsletter for DOS known as <strong>The SCD Report</strong> to inform the pirate community. Over a span of two years, 14 issues of The SCD Report were published and the magazine was well received. However, December 1992 saw some changes, with not only seeing a rewrite of the magazine application, user interface and a re-branding to Software Chronicles Digest, but also it became the final issue.",
		36: "<p>According to Fabulous Furlough's <a href=\"/f/b144a1\">retirement letter</a>, The Humble Guys were founded on the 22 Jan 1990.</p>" +
			"<p>The Humble Guys had a bit of a reputation, partly because they were <a href=\"/f/a93245\">outspoken</a> plus were quite happy to give <a href=\"https://wayback.defacto2.net/the-scene-news-from-1999-september-14/interview-002.html\">interviews</a> or write down their thoughts. And if you believe what is sometimes said, there was virtually no Scene on the PC until they came around and revolutionised the community using their hard earned experience from the competitive Commodore online communities. Personally, I think this view is over sensationalised, and the Scene on the PC was fine before THG. Rather, people online using PCs in the 1980s, were possibly older, cared less about games than on the other microcomputers, so were more relaxed.</p>" +
			"<p>But in saying that, THG really ramp up the release schedule for games and made it a more aggressive and competitive community, for good or ill. " +
			"Many would argue they were the major force throughout 1990 and much of 1991. But became a shell of themselves after a number of people left to form <a href=\"/g/united-software-association*fairlight\">United Software Association</a> (USA) in September and later after more people quit the scene due to the <a href=\"/f/ab232ca\">late January 1992 public arrests</a> of two notable Scene personalities.</p>" +
			// nfo extension
			"<p>We believe THG is the source of the <strong><code>.NFO</code></strong> file extension " +
			"with the minimal textfile <code>KNIGHTS.NFO</code> found in the release of <a href=\"/f/ab3945\">Knights of Legend</a> from the 23rd of January 1990. " +
			"Knights of Legend required <a href=\"/f/b02d22e\">a fix</a> due to <a href=\"/f/b228b1a\">an installation quirk</a> and maybe that is why it is forgotten. " +
			"The following day, the release of <a href=\"/f/aa24d74\">White Death</a> would include a tiny <code>WHITEDET.NFO</code> textfile and so too <code>STUNT.NFO</code> for <a href=\"/f/aa1e84e\">Stunt Track Racer</a> on the 25th. " +
			"By early February the <code>BUBBLE.NFO</code> textfile for <a href=\"/f/ab1eca6\">Bubble Bobble</a> had become bloated, with additional group greets, a big yahoo, multiple boards, a Nashville voicemail phone number and PO BOX." +
			"</p><p>A 1991 backup from the HMS Bounty BBS had the following listings, but note, these 1990 uploads were before the era of <em>0-day</em> wares:</p>" +
			"<pre>" +
			"FAERY-1.ZIP  246086 09/18/89 Faery Tale Adventure.  1 of 3\n" +
			"KNIGHTS0.ZIP 289790 01/25/90 Knights of Legend - Fabulous Furlough's new group - The Humble Guys - 1 of 6\n" +
			"WHTEDTH1.ZIP 128461 01/25/90 White Death - The Humble Guys - 1 of 2\n" +
			"STUNT1.ZIP   268403 01/25/90 Stunt Track Racer - 1 of 2\n" +
			"KOLNSTAL.ZIP    458 01/27/90 Knights of Legend install info -needed\n" +
			"TRUMPS1.ZIP  163611 01/30/90 Trumps Castle - The Humble Guys - 1 of 2\n" +
			"DEJAEGA1.ZIP 173387 01/30/90 Deja Vu II - The Humble Guys - 1 of 2\n" +
			"AJAX1.ZIP    209468 02/01/90 Ajax from The Humble Guys | Uploaded by: Sysop\n" +
			"TERRNHG1.ZIP 166340 02/03/90 Terrain Editor for Sim City from The Humble Guys - 1 of 2 | Uploaded by: Sysop\n" +
			"BUBBLE1.ZIP  262079 02/08/90 Bubble Bobble - The Humble Guys - 1/2\n" +
			"BUBBLCHT.ZIP  42996 02/11/90 Bubble Bobble cheat | Uploaded by: Grebo Guru\n" +
			"CRMEWAV1.ZIP 618772 02/20/90 CRIME WAVE - The Humble Guys - 1 of 4 | cracked\n" +
			"                             2/20/90 by Fabulous Furlough. | Check here on\n" +
			"                             the Bounty for ALL cracks and | releases from\n" +
			"                             The Humble Guys\n" +
			"1989STAT.ZIP  82212 02/23/90 1989 Stats for Earl Weaver - The Humble Guys! | Uploaded by: Fabulous Furlough</pre>",
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
		47: "The Grand Council were a local 1980s release group from the <em>313</em> Detroit and Flint region of Michigan.",
		48: "Untouchables were founded on the 13th February 1993 with the joining of two groups, XAP and UNiQ. " +
			"The new release group lasted less than half a year before abandoning the name and reforming as Pentagram in July.",
		49: "Italsoft was an odd entity out of Argentina that would modify existing pirated games and then claim them as their own. " +
			"But oddly, they would often change the copyright notices in the documentation and software to fake the publisher and the release year.",
		50: "FBI were a Dutch group that created some of the first cracktros on the PC. In 1990 they would collaborate with THG to release European games in the USA.",
		51: "PIRATE was a USA text magazine that covered the Apple II and IBM PC scenes.",
		52: "CiA was an artgroup founded in July 1993 and in the following month, they doubled in size after absorbing ACE.",
		53: "VLA were an early PC programming and demogroup from the USA.",
		54: "The North West Connection was a local group from Washington state, aka the Pacific North West.",
		55: "The Sysops Association Network was collective of elite BBS sysops that exchanged information " +
			"and forums on the latest bulletin board news and technical goings. " +
			"The organisation at some stages grew quite large until the Internet abruptly made it and the member boards redundant. " +
			"At times TSAN claimed it was founded in 1984, but whether this was a single BBS or an actual organisation is unknown.",
		56: "Based in California, American Pirate Industries was an early or possibly first example of a text magazine for the PC BBS Scene.",
		57: "Pirates Sick of Initials formed as a late 1980s cracking group but after the competition became to competitive in early 1990, " +
			"they <a href=\"/f/ad1f4f6\">transitioned</a> into a games utility group releasing cheats, documentation, fixes and tools.",
		58: "Byte Bandits was a lose collection of sysops from North America that supplied games to Sam Brown's Californian BBS, the Byte Bandits for cracking and redistribution.",
		59: "Sorcerers were a Finnish group of teenage programmers who created some of the earliest demos on the PC." +
			" However, this was years after the intros and demos were commonplace on other microcomputer platforms.",
		60: "Katharsis was an Amiga cracking and demo group from Poland that expanded onto both the PC and into the ASCII artscene.",
		61: "NEUA were originally from New York state and went by the name North Eastern Underground Alliance but changed to National Elite in November 1990.",
		62: "Public Enemy probably formed in Janurary 1990 or in 1989 around Montreal and for its first iteration identified as a Canadian group. " +
			"They frequently collaborated with fellow Canadians NASA, the German Amiga group Red Sector Inc. and The Dream Team from Sweden but stopped releasing in mid-1991. ",
		63: "The second iteration of Public Enemy came about out of the USA in February 1993 and lasted until year's end, but with the Blade Runner (514) returning from the original Canadian group.",
		64: "<p>Razor 1911 was a cracking group that was founded in Norway in 1985 on the Commodore 64. " +
			"Some of the founding members jumped to the PC in mid-1991 and unlike other European groups, Razor heavily focused on North American releases. " +
			"For much of 1992, the group expanded and with the exception of their <a href=\"/f/aa4ba1\">7th-anniversary release</a>, they often didn't list members, only couriers and boards. However the growth in the group can be seen in the number of BBS affiliations.</p>" +
			"<p>A comprehensive member list was returned to the May 1993 release of <a href=\"/f/b126c88\">SiLk!</a>, but within a week it got <a href=\"/f/ad1d0c8\">abruptly censored</a>, probably due to unrelated European busts. " +
			"Before finally <a href=\"/f/a8222bc\">listing a reduced membership</a> but with useful roles for the group. However, the original Norwegian members faded into the background and references to the Razor's European origins were removed.</p>" +
			"<p>March 1994 saw the release of the <strong>first demo</strong> by Razor on the PC, named <a href=\"/f/ab445e\">RED</a>. " +
			"It was well received and saw a new <em>Razor 1911 Coding/Training/Art Departments</em> in the <a href=\"/f/af2e8ab\">NFO listings</a> and the return of the PO Box in Norway. " +
			"The famous text Razor 1911 logo by &ltJED> of ACiD was prototyped in <a href=\"/f/b31a533\">Battle Isle 2</a> and revised into its final form for the release of <a href=\"/f/b31a533\">Doom 2</a> in August.</p>" +
			"<p>The start of 1995 saw the outspoken The Renegade Chemist leave Razor and <a href=\"/f/ab3a82\">attempt to kill the group</a> as he departed, but apparently he forgot to get consensus and <a href=\"/f/af1aaf1\">Razor happily continued</a>." +
			" Yet he saw a couple of issues including a lack of money to obtain new games and the reduced number of game releases being published to floppy disks which Razor 1911 worked with." +
			" To raise funds, and decades before the mainstream influencers did the same, Razor 1911 created, advertised and sold merch directly to its online fans. " +
			"Initially, they <a href=\"f/a82163f\">shipped a t-shirt</a> that sold well, but later in the year, they developed a commemorative CD ROM with a large collection of their PC and Amiga releases! " +
			"And to solve the dwindling supply of games being published on floppy disks, Razor <a href=\"/f/ad4a55\">released Tyran</a>, the full compact disc under the new <strong>CD Division</strong> label.</p>",
		65: "<p>While founded in 1985 in North America, by 1990, Red Sector were mostly a European cracking and demo group releasing both on the Commodore 64 and Amiga. " +
			"Their 1990 entry onto the PC was solely wares related, but they <a href=\"/f/b01d2f2\">lacked an experienced PC cracker</a> and would team with existing groups to supply titles and have them crack and release under a cooperation. This collaboration was mostly done with Public Enemy from Canada.</p>" +
			"Over in Europe, Red Sector and the German group Tristar decided to merge mid-year forming the famous Tristar & Red Sector Inc. or TRSi. However, this new branding wasn't reflected on the PC until months later in December with a comical collaboration of five groups and " +
			"the release of <a href=\"/f/a74ac6\">4D Sports Driving presented by</a> Public&nbsp;Enemy&nbsp;/&nbsp;TRSi&nbsp;/&nbsp;Defjam&nbsp;/&nbsp;The Dream Team.",
		66: "Most of 1991 would see TRSi working with the PC in a group collaboration. Initially with Public Enemy and the Swedish based The Dream Team together. But then exclusively with TDT under a TDT / TRSi brand that often got referred to as The Cooperation or the coop. " +
			"That was until September when it was publicly <a href=\"/f/af2c09f\">announced by TDT</a> that this cooperation “is now broken”. And within a month, TDT were opting to instead co-release with fellow Europeans, Skid Row.",
		67: "The remainder of the early 1990s saw TRSi mostly releasing European published titles despite wanting to pivot to to North America where the AAA tier PC development was occurring. Unfortunately in Europe, the PC was a secondary platform for many game developers, often used to dump quick ports of games developed for the Commodore or Atari microcomputers." +
			" And this meant the quality of TRSi brand suffered, though there were <a href=\"/f/ae466d\">exceptions</a>.",
		68: "<p>The Dream Team was based in Sweden and was a name frequently associated with its founder, Hard Core. A programmer who had somewhat of an early reputation for creating second-rate intros on the PC on behalf of the group. Though, we do credit him with creating the first contemporary PC cracktro with this great <a href=\"/f/b01ca10\">musical skull intro</a> from November 1991. As a side note, his intros are the most annoying pieces of shit to preserve, as they have numerous dependencies that were intentionally hidden in releases. If an intro on this page doesn't work right, it is probably missing a hidden file or two.</p>" +
			"The Dream Team enjoyed success and remained active for much of the early 1990s, despite the occasional <a href=\"/f/b5261b0\">hiccup</a>, until quietly retiring at the end of 1993. " +
			"With only the occasional <a href=\"/f/a93193\">guest release</a>, the TDT name disappeared from the scene for 16 months before <a href=\"/f/b72e5db\">returning</a> in 1995 and tried rebuilding itself from the ground up with only four members. But this second attempt wasn't as successful and the group for the most part stopped releasing after four months, and completely vanished in November.",
		69: "<p>ROM 1911 was founded by The Renegade Chemist and Zeus as a dumping ground for PC CD titles that were becoming common place in 1994. At the time Razor 1911 had always been a cracking group that removed (floppy) disk copy protection and the newer CD titles were out of its scope. Zeus had joined Razor in August with the supply of <a href=\"/f/b31a533\">Doom 2</a> and possibly had fast access to PC CD titles that were looking for release. The first iteration of the group was presented as <em>ROM 1911 : Razor 1911 CD-ROM Division</em>.</p>" +
			"<p>It is assumed the group died when The Renegade Chemist had <a href=\"/f/ab3a82\">quit</a> Razor at the end of January 1995. However, months later it was <a href=\"/f/a839dd\">restarted</a> as <em>ROM 1911</em> by Malicious Intent as another small, three member, one board group.  The group ballooned though and in late July, Malicious Intent <a href=\"/f/aa3e2d\">wound it up</a>, and in annoyance with the Razor leadership, quit to start a new group known as RETRiBUTiON.</p>",
		70: "<p>According to a 2008 High Society <a href=\"https://www.high-society.at/history.php\">history page</a>, the group name came about in 1996. However the team were previously known as The Future Boys, who were founded in Austria on the Commodore 64 way back in 1986. Some members of the group later released on the Commodore Amiga as the Home Boys after the Commodore 64 as a viable platform faded from relevance. The Future Boys brand returned in 1995 but moved away from its microcomputer roots and instead released console software for the SNES. It was this entity that renamed itself in 1996 to High Society, and the following year exclusively focused on the PC and consoles such as the N64 and PlayStation.</p>" +
			"<p>The same website mentions that put together, the members have programmed on a huge variety of platforms including the Commodore 16, 64, Amiga; NeoGeo; NES, SNES Gameboy, N64; PC Engine; PlayStation; Dreamcast; PC and cellular phones (pre-iOS/Android).</p>",
		71: "<p>Trinity Reviews were a long running PC-release review group that might have first appeared late 1992 as a subsection of the scene magazine Lancelot 2, <a href=\"/f/a57c2\">issue 3</a>. " +
			"The reviews continued on for years, long after Lancelot 2 stopped publishing and were mostly comprised of an ANSI text file, a viewer and sometimes some attach screenshots.</p>",
		72: "The Real Pirates Guide was a document shared around BBSes of the mid-1980s. While written for the Apple II Scene, it was more broadly applicable to any online Scene of that era. And mostly it is an attempt to get kid-pirates to act more mature while online.",
		73: "According to the introduction, The Amatuer Crackist Tutorial from 1988 was created by PTL Club during a time when there was a shortage of crackers on the PC. It’s a tutorial series that is designed to educate readers new to the basics of software cracking and using unprotect text documents.",
		74: "Church Chat was a newsletter published by PTL Club covering the group activities. The name PTL Club was taken from a <a href=\"https://www.imdb.com/title/tt0125638/\">televangelist talk show</a> and the newsletter’s title also reflects this.",
		75: "Corrupted Programming’s newsletter was an attempt to promote the worthiness of computer viruses and software Trojans as viable topic for the underground online communities of the time.",
		76: "The Official Unprotection Scheme Library were combinations of unprotect documents compiled and republished in a newsletter. Unprotects are texts describing the methods and techniques needed to crack a specific software title.",
		77: "The Elemental's Pirate BBS List was a large collection numbers to pirate bulletin boards online in the USA in 1989. The lists seem to be specific to the PC pirate scene and includes details such as the software in use by the board (Soft), the speed (Baad) and sometimes group affiliations (Comment).",
		78: "Before the era of the world wide web and online news media, knowledge of the release dates for most games were kept to industry, such as software retailers, distributors and game magazines. Using these closed industry sources, Claude Rains RELEASES texts would compile lists of new game releases on PC and their expected release date. Information that was useful for game release groups who often still relied on retailers for software supply.",
		79: "While not Scene related, GIF News is a curiosity of its time. It tries to create a digital game-review magazine using the formatting of a traditional print magazine in providing varied layouts, screenshots, color and multiple fonts in a time before the world wide web and multimedia PC software.",
		80: "While The Mickey Mouse Club was a cracking group, it published the hack and phreak newsletter Hackers Unlimited.",
		81: "Bounty Hunter's National Pirate List was a large list of PC (and occasional Amiga) pirate or warez bulletin boards from the USA in 1990.",
		82: "The Phreakers Handbook was a compilation of texts about phreaking written by other people and probably reused without permission of the original authors. This compilation is the sort of text that would get replaced by subject specific websites of the 1990s.",
		83: "Spectrum was a review magazine for the PC Scene during 1990 and 1991. Scene game releases, cracking groups, and bulletin boards were all review subjects.",
		84: "The Pirate World published by The Pirate Syndicate was a piracy Scene newsletter from 1990 that often reviewed groups of bulletin boards from a specific telephone area code region as well as news and articles on the scene.",
		85: "<p>In a December 1988 <a href=\"http://janeway.exotica.org.uk/target.php?idp=6375&idr=1940&tgt=1\">interview</a>, Fairlight co-founder Strider states his age was 18, and got serious about piracy in 1985. He goes on to say Fairlight was founded on the Commodore 64 and Amiga microcomputer platforms in March 1987 by himself, Gollum and Black Shadow.</p>" +
			"<p>In an April 2008 <a href=\"https://alt.politics.republicans.narkive.com/I7xN7Xnp/san-diego-gop-chairman-co-founded-international-piracy-ring\">article</a> (original is lost) published on <a href=\"https://www.rawstory.com/\">Raw Story</a>, it was revealed that the San Diego GOP (Republican Party) chairman was indeed Strider, who had immigrated to the US 1992 on a H1B visa and became a US citizen in 2003. <q>Online research reveals that Krvaric is the co-founder of Fairlight, a band of software crackers which later evolved into an international video and software piracy group that law enforcement authorities say is among the world's largest such crime rings.</q></p>" +
			"<p>More controversy occurred in August 2020, when the local San Diego PBS television affiliate <a href=\"https://www.kpbs.org/news/evening-edition/2020/08/21/video-surfaces-images-hitler-and-tony-krvaric\">reported</a> on the Fairlight demo <a href=\"https://www.youtube.com/watch?v=X6SS8TE6c4o\">Space Age</a>, released for the Commodore Amiga <a href=\"https://demozoo.org/productions/243679/\">in 1989</a> that displays a few photos of Strider, and some other members plus an animated Hitler sprite bouncing around the screen. Strider's individual photo appears 25 seconds into the demo with the text, <q>Kill a commie, coz here's Strider!</q>, " +
			"though <a href=\"https://www.thesun.co.uk/news/14342020/when-prince-harry-nazi-uniform-why-apology/\">stupid</a>, this probably was tounge in cheek.</p>",
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

package app

const notable = "Notable groups of"

// Milestone is an accomplishment for a year and optional month.
type Milestone struct {
	Year      int     // Year of the milestone.
	Month     int     // Month of the milestone.
	Day       int     // Day of the milestone.
	Prefix    string  // Prefix replacement for the month, such as 'Early', 'Mid' or 'Late'.
	Title     string  // Title of the milestone should be the accomplishment.
	Lead      string  // Lead paragraph, is optional and should usually be the product.
	Content   string  // Content is the main body of the milestone and can be HTML.
	Link      string  // Link is the URL to an article about the milestone or the product.
	LinkTitle string  // LinkTitle is the title of the Link.
	List      Links   // Links is a collection of links that are displayed as a HTML list.
	Highlight bool    // Highlight is a flag to outline the milestone.
	Picture   Picture // Picture is an image or screenshot for a milestone.
}

// Picture is an image or screenshot for a milestone.
type Picture struct {
	Title       string // Title of the picture.
	Alt         string // Alt is the alternative text for the picture.
	Attribution string // attribution is the name of the author of the picture.
	License     string // License is the license of the picture.
	LicenseLink string // LicenseLink is the URL to the license of the picture.
	Webp        string // Webp is the filename of the WebP screenshot.
	Png         string // Png is the filename of the PNG screenshot.
	Jpg         string // Jpg is the filename of the JPG photo.
}

// Links is a collection of Links.
type Links []struct {
	LinkTitle string // LinkTitle is the title of the Link.
	Link      string // Link is the URL to an article about the milestone or the product.
	Forward   string // Forward is an optional name of a group that is prefixed before the link to indicate a merger.
}

// Milestones is a collection of Milestone.
type Milestones []Milestone

// Len is the number of Milestones.
func (m Milestones) Len() int {
	return len(m)
}

func ByDecade1970s() Milestones {
	m := []Milestone{
		{
			Year: 1971, Month: 10, Title: "Secrets of the Little Blue Box", Highlight: true,
			Lead: "Esquire October 1971", LinkTitle: "the complete article",
			Link: "https://www.slate.com/articles/technology/the_spectator/2011/10/the_article_that_inspired_steve_jobs_secrets_of_the_little_blue_.html",
			Content: "Ron Rosenbaum writes the first mainstream article on phone freaks, primarily kids who'd hack and experiment with the global telephone network.<br>" +
				"The piece coins them as phone-<strong>phreaks</strong> and introduces the reader to the kids' use of <strong>pseudonyms</strong> or codenames within their regional <strong>groups</strong> of friends. " +
				"It gives an early example of <strong>social engineering</strong>, defines the community of phreakers as the phone-phreak <strong>underground</strong>, and mentions the newer trend of <strong>computer phreaking</strong>, which we call <u>computer hacking</u> today.",
		},
		{
			Year: 1971, Month: 11, Day: 15, Title: "The first microprocessor",
			Lead: "Intel 4004", LinkTitle: "The Story of the Intel 4004",
			Link:    "https://www.intel.com/content/www/us/en/history/museum-story-of-intel-4004.html",
			Content: "Intel advertises the first-to-market general-purpose programmable processor or microprocessor, the 4-bit Intel 4004.",
		},
		{
			Year: 1972, Month: 4, Title: "The first 8-bit microprocessor",
			Lead: "Intel 8008", LinkTitle: "The Story of the Intel 8008",
			Link:    "https://www.intel.com/content/www/us/en/history/museum-story-of-intel-8008.html",
			Content: "Intel releases the world's first 8-bit microprocessor, the Intel 8008.",
			Picture: Picture{
				Title:       "Intel 8008 CPU chip",
				Alt:         "A photo of an Intel C8008-1 CPU chip.",
				Jpg:         "intel-8008.jpg",
				Attribution: "Konstantin Lanzet",
				License:     "CC BY-SA 4.0",
				LicenseLink: "https://creativecommons.org/licenses/by-sa/4.0/",
			},
		},
		{
			Year: 1972, Prefix: "Early", Title: "Blue boxes",
			Link: "https://explodingthephone.com/", LinkTitle: "about the hackers of the telephone network",
			Content: "Inspired by The Secrets of the Little Blue Box article, Steve Wozniak and a teenage Steve Jobs team up to build and sell 40-100, Wozniak-designed blue boxes to the students of Berkeley University. " +
				"The devices allowed users to hack and manipulate the electromechanical machines that operated the national telephone network.",
		},
		{
			Year: 1974, Month: 4, Title: "The first CPU for microcomputers",
			Lead: "Intel 8080", LinkTitle: "about The Intel 8008 and 8080",
			Link: "https://www.intel.com/content/www/us/en/history/museum-story-of-intel-8008.html",
			Content: "Intel releases the 8-bit 8080 CPU, its second but more successful 8-bit programmable microprocessor. " +
				"This CPU became the processing heart of the earliest popular microcomputers, the Altair 8800, the Sol-20 and the IMSAI.",
		},
		{
			Year: 1975, Month: 1, Title: "The first popular microcomputer",
			Lead: "Altair 8800", LinkTitle: "about the Altair 8800",
			Link:    "https://americanhistory.si.edu/collections/search/object/nmah_334396",
			Content: "The worlds first popular microcomputer appears on the front cover of Popular Electronics in the USA, the Altair 8800 by MITS running an Intel 8080 CPU.",
		},
		{
			Year: 1975, Month: 2, Title: "The first microcomputer software",
			Lead: "Altair BASIC", LinkTitle: "about origins of BASIC",
			Link: "https://time.com/69316/basic/",
			Content: "Paul Allen and Bill Gates program and sell Altair BASIC for the computer they first saw a month prior. " +
				"BASIC (Beginner's All-Purpose Symbolic Instruction Code) was a programming language conceived by John Kemeny and Thomas Jurtz of Dartmouth College in early 1964 to be as approachable as possible.",
			Picture: Picture{
				Title:       "Can anyone beat the Altair System?",
				Alt:         "A May 1976 advertisement for the Altair 8800 computer.",
				Jpg:         "altair-ad.jpg",
				Attribution: "Michael Holley",
				License:     "public domain",
				LicenseLink: "https://commons.wikimedia.org/wiki/File:Altair_Computer_Ad_May_1976.jpg",
			},
		},
		{
			Year: 1975, Month: 3, Day: 5, Title: "The first meeting of the Homebrew Computer Club",
			Lead: "Homebrew Computer Club", LinkTitle: "about the Homebrew Computer Club",
			Link:    "https://www.computerhistory.org/revolution/personal-computers/17/312/1138",
			Content: "While many technology clubs of this type for sharing ideas were common, this Silicon Valley, Bay Area group became famous for its numerous members who later became industry figures.",
		},
		{
			Year: 1976, Month: 1, Title: "Software piracy", Highlight: true,
			Lead: "An Open Letter to Hobbyists", LinkTitle: "the letter",
			Link:    "https://archive.org/details/hcc0201/Homebrew.Computer.Club.Volume.02.Issue.01.Len.Shustek/page/n1/mode/2up",
			Content: "Bill Gates of <em>Micro-Soft</em> writes a letter to the hobbyists of the Homebrew Computer Club requesting they stop stealing Altair BASIC.",
			Picture: Picture{
				Title:       "An Open Letter to Hobbyists",
				Alt:         "A photo of the first page of the letter.",
				Jpg:         "an-open-letter-to-hobbyists.jpg",
				Attribution: "Len Shustek",
				License:     "public domain",
				LicenseLink: "https://commons.wikimedia.org/wiki/File:Bill_Gates_Letter_to_Hobbyists.jpg",
			},
		},
		{
			Year: 1976, Month: 3, Title: "The first Apple computer",
			Lead: "Apple-1", LinkTitle: "about the Apple-1",
			Link:    "https://www.computerhistory.org/revolution/personal-computers/17/312/1132",
			Content: "Steve Wozniak and Steve Jobs release the Apple I, a single-board computer with a 6502 CPU, 4KB of RAM, and a 40-column display controller.",
		},
		{
			Year: 1977, Month: 1, Title: "CP/M operating system",
			LinkTitle: "about CP/M", Link: "https://landley.net/history/mirror/cpm/history.html",
			Content: "Gary Kildall forms Digital Research to sell his hobbyist operating system, CP/M, for the Intel 8080. " +
				"Gary was an occasional consultant for Intel's microprocessor division, which gave him access to hardware and personnel. " +
				"CP/M became the first successful microcomputer operating system. " +
				"It dominated the remainder of the 1970s and is the default platform for most computers running an Intel 8080, 8085 or its compatible competitor, the Zilog Z-80.",
		},
		{
			Year: 1978, Month: 6, Title: "The first x86 CPU",
			Lead: "Intel 8086", LinkTitle: "about the Intel 8086",
			Link: "https://www.pcworld.com/article/535966/article-7512.html",
			Content: "Intel releases the 16-bit programmable microprocessor, the Intel 8086, which is the beginning of the <strong>x86 architecture</strong>.<br>" +
				"Unlike at the start of the decade when Intel broke new ground, this CPU design was a commercial response to market competition. " +
				"While code-compatible with the famous Intel 8080, this product failed to dominate in a market saturated with more affordable 8-bit hardware.",
		},
		{
			Title: "Intel 8088 CPU", Year: 1979, Month: 6,
			Lead: "Intel 8088", LinkTitle: "about the Intel 8088",
			Link: "https://spectrum.ieee.org/chip-hall-of-fame-intel-8088-microprocessor",
			Content: "Intel releases the lesser 16-bit microprocessor, the Intel 8088. " +
				"While fully compatible with the earlier Intel 8086 CPU, this model is intentionally \"castrated\" using an 8-bit external data bus. " +
				"The revision is an improvement for some buyers as it needs less expensive support chips on the mainboard and is compatible with the more readily available 8-bit hardware. " +
				"Software written for either CPU often gets quoted as 8088/86 compatible.",
		},
		{
			Title: "First commercial software for x86",
			Year:  1979, Month: 6, Day: 18,
			Lead: "Microsoft BASIC-86", LinkTitle: "Microsoft introduces BASIC-86",
			Link: "https://thisdayintechhistory.com/06/18/microsoft-introduces-basic-for-8086/",
		},
		{
			Title: "The first operating system for x86", Year: 1980, Month: 8,
			Lead: "Seattle Computer Products QDOS", LinkTitle: "about QDOS",
			Link: "https://www.1000bit.it/storia/perso/tim_paterson_e.asp",
			Content: "Tim Paterson worked on a project at Seattle Computer Products to create an 8086 CPU plugin board for the S-100 bus standard." +
				"Needing an operating system for the 16-bit Intel CPU, he programmed a half-complete, unauthorized clone of the CP/M operating system within four months." +
				"He called it QDOS (Quick and Dirty OS), and it sold few copies.",
		},
		{
			Title: "The first PC", Year: 1981, Month: 8, Day: 12, Highlight: true,
			Lead: "IBM Personal Computer", LinkTitle: "about the IBM PC",
			Link:    "https://www.ibm.com/ibm/history/exhibits/pc25/pc25_birth.html",
			Content: "Built on the 4.77 MHz Intel 8088 microprocessor, 16KB of RAM and Microsoft's PC-DOS, this underpowered machine heralds the <strong>PC platform</strong>.",
			Picture: Picture{
				Title:       "IBM PC 5150",
				Alt:         "A photo of the IBM PC 5150",
				Jpg:         "ibm-pc-5150.jpg",
				Attribution: "Rama & Musée Bolo",
				License:     "CC BY-SA 2.0 FR",
				LicenseLink: "https://creativecommons.org/licenses/by-sa/2.0/fr/deed.en",
			},
		},
		{
			Title: "The first published PC game", Year: 1981, Month: 9,
			Lead: "IBM's Microsoft Adventure", LinkTitle: "about Microsoft Adventure",
			Link:    "https://www.filfre.net/2011/07/microsoft-adventure/",
			Content: "A port of the text only Colossal Cave Adventure.",
		},
		{
			Title: "Initial release of MS-DOS", Year: 1982, Month: 8,
			Lead: "MS-DOS v1.25", LinkTitle: "about MS-DOS 1.25",
			Link:    "https://betawiki.net/wiki/MS-DOS_1.25",
			Content: "Microsoft releases the first edition of MS-DOS v1.25, which is readily available to all OEM computer manufacturers. All prior releases were exclusive to IBM.",
		},
		{
			Title: "Third-party PC games", Year: 1982,
			Content: "The first set of games gets released on the PC platform that IBM does not publish.<br>" +
				"<small>Some early publishers include <a href=\"//s3data.computerhistory.org/brochures/broderbund.software.1982.102646180.pdf\">Brøderbund</a>, " +
				"<a href=\"//archive.org/details/avalon-hill-game-company-catal-fall-1982\">The Avalon Hill Game Company</a>, " +
				"<a href=\"//archive.org/details/strategic-simulations-inc-summer-1982-catalog/mode/2up\">Strategic Simulations, Inc.</a>, " +
				"<a href=\"//www.uvlist.net/companies/info/1023-Windmill+Software\">Windmill Software</a>, " +
				"<a href=\"//retro365.blog/2019/09/23/bits-from-my-personal-collection-the-original-ibm-pc-and-orion-software/\">Orion Software</a> and " +
				"<a href=\"//www.uvlist.net/companies/info/1029-Spinnaker+Software\">Spinnaker Software</a></small>",
		},
		{
			Title: "The first PC clone", Year: 1983, Month: 3,
			Lead: "COMPAQ Portable", LinkTitle: "the advertisement",
			Link: "https://www.computerhistory.org/revolution/personal-computers/17/302/1194",
			Content: "Compaq Computer Corporation releases the first IBM PC compatible computer, the Compaq Portable. " +
				"It is the first PC clone to use the same software and expansion cards as the IBM PC.",
		},
		{
			Title: "PC / MS-DOS 2 released", Year: 1983, Month: 3,
			Lead: "The OS includes ANSI.SYS", LinkTitle: "about MS-DOS ANSI.SYS",
			Link:    "https://github.com/microsoft/MS-DOS/blob/master/v2.0/source/ANSI.txt",
			Content: "Includes for the first time a device driver to view ANSI text graphics in color.",
		},
		{
			Title: "The earliest cracked PC game", Year: 1983,
			Lead: "Atarisoft's Galaxian broken by Koyote Kid", LinkTitle: "about and view the crack",
			Link: "/f/ab2edbc", Highlight: true,
			Picture: Picture{
				Title: "Galaxian broken by Koyote Kid",
				Alt:   "Galaxian broken screenshot",
				Webp:  "ab2edbc.webp",
				Png:   "ab2edbc.png",
			},
		},
		{
			Title: "Major videogame publishers enter the PC market", Year: 1983,
			Content: "Some major arcade and videogame publishers of the era release on the PC.<br>" +
				"<small><a href=\"//dfarq.homeip.net/atarisoft-if-you-cant-beat-em-join-em/\">Atarisoft</a>, " +
				"<a href=\"//www.uvlist.net/companies/info/243-Infocom\">Infocom</a>, " +
				"<a href=\"//www.resetera.com/threads/lets-look-back-at-game-company-datasoft.587093/##post-87110411\">Datasoft</a>, " +
				"<a href=\"//www.uvlist.net/companies/info/83-Mattel%20Electronics\">Mattel</a> and " +
				"<a href=\"//www.wired.com/story/sierra-online-ken-williams-interview-memoir/\">Sierra On-Line</a></small>",
		},
		{
			Title: "The first PC virus", Year: 1986, Month: 1, Day: 19,
			Lead: "Brain", LinkTitle: "about the Brain virus",
			Link:    "https://www.f-secure.com/v-descs/brain.shtml",
			Content: "The first PC virus, Brain, infects the boot sector of floppy disks.",
			Picture: Picture{
				Title:       "A hex dump of the Brain",
				Alt:         "A hex dump of the boot sector of a floppy disk containing the PC virus, Brain.",
				Jpg:         "brain-virus.jpg",
				Attribution: "Avinash Meetoo",
				License:     "CC-BY-2.5",
				LicenseLink: "https://creativecommons.org/licenses/by/2.5/deed.en",
			},
		},
		{
			Title: "Earliest unprotect text", Year: 1983, Month: 5, Day: 12, Highlight: true,
			Lead: "Directions by Randy Day for unprotecting SPOC the Chess Master", LinkTitle: "the unprotect text",
			Link: "/f/a91c702",
			Content: "<code>SPOC.UNP</code><br>" +
				"Unprotects were text documents describing methods to remove software copy protection on floppy disks." +
				"Many authors were legitimate owners who were frustrated that publishers would not permit them to create backup copies of their expensive but fragile 5¼-inch floppy disks for daily driving.",
		},
		{
			Title: "Microsoft Windows announced", Year: 1983, Month: 11, Day: 10,
			Link:      "https://www.poynter.org/reporting-editing/2014/today-in-media-history-in-1983-bill-gates-and-microsoft-introduced-windows/",
			LinkTitle: "about the announcement",
			Content:   "In hindsight, this premature announcement aims to keep Microsoft customers from jumping to competitor graphical user interface software.",
		},
		{
			Title: "Major game publishers enter the PC market", Year: 1984,
			Content: "<a href=\"//www.polygon.com/a/how-ea-lost-its-soul/\">Electronic Arts</a>, " +
				"<a href=\"//www.ign.com/articles/2010/10/01/the-history-of-activision\">Activision</a>, " +
				"<a href=\"//segaretro.org/IBM_PC\">Sega</a> and " +
				"<a href=\"//corporate-ient.com/microprose/\">MicroProse Software</a>* publish on the platform." +
				"<br>* The company founded by Sid Meier",
		},
		{
			Title: "The first 16 color PC game", Year: 1984, Month: 8,
			Lead: "King's Quest", LinkTitle: "the game manual",
			Link: "http://www.sierrahelp.com/Documents/Manuals/Kings_Quest_1_IBM_-_Manual.pdf",
			Content: "The first PC game to use 16 colors, King's Quest, is created by  Sierra On-Line and released by IBM. " +
				"IBM PC graphics cards are limited to 4 colors, but the game is released for the new IBM PCjr that displays upto 16 colors.",
		},
		{
			Title: "The earliest information text", Year: 1984, Month: 10, Day: 17, Highlight: true,
			Lead:      "Zorktools 1 by Software Pirates Inc",
			LinkTitle: "the information text",
			Link:      "/f/ae2da98",
			Content: "<code>INFOCOM.DOC</code><br>" +
				"Information texts were documents included in a release describing how to how to use a utility program."},
		{
			Title: "EGA graphics standard", Year: 1984, Month: 10,
			Lead: "16 colors from a 64 color pallete", LinkTitle: "How 16 colors saved PC gaming",
			Link:    "https://www.custompc.com/retro-tech/ega-graphics",
			Content: "The Enhanced Graphics Adapter standard includes 16 colors, 640×350 pixel resolution and 80×25 text mode.",
		},
		{
			Title: "An early demonstration on the PC", Year: 1984, Month: 10,
			Lead: "Fantasy Land EGA demo by IBM", LinkTitle: "and run the demo",
			Link: "https://www.pcjs.org/software/pcx86/demo/ibm/ega/",
			Content: "The first demo program on the PC, Fantasy Land, is created by IBM to demonstrate the new EGA graphics standard. " +
				"The idea of a demo is to have the program run automatically, without user input, to show off the capabilities of the hardware.",
		},
		{
			Prefix: "The earliest PC groups,", Year: 1984,
			List: Links{
				{LinkTitle: "Against Software Protection <small>ASP</small>", Link: "/g//against-software-protection"},
				{LinkTitle: "Software Pirates Inc <small>SPi</small>", Link: "/g/software-pirates-inc"},
			},
		},
		{
			Title: "The earliest text loader", Year: 1985, Month: 5, Day: 26, Highlight: true,
			Lead:      "Bally Midway's Spy Hunter by Imperial Warlords",
			LinkTitle: "and view the text loader",
			Link:      "/f/aa2be75",
			Content: "Text loaders and ANSI art offer similar results but are different in execution. " +
				"Text loaders are binary programs that display text mode characters and colors. " +
				"ANSI text required the ANSI.SYS device driver included in PC/MS-DOS 2+ to convert plain text files into onscreen animation and color.",
			Picture: Picture{
				Title: "Spy Hunter",
				Alt:   "Spy Hunter by Imperial Warlords screenshot",
				Webp:  "aa2be75.webp",
				Png:   "aa2be75.png",
			},
		},
		{
			Title: "Initial release of Microsoft Windows", Year: 1985, Month: 11, Day: 20,
			Lead: "Windows 1.0", LinkTitle: "about the failure of Windows 1.0",
			Link:    "https://www.theverge.com/2012/11/20/3671922/windows-1-0-microsoft-history-desktop-gracefully-failed",
			Content: "Microsoft Windows 1.0 is released. Expensive hardware requirements and a lack of purpose lead to lackluster sales. It will take a decade and multiple releases before Windows becomes dominant.",
			Picture: Picture{
				Title: "Microsoft Windows 1.01",
				Alt:   "Microsoft Windows 1.01 booting up screenshot",
				Webp:  "windows-version-1.webp",
				Png:   "windows-version-1.png",
			},
		},
		{
			Title: "The earliest \"DOX\"", Year: 1986, Highlight: true,
			Lead: "Dam Buster Documentation by Apocalypse Now BBS", LinkTitle: "the documentation",
			Link: "/f/a61db76",
			Content: "<code>DAMBUST1.DOC</code><br>" +
				"DOX is an abbreviation for documentation, which are text files that provide instructions on playing more complicated games. " +
				"These titles often relied on printed instruction manuals included in the purchased game box to be playable." +
				"<br>Dam Buster is a misname of the game The Dam Busters, a 1984 game published by Accolade.",
		},
		{
			Title: "PC clone sales pickup in Europe", Year: 1986,
			Link:      "https://www.computerhistory.org/revolution/personal-computers/17/302",
			LinkTitle: "about the PC clone market",
			Content: "The European market is dominated by the Amiga and Atari ST, but PC clones gain popularity." +
				" Popular machines include the <a href=\"https://www.dosdays.co.uk/computers/Amstrad%20PC1000/amstrad_pc1000.php\">Amstrad PC1512</a>, " +
				"the Philips P2000T and the <a href=\"https://www.dosdays.co.uk/computers/Olivetti%20M24/olivetti_m24.php\">Olivetti M24</a>.",
		},
		{
			Title: "The first 16 color EGA game", Year: 1986, Month: 3,
			Lead: "Accolade's Mean 18", LinkTitle: "the moby games entry",
			Link: "https://www.mobygames.com/game/152/mean-18/",
		},
		{
			Title: "The earliest PC loaders", Year: 1986, Month: 6, Highlight: true,
			Content: "Loaders were named as they would be the first thing to display each time a cracked game is run. " +
				"These screens were static images in the early days and sometimes contained ripped screens from other games. Some users found these annoying and a cause of file bloat.",
			List: Links{
				{LinkTitle: "Atarisoft's Gremlins by Mr. Turbo", Link: "/f/b44cac"},
				{LinkTitle: "Exodus: Ultima 3 by ESP Pirates", Link: "/f/a83eec"},
				{LinkTitle: "Sega's Frogger II by SPI", Link: "/f/b33404"},
			},
			Picture: Picture{
				Title: "Software Pirates, Inc presents",
				Alt:   "Software Pirates, Inc presents Frogger II  screenshot",
				Webp:  "b33404.webp",
				Png:   "b33404.png",
			},
		},
		{
			Year: 1986, Prefix: notable,
			List: Links{
				{LinkTitle: "Five-O", Link: "/g/five-o"},
				{LinkTitle: "ESP Pirates", Link: "/g/esp-pirates"},
			},
			Picture: Picture{
				Title: "Five O Presents",
				Alt:   "Five O Presents screenshot",
				Webp:  "five-o.webp",
				Png:   "five-o.png",
			},
		},
		{
			Title: "VGA graphics standard", Year: 1987, Month: 4, Day: 2,
			Lead: "256 color graphics", LinkTitle: "about the VGA graphics standard",
			Link:    "https://www.computer.org/publications/tech-news/chasing-pixels/Famous-Graphics-Chips-IBMs-VGA",
			Content: "The Video Graphics Array (VGA) graphics standard is released. It is the first graphics standard to support 256 colors and resolutions up to 640x480.",
		},
		{
			Title: "The earliest PC demo", Year: 1987, Month: 6, Day: 22, Highlight: true,
			Lead:      "3 Dimensional EGA Demonstration",
			LinkTitle: "and view the demo", Link: "/f/ac21460",
			Content: "A demo and a piece of software created purely for aesthetics, usually to show art or animation. " +
				"While earlier demonstration software existed on the PC, they were intended for retailers or distributors and usually not given to the public.",
		},
		{
			Title: "AdLib audio", Year: 1987,
			Lead: "AdLib Music Synthesizer Card", LinkTitle: "about the AdLib sound card",
			Link: "https://www.computinghistory.org.uk/det/23724/AdLib-Music-Synthesizer-Card/",
			Content: "The Music Synthesizer Card sound card is released. It is the first sound card to use FM synthesis and is the first to be widely adopted by game developers. " +
				"Adlib's success is short lived as Creative Labs releases the Sound Blaster in 1989.",
		},
		{
			Year: 1987, Prefix: notable,
			List: Links{
				{LinkTitle: "KGB", Link: "/g/ptl-club"},
				{LinkTitle: `Boys from Company C <small>(BBC)</small>`, Link: "/g/boys-from-company-c", Forward: "Five-O"},
				{LinkTitle: "The PTL Club", Link: "/g/ptl-club"},
				{LinkTitle: "Canadian Pirates Inc <small>(CPI)</small>", Link: "g/canadian-pirates-inc"},
			},
			Picture: Picture{
				Title: "The PTL Club",
				Alt:   "The PTL Club Presents screenshot",
				Webp:  "the-ptl-club.webp",
				Png:   "the-ptl-club.png",
			},
		},
		{
			Title: "The first 32 color VGA game", Year: 1988, Month: 3,
			Lead: "Arcadia's Rockford: The Arcade Game", LinkTitle: "the discussion",
			Link: "https://forum.winworldpc.com/discussion/comment/174818/#Comment_174818",
		},
		{
			Title: "Earliest standalone BBS ad", Year: 1988, Month: 4, Day: 4, Highlight: false,
			Lead: "Swashbucklers II", LinkTitle: "the file",
			Link: "/f/b844ef",
			Picture: Picture{
				Title: "Swashbucklers II",
				Alt:   "Swashbucklers II text advert screenshot",
				Webp:  "b844ef.webp",
				Png:   "b844ef.png",
			},
		},
		{
			Title: "Earliest ANSI ad", Year: 1988, Month: 6, Highlight: false,
			Lead: "Paperboy by BSP", LinkTitle: "the file",
			Link: "/f/a8286b",
			Content: "The earliest ANSI ad is released by BSP for the game Paperboy. " +
				"ANSI art is a computer art form that became widely used to create art and advertisements for BBSes. " +
				"ANSI art is created using ANSI escape codes to create colored text and is usually viewed in a terminal emulator.",
			Picture: Picture{
				Title: "Paperboy by BSP",
				Alt:   "Paperboy by BSP ANSI screenshot",
				Webp:  "a8286b.webp",
				Png:   "a8286b.png",
			},
		},
		{
			Title: "Earliest NFO as a text document", Year: 1988, Month: 7, Day: 30, Highlight: false,
			Lead: "Bentley Sidwell Productions", LinkTitle: "the file", Link: "/f/9f3f4e",
			Content: "The earliest NFO-like document is released by Bentley Sidwell Productions for the game " +
				"Romance of The Three Kingdoms. NFO files are text documents that contain information about a release, such as the release group, " +
				"release date, and release notes. NFO files are usually distributed with pirated software and are often " +
				"used to promote the release group.",
			Picture: Picture{
				Title: "Bentley Sidwell Productions document",
				Alt:   "Romance of The Three Kingdoms by Bentley Sidwell Productions document screenshot",
				Webp:  "9f3f4e.webp",
				Png:   "9f3f4e.png",
			},
		},
		{
			Title: "The earliest ASCII art", Year: 1988, Month: 10, Day: 6, Highlight: true,
			Lead: "Another quality ware from $print", LinkTitle: "the file", Link: "/f/ab3dc1",
			// TODO: fix content
			Content: "The earliest ASCII art is released by $print for the game " +
				"Battle Chess. ASCII art is a computer art form that became widely used to create art and advertisements for BBSes. " +
				"ASCII art is created using ASCII characters and is usually viewed in a terminal emulator.",
			Picture: Picture{
				Title: "Another quality ware from $print",
				Alt:   "Battle Chess by $print ASCII screenshot",
				Webp:  "ab3dc1.webp",
				Png:   "ab3dc1.png",
			},
		},
	}
	return m
}

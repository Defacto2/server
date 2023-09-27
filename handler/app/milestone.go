package app

// Package file milestone.go contains the milestones for The Scene,
// that are used by the home page.

const notable = "Notable group foundings,"

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
	SubTitle  string // SublTitle is the title of the Link in a smaller font and in brackets.
	Link      string // Link is the URL to an article about the milestone or the product.
	Forward   string // Forward is an optional name of a group that is prefixed before the link to indicate a merger.
}

// Milestones is a collection of Milestone.
type Milestones []Milestone

// Len is the number of Milestones.
func (m Milestones) Len() int {
	return len(m)
}

// Collection of Milestones from the 1970s onwards.
func Collection() Milestones {
	m := []Milestone{
		{
			Year: 1971, Month: 10, Title: "Secrets of the Little Blue Box", Highlight: true,
			Lead: "Esquire October 1971", LinkTitle: "the complete article",
			Link: "https://www.slate.com/articles/technology/the_spectator/2011/10/the_article_that_inspired_steve_jobs_secrets_of_the_little_blue_.html",
			Content: "Ron Rosenbaum writes the first mainstream article on phone freaks, primarily kids who'd hack and experiment with the global telephone network.<br>" +
				"The piece coins them as phone-<strong>phreaks</strong> and introduces the reader to the kids' use of <strong>pseudonyms</strong> or codenames within their cliques and <strong>groups</strong> of friends. " +
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
			Year: 1977, Title: "The trinity of microcomputers",
			Lead: "Apple II, Commodore PET, TRS-80", LinkTitle: "about the Apple II, Commodore PET and TRS-80",
			Link: "https://cybernews.com/editorial/the-1977-trinity-and-other-era-defining-pcs/",
			Content: "The Apple II, Commodore PET and TRS-80 are released, the first microcomputers to be readily available to the public. " +
				"By the end of the year, a potential customer in the USA could walk into a mall or retail shop and walk out with a complete personal computer, ready to use.",
		},
		{
			Year: 1978, Month: 2, Title: "The first BBS",
			Lead: "CBBS", LinkTitle: "the Byte Magazine article", Link: "https://vintagecomputer.net/cisc367/byte%20nov%201978%20computerized%20BBS%20-%20ward%20christensen.pdf",
			Content: "Ward Christensen and Randy Suess create the first Bulletin Board System (BBS), the Computerized Bulletin Board System (CBBS) in Chicago. " +
				"The software was custom written in 8080 assembler language which ran on a S-100 bus computer together with the brand new $300, Hayes 110 / 300 baud modem. " +
				"The board became extremely popular, with callers from around the world after articles and logs were published in both Byte and Dr. Dobb's Journal magazines later in the year.",
			Picture: Picture{
				Title:       "A recreation of CBBS",
				Alt:         "A recreation screen capture of the first BBS.",
				Png:         "cbbs.jpg",
				Webp:        "cbbs.webp",
				Attribution: "Aeroid",
				License:     "CC BY-SA 4.0",
				LicenseLink: "https://creativecommons.org/licenses/by-sa/4.0/deed.en",
			},
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
			Title: "The first popular x86 CPU", Year: 1979, Month: 6,
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
			Content: "Microsoft BASIC and its many revisions were the first killer applications for Microsoft in its early years. " +
				"Most microcomputers were sold to enthusiasts or businesses, and the software availability could have been better. " +
				"So many owners resorted to creating software, and the BASIC programming language had the easiest learning curve. " +
				"Microsoft didn't invent the language, but its implementation was considered the gold standard.",
		},
		{
			Title: "The first operating system for x86", Year: 1980, Month: 8,
			Lead: "Seattle Computer Products QDOS", LinkTitle: "about QDOS",
			Link: "https://www.1000bit.it/storia/perso/tim_paterson_e.asp",
			Content: "Tim Paterson worked on a project at Seattle Computer Products to create an 8086 CPU plugin board for the S-100 bus standard. " +
				"Needing an operating system for the 16-bit Intel CPU, he programmed a half-complete, unauthorized clone of the CP/M operating system within four months." +
				"He called it QDOS (Quick and Dirty OS), and it sold few copies.",
		},
		{
			Title: "Computer Software Copyright Act of 1980", Year: 1980, Month: 12, Day: 12, Highlight: true,
			Lead: "Software is defined by copyright laws in the USA", LinkTitle: "about the act",
			Link:    "https://www.c2st.org/the-computer-software-copyright-act-of-1980/",
			Content: "Signed as an amendment to law by President Jimmy Carter, computer programs are defined by copyright law and enable authors to control the copying, selling, and leasing of their software.",
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
			Link: "https://www.filfre.net/2011/07/microsoft-adventure/",
			Content: "A PC port of the text only Colossal Cave Adventure. " +
				"Adventure was a highly influential and popular text-only adventure game for mainframe computers of the 1970s. " +
				"Will Crowther wrote it in FORTRAN for the PDP-10 system and Don Woods at the Stanford AI Lab in California later expanded it. " +
				"The game created the interactive fiction genre, which later led to graphic adventures and story narratives in video games.",
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
			Content: "Around this time, <abbr title=\"graphical user interface\" class=\"initialism\">GUI</abbr> for microcomputing was the hype in the technology industry and media. " +
				"In hindsight, this premature announcement from Microsoft aimed to keep customers from jumping to competitor GUI platforms and offerings. " +
				"It took over a decade before graphical interfaces on the PC replaced text in business computing and even longer before it became commonplace in the home." +
				"<br>Other microcomputer platforms, such as the <span class=\"text-nowrap\">Apple Macintosh <sup>1984</sup></span>, <span class=\"text-nowrap\">Commodore Amiga</span> and <span class=\"text-nowrap\">Atari ST <sup>1985</sup></span> came with a GUI as standard.",
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
				"Information texts were documents included in a release describing how to how to use a utility program.",
		},
		{
			Title: "EGA graphics standard", Year: 1984, Month: 10,
			Lead: "16 colors from a 64 color palette", LinkTitle: "How 16 colors saved PC gaming",
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
			Title: "Razor 1911 is named", Year: 1985, Month: 11,
			Lead: "On the Commodore 64", LinkTitle: "about the early days of Razor 1911",
			Link: "https://csdb.dk/group/?id=431",
			Content: "Razor 1911, the oldest and most famed brand in The Scene, is founded in <strong>Norway</strong> with three members. " +
				"The group released demos and later cracked exclusively for the Commodore 64 and then the Amiga. Co-founder Sector 9 took the brand to the <a href=\"/f/a12d5e\">PC in late 1990</a>.<br>" +
				"The distinctive number suffix was a fad with groups of the Commodore 64 era. <q>1911</q> denotes the decimal value of hexadecimal <code>$777</code>.",
			Picture: Picture{
				Title:       "Amazing Demo I",
				Alt:         "Amazing Demo I by Razor 1911 screenshot",
				Jpg:         "razor-1911-is-founded.png",
				Attribution: "CSDb",
				License:     "© Dr.Jekyll, Sector 9 of Razor 1911",
				LicenseLink: "https://csdb.dk/release/?id=230004",
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
			Content: "While the Commodore, Apple and IBM are common platforms in the US, the European market doesn't always share the same popular platforms. " +
				"Import duties, slow international distribution channels and a lack of localized software and hardware often hampers the adoption of some platforms. " +
				"The Western European market is dominated by Acorn, Amstrad, Commodore, Sinclair but the PC clones produced by local electronic manufactures gain popularity. " +
				"Popular machines include the <a href=\"https://www.dosdays.co.uk/computers/Amstrad%20PC1000/amstrad_pc1000.php\">Amstrad PC1512</a>, " +
				"the Philips P2000T and the <a href=\"https://www.dosdays.co.uk/computers/Olivetti%20M24/olivetti_m24.php\">Olivetti M24</a>.",
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
				{LinkTitle: "ESP Pirates", Link: "/g/esp-pirates"},
				{LinkTitle: "Five-O", Link: "/g/five-o"},
			},
			Picture: Picture{
				Title: "Five O Presents",
				Alt:   "Five O Presents screenshot",
				Webp:  "five-o.webp",
				Png:   "five-o.png",
			},
		},
		{
			Title: "Fairlight is founded", Year: 1987, Month: 3,
			Lead: "On the Commodore 64 and Amiga", LinkTitle: "about the early days of Fairlight",
			Link: "http://janeway.exotica.org.uk/target.php?idp=6375&idr=1940&tgt=1",
			Content: "Fairlight, one of the oldest brands in The Scene, is founded in <strong>Sweden</strong> with just three members. " +
				"The group cracked and released demos exclusively for the Commodore C64 and Amiga platforms before expanding to consoles and the <a href=\"/f/b04615\">PC in February 1991</a>.",
			Picture: Picture{
				Title:       "Fairlight Intro (the Legendary one)",
				Alt:         "Commodore 64, Fairlight Intro (the Legendary one) screenshot",
				Jpg:         "fairlight-is-founded.png",
				Attribution: "CSDb",
				License:     "© Woodo of Fairlight",
				LicenseLink: "https://csdb.dk/release/index.php?id=53390",
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
			Title: "Music audio standard", Year: 1987,
			Lead: "AdLib Music Synthesizer Card", LinkTitle: "about the AdLib sound card",
			Link: "https://www.computinghistory.org.uk/det/23724/AdLib-Music-Synthesizer-Card/",
			Content: "The Music Synthesizer Card sound card is released. It is the first sound card to use FM synthesis and is the first to be widely adopted by game developers. " +
				"Adlib's success is short lived as Creative Labs releases the Sound Blaster in 1989.",
		},
		{
			Year: 1987, Prefix: notable,
			List: Links{
				{
					LinkTitle: `Boys from Company C <small>(BBC)</small>`, Link: "/g/boys-from-company-c",
					Forward: "Five-O",
				},
				{LinkTitle: "Canadian Pirates Inc <small>(CPI)</small>", Link: "g/canadian-pirates-inc"},
				{LinkTitle: "KGB", Link: "/g/ptl-club"},
				{LinkTitle: "The PTL Club", Link: "/g/ptl-club"},
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
			Title: "Earliest standalone 'elite' BBS ad", Year: 1988, Month: 4, Day: 4, Highlight: false,
			Lead: "Swashbucklers II", LinkTitle: "the file",
			Link: "/f/b844ef",
			Content: "Home of PTL/CPI<br>" +
				"100 megs Online!<br>" +
				"85 megs Offline, Request!<br>" +
				"All PTL/CPI Cracks FREE<br>" +
				"All other Major Groups cracks Always Online<br>" +
				"Ask your local Sysop for the number.",
			Picture: Picture{
				Title: "Swashbucklers II",
				Alt:   "Swashbucklers II text advert screenshot",
				Webp:  "b844ef.webp",
				Png:   "b844ef.png",
			},
		},
		{
			Title: "Earliest ANSI ad", Year: 1988, Month: 6, Highlight: false,
			Lead: "Bentley Sidwell Productions", LinkTitle: "and view the file",
			Link: "/f/a8286b",
			Content: "The earliest ANSI ad is released by Bentley Sidwell Productions for the game Paperboy. " +
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
			Lead: "Another quality ware from $print", LinkTitle: "and view the file", Link: "/f/ab3dc1",
			Content: "The earliest ASCII art known so-far is released by $print for the game " +
				"Fire Power. The ASCII logo is relatively crude and is not as detailed as later ASCII art. " +
				"<pre> ╔═══════════════════════════════╗<br>" +
				"╔╝      Another Quality Ware     ╚╗<br>" +
				"║          F  R  O  M             ║<br>" +
				"║                                 ║<br>" +
				"║   ┌┼┼┼ ┌─┐┌──┐ ─┬─ │\\  │──┬──   ║<br>" +
				"║   └┼┼┼┐┼─┘│─┬┘  │  │ \\ │  │     ║<br>" +
				"║   ─┼┼┼┘│  │ └─ ─┴─ │  \\│  │     ║<br>" +
				"╚═════════════════════════════════╝</pre>",
			Picture: Picture{
				Title: "Another quality ware from $print",
				Alt:   "Fire Power by $print ASCII screenshot",
				Webp:  "ab3dc1.webp",
				Png:   "ab3dc1.png",
			},
		},
		{
			Title: "The earliest scene drama", Year: 1988, Month: 11, Day: 25,
			Lead: "TNWC accusing PTL of stealing a release", LinkTitle: "and view the file",
			Link: "/f/aa356d",
			Content: "The earliest scene drama known so-far involves a release by The North West Connection (TNWC) for the game " +
				"Paladin. The drama in the text file accuses PTL of stealing and \"re-releasing\" a release from TNWC. " +
				"Scene drama is often text that is used to call out other groups for bad behavior, stealing releases, " +
				"or to call out other groups for being lame.",
			Picture: Picture{
				Title: "TNWC accusing PTL of stealing a release",
				Alt:   "TNWC accusing PTL of stealing a release screenshot",
				Webp:  "aa356d.webp",
				Png:   "aa356d.png",
			},
		},
		{
			Year: 1988, Prefix: notable,
			List: Links{
				{LinkTitle: "Bentley Sidwell Productions", Link: "/g/bentley-sidwell-productions", SubTitle: "BSP"},
				{LinkTitle: "Boys from Company C", Link: "/g/boys-from-company-c", SubTitle: "BCC", Forward: "Five-O"},
				{LinkTitle: "Crackers in Action", Link: "/g/crackers-in-action", SubTitle: "CIA"},
				{LinkTitle: "Miami Cracking Machine", Link: "/g/miami-cracking-machine", SubTitle: "MCM"},
				{LinkTitle: "Sprint", Link: "/g/sprint"},
				{LinkTitle: "The Grand Council", Link: "/g/the-grand-council", SubTitle: "TGC", Forward: "Dude Man Dude HQ"},
				{
					LinkTitle: "The North West Connection", Link: "/g/the-north-west-connection",
					SubTitle: "TNWC", Forward: "The Neutral Zone",
				},
				{LinkTitle: "The Sysops Association Network", Link: "/g/the-sysops-association-network", SubTitle: "TSAN"},
			},
		},
		{
			Title: "The first 256 color VGA game", Year: 1989, Month: 3,
			Lead: "688 Attack Sub from Electronic Arts", LinkTitle: "the mobygames page",
			Link: "https://www.mobygames.com/game/2099/688-attack-sub",
		},
		{
			Title: "Earliest BBS ANSI loader", Year: 1989, Month: 3,
			Lead: "Rogues Gallery BBS", LinkTitle: "the file",
			Link:    "/f/ad21da8",
			Content: "The Rogues Gallery BBS was based in Long Island, New York, area code 516.",
			Picture: Picture{
				Title: "Rogues Gallery BBS",
				Alt:   "Rogues Gallery BBS ANSI ad screenshot",
				Webp:  "ad21da8.webp",
				Png:   "ad21da8.png",
			},
		},
		{
			Title: "Earliest PC intro", Year: 1989, Month: 4, Highlight: true,
			Lead: "First intro by Sorcerers", LinkTitle: "the file",
			Link: "/f/ab2843",
			Content: "An intro or the later cractrkro are small, usually short, demonstration programs designed to display text with are or animations. " +
				"Oddly, this first intro was created by a group of teenagers out of <strong>Findland</strong>, a country not known for its use of the expensive PC platform. " +
				"Other 16-bit platforms such as the Commodore Amiga and Atari ST offered much better graphics than the CGA on the PC and were more popular in Europe.",

			Picture: Picture{
				Title: "First intro by Sorcerers",
				Alt:   "First intro by Sorcerers screenshot",
				Webp:  "ab2843.webp",
				Png:   "ab2843.png",
			},
		},
		{
			Title: "Earliest PC cracktro", Year: 1989, Month: 4, Day: 29, Highlight: true,
			Lead: "Future Brain Inc", LinkTitle: "and run the cracktro",
			Link: "/f/b83fd7",
			Content: "This first cracktro is released by Future Brain Inc for the game Lombard RAC Rally. " +
				"Future Brain Inc were a group from the <strong>Netherlands</strong>, and were one of the first groups to release cracktros on the PC platform.<br>" +
				"Early cracktros on the PC platform lacked music, and were usually just a simple text screen with a logo. " +
				"On other platforms such as the Commodore 64, Amiga, and Atari ST, cracktros offered music and graphic effects which were easier to program due to their standard hardware. ",
			Picture: Picture{
				Title: "Lombard RAC Rally cracktro",
				Alt:   "Lombard RAC Rally cracktro screenshot",
				Webp:  "b83fd7.webp",
				Png:   "b83fd7.png",
			},
		},
		{
			Title: "First issue of Pirate magazine", Year: 1989, Month: 6, Day: 1,
			Lead: "The earlist known scene newsletter for The Scene on the PC", LinkTitle: "the issues",
			Link:    "/g/pirate",
			Content: "Based in Chicago, Pirate magazine was a bi-monthly newsletter for The Scene on the PC platform.",
		},
		{
			Year: 1989, Prefix: notable,
			List: Links{
				{LinkTitle: "Aces of ANSI Art", Link: "/g/aces-of-ansi-art", SubTitle: "AAA"},
				{LinkTitle: "American Pirate Industries", Link: "/g/american-pirate-industries", SubTitle: "API"},
				{LinkTitle: "Future Brain Inc.", Link: "/g/future-brain-inc", SubTitle: "FBi"},
				{
					LinkTitle: "International Network of Crackers", Link: "/g/international-network-of-crackers",
					SubTitle: "INC", Forward: "MCM, NYC, NCC",
				},
				{LinkTitle: "New York Crackers", Link: "/g/new-york-crackers", SubTitle: "NYC"},
				{LinkTitle: "Norwegian Cracking Company", Link: "/g/norwegian-cracking-company", SubTitle: "NCC"},
				{LinkTitle: "Pirates Sick of Initials", Link: "/g/pirates-sick-of-initials", SubTitle: "PSi"},
				{LinkTitle: "The Firm", Link: "/g/the-firm", Forward: "BCC, Bentley Sidwell Productions"},
				{LinkTitle: "The Underground Council", Link: "/g/the-underground-council", SubTitle: "UGC"},
				{LinkTitle: "Triad", Link: "/g/triad", Forward: "PTL, PSi, Sprint, UGC"},
			},
			Picture: Picture{
				Title: "Another superior FiRM crack by",
				Alt:   "Another superior FiRM crack EGA screenshot",
				Webp:  "the-firm.webp",
				Png:   "the-firm.png",
			},
		},
		{
			Title: "Use of the \".NFO\" file extension", Year: 1990, Month: 1, Day: 23, Highlight: true,
			Lead: "The Humble Guys, Knights of Legend", LinkTitle: "the file",
			Link: "/f/ab3945",
			Content: "<code>KNIGHTS.NFO</code><br>" +
				"The timestamps for the Knights of Legend release predate the Bubble Bobble release by a few days.<br>" +
				"<figure><blockquote class=\"blockquote\"><small>It happened like this, I'd just used \"Unguard\" to crack the SuperLock off of <a href=\"/f/ad4195\">Bubble Bobble</a>, and I said \"I need some file to put the info about the crack in. Hmmm.. Info, NFO!\", and that was it.</small></blockquote>" +
				"<figcaption class=\"blockquote-footer\">Famed, former cracker for The Humble Guys, Fabulous Furlough maintains Bubble Bobble was the first THG release that used the .NFO file extension.</figcaption></figure>" +
				"The <code>.NFO</code> file extension is used to denote a text file containing information about a release. " +
				"Still in use today, the .NFO file contains information about the release group, the release itself, and how to install the release.",
		},
		{
			Title: "Earliest PC cracktro with music", Year: 1990, Month: 12, Day: 2,
			Lead: "The Cat, M1 Tank Plattoon", LinkTitle: "about and view cractrko",
			Link: "/f/ab25f0e",
			Content: "This cracktro was released by The Cat for the game M1 Tank Platoon. " +
				"It is the first known cracktro on the PC platform to feature music. " +
				"But \"music\" in a loose sense, as it relies on the terrible internal PC speaker to produce the tune.<br>" +
				"While the 8-bit consoles and some microcomputers offered dedicated music audio chips, most famously the Commodore 64 with its SID chip, " +
				"the IBM PC which targeted business did not.",
			Picture: Picture{
				Title: "Tank Platoon cracktro",
				Alt:   "Tank Platoon cracktro screenshot",
				Webp:  "ab25f0e.webp",
				Png:   "ab25f0e.png",
			},
		},
		{
			Title: "Digital audio standard", Year: 1990,
			Lead: "SoundBlaster",
			Content: "The SoundBlaster audio standard was released by Creative Labs in 1990. " +
				"It was the first digital audio standard for the IBM PC to be widely adopted on the PC platform, despite its poor quality, mono 8-bit digital audio. " +
				"Previous audio standards such as the AdLib and the MT-32 were limited to FM synthesis or MIDI-like samples.<br>" +
				"The SoundBlaster was the first audio standard to be widely adopted by the PC platform, and was the de facto standard for many years.",
		},
		{
			Title: "CD-ROM media", Year: 1990, Prefix: "Winter",
			Lead: "Mixed-Up Mother Goose", LinkTitle: "the Catalog listing the game",
			Link: "https://archive.org/details/vgmuseum_sierra_sierra-90catalog-alt3/page/n21",
			Content: "The first, widely available enhanced PC game on CD-ROM was Mixed-Up Mother Goose, released by Sierra On-Line in 1990. " +
				"The game was originally released in 1987, but the CD-ROM version featured enhanced VGA graphics and digital audio.",
		},
		{
			Year: 1990, Prefix: notable,
			List: Links{
				{LinkTitle: "ANSI Creators in Demand", Link: "/g/acid-productions", SubTitle: "ACiD", Forward: "Aces of ANSI Art"},
				{LinkTitle: "Bitchin ANSI Design", Link: "/g/bitchin-ansi-design", SubTitle: "BAD"},
				{LinkTitle: "Damn Excellent ANSI Design", Link: "/g//damn-excellent-ansi-design", SubTitle: "Damn"},
				{LinkTitle: "Future Crew", Link: "/g/future-crew", SubTitle: "FC"},
				{LinkTitle: "National Elite Underground Alliance", Link: "/g/national-elite-underground-alliance", SubTitle: "NEUA"},
				{LinkTitle: "Public Enemy", Link: "/g/public-enemy", SubTitle: "PE", Forward: "Red Sector Inc."},
				{LinkTitle: "Razor 1911", Link: "/g/razor-1911", SubTitle: "RZR", Forward: "Razor / Skillion"},
				{LinkTitle: "Software Chronicles Digest", Link: "/g/software-chronicles-digest", SubTitle: "SCD"},
				{
					LinkTitle: "Tristar & Red Sector Inc.", Link: "/g/tristar-ampersand-red-sector-inc",
					SubTitle: "TRSi", Forward: "Red Sector, then in 1991 Skid Row, TDT",
				},
			},
		},
		{
			Title: "Earliest BBS VGA loader", Year: 1991, Month: 3,
			Lead: "XTC Systems BBS", LinkTitle: "the loader", Link: "/f/a41dcd9",
			Content: "<code>XTC-AD.COM</code>",
			Picture: Picture{
				Title: "XTC Systems BBS VGA loader",
				Alt:   "XTC Systems BBS VGA loader screenshot",
				Webp:  "a41dcd9.webp",
				Png:   "a41dcd9.png",
			},
		},
		{
			Title: "Earliest contemporary cracktro", Year: 1991, Month: 3, Day: 12, Highlight: true,
			Lead: "The Dream Team Presents Blues Brothers", LinkTitle: "about and view the cracktro", Link: "/f/b249b1",
			Content: "This cracktro was released by The Dream Team - Tristar - Red Sector Inc. " +
				"Programmed by Hard Core, it is the first known cracktro on the PC platform to feature a contemporary design, " +
				"with a logo, music, and a scroller.<br>" +
				"Cracktros on the PC platform had previously been limited to mostly static logo screens, " +
				"or in the case of the earliest cracktros, no graphics at all.",
			Picture: Picture{
				Title: "Blues Brothers cracktro",
				Alt:   "Blues Brothers cracktro screenshot",
				Webp:  "b249b1.webp",
				Png:   "b249b1.png",
			},
		},
		{
			Title: "Earliest contemporary demoscene", Year: 1991, Month: 7,
			Lead: "Future Crew's Mental Surgery", LinkTitle: "about and view the demo", Link: "/f/ae24168",
			Picture: Picture{
				Title: "Mental Surgery demo",
				Alt:   "Mental Surgery demo screenshot",
				Webp:  "ae24168.webp",
				Png:   "ae24168.png",
			},
		},
		{
			Title: "Earliest elite BBStro", Year: 1991, Month: 10, Day: 21,
			Lead: "Splatterhouse BBS", LinkTitle: "about and view the BBStro", Link: "/f/b11acdf",
			Picture: Picture{
				Title: "Splatterhouse BBS BBStro",
				Alt:   "Splatterhouse BBS BBStro screenshot",
				Webp:  "b11acdf.webp",
				Png:   "b11acdf.png",
			},
		},
		{
			Year: 1991, Prefix: notable,
			List: Links{
				{LinkTitle: "Graphics Rendered in Magnificence", Link: "/g/graphics-rendered-in-magnificence", SubTitle: "GRiM"},
				{LinkTitle: "Insane Creators Enterprise", Link: "/g/insane-creators-enterprise", SubTitle: "iCE"},
				{LinkTitle: "Fairlight", Link: "/g/fairlight", SubTitle: "FLT"},
				{LinkTitle: "Fairlight DOX", Link: "/g/fairlight-dox"},
				{LinkTitle: "Licensed to Draw", Link: "/g/licensed-to-draw", SubTitle: "LTD"},
				{LinkTitle: "Nokturnal Trading Alliance", Link: "/g/nokturnal-trading-alliance", SubTitle: "NTA", Forward: "The Humble Guys"},
				{LinkTitle: "Pirates with Attitude", Link: "/g/pirates-with-attitude", SubTitle: "PWA"},
				{LinkTitle: "Relentless Pursuit of Magnificence", Link: "/g/relentless-pursuit-of-magnificence", SubTitle: "RPM"},
				{LinkTitle: "Razor 1911", Link: "/g/razor-1911", SubTitle: "RZR", Forward: "Razor / Skillion"},
				{LinkTitle: "Skid Row", Link: "/g/skid-row", SubTitle: "SR"},
				{LinkTitle: "The Dream Team", Link: "/g/the-dream-team", SubTitle: "TDT"},
				{LinkTitle: "The Humble Guys F/X", Link: "/g/thg-fx", SubTitle: "THG-FX"},
				{LinkTitle: "Ultra Tech", Link: "/g/ultra-tech", SubTitle: "UT"},
				{
					LinkTitle: "United Software Association", Link: "/g/united-software-association",
					SubTitle: "USA", Forward: "The Humble Guys",
				},
			},
		},
		{
			Year: 1992, Prefix: notable,
			List: Links{
				{LinkTitle: "Artists in Revolt", Link: "/g/artists-in-revolt", Forward: "Fairlight"},
				{LinkTitle: "Mirage", Link: "/g/mirage", Forward: "Licensed to Draw"},
				{LinkTitle: "Pirates Analyze Warez", Link: "/g/pirates-analyze-warez", SubTitle: "PWA"},
				{LinkTitle: "Pyradical", Link: "/g/pyradical"},
				{LinkTitle: "Razor Dox", Link: "/g/razordox", SubTitle: "RZR"},
				{LinkTitle: "Superior Art Creations", Link: "/g/superior-art-creations", SubTitle: "SAC"},
				{LinkTitle: "The One and Only", Link: "/g/the-one-and-only", SubTitle: "TOAO"},
			},
		},
		{
			Year: 1993, Prefix: notable,
			List: Links{
				{LinkTitle: "Drink or Die", Link: "/g/drink-or-die", SubTitle: "DOD"},
				{LinkTitle: "Hybrid", Link: "/g/hybrid", SubTitle: "HBD", Forward: "Pyradical"},
				{LinkTitle: "Legend", Link: "/g/legend", SubTitle: "LND"},
				{LinkTitle: "Paradox", Link: "/g/paradox", SubTitle: "PDX"},
				{LinkTitle: "Pentagram", Link: "/g/pentagram", SubTitle: "PTG", Forward: "Legend"},
				{LinkTitle: "Rise in Superior Couriering", Link: "/g/rise-in-superior-couriering", SubTitle: "RiSC"},
				{LinkTitle: "Scoopex", Link: "/g/scoopex"},
				{
					LinkTitle: "The Untouchables", Link: "/g/the-untouchables",
					SubTitle: "UNT", Forward: "UNiQ, XAP",
				},
			},
		},
		{
			Title: "Earliest CD image release", Year: 1994, Month: 11, Day: 17, Highlight: true,
			Lead: "ROM 1911", LinkTitle: "about the release", Link: "/f/ab3e0b",
			Content: "The earliest known release was a CD image of the game Slob Zone later known as H.U.R.L.<br>" +
				"At the time most Scene boards and FTP sites offered credits for file uploads but hard drive storage was very expensive. " +
				"So whole CD images were undesirable due to the massive file sizes involved, slow internet and even slower modem speeds. " +
				"Plus games published to CDs had little or no copy protection to crack, so were considered too easy, \"lame\" releases.<br>" +
				"ROM 1911 was used by Razor 1911 as the dumping ground for unwanted CDs titles.",
		},
		{
			Year: 1994, Prefix: notable,
			List: Links{
				{LinkTitle: "ROM 1911", Link: "/g/rom-1911", SubTitle: "ROM", Forward: "Razor 1911"},
				{LinkTitle: "Request to Send", Link: "/g/request-to-send", SubTitle: "RTS"},
				{LinkTitle: "Genesis", Link: "/g/genesis", SubTitle: "GNS", Forward: "Pentagram"},
				{LinkTitle: "TDU-Jam", Link: "/g/tdu_jam", SubTitle: "TDU", Forward: "Genesis"},
			},
		},
		{
			Title: "Earliest CD-RIP release", Year: 1995, Month: 6, Day: 3, Highlight: true,
			Lead: "Hybrid", LinkTitle: "about the release", Link: "/f/a938e5",
			Content: "The earliest known CD-RIP release was by Hybrid for the game Virtual Pool from Interplay.<br>" +
				"Hybrid was a group formed by ex-members of Pyradical and Pentagram.<br>" +
				"The CD-RIP came about due to the publishing of games on exclusive to CD-ROM, ignoring the standard floppy disk. " +
				"CD-ROMs were cheaper to produce and had more storage capacity than floppy disks. " +
				"But hard drives were expensive and whole CD image were too large to store. " +
				"So in order for many to play the game, the CD had to be <strong>ripped</strong> to the hard drive with the game fluff such as intro videos removed.",
		},
		{
			Title: "Windows 95 warez release", Year: 1995, Month: 8, Prefix: "Early",
			Lead: "Drink or Die", Link: "/f/a8177", LinkTitle: "about the release",
			Content: "Drink or Die became infamous for releasing the to warez scene, a copy of the CD media for the box retail edition of Windows 95, two weeks before the official worldwide release. " +
				"The release highlighted a significant problem for software and game publishers: some company employees were either members of these warez groups or receiving kickbacks. " +
				"<p><q>Another thing that may raise some questions is that, when you are in MS-DOS-SHELL, and you type 'ver', you will see Windows 95. " +
				"[Version 4.00.950] This does not mean Beta 950, this, in fact (<em>coming directly from my supplier's mouth at MS</em>*) means that this is version 4.0 -ergo- Windows '95.</q></p>* Microsoft",
		},
		{
			Title: "Windows 95", Year: 1995, Month: 8, Day: 24,
			Lead: "Worldwide retail release", LinkTitle: "about the day in history",
			Link:    "https://www.theverge.com/21398999/windows-95-anniversary-release-date-history",
			Content: "Microsoft's biggest and most hyped mainstream product release. It was hugely successful in the market and began the transition away from PC/MS-DOS.",
			Picture: Picture{
				Title: "Windows 95 startup",
				Alt:   "Windows 95 startup screenshot",
				Webp:  "windows-95-startup.webp",
				Png:   "windows-95-startup.png",
			},
		},
		{
			Year: 1995, Prefix: notable,
			List: Links{
				{LinkTitle: "Eclipse", Link: "/g/eclipse", SubTitle: "ECL", Forward: "Hybrid"},
				{LinkTitle: "Hoodlum", Link: "/g/hoodlum", SubTitle: "HLM"},
				{LinkTitle: "Prestige", Link: "/g/prestige", SubTitle: "PTG"},
				{LinkTitle: "Inquisition", Link: "/g/inquisition", SubTitle: "INQ", Forward: "Week in Warez"},
				{LinkTitle: "The Naked Truth", Link: "/g/the-naked-truth-magazine", SubTitle: "NTM"},
				{LinkTitle: "Razor 1911 CD Division", Link: "/g/razor-1911-cd-division", SubTitle: "RZR", Forward: "Razor 1911"},
				{LinkTitle: "Reality Check Network", Link: "/g/reality-check-network", SubTitle: "RCN"},
				{LinkTitle: "The Week in Warez", Link: "/g/the-week-in-warez", SubTitle: "WWW"},
			},
		},
		{
			Title: "The Scene merch", Year: 1996, Month: 1,
			Lead: "Razor 1911 Tenth Anniversary CD-ROM", LinkTitle: "the order form", Link: "/f/a42df1",
			Content: "The first major Scene merchandise was a CD-ROM by Razor 1911 to celebrate their 10th anniversary. " +
				"It was a collection of their PC releases from 1991 to 1995 and was sold for $40 USD each, including worldwide postage. " +
				"Each purchase required the physical cash to be sent in the mail to a PO Box in Florida.<br>" +
				"Other groups followed suit with their own merchandise, with the most popular item being t-shirts.",
			Picture: Picture{
				Title: "Razor 1911 Tenth Anniversary CD-ROM",
				Alt:   "Razor 1911 Tenth Anniversary CD-ROM disc",
				Webp:  "razor-1911-tenth-anniversary-cd-rom.webp",
				Png:   "razor-1911-tenth-anniversary-cd-rom.png",
			},
		},
		{
			Title: "First release standards", Year: 1996, Month: 2,
			Lead: "Standards of Piracy Association", LinkTitle: "the public announcement", Link: "/f/aa3b26",
			Content: "The Standards of Piracy Association (SPA) was formed by the groups " +
				"<a href=\"/g/prestige\">Prestige</a>, " +
				"<a href=\"/g/razor-1911\">Razor 1911</a>, " +
				"<a href=\"/g/mantis\">Mantis</a>, " +
				"<a href=\"/g/napalm\">Napalm</a>, " +
				"and <a href=\"/g/hybrid\">Hybrid</a>. " +
				"After 15 years of games being published on the floppy disk medium, the CD-ROM was now the standard for boxed retail games. " +
				"Unlike the floppy, CD-ROMs were too large for The Scene to copy, crack and illegally distribute. " +
				"And after a number of confusing and broken releases, the SPA was formed to create a set of standards for the release of CD-RIPs, " +
				"where an incomplete but still playable game was accepted as a valid pirated release.",
			List: Links{
				{LinkTitle: "The Faction", Link: "/f/a634e1", SubTitle: "1998"},
				{LinkTitle: "NSA", Link: "/f/a13771", SubTitle: "2000"},
			},
		},
		{
			Year: 1996, Prefix: notable,
			List: Links{
				{LinkTitle: "CD Images For the Elite", Link: "/g/cd-images-for-the-elite", SubTitle: "CiFE"},
				{LinkTitle: "Class", Link: "/g/class", SubTitle: "CLS", Forward: "Prestige"},
				{LinkTitle: "RomLight", Link: "/g/romlight", SubTitle: "RLT", Forward: "Fairlight"},
				{LinkTitle: "Zeus", Link: "/g/zeus", Forward: "Eclipse"},
				{LinkTitle: "Paradigm", Link: "/g/paradigm", SubTitle: "PDM", Forward: "Zeus"},
			},
		},
		{
			Title: "Release standards broken", Year: 1997, Month: 1, Day: 13,
			Lead: "Hybrid presents Diablo", LinkTitle: "the release", Link: "/f/ab49cd",
			Content: "The Standards of Piracy Association CD-RIP standards were broken by founding member Hybrid with this release of Diablo. " +
				"Less than a year prior, SPA had agreed that CD-RIPs should be ripped to a maximum permitted size and any titles where this wasn't possible should be skipped. " +
				"It wasn't uncommon for major games such as <a href=\"https://www.imdb.com/title/tt0131537/\">Sierra's Phantasmagoria</a> to be passed over by release groups due to their massive size and game play reliance on unrippable video and audio content.",
			List: Links{
				{LinkTitle: "Diablo from Razor 1911", Link: "/f/a72ced", SubTitle: "full CD rip"},
			},
		},
		{
			Title: "Earliest ISO release", Year: 1997, Month: 11, Day: 27, Highlight: true,
			Lead: "CD Images For the Elite", LinkTitle: "the release", Link: "/f/ad40ce",
			Content: "An ISO is a file archive format that contains the complete data of a CD, and later DVD discs. " +
				"The trading of ISOs between individuals happened for years prior, but Lords of Magic was the earliest known ISO release to The Scene. " +
				"The formalization of an ISO trading scene for software occurred in late 1997, but it took years before it became a dominate format.",
		},
		{
			Year: 1997, Prefix: notable,
			List: Links{
				{LinkTitle: "CD Images For the Elite", Link: "/g/cd-images-for-the-elite", SubTitle: "CiFE"},
				{LinkTitle: "Divine", Link: "/g/divine", SubTitle: "DVN"},
			},
		},
		{
			Year: 1998, Month: 3, Day: 31,
			Title: "Online CD keys",
			Lead:  "StarCraft by Blizzard",
			Content: "StarCraft was a hugely hyped and popular real-time strategy game by Blizzard Entertainment. " +
				"A major component of the game was its multiplayer mode, which was played online through Blizzard's Battle.net service. " +
				"This was the first retail game to be released with a CD key, a unique code that was required to play the game online.",
		},
		{
			Year: 1998, Month: 4, Day: 1,
			Title: "Starcraft", LinkTitle: "the release", Link: "/f/a9353d",
			Lead: "Razor 1911",
			Content: "The first release of StarCraft was by Razor 1911 and famed cracker Beowulf, who together released the CD-RIP of the game. " +
				"However, the release took a long time to compile and was missing the CD key, which was required to play the desirable online multiplayer. " +
				"<p><q>" +
				"Well, what can I say. This has got to be one of the hardest titles I have ever ripped. " +
				"The crack was trivial, but ripping this game involved understanding and coding utilities for Blizzard's file packer. It is ...a veritable nightmare." +
				"</q></p>",
			List: Links{
				{LinkTitle: "StarCraft Battle.NET Keymaker", Link: "/f/b321b00", SubTitle: "2 April"},
				{LinkTitle: `Starcraft *100% FIX*`, Link: "/f/b13d2c", SubTitle: "3 April"},
			},
		},
		{
			Year:  1998,
			Title: "The ISO scene picks up steam",
			Content: "The ISO scene was still in its infancy, but it grows quickly when some top-tear RIP groups start releasing within the sphere." +
				"<ul class=\"list-unstyled\"><li>" +
				"<a href=\"/f/a82c49\">Razor 1911 merged the ISO division</a> back into the Razor 1911 brand.</li><li>" +
				"The famed courier group RiSC <a href=\"/f/b04dac\">create RiSCiSO</a>, which would become one of largest ISO groups.</li><li>" +
				"<a href=\"/f/ae48b0\">PDM ISO</a> becomes the ISO division of Paradigm and famed supplier Zeus.</li><li>" +
				"And <a href=\"/g/dvniso\">DVNiSO</a> the ISO division of Divine/Deviance." +
				"</li></ul>",
		},
		{
			Year: 1998, Prefix: notable,
			List: Links{
				{LinkTitle: "Fairlight", Link: "/g/fairlight", SubTitle: "FTL"},
				{LinkTitle: "Origin", Link: "/g/origin", SubTitle: "OGN"},
				{LinkTitle: "RiSCiSO", Link: "/g/risciso", Forward: "Rise in Superior Couriering"},
			},
		},
		{
			Year: 1999, Prefix: notable,
			List: Links{
				{LinkTitle: "Razor 1911 Demo", Link: "/g/razor-1911-demo", SubTitle: "RZR", Forward: "Razor 1911"},
				{LinkTitle: "Scienide", Link: "/g/scienide", SubTitle: "SCI"},
			},
		},
		{
			Year: 2000, Prefix: notable,
			List: Links{
				{LinkTitle: "Myth", Link: "/g/myth", Forward: "Paradigm, Origin"},
				{LinkTitle: "Postmortem", Link: "/g/postmortem", SubTitle: "2001"},
				{LinkTitle: "Virility", Link: "/g/virility", SubTitle: "2001"},
				{LinkTitle: "Defacto2 website", Link: "/", SubTitle: "2003"},
				{LinkTitle: "Hoodlum", Link: "/g/hoodlum", SubTitle: "2004"},
				{LinkTitle: "Reloaded", Link: "/g/reloaded", SubTitle: "2004"},
				{LinkTitle: "Rituel", Link: "/g/rituel", SubTitle: "2005"},
				{LinkTitle: "Hatred", Link: "/g/hatred", SubTitle: "2006"},
				{LinkTitle: "Skid Row", Link: "/g/skid-row", SubTitle: "2007"},
			},
		},
		{
			Title: "Digital only scene releases", Year: 2004, Month: 10, Day: 7,
			Lead: "Counter-Strike: Source Final from Emporio", LinkTitle: "the release", Link: "/f/b1282e1",
			Content: "The online multiplayer title, Counter-Strike Source was exclusively distributed on Steam, Valve's digital distribution platform. " +
				"As there was no physical media available, this became a dubious release within The Scene and many groups didn't acknowledge Emporio's package as a legitimate \"retail\" product or a \"final\" release. " +
				"Due to the ease of supply and the constant online patching, at this time digital distribution was not well received. " +
				"<p><q>SOME may contend the fact that this is BETA. <a href=\"https://web.archive.org/web/20050208205808/http://www.steampowered.com/index.php?area=news&archive=yes&id=327\">This is the version that is released on STEAM AS FINAL</a>. " +
				"You cannot do any better than this. The ... thing with STEAM is they can easily release many patches BUT EXPECT the EMPORiO crew to bring each and every patch CRACKED to your doorstep!</q></p>",
		},
		{
			Title: "Digital distribution and online activation", Year: 2004, Month: 11, Day: 16,
			Lead: "Half-Life 2", LinkTitle: "the and view the Steam page", Link: "https://store.steampowered.com/app/220/HalfLife_2",
			Content: "Half-Life 2 was once of the most anticipated games of the decade, and it was the first major game to use Steam, Valve's digital distribution platform. " +
				"Steam was a major shift in the way games were distributed, and it was the first time a AAA game required online activation. " +
				"Steam was not well received by the gaming community, but it was a huge success for Valve, and it paved the way for other digital distribution platforms. " +
				"Half-Life 2 was simultaneously released on Steam, DVD and on CD, but all three formats required Steam activation. ",
		},
		{
			Title: "Half-Life 2 *Retail*", Year: 2004, Month: 11, Day: 28,
			Lead: "Vengeance", LinkTitle: "the release", Link: "/f/b24c10",
			Content: "Half-Life 2 was once of the most anticipated games of the decade, and it was the first major game to use Steam, Valve's digital distribution platform. " +
				"This was the first attempt to crack the Steam activation, and it used an unusual Steam client and activation emulator. " +
				"While playable, the pirated game was crippled both with slower FPS, loadtimes and a lack of multiplayer gameplay.",
			List: Links{
				{LinkTitle: "Half Life 2 DVD *Retail*", Link: "/f/a126f6"},
				{LinkTitle: "Half Life 2 trainer by Ages", Link: "/f/a63666"},
			},
		},
		{
			Title: "End of the line for RIPS", Year: 2005, Month: 10, Day: 9,
			Lead: "Farewell © Myth", LinkTitle: "the release", Link: "/f/a94129",
			Content: "The last release from Myth, a group that was founded as Zeus/<a href=\"/g/paradigm\">Paradigm</a> in 1996 and focused on ripping PC games from CD and later DVDs. " +
				"By the mid 2000s, broadband use was widespread and the desire for ripped CD or DVD games with missing content was dwindling. " +
				"Myth's longtime rival, <a href=\"/f/a53505\">Class, had already quit in early 2004</a>, and the other major ripping group, <a href=\"/g/divine\">Divine</a>, quit in 2006.",
		},
		{
			Year: 2016, Prefix: "", Highlight: true,
			Title: "The twilight of the cracktro",
			Content: "The 2000s was the last time, original quality cracktros were a common sight within The Scene, mostly thanks to a few nostaligic demosceners and piracy sceners. " +
				"However, the number of people who could and were willing to create a decent cracktro dwindled, as the skillset requirements got more specific and complex. " +
				"And so the cracktro was often forsakened for less complicated methods of displaying the release information and branding. ",
			List: Links{
				{LinkTitle: "Fairlight's 500th release", Link: "/f/a61ba0f", SubTitle: "2002"},
				{LinkTitle: "Hoodlum Cracktro #3", Link: "/f/a229a8", SubTitle: "2005"},
				{LinkTitle: "Deviance by Titan", Link: "/f/ac2ea0a", SubTitle: "2005"},
				{LinkTitle: "DEViANCE 2006", Link: "/f/b73b41", SubTitle: "2006"},
				{LinkTitle: "Skid Row by Electric Druggies", Link: "/f/a72d02", SubTitle: "2008"},
				{LinkTitle: "The Settlers 7 Cracktro by Razor 1911", Link: "/f/aa2bba", SubTitle: "2010"},
				{LinkTitle: "CORE 25k by Titan", Link: "/f/a32e91", SubTitle: "2011"},
				{LinkTitle: "Guess Who's Back? Genesis", Link: "/f/b343ed", SubTitle: "2013"},
				{LinkTitle: "Razor 1911 XT-95 Checker Cracktro", Link: "/f/b230776", SubTitle: "2016"},
			},
			Picture: Picture{
				Title: "Razor 1911 XT-95 Checker Cracktro",
				Alt:   "Razor 1911 XT-95 Checker Cracktro screenshot",
				Webp:  "b230776.webp",
				Png:   "b230776.png",
			},
		},
	}
	return m
}

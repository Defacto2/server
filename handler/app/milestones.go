package app

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
			Lead: "Altair BASIC", LinkTitle: "about Altair BASIC",
			Link: "https://time.com/69316/basic/",
			Content: "Paul Allen and Bill Gates program and sell Altair BASIC for the computer they first saw a month prior." +
				"BASIC (Beginner's All-Purpose Symbolic Instruction Code) was a programming language conceived by John Kemeny and Thomas Jurtz of Dartmouth College in early 1964 to be as approachable as possible.",
		},
		{
			Year: 1975, Month: 3, Day: 5, Title: "The first meeting of the Homebrew Computer Club",
			Lead: "Homebrew Computer Club", LinkTitle: "about the Homebrew Computer Club",
			Link:    "https://www.computerhistory.org/revolution/personal-computers/17/312/1138",
			Content: "While many technology clubs of this type for sharing ideas were common, this Silicon Valley, Bay Area group became famous for its numerous members who later became industry figures.",
		},
		{
			Year: 1976, Month: 1, Title: "Software piracy",
			Lead: "An Open Letter to Hobbyists", LinkTitle: "about the Altair 8800",
			Link:    "https://archive.org/details/hcc0201/Homebrew.Computer.Club.Volume.02.Issue.01.Len.Shustek/page/n1/mode/2up",
			Content: "Bill Gates of <em>Micro-Soft</em> writes a letter to the hobbyists of the Homebrew Computer Club requesting they stop stealing Altair BASIC.",
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
			Title: "The first commercial software for x86",
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
			Title: "The first PC", Year: 1981, Month: 8, Day: 12,
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
	}
	return m
}

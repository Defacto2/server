package app

// What ever happened to real bulletin-board systems?
// http://www.textfiles.com/100/bbsdeath.pro
// http://www.textfiles.com/100/copyprot.pro
// http://www.textfiles.com/100/krckwczt.app

// *----------------------------------*
// /      APPLESOURCE BBS LIST        \
// \      *-*-*-*-*-*-*-*-*-*-        /
// *----------------------------------*
// =                                  =
// = Applesource............ 722-4871 =
// = The Board of 'ED'.......369-5972 =
// = RCP/M of Somerville.....526-5988 =
// = Stargazer...............369-5450 =
// = Holiday Inn Cambodia....722-2711 =
// = Easy Access.............526-6044 =
// = C.I.E.E. ...............725-5028 =
// = Pirate's Paradise.......231-1655 =
// = The Two Towers..........463-7941 =
// = ??? BBS.................229-5180 =
// = Board at the Edge of Reality     =
// = ^^^.....................766-9643 =
// = The Mourge..............376-4462 =
// = The Bank BBS............376-1587 =
// = Camelot BBS.............341-7817 =
// = The Rain Forest.........721-7914 =
// = RCP/M of Rutgers Univ...932-3879 =
// = Rutgers PMS.............932-3887 =
// = Colorama................572-0617 =
// = The Spectrum............494-8209 =
// = Castle Westerburg.......238-9118 =
// =                                  =
// *----------------------------------*
// _
// http://www.textfiles.com/bbs/BBSLISTS/bbslist.txt.txt

// The Mall.....................526-0420
// Curcuit Board................369-8319
// The Two Towers...............463-7941
// The Time Machine.............572-5546
// Zeplin BBS...................526-2524
// Paradox......................725-0354
// Stargazer....................369-5450
// Colorama.....................572-0617
// 64 Depot.....................563-1867
// THE ZOO......................722-4871
// http://www.textfiles.com/bbs/BBSLISTS/bbss.txt

// *DAMIEN THE DREADED*
// *=-THE UNTOUCHABLES-=*
// http://www.textfiles.com/messages/compcond.txt

// http://www.textfiles.com/messages/compconna.txt
// 1982
// --------------------------
// - BULLETIN BOARD SYSTEMS -
// --------------------------

// PIRATE SYSTEMS
// ------ -------

// PIRATES COVE...............516-698-4008
// PIRATES TREK...............914-634-1268
// PIRATES BAY................415-775-2384
// PIRATES SHIP...............312-935-2933
// PIRATES LODGE..............516-627-9048
// PIRATES CHEST..............617-891-1349
// HARPOS BAR & GRILL.........618-254-6074
// GALAXY ONE.................215-244-0864
// SEVEN SEAS HOTEL...........516-487-1212
// TREASURE ISLAND............617-444-6821
// TREND NET..................313-775-1649
// TRADERS INN................618-656-3321
// PIRATES PARADISE...........512-477-2672
// GPI CENTRAL ONLINE.........213-556-0433
// --> UNKNOWN OPERATING PIRATE BOARDS <--
// 212-580-9688
// 313-663-0333
// 907-344-8558
// 313-547-7903

// PHREAK SYSTEMS
// ------ -------
// OHIO SCIENTIFIC-OSUNY......914-725-4060
// R A C S  I.................516-482-8491
// R A C S  III...............914-942-2638
// R A C S  IV................212-767-9881
// FANTASY FORUM..............213-827-9175

// BBSs
// ~~~
// Sherwood Forest II and III, OSUNY -- I just thought they were the greatest
//                                      systems ever.
// Pirate's Bay -- Run by Mr. KRACK-MAN, who considered himself the greatest Apple
//                 pirate that ever lived.  It's still up, for all I know.
// The 2600 Magazine BBS -- Run on a piece of Apple BBS software called
//                          TBBS.  It is there that I met David Flory.
// The Police Station -- Remember THAT one?
// The Matrix, IDI, Lunatic Labs -- Three great Bay Area Forum-PC boards.
// Catch-22 -- 25 Users, No Waiting!
// http://phrack.org/issues/33/2.html

const notable = "Notable group foundings,"

// Milestone is an accomplishment for a year and optional month.
type Milestone struct {
	Picture   Picture // Picture is an image or screenshot for a milestone.
	Prefix    string  // Prefix replacement for the month, such as 'Early', 'Mid' or 'Late'.
	Title     string  // Title of the milestone should be the accomplishment.
	Lead      string  // Lead paragraph, is optional and should usually be the product.
	Content   string  // Content is the main body of the milestone and can be HTML.
	Link      string  // Link is the URL to an article about the milestone or the product.
	LinkTitle string  // LinkTitle is the title of the Link.
	List      Links   // Links is a collection of links that are displayed as a HTML list.
	Year      int     // Year of the milestone.
	Month     int     // Month of the milestone.
	Day       int     // Day of the milestone.
	Highlight bool    // Highlight is a flag to outline the milestone.
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
			Year: 1971, Month: 11, Day: 15, Title: "The first civilian microprocessor",
			Lead: "Intel 4004", LinkTitle: "The Story of the Intel 4004",
			Link: "https://www.intel.com/content/www/us/en/history/museum-story-of-intel-4004.html",
			Content: "Intel advertises the first-to-market general-purpose programmable processor or microprocessor, the 4-bit Intel 4004. " +
				"Its main uses were in <a href=\"http://www.vintagecalculators.com/html/busicom_141-pf.html\">calculators</a>, some early automatic teller machines, and other embedded devices." +
				"<p></p>Busicom (formerly Nippon Calculating Machine Corp) <a href=\"http://nascojp.com/about.html\">first commissioned the 4004</a> as part of a chipset for its 141PF Printing Calculator. The 4000 chipset comprises four branded components: the 4001 read-only memory, 4002 RAM, 4003 shift register memory, and the <strong>4004 processor</strong>.",
		},
		{
			Year: 1972, Month: 4, Title: "The first 8-bit microprocessor",
			Lead: "Intel 8008", LinkTitle: "The Story of the Intel 8008",
			Link: "https://www.intel.com/content/www/us/en/history/virtual-vault/articles/the-8008.html",
			Content: "Intel released the world's first 8-bit microprocessor, the Intel 8008. Despite the branding, it was not an 8-bit extension of the 4-bit Intel 4004 but a new architecture." +
				"<br>Computer Terminal Corporation of Texas commissioned the new Intel chip for their cost-effective <a href=\"https://history-computer.com/datapoint-2200-guide/\">Datapoint 2200 computer terminal</a>. Designed as a dumb terminal, CTC realized it could also operate as a programmable device with a central processing unit." +
				"<br>Manufacturing issues with the 8008 and deadlines meant that the Datapoint 2200 ditched the CPU. Instead, CTC followed the common practice of building the internals from discrete transistor-transistor (TTL) logic.",
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
			Content: "Inspired by The Secrets of the Little Blue Box article, Steve Wozniak and a teenage Steve Jobs team up to build and sell dozens and dozens of the Wozniak-designed blue boxes to the students of the University of California, Berkeley. " +
				"The devices allowed users to hack and manipulate the electromechanical machines that operated the national telephone network—enabling them to call anywhere worldwide without incurring the typical prohibitively expensive costs.",
			Picture: Picture{
				Title:       "A blue box device",
				Alt:         "Blue box designed and built by Steve Wozniak.",
				Jpg:         "blue-box.jpg",
				Attribution: "Maksym Kozlenko",
				License:     "CC BY-SA 4.0",
				LicenseLink: "https://creativecommons.org/licenses/by-sa/4.0/",
			},
		},
		{
			Year: 1974, Month: 4, Title: "The first CPU for microcomputers",
			Lead: "Intel 8080", LinkTitle: "about The Intel 8008 and 8080",
			Link: "https://www.intel.com/content/www/us/en/history/virtual-vault/articles/the-8008.html",
			Content: "Intel released the 8-bit <strong>8080 CPU</strong>, its second but far more successful 8-bit programmable microprocessor. " +
				"While the Intel 8008 was the first central processing unit on a chip, the heavily upgraded Intel 8080 became the first mass-produced CPU suitable for personal microcomputing. " +
				"<br>The 8080 and its later descendants, both from Intel and competitors, meant the 8080 architecture came to dominate the 8-bit CPU market of the 1970s and 1980s.<br>" +
				"This CPU became the processing heart of the earliest popular microcomputers, the <a href=\"https://collection.powerhouse.com.au/object/167322\">Altair 8800</a>, " +
				"the <a href=\"http://oldcomputers.net/sol-20.html\">Sol-20</a>, <a href=\"https://collection.powerhouse.com.au/object/153559\">IMSAI</a>, and later in arcade machines, such as the cultural phenomenon that was <a href=\"https://www.computinghistory.org.uk/det/47162/40-Years-of-Space-Invaders/\">Space Invaders</a>.",
		},
		{
			Year: 1975, Month: 1, Title: "The first popular microcomputer",
			Lead: "Altair 8800", LinkTitle: "about the Altair 8800",
			Link: "https://americanhistory.si.edu/collections/search/object/nmah_334396",
			Content: "The worlds first popular microcomputer appears on the <a href=\"https://archive.org/details/197501PopularElectronics\">front cover of Popular Electronics</a> in the USA, the <strong>Altair 8800</strong> by MITS running on the Intel <strong>8080 CPU</strong>. " +
				"Even for the time, the Altair was a primitive device, requiring toggle on/off switches for input and blinking red LED lights for output, and there was no way to save programs. But it was the first widely available programmable computer that didn't cost an arm, a leg, or a house." +
				"<br>Eventually, with the system's popularity and its use of <a href=\"http://www.s100computers.com/History.htm#The%20S-100%20Bus\">the modular S-100 bus interface</a>, an upgraded Altair platform allowed for storage, teletype-keyboard input, printer output and displays.",
		},
		{
			Year: 1975, Month: 2, Title: "The first microcomputer software",
			Lead: "Altair BASIC", LinkTitle: "about origins of BASIC",
			Link: "https://time.com/69316/basic/",
			Content: "Paul Allen and Bill Gates program and sell <strong>Altair BASIC</strong> for the computer they first saw a month prior. " +
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
			Link: "https://www.computerhistory.org/revolution/personal-computers/17/312/1138",
			Content: "<p>While many technology clubs of this type for sharing ideas were common, this Silicon Valley, Bay Area group became famous for its numerous members who later became industry figures.</p>" +
				"<p><q>Are you building your own computer? Terminal? TV Typewriter? I/O device? or some other digital black-magic box?<br>" +
				"Or are you buying time on a time-sharing service?<br>" +
				"If so, you might like to come to a gathering of people with like-minded interests. Exchange information, swap ideas, talk shop, help work on a project, whatever...</q></p>",
			Picture: Picture{
				Title:       "Homebrew Computer Club invitation",
				Jpg:         "homebrew-computer-club.jpg",
				Attribution: "Gotanero",
				License:     "CC BY-SA 3.0",
				LicenseLink: "https://creativecommons.org/licenses/by-sa/3.0/",
			},
		},
		{
			Year: 1976, Month: 1, Title: "Software piracy", Highlight: true,
			Lead: "An Open Letter to Hobbyists", LinkTitle: "the letter",
			Link: "https://archive.org/details/hcc0201/Homebrew.Computer.Club.Volume.02.Issue.01.Len.Shustek/page/n1/mode/2up",
			Content: "<p>Bill Gates of <em>Micro-Soft</em> writes a letter to the hobbyists of the Homebrew Computer Club requesting they <u>stop stealing</u> <strong>Altair BASIC</strong>. " +
				"However, US copyright law generally did not apply to software then. Copying an application was the same as sharing the instructions of a cooking recipe.</p>" +
				"<p><q>As the majority of hobbyists must be aware, most of you steal your software. Hardware must be paid for, but software is something to share. Who cares if the people who worked on it get paid.</q></p>",
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
			Link: "https://www.computerhistory.org/revolution/personal-computers/17/312/1132",
			Content: "Steve Wozniak and Steve Jobs release the Apple&nbsp;I, a single-board computer with a <a href=\"https://spectrum.ieee.org/chip-hall-of-fame-mos-technology-6502-microprocessor\">MOS 6502 CPU</a>, 4KB of RAM, and a 40-column display controller. " +
				"<p>Unlike the more popular and earlier Altair 8800, the Apple Computer wasn't usable out of the box and didn't come with a case. However, <a href=\"https://upload.wikimedia.org/wikipedia/commons/4/48/Apple_1_Advertisement_Oct_1976.jpg\">it did offer a convenient video terminal, cassette, and keyboard interface</a>, requiring owners to supply peripherals for output, storage, and input." +
				"</p><p>The choice of the new, powerful, and affordable <strong>MOS 6502</strong> CPU showed foresight, as it later became the basis of far more successful microcomputer and consoles.<p>" +
				ul0 +
				"<li>Atari 2600 <sup>1977</sup></li>" +
				"<li>Apple II <sup>1977</sup></li>" +
				"<li>Commodore PET <sup>1977</sup></li>" +
				"<li>Commodore VIC-20 <sup>1981</sup></li>" +
				"<li>Commodore&nbsp;64 <sup>1982</sup></li>" +
				"<li>Nintendo Entertainment System <sup>1983</sup></li>" +
				ul1,
		},
		{
			Year: 1977, Month: 1, Title: "CP/M operating system",
			LinkTitle: "about CP/M", Link: "https://landley.net/history/mirror/cpm/history.html",
			Content: "Gary Kildall forms Digital Research to sell his hobbyist operating system, <strong>CP/M</strong>, for the Intel 8080. " +
				"Gary was an occasional consultant for Intel's microprocessor division, which gave him access to hardware and personnel. " +
				"CP/M became the first successful microcomputer operating system. " +
				"It dominated the remainder of the 1970s and is the default platform for most computers running an <strong>Intel 8080</strong>, <strong>8085</strong> or its compatible competitor, the <strong>Zilog Z-80</strong>." +
				"<p></p>IBM and Microsoft's later PC-DOS / MS-DOS took a lot of inspiration from CP/M and supplanted (<a href=\"https://www.wired.com/2012/08/ms-dos-examined-for-thef/\">or usurped</a>) " +
				"it as the dominant, open hardware, microcomputing operating system.",
		},
		{
			Year: 1977, Title: "The trinity of microcomputers",
			Lead: "Apple II, Commodore PET, TRS-80", LinkTitle: "about the Apple II, Commodore PET and TRS-80",
			Link: "https://cybernews.com/editorial/the-1977-trinity-and-other-era-defining-pcs/",
			Content: "<p>The <strong>Apple&nbsp;II</strong>, <strong>Commodore&nbsp;PET</strong> and <strong>TRS-80</strong> are released, the first microcomputers to be readily available to the public. " +
				"By the end of the year, a potential customer in the USA could walk into a mall or retail shop and walk out with a complete personal computer, ready to use.</p>" +
				"<p>The Commodore&nbsp;PET <small>(Personal Electronic Transactor)</small> and the Apple used the <strong>MOS 6502</strong> CPU <sup>1975</sup>, while the TRS-80 <small>(<a href=\"https://cybernews.com/editorial/from-hero-to-zero-meteoric-rise-and-fall-of-tandy-computers/\">Tandy RadioShack</a>)</small> ran off a <strong>Zilog Z-80</strong> <sup>1976</sup>.</p>",
		},
		{
			Year: 1978, Month: 2, Title: "The first Bulletin Board System",
			Lead: "CBBS", LinkTitle: "the Byte Magazine article", Link: "https://vintagecomputer.net/cisc367/byte%20nov%201978%20computerized%20BBS%20-%20ward%20christensen.pdf",
			Content: "Ward Christensen and Randy Suess create the first Bulletin Board System (<strong>BBS</strong>), the <em>Computerized Bulletin Board System</em> (<strong>CBBS</strong>) in Chicago. " +
				"The software was custom written in 8080 assembler language which ran on a <strong>S-100 bus</strong> computer together with the brand new $300, <a href=\"http://www.s100computers.com/Hardware%20Folder/DC%20Hayes/103/103%20Modem.htm\">Hayes 110/300 baud modem</a>. " +
				"The digital bulletin board became extremely popular, with callers from around the world after articles and logs were published in both Byte and Dr.&nbsp;Dobb's Journal magazines later in the year.",
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
			Content: "<p>Intel released the 16-bit programmable microprocessor, the <strong>Intel 8086</strong>, which began the <em>x86-architecture</em> and Intel PC platform.</p>" +
				"<p>In July 1976, the startup Zilog launched its first product, <a href=\"https://spectrum.ieee.org/chip-hall-of-fame-zilog-z80-microprocessor\">the Z80 CPU</a>, an enhanced, cheaper and software-compatible 8080 clone. Eventually, the Z80 became one of the most successful 8-bit CPUs. " +
				"Months later, <a href=\"https://timeline.intel.com/1976/8085-microprocessor\">Intel released the 8085</a>, an update to the 8080 CPU line, improving circuitry power requirements and reducing implementation costs.</p>" +
				"<p>The development and launch of the 8086, a software-compatible 16-bit implementation of the 8080 and the 8085, is a direct response to the Z80 and the market of clone CPUs. However, the 8086 failed to dominate an industry saturated with more affordable 8-bit hardware.</p>",
			Picture: Picture{
				Title:       "A recreation of CBBS",
				Alt:         "A recreation screen capture of the first BBS.",
				Png:         "intel-8086.jpg",
				Attribution: "Thomas Nguyen",
				License:     "CC BY-SA 4.0",
				LicenseLink: "https://creativecommons.org/licenses/by-sa/4.0/deed.en",
			},
		},
		{
			Title: "The first popular x86 CPU", Year: 1979, Month: 6,
			Lead: "Intel 8088", LinkTitle: "about the Intel 8088",
			Link: "https://spectrum.ieee.org/chip-hall-of-fame-intel-8088-microprocessor",
			Content: "Intel releases a lesser 16-bit microprocessor, the <strong>Intel 8088</strong>. " +
				"While <u>fully compatible</u> with the earlier Intel 8086 CPU, this model is intentionally \"castrated\" with an 8-bit external data bus. " +
				"The revision is an improvement for some buyers as it needs less expensive mainboard support chips and is compatible with the more readily available 8-bit hardware. " +
				"<p>Software written for either CPU often gets quoted as <a href=\"https://archive.org/details/msdos-200-users-guide-1983/page/n3/mode/2up\">8086/8088 compatible</a>.</p>",
		},
		{
			Title: "First commercial software for x86",
			Year:  1979, Month: 6, Day: 18,
			Lead: "Microsoft BASIC-86", LinkTitle: "Microsoft introduces BASIC-86",
			Link: "https://thisdayintechhistory.com/06/18/microsoft-introduces-basic-for-8086/",
			Content: "<a href=\"https://www.computerhistory.org/collections/catalog/102623976\">Microsoft BASIC</a> and its many revisions were the first killer applications for Microsoft in its early years. " +
				"Most microcomputers were sold to enthusiasts or businesses, and the software availability could have been better. " +
				"So many owners resorted to creating software, and the BASIC programming language had the easiest learning curve. " +
				"Microsoft didn't invent the language, but its implementation was considered the gold standard.",
		},
		{
			Title: "The early underground", Year: 1979, Highlight: true,
			Lead: "BBSes and the Apple II",
			Content: "<p>Before the Internet, <em>Computerized Bulletin Board Systems</em> were a primary means of online communication between microcomputer owners. In these early days, the setups allowed people to dial in using their microcomputers to share and read public or private messages with other callers.</p>" +
				"<p>The earliest <strong>CBBS</strong> setups ran off <a href=\"http://www.s100computers.com/\">S-100 bus-based computers</a>, which were simple, often incompatible microcomputers or motherboards of the 1970s manufactured by various system manufacturers that shared the same S-100 interface bus. " +
				"When the Apple&nbsp;II received CBBS-like software in 1979, it was commonly called ABBS or an Apple Bulletin Board System. " +
				"By September 1979, nationwide listings <small>(<a href=\"https://mirrors.apple2.org.za/ftp.apple.asimov.net/documentation/magazines/washington_apple_journal/washingtonapplepijournal1979v1no8sep79.pdf\">see page 3 under MODEMania</a>)</small> for dozens of bulletin boards were running on ABBS, CBBS, and other platforms.</p>" +
				"<p>A very early, underground ABBS is the 1979-1981, <a href=\"http://www.textfiles.com/phreak/tuc-intr.phk\">New Jersey-based board*</a>, <strong>Sherwood&nbsp;Forest</strong>, created by Magnetic Surfer, and it runs off a floppy disc and a Micromodem. " +
				"It became a hub for some active telephone phreakers who were early adopters of microcomputers in the New York Tri-state area—many became Scene pirates and notorious computer phreakers and hackers." +
				"<br>*<small>Other sources have claimed Sherwood Forest operated out of Manhattan, NY.</small></p>" +
				"<p>MOM, or <strong>Modem&nbsp;Over&nbsp;Manhattan</strong> (+212-245-4363), as its name suggests, was based in Manhattan, New York, and probably went online in 1980. It is another famous open board with lax subject matter that was popular with the New York phreak community.</p>" +
				"<p>The original <strong>Pirate-Trek</strong> out of New York (+914-634-1268), possibly run by the famed Apple&nbsp;II cracker Krakowicz, <a href=\"http://artscene.textfiles.com/intros/APPLEII/cyclod.gif\">was first announced</a> in 1981.</p>" +
				"<p>There is also the renowned <strong>8BBS</strong> out of San Jose, CA, which ran on a <a href=\"https://www.computerhistory.org/revolution/minicomputers/11/331\">PDP-8 minicomputer</a> in 1980-82 and has a separate article below.</p>" +
				"<p>In the early days of the BBS, the mainstream computer-press paid attention to boards, <a href=\" https://books.google.com.au/books?id=3j4EAAAAMBAJ&pg=PA10&lpg=PA10&dq=%22Modem+Over+Manhattan%22&source=bl&ots=smYwZj_okV&sig=ACfU3U0kYG9RX-3uPfGTakGgtP_mVDcAhA&hl=en&sa=X&ved=2ahUKEwiVs-yi6-qEAxX-oWMGHYpwAPA4ChDoAXoECAIQAw#v=onepage&q=%22Modem%20Over%20Manhattan%22&f=false\">including write-ups</a> and listings of the phone numbers for known underground boards*." +
				"<br>*<small>Though this article in InfoWorld mislabels <strong>8</strong>BBS as BBBS.</small></p>",
		},
		{
			Title: "The first crackers ?", Year: 1979, Highlight: true,
			Content: "<p>We have yet to learn when or who started <strong>cracking</strong>, but it would be after the introduction of copy protection on software on the Apple&nbsp;II and probably after it became more commonplace. " +
				"Andrew McFadden has <a href=\"https://fadden.com/apple2/cassette-protect.html\">written about early copy protection</a> on specific Apple&nbsp;II games released on cassette tapes in 1978 and 1979, but these were probably unusual.</p>" +
				"<p>However, Apple's July 1978 retail debut of the <A href=\"https://collections.museumsvictoria.com.au/articles/2787\">Disk II floppy drive</a> with the first <a href=\"https://www.apple2history.org/history/ah14/#01\">Apple operating system</a> was a significant factor. For the wealthy hobbyists who were Apple&nbsp;II owners, the drive and software became a must-have piece of kit that significantly improved the functionality of their machines. Compared to the initial outlay of a common Apple&nbsp;II setup, the floppy disk drive and interface were quite affordable and quickly caught on.</p>" +
				"<p>The drive offered new benefits for software developers, including speed and reliability and complete control of the floppy drive hardware using software that the developers could write themselves. This ability encouraged them to embed <a href=\"https://www.bigmessowires.com/2015/08/27/apple-ii-copy-protection/\">disk copy protection methods</a> that are <a href=\"https://paleotronic.com/2024/01/28/confessions-of-a-disk-cracker-the-secrets-of-4am/\">still problematic for computer historians today</a>!</p>" +
				"<p>A post on 8BBS (<a href=\"https://archive.org/details/8BBSArchiveP1V1/page/n60/mode/1up\">Message number 4342</a>) from Brain Litzinger at 23:37 02-Dec-80 includes," +
				"<br><code>I also have unbroken: Galaxion,<a href=\"http://artscene.textfiles.com/intros/APPLEII/mlab.gif\">Dogfight</a>,Hi-res shootout,and Astro-Apple</code>.<br>The casual use of <em>unbroken</em> in this 1980 post suggests disk copy protection and a need of cracking.</p>" +
				"<p>In Christmas 1980, <a href=\"https://www.vice.com/en/article/qjvbem/dont-copy-that-floppy-the-untold-history-of-apple-ii-software-piracy\">Omega Software Systems was advertising Lock Smith</a>, a disk copy program that makes a <em>bit-by-bit</em> copy, claiming \"duplication of just about any disk is possible.\" This advertising suggests that disk copy protection was already problematic for Apple&nbsp;II owners who desired software backups and that there was a product market. " +
				"The novel method of disk duplication implies that the anonymous Lock Smith author(s) were well-practiced in bypassing copy protection by the time of print.</p>" +
				"<p>Also, sometime in 1981 is the publication of <a href=\"http://computist.textfiles.com/\">HardCore Computing</a>. A Seattle-based, print magazine for the Apple&nbsp;II that came with \"How to back up your copy-protected disks.\" on the front cover. " +
				"Dave Alpert, the head of Omega Software Inc. and president of the <a href=\"https://archive.org/details/northernillinoisaugpaperlibrary1981\">Northern Illinois Apple Users Group</a>, " +
				"<a href=\"http://computist.textfiles.com/ISSUE.001/page-08.jpg\">is interviewed</a>, and he says Lock Smith took over a year to develop. <a href=\"http://computist.textfiles.com/ISSUE.001/page-10.jpg\">On page 10 of the issue</a>, there is a review section of disk copying programs, including \"Locksmith,\" Copy II Plus, Back-It-Up, Quick and Dirty, and Old Faithful.</p>",
		},
		{
			Title: "The birth of warez ?", Year: 1980, Highlight: true,
			Lead: "The Apple II", Link: "http://artscene.textfiles.com/intros/APPLEII/", LinkTitle: "Apple II crack screens",
			Content: "<p>There's no straightforward answer to this question, but a good guess would be <strong>sometime&nbsp;in&nbsp;1980</strong> within the known technology cities of the USA. " +
				"By then, people were exchanging real-life details, meeting to duplicate and exchange software collections, and finding ways to remove copy protections. " +
				"<a href=\"https://www.apple2history.org/history/ah13/#09\">The first Apple&nbsp;II modem peripherals</a>, the Hayes&nbsp;Micromodem&nbsp;II and the Novation&nbsp;CAT were released in 1979 and 1980, respectively. " +
				"These add-ons enabled microcomputer owners to connect to remote message boards, communicate, and even exchange files using the telephone.</p>" +
				"<p>For over a decade, long-distance <a href=\"https://www.slate.com/articles/technology/the_spectator/2011/10/the_article_that_inspired_steve_jobs_secrets_of_the_little_blue_.html\">phone phreaking had been a well-established movement</a> in the USA, and its monetary entry barrier was relatively low. " +
				"But late-1970s microcomputers were <a href=\"http://www.apple-iigs.info/doc/fichiers/Apple%20Price%20List%201978-08.pdf\">expensive</a>, exotic but mostly useless* machines, and getting them online with modems was challenging. " +
				"So it makes sense that the kids who already phreaked (make time-charged phone calls, gratis) and were into microcomputing would go out of their way to befriend anyone who shared in these niche hobbies, to exchange software, information, and ideas, even if those people resided interstate." +
				"<br>* <small>The first \"killer app\" for the Apple&nbsp;II, <a href=\"https://www.apple2history.org/history/ah18/#07\">VisiCalc, the first spreadsheet for microcomputers</a>, was only released in the last few months of 1979.</small></p>" +
				"<p>Some pirates on the Apple&nbsp;II were <a href=\"http://artscene.textfiles.com/intros/APPLEII/marscars.gif\">dating their cracks</a> in <strong>early 1982</strong>, but many tagged crack screens exist for games published in <strong>1980</strong> and 1981 " +
				"(example <a href=\"http://artscene.textfiles.com/intros/APPLEII/tcommand.gif\">1</a>, <a href=\"http://artscene.textfiles.com/intros/APPLEII/bezmanc.gif\">2</a>, <a href=\"http://artscene.textfiles.com/intros/APPLEII/borgc.gif\">3</a>, <a href=\"http://artscene.textfiles.com/intros/APPLEII/torax.gif\">4</a>.) " +
				"While a copyright year doesn't always mean the game crack is from the same period, it is a fair assumption.</p>" +
				"<p>As for the other microcomputer platforms, <a href=\"http://www.trs-80.org/was-the-trs-80-once-the-top-selling-computer/\">the far more popular TRS-80</a> from Radio Shack <a href=\"http://www.trs-80.org/telephone-interface/\">had a modem peripheral available</a> at the end of 1978. " +
				"However, there needs to be more evidence of an underground culture first developing on this platform. Modem peripherals didn't exist on the <a href=\"http://www.atarimania.com/faq-atari-400-800-xl-xe-what-other-modems-can-i-use-with-my-atari_47.html\">Atari&nbsp;400/800 until 1981</a> and the Commodore&nbsp;64 was years away.</p>",
		},
		{
			Title: "The first group", Year: 1980, Highlight: true,
			Lead: "The Apple Mafia",
			Content: "<p>In 1986, Red Ghost posted <a href=\"/f/a430f7\">The Apple Mafia Story</a>. " +
				"<br><strong>The Untouchables</strong> (examples <a href=\"http://artscene.textfiles.com/intros/APPLEII/freitagc.gif\">1</a>, " +
				"<a href=\"http://artscene.textfiles.com/intros/APPLEII/bellhop.gif\">2</a>, <a href=\"http://artscene.textfiles.com/intros/APPLEII/sraid.gif\">3</a>, " +
				"<a href=\"http://artscene.textfiles.com/intros/APPLEII/kenuston.gif\">4</a>)" +
				"<br><strong>The Apple Mafia</strong> (examples <a href=\"http://artscene.textfiles.com/intros/APPLEII/amafia.gif\">1</a>, " +
				"<a href=\"http://artscene.textfiles.com/intros/APPLEII/digem.gif\">2</a>, <a href=\"http://artscene.textfiles.com/intros/APPLEII/random.gif\">3</a>, " +
				"<a href=\"http://artscene.textfiles.com/intros/APPLEII/snoopyc.gif\">4</a>, <a href=\"http://artscene.textfiles.com/intros/APPLEII/zkeeperc.gif\">5</a>)" +
				"<br><strong>The Dirty Dozen</strong> " +
				"(examples <a href=\"http://artscene.textfiles.com/intros/APPLEII/bilestoadc.gif\">1</a>, <a href=\"http://artscene.textfiles.com/intros/APPLEII/millipedec.gif\">2</a>, " +
				"<a href=\"http://artscene.textfiles.com/intros/APPLEII/plasmania.gif\">3</a>, <a href=\"http://artscene.textfiles.com/intros/APPLEII/wargle.gif\">4</a>) " +
				"<br>In it, he claims these were some of the first-ever pirate groups. He grew up in Queens, New York, and suggests that is where many \"original\" phreakers and pirates originated. But we know that statement is partially incorrect, as in the 1970s, people nationwide, <a href=\"http://www.flyingsnail.com/images/YIPL/YIPL_002.jpg\">were already freaking</a>.</p>" +
				"<p>In the same post, an early 1984 quote from The Godfather states he founded <strong>The Apple Mafia in 1980</strong>, initially as a joke, but it became a more serious project in <strong>1981</strong>." +
				"<br><q style=\"text-transform: lowercase;\">BRIEF HISTORY OF THE APPLE MAFIA. FOUNDED IN 1980 BY THE GODFATHER AS A JOKE. REDONE IN 1981 AS A SEMI SERIOUS GROUP. " +
				"KICKED SOME ASS IN '82. BLEW EVERYONE AWAY IN 83, AND WILL DO MUCH BETTER IN 84. SINCE THE BEGINNING THE GROUP HAS DIED OUT AND BEEN REBORN SEVERAL TIMES, THIS TIME LETS KEEP IT GOING. IS CURRENTLY THE OLDEST ACTIVE GROUP,</q></p>" +
				"<p><a href=\"http://phrack.org/issues/42/3.html\">Phrack Magazine issue 42</a> has a 1993 interview with <a href=\"https://en.wikipedia.org/wiki/Patrick_K._Kroupa\">Lord Digital</a>, who clarifies the Apple Mafia founding." +
				" <q>I played around with various things, ... until " +
				"I got an Apple&nbsp;II+ in 1978. I hung out with a group of people who were also " +
				"starting to get into computers, most of them comprising the main attendees of " +
				"the soon-to-be-defunct TAP (<a href=\"http://www.flyingsnail.com/missingbbs/tap01.html\">1</a>, <a href=\"https://archive.org/details/yipltap/YIPL_and_TAP_Issues_1-91.99-100/page/n165/mode/2up\">2</a>) meetings in NYC, a pretty eclectic collection of " +
				"dudes who have long since gone their separate ways to meet with whatever " +
				"destinies life had in store for them.  Around 1980 there was an Apple Fest<strong>*</strong> that " +
				"we went to, and found even more people with Apples and, from this, formed the " +
				"Apple Mafia, which was, in our minds, really cool sounding and actually became " +
				"the first WAreZ gRoUP to exist for the Apple&nbsp;II.</q>" +
				"<br>* <small><a href=\"https://www.wap.org/journal/showcase/washingtonapplepijournal1981v3no4apr81.pdf\">The first AppleFest was actually held</a> on the weekend of June 6-7 <strong>1981</strong> in Boston.</small></p>" +
				"<p><strong>Super Pirates of Minneapolis</strong>" +
				" (examples <a href=\"http://artscene.textfiles.com/intros/APPLEII/ribbitc.gif\">1</a>, <a href=\"http://artscene.textfiles.com/intros/APPLEII/spirates.gif\">2</a>, <a href=\"http://artscene.textfiles.com/intros/APPLEII/succession.gif\">3</a>)" +
				"<br>Were a famous, early group outside of New York. Below is an uncredited quote from the Red-Sector-A BBS <a href=\"http://www.textfiles.com/bbs/boardsims2.txt\">1987 log file</a>, and the year might be misremembering. However, it does suggest the Super Pirates were around in 1980, the year of <a href=\"https://www.mobygames.com/game/47942/cyber-strike/\">Cyber Strike's publication</a>. " +
				"<br><q>The 1st ware I got was back in 1980. It was Cyber Strike. Along with about 35 other disks, most cracked by the Super Pirates!</q><br>" +
				"There is anecdotal evidence suggesting the SPoM was involved in the first-ever BBS bust, and the members left formed or joined the <strong>Midwest Pirate's Guild</strong>. " +
				"A group strongly associated with the cracker Apple Bandit and his Minneapolis-based board, The&nbsp;Safehouse&nbsp;(+612-724-7066).</p>",
		},
		{
			Title: "8BBS", Year: 1980, Highlight: true,
			Lead: "110, 150 & 300 BAUD SUPPORTED", LinkTitle: "the thousands of message logs", Link: "https://archive.org/details/8BBSArchiveP1V1/mode/1up",
			Content: "<p>In San Jose, CA, <strong>8BBS</strong> (+408-296-5799) came online in March 1980. It is one of the <a href=\"https://everything2.com/title/8BBS\">first BBS message boards</a>, which early microcomputer hobbyists posted, including some early hackers, pirates, and named-dropped phreaker personalities of the era " +
				"(<a href=\"https://privacy-pc.com/articles/history-of-hacking-john-captain-crunch-drapers-perspective.html#Early_BBS_Days\">1</a>, " +
				"<a href=\"https://www.lysator.liu.se/etexts/hacker/digital1.html\">2</a>, " +
				"<a href=\"http://phrack.org/issues/8/2.html\">3</a>, " +
				"<a href=\"http://phrack.org/issues/42/3.html\">4</a>, " +
				"<a href=\"http://phrack.org/issues/10/2.html\">5</a>, " +
				"<a href=\"https://www.theverge.com/c/22889425/susy-thunder-headley-hackers-phone-phreakers-claire-evans\">6</a>). " +
				"But what stands out about the board today is that <a href=\"https://silent700.blogspot.com/2014/12/is-this-something.html\">we have survived thousands of posts</a> from the earliest open online community that anyone in 1980 with the right hardware could access from home. These posts existed before the web, before IRC, before Usenet, before the Internet. </p>" +
				"<p><a href=\"https://archive.org/details/8BBSArchiveP1V1/page/n30/mode/1up\">Message number 3964 from CHUCK HUBERT</a><br>To ALL at 12:52 on 20-Nov-80.<br>Subject! CP/M BBS AND SOFTWARE EXCHANGE</p>" +
				"<p><a href=\"https://archive.org/details/8BBSArchiveP1V1/page/n43/mode/1up\">Message number 4177 from Kevin O'Hare</a><br>To SF (SAN FRANCISCO) PHREAKS at 23:54 on 28-Nov-80.<br>Subject: HELP?</p>" +
				"<p><a href=\"https://archive.org/details/8BBSArchiveP1V1/page/n54/mode/1up\">Message number 4311 from Len Freedman</a><br>To RICK BYRNE at 11:02 on 02-Dec-80.<br>Subject: PROG. TRADING</p>" +
				"<p><a href=\"https://archive.org/details/8BBSArchiveP1V1/page/n76/mode/1up\">Message number 4496 from Susan Thunder</a><br>To Keith Johnson at 03:39 on 07-Dec-80.<br><small>I HAVE BEEN A PHONE PHREAK FOR MANY YEARS AND I WOULD LOVE TO TRADE INFO WITH YOU!!</small></p>" +
				"<p><a href=\"https://archive.org/details/8BBSArchiveP1V1/page/n185/mode/1up\">Message number 7303 from DAVID LEE</a><br>To APPLE USERS at 16:51 on 15-Mar-81.<br>Subject: APPLE SOFTWARE</p>" +
				"<p><a href=\"https://archive.org/details/8BBSArchiveP1V1/page/n197/mode/1up\">Message number 7434 from WALTER HORAT</a><br>To DAVID LEE at 22:22 on 18-Mar-81.<br>Subject: SOFTWARE</p>" +
				"<p><a href=\"https://archive.org/details/8BBSArchiveP1V1/page/n259/mode/1up\">Message number 7853 from Sara Moore</a><br>To DAVID LEE at 05:08 on 02-Apr-81.<br>Subject: SOFTWARE</p>" +
				"<ul><li><a href=\"http://www.flyingsnail.com/missingbbs/login-8BBS.html\">A login capture from 3-Feb-1981.</a></li>" +
				"<li><a href=\"http://www.flyingsnail.com/missingbbs/CHAT-8BBS.html\">Realtime text chat with the system operator.</a></li>" +
				"<li><a href=\"http://www.flyingsnail.com/missingbbs/6116.html\">The ridiculous costs of calling from long-distance.</a></li></ul>",
		},
		{
			Title: "The first operating system for x86", Year: 1980, Month: 8,
			Lead: "Seattle Computer Products QDOS", LinkTitle: "about QDOS",
			Link: "https://www.1000bit.it/storia/perso/tim_paterson_e.asp",
			Content: "<p>Tim Paterson worked on a project at Seattle Computer Products to create an " +
				"<a href=\"http://www.s100computers.com/Hardware%20Folder/Seattle%20Computer%20Products/8086%20CPU%20Board/8086%20Board.htm\">8086 CPU plugin board for the S-100 bus standard</a>. " +
				"Needing an operating system for the 16-bit Intel CPU, he programmed a half-complete, unauthorized clone of the CP/M operating system within four months. " +
				"He called it <strong>QDOS</strong> (Quick and Dirty OS), and it sold few copies.</p>" +
				"<p>Initially, QDOS got bundled with an Intel <a href=\"http://www.s100computers.com/Hardware%20Folder/Seattle%20Computer%20Products/8086%20CPU%20Board/8086%20Board.htm\">8086 CPU and hardware package</a> for the S-100 bus. " +
				"But after poor sales, the OS was promptly renamed with the more business-friendly <a href=\"https://archive.org/details/bitsavers_seattleComanual1980_2120639/mode/2up\">86-DOS</a>.</p>",
			Picture: Picture{
				Title:       "Seattle Computer Products 86-DOS startup",
				Png:         "86-dos.png",
				License:     "CC BY-SA 4.0",
				LicenseLink: "https://creativecommons.org/licenses/by-sa/4.0/deed.en",
				Attribution: "WinWorld",
			},
		},
		{
			Title: "Motorola 68000 16-bit CPU", Year: 1980, Month: 11,
			Lead: "", LinkTitle: "about the 68000", Link: "https://spectrum.ieee.org/chip-hall-of-fame-motorola-mc68000-microprocessor",
			Content: "Available in November 1980, the famed <strong>Motorola 68000</strong> is the 16-bit successor to the 8-bit 6800 CPU from late 1974. " +
				"The Motorola series competed and operated in parallel with the incompatible Intel chips for the burgeoning microprocessor market. " +
				"And like Intel, Motorola found its 8-bit chip designs reversed-engineered, enhanced, and undercut by its other competitors." +
				"<br>But the 68000 was the 16-bit chip of the 1980s, powering everything from the <a href=\"https://www.lifewire.com/history-of-sega-genesis-dawn-729670\">Sega Megadrive/Genesis</a>, the Sega 16, the SNK NeoGeo, and various arcade games. " +
				"<br>Significantly, it was at the heart of a future generation personal computing platforms, the Apple Lisa <sup>1983</sup>, <a href=\"https://spectrum.ieee.org/apple-macintosh\">Apple Macintosh</a> <sup>1984</sup>, " +
				"<a href=\"https://daily.redbullmusicacademy.com/2017/10/atari-st-instrumental-instruments\">Atari&nbsp;ST</a> <sup>1985</sup>, and the <a href=\"https://arstechnica.com/gadgets/2007/07/a-history-of-the-amiga-part-1/\">Commodore Amiga 1000</a> <sup>1985</sup>. " +
				"These incompatible systems offered high-resolution graphics and their own mouse-driven <abbr title=\"Graphical User Interface\">GUI</abbr> operating system as standard. " +
				"At a time when the typical microcomputer or PC relied upon dated, user-hostile text interaction. &nbsp; " +
				"<span title=\"The common input prompt for an IBM or Microsoft disk operating system\"><strong>A>_</strong></span>",
		},
		{
			Title: "Computer Software Copyright Act", Year: 1980, Month: 12, Day: 12, Highlight: true,
			Lead: "Software is defined in US copyright laws", LinkTitle: "about the act",
			Link: "https://www.c2st.org/the-computer-software-copyright-act-of-1980/",
			Content: "Signed as an amendment to law by President Jimmy Carter, computer programs are defined by copyright law and enable authors to control the copying, selling, and leasing of their software." +
				"<br><a href=\"https://repository.law.uic.edu/cgi/viewcontent.cgi?article=1571&context=jitpl\">But the law was confusing</a> as software documentation and software source code are protected, but the object code or the compiled software that ran on the computer hardware is probably not.",
		},
		{
			Title: "The first PC", Year: 1981, Month: 8, Day: 12, Highlight: true,
			Lead: "IBM Personal Computer", LinkTitle: "about the IBM PC",
			Link:    "https://www.ibm.com/ibm/history/exhibits/pc25/pc25_birth.html",
			Content: "Built on the 4.77 MHz <strong>Intel 8088</strong> microprocessor, 16KB of RAM and Microsoft's PC-DOS, this expensive and underpowered machine heralds <strong>the PC platform</strong>.",
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
			Content: "<strong>Microsoft Adventure</strong> is an IBM&nbsp;PC port of the text only <em>Colossal Cave Adventure</em>.<br>" +
				"Adventure was a highly influential and popular text-only adventuring game of exploration and puzzle solving for mainframe computers of the 1970s. " +
				"Will Crowther wrote it in FORTRAN for the PDP-10 system and Don Woods at the Stanford AI Lab in California later expanded it. " +
				"The game created the interactive fiction genre, which later led to graphic adventures and story narratives in video games.",
			Picture: Picture{
				Title:       "IBM Microsoft Adventure",
				Alt:         "A photo of the 1981, Microsoft Adventure floppy disk media.",
				Jpg:         "ibm-microsoft-adventure.jpg",
				Attribution: "Jack Lightbeard & MobyGames",
				License:     "© MobyGames",
				LicenseLink: "https://www.mobygames.com/game/4074/microsoft-adventure/cover/group-3242/cover-176506/",
			},
		},
		{
			Title: "Earliest dated crack", Year: 1982, Month: 3,
			Lead:      "Dung Beetles breakage by Black Bart",
			LinkTitle: "view the crack", Link: "http://artscene.textfiles.com/intros/APPLEII/dungbeetles.gif",
			Content: "The earliest dated or timestamped crack is probably on the Apple&nbsp;II, <strong>Dung Beetles broken by Black Bart in March 1982</strong>. " +
				"<p><a href=\"http://artscene.textfiles.com/intros/APPLEII/thepirate.gif\">A screenshot</a> of an earlier serial key exists, <strong>Broken by The Pirate 09/26/81</strong>, " +
				"apparently for <a href=\"https://www.pouet.net/prod.php?which=33926\">Crush Crumble & Chomp</a>, but the file lacks context.</p>" +
				"<p><em>Broken</em>, <em>unprotect</em> or <em>crack</em> were common words used in this era to describe the removal of disk copy protection.</p>" +
				"Other early dated cracks include" +
				ul0 +
				"<li><a href=\"http://artscene.textfiles.com/intros/APPLEII/marscars.gif\">Mars Cars</a> Cracked by Mr Krac-Man 1982</li>" +
				"<li><a href=\"http://artscene.textfiles.com/intros/APPLEII/tattackm.gif\">Type Attack</a>, (B) 1982 Broken by Krakowicz NY</li>" +
				"<li><a href=\"http://artscene.textfiles.com/intros/APPLEII/aec.gif\">A.E.</a> Cracked by Mr. Krac-Man 12/17/82</li>" +
				"<li><a href=\"http://artscene.textfiles.com/intros/APPLEII/boloc.gif\">Bolo</a> Cracked -- 1982 Trystan II</a> 1982</li>" +
				"<li><a href=\"http://artscene.textfiles.com/intros/APPLEII/ccomputing.gif\">Warp Destroyer</a> The Stack of Corrupt Computing 1982</li>" +
				"<li><a href=\"http://artscene.textfiles.com/intros/APPLEII/sinterceptm.gif\">Shuttle Intercept</a> Copy/OK (B) 1982 <abbr title=\"Apple Pirated Program Library Exchange\">A.P.P.L.E.</abbr> by The Clone Stranger</li>" +
				ul1,
			Picture: Picture{
				Title: "Dung Beetles breakage by Black Bart",
				Alt:   "Dung Beetles broken screenshot",
				Png:   "dungbeetles.png",
				Webp:  "dungbeetles.webp",
			},
		},
		{
			Title: "Initial release of MS-DOS", Year: 1982, Month: 8,
			Lead: "MicroSoft Disk Operating System v1.25", LinkTitle: "about MS-DOS 1 and 1.25",
			Link: "https://www.os2museum.com/wp/dos/dos-1-0-and-1-1/",
			Content: "Microsoft releases the first edition of <strong>MS-DOS</strong> v1.25, <a href=\"https://www.os2museum.com/wp/dos/dos-1-0-and-1-1/msdos-ad-1982/\">readily available</a> to all OEM computer manufacturers. " +
				"Prior releases were exclusive to IBM. The next release, MS-DOS 2, is also sold boxed at retail and will help Microsoft to become the de facto operating system provider for personal computers." +
				"<p>In 2014, the Computer History Museum <a href=\"https://computerhistory.org/blog/microsoft-ms-dos-early-source-code/\">published the source code</a> for this operating system edition, and Microsoft later made a <a href=\"https://github.com/microsoft/MS-DOS\">GitHub repository</a>.</p>",
			Picture: Picture{
				Title:       "Compaq's MS-DOS based on MS-DOS v1.25",
				Alt:         "MS-DOS Version 1.12 for the Compaq Personal Computer.",
				Jpg:         "ms-dos-floppy-disks.jpg",
				Attribution: "Brian R. Lueck",
				License:     "public domain",
				LicenseLink: "https://en.wikipedia.org/wiki/MS-DOS#/media/File:Compaq_mddos_ver1-12.jpg",
			},
		},
		{
			Title: "The earliest cracktro?", Year: 1982, Highlight: true,
			Content: "A <strong>crack intro</strong> should be defined as a computer animation advertising the crackers of a pirated software release. " +
				"So, modified \"<a href=\"http://artscene.textfiles.com/intros/APPLEII/thepirate.gif\">broken by</a>\" texts or <a href=\"http://artscene.textfiles.com/intros/APPLEII/flockland.gif\">static images</a> do not count.</p>" +
				"<p>Unfortunately, it is challenging to date early pirated releases for the PC, Commodore&nbsp;64, or Apple&nbsp;II. Many crackers didn't date their releases, and the systems themselves didn't track time or stamp the files. " +
				"<a href=\"http://artscene.textfiles.com/intros/APPLEII/.thumbs.html\">But given the proliferation</a> of \"broken by\" texts and graphic hacks in 1980-81 and 1982 on the Apple II in the USA, the early crack intro probably evolved here.</p>" +
				"<p>The prolific, early Apple cracker <strong>Mr. Xerox</strong> probably created one of the first intro and scrollers in <a href=\"https://archive.org/details/a2_Starblaster_19xx_C_G_cr_Star_Trek_1983_Sega_cr_Shuttle_Intercept_19xx__cr\">his crack screen</a> " +
				"for <strong>Star Blaster</strong> (c) 1981, which you can compare to the <a href=\"https://archive.org/details/Starblaster4amCrack\">original opening</a>.</p>" +
				"<p>Or cracker <strong>Copycatter</strong> <a href=\"https://archive.org/details/a2_Pro_Football_The_Gold_Edition_1982_System_Design_Lab_cr_Copycatter\">may have created</a> the first scroller in a release of <strong>Pro Football</strong> (c) 1982. " +
				"While newer, the February 1984 <strong>Black Belt</strong> release from <strong>The Apple Mafia</strong> <a href=\"https://archive.org/details/3d0g_022b_Black_Belt\">is a candidate</a> for an early intro, given it has a timestamp and from a well-known group. " +
				"Penqueriel Mazes by Electronic Dimension initially looked like <a href=\"https://archive.org/details/a2_Penqueriel_Mazes_19xx_Sadistic_cr_Electronic_Dimension\">a candidate</a>, but the intro-loader effects are far too modern for the (c) 1982.</p>",
			Picture: Picture{
				Title: "Mr. Xerox's Star Blaster cracktro",
				Png:   "starblaster-mr-xerox.png",
			},
		},
		{
			Title: "Berlin Bear controversy", Year: 1982,
			Content: "<p>Many long argued in the Demoscene that a \"1982\" Berlin Cracking Service <a href=\"https://www.atlantis-prophecy.org/recollection/gfx/BCS.png\">image of the Berlin Bear</a> was the first cracktro (" +
				"<a href=\"https://www.atlantis-prophecy.org/recollection/?load=interviews&id_interview=7\">1</a>, " +
				"<a href=\"https://csdb.dk/release/?id=35670\">2</a>, " +
				"<a href=\"https://m.pouet.net/prod.php?which=17555\">3</a>, " +
				"<a href=\"https://intros.c64.org/main.php?module=showintro&iid=156\">4</a>). " +
				"But this seems far-fetched, and <a href=\"https://www.atlantis-prophecy.org/recollection/?load=crackers_map&country=germany\">anecdotal proof</a> suggests it originates from 1984.</p>" +
				"<p>In late 1982, the Commodore&nbsp;64 was a pricey machine that <a href=\"https://www.power8bit.com/assets/images/screen-shot-2023-03-27-at-5.27.40-pm-508x698.webp\">targeted business users</a> <small>(<a href=\"https://www.power8bit.com/C64.html\">ad source</a>)</small> in the USA and Japan (where it flopped). " +
				"Due to last-minute design changes and poor quality assurance issues, the machine had limited distribution and software that year, according to the book Commodore: a company on the edge. </p>" +
				"<p>By all accounts, the Commodore Braunschweig factory didn't have the (European friendly) PAL Commodore&nbsp;64 machines ready for retail <a href=\"https://www.zock.com/8-Bit/D_C64.HTML\">until 1983</a>. " +
				"In the UK, advertising first hit <a href=\"https://web.archive.org/web/20160611085947if_/http://archive.6502.org/publications/commodore_computing_intl/commodore_computing_intl_1983_02.pdf\">Commodore&nbsp;Computing in February 1983</a> <small>(pages 36-37)</small> " +
				"that focus on the <a href=\"https://static.nosher.net/archives/computers/images/comm64_comci_1983-02-m.jpg\">business&nbsp;market</a>&nbsp;<small>(<a href=\"https://nosher.net/archives/computers/comm64_comci_1983-02?idx=Designed\">ad&nbsp;source</a>)</small> and potential developers, " +
				"and a different business focus advert ran in October 1983 <small>(<a href=\"https://worldradiohistory.com/UK/Practical-Computing/80s/Practical-Computing-1983-10-S-OCR.pdf\">pages 74-75</a>)</small>.</p>" +
				"<p>It seems unlikely that a number of isolated West Berlin kids had access to the unreleased PAL&nbsp;Commodore&nbsp;64 <u>in 1982</u>. The cited year is presumably late 1983 or 1984, misremembered. " +
				"However, the claim is that various Berliners got together in 1982 and formed multiple \"cracking\" groups, to also trade in software and create art on an unreleased business microcomputer, even though most software didn't need cracking then.</p>" +
				"<small>Select quotes from an often referenced <a href=\"https://www.atlantis-prophecy.org/recollection/?load=interviews&id_interview=7\">interview conducted in 2005-06</a>.</small>" +
				ul0 +
				"<li><q>The first intro was a picture of the Berlin Bear from the city flag and was released by BCS in 1982. It was a kind of co-production by several people...</q></li>" +
				"<li><q>A positive moment for me was when the father of a classmate who was working for Apple, was able to through his connections, obtain a C64 on Christmas 1982... Plutonium&nbsp;<u>Crackers</u> 2001 (PC) was born.</q></li>" +
				"<li><q>Some of our close friends/posse in Berlin started their C64 scene-careers nearly at the same time. I'm speaking of <u>Cracking</u> Force Berlin (CFB)... and&nbsp;Berlin&nbsp;<u>Cracking</u>&nbsp;Service&nbsp;(BCS).</q></li>" +
				"<li><q>We were primarily <u>cracking games from 1982</u> until late 1987.</q></li>" +
				"<li><q>Copying games wasn't really illegal in most countries back in 1982 or 1983. ... Most early releases <strong>weren't \"cracked\"</strong>, they were just released or spread.</q></li>" +
				ul1,
		},
		{
			Title: "Third-party PC games", Year: 1982,
			Content: "<p>The first set of published games on the PC platform is sold without IBM's involvement.</p>" +
				"Some early publishers include" +
				ul0 +
				"<li><a href=\"//s3data.computerhistory.org/brochures/broderbund.software.1982.102646180.pdf\">Brøderbund</a></li>" +
				"<li><a href=\"//archive.org/details/avalon-hill-game-company-catal-fall-1982\">The Avalon Hill Game Company</a></li>" +
				"<li><a href=\"//archive.org/details/strategic-simulations-inc-summer-1982-catalog/mode/2up\">Strategic Simulations</a>, Inc.</li>" +
				"<li><a href=\"//www.uvlist.net/companies/info/1023-Windmill+Software\">Windmill Software</a></li>" +
				"<li><a href=\"//retro365.blog/2019/09/23/bits-from-my-personal-collection-the-original-ibm-pc-and-orion-software/\">Orion Software</a></li>" +
				"<li><a href=\"//www.uvlist.net/companies/info/1029-Spinnaker+Software\">Spinnaker Software</a>" +
				ul1,
		},
		{
			Title: "The great online reboot", Year: 1983, Month: 1, Day: 1,
			Lead: "Internetworking", LinkTitle: "the notable computer networks", Link: "https://dl.acm.org/doi/pdf/10.1145/6617.6618",
			Content: "On the 1st of January, 1983, the US Department of Defense coordinated the massive shutdown of its existing experimental wide-area network, <abbr title=\"Advanced Research Projects Agency Network\">ARPAnet</abbr>." +
				"<p>Referred to as \"Flag Day,\" the event meant that all systems associated with the US military network were required to reconnect using a new <abbr title=\"Transfer Control Protocol\">TCP</abbr>/<abbr title=\"Internetwork Protocol\">IP</abbr> protocol.</p>" +
				"<p>The replacement protocol decentralized the networks operations and is somewhat inspired by the earlier French <a href=\"https://www.inria.fr/en/between-stanford-and-cyclades-transatlantic-perspective-creation-internet\">CYCLADES packet-switch network</a>. By demanding the connected hosts handle data delivery and error correction. The connecting of various academic, research, and commercial computer networks is possible, removing the ARPAnet's expense and inability to scale.</p>" +
				"<p>Later in the year, <a href=\"https://www.washingtonpost.com/archive/business/1983/10/04/big-computer-network-split-by-pentagon/d12feaba-c0c7-45fb-a851-25267f8dca9c/\">due to a fear of civilian hackers</a>, the systems associated with the US military were to disconnect again and join a new isolated Defense Data Network (MILnet). The few remaining non-military systems that adopted the TCP/IP protocol standard formed the basis of the new ARPA internetwork or APRA Internet.</p>" +
				"<p>The other alternative networks of the era:</p>" +
				"<abbr title=\"Because It's Time NETwork\">BITNet</abbr> <sup>1981</sup>, a cross-continental, research center and university network for file transfers and messaging." +
				"<br><abbr title=\"European Unix Network\">EUnet</abbr> <sup>1982</sup>, the first public wide area network of Europe." +
				"<br>Janet <sup>1984</sup>, an extensive UK academic network." +
				"<br>Corporate networks from Xerox Internet, DEC Easynet and IBM VNET.",
		},
		{
			Title: "The year of the Commodore 64", Year: 1983, Month: 1,
			Lead: "Computers goes mainstream", LinkTitle: "about the Commodore 64", Link: "http://variantpress.com/books/commodore-a-company-on-the-edge/",
			Content: "January 1983 saw the beginning of the juggernaut that would be the <strong>Commodore&nbsp;64</strong> microcomputer, a platform that became the world's best-selling computer for decades. " +
				"It released in limited numbers in August 1982 for the US market, but sales blew up in the lead to Christmas, and with multiple mass price cuts, it became a massive worldwide success in the following years. " +
				"<p>The Commodore&nbsp;64 became the first mass-market computer and piracy platform.</p>" +
				"Ironically, it is a Scene that at least partly materialized out of Commodore itself, according to Brian Bagnall's book On the Edge. For <a href=\"https://computerhistory.org/profile/bil-herd/\">Bil Herd</a>, " +
				"\"The worst thing you could do was submit a copy of something to the (Commodore) games and applications group.\" " +
				"He felt several bad actors were employed in that department, claiming that by late 1983, \"There were a few nefarious types that would generally make sure a cracked version of the game was available within a week.\"",
		},
		{
			Title: "The first PC clone", Year: 1983, Month: 3,
			Lead: "COMPAQ Portable", LinkTitle: "the advertisement",
			Link: "https://www.computerhistory.org/revolution/personal-computers/17/302/1194",
			Content: "Compaq Computer Corporation releases the first <strong>IBM&nbsp;PC compatible</strong> computer, the Compaq Portable. " +
				"It is the first PC clone to use the same software and expansion cards as the IBM&nbsp;PC.",
		},
		{
			Title: "ANSI.SYS, the unfinished software that leads to ANSI art", Year: 1983, Month: 3,
			Lead: "PC-DOS and MS-DOS version 2 are released", LinkTitle: "about MS-DOS ANSI.SYS",
			Link: "https://github.com/microsoft/MS-DOS/blob/master/v2.0/source/ANSI.txt",
			Content: "<p>For the first time, the IBM&nbsp;PC includes a device driver to view <strong>ANSI text graphics</strong> in color on a microcomputer.</p>" +
				"<p>ANSI was a text terminal display standard from the mid-1970s that formatted onscreen text and controlled cursor movement. The implementation in DOS was only partially complete but became its own sub-standard over time.</p>",
		},
		{
			Title: "The earliest cracked PC game", Year: 1983,
			Lead: "Atarisoft presents: Galaxian broken by The Koyote Kid", LinkTitle: "about and view the crack",
			Link: "/f/ab2edbc", Highlight: true,
			Content: "<p>This modified Galaxian title screen is known as a <strong>crack screen</strong> and was a typical way for crackers on the Apple&nbsp;II to credit themselves. Crackers modified and removed disk copy protection from software for the sole purpose of allowing duplication.</p>" +
				"<p>The online Apple II community commonly used the verbs \"broken\" or unprotected, cracked, and kracked. Given the popularity of the IBM&nbsp;PC in the USA, it is most likely The Koyote Kid was based in the USA and also interacted in the <a href=\"#the-first-crackers\">Apple II underground</a> Scene.</p>" +
				"<p>Atarisoft released Galaxian on a floppy disk for IBM&nbsp;PC in 1983. Compared to the many other console and microcomputer ports, the PC conversion of a highly successful arcade title lacked color and sound.</p>" +
				"<p><a href=\"https://www.mobygames.com/game/137/galaxian/screenshots/pc-booter/951/\">The original text</a> read <code>(C) 1983 ATARI, INC. PRESS SPACE TO CONTINUE.</code></p>",
			Picture: Picture{
				Title: "Galaxian broken by Koyote Kid",
				Alt:   "Galaxian broken screenshot",
				Webp:  "ab2edbc.webp",
				Png:   "ab2edbc.png",
			},
		},
		{
			Title: "Major videogame publishers enter the PC market", Year: 1983,
			Content: "<p>1983 saw some major arcade and video game publishers release software on the PC. Despite the business-centric marketing of the platform, game software sold on a floppy disk was a popular seller. " +
				"For publishers, it is less risky than manufacturing the expensive cartridges required by some other game systems.</p>" +
				ul0 +
				"<li><a href=\"//dfarq.homeip.net/atarisoft-if-you-cant-beat-em-join-em/\">Atarisoft</a></li>" +
				"<li><a href=\"//www.uvlist.net/companies/info/243-Infocom\">Infocom</a></li>" +
				"<li><a href=\"//www.resetera.com/threads/lets-look-back-at-game-company-datasoft.587093/##post-87110411\">Datasoft</a></li>" +
				"<li><a href=\"//www.uvlist.net/companies/info/83-Mattel%20Electronics\">Mattel</a></li>" +
				"<li><a href=\"//www.wired.com/story/sierra-online-ken-williams-interview-memoir/\">Sierra On-Line</a></li>" +
				ul1,
		},
		{
			Title: "Earliest unprotect text", Year: 1983, Month: 5, Day: 12, Highlight: true,
			Lead: "Directions by Randy Day for unprotecting SPOC the Chess Master", LinkTitle: "the unprotect text",
			Link: "/f/a91c702",
			Content: "<code>SPOC.UNP</code><br>" +
				"<p><strong>Unprotects</strong> were text documents describing methods to remove software (floppy) disk copy protection. " +
				"Many authors were legitimate owners who were frustrated that publishers would not permit them to create backup copies of their expensive but fragile 5¼-inch floppy disks for daily driving.</p>" +
				"<p><q>The disk is close to a normal disk. There is one file in the directory, spoc.exe, which is most of the program. However, track 20, sector 5 is a bad sector. In what manner it is bad, I don't know, but nothing can read it.</q></p>" +
				"<p>The origins of the unprotected document go back to the Apple&nbsp;II and other early microcomputer platforms, where BBS users would publically post simple hacks to defeat basic disk copy protection schemes, <a href=\"http://www.textfiles.com/apple/parameters.txt\">such as this 1982 log</a>.</p>",
		},
		{
			Title: "Microsoft Windows announced", Year: 1983, Month: 11, Day: 10,
			Link:      "https://www.poynter.org/reporting-editing/2014/today-in-media-history-in-1983-bill-gates-and-microsoft-introduced-windows/",
			LinkTitle: "about the announcement",
			Content: "<p>Around this time, <abbr title=\"graphical user interface\" class=\"initialism\">GUI</abbr> for microcomputing was all the hype within the technology industry and media. " +
				"In hindsight, this premature announcement from Microsoft aimed to keep customers from jumping ship to competitor platforms and GUI offerings.</p>" +
				"<p>It took a decade before graphical interfaces on the PC replaced text in business computing with Windows NT <sup>1993</sup> and even longer with Windows 95 <sup>1995</sup> before it became commonplace in the home." +
				" Other microcomputer platforms, such as the <span class=\"text-nowrap\">Apple&nbsp;Macintosh <sup>1984</sup></span>, <span class=\"text-nowrap\">Commodore Amiga</span> and <span class=\"text-nowrap\">Atari&nbsp;ST <sup>1985</sup></span> came with a GUI as standard.</p>",
		},
		{
			Title: "First, dial-up Internet connections", Year: 1984,
			Link:      "https://networkencyclopedia.com/serial-line-internet-protocol-slip/",
			LinkTitle: "about SLIP",
			Content: "Rick Adams created the Serial Line Internet Protocol (<strong>SLIP</strong>), the industry-standard protocol to connect dial-up modems to the Internet. " +
				"This protocol allowed for the creation of Internet Service Providers, which provided Internet connections over standard copper telephone lines." +
				"<br>In 1987, Rick would also go on to found one of the earliest ISPs, UUNET. " +
				"Which in the following year would offer the first commercial connection to the Internet.",
		},
		{
			Title: "Major game publishers enter the PC market", Year: 1984,
			Content: "<p>Electronic Arts, Activision, Sega, and MicroProse Software publish on the platform.</p>" +
				ul0 +
				"<li><a href=\"//www.polygon.com/a/how-ea-lost-its-soul/\">Electronic Arts</a> was founded in 1982 by former Apple employee Trip Hawkins and initially developed for the Atari&nbsp;400/800 and later Commodore&nbsp;64.</li>" +
				"<li><a href=\"//www.ign.com/articles/2010/10/01/the-history-of-activision\">Activision</a> originated in late 1979 as the first 3rd-party developer for the Atari&nbsp;2600, comprising former Atari employees.</li>" +
				"<li><a href=\"//segaretro.org/IBM_PC\">Sega</a> was a significant arcade game developer.</li>" +
				"<li><a href=\"//corporate-ient.com/microprose/\">MicroProse Software</a> was the company founded by Sid Meier and Bill Stealey in 1982 to create games for the Atari&nbsp;400/800.</li>" +
				ul1,
		},
		{
			Title: "The first 16 color PC game", Year: 1984, Month: 8,
			Lead: "King's Quest", LinkTitle: "the game manual",
			Link: "http://www.sierrahelp.com/Documents/Manuals/Kings_Quest_1_IBM_-_Manual.pdf",
			Content: "The first PC game to use 16 colors, <a href=\"https://www.mobygames.com/game/122/kings-quest/screenshots/pc-booter/\">King's Quest</a>, is created by Sierra On-Line and released by IBM. " +
				"IBM&nbsp;PC graphics cards are limited to monochrome or 4 colors, but the game is released for the new <strong>IBM&nbsp;PCjr</strong> that displays upto <strong>16 colors</strong>. " +
				"The other pioneering aspect of the game was the pseudo-3D landscape. The player controlled a human avatar from a 3rd person perspective and could use it to walk around set pieces, both in front and from behind, and interact with the onscreen objects.",
		},
		{
			Title: "The earliest information text", Year: 1984, Month: 10, Day: 17, Highlight: true,
			Lead:      "SOFTWARE PIRATES Inc. - ZORKTOOLS 1.0",
			LinkTitle: "the information text",
			Link:      "/f/ae2da98",
			Content: "<code>INFOCOM.DOC</code><br>" +
				"<p><strong>Information texts</strong> were documents stored as plain text and included in a release describing how to use a utility program or game.</p>" +
				"<p>The author of this document is part of <a href=\"http://localhost:1323/g/software-pirates-inc\">Software Pirates Inc.</a>, one of the earliest known groups on the PC underground, dating back to at least 1984. " +
				"Whether an individual or collective, the brand was prolific in writing documentation and coding utilities for the PC but kept themselves anonymous.</p>",
		},
		{
			Title: "EGA graphics standard", Year: 1984, Month: 10,
			Lead: "16 color, 64 color palette, 640x350 resolution!?", LinkTitle: "How 16 colors saved PC gaming",
			Link: "https://www.custompc.com/retro-tech/ega-graphics",
			Content: "The Enhanced Graphics Adapter standard includes 16 colors, 640×350 resolution and 80×25 text mode." +
				"<p><a href=\"http://nerdlypleasures.blogspot.com/2014/01/simcity-for-dos-swiss-army-knife-of.html\">With the odd exception</a>, most PC games that use <strong>EGA</strong> only ever support 160x200 or 320x200 resolutions with 4 or 16 colors on screen. " +
				"There were complications with EGA and its expensive monitor displays, plus the expensive memory requirements needed for higher resolution graphic modes with <strong>16 colors</strong>.</p>",
		},
		{
			Title: "An early demonstration on the PC", Year: 1984, Month: 10,
			Lead: "Fantasy Land EGA demo by IBM", LinkTitle: "and run the demo",
			Link: "https://www.pcjs.org/software/pcx86/demo/ibm/ega/",
			Content: "The first <strong>demo program</strong> on the PC, Fantasy Land, is created by IBM to demonstrate the new <strong>EGA</strong> graphics standard. " +
				"The idea of a demo is to have the program run automatically, without user input, to show off the capabilities of the hardware.",
		},
		{
			Prefix: "The earliest PC groups,", Year: 1984,
			List: Links{
				{LinkTitle: "Against Software Protection <small>ASP</small>", Link: "/g/against-software-protection"},
				{LinkTitle: "Software Pirates Inc <small>SPi</small>", Link: "/g/software-pirates-inc"},
			},
		},
		{
			Title: "The earliest text loader", Year: 1985, Month: 5, Day: 26, Highlight: true,
			Lead:      "Spy Hunter cracked by Spartacus",
			LinkTitle: "and view the text loader",
			Link:      "/f/aa2be75",
			Content: "<p><strong>Loaders</strong> are bits of code that crackers and pirate groups insert to promote themselves and their game releases. As the name suggests, they are loaded and shown before the game starts. " +
				"Loaders first originated on the Apple&nbsp;II and later the Commodore&nbsp;64 piracy Scenes.</p>" +
				"<p>While text loaders and ANSI art look similar, the execution is entirely different. ANSI art relies on plain text files encoded with ASCII escape control codes. " +
				"In contrast, text loaders are computer applications that use the computer's text characters stored in the system <a href=\"https://minuszerodegrees.net/video/bios_video_modes.htm\">graphics card ROM</a>, acting as a text programming interface.</p>" +
				"<p>Little is known about the Imperial Warlords that released this 1984 PC game port, though the two BBS advertised are from San Francisco and Minneapolis, which suggests a national group.</p>",
			Picture: Picture{
				Title: "Spy Hunter",
				Alt:   "Spy Hunter by Imperial Warlords screenshot",
				Webp:  "aa2be75.webp",
				Png:   "aa2be75.png",
			},
		},
		{
			Title: "Earliest ANSI ad", Year: 1985, Month: 8, Highlight: false,
			Lead: "The Game Gallery 300 1200", LinkTitle: "and view the file",
			Link: "/f/ba2bcbb",
			Content: "<p>The earliest <strong>ANSI ad</strong>vert is for the Manhattan based BBS, The Game Gallery (+212-799-6987). ANSI art is a computer art form that became widely used to create art and advertisements for online bulletin board systems.</p>" +
				"<p>The output uses ANSI escape codes, a standard Digital Equipment Corporation (DEC) pioneered for its <a href=\"https://vt100.net/dec/vt_history\">minicomputer video terminals</a>. Later, it was used on IBM and other PCs using software drivers and video <a href=\"https://vt100.net/emu/\">terminal emulators</a>.</p>",
			Picture: Picture{
				Title: "The Game Gallery",
				Alt:   "The Game Gallery ad screenshot",
				Webp:  "ba2bcbb.webp",
				Png:   "ba2bcbb.png",
			},
		},
		{
			Title: "Razor 1911 is named", Year: 1985, Month: 11,
			Lead: "On the Commodore 64", LinkTitle: "about the early days of Razor 1911",
			Link: "https://csdb.dk/group/?id=431",
			Content: "<p><strong>Razor 1911</strong>, the oldest and most famed brand in the Scene, was founded in <strong>Norway</strong> and has three members. " +
				"The group released demos and later cracked exclusively for the Commodore&nbsp;64 and then the Amiga. Co-founder Sector 9 took the brand to the <a href=\"/f/a12d5e\">PC in late 1990</a>.</p>" +
				"<p>The distinctive number suffix was a fad with groups of the Commodore&nbsp;64 era*. <q>1911</q> denotes the decimal value of hexadecimal <code>$777</code>.</p>" +
				"<p><small>* other named examples, 1001 Crew, 1701 Crackware, The Gamebusters 1541, The Professionals 2010.</small></p>",
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
			Content: "<strong>Microsoft Windows</strong> 1.0 was released but failed in the market. The expensive, minimum hardware requirements and a lack of software led to lackluster sales. It will take a decade and multiple releases before Windows becomes dominant.",
			Picture: Picture{
				Title: "Microsoft Windows 1.01",
				Alt:   "Microsoft Windows 1.01 booting up screenshot",
				Webp:  "windows-version-1.webp",
				Png:   "windows-version-1.png",
			},
		},
		{
			Title: "The earliest \"DOX\"", Year: 1986, Highlight: true,
			Lead: "Dam Buster documentation by Brew Associates", LinkTitle: "the documentation",
			Link: "/f/a61db76",
			Content: "<code>DAMBUST1.DOC</code><br>" +
				"<p><strong>DOX</strong> is an abbreviation for documentation, which are text files that provide instructions on playing more complicated games. " +
				"Games not in the arcade or action genre were usually unintuitive and relied on printed gameplay instruction manuals sold with the purchased game box to be usable. " +
				"These titles often relied on <a href=\"https://archive.org/details/extras_msdos_Microsoft_Flight_Simulator_v1.0_1982/mode/2up\">printed instruction manuals</a> included in the purchased game box to be usable.</p>" +
				"<p>Piracy groups had been including forms of gameplay instructions as text documents for the more complicated game releases for years, so it is unlikely this example is the first DOX.</p>" +
				"<p><q>The primary reason for the writing of this file is the fact that people may not be fully appreciating the Dam Buster game.  I have seen some documentation out, but it is lame at best. What I have given you here is the actual text of the actual documentation distributed with the game. Enjoy!</q></p>" +
				"<p>Dam Buster is a misname of <a href=\"https://archive.org/details/msdos_The_Dam_Busters_1985\">The Dam Busters</a>, a 1984-85 game published by Accolade.</p>",
		},
		{
			Title: "PC clone sales pickup in Europe", Year: 1986,
			Link:      "https://www.computerhistory.org/revolution/personal-computers/17/302",
			LinkTitle: "about the PC clone market",
			Content: "While the Commodore, Apple and IBM are common platforms in the US, the European market doesn't always share the same popular platforms. " +
				"Import duties, slow international distribution channels and a lack of localized software and hardware often hampers the adoption of some platforms. " +
				"<br>The Western European market is dominated by Acorn, Amstrad, Commodore, Sinclair but the PC clones produced by local electronic manufactures gain popularity. " +
				"Popular machines include the <a href=\"https://www.dosdays.co.uk/computers/Amstrad%20PC1000/amstrad_pc1000.php\">Amstrad PC1512</a>, " +
				"the Philips P2000T and the <a href=\"https://www.dosdays.co.uk/computers/Olivetti%20M24/olivetti_m24.php\">Olivetti M24</a>.",
			Picture: Picture{
				Title:       "The Olivetti M24",
				Jpg:         "olivetti-m24.jpg",
				Attribution: "Federigo Federighi",
				License:     "CC-BY-4.0",
				LicenseLink: "https://creativecommons.org/licenses/by-sa/4.0",
			},
		},
		{
			Title: "The first PC virus", Year: 1986, Month: 1, Day: 19,
			Lead: "Brain", LinkTitle: "about the Brain virus",
			Link:    "https://www.f-secure.com/v-descs/brain.shtml",
			Content: "The first PC virus, <em>Brain</em>, infects the boot sector of floppy disks. It acted more as annoying spam than a nefarious application, but it did have the unwanted consequence of slowing down the systems it infected.",
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
			Link:    "https://www.mobygames.com/game/152/mean-18/",
			Content: "It may seem strange today, but golf games were popular in the 1980s and 1990s. The real-life sport was aspirational for plenty of white collar US and Japanese workers, so it isn't surprising that video game golf simulations targeting expensive computer platforms and arcades became popular.",
			Picture: Picture{
				Title:       "Mean 18",
				Alt:         "Mean 18 by Accolade screenshot",
				Png:         "mean-18-ega.png",
				License:     "© Accolade",
				LicenseLink: "https://www.mobygames.com/game/152/mean-18/cover-art",
				Attribution: "Trixter & MobyGames",
			},
		},
		{
			Title: "The earliest PC loaders", Year: 1986, Month: 6, Highlight: true,
			Content: "<p><strong>Loaders</strong> were named as they would be the first thing to load and display each time the cracked game is run. " +
				"These screens were static images in the early days and sometimes contained ripped screens from other games. Some users found these annoying and a cause of unwanted file bloat.</p>" +
				"<p>The first static loaders originated on the Apple&nbsp;II underground, such as <a href=\"http://artscene.textfiles.com/intros/APPLEII/cbaseball.gif\">this example</a> by The Digital Gang for the crack release of Championship Baseball that likely came out in 1983.</p>",
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
			Content: "<strong>Fairlight</strong>, one of the oldest brands in the Scene, is founded in <strong>Sweden</strong> with just three members. " +
				"The group cracked and released demos exclusively for the Commodore&nbsp;64 and Amiga platforms before expanding to consoles and the <a href=\"/f/b04615\">PC in February 1991</a>.",
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
			Content: "The Video Graphics Array (<strong>VGA</strong>) graphics standard is released. It is the first graphics standard to support <strong>256 colors</strong> and resolutions up to 640x480.",
		},
		{
			Title: "The earliest PC demo", Year: 1987, Month: 6, Day: 22, Highlight: true,
			Lead:      "3 Dimensional EGA Demonstration",
			LinkTitle: "and view the demo", Link: "/f/ac21460",
			Content: "A demo is a piece of software created purely for aesthetics, usually to show art and animation. " +
				"While earlier demonstration software existed on the PC, they were intended for retailers or distributors and usually not given to the public.",
		},
		{
			Title: "Music audio standard", Year: 1987,
			Lead: "AdLib Music Synthesizer Card", LinkTitle: "about the AdLib sound card",
			Link: "https://www.computinghistory.org.uk/det/23724/AdLib-Music-Synthesizer-Card/",
			Content: "The Music Synthesizer Card sound card is released. It was the first sound card to use FM synthesis and the first widely adopted by game developers. " +
				"<strong>AdLib</strong>'s success was short-lived, as competitor <a href=\"https://www.creative.com\">Creative Labs</a> released the <a href=\"https://www.vgmpf.com/Wiki/index.php?title=Sound_Blaster\">Sound Blaster</a> in 1989, " +
				"a clone of the AdLib card that included a simple digital sound processor for speech and sound effects.",
			Picture: Picture{
				Title:       "An AdLib Music Synthesizer ISA slot card",
				Jpg:         "adlib-card.jpg",
				Attribution: "TheAlmightyGuru",
				License:     "GNU FDL",
				LicenseLink: "https://www.vgmpf.com/Wiki/index.php?title=File:AdLib_-_1987.jpg",
			},
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
			Content: "<p><a href=\"https://www.mobygames.com/game/4019/rockford-the-arcade-game/\">Rockford</a> is a strange game, being a port of the arcade game of the same name—a machine, created as a port of the then-popular microcomputer videogame series, " +
				"<a href=\"https://boulder-dash.com/history/\">Boulder Dash</a>.</p>" +
				"<p>More unusual is the use of <strong>32 color VGA</strong> for a home computer port of an arcade game on the PC, in an era when ports were done on the cheap using lowest common denominator 4 color CGA graphics. " +
				"The crossover of players who owned expensive VGA graphic cards and monitors in 1988 who were playing arcade ports was probably low.</p>",
			Picture: Picture{
				Title:       "Rockford: The Arcade Game",
				Alt:         "Rockford: The Arcade Game screenshot",
				Png:         "rockford-32-color-vga.png",
				Attribution: "486pc & MobyGames",
				License:     "© Arcadia",
				LicenseLink: "https://www.mobygames.com/game/4019/rockford-the-arcade-game",
			},
		},
		{
			Title: "Earliest, standalone elite BBS advert", Year: 1988, Month: 4, Day: 4, Highlight: false,
			Lead: "Swashbucklers II", LinkTitle: "the file",
			Link: "/f/b844ef",
			Content: "<p>While novel in 1988, <strong>BBS adverts</strong> like these would plague releases as spam in the years to come, with boards injecting numerous texts and tagging the releases with their names, often under the guise of documentation or readme texts.</p>" +
				"<pre>Another Quality Ware Downloaded off:<br>" +
				"S W A S H B U C K L E R S   I I<br>" + "Home of PTL/CPI<br>" +
				"100 megs Online!<br>" +
				"85 megs Offline, Request!<br>" +
				"All PTL/CPI Cracks FREE<br>" +
				"All other Major Groups cracks Always Online<br>" +
				"Ask your local Sysop for the number..</pre>",
			Picture: Picture{
				Title: "Swashbucklers II",
				Alt:   "Swashbucklers II text advert screenshot",
				Webp:  "b844ef.webp",
				Png:   "b844ef.png",
			},
		},
		{
			Title: "Earliest NFO-like text document", Year: 1988, Month: 7, Day: 30, Highlight: false,
			Lead: "Bentley Sidwell Productions", LinkTitle: "the file", Link: "/f/ad417f",
			Content: "<p><strong>NFO</strong> information text files are usually distributed with pirated software to provide usage instructions, promote the release group, and occasionally encourage group propaganda.</p>" +
				"<p>Bentley Sidwell Productions may have released the earliest NFO-like document for the 1988 game <a href=\"https://www.mobygames.com/game/9093/romance-of-the-three-kingdoms/cover/group-9976/cover-249195/\">Romance of The Three Kingdoms</a>.</p>" +
				"<pre>" +
				"************************************************************************" + br +
				br +
				"  Romance of The Three Kingdoms" + br +
				"    - (KOEI) -" + br +
				"  \"We Supply The Past, You Make The History\"" + br +
				br +
				"***********************************************************************" + br +
				br +
				"Welcome to the wonderful world of second century China.. China's" + br +
				"second dynasty has collapsed and the entire nation battles itself for" + br +
				"supremacy in this most interesting action game from Koei.." + br +
				br +
				"Floppy users: UnARC - ROTK-1.ARC onto Disk #1..." + br +
				"                      ROTK-2.ARC onto Disk #2..." + br +
				"                      ROTK-3.ARC onto Disk #3..." + br +
				br +
				"Hard Drive:   UnARC - All Files Into One Directory..." + br +
				br +
				"Nothing to edit... nothing.</pre>",
		},
		{
			Title: "The earliest ASCII art on PC", Year: 1988, Month: 10, Day: 6, Highlight: true,
			Lead: "Another quality ware from $print", LinkTitle: "and view the file", Link: "/f/ab3dc1",
			Content: "The earliest <strong>ASCII art</strong> known so-far is released by $print for the game " +
				"Fire Power. The ASCII text logo is relatively crude and is not as detailed as later ASCII art. " +
				"<pre> ╔═══════════════════════════════╗<br>" +
				"╔╝      Another Quality Ware     ╚╗<br>" +
				"║          F  R  O  M             ║<br>" +
				"║                                 ║<br>" +
				"║   ┌┼┼┼ ┌─┐┌──┐ ─┬─ │\\  │──┬──   ║<br>" +
				"║   └┼┼┼┐┼─┘│─┬┘  │  │ \\ │  │     ║<br>" +
				"║   ─┼┼┼┘│  │ └─ ─┴─ │  \\│  │     ║<br>" +
				"╚═════════════════════════════════╝<br>" +
				"║  The Ultimate Empire [USA]      ║<br>" +
				"║  Warez R Us          [CAN]      ║<br>" +
				"╚═════════════════════════════════╝</pre>",
			Picture: Picture{
				Title: "Another quality ware from $print",
				Alt:   "Fire Power by $print ASCII screenshot",
				Webp:  "ab3dc1.webp",
				Png:   "ab3dc1.png",
			},
		},
		{
			Title: "The earliest PC Scene drama", Year: 1988, Month: 11, Day: 25,
			Lead: "TNWC accusing PTL of stealing a release", LinkTitle: "and view the file",
			Link: "/f/aa356d",
			Content: "<p>The earliest <strong>scene drama</strong> known so far involves a release by <a href=\"/g/the-north-west-connection\">The North West Connection</a> (TNWC) for the game Paladin. " +
				"The drama accuses <a href=\"/g/ptl-club?\">PTL Club</a> of stealing and \"re-releasing\" an early game released by TNWC. " +
				"Scene drama often involves texts that call out other groups for poor behavior, breaking commonly accepted rules, or being \"lame\".</p>" +
				"<p><q>DO NOT TAKE THIS FILE FROM THE ARCHIVE!!!!<br>" +
				"Well unlike PTL I won't sacrifice some game code to put up a fancy title screen for the group that released this (TNWC). " +
				"This is officially out third release, but really it's our second major one since PTL took Paladin and \"re-released\" it by taking off the doc check.<br>" +
				"Anyway - on with the game.  This game is a great role-playing game with some of the best graphics I've seen in an RPG (which is not what you'd expect from Infocom) so enjoy it.</q></p>",
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
			Link:    "https://www.mobygames.com/game/2099/688-attack-sub",
			Content: "Driving, flying, and military simulation games were once a popular genre of video games on the PC. Before dedicated <abbr title=\"graphics processing unit\">GPUs</abbr> existed, this genre created demanding open-world landscapes requiring expensive CPUs and even co-processors! Which was great for those with high-end hardware who wanted to show off.",
			Picture: Picture{
				Title:       "688 Attack Sub",
				Alt:         "688 Attack Sub in-game screenshot",
				Png:         "688-attack-sub.png",
				Attribution: "Defacto2",
				License:     "© Electronic Arts",
				LicenseLink: "https://www.mobygames.com/game/2099/688-attack-sub/screenshots/dos/9155/",
			},
		},
		{
			Title: "Earliest ANSI loader", Year: 1989, Month: 3,
			Lead: "The Rogues Gallery", LinkTitle: "the file",
			Link: "/f/ad21da8",
			Content: "<p><strong>ANSI loaders</strong> were text files with ASCII escape control characters to provide color and cursor movement. However, a specific display driver on IBM and other PCs often needed to load at boot before viewing the texts. So, to avoid this, Sceners converted their ANSI artworks into simple, self-displaying applications or \"loaders.\"</p>" +
				"<p><a href=\"https://demozoo.org/bbs/1762/\">The Rogues Gallery</a> (+516-361-9846) was a BBS based in Long Island, New York.</p>",
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
			Content: "An <strong>intro</strong> or the later cractrkro are small, usually short, demonstration programs designed to display text with are or animations. " +
				"Oddly, this first intro was created by a group of teenagers out of <strong>Findland</strong>, a country not known for its use of the expensive PC platform. " +
				"<br>Other more popular 16-bit platforms such as the Commodore Amiga and Atari&nbsp;ST offered much better graphics and audio than CGA on the PC.",

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
			Content: "This first <strong>cracktro</strong> is released by Future Brain Inc for the game <a href=\"https://www.mobygames.com/game/2161/lombard-rac-rally/cover/group-99392/cover-270796/\">Lombard RAC Rally</a>. " +
				"Future Brain Inc were a group from the <strong>Netherlands</strong>, and were one of the first groups to release cracktros on the PC platform.<br>" +
				"Early cracktros on the PC platform lacked music, and were usually a simple screen of text and a logo. " +
				"On other platforms, the Commodore&nbsp;64, Amiga 500, and Atari&nbsp;ST, cracktros offered music and graphic effects which were easier to create due to their unified hardware. ",
			Picture: Picture{
				Title: "Lombard RAC Rally cracktro",
				Alt:   "Lombard RAC Rally cracktro screenshot",
				Webp:  "b83fd7.webp",
				Png:   "b83fd7.png",
			},
		},
		{
			Title: "First issue of Pirate magazine", Year: 1989, Month: 6, Day: 1,
			Lead: "The earlist known scene newsletter for the Scene on the PC", LinkTitle: "the issues",
			Link: "/g/pirate",
			Content: "<p>Created in Chicago, Pirate magazine was a bi-monthly text newsletter for the Scene on the PC platform and distributed through bulletin boards. It ran for at least five issues between June 1989 and April 1990.</p>" +
				"<q>What's a pirate? COMPUTER PIRACY is copying and distribution of copyright software (warez). Pirates are hobbyists who enjoy collecting and playing with the latest programs. " +
				"Most pirates enjoy collecting warez, getting them running, and then generally archive them, or store them away. A PIRATE IS NOT A BOOTLEGGER. " +
				"Bootleggers are to piracy what a chop-shop is to a home auto mechanic. Bootleggers are people who DEAL stolen merchandise for personal gain. " +
				"Bootleggers are crooks. They sell stolen goods. Pirates are not crooks, and most pirates consider bootleggers to be lower life forms...</q>" +
				"<br><q>Pirates SHARE warez to learn, trade information, and have fun! But, being a pirate is more than swapping warez. It's a life style and a passion.</q>",
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
			Title: "Origin of the \".NFO\" file extension", Year: 1990, Month: 1, Day: 23, Highlight: true,
			Lead: "The Humble Guys, Knights of Legend or Bubble Booble", LinkTitle: "the file",
			Link: "/f/ab3945",
			Content: "<code>KNIGHTS.NFO</code> <small>or</small> <code>BUBBLE.NFO</code><br>" +
				"<p>The <code>.NFO</code> file extension is used to denote a text file containing information about a release. " +
				"Still in use today, the dot nfo file contains information about the release group, the release itself, and how to install.</p>" +
				"<p>While disputed, it is not too important which release from <strong>The Humble Guys</strong> is the first to use the dot \"nfo\" file extension. " +
				"The timestamps of the release files suggest the text file for Knights of Legend predates Bubble Bobble by a few days. But famed, former cracker Fabulous Furlough has always stated Bubble Bobble was the release that first used the naming standard.</p>" +
				"<p>Bubble Bobble was the far more notable game of the era and may have been the more memorable game title when recalling the event. " +
				"Meanwhile, Knights didn't require a crack, and the filenaming convention was possibly retroactively applied.</p>" +
				"<figure><blockquote class=\"blockquote\"><q><small>It happened like this, I'd just used <a href=\"http://nerdlypleasures.blogspot.com/2011/05/scourge-of-preservation-disk-based-copy.html\">\"Unguard\"</a> to crack the SuperLock off of <a href=\"/f/ad4195\">Bubble Bobble</a>, and I said \"I need some file to put the info about the crack in. Hmmm.. Info, NFO!\", and that was it.</small></q></blockquote>" +
				"<figcaption class=\"blockquote-footer\">Famed, former cracker for The Humble Guys, Fabulous Furlough recalls Bubble Bobble as the first THG release that used the .NFO file extension.</figcaption></figure>" +
				"<p><q>Bubble Bobble by Nova Logic Through Taito<br>Broken by Fabulous Furlough<br>Normal Taito Loader - 5 minutes</q></p>" +
				"<p><q>Knights of Legend by Origin Systems<br>It seems to be unprotected, if you find anything leave us a message..</q></p>",
		},
		{
			Title: "Earliest PC cracktro with music", Year: 1990, Month: 12, Day: 2,
			Lead: "The Cat, M1 Tank Plattoon", LinkTitle: "about and view cractrko",
			Link: "/f/ab25f0e",
			Content: "<p>This <strong>cracktro</strong> was released by The Cat for the game <a class=\"text-nowrap\" href=\"https://www.mobygames.com/game/1499/m1-tank-platoon/cover/group-3004/cover-230986/\">M1 Tank Platoon</a>. " +
				"It is the first known cracktro on the PC platform to feature music. " +
				"But <em>music</em> in a loose sense, as it relies on the terrible internal PC speaker to produce the melody.</p>" +
				"<p>While the 8-bit consoles and some microcomputers offered dedicated music audio chips, most famously the Commodore&nbsp;64 with its SID chip, " +
				"the IBM&nbsp;PC which targeted business did not.</p>",
			Picture: Picture{
				Title: "Tank Platoon cracktro",
				Alt:   "Tank Platoon cracktro screenshot",
				Webp:  "ab25f0e.webp",
				Png:   "ab25f0e.png",
			},
		},
		{
			Title: "Digital audio standard", Year: 1990,
			Lead:      "SoundBlaster",
			LinkTitle: "The Sound Blaster Story", Link: "https://www.custompc.com/retro-tech/the-sound-blaster-story",
			Content: "<p>The <strong>Sound Blaster</strong> audio standard came about in 1990 after the Sound Blaster 1.5 audio card was released by Creative Labs, with the box proudly proclaiming" +
				" it <q><a href=\"https://vgmpf.com/Wiki/index.php?title=File:Sound_Blaster_1.5_-_Box_-_Back.jpg\">The PC Sound Standard</a></q>. " +
				"It was the first digital audio standard for the IBM&nbsp;PC to be widely adopted on the PC platform, despite its poor quality, mono 8-bit digital audio." +
				"Previous audio standards such as the AdLib and the MT-32, were limited to FM synthesis or MIDI-like samples.</p>" +
				"<p>The Sound Blaster was the first audio standard widely adopted by the PC platform and was the de facto audio option in games for many years.</p>",
		},
		{
			Title: "CD-ROM multimedia", Year: 1990, Prefix: "Winter",
			Lead: "Mixed-Up Mother Goose", LinkTitle: "the Catalog listing the game",
			Link: "https://archive.org/details/vgmuseum_sierra_sierra-90catalog-alt3/page/n21",
			Content: "<p>The first widely available enhanced PC game on <strong>CD-ROM</strong> was <a href=\"https://www.mocagh.org/sierra/mothergoose-alt-manual.pdf\">Mixed-Up Mother Goose</a>, announced by Sierra On-Line in 1990 and released in 1991. " +
				"The children's game was a high-technology remake of <a href=\"https://www.mobygames.com/game/758/mixed-up-mother-goose/cover/group-27001/cover-70129/\">a fun title from 1987</a>, but the CD-ROM remake featured new, enhanced VGA graphics and interface, digital audio with speech, singing, and music.</p>" +
				"<p>With the newest of the technology and a lack of standards for CD media, <a href=\"https://sierrachest.com/index.php?a=games&id=544&title=mother-goose-vga&fld=box&pid=3\">the box came with two identical discs</a>, one red and one blue. The red disc supported Red Book CD audio, while the blue disc lower quality digital samples playback.</p>",
		},
		{
			Year: 1990, Prefix: notable,
			List: Links{
				{LinkTitle: "ANSI Creators in Demand", Link: "/g/acid-productions", SubTitle: "ACiD", Forward: "Aces of ANSI Art"},
				{LinkTitle: "Bitchin ANSI Design", Link: "/g/bitchin-ansi-design", SubTitle: "BAD"},
				{LinkTitle: "Damn Excellent ANSI Design", Link: "/g/damn-excellent-ansi-design", SubTitle: "Damn"},
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
			Content: "<p><code>XTC-AD.COM</code></p>" +
				"<p>This <strong>VGA loader</strong> is an advert for the well-known bulletin board <a href=\"https://demozoo.org/bbs/4009/\">XTC Systems</a> in Dallas, Texas. " +
				"It served as the <em>World Headquarters</em> for the famed art group <a href=\"/g/acid-productions\">ACiD Productions</a> and as a distribution board for <a href=\"/g/fairlight\">Fairlight</a>, <a href=\"/g/razor-1911\">Razor 1911</a>, and some popular magazines.",
			Picture: Picture{
				Title: "XTC Systems BBS VGA loader",
				Alt:   "XTC Systems BBS VGA loader screenshot",
				Webp:  "a41dcd9.webp",
				Png:   "a41dcd9.png",
			},
		},
		{
			Title: "The contemporary PC cracktro", Year: 1991, Month: 3, Day: 12, Highlight: true,
			Lead: "The Dream Team Presents Blues Brothers", LinkTitle: "about and view the cracktro", Link: "/f/b249b1",
			Content: "A collaboration of <a href=\"/g/the-dream-team\">The Dream Team</a> with <a href=\"/g/tristar-ampersand-red-sector-inc\">Tristar &amp; Red Sector Inc.</a> released this 1991 <strong>cracktro</strong>. " +
				"TDT founder <a href=\"/p/hard-core\">Hard Core</a> programmed it, which is the first known cracktro on the PC platform to feature a modern presentation with a logo, music, and a scroller. " +
				"Cracktros on the PC platform had previously been limited to primarily static logo screens or, in the case of the earliest cracktros, no graphics.",
			Picture: Picture{
				Title: "Blues Brothers cracktro",
				Alt:   "Blues Brothers cracktro screenshot",
				Webp:  "b249b1.webp",
				Png:   "b249b1.png",
			},
		},
		{
			Title: "The contemporary PC Demoscene", Year: 1991, Month: 7,
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
			Content: "<p><a href=\"https://demozoo.org/bbs/7179/\">Splatterhouse, or Splatter House</a>, was a San Jose, California bulletin board " +
				"heavily affiliated with the <a href=\"/g/international-network-of-crackers\">International Network of Crackers</a>, the art group <a href=\"/g/acid-productions\">ACiD Productions</a>, " +
				"and the designers of this <strong>BBStro</strong>, <a href=\"/g/insane-creators-enterprise\">Insane Creators Enterprise</a>.",
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
			Title: "First SuperVGA / VESA game", Year: 1992, Month: 6,
			Lead: "Links 386 Pro", LinkTitle: "the mobygames page", Link: "https://www.mobygames.com/game/3757/links-386-pro/",
			Content: "<p>The first widely available <strong>SuperVGA</strong> game was Links 386 Pro from Access. Here, another popular golf simulation pushed the baseline PC gaming requirements with the need for higher-end hardware. " +
				"The 386 in the title stated the minimum requirement of an Intel 386 CPU when 286 systems were the commodity.</p>" +
				"<p>The problem for consumers is that ordinarily, most PC software never took advantage of the enhancements offered by the more expensive Intel 386 or 486 CPUs.</p>" +
				"<p>Some caveats to the first SVGA/VESA claim: we are talking about a retail, boxed game requiring a resolution/color depth that a standard VGA setup cannot handle, " +
				"so at least a constant 600x400 resolution with 256 colors.</p>",
			Picture: Picture{
				Title:       "Links 386 Pro",
				Alt:         "Links 386 Pro in-game screenshot",
				Png:         "links-386-pro-svga.png",
				Attribution: "Servo & MobyGames",
				License:     "© Access Software",
				LicenseLink: "https://www.mobygames.com/game/3757/links-386-pro/",
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
			Picture: Picture{
				Title: "The One and Only",
				Webp:  "b13a93.webp",
				Png:   "b13a93.png",
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
			Content: "<p>The earliest known release was a <strong>CD image</strong> of the game <a href=\"https://www.mobygames.com/game/3350/hurl/cover/group-2469/cover-13273/\">Slob Zone</a> later known as H.U.R.L.</p>" +
				"CD images were generally disliked in the Scene, as they had no copy protection to crack and occupied too much space on the file site or bulletin board's costly hard drives. " +
				"<p>ROM 1911 was used by Razor 1911 as a dumping ground for CD titles.</p>" +
				"<q>ROM 1911 is dedicated to bringing you the best CD titles available on PC, with a catch : we must be able to zip these games down to 10 1.44 meg files or less, or we won't release it. " +
				"We know that a large portion of the scene would like to play these games, but the size of many CD titles is prohibitive to the pirate gamer. But not all games are too big to release! Which brings us here ..</q>",
		},
		{
			Title: "Copyright infringement legal precedent", Year: 1994, Month: 12, Day: 28, Highlight: true,
			Lead: "No criminal liability for the sharing of software", LinkTitle: "David LaMacchia Defense Fund with press releases", Link: "https://web.archive.org/web/19990224000548/http://photo.net/dldf/home.html",
			Content: "In April 1994, David LaMacchia, a 20 year old junior at the Massachusetts Institute of Technology, <a href=\"/f/b628640\">is indicted</a> for conspiring to commit wire fraud. A 1950s law intended to stop defrauding another out of money using the U.S. landline telephone network. " +
				"<br>David ran two anonymous <a href=\"https://fsp.sourceforge.net/index.html\">File Service Protocol</a> sites using MIT's internal network connected to the Internet to share software with users without financial gain. The primary site, <a href=\"https://web.archive.org/web/19991018194139/http://photo.net/dldf/indictment.html\">Cynosure, offered downloads, while Cynosure II</a> also permitted uploads with requests." +
				"<p></p>Months later, David's defense lawyers filed a motion to dismiss, " +
				"<q>LaMacchia contends that the indictment invents a criminal charge, primarily by distorting the wire fraud statute, in order to circumvent Congress's decision not to apply a criminal sanction to LaMacchia's alleged conduct.</q>" +
				"<br>And just days after Christmas, the motion to dismiss was allowed by District Judge Stearns." +
				"<br><q>The Court dismissed the indictment, holding that <strong>there was no clearly expressed Congressional intent to permit prosecution of copyright " +
				"infringement</strong> under the wire fraud statute. There was no allegation that LaMacchia infringed copyrighted software for commercial advantage or private financial gain.</q>",
		},
		{
			Year: 1994, Prefix: notable,
			List: Links{
				{LinkTitle: "ROM 1911", Link: "/g/rom-1911", SubTitle: "ROM", Forward: "Razor 1911"},
				{LinkTitle: "Request to Send", Link: "/g/request-to-send", SubTitle: "RTS"},
				{LinkTitle: "Genesis", Link: "/g/genesis", SubTitle: "GNS", Forward: "Pentagram"},
				{LinkTitle: "TDU-Jam", Link: "/g/tdu_jam", SubTitle: "TDU", Forward: "Genesis"},
			},
			Picture: Picture{
				Title: "Genesis brand",
				Webp:  "a228ca.webp",
				Png:   "a228ca.png",
			},
		},
		{
			Title: "Earliest CD-RIP release", Year: 1995, Month: 6, Day: 3, Highlight: true,
			Lead: "Hybrid", LinkTitle: "about the release", Link: "/f/a938e5",
			Content: "<p>A play on the media, CD-ROM, the earliest known <strong>CD-RIP</strong> (later simplified to \"rip\") release, was by Hybrid for the game <a href=\"https://www.mobygames.com/game/3328/virtual-pool/cover/group-119259/cover-316591/\">Virtual Pool from Interplay</a>. " +
				"Hybrid was a group formed by ex-members of <a href=\"/g/pyradical\">Pyradical</a> and <a href=\"/g/pentagram\">Pentagram</a>.</p>" +
				"The <u>CD RIP</u> type came about due to CD-ROM-only games being unable to get a proper Scene release. For PC game publishers, " +
				"CD-ROMs were cheaper to produce and had far more storage capacity than the standard floppy disks. However, large hard drives were too expensive to store the content of complete CD images. " +
				"So, for many pirates to play a game published on CD, the disc's content had to be ripped and repackaged to a hard drive, but with the removal of the game's fluff, such as intro videos, music, and speech.",
		},
		{
			Title: "Windows 95 warez release", Year: 1995, Month: 8, Day: 12,
			Lead: "Drink or Die", Link: "/f/bb2b71f", LinkTitle: "about the release",
			Content: "<p><strong>Drink or Die</strong> became notorious for releasing the CD media for the box retail edition of <strong>Windows&nbsp;95</strong> two weeks before the official worldwide release.</p>" +
				"<p>In an era when global, same-day product launches were logistically costly and uncommon, this operating system launch was probably the most hyped Microsoft consumer product ever. " +
				"Over a decade before Apple cemented the marketing tactic, <a href=\"https://rarehistoricalphotos.com/windows-95-launch-day-1995/\">Windows 95 had fans queuing at midnight in retail stores worldwide</a>.</p>" +
				"<p>The release also highlighted a significant problem for software and game publishers: for pirates to get access to the retail packaging weeks before launch meant some company employees were either members of these warez groups or receiving kickbacks.</p>" +
				"<p>Years later, competitor <a href=\"/g/pirates-with-attitudes\">Pirates With Attitudes</a> would release the <a href=\"/f/a52a8c\" class=\"text-nowrap\">Windows 98 media</a> five weeks and <a href=\"/f/b42e2f6\">Windows 2000</a> two months before the official launches! " +
				"But in the following decade, a global, coordinated law enforcement effort would take down both groups.</p>" +
				"<p>The other Microsoft-sourced releases from DOD during this two-week period were the Windows 95 <a href=\"/f/b82406f\" class=\"text-nowrap\">floppy edition</a>, <a href=\"/f/b721b5\" class=\"text-nowrap\">upgrade edition</a>, <a href=\"/f/b92697\" class=\"text-nowrap\">Plus Pack</a>, Microsoft <a href=\"/f/ba28e0f\">BOB</a>, and <a href=\"/f/bc2dc2f\">Word</a>.</p>",
		},
		{
			Title: "Windows 95", Year: 1995, Month: 8, Day: 24,
			Lead: "Worldwide retail release", LinkTitle: "about the day in history",
			Link: "https://www.theverge.com/21398999/windows-95-anniversary-release-date-history",
			Content: "<p>Microsoft's biggest and most hyped mainstream product release was hugely successful in the market and finally began the PC's transition away from the archaic IBM and Microsoft DOS <small>(Disk&nbsp;Operating&nbsp;System)</small>.</p>" +
				"<p>Windows 95 had been a long time coming, over a decade late, and offered a fully graphical user interface as the default. It also introduced <a href=\"https://arstechnica.com/gadgets/2015/08/the-windows-start-menu-saga-from-1993-to-today/\">the famed Start menu concept</a> that would later become favored by many Windows and, ironically, <a href=\"https://fossforce.com/2019/07/why-gnome-2-continues-to-win-the-desktop-popularity-contest/\">desktop Linux users</a>.</p>",
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
			Picture: Picture{
				Title: "Hoodlum cracktro",
				Webp:  "a8450c.webp",
				Png:   "a8450c.png",
			},
		},
		{
			Title: "The Scene merch", Year: 1996, Month: 1,
			Lead: "Razor 1911 Tenth Anniversary CD-ROM", LinkTitle: "the order form", Link: "/f/a42df1",
			Content: "<p>The first major <strong>Scene merchandise</strong> was selling a CD-ROM by <a href=\"/g/razor-1911\">Razor 1911</a> to celebrate their 10th anniversary. " +
				"The disc was a collection of their PC releases from 1991 to 1995 and, including worldwide postage, sold for $40 each, or about the cost of a full-priced, boxed PC game. " +
				"Before online or consumer digital transactions, buyers had to post in an envelope the physical cash and an order form to a PO Box in Florida.</p>" +
				"<p>The disc was controversial as <strong>reselling</strong> scene-released software was criminal and frowned upon. But its success meant other group merchandise soon followed suit, with the most popular items being branded <a href=\"/f/b4484e\">t-shirts</a>.</p>",
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
			Content: "<p>The Standards of Piracy Association (<strong>SPA</strong>) was formed by the groups " +
				"<a href=\"/g/prestige\">Prestige</a>, " +
				"<a href=\"/g/razor-1911\">Razor 1911</a>, " +
				"<a href=\"/g/mantis\">Mantis</a>, " +
				"<a href=\"/g/napalm\">Napalm</a>, " +
				"and <a href=\"/g/hybrid\">Hybrid</a>.</p>" +
				"<p>For the prior 15 years, PC publishers used 5¼ and 3½ inch floppy disks to distribute software, whereas the CD-ROM was now the standard medium for boxed retail games. " +
				"But CD-ROMs were too large for the Scene to copy, crack, and spread properly. After several confusing and broken releases, an association of groups created a set of standards for releasing <strong>CD-RIP</strong>s. " +
				"While floppy disk distributed releases always included the complete and cracked game, ripped CD releases were playable but missing key gameplay features, such as cutscenes, music, instruction manuals, and speech.</p>" +
				"<p><em>CD ripping made an incomplete but technically playable game accepted as a valid pirated release, as this was not the case prior.</em></p>" +
				"<p><q>the SPA is an agreement between the 5 top PC games groups that lays down the official \"rules of engagement\" to be used in the battle to release the most</q></p>",
			List: Links{
				{LinkTitle: "The Faction", Link: "/f/a634e1", SubTitle: "1998"},
				{LinkTitle: "NSA", Link: "/f/a13771", SubTitle: "2000"},
			},
		},
		{
			Title: "The first popular 3D graphics chipset", Year: 1996, Month: 10,
			Lead: "3Dfx Voodoo 1", LinkTitle: "The Voodoo That They Righteously Do", Link: "https://computeme.tripod.com/voodoo1.html",
			Content: "<p>The <a href=\"https://www.pixelrefresh.com/3dfx-orchid-righteous-voodoo-1-where-3d-acceleration-truly-began\">Orchid Righteous</a> is available in retail. " +
				"Later cards from other manufacturers, such as the <a href=\"https://www.tomshardware.com/reviews/3d-accelerator-card-reviews,42-7.html\">Diamond Monster 3D</a>, quickly followed, and within a year, the 3Dfx chipset dominated the market.</p>" +
				"<p>Before the 3Dfx Voodoo release, the fragmented 3D graphics market for consumer PCs had abysmal software support. " +
				"3Dfx coordinated with publishers to target their Glide API with new game releases so gamers had confidence in their Voodoo card purchases.</p>" +
				"<p>3Dfx also extended the life of existing PC hardware and broke the endless cycle of aggressive, expensive CPU upgrades to support the current generation of games. " +
				"A new 3Dfx card would double the resolution, add fantastic color support, and even improve the frames-per-second on what would otherwise be an older machine.</p>",
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
			Content: "<p>The Standards of Piracy Association <a href=\"/f/aa3b26\">CD-RIP standards</a> were <strong>broken by founding member</strong> " +
				"Hybrid with this release of <a href=\"https://news.blizzard.com/en-us/diablo3/22887361/diablo-now-available-on-gog-com\">Diablo</a>.</p>" +
				"Less than a year prior, <strong>SPA</strong> had agreed that <strong>CD-RIP</strong>s should be ripped to a maximum permitted size and any titles where this wasn't possible should be skipped. " +
				"It wasn't uncommon for major games such as <a href=\"https://www.imdb.com/title/tt0131537/\">Sierra's Phantasmagoria</a> " +
				"to be passed over by release groups due to their massive size and game play reliance on unrippable video and audio content.",
			List: Links{
				{LinkTitle: "Diablo from Razor 1911", Link: "/f/a72ced", SubTitle: "full CD rip"},
			},
		},
		{
			Title: "Earliest ISO release", Year: 1997, Month: 11, Day: 27, Highlight: true,
			Lead: "CD Images For the Elite", LinkTitle: "the release", Link: "/f/ad40ce",
			Content: "An <a href=\"https://www.loc.gov/preservation/digital/formats/fdd/fdd000348.shtml\">ISO is a common file archive format</a> containing the entire CD and later DVD data. It enables the copying and exact replication of data onto consumable blank discs." +
				"<p>Trading <strong>ISO images</strong> between individuals has happened for years prior, but <a href=\"https://www.mobygames.com/game/2082/lords-of-magic/covers/\">Lords of Magic</a> was the earliest known ISO release pushed to the Scene.</p>" +
				"<p>A formalization of an ISO trading scene occurred sometime in late 1997, but it took years before the medium became the dominant format.</p>",
		},
		{
			Title: "Overnight, many Warez Scene activities become criminal", Year: 1997, Month: 12, Day: 16, Highlight: true,
			Lead: "No Electronic Theft (NET) Act", LinkTitle: "Forbes article Stealing code", Link: "https://www.forbes.com/1997/11/28/feat.html?sh=5fc911fb2c1c",
			Content: "<p><q><strong>The NET Act</strong> was signed into law by President Clinton in December 1997, <u>making it illegal to reproduce or distribute copyrighted works</u>, such as software programs and musical recordings, even if the defendant acts without a commercial purpose or for private financial gain.</q></p>" +
				"<p>The law is a response to the failed prosecution against David LaMacchia from a few years earlier. The dismissal of his case brought attention to the legal anomaly named after his win, <a href=\"https://www.newscientist.com/article/mg15621113-000-publish-on-the-net-and-be-damned/\">the LaMacchia loophole</a>.</p>" +
				"<p>Under the new law, <q>if the defendant reproduces or distributes 10 or more copyrighted works that have a total value of more than $2,500, he or she can be charged with a felony, and faces a sentence of up to 3 years imprisonment and a fine of up to $250,000.</p>" +
				"<p>A defendant who reproduces or distributes one or more copies of copyrighted works with a value of more than $1,000 can be charged with a misdemeanor, and face up to one year in prison and a fine of up to $100,000.</q></p>",
		},
		{
			Year: 1997, Prefix: notable,
			List: Links{
				{LinkTitle: "CD Images For the Elite", Link: "/g/cd-images-for-the-elite", SubTitle: "CiFE"},
				{LinkTitle: "Divine", Link: "/g/divine", SubTitle: "DVN"},
			},
			Picture: Picture{
				Title: "Divine cracktro",
				Webp:  "a424a4c.webp",
				Png:   "a424a4c.png",
			},
		},
		{
			Year: 1998, Month: 3, Day: 31,
			Title: "Online keys",
			Lead:  "StarCraft by Blizzard",
			Content: "<a href=\"https://www.mobygames.com/game/378/starcraft/cover/group-9232/cover-2059/\">StarCraft</a> was a hugely hyped and popular real-time strategy game by Blizzard Entertainment.<br>" +
				"A significant gameplay component was its online multiplayer mode through Blizzard's&nbsp;<a href=\"https://www.myabandonware.com/game/starcraft-epy\">Battle.net</a>. " +
				"The service required player registration and <strong>a unique unlock code</strong> in each copy of the game, making StarCraft the first retail game to issue CD keys.",
			Picture: Picture{
				Title:       "Rear of the StarCraft CD case",
				Alt:         "Rear of the StarCraft CD case screenshot",
				Attribution: "MES392",
				License:     "©",
				LicenseLink: "https://www.reddit.com/r/starcraft/comments/aaz4es/cleaned_up_the_office_who_needs_an_original/",
				// Webp:  "starcraft-case.webp",
				Jpg: "starcraft-case.jpg",
			},
		},
		{
			Year: 1998, Month: 4, Day: 1,
			Title: "Starcraft", LinkTitle: "the release", Link: "/f/a9353d",
			Lead: "Razor 1911",
			Content: "The credited release of StarCraft went to <a href=\"/g/razor-1911\">Razor 1911</a> and famed cracker <a href=\"/p/beowulf\">Beowulf</a>, who together put out a CD-RIP of the game. " +
				"However, the package took time to compile and lacked the unique CD keys required to play the desirable online multiplayer." +
				"<p><q>Well, what can I say. This has got to be one of the hardest titles I have ever ripped. " +
				"The crack was trivial, but ripping this game involved understanding and coding utilities for Blizzard's file packer. It is ...a veritable nightmare.</q></p>",
			List: Links{
				{LinkTitle: "StarCraft Battle.NET Keymaker", Link: "/f/b321b00", SubTitle: "2 April"},
				{LinkTitle: `Starcraft *100% FIX*`, Link: "/f/b13d2c", SubTitle: "3 April"},
			},
			Picture: Picture{
				Title: "Razor 1911 Starcraft Broodwar cracktro",
				Webp:  "b22b15d.webp",
				Png:   "b22b15d.png",
			},
		},
		{
			Year:  1998,
			Title: "ISO scene picks up steam",
			Content: "<p>The <strong>ISO scene</strong> is still in its infancy but snowballs after some top groups start releasing with the file format.</p>" +
				ul0 +
				"<li>Razor 1911 merges the separate <a href=\"/f/a82c49\">ISO division</a> back into the Razor 1911 label.</li>" +
				"<li><a href=\"/f/ac2be5\">Fairlight returns</a> after 4-years and is exclusively releasing with the format.</li>" +
				"<li>The famed couriers RiSC created <a href=\"/f/b04dac\">RiSCiSO</a> to become one of the largest ISO release groups.</li>" +
				"<li><a href=\"/f/ae48b0\">PDM ISO</a> becomes the ISO division of <a href=\"/g/paradigm\">Paradigm</a> and Zeus.</li>" +
				"<li><a href=\"/g/dvniso\">DVNiSO</a> becomes an ISO division of Divine and Deviance.</li>" +
				ul1 +
				"<p>Other early users of the format include<br><a class=\"text-nowrap\" href=\"/g/cd-images-for-the-elite\">CD Images for the Elite</a> (CiFE), <a href=\"/g/kalisto\">Kalisto</a>, <a href=\"/g/isolation\">ISOlation</a>, <a class=\"text-nowrap\" href=\"/g/in-search-of-cd\">In Search of CD</a>, and CaLiSO.</p>" +
				"<p><q>Paradigm - we do rips, we do ISO - we do it all with style</q></p>",
		},
		{
			Year: 1998, Prefix: notable,
			List: Links{
				{LinkTitle: "Fairlight", Link: "/g/fairlight", SubTitle: "FTL", Forward: "Fairlight returns after a few years absent"},
				{LinkTitle: "Origin", Link: "/g/origin", SubTitle: "OGN"},
				{LinkTitle: "RiSCiSO", Link: "/g/risciso", Forward: "Rise in Superior Couriering"},
			},
		},
		{
			Year:  1999,
			Title: "3Dfx vs. Nvidia", LinkTitle: "a short story of 3dfx - 5 steps to fall", Link: "https://level2.vc/a-short-story-of-3dfx/",
			Lead: "1999 was a complex year for PC gamers",
			Content: "The market pioneer, 3Dfx, with its Voodoo 3 GPU, had abandoned OEM manufacturers and decided to produce both the chips and graphic boards in-house. The change, intended to boost profits, led to manufacturing and global distribution shortages and decreased retail shelf space for 3Dfx products." +
				"<br>In the same year, Nvidia released its TNT and <strong>GeForce series</strong> of GPUs and became the go-to supplier of chips for OEM card manufacturers. Unlike 3Dfx, Nvidia was API agnostic and happy to prioritize Direct3D and OpenGL." +
				"<br>For gamers, the new 3Dfx cards were more challenging to obtain but offered the best compatibility for 3D games of the past few years. Plus, current games ran fast with better frames per second." +
				"<br>The high-end Nvidia products offered improved resolutions and graphic feature sets but poorer compatibility for older games developed primarily for the proprietary 3Dfx Glide API." +
				"<br>But by the end of 2000, 3Dfx was bankrupt, having taken on too much debt and railroaded themselves into a dead-end architecture. <a href=\"https://www.cnet.com/culture/nvidia-buys-out-3dfx-graphics-chip-business\">By April 2002</a>, the company's assets and intellectual property were owned by Nvidia.",
		},
		{
			Year: 1999, Prefix: notable,
			List: Links{
				{LinkTitle: "Razor 1911 Demo", Link: "/g/razor-1911-demo", SubTitle: "RZR", Forward: "Razor 1911"},
				{LinkTitle: "Scienide", Link: "/g/scienide", SubTitle: "SCI"},
			},
			Picture: Picture{
				Title: "Razor 1911 Demo production",
				Webp:  "a92f47.webp",
				Png:   "a92f47.png",
			},
		},
		{
			Title: "The copy party is over", Year: 2000, Month: 5, Day: 5, Highlight: true,
			Lead: "The end of Pirates with Attitude", LinkTitle: "the US DOJ press release", Link: "https://web.archive.org/web/20120114174415/http://www.justice.gov/criminal/cybercrime/pirates.htm",
			Content: "The US Department of Justice <strong>indicted 17 members</strong> of <a href=\"/g/pirates-with-attitudes\">Pirates with Attitudes</a> who got caught up in a honey pot scheme where, for months, Canadian law enforcement had taken control of the primary PWA FTP distribution (World HQ) site, Sentinel, running out of the University of Sherbrooke in Quebec." +
				"<p>A day later, <a href=\"/f/a23b69\">PWA published their final release</a>, a farewell NFO written by the fugitive Shiffie out of Belgium.</p>" +
				"Of the US defendants, 13 pleaded guilty. Five members were active employees of Intel Corp, and one was an employee of Microsoft Corp." +
				"<p><br>Less than a week later, Christian Morley, aka <q>Mercy,</q> a former senior organizer of PWA, <a href=\"https://ipmall.law.unh.edu/sites/default/files/hosted_resources/CyberCrime/pwa_verdict.pdf\">became the first person to be found guilty</a> under the No Electronic Theft (NET) Act and <u>the first to be guilty of conspiracy to infringe software copyrights</u>.</p>",
		},
		{
			Title: "Direct3D, the 3D graphic standard", Year: 2000, Month: 11,
			Lead: "DirectX 8.0 (4.08.00.0400)", LinkTitle: "the press release", Link: "https://news.microsoft.com/2000/11/09/microsoft-announces-release-of-directx-8-0",
			Content: "<p>The release of Microsoft's <strong>Direct3D</strong> 8 for all active editions of Windows from 95 through to XP was the beginning of the dominance of the proprietary 3D graphics API, as it is the first release offering compelling features for game developers.</p>" +
				"<p>For Microsoft, this helps to lock in Windows as the only operating system for modern PC gaming.</p>" +
				"<p>Prior editions of Direct3D, since 1996, have been clumsy and lacking features compared to the competing proprietary 3Dfx Glide or the industry OpenGL standard. Direct3D was instead a hardware fallback API for developers to support.</p>",
		},
		{
			Year: 2000, Prefix: notable + " onward,",
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
			Title: "The global takedown of Drink or Die", Year: 2001, Month: 12, Day: 11,
			Lead: "Operation Buccaneer", LinkTitle: "the Department of Justice overview", Link: "https://www.ojp.gov/ncjrs/virtual-library/abstracts/cross-national-investigation-and-prosecution-intellectual-property",
			Content: "<p><strong>Operation Buccaneer</strong> was the first global effort to target a specific warez group for criminal prosecution. Because of the nature of warez groups on the Internet, members operate in numerous countries and time zones. " +
				"The operation needed multiple law enforcement agencies in the USA, UK, Australia, Norway, Sweden, and Finland to coordinate the execution of <strong>70&nbsp;search&nbsp;warrants</strong>.</p>" +
				"<p>The target, <a href=\"/g/drink-or-die\">Drink&nbsp;or&nbsp;Die</a> is singled out for their multiple, pirate releases of Microsoft&nbsp;<a href=\"/f/bb2b71f\">Windows&nbsp;95</a> that occurred back in August 1995, over six years prior with a different group membership.</p>",
		},
		{
			Title: "Digital only scene releases", Year: 2004, Month: 10, Day: 7,
			Lead: "Counter-Strike: Source Final from Emporio", LinkTitle: "the release", Link: "/f/b1282e1",
			Content: "The online multiplayer title, <a href=\"https://www.mobygames.com/game/15518/counter-strike-source/cover/group-80271/cover-733563/\">Counter-Strike Source</a> was exclusively distributed on <strong><a href=\"https://steampowered.com\">Steam</a></strong>, Valve's digital distribution platform. " +
				"As there was no physical media available, this became a <strong>dubious&nbsp;release</strong> within the Scene and many groups didn't acknowledge the <a href=\"/g/emporio\">Emporio</a> package as a legitimate \"retail\" product or a \"final\" release. " +
				"<p>Due to the ease of supply and constant online patching, digital distribution was not well received.</p>" +
				"<p><q>SOME may contend the fact that this is BETA. <a href=\"https://web.archive.org/web/20050208205808/http://www.steampowered.com/index.php?area=news&archive=yes&id=327\">This is the version that is released on STEAM AS FINAL</a>. " +
				"You cannot do any better than this. The ... thing with STEAM is they can easily release many patches BUT EXPECT the EMPORiO crew to bring each and every patch CRACKED to your doorstep!</q></p>",
		},
		{
			Title: "Digital distribution and online activation", Year: 2004, Month: 11, Day: 16,
			Lead: "Half-Life 2", LinkTitle: "the and view the Steam page", Link: "https://store.steampowered.com/app/220/HalfLife_2",
			Content: "Half-Life 2 was once of the most anticipated games of the decade, and it was the first major game to use <strong><a href=\"https://steampowered.com\">Steam</a></strong>, Valve's digital distribution platform. " +
				"Steam was a huge shift in the way games were distributed, and it was the first time a major game <strong>required online activation</strong>. " +
				"<a href=\"https://www.amazon.com/product-reviews/B00006I02Z/ref=acr_dp_hist_1?ie=UTF8&filterByStar=one_star&reviewerType=all_reviews#reviews-filter-bar\">Steam often was not well received</a> by the gaming community, but it was a big success for Valve, and it paved the way for other digital distribution platforms. " +
				"Half-Life 2 was simultaneously released on <a href=\"https://store.steampowered.com/app/220/HalfLife_2/\">Steam</a>, " +
				"<a href=\"https://www.mobygames.com/game/15564/half-life-2/cover/group-90348/cover-246334/\">DVD</a> " +
				"and <a href=\"https://www.mobygames.com/game/15564/half-life-2/cover/group-16318/cover-38738/\">on CD</a>, but all three formats required Steam activation. ",
		},
		{
			Title: "Half-Life 2 *Retail*", Year: 2004, Month: 11, Day: 28,
			Lead: "Vengeance", LinkTitle: "the release", Link: "/f/b24c10",
			Content: "Half-Life 2 was one of the most anticipated games of the decade, and it was the first major game to use <strong>Steam</strong>, Valve's digital distribution platform. " +
				"<a href=\"/g/vengeance\">Vengeance</a> is the <u>first attempt to crack the Steam activation</u>, and it used an unusual Steam client and activation emulator. " +
				"But while playable, their pirate release of the game suffered with slower frame rates, load times, and the lack of multiplayer gameplay." +
				"<p>Vengeance would release the DVD *Retail* version with a tweaked crack two days later.</p>",
			List: Links{
				{LinkTitle: "Half Life 2 *Retail* [CD]", Link: "/f/b24c10"},
				{LinkTitle: "Half Life 2 DVD *Retail*", Link: "/f/a126f6"},
				{LinkTitle: "Half-Life 2 *Retail* Offline Installer", Link: "/f/b31a0b7"},
				{LinkTitle: "Half-Life 2 CDVersion Upgrade", Link: "/f/bc300c7"},
				{LinkTitle: "Half Life 2 trainer by Ages", Link: "/f/a63666"},
			},
		},
		{
			Title: "End of the line for RIPS", Year: 2005, Month: 10, Day: 9,
			Lead: "Farewell © Myth", LinkTitle: "the release", Link: "/f/a94129",
			Content: "Farewell © Myth is the <strong>final release</strong> from Myth, a group founded as <a href=\"/f/a53f3c\">Zeus</a> then <a href=\"/g/paradigm\">Paradigm</a> in 1996 and focused on <strong>ripping</strong> PC games from CD and later DVDs. " +
				"By the mid-2000s, broadband use was widespread and the desire for ripped CD or DVD games with missing content was dwindling. " +
				"Myth's longtime rival, <a href=\"/f/a53505\">Class, had already quit in early 2004</a>, and the other major competitor, <a href=\"/g/divine\">Divine</a>, finished up the following year." +
				"<p><q>We believe that the rip scene is one of incredible skill. " +
				"Not only is there the cracking talent needed to be successful like that of ISO, you must have dedicated coders and rippers to fully complete the task. " +
				"Much time is needed to perfect a rip like that of Neverwinter Nights. (We'll never forget you old friend) With the faster speed of the internet, equates to less usage of rips and just makes it not worth it. " +
				"When you are releasing upwards of 30 games a month some months, and you know not many people are downloading them. " +
				"It hardly gives you the rush of winning the title. We see groups throw out games now with stolen cracks and are completely non-working. These titles are not nuked, as no one even notices anymore, indeed a sad time in the scene.</q></p>",
		},
		{
			Year: 2016, Prefix: "", Highlight: true,
			Title: "The twilight of the cracktro",
			Content: "The first decade of the 2000s was the last time original <strong>quality cracktros</strong> were common in the Scene, primarily thanks to a few nostalgic Demosceners and warez sceners. " +
				"However, the number of people who could and were willing to create a decent cracktro dwindled as the skillset requirements got more specific and complex. " +
				"And so the cracktro was often forsaken for more straightforward methods of displaying the release information and branding. ",
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

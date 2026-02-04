package app

// Package file milestone.go contains the listings of the milestone page.

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
	Attribution string // Attribution is the name of the author of the picture.
	License     string // License is the license of the picture.
	LicenseLink string // LicenseLink is the URL to the license of the picture.
	Webp        string // Webp is the filename of the WebP screenshot.
	Png         string // Png is the filename of the PNG screenshot.
	Jpg         string // Jpg is the filename of the JPG photo.
	Avif        string // Avif is the filename of the AVIF photo.
	Webm        string // Webm is the filename of the WebM multimedia container, such as a video.
}

// Links is a collection of Links.
type Links []struct {
	LinkTitle string // LinkTitle is the title of the Link.
	SubTitle  string // SubTitle is the title of the Link in a smaller font and in brackets.
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
			Content: "<p>Ron Rosenbaum writes the first mainstream article on phone freaks, primarily kids who'd hack and experiment with the global telephone network.</p>" +
				"<p>The piece coins them as phone-freaks (<strong>phreaks</strong>) and introduces the reader to the kids' use of <strong>pseudonyms</strong> or codenames within their cliques and <strong>groups</strong> of friends. " +
				"It gives an early example of <strong>social engineering</strong>, defines the community of phreakers as the phone-phreak <strong>underground</strong>, and mentions the newer trend of <strong>computer phreaking</strong>, which we call <u>computer&nbsp;hacking</u> today.</p>",
		},
		{
			Year: 1971, Month: 11, Day: 15, Title: "The first civilian microprocessor",
			Lead: "Intel 4004", LinkTitle: "The Story of the Intel 4004",
			Link: "https://www.intel.com/content/www/us/en/history/museum-story-of-intel-4004.html",
			Content: "<p>Intel advertises the first-to-market general-purpose programmable processor or microprocessor, the 4-bit Intel&nbsp;4004. " +
				"Its main uses were in <a href=\"http://www.vintagecalculators.com/html/busicom_141-pf.html\">calculators</a>, some early automatic teller machines, and other embedded devices.</p>" +
				"<p>Busicom (formerly Nippon Calculating Machine Corp) <a href=\"http://nascojp.com/about.html\">first commissioned</a> the 4004 as part of a chipset for its 141PF Printing Calculator. " +
				"The 4000 chipset comprises four branded components: the 4001 read-only memory, 4002 RAM, 4003 shift register memory, and the <strong>4004 processor</strong>.</p>",
		},
		{
			Year: 1972, Month: 4, Title: "The first 8-bit microprocessor",
			Lead: "Intel 8008", LinkTitle: "The Story of the Intel 8008",
			Link: "https://www.intel.com/content/www/us/en/history/virtual-vault/articles/the-8008.html",
			Content: "<p>Intel released the world's first 8-bit microprocessor, the Intel&nbsp;8008. Despite the branding, it was not an 8-bit extension of the 4-bit Intel&nbsp;4004 but a new architecture.</p>" +
				"<p>Computer Terminal Corporation of Texas commissioned the new Intel chip for their cost-effective <a href=\"https://history-computer.com/datapoint-2200-guide/\">Datapoint 2200</a> computer terminal. Designed as a dumb terminal, CTC realized it could also operate as a programmable device with a central processing unit.</p>" +
				"<p>Manufacturing issues with the 8008 and deadlines meant that the Datapoint 2200 ditched the CPU. Instead, CTC followed the common practice of building the internals from discrete transistor-transistor (TTL) logic.</p>",
			Picture: Picture{
				Title:       "Intel 8008 CPU chip",
				Alt:         "A photo of an Intel C8008-1 CPU chip.",
				Jpg:         "intel-8008.jpg",
				Avif:        "intel-8008.avif",
				Attribution: "Konstantin Lanzet",
				License:     "CC BY-SA 4.0",
				LicenseLink: "https://creativecommons.org/licenses/by-sa/4.0/",
			},
		},
		{
			Year: 1972, Prefix: "Early", Title: "Blue boxes",
			Link: "https://explodingthephone.com/", LinkTitle: "about the hackers of the telephone network",
			Content: "Inspired by <a href=\"#secrets-of-the-little-blue-box\">The Secrets of the Little Blue Box</a> article, Steve Wozniak and a teenage Steve Jobs team up to build and sell dozens and dozens of the Wozniak-designed blue boxes to the students of the University of California, Berkeley. " +
				"The devices allowed users to hack and manipulate the electromechanical machines that operated the national telephone network—enabling them to call anywhere worldwide without incurring the typical prohibitively expensive costs.",
			Picture: Picture{
				Title:       "A blue box device",
				Alt:         "Blue box designed and built by Steve Wozniak.",
				Jpg:         "blue-box.jpg",
				Avif:        "blue-box.avif",
				Attribution: "Maksym Kozlenko",
				License:     "CC BY-SA 4.0",
				LicenseLink: "https://creativecommons.org/licenses/by-sa/4.0/",
			},
		},
		{
			Year: 1972, Title: "The first user-focused platform",
			Lead: "The PLATO IV", LinkTitle: "about the PLATO", Link: "https://arstechnica.com/gadgets/2023/03/plato-how-an-educational-computer-system-from-the-60s-shaped-the-future/",
			Content: "<p>In 1972, the PLATO system IV network came online as the second iteration of the University of Illinois's class-room education platform. " +
				"Its concept is to provide computer-based education on various broad subjects, not just computer literacy. This objective affected the network's design, end-user terminals, and software, leading to many computing and computer game design firsts.</p>" +
				"<p>The terminals connect to a supercomputer mainframe that eventually could support over 1,000 simultaneous users in various universities, colleges, and schools throughout Illinois and later setups out of state. However, the terminals and the special-purpose programming language used to develop the software make the network unusually special. One cannot overstate how advanced this platform and software is in its time.</p>" +
				"<p>Each monochrome terminal supports vector and bitmap graphics and offers an exceedingly high resolution of 512x512 pixels! This was twelve years before the Apple Macintosh System&nbsp;1 operating system, which only provided 512x342 resolution. The terminals and software provided keyboard text and user-friendly touchscreen input almost 40 years before the modern tablet.</p>" +
				"<p>Equally as important was the " +
				`<a href="https://distributedmuseum.illinois.edu/exhibit/tutor-programming-language/">TUTOR programming language</a> ` +
				"used to develop software on the platform. Designed for non-programmers and educators to build coursework delivered on the network, the language allowed easy access to all terminal and network hardware elements, such as vector and sprite graphics, custom fonts, communication, and touch input.</p>" +
				"<p>The photo shows a boy named Reid playing a touch game called PICTURE SHOW. " +
				`The PLATO IV had an optional audio peripheral that, in a <a href="https://files.eric.ed.gov/fulltext/ED148298.pdf">1977 report</a>, stated it was of poor quality and unreliable. ` +
				"Still, the image has to be one of the earliest examples of interactive multimedia, edutainment software and touch-first design. " +
				`Also, conflicting metadata makes it unknown if the photo is from <a href="https://umedia.lib.umn.edu/item/p16022coll91:445">1972</a>` +
				` or <a href="https://computerhistory.org/blog/meet-2021-chm-fellow-honoree-raymond-ozzie/">1976</a>,` +
				` but a <a href="https://umedia.lib.umn.edu/item/p16022coll91:193">1975 photo</a> of a girl of similar age using the same terminal model, headphones and touch exists, maybe interacting with the same software?</p>`,
			Picture: Picture{
				Title:       "A child using the PLATO IV system",
				Alt:         "A photo of a boy using the touch interface of the PLATO IV system.",
				Jpg:         "plato-iv.jpg",
				Avif:        "plato-iv.webp",
				Attribution: `is uncertain; the owner maybe "University of Illinois developer", Raymond Ozzie or a university`,
				License:     "source",
				LicenseLink: "https://grainger.illinois.edu/news/magazine/plato",
			},
		},
		{
			Year: 1973, Title: "The first online communities",
			Lead:      "PLATO IV Notes, Talkomatic and online games",
			LinkTitle: "about PLATO emulation delivered over the Internet", Link: "https://www.cyber1.org/",
			Content: "<p>Not long after the rollout of the PLATO IV system to various locations and the creation of specific software, online communities of friends and users started to develop. Most probably a first, people intentionally used the network outside of class or work to hang out, chitchat with others, and play multiplayer games online.</p>" +
				`<p>This all began with the August release of <a href="https://just.thinkofit.com/plato-the-emergence-of-online-community/">Notes by David Woolley</a>, a 17-year-old student and programmer. ` +
				`He was <a href="http://www.platohistory.org/blog/2013/08/plato-notes-released-40-years-ago-today.html">asked</a> to develop an app allowing PLATO users to ` +
				`<a href="https://digital.library.illinois.edu/collections/7bfaf980-0727-0130-c5bb-0019b9e633c5-e/tree">post bug reports</a> ` +
				`and for staff to reply with <a href="https://just.thinkofit.com/wp-content/uploads/1994/01/plato-base-note-nestedloops.jpg">back-and-forth communication</a>. ` +
				`A year later, Personal Notes by Kim Mast was released, allowing users to have private notes and, more importantly, to send notes directly to individuals as <strong>electronic messages</strong>.</p>` +
				"<p>Doug Brown released Talkomatic in the fall of 1973. This program allowed multiple people to occupy a <strong>chat room and talk</strong> in real time. Each user had " +
				`<a href="https://just.thinkofit.com/wp-content/uploads/1994/01/talko-comb.png">their own window</a>, ` +
				"and the text characters printed as they typed. After its success, the PLATO staff incorporated a form of direct chat into the system, allowing people to notify and page others for a real-time one-on-one chat like an <strong>instant message service</strong>.</p>" +
				"<p>At the start of 1976, Group Notes became the final evolution of the Notes concept, with the advice and feedback of many users and David's work. Groups allowed unlimited public and private notefiles for broad subject or <strong>topic-orientated discussions</strong>, such as books, music, movies, religion, science fiction, etc., years before Usenet or the " +
				`<abbr title="Computerized Bulletin Board System">CBBS</abbr>.` +
				`<p>Some people also used notefiles as a form of <strong>blogging</strong>, such as ` +
				`<a href="https://distributedmuseum.illinois.edu/exhibit/bruce-parello/">The Red Sweater</a>'s Newsreport or ` +
				`Dr. Gräper's <a href="http://www.grapenotes.com/">=grapenotes=</a>, and these could be inserted with <strong><a href="http://www.platopeople.com/emoticons.html">emoticons</a></strong>.</p>` +
				`<p>It seems out of the gate that various students and possibly staff started using the TUTOR programming language in 1972 to create multiplayer ` +
				`<a href="https://www.uvlist.net/platforms/games-list/181">games</a> on the PLATO IV. Titles include Chess, Dogfight, Backgammon, LIFE, Darwin1 and Moonwar. ` +
				`In Computer Lib/Dream Machines, <a href="https://archive.org/details/computer-lib-dream-machines/page/n29/mode/1up">Ted Nelson extensively wrote</a> about his visit and use of the PLATO IV in 1973 and dedicated a couple of pages to the games he uncovered on the network back then.</p>` +
				`<p>The most famous early multiplayer game on the PLATO was <a href="http://www.daleske.com/plato/empire.php">John Daleske's Empire</a>, released in May 1973. The original game supported up to eight players in a competitive strategic economic simulation.</p>` +
				`<p>A revised edition of Empire II was released in September and offered <strong>50 simultaneous players</strong> in eight teams a new game mechanic: spaceship tactical combat. The older economic simulation game was taken over by Silas Warner and redeveloped as ` +
				`<a href="https://datadrivengamer.blogspot.com/2019/07/games-79-80-empire-and-road-to-wizardry.html">Conquest</a>. John gave an optimization update to Empire II, which became known as ` +
				`<a href="https://datadrivengamer.blogspot.com/2019/07/games-79-80-empire-and-road-to-wizardry.html">Empire III</a>, with the same gameplay but on a much bigger playfield.</p>` +
				`<p>Inspired by the 1974 publication of Dungeons & Dragons, numerous authors created fantasy, computerized role-playing games (<strong>CRPG</strong>) ` +
				`<a href="https://crpgaddict.blogspot.com/2021/06/brief-everything-we-know-about-1970s.html">on the PLATO</a> system. Titles such as ` +
				`<a href="https://crpgaddict.blogspot.com/2019/01/revisiting-dungeon-1975.html">The Dungeon</a>, ` +
				`<a href="https://crpgaddict.blogspot.com/2019/01/revisiting-game-of-dungeons-1975.html">The Game of Dungeons</a>, ` +
				`<a href="https://crpgaddict.blogspot.com/2013/11/game-123-orthanc-1977.html">Orthanc</a>, Moria, and various games called Dungeon began in that year or 1975.</p>` +
				`<p>Unlike the solo CRPG games that were developed on microcomputers years later, these games, even when played solo, had a solid online component with competitive high scores, active player listings and <strong>permadeath</strong>. ` +
				`Games such as <a href="https://crpgaddict.blogspot.com/2013/11/game-121-moria-1975.html">Moria</a> and later ` +
				`<a href="https://crpgaddict.blogspot.com/2013/11/game-124-avatar-1979.html">Avatar</a> offered players to <strong>play together in co-op</strong> as members of a party exploring multiple levels on a large playworld.</p>` +
				`<p>Brand Fortner's Airfight from 1974 was a 3D combat flight simulator in which you did your best to take out the enemy being flown by human opponents in a <strong>multiplayer death match</strong>. The title is believed to be the first of the <strong>flight simulator genre</strong>. Meanwhile, 1975's Panther by John Haefeli looked much like Atari's arcade Battlezone from 1980, except you played against online humans!</p>` +
				`<p>Yet all games created on PLATO were passion projects by their authors. Unlike the pay-the-hour commercial online services that came much later or the physical media sale opportunities that would eventuate on microcomputers, the PLATO author had no means of monetizing if the thought ever crossed their mind.</p>`,
		},
		{
			Year: 1974, Month: 4, Title: "The first CPU for microcomputers",
			Lead: "Intel 8080", LinkTitle: "about The Intel 8008 and 8080",
			Link: "https://www.intel.com/content/www/us/en/history/virtual-vault/articles/the-8008.html",
			Content: "<p>Intel released the 8-bit <strong>8080 CPU</strong>, its second but far more successful 8-bit programmable microprocessor, " +
				"and the first mass-produced CPU suitable for personal microcomputing. " +
				"The 8080 and its later descendants, both from Intel and competitors, meant the 8080 architecture came to dominate the 8-bit CPU market of the 1970s and 1980s.</p>" +
				"<p>This CPU became the processing heart of the earliest popular microcomputers, the <a href=\"https://collection.powerhouse.com.au/object/167322\">Altair&nbsp;8800</a>, " +
				"the <a href=\"http://oldcomputers.net/sol-20.html\">Sol-20</a>, <a href=\"https://collection.powerhouse.com.au/object/153559\">IMSAI</a>, and later in arcade machines, " +
				"such as the cultural phenomenon that was <a href=\"https://www.computinghistory.org.uk/det/47162/40-Years-of-Space-Invaders/\">Space Invaders</a>.</p>",
		},
		{
			Year: 1975, Month: 1, Title: "The first popular microcomputer",
			Lead: "Altair 8800", LinkTitle: "about the Altair 8800",
			Link: "https://americanhistory.si.edu/collections/search/object/nmah_334396",
			Content: "<p>The worlds first popular microcomputer appears on the <a href=\"https://archive.org/details/197501PopularElectronics\">front cover</a> of Popular Electronics in the USA, the <strong>Altair&nbsp;8800</strong> by MITS running on the Intel <strong>8080 CPU</strong>. " +
				"Even for the time, the Altair was a primitive device, requiring toggle on/off switches for input and blinking red LED lights for output, and there was no way to save programs. But it was the first widely available programmable computer that didn't cost an arm, a leg, or a house.</p>" +
				"<p>Eventually, with the system's popularity and its use of the modular <a href=\"http://www.s100computers.com/History.htm#The%20S-100%20Bus\">S-100 bus interface</a>, an upgraded Altair platform allowed for storage, teletype-keyboard input, printer output and displays.</p>",
		},
		{
			Year: 1975, Month: 2, Title: "The first microcomputer software",
			Lead: "Altair BASIC", LinkTitle: "about origins of BASIC",
			Link: "https://time.com/69316/basic/",
			Content: "Paul Allen and Bill Gates program and sell <strong>Altair&nbsp;BASIC</strong> for the computer they first saw a month prior. " +
				"BASIC (Beginner's All-Purpose Symbolic Instruction Code) was a programming language conceived by John Kemeny and Thomas Jurtz of Dartmouth College in early 1964 to be as approachable as possible.",
			Picture: Picture{
				Title:       "Can anyone beat the Altair System?",
				Alt:         "A May 1976 advertisement for the Altair 8800 computer.",
				Jpg:         "altair-ad.jpg",
				Avif:        "altair-ad.avif",
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
				Avif:        "homebrew-computer-club.avif",
				Attribution: "Gotanero",
				License:     "CC BY-SA 3.0",
				LicenseLink: "https://creativecommons.org/licenses/by-sa/3.0/",
			},
		},
		{
			Year: 1976, Month: 1, Title: "Software piracy", Highlight: true,
			Lead: "An Open Letter to Hobbyists", LinkTitle: "the letter",
			Link: "https://archive.org/details/hcc0201/Homebrew.Computer.Club.Volume.02.Issue.01.Len.Shustek/page/n1/mode/2up",
			Content: "<p>Bill Gates of <em>Micro-Soft</em> pens a letter to the hobbyists of the Homebrew Computer Club requesting they <u>stop stealing</u> <strong>Altair&nbsp;BASIC</strong>. " +
				"However, while US copyright law protected the software author from plagiarism, it did not allow for restrictions to be placed on software usage. " +
				"For many hobbyists, the copying and sharing of retail software got viewed the same as copying to paper the instructions of a great recipe taken from a cookbook.</p>" +
				"<p><q>As the majority of hobbyists must be aware, most of you steal your software. Hardware must be paid for, but software is something to share. Who cares if the people who worked on it get paid.</q></p>",
			Picture: Picture{
				Title:       "An Open Letter to Hobbyists",
				Alt:         "A photo of the first page of the letter.",
				Jpg:         "an-open-letter-to-hobbyists.jpg",
				Avif:        "an-open-letter-to-hobbyists.avif",
				Attribution: "Len Shustek",
				License:     "public domain",
				LicenseLink: "https://commons.wikimedia.org/wiki/File:Bill_Gates_Letter_to_Hobbyists.jpg",
			},
		},
		{
			Year: 1976, Month: 3, Title: "The Apple Computer",
			Lead: "By the APPLE Computer Company", LinkTitle: "about the Apple-1",
			Link: "https://www.computerhistory.org/revolution/personal-computers/17/312/1132",
			Content: "<p>Steve Wozniak and Steve Jobs released The Apple Computer, later rebranded as the Apple I. It was a single-board device for electronic hobbyists with a " +
				"<a href=\"https://spectrum.ieee.org/chip-hall-of-fame-mos-technology-6502-microprocessor\">MOS 6502</a> CPU, 4KB of RAM, and a 40-column display controller.</p>" +
				"<p>Unlike the far more popular Altair&nbsp;8800, The Apple Computer wasn't usable out of the box and didn't come with a case. However, <a href=\"https://upload.wikimedia.org/wikipedia/commons/4/48/Apple_1_Advertisement_Oct_1976.jpg\">it did offer</a> a convenient video terminal, cassette, and keyboard interface, but requires owners to supply peripherals for input, output, and storage.</p>" +
				"<p>The board is a commercial failure, selling less than 200 units, and could be considered more of a prototype for the company and third-party investors. The following year, the product line was <a href=\"https://www.applefritter.com/node/2706\">replaced with circuit boards</a> housing an Apple II.</p>" +
				"<p>The choice of the new <strong>MOS 6502 CPU</strong> showed foresight, as it became the foundation of many successful microcomputers and consoles.<p>" +
				ul0 +
				"<li>Atari&nbsp;2600 <sup>1977</sup></li>" +
				"<li>Apple&nbsp;II <sup>1977</sup></li>" +
				"<li>Commodore&nbsp;PET <sup>1977</sup></li>" +
				"<li>Commodore&nbsp;VIC-20 <sup>1981</sup></li>" +
				"<li>Commodore&nbsp;64 <sup>1982</sup></li>" +
				"<li>Nintendo&nbsp;Entertainment&nbsp;System <sup>1983</sup></li>" +
				ul1,
		},
		{
			Year: 1977, Title: "Apple II, Commodore PET, Tandy TRS-80",
			Lead: "The second generation of microcomputers", LinkTitle: "about the Apple II, Commodore PET and Tandy TRS-80",
			Link: "https://cybernews.com/editorial/the-1977-trinity-and-other-era-defining-pcs/",
			Content: "<p>The <strong>Commodore&nbsp;PET</strong>, <strong>Apple&nbsp;II</strong>, and the <strong>Tandy TRS-80</strong> " +
				"became the first successful microcomputers marketed to a mainstream consumer rather than an electronics hobbyist. " +
				"By the end of the year, a potential customer in the USA could walk into a mall or specialist retail shop and walk out with a complete personal computer ready to use. However, in 1977, things began slowly for Commodore and Apple.</p>" +
				"<p>In the January 1978 issue of Creative Computing, <a href=\"https://archive.org/details/197801ROMV1I07/page/58/mode/2up\">the article</a>, <em>Home Computers: A look at what's coming</em>, didn't even review the Apple microcomputers; instead, it previewed affordable machines by RCA, Bally, and National Semiconductor, none of which are well known today.</p>" +
				"<strong>Commodore PET</strong> <em>Personal Electronic Transactor</em><br>" +
				"<p>Commodore was the first to announce its machine in January at CES, but shipping only occurred in mid-October. Even then, the numbers were tiny, with the end-of-year batches reaching just 500 boxed machines.</p>" +
				"<strong>Apple II</strong><br>" +
				"<p>Apple didn't fare much better, as its <a href=\"https://www.fastcompany.com/4001956/apples-sales-grew-150x-between-1977-1980-2\">revenue until the end of September 1977</a> was just USD&nbsp;774,000, which includes sales of both the Apple&nbsp;I and the mid-April launch of the Apple&nbsp;II. " +
				"Its <a href=\"https://web.archive.org/web/20140124082855/https://www.swtpc.com/mholley/Apple/Apple_IPO.pdf\">December 1980 stock perspective</a> states, <q>Net sales in fiscal 1977 occurred primarily in the fourth fiscal quarter and consisted principally of sales of the basic Apple II mainframe computer.</q> " +
				"Given the expensive Apple&nbsp;II <a href=\"https://www.applefritter.com/node/2703\">is priced at</a> $1300-2600, the number of machines sold could have been in the hundreds.</p>" +
				"<strong>Tandy TRS-80</strong><br>" +
				"<p>Sales of the Tandy were considerable. It was <a href=\"https://www.radioshackcatalogs.com/flipbook/c1977_rsc-01.html\">announced at</a> the end of July and priced from $400 or $500, including a display. " +
				"It was widely available nationally through the thousands of RadioShack retail stores, and took 10,000 unit <a href=\"https://www.wired.com/2010/08/0803trs-80-computer-launch/\">orders in the first month</a>, birthing the microcomputer revolution! " +
				"The November 1977 <a href=\"https://archive.org/details/ROM05_201806/page/n50/mode/1up\">issue</a> of ROM announced, <em>Radio Shack is for real with its realistically priced (if not so named) micro. The ready-to-plug-in-and-run TRS-80 sells for $599.95 complete with a fifty-three-key keyboard, regulated power supply, interfaced cassette recorder, and twelve-inch video display monitor. " +
				"As if the low price isn’t enough, the real marketing con is the instant availability of five prerecorded programs. For a complete library Radio Shack is still the premier purveyor of ready-to-run systems with something to run. " +
				"Applications software so far includes the demonstration blackjack and backgammon cassette that comes with the unit as well as a payroll program, a math education program, and a personal finance program. More on the way. All on prerecorded cassettes. At your local Radio Shack.</em></p>" +
				"<p>Creative Computing would <a href=\"https://archive.org/details/CreativeComputingbetterScan197809/page/n37/mode/1up\">report</a> on the sales up to mid-1978, saying Commodore had shipped 15000 PETS, Tandy had shipped somewhere between 8000-20000 TRS-80 machines, and calculated that the secretive Apple had shipped 25000 units.</p>",
		},
		{
			Year: 1978, Title: "CP/M operating system",
			Lead:      "The forgotten origins of Microsoft Windows",
			LinkTitle: "The History of CP/M", Link: "https://archive.org/details/CreativeComputingbetterScan198311/page/n205/mode/2up",
			Content: "<p>" +
				"Digital Research releases version 1.4 of CP/M, the operating system for the Intel 8080 CPU." +
				"</p><p>" +
				"In 1973, <a href=\"https://www.youtube.com/watch?v=V5S8kFvXpo4\">Gary Kildall</a>, an occasional consultant for Intel's microprocessor division, began collecting hardware that would form a complete microcomputer system based on the new Intel 8080 CPU. This was in the era before off-the-shelf systems could be found." +
				"</p><p>" +
				"Gary needed a way to link all the hardware components together in software, so he wrote a simple operating system in a high-level programming language he had created for Intel, the Program Language for Microcomputers or PL/M. " +
				"The new operating system was later given the name <strong>C</strong>ontrol <strong>P</strong>rogram/<strong>M</strong>onitor, more commonly called CP/M." +
				"</p><p>" +
				"Gary attempted to get Intel involved in his pet project, but they showed no interest. This wasn't surprising, given the limited availability of microcomputers and Intel's own operating system development for the 8080 CPU, the Intel System Implementation Supervisor." +
				"</p><p>" +
				"After the rejection, Gary and his wife Dorothy went out on their own in 1974, forming Intergalactic Digital Research to further develop and market the software. Initially, marketing it directly to hobbyists, but later discovered the new market of hardware manufacturers. " +
				"In 1975, several small companies were selling microcomputers to hobbyists, which included both custom hardware and their own simple operating systems. However, developing system software was time-consuming and expensive, so many of these small companies adopted Gary's CP/M. " +
				"By doing so, they could focus on the hardware, and the CP/M platform evolved into a de facto standard." +
				"</p><p>" +
				"CP/M was an 8-bit operating system that worked on 8-bit microprocessors like Intel's 8080 and the Z80 by Zilog. " +
				"However, in 1980, a couple of years after Intel's first 16-bit processor entered the market. " +
				"It was not Digital Research, but a small hardware manufacturer named Seattle Computer Products, that was one the first to release a purpose-built 16-bit microcomputer operating system, 86-DOS. " +
				"A scrapy and rushed system that was patterned after CP/M&nbsp;version&nbsp;1.4, but was incompatible due to the methods it used to handle disk data." +
				"</p><p>" +
				"86-DOS would be purchased by Microsoft for a secret IBM contract, and rebranded as PC-DOS for the IBM&nbsp;PC. " +
				"Microsoft would rewrite the software from scratch and release it as Microsoft&nbsp;MS-DOS&nbsp;v2, but it still kept the same CP/M patterns and commands. " +
				"Controversially, MS&hyphen;DOS would take over the markets of both Digital Research and IBM, and become the basis of Microsoft Windows, later evolving into the Windows Command Prompt. " +
				"And while there were hundreds of enhancements to MS-DOS, the Command Prompt and the more recent Windows Terminal, " +
				"for backward compatibility and user muscle memory, Microsoft always kept the original CP/M design patterns. " +
				"Modern annoyances or features such as drive letters, the use of back slashes, three-letter filename extensions, CR+LF newlines, the end-of-file marker, and " +
				"commands: DIR, REN, TYPE, etc." +
				"</p>",
		},
		{
			Year: 1978, Month: 2, Title: "The first computerized bulletin board system",
			Lead: "CBBS", LinkTitle: "the Byte Magazine article", Link: "https://vintagecomputer.net/cisc367/byte%20nov%201978%20computerized%20BBS%20-%20ward%20christensen.pdf",
			Content: "<a href=\"https://portcommodore.com/dokuwiki/doku.php?id=larry:comp:bbs:about_cbbs\">Ward Christensen</a> and Randy Suess create the first bulletin board system (<strong>BBS</strong>), the <em>Computerized Bulletin Board System</em> (<strong>CBBS</strong>) in Chicago. " +
				"The software was custom written in 8080 assembler language which ran on a <strong>S-100 bus</strong> computer together with the brand new $300, <a href=\"http://www.s100computers.com/Hardware%20Folder/DC%20Hayes/103/103%20Modem.htm\">Hayes 110/300</a> baud modem. " +
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
			Content: "<p>Intel released the 16-bit programmable microprocessor, the <strong>Intel&nbsp;8086</strong>, which began the <em>x86-architecture</em> and Intel PC platform.</p>" +
				"<p>In July 1976, the startup Zilog launched its first product, the <a href=\"https://spectrum.ieee.org/chip-hall-of-fame-zilog-z80-microprocessor\">Z80 CPU</a>, an enhanced, cheaper and software-compatible 8080 clone. " +
				"Eventually, the Z80 became one of the most successful 8-bit CPUs. " +
				"Months later, Intel released the <a href=\"https://timeline.intel.com/1976/8085-microprocessor\">8085</a>, an update to the 8080 CPU line, improving circuitry power requirements and reducing implementation costs.</p>" +
				"<p>The development and launch of the 8086, a software-compatible 16-bit implementation of the 8080 and the 8085, is a direct response to the Z80 and the market of clone CPUs. " +
				"However, the 8086 failed to dominate an industry saturated with more affordable 8-bit hardware.</p>",
			Picture: Picture{
				Title:       "A recreation of CBBS",
				Alt:         "A recreation screen capture of the first BBS.",
				Png:         "intel-8086.jpg",
				Avif:        "intel-8086.avif",
				Attribution: "Thomas Nguyen",
				License:     "CC BY-SA 4.0",
				LicenseLink: "https://creativecommons.org/licenses/by-sa/4.0/deed.en",
			},
		},
		{
			Title: "The first copy protections", Year: 1978,
			Lead: "Cassette tapes", LinkTitle: "Andy McFadden's article",
			Link: "https://fadden.com/apple2/cassette-protect.html",
			Content: "<p>While forgotten today, cassette tapes were the popular medium of publishing <a href=\"https://retro365.blog/2024/09/25/bomber-bob-bishop-pioneering-the-apple-ii/\">retail software</a> on microcomputers in the late 1970s. " +
				"Compared to the more expensive floppy disks, compact cassette tapes were less durable and harder to pirate due to their analog nature, but were terribly slow when used for data.</p>" +
				"<p>Andrew McFadden wrote about copy protection program routines found in some of the first Apple II games that were published on cassettes. He discovered the following games had some form of copy protection.</p>" +
				"<ul>" +
				"<li>1978 <em>Personal Software</em>, <a href=\"https://www.mobygames.com/game/29710/microchess/\">Microchess 2.0</a></li>" +
				"<li>1978 <em>Softape</em>, Instant Library: <a href=\"https://www.mobygames.com/game/146664/instant-library-module-6/\">Module 6</a> (Blackjack)</li>" +
				"<li>1978 <em>Softape</em>, Sargon II</li>" +
				"<li>1979 <em>Hayden Book Company</em>, <a href=\"https://www.mobygames.com/game/22058/sargon-ii/\">Sargon II</a></li>" +
				"</ul>" +
				"<p>Roland Gustafsson, an early creator of disk protection schemes on the Apple II, stated in a " +
				"<a href=\"http://underground2e.free.fr/Underground/Deplombage/Interviews/The_Wizard/Roland_Gustafsson.html\">2014 interview</a> that the first copy protection he discovered in the wild was the cassette tape release of " +
				"<a href=\"https://www.mobygames.com/game/18186/flight-simulator/\">Flight Simulator</a>. The game was developed in 1979 by Bruce Artwick and published by subLOGIC in either January or <q>early</q> 1980.</p>" +
				"<p>Roland would graduate from high school in 1981 and work freelance to develop custom disk copy protections for SSI, Brøderbund Software, Gebelli Software, and Silicon Valley Systems. " +
				"And for SSI, he would create RDOS for the strategy game publisher, a custom <q>protected</q> disk operating system that used a tiny memory footprint and was very fast. " +
				"A similar, performant self-booting custom disk operating system concept got <a href=\"https://www.mobygames.com/platform/pc-booter/sort:date/\">popular</a> on the IBM PC in 1982.</p>" +
				"<p>The <a href=\"https://archive.org/details/CreativeComputingbetterScan198001/page/n135/mode/1up\">correspondence</a> for Creative Computing January 1980 wrote, <em>" +
				"...pointed out some software problems that are also appropriate for Apple owners. The particular problem has to do with protected software. " +
				"That is, software designed to prevent you from making copies or changes, or that destroys itself if you make such an attempt. " +
				"If software theft is a real problem then there is possibly some advantage to the seller. I say possibly because a good programmer can “fix” the software anyway. And, with special equipment, " +
				"anything that is recorded on a magnetic surface can be copied onto a magnetic surface. For the most part, the attempts to prevent theft will only cause problems for the purchaser. If the software can’t be copied for back-up then the user has to buy another when the " +
				"original wears out.</em></p>" +
				"<p>Later in the year, the Creative Computing <a href=\"https://archive.org/details/CreativeComputingbetterScan198011/page/n145/mode/1up\">article</a>, " +
				"<q>TRcopy and the Pirate</q> contained, <em>The instruction book stated that <a href=\"https://americanhistory.si.edu/collections/object/nmah_1756588\">TRcopy</a> couldn’t be used to copy itself. Hmmmm. Of course I tried and they’re right. ... That the programmer of TRcopy felt the need to build in a self-defense mechanism says something very interesting about the program ... and about the state of the software industry in general. Let’s face it. TRcopy and the others, Duplik, SYSCOP, COPSYS, Clone, etc. are the computer world’s software answer to the Xerox machine. They are programs designed to violate copyright laws.<br>" +
				"I took my TRcopy to a local Radio Shack. With the manager’s help, I managed to copy nearly $200 worth of software onto a $2 cassette tape inside of twenty minutes. TRcopy, thus, becomes an interesting means of shoplifting. It is also curious in that the local Radio Shack isn’t out anything. They’ve made a $2 sale, where they probably wouldn’t have made the $200 sale. " +
				"Perhaps TRcopy should be illegal. Still, the program is irrelevant. (The Xerox machine isn’t the counterfeiter.) The duplicating programs are inevitable and the fact it can be done means that it will be done. " +
				"<br>It is a mark of the maturity of our industry that we have finally produced our own pirate industry.</em></p>",
		},
		{
			Title: "The first copy protections", Year: 1979,
			Lead:      "Floppy disks",
			LinkTitle: "the Who Copied What? table", Link: "https://archive.org/details/hardcore-computing-1/page/n13/mode/2up",
			Content: "<p>Using the contemporary print media of the era, we can propose that copy protections for floppy disks began appearing in software in late 1979 and became commonplace in late 1980.</p>" +
				"<p>The October 1980 issue of Byte wrote, <em>While attempts to eliminate software piracy are commendable, they very often fail because of the cleverness of personal-computer users; many take the anticopy measures as a challenge. The problem lies in making the protection scheme easy enough to be affordable, but complex enough to work.</em></p>" +
				"<p>SoftTalk <a href=\"https://vintageapple.org/softalk/pdf/SOFTALK_8101_v1_n05.pdf\">January 1981</a> interviews an executive of California Pacific. " +
				"<em>Good copy protection has a high priority at California Pacific. <q>In areas where Super Invaders was sold on cassette and unprotected, Trilogy (ed: with disk copy protection) outsold Invaders ten to one</q></em>. " +
				"The article writes, <em>California Pacific's Super Invaders hit northern California in mid-October of 1979 and Trilogy* followed in December.</em> <small>*Bill&nbsp;Budge's&nbsp;Trilogy&nbsp;of&nbsp;Games</small>" +
				"<p>The debut of HardCore Computing was published in June 1981, with the articles written earlier. " +
				"Of note is the piece found on page 10, <em>bit copy programs - that will copy the <q>uncopyables</q></em> by Karen Fitzpatrick. " +
				"It writes, <em>A bit-copier is a MUST for anyone who purchases <q>protected</q> software</em> and goes on to review three floppy duplication programs:<br>" +
				"<a href=\"https://archive.org/details/Locksmith_2.02.13.03.14.1\">Locksmith</a> by Omega Software Products first published in January 1981.<br>" +
				"<a href=\"https://archive.org/details/Compilation_Super_Disk_Copy_3.6_PDQ_Copy_Copy_II_Plus_3.0_Locksmith_4.1_Nibbles_\">Back-It-UP</a> by Sensible Software, from 1981 that offered <em>quick & dirty</em>+<em>old faithful</em>.<br>" +
				"<a href=\"https://archive.org/details/wozaday_Copy_II_Plus_1_0\">Copy II Plus</a> by Central Point Software from 1981.</p>" +
				"<p>On page 12 the article contains the table <em>Who Copied What?</em> and lists a collection of Apple II software tested, all of which must have some form of disk copy protection.</p>" +
				"<ul>" +
				"<li>1980 <u>Dec</u> <em>Brøderbund Software</em>, <a href=\"https://www.mobygames.com/game/81438/apple-galaxian/\">Apple Galaxian</a></li>" +
				"<li>1980 <u>Dec</u> <em>California Pacific</em>, <a href=\"https://www.mobygames.com/game/1256/akalabeth-world-of-doom/\">Akalabeth: World of Doom</a></li>" +
				"<li>1981 <u>Feb</u> <em>Hayden Software</em>, <a href=\"\">Reversal</a></li>" +
				"<li>1981 <u>Apr</u> <em>Highlands Computer Services</em>, <a href=\"https://www.mobygames.com/game/116064/creature-venture/\">Creature Venture</a></li>" +
				"<li>1981 <u>May</u> <em>M.D Software</em>, Disc-O-Doc <small>disk utility advertised on page 60 of Softalk May 1981</small></li>" +
				"<li>1980 <u>Nov</u> <em>Micro Lab</em>, The Data Factory <small>a database application</small></li>" +
				"<li>1980 <u>Dec</u> <em>MUSE</em>, <a href=\"https://www.mobygames.com/game/63061/abm/\">ABM: Anti-Ballistic Missile Game</a></li>" +
				"<li>1979 <em>IUS</em>, <a href=\"https://elisoftware.org/w/index.php?title=Information_Unlimited_Software\">EasyWriter</a> <small>co-authored by notorious phreaker John Draper</small></li>" +
				"<li>1980 <u>Sep</u> <em>On-line Systems</em>, <a href=\"https://www.mobygames.com/game/1761/hi-res-adventure-2-the-wizard-and-the-princess/\">The Wizard and the Princess</a></li>" +
				"<li>1980 Dec <em>On-line Systems</em>, <a href=\"https://www.mobygames.com/game/15282/hi-res-adventure-0-mission-asteroid/\">Mission Asteroid</a></li>" +
				"<li>1981 Feb <em>Personal Software</em>, <a href=\"https://www.mobygames.com/game/50/zork-the-great-underground-empire/\">Zork: The Great Underground Empire</a></li>" +
				"<li>1980 <em>Sirius Software</em>, <a href=\"https://www.mobygames.com/game/92109/both-barrels/\">Both Barrels</a></li>" +
				"<li>1980 <u>Nov</u> <em>Sirius Software</em>, <a href=\"https://www.mobygames.com/game/47942/cyber-strike/\">Cyber Strike</a></li>" +
				"<li>1980 <em>Sirius Software</em>, <a href=\"https://allincolorforaquarter.blogspot.com/2015/08/nasir-gebelli-and-early-days-of-sirius.html\">E-Z Draw</a></li>" +
				"<li>1980 <u>Dec</u> <em>Sirius Software</em>, <a href=\"https://www.mobygames.com/game/70193/phantoms-five/\">Phantoms Five</a></li>" +
				"<li>1980 <u>Dec</u> <em>Sirius Software</em>, <a href=\"https://www.mobygames.com/game/43500/star-cruiser/\">Star Cruiser</a></li>" +
				"<li>1980 <u>Dec</u> <em>SSI</em>, <a href=\"https://www.mobygames.com/game/54998/computer-air-combat/\">Computer Air Combat</a></li>" +
				"<li>1980 <em>SSI</em>, <a href=\"https://www.mobygames.com/game/157493/computer-ambush/\">Computer Ambush</a></li>" +
				"<li>1980 <u>Sep</u> <em>SSI</em>, <a href=\"https://www.mobygames.com/game/50900/computer-quarterback/\">Computer Quarterback</a></li>" +
				"<li>1981 <u>Mar</u> <em>SSI</em>, <a href=\"https://www.mobygames.com/game/2907/the-warp-factor/\">The Warp Factor</a></li>" +
				"<li>1980 <em>Top of the Orchard</em>, <a href=\"https://mirrors.apple2.org.za/ftp.apple.asimov.net/documentation/applications/misc/Bill%20Budges%203-D%20Graphics%20System%20and%20Game%20Tool.pdf\">Bill Budge's 3-D Graphics System and Game Tool</a></li>" +
				"</ul><small><u>Month</u>, found in a Softalk magazine advert or review.</small>",
		},
		{
			Title: "The first popular x86 CPU and commercial software", Year: 1979, Month: 6,
			Lead: "Intel 8088 + Microsoft BASIC-86", LinkTitle: "about the Intel 8088",
			Link: "https://spectrum.ieee.org/chip-hall-of-fame-intel-8088-microprocessor",
			Content: "Intel releases a lesser 16-bit microprocessor, the <strong>Intel&nbsp;8088</strong>. " +
				"While <u>fully compatible</u> with the earlier Intel&nbsp;8086 CPU, this model is intentionally \"castrated\" with an 8-bit external data bus. " +
				"The revision is an improvement for some buyers as it needs less expensive mainboard support chips and is compatible with the more readily available 8-bit hardware. " +
				"<p>Software written for either CPU often gets quoted as <a href=\"https://archive.org/details/msdos-200-users-guide-1983/page/n3/mode/2up\">8086/8088 compatible</a>.</p>" +
				"<p>Also in June on the 18th, Microsoft <a href=\"https://thisdayintechhistory.com/06/18/microsoft-introduces-basic-for-8086/\">published</a> BASIC on the x86 platform. " +
				"<a href=\"https://www.computerhistory.org/collections/catalog/102623976\">Microsoft BASIC</a> and its many revisions were the first killer applications for Microsoft in its early years. " +
				"Microcomputers were often sold to enthusiasts or businesses, but the software availability for these machines was lacking. " +
				"So many owners resorted to building software, and the BASIC programming language had an easy learning curve. " +
				"Though Microsoft didn't invent the language, its implementation was considered the gold standard.</p>",
		},
		{
			Title: "The early online underground", Year: 1979, Highlight: true,
			Lead: "CBBS, ABBS, and the Apple II microcomputer",
			Content: "<p>Even this early, in the USA at least, there were commercial online services for microcomputers owners with modems being provided by CompuServe and <a href=\"https://archive.org/details/CreativeComputingbetterScan197910/page/n77/mode/2up\">The Source</a>. " +
				"At the time, they offered real-time chat, electronic mail, sports, news, weather, stocks, and interactive entertainment for a high, hourly fee.<sup><a href=\"#the-early-underground-fn3\">[3]</a></sup></p>" +
				"<p>However, for those who didn't want to pay the usage charges of the commercial offerings, the <em>Computerized Bulletin Board System</em> was the primary tool for communication between microcomputer owners. " +
				"In these early days, the setups allowed people to dial in using their computers to share and read public or private messages with other callers.</p>" +
				"<p>The earliest <strong>CBBS</strong> setups ran off <a href=\"http://www.s100computers.com/\">S-100 bus-based computers</a>. " +
				"These systems shared a common \"S-100 interface bus\" but otherwise, were incompatible platforms fabricated by many manufacturers of the 1970s. When the Apple&nbsp;II received CBBS-like software in 1979, it was typically called <strong>ABBS</strong>, an Apple Bulletin Board System or Service. " +
				"By September 1979, nationwide listings<sup><a href=\"#the-early-underground-fn1\">[1]</a></sup> for dozens of bulletin boards were running as ABBS, CBBS, and on other platforms.</p>" +
				"<p>1979 also saw the introduction of Corvus Systems and their 10MB hard drive solutions for these same microcomputers. " +
				"While the drives were prohibitively expensive, in 1981, the units could be shared between numerous microcomputers using a local area network configuration named <a href=\"http://www.bitsavers.org/pdf/corvus/brochures/PC_Omninet_Brochure.pdf\">Omninet</a>.</p>" +
				// press attention
				"<p>In the first days of the BBS, the mainstream computer press paid attention to boards, " +
				"<a href=\" https://books.google.com.au/books?id=3j4EAAAAMBAJ&pg=PA10&lpg=PA10&dq=%22Modem+Over+Manhattan%22&source=bl&ots=smYwZj_okV&sig=ACfU3U0kYG9RX-3uPfGTakGgtP_mVDcAhA&hl=en&sa=X&ved=2ahUKEwiVs-yi6-qEAxX-oWMGHYpwAPA4ChDoAXoECAIQAw#v=onepage&q=%22Modem%20Over%20Manhattan%22&f=false\">including write-ups</a>" +
				"<sup><a href=\"#the-early-underground-fn2\">[2]</a></sup> and listings of the phone numbers for known underground boards.</p>" +
				"<p>The <u>underground</u> terminology may have <a href=\"https://archive.org/details/197708ROMV1I02/page/n6/mode/1up\">originated</a> from the CB (Citizens Band) <q>ham</q> radio communities, which were among the earliest adopters of single-board and micro-computers.</p>" +
				sect0 +
				"<div id=\"the-early-underground-fn1\">[1] See page 3 under <em>MODEMania</em> in the <a href=\"https://mirrors.apple2.org.za/ftp.apple.asimov.net/documentation/magazines/washington_apple_journal/washingtonapplepijournal1979v1no8sep79.pdf\">Washington Apple Journal</a>.</div>" +
				"<div id=\"the-early-underground-fn2\">[2] In the Innovative Bulletin Boards list, InfoWorld mislabels <strong>8</strong>BBS as BBBS.</div>" +
				"<div id=\"the-early-underground-fn3\">[3] An hour of online usage on The Source was more expensive than a cinema movie ticket.</div>" +
				sect1,
		},
		{
			Title: "The first boards", Year: 1979, Highlight: true,
			Lead: "Some of the early underground boards and online communities",
			Content: "" +
				// Sherwood Forest
				"<strong>Sherwood Forest</strong><br>" +
				"<p>A very early, underground ABBS is the 1979-1981 New Jersey-based<sup><a href=\"#the-first-boards-fn1\">[1]</a></sup> board, <strong>Sherwood&nbsp;Forest</strong>, created by Magnetic Surfer. " +
				"It runs off a floppy disk and a Micromodem and became a hub for some active telephone hackers who were early adopters of microcomputers in the New York Tri-state area—many became Scene pirates and notorious computer phreakers and hackers.</p>" +
				// Modem over Manhattan
				"<strong>Modem Over Manhattan</strong><br>" +
				"<p>As its name suggests, <strong>MOM</strong>, or <strong>Modem&nbsp;Over&nbsp;Manhattan</strong> (+212-245-4363, +212-912-9141), was based in Manhattan, New York, and probably went online in 1980. " +
				"It is another famous open board with lax rules that was popular with the New York phreak community.</p>" +
				// Pirate's Harbor
				"<strong>Pirate's Harbor</strong><br>" +
				"<p>Pirate's Harbor was an early pirate discussion board in Boston that also shared cracking techniques, guides and likely later on wares. " +
				"We know it was online in 1981 due to an <a href=\"https://archive.org/details/hardcore-computing-3/page/n19/mode/2up\">article</a> by Mike Flynn in HardCore Computing #3 from 1982 who wrote about the board being frequented by one of the developers of the famous game " +
				"<a href=\"https://www.mobygames.com/game/1209/wizardry-proving-grounds-of-the-mad-overlord/\">Wizardry</a> by Sir-tech Software.</p>" +
				// Pirate Trek
				"<strong>Pirate-Trek</strong><br>" +
				"<p>An early pirate board, the original <strong>Pirate-Trek</strong> out of New York (+914-634-1268), possibly run by the famed Apple&nbsp;II " +
				"<a href=\"https://ascii.textfiles.com/archives/828\">cracker Krakowicz</a>, " +
				"was <a href=\"http://artscene.textfiles.com/intros/APPLEII/cyclod.gif\">first announced</a> in 1981.</p>" +
				// 8BBS
				"<strong>8BBS</strong><br>" +
				"<p>There is also the renowned <strong>8BBS</strong> that operated near San Jose, CA, from 1980 to 1982 and ran on a <a href=\"https://www.computerhistory.org/revolution/minicomputers/11/331\">PDP-8 minicomputer</a>. " +
				"Unlike the other early underground boards, a chunk of the message base has been paperprinted, scanned, and preserved online! So it has its own <a href=\"#8bbs\">8BBS milestone article</a>.</p>" +
				sect0 +
				"<div id=\"the-first-boards-fn1\">[1] In a 1987 interview, <a href=\"http://www.textfiles.com/phreak/tuc-intr.phk\">TUC states</a> the first Sherwood Forest was in New Jersey, but other sources suggest it was in Manhattan, NY.</div>" +
				sect1,
		},
		{
			Title: "Widespread disk copying leads to copy protections", Year: 1979,
			Link: "https://vintageapple.org/softalk/pdf/SOFTALK_8010_v1_n02.pdf", LinkTitle: "the October 1980 issue of Softalk",
			Content: "" +
				"<p>It's easy to imagine software piracy in the early microcomputer era as online exchanges. " +
				"Online digital services existed in the late 1970s ~ early 1980s, and one might assume that is how piracy was always done. " +
				"However, that is not the case, both due to the hardware limitations of the time and the hyperlocalization of computer users. " +
				"While some modems existed, they were unusable with file transfers for most, and unaffordable hard drives were very rare. " +
				"Many online providers such as computerized bulletin boards, only facilitated message posting and replying.</p>" +
				"<p>Some people did used those online messaging services to coordinate in-person meetups, to converse, share ideas, programming, and of course, exchange commercial software. " +
				"This coordination wasn't exclusive to online digital services; traditional advertising in newspapers, print magazines, and paper flyers was far more popular, and local computer clubs would advertise themselves, renting out venues and meeting regularly.</p>" +
				"<p>The October 1980 of Softalk <q>Pirate, Thief. Who Dares to Catch Him?</q> is one of the first to document the problem of software piracy, which is described as <q>very young, and it can be stopped.</q></p>" +
				"<p><strong>Just One for My Buddy.</strong> <em>Apple ownership calls forth the enthusiastic brand loyalty once only associated with a particular make of automobile. " +
				"But concomitant with the explosion of products to support the Apple has come an acquisitiveness on the part of many users that threatens the future health of the industry. These owners either become, or trade with, software pirates. " +
				"<br> Starting by making copies for enthusiastic friends, some personal computer users move on to cranking out tens to hundreds of copies that they nonchalantly pass on to their friends' friends and mere acquaintances. " +
				"To those who buy their goods, software pirates are great money savers; to their victims in the industry, they're thieves.</em></p>" +
				// user groups
				"<p><strong>User Groups Under Fire.</strong> <em>Many manufacturers and retailers believe that user groups, at least those computer clubs whose members meet to swap information and programs with each other, are the most common perpetrators of unlawful copying. " +
				"When microcomputers were first introduced to the home, few were able to use them with a great deal of efficiency. " +
				"Because information and help were scarce, the best way for owners to learn more about their new investments was to meet and share ideas with other owners. " +
				"As computers gained popularity, user groups expanded in size and proliferated. Exchanging information and homemade programs was fine; " +
				"the problems arose when group members began trading commercial software as freely as they did their own.</em></p>" +
				// retailers
				"<p><strong>Piracy in the Retail Ranks.</strong> <em>Although the vast majority of retailers depend on software sales as much as computer sales to make their nut and would easily see the long-range consequences of ripping off their suppliers as disastrous, a few do not, and these few cause painful times for manufacturers. " +
				"Some dealers won't order a new product; they won't risk money on products they have to buy sight unseen, especially when, as .is the policy of most software companies, they have no recourse if they cannot sell what they purchase." +
				"<br>Instead, several retailers chip in and purchase one original from which they make copies for themselves. " +
				"The dealers who like the product after running their copies may decide to place orders. " +
				"But some dealers, even when they consider a program a winner, still won't purchase any for their stores. " +
				"What they might do is make and sell copies of their copies." +
				"<br>Lipson of Progressive thinks retailers are the major perpetrators of software piracy. " +
				"He refers to several retailers who never fail to order one copy of any new software product he produces. " +
				"But none of them ever reorders a program. <q>A customer on the brink of buying a system says he'll buy it if he can have this or that program with it. " +
				"Naturally, the retailer agrees, and the computer sale is made. But instead of taking financial responsibility for the plum and throwing in the program at his own expense, the retailer makes the customer a copy and retains the original.</q>" +
				"</em></p>",
		},
		{
			Title: "The first software crackers", Year: 1979, Highlight: true,
			Lead: "Disk copy protection hackers and crackers",
			Content: "<p>We have yet to learn who started <em>cracking</em>, when, or why, but it was certainly anonymous and probably born from curiosity " +
				"and for the technical challenge of <em>breaking</em> and <em>unlocking</em> protected software. " +
				"Yet cracking was also a response to the insertion of copy protection into software, likely first done on the Apple&nbsp;II.</p>" +
				// early examples
				"<p>Andrew McFadden wrote about early <a href=\"https://fadden.com/apple2/cassette-protect.html\">copy protection on software cassette tapes</a> in 1978 and 1979, but, they were unusual. " +
				"However, the July 1978 retail debut of the <a href=\"https://collections.museumsvictoria.com.au/articles/2787\">Disk II</a> floppy drive ecosystem with the first Apple " +
				"<a href=\"https://www.apple2history.org/history/ah14/#01\">Disk Operating System</a> was significant. " +
				// disk copy protection
				"It offered new benefits for software developers, including speed, reliability and complete control of the floppy drive hardware using custom software. " +
				"A critical mass of floppy drive owners with the new capabilities encouraged developers to use the media and embed novel <a href=\"https://www.bigmessowires.com/2015/08/27/apple-ii-copy-protection/\">disk copy protection methods</a> into their software intended for sale. " +
				"Interestingly, these ancient protection schemes are <a href=\"https://paleotronic.com/2024/01/28/confessions-of-a-disk-cracker-the-secrets-of-4am/\">still problematic</a> for computer historians today.</p>" +
				// roland cite
				"<p>Roland Gustafsson an early pioneer in creating disk copy protections answered a question about the discovery of the novel approach to using the disk drives. " +
				"<em>Initially by disassembling the Apple disk I/O routines and trying to figure out what they did. Also, quite significantly, I met Steve Wozniak after a San Francisco Apple Core Users Group meeting in a deli and he happened to be standing next to me in line waiting to order a sandwich, I picked his brain on how the disk mechanism worked. The brief 5 minutes of questioning there was enough for me to go and get started!</em></p>" +
				// jeffrey cite
				"<p>The December 1980 issue of Softalk magazine has Jeffrey Stanton <a href=\"http://underground2e.free.fr/Underground/Deplombage/Interviews/The_Wizard/Scans/Softalk_198012_Thief_p03.jpg\">commenting</a> on crackers, <br><em>" +
				"An interesting sidelight to the computer piracy game has resulting in people buying protected software for the challenge of breaking it. This concept may seem strange considering the price of software, but these people thrive on the most sophisticated protection schemes. To them, it is the ultimate \"adventure game.\"<br>" +
				"I've met many Apple owners who have spent much more time breaking a game disk than they ever spent playing the game. And a good portion of these people purchased that disk. In some cases, particularly among the more addicted experts, <strong>friends will gladly loan them any program in exchange for an unprotected copy that they can use for trading purposes</strong>. Hence, the danger of widespread trading or piracy of a disk doesn't always lie with the person who breaks the disk, but with their loss of control once their friends obtain a copy." +
				"</em></p>" +
				"<p>The October 1980 issue of Byte also reaffirms the existence of crackers, <em>While attempts to eliminate software piracy are commendable, they very often fail because of the cleverness of personal-computer users; many take the anticopy measures as a challenge. The problem lies in making the protection scheme easy enough to be affordable, but complex enough to work.</em></p>" +
				"<p>Jeffrey's portrayal of a loss of control could help to explain why some crackers started to inject their name or persona into their unprotected software in the form of " +
				"<a href=\"/image/milestone/tcommand.png\">digital graffiti</a> and filename disk hacks<F6>.</p>" +
				`<pre style="font-size:22px;" class="font-dos-mda reader reader-invert border border-black rounded-1 p-1">` +
				`]CATALOG<br><br>DISK VOLUME 254<br><br>` +
				`  T 012 SAVEGAME<br>` +
				`* S 000 ********************<br>` +
				`* S 000 * THIS UNPROTECTED *<br>` +
				`* S 000 * COPY PROVIDED BY *<br>` +
				`* S 000 * PIRATED SOFTWARE *<br>` +
				`* S 000 ********************<br>` +
				`  T 064 LIST1.MW<br>` +
				`  T 036 LIST1.MW<br>` +
				`* A 002 HELLO<br><br>]<span class="blinking">#</span>` +
				`</pre>`,
		},
		{
			Title: "The birth of wares", Year: 1980, Highlight: true,
			Lead: "It began on the Apple II microcomputer", Link: "http://artscene.textfiles.com/intros/APPLEII/", LinkTitle: "and browse the Apple II crack screens",
			Content: // kids with micros
			"<p>Without software, <a href=\"http://www.apple-iigs.info/doc/fichiers/Apple%20Price%20List%201978-08.pdf\">expensive</a> microcomputers of the era were mostly useless machines. " +
				"Getting them online with modems was also challenging.<sup><a href=\"#the-birth-of-warez-fn2\">[2]</a></sup><sup><a href=\"#the-birth-of-warez-fn5\">[5]</a></sup> " +
				"So understandably, the micro owners who were into computing would befriend fellow hobbyists, form communities, share information, and exchange software.</p>" +
				"<strong>How did this come about?</strong><br>" +
				// apple modems
				"<p>1979 saw the sale of the first Apple&nbsp;II <a href=\"https://www.apple2history.org/history/ah13/#09\">modem peripheral</a>, the Hayes&nbsp;Micromodem&nbsp;II and later, the Novation&nbsp;CAT. " +
				"These modems and the development of usable modem software such as ASCII&nbsp;Express in 1980, enabled Apple owners to connect to electronic message boards, communicate, and even exchange files remotely using the telephone.</p>" +
				// telephone costs
				"<p>One problem with telephones was that the expense of making calls outside the caller's local area was charged by the minute. " +
				"So, combining a slow microcomputer with an even slower modem on the phone network often led to a prohibitively costly phone bill. But " +
				"<a href=\"https://www.slate.com/articles/technology/the_spectator/2011/10/the_article_that_inspired_steve_jobs_secrets_of_the_little_blue_.html\">phone phreaking</a> had been a well-established, anti-corporate movement, " +
				" allowing callers to trick a phone company into misbilling or giving away expensive, long-distance phone calls.</p>" +
				// birth of warez
				"<p>So when was the birth of wares<sup><a href=\"#the-birth-of-warez-fn6\">[1]</a></sup> and a Warez scene? " +
				"There's no exact answer, but a good guess would be <strong>sometime&nbsp;in&nbsp;1980</strong> in the United States, maybe in Greater New York, Greater Boston, San Francisco Bay Area, or elsewhere. " +
				"By then, microcomputer owners exchanged details to meet in real life and online to duplicate and exchange software collections. And, importantly, to find ways to remove Apple II disk copy protections and show off the results. " +
				// warez dating
				"The pirates, also often identified as phone phreaks, removed or cracked disk copy protection on the Apple&nbsp;II and were dating their activity towards the end of 1980<sup><a href=\"#the-birth-of-warez-fn4\">[3]</a></sup> and in 1981. " +
				"Likewise, many modified, <q>cracked</q>, or <q>broken</q> ingame title screens exist for games published in those years.</p>" +
				// byter interview
				`<p>In a 1991 interview<sup><a href="#the-birth-of-warez-fn6">[6]</a></sup> for The Humble Review, Byter briefly talks about his early time on the Apple II scene he discovered in 1981. ` +
				`He states in those early Apple II days the boards were mostly message systems and occasional file transfer systems. However, the limited storage and slow modem speeds in those days meant most people chatted rather than pirated software. ` +
				`He goes on to confirm "In those days there wasn't any such thing as cracking groups... most everything which was cracked was credited solely to individuals."</p>` +
				// other platforms
				"<p>As for the other microcomputer platforms, the far more <a href=\"http://www.trs-80.org/was-the-trs-80-once-the-top-selling-computer/\">popular</a> " +
				"TRS-80 from Tandy had a <a href=\"http://www.trs-80.org/telephone-interface/\">modem peripheral</a> available at the end of 1978. " +
				"However, there is no evidence of an underground culture developing on the machine. A modem didn't sell for the " +
				"Atari&nbsp;400/800 <a href=\"http://www.atarimania.com/faq-atari-400-800-xl-xe-what-other-modems-can-i-use-with-my-atari_47.html\">until 1981</a>, " +
				"with its first dated cracks <a href=\"https://demozoo.org/productions/382174\">appearing in 1982</a>.</p>" +
				sect0 +
				"<div id=\"the-birth-of-warez-fn1\">[1] Warez was originally spelt with an <q>s</q> after the dictionary spelling.</div>" +
				"<div id=\"the-birth-of-warez-fn2\">[2] <a href=\"https://www.apple2history.org/history/ah18/#07\">VisiCalc</a>, the first useful <q>killer app</q>, was only published in the last few months of 1979.</div>" +
				"<div id=\"the-birth-of-warez-fn3\">[3] See, \"The earliest dated software crack and text art\"</div>" +
				"<div id=\"the-birth-of-warez-fn4\">[4] Crack screens with a Copyright 1980 and 1981 notice " +
				"<a href=\"http://artscene.textfiles.com/intros/APPLEII/tcommand.gif\">1</a>, " +
				"<a href=\"http://artscene.textfiles.com/intros/APPLEII/bezmanc.gif\">2</a>, " +
				"<a href=\"http://artscene.textfiles.com/intros/APPLEII/borgc.gif\">3</a>, " +
				"<a href=\"http://artscene.textfiles.com/intros/APPLEII/torax.gif\">4</a>.</div>" +
				"<div id=\"the-birth-of-warez-fn5\">[5] Early microcomputer peripherals' included software was often bare bones and only intended to confirm the hardware's operation. " +
				"New owners were expected to <a href=\"https://mirrors.apple2.org.za/ftp.apple.asimov.net/documentation/hardware/io/Hayes%20Micromodem%20II%20Manual.pdf\">program their own software</a> to use with their purchase.</div>" +
				"<div id=\"the-birth-of-warez-fn6\">[6] " +
				"<a href=\"https://defacto2.net/f/a56d0\">The Humble Review</a> issue #1, an interview with byter (1/2)." +
				div1 +
				sect1,
			Picture: Picture{
				Title:       "Tank Command - Kraked By Copy/Cat - No Rights Reserved",
				Alt:         "Tank Command kracked by screenshot on the Apple II",
				Png:         "tcommand.png",
				Attribution: "Jason Scott",
			},
		},
		{
			Title: "8BBS", Year: 1980, Month: 3, Day: 15, Highlight: true,
			Lead: "(408) 296-5799", LinkTitle: "the thousands of message logs", Link: "https://archive.org/details/8BBSArchiveP1V1/mode/1up",
			Content: "<p>Nearby San Jose, CA, <strong>8BBS#1</strong> <small>(eight-BBS number one)</small> came online in March 1980. It is one of the first electronic message boards" +
				" that early microcomputer hobbyists used, and is home to posts by some early hackers, pirates, and named-drop phreaker personalities of the era<sup><a href=\"#8bbs-fn1\">[1]</a></sup>. " +
				// message logs
				"But what stands out about the board today is that we have survived thousands of posts " +
				"from the earliest open online community that anyone in 1980 with the proper hardware could access from home—allowing for a more relaxed conversation that may not have been available in a work or academic environment. " +
				"These posts exist before Reddit, the web, Usenet, or the Internet.</p>" +
				`<pre style="font-size:22px;" class="font-dos-mda reader reader-invert border border-black rounded-1 p-1">` +
				"8BBS VER 5.5\n03-FEB-81 19:53:44\nPHONE: (408) 296-5799, 24 HOURS A DAY, EVERY DAY.\n" +
				"110, 150 & 300 BAUD SUPPORTED.\n* * * WELCOME TO BERNARD AND DICK'S\n* * * 8BBS#1 / SANTA CLARA, CA\n" +
				"* * * THE WORLD'S FIRST PDP8 BASED BULLETIN BOARD SYSTEM.\n* * * IN OPERATION SINCE MARCH 15, 1980</pre>" +
				"<p><a href=\"https://archive.org/details/8BBSArchiveP1V1/page/n30/mode/1up\">Message number 3964 from CHUCK HUBERT</a><br>To ALL at 12:52 on 20-Nov-80. Subject: CP/M BBS AND SOFTWARE EXCHANGE</p>" +
				"<p><a href=\"https://archive.org/details/8BBSArchiveP1V1/page/n43/mode/1up\">Message number 4177 from Kevin O'Hare</a><br>To SF (SAN FRANCISCO) PHREAKS at 23:54 on 28-Nov-80. Subject: HELP?</p>" +
				"<p><a href=\"https://archive.org/details/8BBSArchiveP1V1/page/n54/mode/1up\">Message number 4311 from Len Freedman</a><br>To RICK BYRNE at 11:02 on 02-Dec-80. Subject: PROG. TRADING</p>" +
				"<p><a href=\"https://archive.org/details/8BBSArchiveP1V1/page/n76/mode/1up\">Message number 4496 from Susan Thunder</a><br>To Keith Johnson at 03:39 on 07-Dec-80.<br><small>I HAVE BEEN A PHONE PHREAK FOR MANY YEARS AND I WOULD LOVE TO TRADE INFO WITH YOU!!</small></p>" +
				"<p><a href=\"https://archive.org/details/8BBSArchiveP1V1/page/n185/mode/1up\">Message number 7303 from DAVID LEE</a><br>To APPLE USERS at 16:51 on 15-Mar-81. Subject: APPLE SOFTWARE</p>" +
				"<p><a href=\"https://archive.org/details/8BBSArchiveP1V1/page/n197/mode/1up\">Message number 7434 from WALTER HORAT</a><br>To DAVID LEE at 22:22 on 18-Mar-81. Subject: SOFTWARE</p>" +
				"<p><a href=\"https://archive.org/details/8BBSArchiveP1V1/page/n259/mode/1up\">Message number 7853 from Sara Moore</a><br>To DAVID LEE at 05:08 on 02-Apr-81. Subject: SOFTWARE</p>" +
				"<ul>" +
				"<li><a href=\"http://www.flyingsnail.com/missingbbs/login-8BBS.html\">A login capture from 3-Feb-1981.</a></li>" +
				"<li><a href=\"http://www.flyingsnail.com/missingbbs/CHAT-8BBS.html\">Realtime text chat with the system operator.</a></li>" +
				"<li><a href=\"http://www.flyingsnail.com/missingbbs/6116.html\">The ridiculous costs of calling from long-distance.</a></li>" +
				"<li><a href=\"https://everything2.com/user/FTCnet/writeups/8BBS\">8BBS (thing) writeup from 2006.</a></li>" +
				"<li><a href=\"https://silent700.blogspot.com/2014/12/is-this-something.html\">tl;dr: I was given some old BBS session logs and I scanned them.</a></li>" +
				"</ul>" +
				sect0 +
				"<div id=\"8bbs-fn1\">[1] Phreaker personalities who mention 8BBS, " +
				"<a href=\"https://privacy-pc.com/articles/history-of-hacking-john-captain-crunch-drapers-perspective.html#Early_BBS_Days\">Captain&nbsp;Crunch</a>, " +
				"<a href=\"https://www.lysator.liu.se/etexts/hacker/digital1.html\">The&nbsp;Hacker&nbsp;Crackdown</a>, " +
				"<a href=\"http://phrack.org/issues/8/2.html\">TUC</a>, " +
				"<a href=\"http://phrack.org/issues/42/3.html\">Lord&nbsp;Digital</a>, " +
				"<a href=\"http://phrack.org/issues/10/2.html\">Dave&nbsp;Starr</a>, " +
				"<a href=\"https://www.theverge.com/c/22889425/susy-thunder-headley-hackers-phone-phreakers-claire-evans\">Susan&nbsp;Headley</a>. " +
				div1 +
				sect1,
		},
		{
			Title: "The first operating system for x86", Year: 1980, Month: 8,
			Lead: "Seattle Computer Products QDOS", LinkTitle: "about QDOS",
			Link: "https://www.1000bit.it/storia/perso/tim_paterson_e.asp",
			Content: "<p>Tim Paterson worked on a project at Seattle Computer Products to create an " +
				"8086 CPU <a href=\"http://www.s100computers.com/Hardware%20Folder/Seattle%20Computer%20Products/8086%20CPU%20Board/8086%20Board.htm\">plugin&nbsp;board</a> for the S-100 bus standard. " +
				"Needing an operating system for the 16-bit Intel CPU, he programmed a half-complete, unauthorized clone of the CP/M operating system within four months. " +
				"He called it <strong>QDOS</strong> (Quick and Dirty OS), and it sold few copies.</p>" +
				"<p>Initially, QDOS got bundled with an Intel&nbsp;8086 CPU and hardware <a href=\"http://www.s100computers.com/Hardware%20Folder/Seattle%20Computer%20Products/8086%20CPU%20Board/8086%20Board.htm\">package</a> for the S-100 bus. " +
				"But after poor sales, the OS was promptly renamed with the more business-friendly <a href=\"https://archive.org/details/bitsavers_seattleComanual1980_2120639/mode/2up\">86-DOS</a>.</p>" +

				`<pre style="font-size:1.5em;line-height:1em;" class="font-dos-mda reader reader-invert border border-black rounded-1 p-1">` +
				"<br><p>SCP 8086 Monitor 1.5<br>>B<br>☺︎<br>" +
				"86-DOS version 1.00<br>" +
				"Copyright 1980,81 Seattle Computer Products, Inc.<br>" +
				"Enter today's date (m-d-y): 01-01-80<br>" +
				"<br>COMMAND v. 1.00<br>" +
				"<br>A:chkdsk a:<br>" +
				"              19 disk files<br>" +
				"          245760 bytes total disk space<br>" +
				"          146944 bytes remain available<br>" +
				"<br>               0 bytes total system RAM<br>" +
				"         1036448 bytes free<br>" +
				"<br>A:edlin news.doc<br>" +
				"<br>EDLIN version 1.00<br>" +
				`End of input file<br>*<span class="blinking">_</span><br>` +
				`</p></pre>`,
		},
		{
			Title: "Motorola 68000 16-bit CPU", Year: 1980, Month: 11,
			Lead: "", LinkTitle: "about the 68000", Link: "https://spectrum.ieee.org/chip-hall-of-fame-motorola-mc68000-microprocessor",
			Content: "<p>Available in November 1980, the famed <strong>Motorola 68000</strong> is the 16-bit successor to the 8-bit 6800 CPU from late 1974. " +
				"The Motorola series competed and operated in parallel with the incompatible Intel chips for the burgeoning microprocessor market. " +
				"And like Intel, Motorola found its 8-bit chip designs reversed-engineered, enhanced, and undercut by its other competitors.</p>" +
				"<p>But the 68000 was the 16-bit chip of the 1980s, powering everything from the Sega <a href=\"https://www.lifewire.com/history-of-sega-genesis-dawn-729670\">Megadrive/Genesis</a>, the Sega 16, the SNK NeoGeo, and various arcade games.</p>" +
				"<p>Significantly, it was at the heart of a future generation personal computing platforms, the Apple&nbsp;Lisa&nbsp;<sup>1983</sup>, <a href=\"https://spectrum.ieee.org/apple-macintosh\">Apple&nbsp;Macintosh</a>&nbsp;<sup>1984</sup>, " +
				"<a href=\"https://daily.redbullmusicacademy.com/2017/10/atari-st-instrumental-instruments\">Atari&nbsp;ST</a>&nbsp;<sup>1985</sup>, and the Commodore&nbsp;<a href=\"https://arstechnica.com/gadgets/2007/07/a-history-of-the-amiga-part-1/\">Amiga&nbsp;1000</a>&nbsp;<sup>1985</sup>. " +
				"These incompatible systems offered high-resolution graphics and their own mouse-driven <abbr title=\"Graphical User Interface\">GUI</abbr> operating system as standard. " +
				"At a time when the typical microcomputer or PC relied upon dated, user-hostile text interaction." +
				"<br><span title=\"The common input prompt for an IBM or Microsoft disk operating system\"><strong>A>_</strong></span></p>",
		},
		{
			Title: "The earliest dated software crack and text art", Year: 1980, Month: 11,
			Lead:      "So far, Cyber Strike broken by The Tornado",
			LinkTitle: "about and emulate the crack", Link: "https://archive.org/details/Sabotage_Reversal_Cyber_Strike_Dungeons",
			Content: // dung beetles
			"<p>The earliest-dated crack is probably on the Apple&nbsp;II. A likely example is " +
				"<q><a href=\"https://www.ebay.com/itm/204747521812\">Cyber Strike</a> broken by The Tornado</q> in <strong>November 1980</strong>. " +
				"The static crack credit and text art is loaded at the start of the game before the game's title screen.</p>" +
				"<p>The game is authored by Nasir Gebelli and published by Sirius Software, a company formed in 1980 and known for their disk copy protections. " +
				"The game also entered the Softalk Bestsellers <u>November 1980</u> charts at position 6, meaning the game likely went on sale in October or November.</p>" +
				"Other dated cracks include" +
				ul0 +
				"<li><a href=\"https://demozoo.org/productions/381802\">Pulsar II / Worm Wall</a> <small>1981, Sirius Software for Apple II <q>Sliced by -The Razor- April 1981</q></small></li>" +
				"<li><a href=\"https://demozoo.org/productions/147055\">Crush Crumble &amp; Chomp</a> <small>1981, Automated Simulations for Apple II <q>Broken by The Pirate 09/26/81</q></small></li>" +
				"<li><a href=\"https://demozoo.org/productions/382174\">Submarine Commander</a> <small>1982, Thorn EMI for Atari 400/800, <q>Cracked 1982 by The Code Cracker</q></small></li>" +
				"<li><a href=\"https://demozoo.org/productions/382162\">Alien Swarm</a> <small>1982, Inhome Software for Atari 400/800, <q>Copyright Disks Ahoy 1982</q></small></li>" +
				"<li><a href=\"https://demozoo.org/productions/380720\">Dung Beatles</a> <small>1982, Datasoft for Apple II, <q>Broken by Black Bart March 1982</q></small></li>" +
				"<li><a href=\"https://demozoo.org/productions/381276\">Apple World</a> <small>1980, United Software for Apple II, <q>(c) cracked 1982 by The Mulcher ][</q></small></li>" +
				"<li><a href=\"https://demozoo.org/productions/380878\">Flockland Island Crisis</a> <small>1982, Vital Information for Apple II, <q>cracked (c) 1982 by mr. krac-man</q></small></li>" +
				"<li><a href=\"https://demozoo.org/productions/290088\">Type Attack</a> <small>1982, Sirius Software for Apple II, <q>(B)1982 Broken by Krakowicz NY</q></small></li>" +
				"<li><a href=\"https://demozoo.org/productions/382776\">Hard Hat Mack</a> <small>Oct 1983, EA for Commodore 64, <q>cracked AD 1983 by Oleander</q></small></li>" +
				"<li><a href=\"https://demozoo.org/productions/383213\">Space Sentinel</a> <small>1983, T&F for Commodore 64, <q>broken by mike freeze 830915</q></small></li>" +
				ul1,
			Picture: Picture{
				Title: "Cyber Strike broken by The Tornado - Nov '80",
				Alt:   "Cyber Strike broken by screenshot on the Apple II",
				Png:   "cyber_strike_the_tornato.png",
				Webp:  "cyber_strike_the_tornato.webp",
			},
		},
		{
			Title: "Computer Software Copyright Act", Year: 1980, Month: 12, Day: 12, Highlight: true,
			Lead: "Software is defined in US copyright laws", LinkTitle: "about the act",
			Link: "https://www.c2st.org/the-computer-software-copyright-act-of-1980/",
			Content: "<p>Signed as an amendment to law by President Jimmy Carter, computer programs are defined by copyright law and enable authors to control the copying, selling, and leasing of their software.</p>" +
				"<p>But the law was <a href=\"https://repository.law.uic.edu/cgi/viewcontent.cgi?article=1571&context=jitpl\">confusing</a> as software documentation and software source code are protected, but the object code or the compiled software that ran on the computer hardware is probably not.</p>" +
				"<p>The screenshot shows a heavy-handed copyright 1979 notice for an Apple Computer published game. It is missing the notices that software copyright infringement is illegal and criminal. In the day, Apple could only threaten to sue for civil damages; however, even that is mostly scaremongering.</p>" +
				"<p>The <a href=\"https://www.mobygames.com/game/125083/apple-bowl/cover/group-243846/cover-621747/\">game was sold</a> on an audio cassette tape, making it feasible for a radio or TV station to broadcast the software over the air for duplication, so Apple included the \"duplicated or transmitted\" assertion.</p>",
			Picture: Picture{
				Title: "© 1979 notice for an Apple Computer published game",
				Alt:   "A screenshot of the copyright notice from the game Apple Bowl",
				Png:   "1979-apple-bowl-copyright.png",
				Webp:  "1979-apple-bowl-copyright.webp",
			},
		},
		{
			Title: "The first groups", Year: 1981, Highlight: true,
			Lead: "Possibly late 1981, but probably 1982",
			Content: "<p>Discussions of cracking groups from the Apple II era often claim they were around in 1980. " +
				"However, of the cracks that survive today, the ones by cracking groups are for games that got published for Christmas 1981 and in <strong>1982</strong>. " +
				"While there are many 1980 and 1981 cracks with authorship, these were released by individual crackers rather than in a collaboration as part of a cracking group.</p>" +
				"<p>Some of the famous, <q>first</q> cracking groups from the Apple II era are, " +
				"<a href=\"https://demozoo.org/groups/153053\">Super Pirates of Minneapolis</a>, " +
				"<a href=\"https://demozoo.org/groups/61767\">The Apple Mafia</a>, " +
				"The Software Pirates, <a href=\"https://demozoo.org/groups/61754\">Digital Gang</a>, " +
				"<a href=\"https://demozoo.org/groups/153120\">The Dirty Dozen</a>, " +
				"<a href=\"https://demozoo.org/groups/153093\">The Untouchables</a>, " +
				"and <a href=\"https://demozoo.org/groups/153047\">Apple Pirated Program Library Exchange</a> <small>aka A.P.P.L.E.</small>." +
				"</p>" +
				// byter interview
				`<p>In a 1991 interview<sup><a href="#the-first-group-fn1">[1]</a></sup> for The Humble Review, Byter talks about the early Apple II scene. He confirms, "In those days [a decade ago] there wasn't any such thing as cracking groups... most everything which was cracked was credited solely to individuals."` +
				` He continues, "As for cracking groups, they're changed as well. Apple ][ cracking groups (when they weren't simply individuals), were always small. Only members essential to the groups activities were members. This included (at times) a leader, a cracker or two and sometimes an artist and a programmer. ` +
				`It was rare for a group to have more than five members. Suppliers were never part of the group, nor were sysops or boards."</p>` +
				// the apple marfia story
				"<p><strong>The Apple Mafia</strong>, <strong>The Untouchables</strong>, <strong>The Dirty Dozen</strong><br>" +
				"In 1986, Red Ghost posted <a href=\"/f/a430f7\">The Apple Mafia Story</a>, claiming " +
				"The&nbsp;Untouchables, The&nbsp;Apple&nbsp;Mafia, and&nbsp;The&nbsp;Dirty&nbsp;Dozen " +
				"were some of the first-ever pirate groups. But he admits he wasn't there and wasn't even into computers then. He grew up in Queens, New York, and suggests that is where many original phreakers and pirates originated. " +
				"But that is debatable, as he was probably unaware that phone freaking was a <a href=\"http://www.flyingsnail.com/images/YIPL/YIPL_002.jpg\">nationwide</a> activity in the 1960s and 1970s. " +
				"The YIPL July 1971 newsletter wrote, <em>Blue Box is linked to phone call fraud - After interviewing engineering students <u>around the country</u>, I found that the blue box...</em>. " +
				"<br>And of the pirate groups mentioned, they only show cracks for games from 1982 and 1983.</p>" +
				// godfather quote
				"<p><strong><q>A Brief History of the Apple Mafia</q></strong><br>" +
				"In the named post from late 1983 or early 1984, The Godfather states he founded The Apple Mafia in 1980, first as a joke, then as a serious project in 1981. " +
				"Maybe the name was used for phone phreaking and later shifted towards piracy? Or maybe he was suffering from some memory bias? " +
				"<em style=\"text-transform: lowercase;\">BRIEF HISTORY OF THE APPLE MAFIA. FOUNDED IN 1980 BY THE GODFATHER AS A JOKE. REDONE IN 1981 AS A SEMI SERIOUS GROUP. " +
				"KICKED SOME ASS IN '82. BLEW EVERYONE AWAY IN 83, AND WILL DO MUCH BETTER IN 84. ..." +
				"IS CURRENTLY THE OLDEST <u>ACTIVE</u> GROUP, NEXT (OF PEOPLE WHO WOULD STILL BE AROUND) ARE " +
				"<a href=\"https://demozoo.org/groups/118450\">THE WARE LORDS</a> ('83 I BEILIEVE) AND " +
				"<a href=\"https://demozoo.org/groups/115539\">THE 1200 CLUB</a> ('83 ALSO, I THINK). THAT'S IT.</em></p>" +
				// phrack magazine quote
				"<p><strong>The Apple Mafia, <q>the first WAreZ gRoUP</q></strong><br>" +
				"Phrack Magazine issue 42 has a 1993 <a href=\"http://phrack.org/issues/42/3.html\">interview</a> with hacker and former Apple pirate <a href=\"https://en.wikipedia.org/wiki/Patrick_K._Kroupa\">Lord&nbsp;Digital</a>. " +
				"The interview claims around 1980, he and some New York friends traveled to the AppleFest conference, discovered some other Apple owners. And afterwards, formed The Apple Mafia to make it the first warez group for the Apple II. " +
				"However, the story is inaccurate, as AppleFest was first held on June 1981<sup><a href=\"#the-first-group-fn3\">[3]</a></sup>, in Boston. " +
				"<q>I played around with various things, ... until " +
				"I got an Apple&nbsp;II+ in 1978. I hung out with a group of people who were also " +
				"starting to get into computers, most of them comprising the main attendees of " +
				"the soon-to-be-defunct TAP<sup><a href=\"#the-first-group-fn2\">[2]</a></sup> meetings in NYC... " +
				"Around 1980 there was an Apple Fest that we went to, and found even more people with Apples and, from this, formed the " +
				"Apple Mafia, which was, in our minds, really cool sounding and actually became the first WAreZ gRoUP to exist for the Apple&nbsp;II.</q>" +
				"<p></p>" +
				"Given the inconsistencies about The Apple Mafia, a guess would be they formed in the second half of 1981, post the first AppleFest conference. " +
				"But, probably as a phone or computer phreaking clique that later got into cracking software on the Apple II, after being inspired by <a href=\"/image/milestone/cyber_strike_the_tornato.webp\">others</a>. " +
				"Did this switch make them the <q>first</q> cracking group, who knows?" +
				"</p>" +
				// super pirates
				"<p><strong>Super Pirates of Minneapolis</strong>" +
				"<br>The Super Pirates were a famous, early group from outside of New York. " +
				"A claim suggests the Super Pirates were around in 1980, the same year the game <a href=\"https://www.mobygames.com/game/47942/cyber-strike\">Cyber&nbsp;Strike</a> from Sirius Software was published (in the forth quarter). " +
				"However, associating Super Pirates with this year should be viewed with skepticism, as their <a href=\"https://demozoo.org/groups/153053/\">known cracks</a> are for games with <a href=\"https://www.mobygames.com/game/17995/horizon-v/screenshots/apple2/113190/\">&copy;1982</a>.<br>" +
				"<em>The 1st ware I got was back in 1980. <a href=\"https://demozoo.org/productions/380718\">It was Cyber Strike</a>. Along with about 35 other disks, most cracked by the Super Pirates!</em> " +
				"says <a href=\"https://demozoo.org/sceners/153128\">The Incognito</a> in the <q>Pirate History</q> repost found on the <a href=\"https://demozoo.org/bbs/240\">Red Sector A BBS</a> <small>(313) 591-1024</small> and also in the <a href=\"http://www.textfiles.com/bbs/boardsims2.txt\">Board Simulations 2</a> text from 1987.</p>" +
				// midwest guild
				"<p>Anecdotal evidence suggests the Super Pirates were involved in the first-ever BBS bust. The members left to form or joined the <a href=\"https://demozoo.org/groups/86223\">Midwest Pirate's Guild</a>, " +
				"a group strongly associated with the cracker <a href=\"https://demozoo.org/sceners/118462\">Apple Bandit</a> and his Minneapolis-based board, <a href=\"https://demozoo.org/bbs/359\">The&nbsp;Safehouse</a>&nbsp;<small>(612) 724-7066</small>.</p>" +
				sect0 +
				"<div id=\"the-first-group-fn1\">[1] " +
				"<a href=\"https://defacto2.net/f/a56d0\">The Humble Review</a> issue #1, an interview with byter (1/2)." +
				"<div id=\"the-first-group-fn2\">[2] <a href=\"http://www.flyingsnail.com/missingbbs/tap01.html\">TAP</a> was formerly named as " +
				"The <a href=\"https://archive.org/details/yipltap/YIPL_and_TAP_Issues_1-91.99-100/page/n165/mode/2up\">Youth International Party Line</a> (YIPL).</div>" +
				"<div id=\"the-first-group-fn3\">[3] <q>For the first time ever, a computer show devoted exclusively to the Apple computers. Applefest '81</q> advert in the <a href=\"https://www.wap.org/journal/showcase/washingtonapplepijournal1981v3no4apr81.pdf\">April 1981 issue of Washington Apple Pi</a>.</div>" +
				div1 +
				sect1,
		},
		{
			Title: "The earliest cracktros", Year: 1981, Month: 4, Highlight: true,
			Lead: "Mr. Xerox's Starblaster, and Sliced by -The Razor-",
			Content: "<p><strong>Cracktros</strong> and crack intros are programmed and animated vanity title screens " +
				"that gives credit to the removal of disk copy protection schemes." +
				"</p><p>" +
				// apple ii
				"It is challenging to place early pirated releases for the Apple&nbsp;II, Atari, or PC. " +
				"Many early crackers didn't date their releases, and the systems themselves didn't track time or stamp the files. " +
				"But given the <a href=\"http://artscene.textfiles.com/intros/APPLEII/.thumbs.html\">proliferation</a> of <q>broken</q> and <q>cracked</q> by texts injected into Apple&nbsp;II software during 1980, 1981, and 1982, " +
				"it can be assumed the early cracktro evolved on this system." +
				"</p><p>" +
				// mr xerox
				"<strong>Candidates</strong><br>" +
				"The prolific Apple II cracker <strong><a href=\"https://demozoo.org/sceners/153043/\">Mr. Xerox</a></strong> probably created one of the <u>first proper crack-intros</u> and uses a vertical and horizontal scroller in his animated cracked-by " +
				"<a href=\"https://archive.org/details/a2_Starblaster_19xx_C_G_cr_Star_Trek_1983_Sega_cr_Shuttle_Intercept_19xx__cr\">introduction</a> for StarBlaster. " +
				"When compared to the <a href=\"https://archive.org/details/Starblaster4amCrack\">startup</a> of the original game, the Mr. Xerox animation clearly involved additional code injected by the cracker. Confusing, there were at least two games named StarBlaster. " +
				"The game cracked by Xerox is from Piccadilly Software and has &copy;1981 but was announced as available at retail by <strong>June 1982</strong> in the " +
				"<a href=\"https://archive.org/details/softalkv2n10jun1982/page/n101/mode/2up?q=%22Star+Blaster%22\">computing press</a>." +
				"</p><p>" +
				// the copycatter
				"The Apple II cracker <strong><a href=\"https://demozoo.org/sceners/153624/\">The Copycatter</a></strong> may have vibed the first <u>horizontal scroller</u> found in a " +
				"<a href=\"https://archive.org/details/a2_Pro_Football_The_Gold_Edition_1982_System_Design_Lab_cr_Copycatter\">release</a> of Pro&nbsp;Football The Gold Edition, " +
				"however, it is not a true crack-intro, but a text edit using a hex editor. " +
				"Pro Football was not a game, but an expensive application to predict the results of Gridiron football matches. " +
				"The Gold Edition update was announced and <a href=\"https://archive.org/details/softalkv2n11jul1982/page/126/mode/2up?q=%22pro+football%22\">advertised</a> in Softalk <strong>June 1982</strong>. " +
				"The crack scrolls the following message, <br><em>BROKEN BY -\\[THE COPYCATTER]/-  THANKS STOSH</em>." +
				"</p><p>" +
				// the razor
				"So far, a two-frame animated <a href=\"https://demozoo.org/productions/381802/\">loader</a> for a pair of Apple II games is the <u>earliest dated intro</u> known. " +
				"Credited to <strong><a href=\"https://demozoo.org/sceners/153666/\">-THE RAZOR-</a></strong> and dated to <strong>April 1981</strong>, the intro was used for the games Pulsar II and Worm Wall, both of which were sold as a single floppy at retail. " +
				"-The Razor- <q>sliced</q> and likely cracked the games as two separate releases. " +
				"The games are authored by Nasir Gebelli and published by Sirius Software, both were known to do early copy protection, also making this is a possible crack-intro. " +
				"</p>",
			Picture: Picture{
				Title: "Mr. Xerox's Star Blaster cracktro",
				Alt:   "Mr. Xerox's Star Blaster cracktro on the Apple II",
				Png:   "starblaster-mr-xerox.png",
				Webp:  "starblaster-mr-xerox.webp",
				Webm:  "starblaster-mr-xerox.webm",
			},
		},
		{
			Title: "Atari's Graphics/Sound Demonstration", Year: 1981, Month: 5,
			Link:      "https://www.atarimania.com/8bit/files/APX_Graphics_Sound_Demonstration.pdf",
			LinkTitle: "the Graphics/Sound Demonstration manual",
			Content: "Under its <a href=\"https://archive.org/details/APXCatalogWinter1981/page/n41/mode/2up?view=theater\">Atari Program Exchange</a> (APX) label, " +
				"Atari publishes the Graphics/Sound Demonstration, a mail order title containing a diskette and manual with instructions on running several graphic and sound demonstrations on the Atari 400/800 line of computers. " +
				"The disk also includes the assembly and BASIC source codes, allowing programmers and hobbyists to adapt these vanity effects in their software.",
			Picture: Picture{
				Title: "Graphics/Sound Demonstration catalog page",
				Alt:   "A photo snippet of the 1981, Atari Program Exchange catalog page for the Graphics/Sound Demonstration.",
				Jpg:   "atari-graphics-sound-demonstration.png",
				Avif:  "atari-graphics-sound-demonstration.avif",
			},
		},
		{
			Title: "The first PC", Year: 1981, Month: 8, Day: 12, Highlight: true,
			Lead: "IBM Personal Computer", LinkTitle: "about the IBM PC",
			Link:    "https://www.ibm.com/ibm/history/exhibits/pc25/pc25_birth.html",
			Content: "Built on the 4.77 MHz <strong>Intel&nbsp;8088</strong> microprocessor, 16KB of RAM and Microsoft's PC-DOS, this expensive and underpowered machine heralds the <u><strong>PC platform</strong></u>.",
			Picture: Picture{
				Title:       "IBM PC 5150",
				Alt:         "A photo of the IBM PC 5150",
				Jpg:         "ibm-pc-5150.jpg",
				Avif:        "ibm-pc-5150.avif",
				Attribution: "Rama & Musée Bolo",
				License:     "CC BY-SA 2.0 FR",
				LicenseLink: "https://creativecommons.org/licenses/by-sa/2.0/fr/deed.en",
			},
		},
		{
			Title: "The first published PC game", Year: 1981, Month: 9,
			Lead: "Microsoft Adventure from IBM", LinkTitle: "about Microsoft Adventure",
			Link: "https://www.filfre.net/2011/07/microsoft-adventure/",
			Content: "<p><strong>Microsoft Adventure</strong> is an IBM&nbsp;PC port of the text game <em>Colossal Cave Adventure</em>.</p>" +
				"<p>Adventure was a highly influential and popular text-only adventuring game of exploration and puzzle solving for mainframe computers of the 1970s. " +
				"Will Crowther wrote it in FORTRAN for the PDP-10 system and Don Woods at the Stanford AI Lab in California later expanded it. " +
				"The game created the interactive fiction genre, which later led to graphic adventures and story narratives in video games.</p>",
			Picture: Picture{
				Title:       "IBM Microsoft Adventure",
				Alt:         "A photo of the 1981, Microsoft Adventure floppy disk media.",
				Jpg:         "ibm-microsoft-adventure.jpg",
				Avif:        "ibm-microsoft-adventure.avif",
				Attribution: "Jack Lightbeard & MobyGames",
				License:     "© MobyGames",
				LicenseLink: "https://www.mobygames.com/game/4074/microsoft-adventure/cover/group-3242/cover-176506/",
			},
		},
		{
			Title: "The first demo", Year: 1981, Month: 12, Highlight: true,
			Lead:      "So far, Merry Christmas by CB'81, on the Atari 400/800",
			LinkTitle: "the Demozoo entry with a YouTube link",
			Link:      "https://demozoo.org/productions/144652/",
			Content: "<p>The earliest known demo or demonstration program is probably this great but untitled animated Christmas greeting created on the Atari 400 or 800 and signed as <q>CB'81</q>. " +
				"CB is believed to be Claus Buchholz, a known <a href=\"https://archive.org/details/Atari40048KUpgrade/mode/2up\">hardware hacker</a> for the platform. " +
				"We presume this demo software got shared on Atari-centric bulletin boards in the USA around late 81.</p>" +
				"<p>Earlier demonstration software existed for various machines, including " +
				"1978's <a href=\"https://demozoo.org/productions/121614/\">Apple&nbsp;Vision</a>, " +
				"1979's <a href=\"https://demozoo.org/productions/151537/\">Dancing&nbsp;Demon</a> on the TRS-80, and " +
				"1980's <a href=\"https://demozoo.org/productions/98550/\">Atari In-Store Demonstration Program</a>. " +
				"However, these were commercials created by Apple, Radio Shack, or Atari employees and designed to demonstrate the machines' capabilities in a retail store.</p>" +
				"<p>The untitled Christmas greeting by CB is the earliest known demonstration software created by a hobbyist with no commercial intent.</p>",
			Picture: Picture{
				Title: "The untitled Christmas greeting by CB",
				Alt:   "A screenshot of the 1981 Christmas greeting, on the Atari 400/800.",
				Png:   "cb-81.png",
			},
		},
		{
			Title: "MS-DOS", Year: 1982, Month: 8,
			Lead: "MicroSoft Disk Operating System v1.25", LinkTitle: "about MS-DOS 1 and 1.25",
			Link: "https://www.os2museum.com/wp/dos/dos-1-0-and-1-1/",
			Content: "Microsoft releases the first edition of <strong>MS-DOS</strong> v1.25, <a href=\"https://www.os2museum.com/wp/dos/dos-1-0-and-1-1/msdos-ad-1982/\">readily available</a> to all OEM computer manufacturers. " +
				"Prior releases were exclusive to IBM. The next release, MS-DOS 2, is also sold boxed at retail and will help Microsoft to become the de facto operating system provider for personal computers." +
				"<p>In 2014, the Computer History Museum published the <a href=\"https://computerhistory.org/blog/microsoft-ms-dos-early-source-code/\">source code</a> for this operating system edition, and Microsoft later made a GitHub <a href=\"https://github.com/microsoft/MS-DOS\">repository</a>.</p>",
			Picture: Picture{
				Title:       "Compaq's MS-DOS based on MS-DOS v1.25",
				Alt:         "MS-DOS Version 1.12 for the Compaq Personal Computer.",
				Jpg:         "ms-dos-floppy-disks.jpg",
				Avif:        "ms-dos-floppy-disks.avif",
				Attribution: "Brian R. Lueck",
				License:     "public domain",
				LicenseLink: "https://en.wikipedia.org/wiki/MS-DOS#/media/File:Compaq_mddos_ver1-12.jpg",
			},
		},
		{
			Title: "Third-party PC games", Year: 1982,
			Content: "<p>The first set of published games on the PC platform is sold without IBM's involvement.</p>" +
				"Some early publishers include" +
				ul0 +
				"<li><a href=\"//s3data.computerhistory.org/brochures/broderbund.software.1982.102646180.pdf\">Brøderbund</a> was one of the major publishers of the Apple II.</li>" +
				"<li><a href=\"//archive.org/details/avalon-hill-game-company-catal-fall-1982\">The Avalon Hill Game Company</a> is the famed war and strategic board game publisher. </li>" +
				"<li><a href=\"//archive.org/details/strategic-simulations-inc-summer-1982-catalog/mode/2up\">Strategic Simulations</a>, Inc. acquired the Dungeons and Dragons computer game license and became a pioneer of the CRPG genre.</li>" +
				"<li><a href=\"//www.uvlist.net/companies/info/1023-Windmill+Software\">Windmill Software</a> was one of the first developers to create games exclusively on the PC.</li>" +
				"<li><a href=\"//retro365.blog/2019/09/23/bits-from-my-personal-collection-the-original-ibm-pc-and-orion-software/\">Orion Software</a> created some of the earliest games on the PC.</li>" +
				"<li><a href=\"//www.uvlist.net/companies/info/1029-Spinnaker+Software\">Spinnaker Software</a>" +
				ul1 +
				"<p>The following year saw some major arcade and video game publishers release software on the PC. Despite the business-centric marketing of the platform, game software sold on a floppy disk was a popular seller. " +
				"For publishers, it is less risky than manufacturing the expensive cartridges required by some other game systems.</p>" +
				ul0 +
				"<li><a href=\"//dfarq.homeip.net/atarisoft-if-you-cant-beat-em-join-em/\">Atarisoft</a> was the publishing arm of the computer, console, and arcade game maker.</li>" +
				"<li><a href=\"//www.uvlist.net/companies/info/243-Infocom\">Infocom</a> founded by the Massachusetts Institute of Technology staff and students to create story narrative games.</li>" +
				"<li><a href=\"//www.resetera.com/threads/lets-look-back-at-game-company-datasoft.587093/##post-87110411\">Datasoft</a> created licensed film, television assets, and arcade ports.</li>" +
				"<li><a href=\"//www.uvlist.net/companies/info/83-Mattel%20Electronics\">Mattel</a> was the creator of the Intellivision console and numerous games.</li>" +
				"<li><a href=\"//www.wired.com/story/sierra-online-ken-williams-interview-memoir/\">Sierra On-Line</a> became one of the biggest PC publishers of the 1980s and the flag-barrier of the graphic adventure genre.</li>" +
				ul1,
		},
		{
			Title: "The great online reboot - the birth of an inter-network", Year: 1983, Month: 1, Day: 1,
			Lead: "APRA Internet", LinkTitle: "the Notable computer networks", Link: "https://dl.acm.org/doi/pdf/10.1145/6617.6618",
			Content: "On January 1, 1983, the US Department of Defense coordinated the massive shutdown of its existing experimental wide-area network, <abbr title=\"Advanced Research Projects Agency Network\">ARPAnet</abbr>. " +
				"Referred to as <q>Flag Day,</q> the event required all systems associated with the US military network to reconnect using a new <abbr title=\"Transfer Control Protocol\">TCP</abbr>/<abbr title=\"Internetwork Protocol\">IP</abbr> protocol. " +
				"The replacement protocol decentralized the network's operations and is somewhat inspired by the earlier French " +
				"<a href=\"https://www.inria.fr/en/between-stanford-and-cyclades-transatlantic-perspective-creation-internet\">CYCLADES</a> packet-switch network. " +
				"By demanding that the connected hosts handle data delivery and error correction, connecting various academic, research and commercial computer networks is possible, removing ARPAnet's excessive expense and inability to scale.</p>" +
				"<p>Later in the year, due to a <a href=\"https://www.washingtonpost.com/archive/business/1983/10/04/big-computer-network-split-by-pentagon/d12feaba-c0c7-45fb-a851-25267f8dca9c/\">fear of civilian hackers</a>, the systems associated with the US military were to disconnect again and join a new isolated Defense Data Network (MILnet). The few remaining non-military systems that adopted the TCP/IP protocol standard formed the basis of the new ARPA internetwork or APRA Internet.</p>" +
				"<p>The other alternative networks of the era:</p>" +
				ul0 +
				"<li><abbr title=\"Because It's Time NETwork\">BITNet</abbr> <sup>1981</sup>, a cross-continental, research center and university network for file transfers and messaging." +
				"<li><abbr title=\"European Unix Network\">EUnet</abbr> <sup>1982</sup>, the first public wide area network of Europe.</li>" +
				"<li>Janet <sup>1984</sup>, an extensive UK academic network.</li>" +
				"<li>Corporate networks from Xerox Internet, DEC Easynet and IBM VNET.</li>" +
				ul1,
		},
		{
			Title: "Commercial Disk Copy Protections", Year: 1983, Month: 1,
			Lead: "PROLOK and the end of software piracy", LinkTitle: "the PROLOK deep-dive for MartyPC",
			Link: "https://martypc.blogspot.com/2024/09/pc-floppy-copy-protection-vault-prolok.html",
			Content: "<p>" +
				"PROLOK, the first <a href=\"https://archive.org/details/byte-magazine-1984-10-rescan/mode/2up?q=prolok\">heavily marketed</a> " +
				"disk copy protection ecosystem for Apple and PCs, was shown at the CP/M '83 conference held in January 1983 in San Francisco. CP/M by Digital Research was a major PC operating system predominantly used by businesses." +
				"</p><p>" +
				"Creative Computing <a href=\"https://archive.org/details/CreativeComputingbetterScan198308/page/n199/mode/2up?q=prolok\">reports</a>, " +
				"<em>Urban Pacific Data Service came out with Prolok, which they say will <q>all but eliminate piracy</q>. Here's how it works. Software producers and others buy Prolock disks, which have a built-in fingerprint, that is, a series of random program encryptions <q>and other devious programming techniques,</q> which protect the program.</em>" +
				"</p><p>" +
				"There's no confirmation of this, but it seems likely Urban Pacific Data Service was or became Vault Corporation, and in May 1983, filed the " +
				"<a href=\"https://tsdr.uspto.gov/#caseNumber=73425657&caseSearchType=US_APPLICATION&caseType=DEFAULT&searchType=statusSearch\">trademark</a> for Prolok. " +
				"The ecosystem was heavily advertised in the tech press, often with the bold claims of <q>the end of software piracy</q>, and became very popular, with thousands of customers at its peak. However, at the end of 1984, the reputation of Prolok was " +
				"<a href=\"https://archive.org/details/PC-Mag-1985-01-22/mode/2up?q=prolok\">destroyed</a> after the company began promoting the idea of a <q>Plus</q> update to the tool, which enabled malware-like behavior." +
				"</p><p>" +
				"Other tools, duplication services and protections from the era," +
				ul0 +
				"<li><a href=\"https://martypc.blogspot.com/2024/08/pc-floppy-copy-protection-formaster.html\">Copy-Lock</a> by Formaster <small>for Apple, Commodore, IBM PC</small></li>" +
				"<li><a href=\"https://martypc.blogspot.com/2024/08/pc-floppy-copy-protection-softguard.html\">SUPERLoK</a> by Softguard Systems <small>used by Lotus, Ashton Tate, Sierra On-line</small></il>" +
				"<li><a href=\"https://martypc.blogspot.com/2024/10/pc-floppy-copy-protection-xemag-xelok.html\">Xelok</a> by XEMAG <small>for Apple, Commodore, IBM PC</small></li>" +
				"<li><a href=\"https://archive.org/details/PC_Tech_Journal_vol01_n05/page/n75/mode/2up?q=%22SECURE-WARE%22\">SECURE-WARE</a> by Remote Systems Inc.</li>" +
				"<li>COPYLOCK by Export Software International (UK)</li>" +
				"<li>Software Protection Device by <a href=\"https://archive.org/details/PC_Tech_Journal_vol01_n05/page/n75/mode/2up?q=cslabs\">CSLabs</a></li>" +
				"<li><a href=\"https://martypc.blogspot.com/2024/09/pc-floppy-copy-protection-electronic.html\">Interlock</a> by Electronic Arts <small>used internally for their PC games of 1984-87</small></li>" +
				ul1 +
				"</p>",
		},
		{
			Title: "Microsoft DOS v2, ANSI, and the PC clones", Year: 1983, Month: 3,
			Lead: "Origins of ansi art on microcomputers", LinkTitle: "about MS-DOS ANSI.SYS",
			Link: "https://github.com/microsoft/MS-DOS/blob/master/v2.0/source/ANSI.txt",
			Content: "<p>" +
				"March saw the release of the Microsoft DOS version 2. Reprogrammed from scratch to ultimately distance Microsoft from its 86-DOS licensing contract with Seattle Computer Products, as well as any conceivable claims of code theft of Digital Research's CP/M operating system, which was the inspiration for 86-DOS." +
				"</p><p>" +
				"MS-DOS 2 included a new special device driver, ANSI.SYS, to allow the IBM PC to view ANSI escape control formatting and color text on the microcomputer. However, the implementation was incomplete, and in typical Microsoft fashion, future updates deviated from the documented standard." +
				"</p><p>" +
				"Also, the month saw Compaq Computer Corporation release the first unauthorised IBM PC compatible computer, the <a href=\"https://www.computerhistory.org/revolution/personal-computers/17/302/1194\">Compaq Portable</a>. " +
				"And Compaq would use Microsoft's operating system." +
				"</p>",
		},
		{
			Title: "Microsoft Windows announced", Year: 1983, Month: 11, Day: 10,
			Link:      "https://www.poynter.org/reporting-editing/2014/today-in-media-history-in-1983-bill-gates-and-microsoft-introduced-windows/",
			LinkTitle: "the announcement",
			Content: "<p>Around this time, <abbr title=\"graphical user interface\" class=\"initialism\">GUI</abbr> for microcomputing was all the hype within the technology industry and media. " +
				"In hindsight, this premature announcement from Microsoft aimed to keep customers from jumping ship to competitor platforms and GUI offerings.</p>" +
				"<p>It took a decade before graphical interfaces on the PC replaced text in business computing with Windows&nbsp;NT&nbsp;<sup>1993</sup> and even longer with Windows&nbsp;95&nbsp;<sup>1995</sup> before it became commonplace in the home." +
				" Other microcomputer platforms, such as the <span class=\"text-nowrap\">Apple&nbsp;Macintosh <sup>1984</sup></span>, <span class=\"text-nowrap\">Commodore&nbsp;Amiga</span> and <span class=\"text-nowrap\">Atari&nbsp;ST&nbsp;<sup>1985</sup></span> came with a GUI as standard.</p>",
		},
		{
			Title: "Earliest Unprotect texts", Year: 1983, Highlight: true,
			Lead: "So far, Unprotects for Lotus 1-2-3", LinkTitle: "the Unprotect texts",
			Link: "/files/how-to",
			Content: "<code>123.UNP</code><br>" +
				"<p>The January release of <a href=\"https://www.wired.com/2009/01/jan-26-1983-spreadsheet-as-easy-as-1-2-3-2/\">1-2-3</a> from Lotus " +
				"became the killer application for the IBM PC and helped the platform dominate in business and the home in the USA. " +
				"Like VisiCalc on the Apple 2, it was a spreadsheet application running on the powerful IBM personal computer that allowed for a more extensive feature set and usability.</p>" +
				"<p>All the early editions of the 1-2-3 came with floppy disk copy protection, which allowed for hard drive installation but required the original purchased floppy disk when loading the program. " +
				"The loss or easy damage of this key disk left a buyer unable to use their expensive 1983-$500 software.</p>" +
				"<p>Many Unprotect texts provide instructions on how end users can <a href=\"/f/af29fa4\">hack and edit</a> Lotus 1-2-3 to remove its copy protection. " +
				"It seems that so many people were frustrated with this copy protection that Lotus eventually abandoned it. " +
				"However, it is not sure if 1-2-3 is the origin of Unprotect on PC or if it merely popularized. " +
				"But the form of removal was also used on the Apple 2, such as in this <a href=\"http://www.textfiles.com/apple/parameters.txt\">1982 log</a>.</p>",
		},
		{
			Title: "The first 16 color PC game", Year: 1984, Month: 8,
			Lead: "King's Quest", LinkTitle: "the game manual",
			Link: "http://www.sierrahelp.com/Documents/Manuals/Kings_Quest_1_IBM_-_Manual.pdf",
			Content: "<p>" +
				"The first PC game to use 16 colors, <a href=\"https://www.mobygames.com/game/122/kings-quest/screenshots/pc-booter/\">King's Quest</a>, is created by Sierra On-Line and released by IBM. " +
				"IBM&nbsp;PC graphics cards are limited to monochrome or 4 colors, but the game is released for the new <strong>IBM&nbsp;PCjr</strong> that displays upto <strong>16 colors</strong>. " +
				"The other pioneering aspect of the game was the pseudo-3D landscape. The player controlled a human avatar from a 3rd person perspective and could use it to walk around set pieces, both in front and from behind, and interact with the onscreen objects." +
				"</p><p>" +
				"King's Quest did not run off the PC's disk operating system; instead, the game floppy disk had its own self-booting loader, today referred to as a PC booter. For the time, the game had aggressive <a href=\"https://martypc.blogspot.com/2024/08/pc-floppy-copy-protection-formaster.html\">copy protection</a> using Formaster's Copy-Lock.</p>",
		},
		{
			Title: "The earliest information text on PC", Year: 1984, Month: 10, Day: 17, Highlight: true,
			Lead:      "So far, Software Pirates Inc. - ZorkTools 1.0",
			LinkTitle: "the information text",
			Link:      "/f/ae2da98",
			Content: "<code>INFOCOM.DOC</code><br>" +
				"<p><strong>Information texts</strong> were documents stored as plain text and included in a release describing how to use a utility program or game. " +
				"While the texts were common on the Apple II, it took years for them to appear on the PC.</p>" +
				"<p>The author of this document is part of <a href=\"/g/software-pirates-inc\">Software Pirates Inc.</a>, one of the earliest known groups on the PC underground, dating back to at least 1984. " +
				"Whether an individual or collective, the brand was prolific in writing documentation and coding utilities for the PC but kept themselves anonymous.</p>" +
				"<p>May 1985 saw the release of the <strong>ARC archiving and compression tool</strong> that immediately caught on with the PC BBS community. " +
				"It also allowed pirate groups to package releases with multiple files, such as information texts. " +
				"This <a href=\"/f/b32077c\">SPI release</a> of The World's Greatest Baseball Game, packed in December 1985, includes a <code>BASEBALL.DOC</code> textfile describing the game and how to run it.</p>",
		},
		{
			Title: "EGA graphics standard", Year: 1984, Month: 10,
			// Lead: "16 color, 64 color palette, 640x350 resolution!?",
			LinkTitle: "How 16 colors saved PC gaming",
			Link:      "https://www.custompc.com/retro-tech/ega-graphics",
			Content: "<p>The new Enhanced Graphics Adapter standard from IBM uses: </p>" +
				ul0 +
				"<li>16 colors onscreen</li>" +
				"<li>64 color palette</li>" +
				"<li>maximum 640 x 350 resolution</li>" +
				"<li>80x25 character text mode</li>" +
				ul1 +
				"<p><a href=\"http://nerdlypleasures.blogspot.com/2014/01/simcity-for-dos-swiss-army-knife-of.html\">With the odd exception</a>, most PC games that use <strong>EGA</strong> only ever support 160x200 or 320x200 resolutions with 4 or 16 colors on screen. " +
				"There were complications with EGA and its expensive monitor displays, plus the expensive memory requirements needed for higher resolution graphic modes with <strong>16 colors</strong>.</p>" +
				"<p>IBM would also create the first <strong>demo program</strong> on the PC, <a href=\"https://www.pcjs.org/software/pcx86/demo/ibm/ega/\">Fantasy Land EGA</a>, is released to demonstrate the new <strong>EGA</strong> graphics standard. " +
				"The idea of a demo is to have the program run automatically, without user input, to show off the capabilities of the hardware.",
			Picture: Picture{
				Title: "The 1984 PC game, Ancient Art of War, updated in 1987 with EGA colors",
				Alt:   "Character selection screenshot for the EGA update to Ancient Art of War",
				Png:   "ega-ancient_art_of_war.png",
			},
		},
		{
			Title: "The earliest PC cracked releases", Year: 1984, Highlight: true,
			Lead: "So far, The Duplicators and 'public domain'",
			Link: "/g/the-duplicators", LinkTitle: "about these pioneering crackers",
			Content: "<p>This modified, tagged, or graffitied title screen is a <strong>crack&nbsp;screen</strong> " +
				"and was a typical way for crackers on the Apple, Atari microcomputers, and the IBM PC to credit themselves. " +
				"Crackers altered games and removed disk copy protection from software to permit copying and redistribution.</p>" +
				"<p>The earliest examples we have on the IBM PC are cracked games from mid-1984, attributed to <a href=\"/g/the-duplicators\">(C) 1984 The Duplicators</a>. " +
				"The plurality in the name suggests it was a small group, but it could have been a solo cracker. " +
				"And they probably did some prior cracks on the Apple II such as <a href=\"https://demozoo.org/productions/381192\">a crack</a> for Track Attack.</p>" +
				"<p>Also, an oddity on the PC are the anonymous cracked games where the copyright information gets replaced with text proclaiming the game is public domain, such as this 1984 " +
				"<a href=\"/f/ab27d16\">example of Stargate</a>. " +
				"The reasoning for this is uncertain.</p>",
			Picture: Picture{
				Title: "HHM broken by the Duplicators",
				Alt:   "Copyright 1984 the Duplicators screenshot on the PC",
				Webp:  "a319104.webp",
				Png:   "a319104.png",
			},
		},
		{
			Title: "The year of the Commodore 64", Year: 1984,
			Lead: "Computers goes mainstream", LinkTitle: "about the Commodore 64", Link: "http://variantpress.com/books/commodore-a-company-on-the-edge/",
			Content: "<p>" +
				"While the Commodore 64, or C&hyphen;64, would first hit the market in August 1982, manufacturing constraints and quality control issues would result in a tiny number of machines solely in the USA and Japan. " +
				"At year's end, there were around 50,000 Commodore 64s worldwide or back-ordered, and one million Commodore VIC&hyphen;20 microcomputers, the less capable precursor. " +
				"Of that million, 800,000 were sold in the USA, almost 200,000 in Europe, with half in the UK.</p>" +
				"<p>The C&hyphen;64 problems are reflected in the press of the time, with Creative Computing <a href=\"https://archive.org/details/CreativeComputingbetterScan198301/page/n22/mode/1up\">reviewing</a> a pre-production unit for the January 1983 magazine, which praised the machine but complained about build quality, especially the television output. " +
				"The issue features a prime C&hyphen;64 <a href=\"https://archive.org/details/CreativeComputingbetterScan198301/page/n346/mode/1up\">advert</a> from Commodore. However, in the subsequent magazine issues, Commodore <a href=\"https://archive.org/details/CreativeComputingbetterScan198303/page/n112/mode/1up\">replaced</a> the new C&hyphen;64 ad with advertising for the older VIC&hyphen;20. " +
				"A <a href=\"https://archive.org/details/CreativeComputingbetterScan198304/page/n293/mode/1up\">citation</a> in the April 1983 issue may suggest why, <em>According to Neil Harris, in 1980, 10,000 Vic 20 units were sold nationwide. Toward the end of 1982, Commodore was manufacturing 10,000 Vic 20 units per day. And the new machine, the 64, is back-ordered in the tens of thousands of units.</em>" +
				"</p><p>" +
				"Another issue that plagued the C&hyphen;64 in the USA was the unavailability of the disk drive. Creative Computing <a href=\"https://archive.org/details/CreativeComputingbetterScan198308/page/n230/mode/1up\">wrote</a> in August 1983, " +
				"<em>You could make an investment in the speed, convenience, and reliability of a disk drive. The only problem with this approach is cost, which in some cases exceeds that of the computer itself. " +
				"But if you chose not to part with $400 for a drive, you were stuck with the very dreary prospect of cassette storage.</em> " +
				"Compute! December 1983 would <a href=\"https://archive.org/details/1983-12-computegazette/page/n7/mode/1up\">report</a>, <em>We are hearing that 1541 drives are virtually unavailable, and that many drives purchased before the supply dried up suffer from reliability problems</em> " +
				"and later confirming <em>the nearly total absence of 1541s from dealers' shelves in August and September.</em> " +
				"A problem for most buyers wanting the drive, " +
				"[C-64] <em>sales with disk drives are running at 90 percent.</em>" +
				"</p><p>" +
				"However, the biggest problem for the platform throughout 1983 was the lack of software availability, which was emphasized in many publications. " +
				"Compute! <a href=\"https://archive.org/details/1983-08-computegazette/page/n25/mode/1up\">wrote</a> in August, <em>Although the Commodore 64 has been around for almost a year now, software is still scarce. There are many good programs available, but merchants and customers are frustrated that there aren't more.</em> " +
				"Ahoy! of January 1984 <a href=\"https://archive.org/details/ahoy-magazine-01/page/n38/mode/1up\">claimed</a>, " +
				"<em>The C-64's main failing point has been the relative scarcity of software, and while the computer's sales success is changing all that, the gap between a 64 owner's selection and a IBM PCJr/PC's is wide and not soon to be bridged-if ever.</em>" +
				"</p><p>" +
				"But Creative Computing was more optimistic, in October 1983 <a href=\"https://archive.org/details/CreativeComputingbetterScan198311/page/n135/mode/1up\">reporting</a>, <em>Sierra On-Line has entered the slowly maturing Commodore 64 software market with three converted Apple games</em>. " +
				"And in December <a href=\"https://archive.org/details/CreativeComputingbetterScan198312/page/n328/mode/1up\">writing</a>, " +
				"<em>Not to slight original efforts for the 64, but frankly, the best software packages available for the Commodore 64 right now are translations from the Apple and Atari</em> [game ports]. " +
				"<em>The top-notch houses, including Sierra On-Line, Sirius, and Synapse, are working night and day to translate their hits for the 64.</em> " +
				"</p><p>" +
				"While 1983 was an amazing year for Commodore, it was likely due to massive sales of the VIC&hyphen;20. It would take until 1984 for the Commodore 64 to solve many of the supply issues, to improve quality control, and  continue to see reductions in prices. " +
				"1984 would launch or see the start of several dedicated magazines, including RUN and Ahoy! in the USA; 64'er and Input 64 in Germany, and UK's Your Commodore. But importantly, in 1984, the Commodore 64 would see wider support from software publishers, with the new year opening with " +
				"<em><a href=\"https://archive.org/details/1984-01-computegazette/page/n3/mode/2up\">Electronics Arts Comes To The Commodore</a></em> and other majors, like Strategic Simulations Inc., <a href=\"https://archive.org/details/1984-01-computegazette/page/n84/mode/1up\">following</a>." +
				"</p><p>" +
				"The Commodore 64 became the all time, best-selling microcomputer, with many millions sold.</p>",
		},
		{
			Title: "The Berlin Bear controversy", Year: 1984,
			Lead: "Commodore 64",
			Content: "<p>" +
				"Way back in the 2000s, many in the Demoscene <a href=\"https://www.pouet.net/prod.php?which=17555\">argued</a> that a 1982 Berlin Bear image drawn for the Commodore 64 cracker group " +
				"<a href=\"https://csdb.dk/group/?id=2845\">Berlin Cracking Service</a> was the first ever Scene intro and cracktro. However, the <a href=\"https://www.atlantis-prophecy.org/recollection/?load=interviews&id_interview=7\">claim</a> was outlandish for multiple reasons, and either it was fabricated or a memory bias. Unfortunately, memory bias and conjecture are quite common when reflecting on the early Scene." +
				"</p><p>" +
				"The Scener, Jazzcat <a href=\"https://www.atlantis-prophecy.org/recollection/?load=crackers_map&country=germany\">wrote</a> of the cracktro, [the group] <em>for some time claimed the glory of having the first real crack intro which was the famous screen</em>. Of the image itself, <em>the picture was discovered to be in Paint Magic format which did not appear until 1983.</em> " +
				"</p><p>" +
				"Paint Magic, authored by Mark Riley, was a drawing tool for the Commodore 64 that sold for $50 by the Californian company Datamost. " +
				"However, despite the &copy;1983 in the print <a href=\"https://archive.org/details/game_manual_Paint_Magic/page/n3/mode/2up\">manual</a>, it was likely released in 1984. " +
				"Given that it was <a href=\"https://archive.org/details/cbm_magazine_index-power_play/power_play/1984/power_play-08-198403/page/n21/mode/1up?q=datamost+paint+magic\">showcased</a> at January's CES '84, " +
				"and the <a href=\"https://archive.org/details/ahoy-magazine-08/page/n48/mode/1up\">reviews</a> and " +
				"<a href=\"https://archive.org/details/the-everything-book-for-the-commodor-c-64-vic-20-home-computer-summer-1984/page/22/mode/1up?q=datamost+paint+magic\">reseller ads</a> are only found in magazines of 1984 and 1985." +
				"</p><p>" +
				"So far, the image itself has only been discovered in cracks of games that were published in mid-1984. These happen to be ports of Activision titles that were advertised as <em><a href=\"https://archive.org/details/computes-gazette-issue-015-september-1984/page/18/mode/2up\">Introducing Activision For Your Commodore 64</a></em> in the August and September issues of various Commodore magazines. " +
				"Ahoy <u>September 1984</u> <a href=\"https://archive.org/details/ahoy-magazine-09/page/n8/mode/1up\">writes</a> under New Games Update, " +
				"<em>Activision's Pitfall II: Lost Caverns, forecast in these pages in July, is now available</em>, which dates <a href=\"https://demozoo.org/productions/382466/\">this</a> Berlin Bear usage." +
				"</p><p>" +
				"And assuming the Berlin Bear artwork is from early or late 1984, there are dozens of examples of prior Scene art on the Commodore." +
				"</p><p>" +
				"Today, the most obvious counterpoint to <q>being first</q> is the Apple II cracking Scene productions that existed years before the Commodore 64. " +
				"There's the text artwork <a href=\"https://demozoo.org/productions/380718/\">loader</a> by The Tornado that is self-dated to November 1980. " +
				"As well as the animated <a href=\"https://demozoo.org/productions/381802/\">loader</a> created by &hyphen;The Razor&hyphen; for a game repack self-dated to April 1981. " +
				"And Apple cracking groups such as the Midwest Pirate Guild were using custom art in their 1983 <a href=\"https://demozoo.org/productions/288324/\">loaders</a>." +
				"</p>",
			Picture: Picture{
				Title:       "Berlin Bear upside down",
				Alt:         "A screenshot of the Berlin Bear image for the Commodore 64.",
				Png:         "berlin-bear.png",
				Attribution: "Jazzcat but flipped by us",
			},
		},
		{
			Title: "First, dial-up Internet connections", Year: 1984,
			Link:      "https://networkencyclopedia.com/serial-line-internet-protocol-slip/",
			LinkTitle: "about SLIP",
			Content: "<p>" +
				"Rick Adams created the Serial Line Internet Protocol (<strong>SLIP</strong>), the industry-standard protocol to connect dial-up modems to the Internet. " +
				"This protocol allowed for the creation of Internet Service Providers, which provided Internet connections over standard copper telephone lines." +
				"<br>In 1987, Rick would also go on to found one of the earliest ISPs, UUNET. " +
				"Which in the following year would offer the first commercial connection to the Internet.</p>" +
				"<p>Below is a <a href=\"https://www.ascilite.org/archived-journals/aset/confs/edtech94/rw/rehn.html\">mockup</a>, using SLIP in 1993 to connect to a Western Australian university provider.</p>" +
				`<pre style="font-size:1.5em;line-height:1em;" class="font-dos-mda reader reader-invert border border-black rounded-1 p-1"><p>` +
				"<br>		WELCOME<br><br>" +
				"login: defacto2<br>" +
				"Password:<br>" +
				"Last login: Tue June 28 18:44:50<br>" +
				"SunOS Release 4.1.3 (CSI) #2: Mon Mar 8 13:58:16 WST 1993<br><br>" +
				"Welcome to cleo.murdoch.edu.au, Academic Services Unit, Murdoch<br><br>" +
				"For user support via email, email to userhelp@cleo<br><br>" +
				"ELECTRONIC MAIL: To access your email, type the command \"pine\".<br>" +
				"LIBRARY CATALOGUES: To access remote library catalogues, type \"nis\".<br>" +
				"GOPHER SERVER: To access the gopher information server, type \"gopher\".<br>" +
				"WWW SERVER: To access Murdoch's WWW servers, type \"lynx\".<br><br>" +
				"You have new mail.<br><br>" +
				"cleo><span class=\"blinking\">█</a>" +
				`</p></pre>`,
		},
		{
			Prefix: "The earliest PC groups,", Year: 1984,
			List: Links{
				{
					LinkTitle: "Against Software Protection <small>ASP</small>",
					Link:      "/g/against-software-protection",
					Forward:   "Atlanta, Connecticut, Miami",
				},
				{LinkTitle: "The Duplicators", Link: "/g/the-duplicators"},
				{LinkTitle: "The IPL", Link: "/g/ipl", Forward: "Chicago"},
				{
					LinkTitle: "Software Pirates Inc <small>SPi</small>",
					Link:      "/g/software-pirates-inc",
					Forward:   "Texas and Sunnyvale, CA (?)",
				},
				{LinkTitle: "Faked 'public domain' releases", Link: "/g/public-domain"},
			},
			Picture: Picture{
				Title: "Copyright invalid in 1984",
				Alt:   "Copyright invalid in 1984 by SPI screenshot on the PC",
				Png:   "b92e146.png",
			},
		},
		{
			Title: "The release of ARC", Year: 1985, Month: 3,
			Lead:      "The file ARChive utility",
			LinkTitle: "about the tool",
			Link:      "/compression",
			Content: "<p>Authored by Thom Henderson and released sometime in March 1995, " +
				"ARC quickly took the PC BBS scene by storm by allowing boards and users to use a single application to both archive and compress a directory of files into a single package. " +
				"The adoption was rapid, with contemporary texts claiming it was in widespread use by the year's end.</p>" +
				"<p>Its impact on the scene allowed groups like Software Pirates Inc. to bundle additional help and description files in their releases and would later leave the opinion of including separate BBS ads, intros, cracktros with the release.</p>" +
				`<pre style="font-size:1.5em;line-height:1em;" class="font-dos-mda reader reader-invert border border-black rounded-1 p-1">` +
				"<br><p>ARC - Archive utility, Version 3.10, created on 05/01/85 at 22:34:50<br>" +
				"(C) COPYRIGHT 1985 by System Enhancement Associates; ALL RIGHTS RESERVED" +
				`</p></pre>`,
		},
		{
			Title: "The earliest text loader on PC", Year: 1985, Month: 5, Day: 26, Highlight: true,
			Lead:      "So far, Spy Hunter cracked by Spartacus",
			LinkTitle: "and view the text loader",
			Link:      "/f/aa2be75",
			Content: "<p><strong>Loaders</strong> are bits of code that crackers and pirate groups insert to promote themselves and their game releases. As the name suggests, they are loaded and shown before the game starts. " +
				"Loaders originated on the Apple&nbsp;II and later the Commodore&nbsp;64 piracy Scenes.</p>" +
				"<p>While text loaders and ANSI art look similar, the execution is entirely different. ANSI art relies on plain text files encoded with ASCII escape control codes. " +
				"In contrast, text loaders are computer applications that use the computer's text characters stored in the system graphics card <a href=\"https://minuszerodegrees.net/video/bios_video_modes.htm\">ROM</a>, acting as a text programming interface.</p>" +
				"<p>Little is known about the Imperial Warlords that released this 1984 PC game port, though the two BBS advertised are from San Francisco and Minneapolis, which suggests a national group.</p>" +
				`<pre style="font-size:9px;" class="font-dos-cga reader reader-invert border border-black rounded-1 p-1">` +
				`<br><div><span style="color:#00c400;background-color:#000;">                                                                                </span><br>` +
				`<span style="color:#00c400;background-color:#000;">              </span><span style="color:#4ef3f3;background-color:#000;">And now... Presenting... the fourth of the series...</span><br><span style="color:#00c400;background-color:#000;">              </span><br>` +
				`<span style="color:#00c400;background-color:#000;">                                                                                </span><br>` +
				`<span style="color:#00c400;background-color:#000;">              </span><span style="color:#00c4c4;background-color:#000;"> </span><span style="color:#00c400;background-color:#000;">                  </span><span style="color:#c47e00;background-color:#000;">          </span><span style="color:#4ef3f3;background-color:#000;">    </span><span style="color:#00c400;background-color:#000;">           \/\/\/\/\/\/          </span><br>` +
				`<span style="color:#00c400;background-color:#000;">                     ▓▓▓▓▓▓▓     </span><span style="color:#c47e00;background-color:#000;">─┬─</span><span style="color:#00c400;background-color:#000;">      </span><span style="color:#c47e00;background-color:#000;">─┬─ </span><span style="color:#4ef3f3;background-color:#000;"> </span><span style="color:#00c400;background-color:#000;">            \/ / \  \/           </span><br>` +
				`<span style="color:#00c400;background-color:#000;">                     ▓▓▓▓▓▓▓▓▓▓  </span><span style="color:#4ef3f3;background-color:#000;"> </span><span style="color:#c47e00;background-color:#000;">├────────┴─┐</span><span style="color:#4ef3f3;background-color:#000;"> </span><span style="color:#c40000;background-color:#000;"> </span><span style="color:#00c400;background-color:#000;"> </span><span style="color:#4ef3f3;background-color:#000;"> </span><span style="color:#00c400;background-color:#000;">          \/ \/\ /\           </span><br>` +
				`<span style="color:#00c400;background-color:#000;">                     ▓▓▓▓▓▓▓▓▓▓▓▓</span><span style="color:#4ef3f3;background-color:#000;"> </span><span style="color:#c47e00;background-color:#000;">│</span><span style="color:#f3f34e;background-color:#000;">SPY</span><span style="color:#00c400;background-color:#000;"> </span><span style="color:#4edc4e;background-color:#000;">HUNTER</span><span style="color:#c47e00;background-color:#000;">│</span><span style="color:#4ef3f3;background-color:#000;"> </span><span style="color:#00c400;background-color:#000;">═ ═ ═ ═ ═ ═   \/ / / /           </span><br>` +
				`<span style="color:#00c400;background-color:#000;">                     ▓▓▓▓▓▓▓▓▓▓  </span><span style="color:#4ef3f3;background-color:#000;"> </span><span style="color:#c47e00;background-color:#000;">├────────┬─┘</span><span style="color:#0000c4;background-color:#000;"> </span><span style="color:#00c400;background-color:#000;">               \ \/ /            </span><br>` +
				`<span style="color:#00c400;background-color:#000;">                     ▓▓▓▓▓▓▓    </span><span style="color:#c47e00;background-color:#000;"> ─┴─</span><span style="color:#00c400;background-color:#000;">      </span><span style="color:#c47e00;background-color:#000;">─┴─ </span><span style="color:#00c400;background-color:#000;">                 \//\             </span><br>` +
				`<span style="color:#00c400;background-color:#000;">                                                                                </span><br>` +
				`<span style="color:#00c400;background-color:#000;">                                                                                </span><br>` +
				`<span style="color:#00c400;background-color:#000;">                                   </span><span style="color:#4ef3f3;background-color:#000;">Cracked by</span><span style="color:#00c400;background-color:#000;">                                   </span><br>` +
				`<span style="color:#00c400;background-color:#000;">                                                                                </span><br>` +
				`<span style="color:#00c400;background-color:#000;">                               </span><span style="color:#f34ef3;background-color:#000;">!</span><span style="color:#00c400;background-color:#000;">   </span><span style="color:#4edc4e;background-color:#000;">_________</span><span style="color:#00c400;background-color:#000;">   </span><span style="color:#f34ef3;background-color:#000;">!</span><span style="color:#00c400;background-color:#000;">                                </span><br>` +
				`<span style="color:#00c400;background-color:#000;">                                </span><span style="color:#f34ef3;background-color:#000;">\</span><span style="color:#00c400;background-color:#000;"> </span><span style="color:#4edc4e;background-color:#000;">/</span><span style="color:#00c400;background-color:#000;">         </span><span style="color:#4edc4e;background-color:#000;">\</span><span style="color:#00c400;background-color:#000;"> </span><span style="color:#f34ef3;background-color:#000;">/</span><span style="color:#00c400;background-color:#000;">                                 </span><br>` +
				`<span style="color:#00c400;background-color:#000;">       </span><span style="color:#0000c4;background-color:#000;"> </span><span style="color:#00c400;background-color:#000;">                      </span><span style="color:#f34ef3;background-color:#000;">!--</span><span style="color:#4edc4e;background-color:#000;">X</span><span style="color:#00c400;background-color:#000;"> </span><span style="color:#fff;background-color:#000;">SPARTACUS</span><span style="color:#00c400;background-color:#000;"> </span><span style="color:#f34ef3;background-color:#000;">X--!</span><span style="color:#4e4e4e;background-color:#000;"> </span><span style="color:#00c400;background-color:#000;">                              </span><br>` +
				`<span style="color:#00c400;background-color:#000;">                                </span><span style="color:#f34ef3;background-color:#000;">/</span><span style="color:#4e4e4e;background-color:#000;"> </span><span style="color:#4edc4e;background-color:#000;">\_________/</span><span style="color:#00c400;background-color:#000;"> </span><span style="color:#f34ef3;background-color:#000;">\</span><span style="color:#4e4e4e;background-color:#000;"> </span><span style="color:#00c400;background-color:#000;">                                </span><br>` +
				`<span style="color:#00c400;background-color:#000;">     </span><span style="color:#00c4c4;background-color:#000;">┌───────────────┐</span><span style="color:#00c400;background-color:#000;">         </span><span style="color:#f34ef3;background-color:#000;">!</span><span style="color:#00c400;background-color:#000;">           </span><span style="color:#f34ef3;background-color:#000;"> </span><span style="color:#00c400;background-color:#000;">   </span><span style="color:#f34ef3;background-color:#000;">!</span><span style="color:#00c400;background-color:#000;">           </span><span style="color:#f34ef3;background-color:#000;"> </span><span style="color:#00c400;background-color:#000;">                    </span><br>` +
				`<span style="color:#00c400;background-color:#000;">   </span><span style="color:#00c4c4;background-color:#000;">┌─┘</span><span style="color:#c47e00;background-color:#000;"> </span><span style="color:#c47e00;background-color:#000;">LOADING  GAME</span><span style="color:#00c4c4;background-color:#000;"> └─┐</span><span style="color:#00c400;background-color:#000;">                                                        </span><br>` +
				`<span style="color:#00c400;background-color:#000;">   </span><span style="color:#00c4c4;background-color:#000;">│</span><span style="color:#00c400;background-color:#000;"> </span><span style="color:#c47e00;background-color:#000;">PLEASE  STAND  BY</span><span style="color:#c47e00;background-color:#000;"> </span><span style="color:#00c4c4;background-color:#000;">│</span><span style="color:#00c400;background-color:#000;">             </span><span style="color:#4ef3f3;background-color:#000;">Of the</span><span style="color:#00c400;background-color:#000;">                                     </span><br>` +
				`<span style="color:#00c400;background-color:#000;">   </span><span style="color:#00c4c4;background-color:#000;">└───────────────────┘</span><span style="color:#00c400;background-color:#000;">                                                        </span><br>` +
				`<span style="color:#00c400;background-color:#000;">                                  </span><span style="color:#f34ef3;background-color:#000;">╔══════════╗</span><span style="color:#00c400;background-color:#000;">                                  </span><br>` +
				`<span style="color:#00c400;background-color:#000;">                                  </span><span style="color:#f34ef3;background-color:#000;">║</span><span style="color:#00c400;background-color:#000;"> </span><span style="color:#4ef3f3;background-color:#000;">IMPERIAL</span><span style="color:#00c400;background-color:#000;"> </span><span style="color:#f34ef3;background-color:#000;">║</span><span style="color:#00c400;background-color:#000;">                                  </span><br>` +
				`<span style="color:#00c400;background-color:#000;">                                  </span><span style="color:#f34ef3;background-color:#000;">║</span><span style="color:#00c400;background-color:#000;"> </span><span style="color:#f3f34e;background-color:#000;">WARLORDS</span><span style="color:#00c400;background-color:#000;"> </span><span style="color:#f34ef3;background-color:#000;">║</span><span style="color:#00c400;background-color:#000;">                                  </span><br>` +
				`<span style="color:#00c400;background-color:#000;">                                  </span><span style="color:#f34ef3;background-color:#000;">╚══════════╝</span><span style="color:#00c400;background-color:#000;">                                  </span><br>` +
				`</div><br></pre>`,
			// Picture: Picture{
			// 	Title: "Spy Hunter",
			// 	Alt:   "Spy Hunter by Imperial Warlords screenshot",
			// 	Webp:  "aa2be75.webp",
			// 	Png:   "aa2be75.png",
			// },
		},
		{
			Title: "The earliest PC ASCII art", Year: 1985, Month: 7, Day: 24, Highlight: true,
			Lead: "So far, How to WIN at KING's QUEST from The Illinois Pirates", LinkTitle: "and view the file", Link: "/f/bc30a5b",
			Content: "<p><strong>The Illinois Pirates</strong> walk-through for the PC exclusive game King's Quest released the earliest known PC <strong>ASCII art</strong> or Codepage 437 art. " +
				"The ASCII text logo uses block and line art characters that were exclusive to the IBM PC platform.</p>" +
				`<pre style="font-size:28px;" class="font-dos-mda reader reader-invert border border-black rounded-1 p-1">` +
				"<br>" +
				`/////////// How to WIN at KING's QUEST \\\\\\\\\\\\\\\<br>` +
				`\\\\\\\\\\\    on the IBM PC/PCjr      ///////////////<br><br>` +
				"                             as tabulated by<br>" +
				" The    ███████  █    █    ▀ ██    █ ██████ ▀ █████<br>" +
				"           █     █    █    █ █ █   █ █    █ █ █<br>" +
				"           █     █    █    █ █  █  █ █    █ █ █████<br>" +
				"           █     █    █    █ █   █ █ █    █ █     █<br>" +
				"        ███████  ████ ████ █ █    ██ ██████ █ █████<br>" +
				"           ╔════╗                          ╔═════╕<br>" +
				"           ║    ║          ══════╦══════   ║     │<br>" +
				"           ║    ║ ║              ║         ║<br>" +
				"           ╠════╝   ╠══╗ ╔═══╗   ║   ╔═══  ║<br>" +
				"           ║      ║ ║    ║   ║   ║   ║     ╚═════╗<br>" +
				"           ║      ║ ║    ╠═══╣   ║   ╠═          ║<br>" +
				"           ║      ║ ║    ║   ║   ║   ║     │     ║<br>" +
				"                                     ╚═══  ╘═════╝<br>" +
				"</pre>",
		},
		{
			Title: "Earliest ANSI ad", Year: 1985, Month: 8, Highlight: false,
			Lead: "So far, The Game Gallery", LinkTitle: "and view the file",
			Link: "/f/ba2bcbb",
			Content: "<p>The earliest <strong>ANSI ad</strong>vertisement is for the Manhattan based BBS, <strong>The&nbsp;Game&nbsp;Gallery</strong>&nbsp;(+212-799-6987). ANSI art is a computer art form that became widely used to create art and advertisements for online bulletin board systems.</p>" +
				"<p>The output uses ANSI escape codes, a standard Digital Equipment Corporation (DEC) pioneered for its minicomputer <a href=\"https://vt100.net/dec/vt_history\">video terminals</a>. Later, it was used on IBM and other PCs using software drivers and video <a href=\"https://vt100.net/emu/\">terminal emulators</a>.</p>" +
				`<pre style="font-size:inherit;" class="font-dos-cga reader reader-invert border border-black rounded-1 p-1"><div style="color:#aaa;background-color:#000;"><br>` +
				`<span style="color:#aaa;">         </span><span style="color:#fff;background-color:#00a;">Hi score </span><span style="color:#fff;background-color:#a00;">212-799-6987</span><br>` +
				`<span style="color:#aaa;">       </span><span style="color:#fff;background-color:#00a;">╔════════════════════════════════════╗</span><br>` +
				`<span style="color:#aaa;">       </span><span style="color:#fff;background-color:#00a;">║∙ █ ∙  ██ THE GAME GALLERY∙ ██ ∙ █ ■║</span><br>` +
				`<span style="color:#aaa;">       </span><span style="color:#fff;background-color:#00a;">║ ∙██ ∙  █  ∙∙∙ 300 1200 ∙∙∙ ███∙ █ .║</span><br>` +
				`<span style="color:#aaa;">       </span><span style="color:#fff;background-color:#00a;">║∙∙ ██ ∙∙∙∙∙∙∙∙∙∙ ██████ . ███ ∙ ∙█ .║</span><br>` +
				`<span style="color:#aaa;">       </span><span style="color:#fff;background-color:#00a;">║∙∙ ███ █∙ ██████ ∙∙∙ ██ ...∙∙∙ ███ .║</span><br>` +
				`<span style="color:#aaa;">       </span><span style="color:#fff;background-color:#00a;">║∙∙∙ █∙∙█∙∙  █∙∙ ████ ∙ ∙ ██████  </span><span class="blinking" style="color:#fff;background-color:#0aa;">` + "\x01" + `</span><span style="color:#fff;background-color:#00a;"> .║</span><br>` +
				`<span style="color:#aaa;">       </span><span style="color:#fff;background-color:#00a;">║</span><span style="color:#aaa;">               </span><span style="color:#fff;background-color:#00a;">` + "\x02" + `......</span><span class="blinking" style="color:#fff;background-color:#a00;">` + "\x01" + `</span><span style="color:#fff;background-color:#00a;">.............║</span><br>` +
				`<span style="color:#aaa;">       </span><span style="color:#fff;background-color:#00a;">║∙∙   ∙∙∙∙∙ ∙ .███ ███.█.... █ .███ .║</span><br>` +
				`<span style="color:#aaa;">       </span><span style="color:#fff;background-color:#00a;">║∙█∙    ████ ∙.█  </span><span class="blinking" style="color:#ff5;background-color:#a0a;">` + "\x01" + `</span><span style="color:#fff;background-color:#00a;">  █.███████ .█ ...║</span><br>` +
				`<span style="color:#aaa;">       </span><span style="color:#fff;background-color:#00a;">║∙██     ███∙ .█  </span><span class="blinking" style="color:#fff;background-color:#aaa;">` + "\x01" + `</span><span style="color:#fff;background-color:#00a;"> </span><span class="blinking" style="color:#fff;background-color:#000;">` + "\x01" + `</span><span style="color:#fff;background-color:#00a;">█.█ .......█████║</span><br>` +
				`<span style="color:#aaa;">       </span><span style="color:#fff;background-color:#00a;">║∙███     ██∙∙.███████.....█ .█......║</span><br>` +
				`<span style="color:#aaa;">       </span><span style="color:#fff;background-color:#00a;">║∙∙∙∙ ██  ∙∙∙∙.........█████ .█. ████║</span><br>` +
				`<span style="color:#aaa;">       </span><span style="color:#fff;background-color:#00a;">║∙∙∙∙∙ █    █∙███████████ ....█......║</span><br>` +
				`<span style="color:#aaa;">       </span><span style="color:#fff;background-color:#00a;">║∙█  ███</span><span style="color:#aaa;">                </span><span style="color:#fff;background-color:#00a;">█ <span class="blinking">` + "\x01" + `</span>..... ███.║</span><br>` +
				`<span style="color:#aaa;">       </span><span style="color:#fff;background-color:#00a;">║∙█∙∙∙∙∙∙     300 1200    ...██... █.║</span><br>` +
				`<span style="color:#aaa;">       </span><span style="color:#fff;background-color:#00a;">║ ∙ ██ ██   212-799-6987  █████. ███.║</span><br>` +
				`<span style="color:#aaa;">       </span><span style="color:#fff;background-color:#00a;">║∙∙∙∙ █∙∙  24HRS WEEKDAYS   .......■.║</span><br>` +
				`<span style="color:#aaa;">       </span><span style="color:#fff;background-color:#00a;">╚════════════════════════════════════╝</span><br>` +
				`<span style="color:#aaa;">        </span><span style="color:#fff;background-color:#00a;">For those who use the computer for</span><br>` +
				`<span style="color:#aaa;">        </span><span style="color:#fff;background-color:#00a;">recreation. </span><span style="color:#fff;background-color:#a00;">THE GAME</span><span style="color:#fff;background-color:#00a;"> </span><span style="color:#fff;background-color:#a00;">GALLERY.</span><br><br></div></pre>`,
		},
		{
			Title: "Razor 1911 is named", Year: 1985, Month: 11,
			Lead: "On the Commodore 64", LinkTitle: "about the early days of Razor 1911",
			Link: "https://csdb.dk/group/?id=431",
			Content: "<p><strong>Razor 1911</strong>, the oldest and most famed brand in the Scene, was founded in <strong>Norway</strong> and has three members. " +
				"The group released demos and later cracked exclusively for the Commodore&nbsp;64 and then the Amiga. Co-founder Sector 9 took the brand to the PC in <a href=\"/f/a12d5e\">late 1990</a>.</p>" +
				"<p>The distinctive number suffix was a fad with groups of the Commodore&nbsp;64 era<sup><a href=\"#razor-1911-is-named-fn1\">[1]</a></sup>.<br><q>1911</q> denotes the decimal value of hexadecimal <code>$777</code>.</p>" +
				sect0 +
				"<div id=\"razor-1911-is-named-fn1\">[1] Other named examples include, 1001&nbsp;Crew, 1701&nbsp;Crackware, The&nbsp;Gamebusters&nbsp;1541, The&nbsp;Professionals&nbsp;2010.</div>" +
				sect1,
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
				Png:   "windows-version-1.png",
			},
		},
		{
			Title: "Earliest, \"proto\" NFO text", Year: 1985, Month: 12, Day: 26, Highlight: false,
			Lead: "So far, Software Pirates Inc", LinkTitle: "the file", Link: "/f/b32077c",
			Content: "<p><strong>NFO</strong> information text files are usually distributed with pirated software to provide usage instructions, promote the release group, and occasionally encourage group propaganda.</p>" +
				"<p>Software Pirates Inc may have released the earliest NFO-like document for the late 1985 packaged release of " +
				"<a href=\"https://www.mobygames.com/game/22398/the-worlds-greatest-baseball-game/\">The World's Greatest Baseball Game</a>.</p>" +
				`<pre style="font-size:24px;" class="font-dos-mda reader reader-invert border border-black rounded-1 p-1">` +
				"Welcome to the Software Pirates, Inc.  version of Baseball" + br +
				"If you are new to the Software Pirates concept of DOS" + br +
				"files of your favorite protected program then you can help" + br +
				"us.  Send us your favorite protected diskette and we will" + br +
				"return it as DOS compatible file(s).  We hope you can help" + br +
				"this worthy cause.  We offer an exclusive money back" + br +
				"guarantee and warranty for the life of the program, if it" + br +
				"should ever fail you.  If you are not new to the SPI" + br +
				"concept, we still welcome donations of your protected" + br +
				"diskettes." + br + br +
				"Instructions for playing Baseball." + br +
				"Baseball is a 3 file set, including this documentation" + br +
				"file.  The other two files are 1.  BASEBALL.COM, the" + br +
				"loader and diskette emulator, 2.  BASEBALL.SPI, the" + br +
				"diskette image These files are distributed under the ARC" + br +
				"format, to retain their consistency." + br + br +
				"Starting" + br +
				"Change the DOS default prompt to the drive containing" + br +
				"BASEBALL.SPI and execute the command BASEBALL." + br +
				"...</pre>",
		},
		{
			Prefix: "The earliest PC groups,", Year: 1985,
			List: Links{
				{LinkTitle: "Imperial Warlords", Link: "/g/imperial-warlords", Forward: "San Francisco and Minnesota"},
				{LinkTitle: "The Illinois Pirates <small>TIP</small>", Link: "/g/the-illinois-pirates", Forward: "Chicago"},
			},
			Picture: Picture{
				Title: "The Illinois Pirates hack",
				Alt:   "The Illinois Pirates in-game hack on the PC screenshot",
				Png:   "ad1d67e.png",
			},
		},
		{
			Title: "The earliest PC \"DOX\"", Year: 1986, Highlight: true,
			Lead: "So far, Dam Buster documentation by Brew Associates", LinkTitle: "the documentation",
			Link: "/f/a61db76",
			Content: "<code>DAMBUST1.DOC</code><br>" +
				"<p><strong>DOX</strong> is an abbreviation for documentation, which are text files that provide instructions on playing more complicated games. " +
				"Games not in the arcade or action genre were usually unintuitive and relied on printed gameplay " +
				"<a href=\"https://archive.org/details/extras_msdos_Microsoft_Flight_Simulator_v1.0_1982/mode/2up\">instruction manuals</a> sold with the purchased game box to be usable.</p>" +
				"<p><q>The primary reason for the writing of this file is the fact that people may not be fully appreciating the Dam Buster game.  " +
				"I have seen some documentation out, but it is lame at best. What I have given you here is the actual text of the actual documentation distributed with the game. Enjoy!</q> " +
				"Dam Buster is a misname of <a href=\"https://archive.org/details/msdos_The_Dam_Busters_1985\">The Dam Busters</a>, a 1984-85 game published by Accolade.</p>" +
				"<p>Piracy groups had been including forms of gameplay instructions as text documents for the more complicated game releases for years, so it is unlikely this example is the first PC DOX. " +
				"An oddity is that for much of the 1980s, the PC was not the primary development platform for games. " +
				"This instead occurred on the Apple, Atari, and later the Commodore microcomputers, and afterwards the games were ported to the PC. " +
				"Pirates on the PC would often <a href=\"/f/b5258ae\">reuse</a> the \"DOX\" documents that got authored for those microcomputers rather than writing their own.</p>",
		},
		{
			Title: "PC clone sales pickup in Europe", Year: 1986,
			Link:      "https://www.computerhistory.org/revolution/personal-computers/17/302",
			LinkTitle: "about the PC clone market",
			Content: "While the Commodore, Apple and IBM are common platforms in the US, the European market doesn't always share the same popular platforms. " +
				"Import duties, slow international distribution channels and a lack of localized software and hardware often hampers the adoption of some platforms. " +
				"<br>The Western European market is dominated by Acorn, Amstrad, Commodore, Sinclair but the PC clones produced by local electronic manufactures gain popularity. " +
				"Popular machines include the <a href=\"https://www.dosdays.co.uk/computers/Amstrad%20PC1000/amstrad_pc1000.php\">Amstrad&nbsp;PC1512</a>, " +
				"the Philips&nbsp;P2000T and the <a href=\"https://www.dosdays.co.uk/computers/Olivetti%20M24/olivetti_m24.php\">Olivetti&nbsp;M24</a>.",
			Picture: Picture{
				Title:       "The Olivetti M24",
				Jpg:         "olivetti-m24.jpg",
				Avif:        "olivetti-m24.avif",
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
				Avif:        "brain-virus.avif",
				Attribution: "Avinash Meetoo",
				License:     "CC-BY-2.5",
				LicenseLink: "https://creativecommons.org/licenses/by/2.5/deed.en",
			},
		},
		{
			Title: "The first 16 color EGA game", Year: 1986, Month: 3,
			Lead: "Accolade's Mean 18", LinkTitle: "the moby games entry",
			Link: "https://www.mobygames.com/game/152/mean-18/",
			Content: "It may seem strange today, but golf games were popular in the 1980s and 1990s. " +
				"Real-life sports were aspirational for many white-collar US and Japanese workers, " +
				"so it isn't surprising that video game golf simulations targeting expensive computer platforms and arcades have become popular.",
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
			Title: "The earliest PC loaders", Year: 1986, Month: 3, Highlight: true,
			Content: "<p><strong>Loaders</strong> acted as they were named, given that they would be the first thing to load and display each time the cracked game was run. " +
				"These screens were static images created in <a href=\"https://winworldpc.com/product/pc-paint/100a\">PC Paint</a> in the early days, and sometimes contained ripped screens from other games. Some users found these annoying and a cause of unwanted file bloat.</p>" +
				"<p>The first static loaders originated on the Apple&nbsp;II underground, such as <a href=\"http://artscene.textfiles.com/intros/APPLEII/cbaseball.gif\">this example</a> " +
				"by The&nbsp;Digital&nbsp;Gang for the crack release of Championship&nbsp;Baseball that likely came out in 1983.</p>",
			List: Links{
				{LinkTitle: "Alley Cat from Five O", Link: "/f/b01c518"},
				{LinkTitle: "Conquest from T.O.A.D.S.", Link: "/f/bb2e428"},
				{LinkTitle: "Tapper from T.O.A.D.S.", Link: "/f/a6197ae"},
			},
			Picture: Picture{
				Title: "Software Pirates, Inc presents",
				Alt:   "Software Pirates, Inc presents Frogger II  screenshot",
				Png:   "a6197ae.png",
			},
		},
		{
			Year: 1986, Prefix: notable,
			List: Links{
				{LinkTitle: "ESP Pirates", Link: "/g/esp-pirates", Forward: "Arizona"},
				{LinkTitle: "Five-O", Link: "/g/five-o", Forward: "Minnesota"},
				{
					LinkTitle: "T.O.A.D.S. <small>TOADS</small>",
					Link:      "/g/toads",
					Forward:   "Chicago and San Francisco",
				},
				{LinkTitle: "Cracking On the IBMpc", Link: "/g/cracking-101", Forward: "Buckaroo Banzai"},
			},
			Picture: Picture{
				Title: "Five O Presents",
				Alt:   "Five O Presents screenshot",
				Png:   "ac1b5ea.png",
			},
		},
		{
			Title: "Fairlight is founded", Year: 1987, Month: 3,
			Lead: "On the Commodore 64 and Amiga", LinkTitle: "about the early days of Fairlight",
			Link: "http://janeway.exotica.org.uk/target.php?idp=6375&idr=1940&tgt=1",
			Content: "<strong>Fairlight</strong>, one of the oldest brands in the Scene, is founded in <strong>Sweden</strong> with just three members. " +
				"The group cracked and released demos exclusively for the Commodore&nbsp;64 and Amiga platforms before expanding to consoles and the <a href=\"/f/b04615\">PC</a> in February 1991.",
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
			// Lead: "256 color graphics",
			LinkTitle: "about the VGA graphics standard",
			Link:      "https://www.computer.org/publications/tech-news/chasing-pixels/Famous-Graphics-Chips-IBMs-VGA",
			Content: "<p>The new Video Graphics Array standard from IBM uses: </p>" +
				ul0 +
				"<li>256 colors onscreen</li>" +
				"<li>262144 color palette</li>" +
				"<li>maximum 640 x 480 resolution</li>" +
				"<li>80x25 character text mode</li>" +
				ul1 +
				"Unlike IBM's other 18-bit color standards, VGA is the first standard to support <strong>256 colors</strong> onscreen, resolutions up to 640x480, but also maintain backwards compatibility with software for CGA <u>and EGA</u>. " +
				"However, it would be years before game developers fully adopted the improved color palettes. " +
				"Initially it was used to mimick the Commodore Amiga, 32 of 4096 colors in game ports, " +
				"before games on PC embraced all 256 colors onscreen around 1990, give or take a year. " +
				"Both, with the use of digitalized photography, scanned images in game, and later multimedia games released on CD-ROM. " +
				"But VGA games using the 640x480 resolution would be less common.",
		},
		{
			Title: "Music audio standard", Year: 1987,
			Lead: "AdLib Music Synthesizer Card", LinkTitle: "about the AdLib sound card",
			Link: "https://www.computinghistory.org.uk/det/23724/AdLib-Music-Synthesizer-Card/",
			Content: "The Music Synthesizer Card sound card is released. It was the first sound card to use FM synthesis and the first widely adopted by game developers. " +
				"<strong>AdLib</strong>'s success was short-lived, as competitor <a href=\"https://www.creative.com\">Creative&nbsp;Labs</a> released the <a href=\"https://www.vgmpf.com/Wiki/index.php?title=Sound_Blaster\">Sound&nbsp;Blaster</a> in 1989, " +
				"a clone of the AdLib card that included a simple digital sound processor for speech and sound effects.",
			Picture: Picture{
				Title:       "An AdLib Music Synthesizer ISA slot card",
				Jpg:         "adlib-card.jpg",
				Avif:        "adlib-card.avif",
				Attribution: "TheAlmightyGuru",
				License:     "GNU FDL",
				LicenseLink: "https://www.vgmpf.com/Wiki/index.php?title=File:AdLib_-_1987.jpg",
			},
		},
		{
			Year: 1987, Prefix: notable,
			List: Links{
				{
					LinkTitle: "Bentley Sidwell Productions", Link: "/g/bentley-sidwell-productions",
					SubTitle: "BSP", Forward: "Texas",
				},
				{
					LinkTitle: `Boys from Company C <small>(BCC)</small>`, Link: "/g/boys-from-company-c",
					Forward: "Virginia and D.C. region",
				},
				{
					LinkTitle: "Canadian Pirates Inc <small>(CPI)</small>", Link: "/g/canadian-pirates-inc",
					Forward: "🇨🇦 Ontario",
				},
				{
					LinkTitle: "-=C&M=-",
					Link:      "/g/c-ampersand-m",
					Forward:   "Maryland",
				},
				{
					LinkTitle: "KGB", Link: "/g/kgb",
					Forward: "🇨🇦 Ontario",
				},
				{
					LinkTitle: "The PTL Club", Link: "/g/ptl-club",
					Forward: "Illinois",
				},
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
			Content: "<p><a href=\"https://www.mobygames.com/game/4019/rockford-the-arcade-game/\">Rockford</a> is a strange game. " +
				"It is a port of the arcade game of the same name—a machine created as a port of the then-popular microcomputer video game series, " +
				"<a href=\"https://boulder-dash.com/history/\">Boulder Dash</a>.</p>" +
				"<p>More unusual is the use of <strong>32-color VGA</strong> for a home computer port of an arcade game on the PC in an era when ports were done on the cheap using the lowest common denominator four-color CGA graphics. " +
				"The crossover of players who owned expensive VGA graphic cards and monitors in 1988 who were playing arcade ports was low.</p>",
			Picture: Picture{
				Title:       "Rockford: The Arcade Game",
				Alt:         "Rockford: The Arcade Game screenshot",
				Png:         "rockford-32-color-vga.png",
				Avif:        "rockford-32-color-vga.avif",
				Attribution: "486pc & MobyGames",
				License:     "© Arcadia",
				LicenseLink: "https://www.mobygames.com/game/4019/rockford-the-arcade-game",
			},
		},
		{
			Title: "Earliest, standalone \"elite\" BBS ad", Year: 1988, Month: 4, Day: 4, Highlight: false,
			Lead: "So far, Swashbucklers II", LinkTitle: "the file",
			Link: "/f/b844ef",
			Content: "<p>While novel in 1988, <strong>BBS adverts</strong> like this <code>README.!!!</code> text file would plague releases as spam in the years to come, " +
				"with boards injecting numerous texts and tagging the releases with their names, often under the guise of documentation or readme texts.<br><br></p>" +
				`<pre style="font-size:28px;" class="font-dos-mda reader reader-invert border border-black rounded-1 p-1"><br>` +
				"Another Quality Ware Downloaded off:<br><br>" +
				"           S W A S H B U C K L E R S   I I<br>" +
				"                 Home of PTL/CPI<br>" +
				"                 100 megs Online!<br>" +
				"             85 megs Offline, Request!<br>" +
				"              All PTL/CPI Cracks FREE<br>" +
				"    All other Major Groups cracks Always Online<br>" +
				"       Ask your local Sysop for the number..<br>" +
				"We are a private system, but do accept the occasional new GOOD user. If " +
				"you have something to offer, call us. Once on, you won't have to call any " +
				"further.<br><br>" +
				"If all you want are the Latest warez FIRST call us we have them, or " +
				"we've just cracked them.<br><br></pre>",
		},
		{
			Title: "The earliest PC Scene drama", Year: 1988, Month: 11, Day: 25,
			Lead: "So far, TNWC accusing PTL of stealing a release", LinkTitle: "and view the file",
			Link: "/f/aa356d",
			Content: "<p>The earliest <strong>scene drama</strong> known so far involves a release by " +
				"<a href=\"/g/the-north-west-connection\">The&nbsp;North&nbsp;West&nbsp;Connection</a>&nbsp;(TNWC) for the game Paladin. " +
				"The drama accuses <a href=\"/g/ptl-club?\">PTL Club</a> of stealing and <q>re-releasing</q> an early game released by TNWC. " +
				"Scene drama often involves texts that call out other groups for poor behavior, breaking commonly accepted rules, or being <q>lame.</q></p>" +
				`<pre style="font-size:28px;" class="font-dos-mda reader reader-invert border border-black rounded-1 p-1"><br>` +
				"<p>DO NOT TAKE THIS FILE FROM THE ARCHIVE!!!!<br>" +
				"Well unlike PTL I won't sacrifice some game code to put up a fancy title screen for the group that released this (TNWC). " +
				"This is officially out third release, but really it's our second major one since PTL took Paladin and \"re-released\" it by taking off the doc check.<br>" +
				"Anyway - on with the game.  This game is a great role-playing game with some of the best graphics I've seen in an RPG (which is not what you'd expect from Infocom) so enjoy it.</p>" +
				"</pre>",
		},
		{
			Year: 1988, Prefix: notable,
			List: Links{
				{
					LinkTitle: "Crackers in Action", Link: "/g/crackers-in-action",
					SubTitle: "CIA", Forward: "Colorado",
				},
				{
					LinkTitle: "Future Brain Inc.", Link: "/g/future-brain-inc",
					SubTitle: "FBi", Forward: "🇳🇱 First Dutch group on the PC",
				},
				{
					LinkTitle: "Miami Cracking Machine", Link: "/g/miami-cracking-machine",
					SubTitle: "MCM", Forward: "Florida",
				},
				{LinkTitle: "Sprint", Link: "/g/sprint", Forward: "Ohio and 🇨🇦 Ontario"},
				{
					LinkTitle: "The Grand Council", Link: "/g/the-grand-council",
					SubTitle: "TGC", Forward: "Michigan",
				},
				{
					LinkTitle: "The North West Connection", Link: "/g/the-north-west-connection",
					SubTitle: "TNWC", Forward: "Washington",
				},
				{LinkTitle: "The Sysops Association Network", Link: "/g/the-sysops-association-network", SubTitle: "TSAN"},
			},
		},
		{
			Title: "The first 256 color VGA game", Year: 1989, Month: 3,
			Lead: "688 Attack Sub from Electronic Arts", LinkTitle: "the mobygames page",
			Link: "https://www.mobygames.com/game/2099/688-attack-sub",
			Content: "Driving, flying, and military simulation games were once a popular genre of video games on the PC. " +
				"Before dedicated <abbr title=\"graphics processing unit\">GPUs</abbr> existed, this genre created demanding open-world landscapes requiring expensive CPUs and even co-processors! " +
				"Which was great for those with high-end hardware who wanted to show off.",
			Picture: Picture{
				Title:       "688 Attack Sub",
				Alt:         "688 Attack Sub in-game screenshot",
				Png:         "688-attack-sub.png",
				Avif:        "688-attack-sub.avif",
				Attribution: "Defacto2",
				License:     "© Electronic Arts",
				LicenseLink: "https://www.mobygames.com/game/2099/688-attack-sub/screenshots/dos/9155/",
			},
		},
		{
			Title: "Earliest ANSI loader", Year: 1989, Month: 3,
			Lead: "So far, The Rogues Gallery", LinkTitle: "and view the loader",
			Link: "/f/ad21da8",
			Content: "<p><strong>ANSI loaders</strong> were text files with ASCII escape control characters to provide color and cursor movement. " +
				"However, a specific display driver on IBM and other PCs often needed to load at boot before viewing the texts. " +
				"So, to avoid this, Sceners converted their ANSI artworks into simple, self-displaying applications or <q>loaders.</q></p>" +
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
			Lead: "So far, First intro by Sorcerers", LinkTitle: "and run the intro",
			Link: "/f/ab2843",
			Content: "<p>An <strong>intro</strong>, or the later cracktro, is a small, usually short, demonstration program designed to display text with graphics or animations. " +
				"Oddly, the <q>First Intro</q> was written by some teenagers in Finland, a country not known for using expensive PC platforms.</p>" +
				"<p>Intros on the other popular 16-bit microcomputers had a higher creative expectation, with the machines offering much better graphics and audio capabilities than a common 1980's PC using a 4-color graphics adapter.</p>",
			Picture: Picture{
				Title: "First intro by Sorcerers",
				Alt:   "First intro by Sorcerers screenshot",
				Webp:  "ab2843.webp",
				Png:   "ab2843.png",
			},
		},
		{
			Title: "Earliest PC cracktro", Year: 1989, Month: 4, Day: 29, Highlight: true,
			Lead: "So far, Future Brain Inc.", LinkTitle: "and run the cracktro",
			Link: "/f/b83fd7",
			Content: "<p><strong>Future Brain Inc.</strong>, a group from the <strong>Netherlands</strong> that was among the first to release a cracktro on the PC platform, " +
				"released this for the game <a href=\"https://www.mobygames.com/game/2161/lombard-rac-rally/cover/group-99392/cover-270796/\">Lombard RAC Rally</a>.</p>" +
				"<p>Early cracktros on the PC lacked music and were usually a simple screen of text and a logo. " +
				"On other microcomputer platforms, the Commodore&nbsp;64, Amiga&nbsp;500, and Atari&nbsp;ST, cracktros offered music and graphic effects that were easier to create due to their unified hardware.</p>",
			Picture: Picture{
				Title: "Lombard RAC Rally cracktro",
				Alt:   "Lombard RAC Rally cracktro screenshot",
				Webp:  "b83fd7.webp",
				Png:   "b83fd7.png",
			},
		},
		{
			Title: "First issue of Pirate magazine", Year: 1989, Month: 6, Day: 1,
			Lead: "The earliest known scene newsletter for the Scene on the PC", LinkTitle: "the issues",
			Link: "/g/pirate",
			Content: "<p>Created in Chicago, Pirate magazine was a bi-monthly text newsletter for the Scene on the PC platform and distributed through bulletin boards. " +
				"It ran for at least five issues between June 1989 and April 1990.</p>" +
				`<pre style="font-size:28px;" class="font-dos-mda reader reader-invert border border-black rounded-1 p-1"><br>` +
				"<p>What's a pirate? COMPUTER PIRACY is copying and distribution of copyright software (warez). Pirates are hobbyists who enjoy collecting and playing with the latest programs. " +
				"Most pirates enjoy collecting warez, getting them running, and then generally archive them, or store them away. A PIRATE IS NOT A BOOTLEGGER. " +
				"Bootleggers are to piracy what a chop-shop is to a home auto mechanic. Bootleggers are people who DEAL stolen merchandise for personal gain. " +
				"Bootleggers are crooks. They sell stolen goods. Pirates are not crooks, and most pirates consider bootleggers to be lower life forms..." +
				"</pre>" +
				`<pre style="font-size:28px;" class="font-dos-mda reader reader-invert border border-black rounded-1 p-1"><br>` +
				"Pirates SHARE warez to learn, trade information, and have fun! But, being a pirate is more than swapping warez. It's a life style and a passion." +
				"<br><br></pre>",
		},
		{
			Year: 1989, Prefix: notable,
			List: Links{
				{
					LinkTitle: "Aces of ANSI Art", Link: "/g/aces-of-ansi-art", SubTitle: "AAA",
					Forward: "The beginning of The Art Scene",
				},
				{
					LinkTitle: "American Pirate Industries", Link: "/g/american-pirate-industries",
					SubTitle: "API", Forward: "California",
				},
				{
					LinkTitle: "Future Crew", Link: "/g/future-crew",
					SubTitle: "FC", Forward: "🇫🇮 The first mainstream PC group",
				},
				{
					LinkTitle: "International Network of Crackers", Link: "/g/international-network-of-crackers",
					SubTitle: "INC", Forward: "MCM, NYC, NCC",
				},
				{
					LinkTitle: "New York Crackers", Link: "/g/new-york-crackers",
					SubTitle: "NYC", Forward: "New York",
				},
				{
					LinkTitle: "Norwegian Cracking Company", Link: "/g/norwegian-cracking-company",
					SubTitle: "NCC", Forward: "🇳🇴 First Norwegian group on the PC",
				},
				{LinkTitle: "Pirates Sick of Initials", Link: "/g/pirates-sick-of-initials", SubTitle: "PSi"},
				{
					LinkTitle: "Sorcerers", Link: "/g/sorcerers",
					Forward: "🇫🇮 First PC demo group and Finnish group on the PC",
				},
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
			Title: "\".NFO\" file extension origins", Year: 1990, Month: 1, Day: 23, Highlight: true,
			Lead: "The Humble Guys", LinkTitle: "the list of THG releases",
			Link: "/g/the-humble-guys",
			Content: "" +
				// extension
				"<p>The <strong>.NFO</strong> file extension denotes a text file containing information about a release. " +
				"Still in use today, the dot nfo file contains information about the release group, the release itself, and how to install.</p>" +
				// thg
				"<p>While disputed, it is not too important which release from <strong>The&nbsp;Humble&nbsp;Guys</strong> is the first to use the dot <q>nfo</q> file extension. " +
				// knights
				"The timestamps of the release files and BBS tape backups suggest there were a number of THG game releases that predate Bubble Bobble by weeks. " +
				"But famed THG founder and former cracker, <a href=\"/p/fabulous-furlough\">Fabulous Furlough</a> has often stated Bubble Bobble was the release that first used the naming standard.</p>" +
				// quote
				"<figure><blockquote class=\"blockquote\"><q><small>It happened like this, I'd just used " +
				"<q><a href=\"http://nerdlypleasures.blogspot.com/2011/05/scourge-of-preservation-disk-based-copy.html\">Unguard</a></q> " +
				"to crack the SuperLock off of <a href=\"/f/ad4195\">Bubble&nbsp;Bobble</a>, and I said " +
				"<q>I need some file to put the info about the crack in. Hmmm.. Info, NFO!</q>, and that was it.</small></q></blockquote>" +
				"<figcaption class=\"blockquote-footer\">Founder of The&nbsp;Humble&nbsp;Guys, Fabulous&nbsp;Furlough recalls Bubble Bobble as the first THG release that used the .NFO file extension.</figcaption></figure>" +
				// bubble bobble
				"<p>Bubble Bobble was the more notable game of the period and may have been a more memorable game title when recalling the event.</p>" +
				`<pre style="font-size:1.5em;line-height:1em;" class="font-dos-mda reader reader-invert border border-black rounded-1 p-1">` +
				"KNIGHTS.NFO WHITEDET.NFO STUNT.NFO TRUMP.NFO DEJAVUII.NFO AJAX.NFO TERRAIN.NFO BUBBLE.NFO TREK.NFO CRMEWAVE.NFO STRIDER.NFO GUNBOAT.NFO 1989STAT.NFO ..." +
				"</pre>",
		},
		{
			Title: "Earliest PC cracktro with music", Year: 1990, Month: 12, Day: 2,
			Lead: "So far, The Cat's M1 Tank Platoon", LinkTitle: "about and view cractrko",
			Link: "/f/ab25f0e",
			Content: "<p>The Cat released this cracktro for the game <a class=\"text-nowrap\" href=\"https://www.mobygames.com/game/1499/m1-tank-platoon/cover/group-3004/cover-230986/\">M1 Tank Platoon</a>. " +
				"It is the first known cracktro on the PC platform to feature music. But music is in a loose sense, as it relies on the terrible internal PC speaker to produce the melody.</p>" +
				"<p>While 8-bit consoles and some microcomputers offered dedicated music audio chips, most famously the Commodore&nbsp;64 with its SID chip, the IBM&nbsp;PC, which targeted business, did not.</p>",
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
			Content: "<p>The <strong>Sound&nbsp;Blaster</strong> audio standard came about in 1990 after the Sound&nbsp;Blaster 1.5 audio card was released by Creative&nbsp;Labs, with the box proudly proclaiming" +
				" it <q><a href=\"https://vgmpf.com/Wiki/index.php?title=File:Sound_Blaster_1.5_-_Box_-_Back.jpg\">The PC Sound Standard</a></q>. " +
				"It was the first digital audio standard for the IBM&nbsp;PC to be widely adopted on the PC platform, despite its poor quality, mono 8-bit digital audio. " +
				"Previous audio standards such as the AdLib and the MT-32, were limited to FM synthesis or MIDI-like samples.</p>" +
				"<p>The Sound&nbsp;Blaster was the first audio standard widely adopted by the PC platform and was the de facto audio option in games for many years.</p>",
		},
		{
			Title: "CD-ROM multimedia", Year: 1990, Prefix: "Winter",
			Lead: "Mixed-Up Mother Goose", LinkTitle: "the catalog listing the game",
			Link: "https://archive.org/details/vgmuseum_sierra_sierra-90catalog-alt3/page/n21",
			Content: "<p>The first widely available enhanced PC game on <strong>CD-ROM</strong> was <a href=\"https://www.mocagh.org/sierra/mothergoose-alt-manual.pdf\">Mixed-Up Mother Goose</a>, announced by Sierra On-Line in 1990 and released in 1991. " +
				"The children's game was a high-technology remake of <a href=\"https://www.mobygames.com/game/758/mixed-up-mother-goose/cover/group-27001/cover-70129/\">a fun title</a> from 1987, but the CD-ROM remake featured new, enhanced VGA graphics and interface, digital audio with speech, singing, and music.</p>" +
				"<p>With the newest technology and a lack of standards for CD media, <a href=\"https://sierrachest.com/index.php?a=games&id=544&title=mother-goose-vga&fld=box&pid=3\">the box</a> " +
				"came with two identical discs, one red and one blue. " +
				"The red disc supported Red Book CD audio, while the blue disc supported lower-quality digital playback samples.</p>",
		},
		{
			Year: 1990, Prefix: notable,
			List: Links{
				{LinkTitle: "ANSI Creators in Demand", Link: "/g/acid-productions", SubTitle: "ACiD", Forward: "Aces of ANSI Art"},
				{LinkTitle: "Katharsis", Link: "/g/katharsis", Forward: "🇵🇱 First Polish group on the PC"},
				{
					LinkTitle: "National Elite Underground Alliance", Link: "/g/national-elite-underground-alliance",
					SubTitle: "NEUA", Forward: "New York",
				},
				{LinkTitle: "🇺🇸 Public Enemy", Link: "/g/public-enemy", SubTitle: "PE", Forward: "🇩🇪 First German PC group, Red Sector Inc."},
				{
					LinkTitle: "Software Chronicles Digest", Link: "/g/software-chronicles-digest",
					SubTitle: "SCD", Forward: "California",
				},
				{LinkTitle: "The Dream Team", Link: "/g/the-dream-team", SubTitle: "TDT", Forward: "🇸🇪 First PC group from Sweden"},
				{
					LinkTitle: "The Humble Guys", Link: "/g/the-humble-guys",
					SubTitle: "THG", Forward: "Tennessee",
				},
				{
					LinkTitle: "🇩🇪 Tristar & Red Sector Inc.", Link: "/g/tristar-ampersand-red-sector-inc",
					SubTitle: "TRSi", Forward: "🇩🇪 Red Sector, then in 1991 Skid Row, TDT",
				},
				{LinkTitle: "Ultra Tech", Link: "/g/ultra-tech", SubTitle: "UT"},
			},
		},
		{
			Title: "The first application and utility groups", Year: 1991, Month: 1, Highlight: true,
			Lead: "Nokturnal Trading Alliance and IUD",
			Content: "<p>The PC's first dedicated application and software utility groups emerged at the beginning of 1991. " +
				"Groups such as <a href=\"/g/nokturnal-trading-alliance\">Nokturnal Trading Alliance</a>, and later, <a href=\"/g/the-hill-people\">The Hill People</a> and " +
				"<a href=\"/g/inc-utility-division\">IUD</a> <em><a href=\"/g/international-network-of-crackers\">International Network of Crackers</a> Utility Division</em> start to package, " +
				"crack and exclusively release commercial applications, system utilities and productivity software.</p>" +
				"<p>Yet this form of software piracy <a href=\"f/ab25292\">dominated</a> the elite bulletin boards for the PC and had done so for a long while. " +
				"Typically, individuals compiled these \"app\" releases anonymously or for upload to their local bulletin boards instead of under a Scene group for competition. " +
				"Was this solo anonymity the legacy of do-it-yourself cracking and <a href=\"/files/how-to\">Unprotection documentation</a> common on the PC in the 1980s, or maybe a fear of big tech and their lawyers?</p>" +
				"<p>The most famous application group, <a href=\"/g/pirates-with-attitudes\">Pirates with Attitudes</a> (PWA), also was founded in 1991 but focused on game titles for their first two years.</p>" +
				"<p>A typical PC piracy BBS from the 1980s would mostly have system utilities and the occasional application uploaded with no individual or group credited and no additional help textfiles." +
				`<p><pre style="font-size:21px;" class="font-dos-mda reader reader-invert border border-black rounded-1 p-1">` +
				"IBMSPLIT.ARC   9200 01/05/89 Get WARPUTIL instead - handles MFM too!!\n" +
				"COPY606.ARC   28672 01/18/89 \n" +
				"NODMON25.ZIP  45028 01/18/89 \n" +
				"DSZREG.ARC     9216 02/12/89 Registers your DSZ. Press space when flashes\n" +
				"DSZ0223.ARC   81870 02/23/89 Latest DSZ\n" +
				"HELP33.ARC   140596 02/26/89 This is a nice utility to have around for | DOS 3.3\n" +
				"PRODOR29.ZIP 170833 03/01/89 \n" +
				"ARC601.EXE   138807 03/16/89 Latest vers. of IBM ARC - run to unpack..\n" +
				"PKZ092.EXE   102499 03/16/89 Latest vers. - run to unpack...\n" +
				"OPTUNE.ZIP    74741 03/17/89 OPTune Disk Optimizer From Gazelle Systems\n" +
				"PROD30B1.ZIP  88688 03/22/89 PCB PRODOOR V3.01B\n" +
				"COBOL.ZIP    163831 04/05/89 \n" +
				"AM42.ZIP     115180 04/13/89 Arcmaster 4.2\n" +
				"VARISLOW.ZIP   1922 04/20/89 Slow down the AT toplay games..\n" +
				"NORTCOM.ARC   54070 04/21/89 Norton Commander\n" +
				"TDRAW320.EXE 189659 04/27/89 \n" +
				"DRDOOM1.ZIP  269384 05/25/89  \n" +
				"DRDOOM2.ZIP  340992 05/25/89 " +
				"</pre></p>",
		},
		{
			Title: "Earliest BBS VGA loader", Year: 1991, Month: 3,
			Lead: "So far, XTC Systems BBS", LinkTitle: "the loader", Link: "/f/a41dcd9",
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
			Content: "This 1991 cracktro was released by a collaboration of " +
				"<a href=\"/g/the-dream-team\">The Dream Team</a> with <a href=\"/g/tristar-ampersand-red-sector-inc\">Tristar, and Red Sector Inc.</a>. " +
				"Dream Team founder <a href=\"/p/hard-core\">Hard Core</a> programmed it, which is the first known cracktro on the PC platform to feature a modern presentation with a logo, music, and a scroller. " +
				"Cracktros on the PC had previously been limited to primarily static logo screens or, in the case of the earliest cracktros, no graphics.",
			Picture: Picture{
				Title: "Blues Brothers cracktro",
				Alt:   "Blues Brothers cracktro screenshot",
				Avif:  "b249b1.avif",
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
			Title: "Earliest \"elite\" PC BBStro", Year: 1991, Month: 10, Day: 21,
			Lead: "So far, Splatterhouse BBS", LinkTitle: "about and view the BBStro", Link: "/f/b11acdf",
			Content: "<p><a href=\"https://demozoo.org/bbs/7179/\">Splatterhouse, or Splatter House</a>, was a San Jose, California bulletin board " +
				"heavily affiliated with the <a href=\"/g/international-network-of-crackers\">International Network of Crackers</a>, the art group <a href=\"/g/acid-productions\">ACiD Productions</a>, " +
				"and the designers of this <strong>BBStro</strong>, <a href=\"/g/insane-creators-enterprise\">Insane Creators Enterprise</a>.</p>" +
				"<p>While there were many earlier PC BBS ads, this was the first that combined music and animation.</p>",
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
				{LinkTitle: "HiPE", Link: "/g/hipe"},
				{LinkTitle: "Insane Creators Enterprise", Link: "/g/insane-creators-enterprise", SubTitle: "iCE"},
				{LinkTitle: "🇸🇪 Fairlight PC", Link: "/g/fairlight", SubTitle: "FLT"},
				{LinkTitle: "Licensed to Draw", Link: "/g/licensed-to-draw", SubTitle: "LTD", Forward: "DREAM"},
				{LinkTitle: "Mirage", Link: "/g/mirage", Forward: "Licensed to Draw"},
				{
					LinkTitle: "Nokturnal Trading Alliance", Link: "/g/nokturnal-trading-alliance",
					SubTitle: "NTA", Forward: "California",
				},
				{
					LinkTitle: "Pirates with Attitude", Link: "/g/pirates-with-attitude",
					SubTitle: "PWA", Forward: "Michigan and Minnesota",
				},
				{LinkTitle: "🇺🇸 Razor 1911 (on PC)", Link: "/g/razor-1911", SubTitle: "RZR", Forward: "🇳🇴 Razor / 🇪🇺 Skillion"},
				{LinkTitle: "Razor Dox", Link: "/g/razordox", SubTitle: "RZR"},
				{LinkTitle: "Relentless Pursuit of Magnificence", Link: "/g/relentless-pursuit-of-magnificence", SubTitle: "RPM"},
				{LinkTitle: "🇪🇺 Skid Row (on PC)", Link: "/g/skid-row", SubTitle: "SR"},
				{LinkTitle: "🇩🇪🇨🇭 Scoopex (IBM)", Link: "/g/scoopex"},
				{LinkTitle: "Silicon Dream Artists", Link: "/g/silicon-dream-artists", SubTitle: "SdA", Forward: "MAi / Maximized ANSi Designers"},
				{
					LinkTitle: "The Cracking Lords", Link: "/g/the-cracking-lords", SubTitle: "TCL",
					Forward: "🇮🇹 First PC group from Italy",
				},
				{LinkTitle: "The Humble Guys F/X", Link: "/g/thg-fx", SubTitle: "THG-FX"},
				{
					LinkTitle: "United Software Association", Link: "/g/united-software-association*fairlight",
					SubTitle: "USA", Forward: "The Humble Guys",
				},
			},
		},
		{
			Title: "Earliest CD release", Year: 1992, Month: 3, Day: 3, Highlight: true,
			Lead: "Battle Chess MPC", LinkTitle: "about the release", Link: "/f/aa209be",
			Content: "<p>The first known release of a game on CD was probably Battle Chess MPC (multimedia PC) released by International Network of Crackers on the 3rd of March 1992. " +
				"Being a novel medium for software distribution, the INC release was a mess requiring the user to have access to 28 floppy disks and then a third party tool to copy and \"splice\" the disks to a hard drive. " +
				"Copying to this many floppy disks for a single game would have been slow, tedious, and expensive, both in time and hardware.</p>" +
				"<p>Later in the month on the 22nd, " +
				"Razor 1911 would release Stellar 7 CD-ROM (now lost) that was reviewed in <a href=\"/f/b42bdee\">DMZ Review #4</a> and " +
				"$yndicate would release the " +
				`<a href="/f/b126bd6">CD ROM edition</a> ` + "of Wing Commander that didn't have complex installation process, and INC would attempt some other MPC titles. But in 1992, CD piracy didn't make sense or take off.</p>" +
				"<p>However in late 1994, scene personalities, The Renegade Chemist and Zeus would team up to form ROM 1911 : Razor 1911 CD-ROM Division. An early or possibly the first CD release from this pair was a game named " +
				`<a href="/f/ab3e0b">Slob Zone</a>` + ", an 8 floppy disk release. But because game publishers often didn't add copy protection on their CD titles, Razor 1911 didn't want any scene credit for the release.</p>",
		},
		{
			Title: "First SuperVGA / VESA game", Year: 1992, Month: 6,
			Lead: "Links 386 Pro", LinkTitle: "the mobygames page", Link: "https://www.mobygames.com/game/3757/links-386-pro/",
			Content: "<p>The first widely available <strong>SuperVGA</strong> game was Links 386 Pro from Access. Here, another popular golf simulation pushed the baseline PC gaming requirements with the need for higher-end hardware. " +
				"The 386 in the title stated the minimum requirement of an Intel&nbsp;386 CPU when 286 systems were the commodity.</p>" +
				"<p>The problem for consumers is that ordinarily, most PC software never took advantage of the enhancements offered by the more expensive Intel&nbsp;386 or 486 CPUs.</p>" +
				"<p>Some caveats to the first SVGA/VESA claim: we are talking about a retail, boxed game requiring a resolution/color depth that a standard VGA setup cannot handle, " +
				"so at least a constant 600x400 resolution with 256 colors.</p>",
			Picture: Picture{
				Title:       "Links 386 Pro",
				Alt:         "Links 386 Pro in-game screenshot",
				Png:         "links-386-pro-svga.png",
				Avif:        "links-386-pro-svga.avif",
				Attribution: "Servo & MobyGames",
				License:     "© Access Software",
				LicenseLink: "https://www.mobygames.com/game/3757/links-386-pro/",
			},
		},
		{
			Year: 1992, Prefix: notable,
			List: Links{
				{LinkTitle: "Artists in Revolt", Link: "/g/artists-in-revolt", Forward: "Fairlight"},
				{LinkTitle: "Damn Excellent Art Designers", Link: "/g/damn-excellent-ansi-design", SubTitle: "DeAD", Forward: "Damn Excellent ANSI Design"},
				{LinkTitle: "Graphics Rendered in Magnificence", Link: "/g/graphics-rendered-in-magnificence", SubTitle: "GRiM", Forward: "Silicon Dream Artists / NC-17"},
				{LinkTitle: "HYPE", Link: "/g/hype"},
				{LinkTitle: "Pyradical", Link: "/g/pyradical"},
				{LinkTitle: "🇩🇪 Superior Art Creations", Link: "/g/superior-art-creations", SubTitle: "SAC"},
				{
					LinkTitle: "The One and Only", Link: "/g/the-one-and-only",
					SubTitle: "TOAO", Forward: "New Jersey",
				},
			},
			Picture: Picture{
				Title: "The One and Only",
				Avif:  "b13a93.avif",
				Png:   "b13a93.png",
			},
		},
		{
			Year: 1993, Prefix: notable,
			List: Links{
				{LinkTitle: "🇷🇺 Drink or Die", Link: "/g/drink-or-die", SubTitle: "DOD"},
				{LinkTitle: "Hybrid", Link: "/g/hybrid", SubTitle: "HBD", Forward: "Pyradical"},
				{LinkTitle: "Legend", Link: "/g/legend", SubTitle: "LND"},
				{LinkTitle: "Paradox (on PC)", Link: "/g/paradox", SubTitle: "PDX"},
				{LinkTitle: "Pentagram", Link: "/g/pentagram", SubTitle: "PTG", Forward: "Legend"},
				{LinkTitle: "Rise in Superior Couriering", Link: "/g/rise-in-superior-couriering", SubTitle: "RiSC"},
				{
					LinkTitle: "Untouchables", Link: "/g/untouchables",
					SubTitle: "UNT", Forward: "UNiQ, XAP",
				},
			},
		},
		{
			Title: "First mention of \"CD-RIP\"", Year: 1994, Month: 9, Day: 4, Highlight: true,
			Lead: "So far, Hybrid", LinkTitle: "about the release", Link: "/f/ab27459",
			Content: "<p>A play on the media, CD-ROM, the earliest mention of <strong>CD-RIP</strong> (later simplified to <q>rip</q>) release, " +
				"was by Hybrid for the game Shanghai: Great Moments. " +
				"Hybrid was a group formed by ex-members of <a href=\"/g/pyradical\">Pyradical</a> and <a href=\"/g/pentagram\">Pentagram</a>.</p>" +
				"The <u>CD RIP</u> type came about due to CD-ROM-only games being unable to get a proper Scene release. For PC game publishers, " +
				"CD-ROMs were cheaper to produce and had far more storage capacity than the standard floppy disks. However, large hard drives were too expensive to store the content of complete CD images. " +
				"So, for many pirates to play a game published on CD, the disc's content had to be ripped and repackaged to a hard drive, but with the removal of the game's fluff, such as intro videos, music, and speech.",
		},
		{
			Title: "Copyright infringement legal precedent", Year: 1994, Month: 12, Day: 28, Highlight: true,
			Lead: "No criminal liability for the sharing of software", LinkTitle: "the David LaMacchia Defense Fund with press releases", Link: "https://web.archive.org/web/19990224000548/http://photo.net/dldf/home.html",
			Content: "<p>In April 1994, David LaMacchia, a 20-year-old junior at the Massachusetts Institute of Technology, was <a href=\"/f/b628640\">indicted</a> for conspiring to commit wire fraud. " +
				"A 1950s law intended to stop defrauding another out of money using the U.S. landline telephone network.</p>" +
				"<p>David ran two anonymous <a href=\"https://fsp.sourceforge.net/index.html\">File Service Protocol</a> sites using MIT's internal network connected to the Internet to share software with users without financial gain. " +
				"The primary site, <a href=\"https://web.archive.org/web/19991018194139/http://photo.net/dldf/indictment.html\">Cynosure</a>, offered downloads, while Cynosure II also permitted uploads with requests.</p>" +
				"<p>Months later, David's defense lawyers filed a motion to dismiss, " +
				"<q>LaMacchia contends that the indictment invents a criminal charge, primarily by distorting the wire fraud statute, in order to circumvent Congress's decision not to apply a criminal sanction to LaMacchia's alleged conduct.</q></p>" +
				"<p>And just days after Christmas, the motion to dismiss was allowed by District Judge Stearns.</p>" +
				"<p><q>The Court dismissed the indictment, holding that <u>there was no clearly expressed Congressional intent to permit prosecution of copyright infringement</u> " +
				"under the wire fraud statute. There was no allegation that LaMacchia infringed copyrighted software for commercial advantage or private financial gain.</q></p>",
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
				Title: "TDU-Jam! branding",
				Webp:  "af2b6a5.webp",
				Png:   "af2b6a5.png",
			},
		},
		{
			Title: "Windows 95 warez release", Year: 1995, Month: 8, Day: 12,
			Lead: "Drink or Die", Link: "/f/bb2b71f", LinkTitle: "about the release",
			Content: "<p><strong>Drink or Die</strong> became notorious for releasing the CD media for the box retail edition of <strong>Windows&nbsp;95</strong> " +
				"two weeks before the official worldwide release.</p>" +
				"<p>In an era when global, same-day product launches were logistically costly and uncommon, this operating system launch was probably the most hyped Microsoft consumer product ever. " +
				"Over a decade before Apple cemented the marketing tactic, Windows&nbsp;95 had fans <a href=\"https://rarehistoricalphotos.com/windows-95-launch-day-1995/\">queuing&nbsp;at&nbsp;midnight</a> in retail stores worldwide.</p>" +
				"<p>The release also highlighted a significant problem for software and game publishers: for pirates to get access to the retail packaging weeks before launch meant some company employees were either members of these warez groups or receiving kickbacks.</p>" +
				"<p>Years later, competitor <a href=\"/g/pirates-with-attitudes\">Pirates&nbsp;With&nbsp;Attitudes</a> would release the <a href=\"/f/a52a8c\" class=\"text-nowrap\">Windows 98 media</a> five weeks and <a href=\"/f/b42e2f6\">Windows&nbsp;2000</a> two months before the official launches! " +
				"However, a global, coordinated law enforcement effort would take down both groups in the following decade.<sup><a href=\"#windows-95-warez-release-fn1\">[1]</a></sup></p>" +
				"<p>The other Microsoft-sourced releases from DOD during these two weeks were the Windows&nbsp;95 <a href=\"/f/b82406f\" class=\"text-nowrap\">floppy edition</a>, <a href=\"/f/b721b5\" class=\"text-nowrap\">upgrade edition</a>, <a href=\"/f/b92697\" class=\"text-nowrap\">Plus Pack</a>, Microsoft <a href=\"/f/ba28e0f\">BOB</a>, and <a href=\"/f/bc2dc2f\">Word</a>.</p>" +
				sect0 +
				"<div id=\"windows-95-warez-release-fn1\">[1] In <a href=\"#the-copy-party-is-over\">Operation Cyberstrike</a> and <a href=\"#the-global-takedown-of-drink-or-die\">Operation Buccaneer</a>.</div>" +
				sect1,
		},
		{
			Title: "Windows 95", Year: 1995, Month: 8, Day: 24,
			Lead: "Worldwide retail release", LinkTitle: "about the day in history",
			Link: "https://www.theverge.com/21398999/windows-95-anniversary-release-date-history",
			Content: "<p>Microsoft's biggest and most hyped mainstream product release was hugely successful in the market and " +
				"finally began the PC's transition away from the archaic IBM and Microsoft DOS <small>(Disk&nbsp;Operating&nbsp;System)</small>.</p>" +
				"<p>Windows&nbsp;95 had been a long time coming, over a decade late, and offered a fully graphical user interface as the default. " +
				"It also introduced the famed <a href=\"https://arstechnica.com/gadgets/2015/08/the-windows-start-menu-saga-from-1993-to-today/\">Start menu</a> concept " +
				"that would later become favored by many Windows and, ironically, <a href=\"https://fossforce.com/2019/07/why-gnome-2-continues-to-win-the-desktop-popularity-contest/\">desktop</a> Linux users.</p>",
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
				{LinkTitle: "Hoodlum (on PC)", Link: "/g/hoodlum", SubTitle: "HLM"},
				{
					LinkTitle: "Prestige (on PC)", Link: "/g/prestige",
					SubTitle: "PTG", Forward: "Ohio and 🇳🇱 The Netherlands",
				},
				{LinkTitle: "Inquisition", Link: "/g/inquisition", SubTitle: "INQ", Forward: "Week in Warez"},
				{LinkTitle: "The Naked Truth", Link: "/g/the-naked-truth-magazine", SubTitle: "NTM"},
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
				"Before online or consumer digital transactions, buyers had to post the physical cash and an order form in an envelope to a PO Box in Florida.</p>" +
				"<p>The disc was controversial as <strong>reselling</strong> scene-released software was criminal and frowned upon. " +
				"But its success meant other group merchandise soon followed suit, with the most popular items being branded <a href=\"/f/b4484e\">t-shirts</a>. " +
				"Though the t-shirt merch was probably first introduced on the PC scene by <a href=\"/g/the-dream-team\">The Dream Team</a> with their 1992 <a href=\"/f/b32f0c\">T-Shirt Series #1</a>.</p>",
			Picture: Picture{
				Title: "Razor 1911 Tenth Anniversary CD-ROM",
				Alt:   "Razor 1911 Tenth Anniversary CD-ROM disc",
				Png:   "razor-1911-tenth-anniversary-cd-rom.png",
				Avif:  "razor-1911-tenth-anniversary-cd-rom.avif",
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
				"Later, cards from other manufacturers, such as the <a href=\"https://www.tomshardware.com/reviews/3d-accelerator-card-reviews,42-7.html\">Diamond Monster 3D</a>, quickly followed, and within a year, the 3Dfx chipset dominated the market.</p>" +
				"<p>Before the 3Dfx Voodoo release, consumer PCs' fragmented 3D graphics market needed more software support. " +
				"3Dfx coordinated with publishers to target their Glide API with new game releases so gamers had confidence in their Voodoo card purchases.</p>" +
				"<p>3Dfx also extended the life of existing PC hardware and broke the endless cycle of aggressive, expensive CPU upgrades to support the current generation of games. " +
				"A new 3Dfx card would double the resolution, add fantastic color support, and even improve the frames-per-second on what would otherwise be an older machine.</p>",
		},
		{
			Year: 1996, Prefix: notable,
			List: Links{
				//	{LinkTitle: "CD Images For the Elite", Link: "/g/cd-images-for-the-elite", SubTitle: "CiFE"},
				{LinkTitle: "Class", Link: "/g/class", SubTitle: "CLS", Forward: "Prestige"},
				{LinkTitle: "RomLight", Link: "/g/romlight", SubTitle: "RLT", Forward: "Fairlight"},
				{LinkTitle: "Paradigm", Link: "/g/paradigm", SubTitle: "PDM", Forward: "Eclipse"},
			},
		},
		{
			Title: "Release standards broken", Year: 1997, Month: 1, Day: 13,
			Lead: "Hybrid presents Diablo", LinkTitle: "the release", Link: "/f/ab49cd",
			Content: "<p>Founding member Hybrid is the first to break the CD-RIP standard <a href=\"/f/aa3b26\">rules</a> set by The Standards of Piracy Association with the release of " +
				"<a href=\"https://news.blizzard.com/en-us/diablo3/22887361/diablo-now-available-on-gog-com\">Diablo</a>.</p>" +
				"<p>Less than a year prior, SPA had agreed that CD-RIPs should be ripped to a maximum permitted size and titles that weren't possible should be skipped. " +
				"Release groups often passed over significant games such as Sierra's <a href=\"https://www.imdb.com/title/tt0131537/\">Phantasmagoria</a> due to their massive size and gameplay reliance on un-rippable video and audio content.</p>",
			List: Links{
				{LinkTitle: "Diablo from Razor 1911", Link: "/f/a72ced", SubTitle: "full CD rip"},
			},
		},
		{
			Title: "Earliest ISO release", Year: 1997, Month: 11, Day: 27, Highlight: true,
			Lead: "So far, CD Images For the Elite (CiFE)", LinkTitle: "the release", Link: "/f/ad40ce",
			Content: "An <a href=\"https://www.loc.gov/preservation/digital/formats/fdd/fdd000348.shtml\">ISO</a> is a standard file archive format containing the entire CD and later DVD data. It enables the copying and exact replication of data onto consumable blank discs. " +
				"Trading <strong>ISO images</strong> between individuals has happened for years prior, but <a href=\"https://www.mobygames.com/game/2082/lords-of-magic/covers/\">Lords of Magic</a> was the earliest known ISO release pushed to the Scene.</p>" +
				"<p>A formalization of an ISO trading scene occurred sometime in late 1997, but it took years before the medium became the dominant format in the Scene.</p>",
		},
		{
			Title: "Overnight, Warez becomes criminal", Year: 1997, Month: 12, Day: 16, Highlight: true,
			Lead: "No Electronic Theft (NET) Act", LinkTitle: "the Forbes article Stealing code", Link: "https://www.forbes.com/1997/11/28/feat.html?sh=5fc911fb2c1c",
			Content: "<p><q><strong>The NET Act</strong> was signed into law by President Clinton in December 1997, <u>making it illegal to reproduce or distribute copyrighted works</u>, such as software programs and musical recordings, even if the defendant acts without a commercial purpose or for private financial gain.</q></p>" +
				"<p>The law is a response to the failed prosecution against David LaMacchia from a few years earlier. The dismissal of his case brought attention to the legal anomaly named after his win, the <a href=\"https://www.newscientist.com/article/mg15621113-000-publish-on-the-net-and-be-damned/\">LaMacchia loophole</a>.</p>" +
				"<p>Under the new law, <q>if the defendant reproduces or distributes 10 or more copyrighted works that have a total value of more than $2,500, he or she can be charged with a felony, and faces a sentence of up to 3 years imprisonment and a fine of up to $250,000. " +
				"A defendant who reproduces or distributes one or more copies of copyrighted works with a value of more than $1,000 can be charged with a misdemeanor, and face up to one year in prison and a fine of up to $100,000.</q></p>",
		},
		{
			Year: 1997, Prefix: notable,
			List: Links{
				{LinkTitle: "CD Images For the Elite", Link: "/g/cd-images-for-the-elite", SubTitle: "CiFE"},
				{LinkTitle: "Divine", Link: "/g/divine", SubTitle: "DVN"},
			},
			Picture: Picture{
				Title: "Divine cracktro",
				Avif:  "a424a4c.avif",
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
				Jpg:         "starcraft-case.jpg",
				Avif:        "starcraft-case.avif",
			},
		},
		{
			Year: 1998, Month: 4, Day: 1,
			Title: "Starcraft", LinkTitle: "about the release", Link: "/f/a9353d",
			Lead: "Razor 1911",
			Content: "<p><a href=\"/g/razor-1911\">Razor 1911</a> and famed cracker <a href=\"/p/beowulf\">Beowulf</a> were credited with the release of StarCraft. " +
				"Together, they released a CD-RIP of the game. However, the package took time to compile and lacked the unique CD keys required to play the desirable online multiplayer.</p>" +
				"<p><q>Well, what can I say. This has got to be one of the hardest titles I have ever ripped. " +
				"The crack was trivial, but ripping this game involved understanding and coding utilities for Blizzard's file packer. It is ...a veritable nightmare.</q></p>",
			List: Links{
				{LinkTitle: "StarCraft Battle.NET Keymaker", Link: "/f/b321b00", SubTitle: "2 April"},
				{LinkTitle: `Starcraft *100% FIX*`, Link: "/f/b13d2c", SubTitle: "3 April"},
			},
			Picture: Picture{
				Title: "Razor 1911 Starcraft Broodwar cracktro",
				Avif:  "b22b15d.avif",
				Png:   "b22b15d.png",
			},
		},
		{
			Year:  1998,
			Title: "ISO scene picks up steam",
			Content: "<p>The <strong>ISO scene</strong> is still in its infancy but snowballs after some top groups start releasing with the file format.</p>" +
				"<p>Some key events of 1998.</p>" +
				ul0 +
				"<li>Razor 1911 merges the separate <a href=\"/f/a82c49\">ISO division</a> back into the Razor 1911 label.</li>" +
				"<li><a href=\"/f/ac2be5\">Fairlight returns</a> after 4-years and is exclusively released with the format.</li>" +
				"<li>The famed couriers RiSC created <a href=\"/f/b04dac\">RiSCiSO</a> to become one of the largest ISO release groups.</li>" +
				"<li><a href=\"/f/b3202e0\">PDM ISO</a> is the ISO division of <a href=\"/g/paradigm\">Paradigm</a> and Zeus.</li>" +
				"<li><a href=\"/g/deviance\">DVNiSO</a> is the ISO division of Divine.</li>" +
				"<li><a href=\"/f/a94b94\">SHOCKiSO</a> is the ISO division of Shock.</li>" +
				ul1 +
				"<p>Other early users of the format include " +
				"<a class=\"text-nowrap\" href=\"/g/cd-images-for-the-elite\">CD Images for the Elite</a> (CiFE), " +
				"<a href=\"/g/kalisto\">Kalisto</a>, <a href=\"/g/isolation\">ISOlation</a>, " +
				"<a class=\"text-nowrap\" href=\"/g/in-search-of-cd\">In Search of CD</a>, and CaLiSO.</p>" +
				"<p><q>Paradigm - we do rips, we do ISO - we do it all with style</q></p>",
		},
		{
			Year: 1998, Prefix: notable,
			List: Links{
				{LinkTitle: "DVNiSO / Deviance", Link: "/g/deviance"},
				{LinkTitle: "Fairlight", Link: "/g/fairlight", SubTitle: "FTL", Forward: "Fairlight returns after a few years absent"},
				{LinkTitle: "Origin", Link: "/g/origin", SubTitle: "OGN"},
				{LinkTitle: "RiSCiSO", Link: "/g/risciso", Forward: "Rise in Superior Couriering"},
			},
		},
		{
			Year:  1999,
			Title: "3Dfx vs. Nvidia", LinkTitle: "a short story of 3dfx - 5 steps to fall", Link: "https://level2.vc/a-short-story-of-3dfx/",
			Lead: "1999 was a complex year for PC gamers",
			Content: "<p>The market pioneer, 3Dfx, with its Voodoo 3 GPU, had abandoned OEM manufacturers and decided to produce both the chips and graphic boards in-house. " +
				"The change, intended to boost profits, led to manufacturing and global distribution shortages and decreased retail shelf space for 3Dfx products.</p>" +
				"<p>In the same year, Nvidia released its TNT and <strong>GeForce series</strong> of GPUs and became the go-to supplier of chips for OEM card manufacturers. " +
				"Unlike 3Dfx, Nvidia was API agnostic and happy to prioritize Direct3D and OpenGL.</p>" +
				"<p>For gamers, the new 3Dfx cards were more challenging to obtain but offered the best compatibility for 3D games of the past few years. " +
				"Plus, current games ran fast with better frames per second.</p>" +
				"<p>The high-end Nvidia products offered improved resolutions and graphic feature sets but poorer compatibility for older games developed primarily for the proprietary 3Dfx Glide API. " +
				"But by the end of 2000, 3Dfx was bankrupt, having taken on too much debt and railroaded themselves into a dead-end architecture. <a href=\"https://www.cnet.com/culture/nvidia-buys-out-3dfx-graphics-chip-business\">By April 2002</a>, the company's assets and intellectual property were owned by Nvidia.</p>",
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
				Avif:  "a92f47.avif",
			},
		},
		{
			Title: "The giveaway safe habor is over", Year: 2000, Month: 5, Day: 5, Highlight: true,
			Lead: "The end of Pirates with Attitude", LinkTitle: "the US DOJ press release", Link: "https://web.archive.org/web/20120114174415/http://www.justice.gov/criminal/cybercrime/pirates.htm",
			Content: "<p>The US Department of Justice <strong>indicted 17 members</strong> of <a href=\"/g/pirates-with-attitudes\">Pirates&nbsp;with&nbsp;Attitudes</a> " +
				"who got caught up in a honey pot scheme where, for months, Canadian law enforcement had taken control of the primary " +
				"PWA FTP distribution site, Sentinel, running out of the University of Sherbrooke in Quebec. " +
				"A day later, PWA published its <a href=\"/f/a23b69\">final release</a>, a farewell NFO by the fugitive Shiffie out of Belgium.</p>" +
				"<p>Of the US defendants, 13 pleaded guilty. Five members were active employees of Intel Corp, and one was an employee of Microsoft Corp. " +
				"Less than a week later, Christian Morley, aka <q>Mercy,</q> a former senior organizer of PWA, became the first person to be " +
				"<a href=\"https://ipmall.law.unh.edu/sites/default/files/hosted_resources/CyberCrime/pwa_verdict.pdf\">found guilty</a> under the No Electronic Theft (NET) Act and " +
				"the first to be guilty of <u>conspiracy to infringe software copyrights</u>.</p>" +
				`<pre style="font-size:1.5em;line-height:1em;" class="font-dos-mda reader reader-invert border border-black rounded-1 p-1">` +
				"                         PWA Sites<sup><a href=\"#the-copy-party-is-over-fn1\">[1]</a></sup><br>" +
				"┌──────────────────────┬─────────────────────┬────────────┐<br>" +
				"│ FTP Site Names       │ Status ············ │ SiteOP ··· │<br>" +
				"├──────────────────────┼─────────────────────┼────────────┤<br>" +
				"│ Sentinel ··········  │ World HQ ·········· │ Xxxxxxx ·· │<br>" +
				"│ The Rock      ·····  │ US HQ ············· │ Xxxxxxx ·· │<br>" +
				"│ Major Malfunction ·  │ EURO HQ ··········· │ Xxxxxxx ·· │<br>" +
				"│ MidNite Resistence·  │ World Courier HQ ·· │ Xxxxxxx ·· │<br>" + //nolint:misspell
				"</pre>" +
				sect0 +
				"<div id=\"the-copy-party-is-over-fn1\">[1] Unlike other groups, PWA were still <a href=\"/f/ac38f0\">advertising</a> their sites in 1999.</div>" +
				sect1,
			//
		},
		{
			Title: "Direct3D, the 3D graphic standard", Year: 2000, Month: 11,
			Lead: "DirectX 8.0 (4.08.00.0400)", LinkTitle: "the press release", Link: "https://news.microsoft.com/2000/11/09/microsoft-announces-release-of-directx-8-0",
			Content: "<p>The release of Microsoft's <strong>Direct3D</strong> 8 for all active editions of Windows from 95 through to XP was the beginning of the dominance " +
				"of the proprietary 3D graphics API, as it is the first release offering compelling features for game developers.</p>" +
				"<p>For Microsoft, this helps to lock in Windows as the only operating system for modern PC gaming. " +
				"Since 1996, prior editions of Direct3D have been clumsy and lacking features compared to the competing proprietary 3Dfx Glide or the industry OpenGL standard. " +
				"Direct3D was instead a hardware fallback API for developers to support.</p>",
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
			Title: "The global takedown of Drink or Die", Year: 2001, Month: 12, Day: 11, Highlight: true,
			Lead: "Operation Buccaneer", LinkTitle: "the Department of Justice overview", Link: "https://www.ojp.gov/ncjrs/virtual-library/abstracts/cross-national-investigation-and-prosecution-intellectual-property",
			Content: "<p><strong>Operation Buccaneer</strong> was the first global effort to target a specific warez group for criminal prosecution. " +
				"Because of the nature of warez groups on the Internet, members operate in numerous countries and time zones. " +
				"The operation needed multiple law enforcement agencies in the USA, UK, Australia, Norway, Sweden, and Finland to coordinate the execution " +
				"of <strong>70&nbsp;search&nbsp;warrants</strong>.</p>" +
				"<p>The target, <a href=\"/g/drink-or-die\">Drink&nbsp;or&nbsp;Die</a> is singled out for its multiple pirate releases of " +
				"Microsoft&nbsp;<a href=\"/f/bb2b71f\">Windows&nbsp;95</a> that occurred back in August 1995, over six years prior with a different group membership.</p>",
		},
		{
			Title: "Digital only scene releases", Year: 2004, Month: 10, Day: 7,
			Lead: "Counter-Strike: Source Final from Emporio", LinkTitle: "the release", Link: "/f/b1282e1",
			Content: "<p>Counter-Strike <a href=\"https://www.mobygames.com/game/15518/counter-strike-source/cover/group-80271/cover-733563/\">Source</a>, the online multiplayer title, was exclusively distributed on <a href=\"https://steampowered.com\">Steam</a>, Valve's digital distribution platform. " +
				"As no physical media was available, this became a dubious release within the Scene, and many groups didn't acknowledge the <a href=\"/g/emporio\">Emporio</a> package as a legitimate <q>retail</q> " +
				"product or a <q>final</q> release. The release of Steam-only games was poorly received due to the ease of supply and constant online patching.</p>" +
				"<p><q>SOME may contend the fact that this is BETA. This is the version that is released on <a href=\"https://web.archive.org/web/20050208205808/http://www.steampowered.com/index.php?area=news&archive=yes&id=327\">STEAM AS FINAL</a>. " +
				"You cannot do any better than this. The ... thing with STEAM is they can easily release many patches BUT EXPECT the EMPORiO crew to bring each and every patch CRACKED to your doorstep!</q></p>",
		},
		{
			Title: "Digital distribution and online activation", Year: 2004, Month: 11, Day: 16,
			Lead: "Half-Life 2", LinkTitle: "the and view the Steam page", Link: "https://store.steampowered.com/app/220/HalfLife_2",
			Content: "<p>Half-Life 2 was one of the most anticipated games of the decade, and it was the first major game to use <a href=\"https://steampowered.com\">Steam</a>, " +
				"Valve's digital distribution platform. Steam was a massive shift in how games got distributed, and it was the first time a significant game required online activation. " +
				"Steam often was not well received by the gaming <a href=\"https://www.amazon.com/product-reviews/B00006I02Z/ref=acr_dp_hist_1?ie=UTF8&filterByStar=one_star&reviewerType=all_reviews#reviews-filter-bar\">community</a>, " +
				"but it was a big success for Valve and paved the way for other digital distribution platforms. " +
				"Half-Life 2 was released simultaneously on <a href=\"https://store.steampowered.com/app/220/HalfLife_2/\">Steam</a>, " +
				"<a href=\"https://www.mobygames.com/game/15564/half-life-2/cover/group-90348/cover-246334/\">DVD</a>, " +
				"and <a href=\"https://www.mobygames.com/game/15564/half-life-2/cover/group-16318/cover-38738/\">CD</a>, but all three formats required Steam activation.</p>",
		},
		{
			Title: "Half-Life 2 *Retail*", Year: 2004, Month: 11, Day: 28,
			Lead: "Vengeance", LinkTitle: "the release", Link: "/f/b24c10",
			Content: "<p>Half-Life 2 was one of the most anticipated games of the decade, and it was the first major game to use Steam, Valve's digital distribution platform.</p>" +
				"<p><a href=\"/g/vengeance\">Vengeance</a> is the first attempt to crack the Steam activation, and it used an unusual Steam client and activation emulator. " +
				"But while playable, their pirate release of the game suffered with slower frame rates, load times, and the lack of multiplayer gameplay. " +
				"Vengeance would release the DVD *Retail* version with a tweaked crack two days later.</p>",
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
			Content: "<p>Farewell © Myth is the final release from Myth, a group founded as <a href=\"/f/a53f3c\">Zeus</a>, " +
				"then <a href=\"/g/paradigm\">Paradigm</a> in 1996 and focused on ripping PC games from CD and later DVDs. " +
				"By the mid-2000s, broadband use was widespread, and the desire for ripped CD or DVD games with missing content was dwindling. " +
				"Myth's longtime rival, Class, had already <a href=\"/f/a53505\">quit</a> in early 2004, and the other major competitor, " +
				"<a href=\"/g/divine\">Divine</a>, finished up the following year.</p>" +
				"<p><q>We believe that the rip scene is one of incredible skill. " +
				"Not only is there the cracking talent needed to be successful like that of ISO, you must have dedicated coders and rippers to fully complete the task. " +
				"Much time is needed to perfect a rip like that of Neverwinter Nights. (We'll never forget you old friend) With the faster speed of the internet, " +
				"equates to less usage of rips and just makes it not worth it. " +
				"When you are releasing upwards of 30 games a month some months, and you know not many people are downloading them. " +
				"It hardly gives you the rush of winning the title. We see groups throw out games now with stolen cracks and are completely non-working. " +
				"These titles are not nuked, as no one even notices anymore, indeed a sad time in the scene.</q></p>",
		},
		{
			Year: 2016, Prefix: "", Highlight: true,
			Title: "The twilight of the cracktro",
			Content: "<p>The first decade of the 2000s was the last time original-quality cracktros were common in the Scene, " +
				"primarily thanks to a few nostalgic Demosceners and warez crackers. " +
				"However, the number of people who could and were willing to create a decent cracktro dwindled as the skillset requirements got more specific and complex. " +
				"So, cracktros were often forsaken by more straightforward methods of displaying the release information and branding.</p>",
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
				Avif:  "b230776.avif",
				Png:   "b230776.png",
			},
		},
	}
	return m
}

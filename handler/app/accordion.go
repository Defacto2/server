package app

// Package file website.go contains the website list and handler.

// Accordion is a collection of websites grouped by a category.
type Accordion = []struct {
	Name  string // Name of the category.
	Title string // Title tab for the category.
	ID    string // ID of the category.
	Sites Sites  // Sites are the websites shown in the category.
	Open  bool   // Whether the category is displayed or closed.
}

// Sites is a collection of websites.
type Sites = []Site

// Site is a website.
type Site struct {
	Title string // Title of the website.
	URL   string // URL of the website, including the HTTP or HTTPS protocol.
	Info  string // A short description of the website.
}

// List is a collection of websites grouped by a category.
func List() Accordion {
	return Accordion{
		{
			"Text art scene", "Text art pages",
			"text", art(), false,
		},
		{
			"Bulletin Board System", "BBS pages",
			"bbs", bbses(), false,
		},
		{
			"Crack and pirate scene", "Pirate pages",
			"pirate", pir8(), false,
		},
		{
			"The demoscene", "Demoscene pages",
			"demo", demos(), false,
		},
		{
			"Former groups", "Pages of former groups",
			"exgroup", groups(), false,
		},
		{
			"YouTube videos", "YouTube",
			"youtube", youtube(), false,
		},
		{
			"Podcasts", "Podcasts",
			"podcast", podcasts(), false,
		},
		{
			"Books", "Books",
			"book", books(), false,
		},
		{
			"Videos series and documentaries", "Videos",
			"video", videos(), false,
		},
		{
			"AmA ~ ask me anything", "AMAs",
			"ama", ama(), false,
		},
	}
}

// youtube returns a list of YouTube videos.
func youtube() []Site {
	return Sites{
		Site{
			"I was a video game software pirate", "https://www.youtube.com/watch?v=ockNRSt3Nsk",
			"Modern Vintage Gamer talks about his past as a video game pirate on the Commodore microcomputers.",
		},
		Site{
			"Clever floppy disk anti-piracy", "https://www.youtube.com/watch?v=VheNpiSZxf0",
			"Modern Vintage Gamer covers the clever copy protection of Dungeon Master.",
		},
		Site{
			"Dongle anti-piracy fail", "https://www.youtube.com/watch?v=W1cryx7TzqM",
			"Modern Vintage Gamer on the hardware copy protection of Ocean's Robocop 3 for the Commodore Amiga that was broken by Fairlight.",
		},
		Site{
			"Codewheels, early Anti-Piracy that was easy to bypass", "https://www.youtube.com/watch?v=S_Tz0YpDa6o",
			"Modern Vintage Gamer covers the gimmic, paper codewheel copy protection of the 1980s.",
		},
		Site{
			"Datel Action Replay", "https://www.youtube.com/watch?v=WH3ja70_okA",
			"Modern Vintage Gamer covers the Datel Action Replay, a cheat device that was also used to bypass copy protection on the Commodore Amiga.",
		},
		Site{
			"SecureROM", "https://www.youtube.com/watch?v=u8ltfyqD3lM",
			"Modern Vintage Gamer explores the SecureROM copy protection that was implemented onto PC CD-ROM games from 1997.",
		},
		Site{
			"LensLock DRM", "https://www.youtube.com/watch?v=Wpn9sLNg-6k",
			"Modern Vintage Gamer covers the LensLock, an odd, platic hardware copy protection device that was used on the Commodore 64.",
		},
		Site{
			"StarForce", "https://www.youtube.com/watch?v=p-wyIalhdPU",
			"Modern Vintage Gamer talks about the StarForce copy protection that was used on PC CD-ROM games from the mid 2000s.",
		},
		Site{
			"History of Denuvo", "https://www.youtube.com/watch?v=y_6zYVcJIKM",
			"Modern Vintage Gamer covers the Denuvo copy protection, the DRM for DRMs.",
		},
		Site{
			"How cracking groups ripped original Xbox discs", "https://www.youtube.com/watch?v=uY8KNl88Lqc",
			"Modern Vintage Gamer explores the Xbox 360 piracy scene that used PC DVD drives to rip game discs.",
		},
		Site{
			"Rockstar Games busted selling cracked versions of their games", "https://www.youtube.com/watch?v=XEKPUARYckc",
			"Modern Vintage Gamer discusses Rockstar Games using the No CD Crack by Razor 1911.",
		},
		Site{
			"History of DRM & copy protection in computer games", "https://www.youtube.com/watch?v=HjEbpMgiL7U",
			"Lazy Game Reviews covers the history of copy protection in computer games.",
		},
		Site{
			"The lost art of video game anti-piracy", "https://www.youtube.com/watch?v=ha7w96FQ-y4",
			"nimk covers the anti-piracy ingame death and warning screens both modern and old.",
		},
		Site{
			"Anti-piracy screens are unnerving. But, why?", "https://www.youtube.com/watch?v=dL9gUli_7L0",
			"Gearisko asks why anti-piracy screens console games are so off-putting.",
		},
		Site{
			"10 punishments for video game piracy", "https://www.youtube.com/watch?v=6avtHAmz6js",
			"gameranx shows 10 examples where game developers trolled the pirates.",
		},
		Site{
			"The horror of anti-piracy screens", "https://www.youtube.com/watch?v=wUUZFu0YmXw",
			"lzzzyzzz dives into the bizzarre world of video game anti-piracy screens.",
		},
		Site{
			"7 crazy ways video game pirates were punished", "https://www.youtube.com/watch?v=FVVx5WYmO2s",
			"Grunge lists 7 examples of how game developers punished the pirates in creative ways.",
		},
		Site{
			"The Amiga's hidden and funny developer messages", "https://www.youtube.com/watch?v=xXYBrvKEXKw",
			"Kim Justice covers the hidden and funny developer messages in Commodore Amiga games.",
		},
		Site{
			"The greatest video game pirate of all time", "https://www.youtube.com/watch?v=ZUioVa-wdDk",
			"Przle covers the story of the controversial cracker EMPRESS.",
		},
		Site{
			"How old school computers and games work", "https://www.youtube.com/playlist?list=PLfABUWdDse7bfBp4HvkN_RSKdXygMO71Z",
			"A playlist from the 8-bit Guy that covers the technical side of microcomputer hardware and software.",
		},
		Site{
			"Back to the BBS - The Underground", "https://www.youtube.com/watch?v=z_heZ-lgzq0",
			"Al's Geek Lab covers the warez and HPAVCC bulletin board system scene.",
		},
		Site{
			"Bulletin Board System (BBS) - The Internet's first community", "https://www.youtube.com/watch?v=I18ifd8I6P8",
			"Off the Cuff interviews Jason Scott on the origins of the BBS.",
		},
		Site{
			"Game piracy explained", "https://www.youtube.com/watch?v=8uUJFvSkTfI",
			"Overlord Gaming offers a primer of Internet software piracy of the 2010s.",
		},
	}
}

// ama is a collection of ask me anything posts.
func ama() []Site {
	return Sites{
		Site{
			"Evil Current <sup>2012</sup>", "https://www.reddit.com/r/IAmA/comments/xusji/iama_former_member_of_razor_1911_amongst_many/",
			"The former head member of Razor 1911, and ex-member of The Cartel, Drink or Die, SCuM, Tyranny, Napalm, Pirates with Attitudes, and others.",
		},
		Site{
			"The Playboy <sup>2010</sup>", "https://www.reddit.com/r/IAmA/comments/ckobg/iama_exbbs_warez_scene_guy_ama/",
			"I worked for a major computer game company and was a member of Razor1911. I supplied and was responsible for releasing MAJOR releases such as C&C Red Alert, Z, and others.",
		},
		Site{
			"BiGrAr <sup>2004</sup>", "https://yro.slashdot.org/story/02/10/04/144217/former-drinkordie-member-chris-tresco-answers",
			"Slashdot AMA with ex Drink Or Die member, taken after being busted in Operation Buccaneer but before serving a 33 month jail sentence.",
		},
		Site{
			"ex-MP3 scener <sup>2010</sup>", "https://www.reddit.com/r/IAmA/comments/c451i/iama_ex_warez_scene_member_ama/",
			"I didn't have involvement with cracking etc. and was involved in the far less glamorous MP3 side of the scene, which in a way I'd consider to be a now redundant section.",
		},
		Site{
			"ex-DVD scener <sup>2009</sup>", "https://www.reddit.com/r/IAmA/comments/9l1j3/iama_former_distributor_of_warez_on_the_top_level/",
			"I was a member of the DVD ripping scene for a few years.",
		},
	}
}

// art is a collection of text art websites.
func art() []Site {
	return Sites{
		Site{
			"16colors", "https://16colo.rs/",
			"You're looking at retro computer graphics gallery. We make ANSI/ASCII art available for web display.",
		},
		Site{
			"aSCII aRENA", "https://www.asciiarena.se/",
			"aSCII aRENA is a website dedicated to the art scene and the artists who create it.",
		},
		Site{
			"Blocktronics", "http://blocktronics.org/",
			"We're an international network of digital textmode artists that releases ANSi artpacks periodically.",
		},
		Site{
			"Art Scene", "http://artscene.textfiles.com/",
			"The textfiles.com computer art collection.",
		},
	}
}

// bbses is a collection of BBS websites.
func bbses() []Site {
	return Sites{
		Site{
			"BBS Ads Collection", "https://mbox.bz/slurp/ascii/bbsads/",
			"One of the most complete BBS textmode ad collections, containing over 1,500 single ads from various platforms and scenes.",
		},
		Site{
			"Break Into Chat", "https://breakintochat.com/blog/",
			"Break Into Chat is a blog about BBS history, retro computing and technology reminiscences.",
		},
		Site{
			"BBS Documentary", "http://www.bbsdocumentary.com/",
			"Jason Scott's documentary about the history of the BBS.",
		},
		Site{
			"The BBS Archives", "https://archives.thebbs.org/",
			"The BBS Archives is a collection of BBS files from the 80's and 90's.",
		},
		Site{
			"The BBS Software Directory", "http://www.bbsdocumentary.com/software/",
			"A canonical list of BBS software packages for all platforms.",
		},
	}
}

// books is a collection of books about the scene.
func books() []Site {
	return Sites{
		Site{
			"The Modem World <sup>2022</sup>", "https://yalebooks.yale.edu/book/9780300248142/modem-world/",
			"The Modem World is the first book to chronicle the history of the social, political, and technical changes wrought by the invention of the modem.",
		},
		Site{
			"Exploding the Phone <sup>2013</sup>", "http://explodingthephone.com/",
			"Before smartphones, back even before the Internet and personal computer, a misfit group of technophiles, blind teenagers, hippies, and outlaws figured out how to hack the world's largest machine: the telephone system.",
		},
		Site{
			"Warez: The Infrastructure and Aesthetics of Piracy <sup>2021</sup>", "https://punctumbooks.pubpub.org/pub/m5fu2twe",
			"Is the first scholarly research book about this underground subculture, which began life in the pre-internet era Bulletin Board Systems and moved to internet File Transfer Protocol servers (“topsites”) in the mid- to late-1990s.",
		},
	}
}

// demos is a collection of demoscene websites.
func demos() []Site {
	return Sites{
		Site{
			"Demozoo", "https://demozoo.org/",
			"Demozoo is the database of the demoscene.",
		},
		Site{
			"Pouët", "https://www.pouet.net/",
			"Pouët is a website dedicated to the art of demoscene and the demoscene culture.",
		},
		Site{"Scene.org", "https://www.scene.org/", ""},
		Site{
			"Demoscene Documentary", "https://www.youtube.com/user/demoscenedoc",
			"A ten episode documentary about the Finnish demoscene, subtitled in English.",
		},
		Site{
			"DSP", "https://www.docsnyderspage.com/",
			"Commodore 64 cracker intros in your browser.",
		},
		Site{
			"Flashtro", "http://www.flashtro.com/",
			"Amiga cracktros in your browser.",
		},
		Site{
			"GDI Mayhem", "https://gdimayhem.com/",
			"Curated GDI/OpenGL/D3D effects from the PC cracking scene.",
		},
		Site{
			"The Hornet Archive", "https://hornet.org/",
			"Digital art, rendered in realtime, from the dawn of the PC era.",
		},
		Site{
			"Scenery", "https://www.exotica.org.uk/wiki/Scenery",
			"Scenery is the guide to the C64 and Amiga demoscenes with comprehensive information on releases, parties and groups.",
		},
	}
}

// groups is a collection of scene group websites.
func groups() []Site {
	return Sites{
		Site{
			"ACiD Productions", "https://www.acid.org/",
			"The pioneering ANSI art group.",
		},
		Site{
			"iCE", "https://www.ice.org/",
			"The other pioneering ANSI art group.",
		},
		Site{
			"Deviance Demo", "http://deviance.untergrund.net/",
			"The former demo-division of the Deviance pirate group.",
		},
		Site{
			"Quartex", "https://www.quartex.org/",
			"The Amiga and console pirate group, with a site under construction since 2001.",
		},
		Site{
			"Scoopex", "http://www.scoopex1988.org/",
			"The Amiga pirate and demo group.",
		},
		Site{
			"Razor 1911", "https://www.razor1911.com/",
			"The famed pirate and demo group, the website is down as of 2023.",
		},
		Site{
			"Titan", "http://www.titancrew.org/",
			"The active, multi-platform demo group and former cracktro producers.",
		},
		Site{
			"TRIAD", "https://www.triad.se/",
			"The Commodore 64 demo and cracking group, active since 1986.",
		},
		Site{
			"TRSi", "http://www.trsi.org/",
			"The Amiga and PC demo and cracking group.",
		},
	}
}

// pir8 is a collection of pirate websites.
func pir8() []Site {
	return Sites{
		Site{
			"SCiZE's classic collection", "https://scenelist.org/",
			"Browse and search through thousands of 1990s PC releases from the BBS and early internet era.",
		},
		Site{
			"replacementdocs", "http://www.replacementdocs.com/",
			"The original web archive of game manuals for all platforms.",
		},
		Site{
			"RECOLLECTION", "http://www.atlantis-prophecy.org/recollection",
			"Stories from the Commodore 64 underground scenes.",
		},
		Site{
			"mp3Scene", "https://mp3scene.info/",
			"An archive of the MP3 scene.",
		},
		Site{
			"GameCopyWorld", "https://gamecopyworld.com/",
			"An ad heavy website but with a massive, historical collection of game trainers and fixes.",
		},
		Site{
			"ReScene", "https://rescene.wikidot.com/",
			"ReScene is a mechanism for backing up and restoring the metadata from Scene RAR and music releases.",
		},
	}
}

// podcasts returns a list of podcasts.
func podcasts() []Site {
	return Sites{
		Site{
			"Modem Mischief Podcast <sup>2021 - ongoing</sup>", "https://modemmischief.com/",
			"Modem Mischief is a true cybercrime podcast.",
		},
		Site{
			"Apple II pirate lore <sup>2003</sup>", "https://archive.org/details/Apple-II-Pirate-Lore",
			"Overview of the Apple II Piracy Community of the early to mid 1980's, presented by historian Jason Scott at the 5th Rubi-Con conference.",
		},
		Site{
			"Open Apple #66:Glenda Adams <sup>2016</sup>", "https://www.open-apple.net/2016/12/28/show-066-glenda-the-atom-adams-software-piracy/",
			"Glenda Adams, also known as The Atom, was a cracker of some note back in the 1980s, and she shares great stories with us of her exploits in boot tracing, cracking, and distributing software in the glory days of the Apple II BBS scene.",
		},
		Site{
			"100 Years of the Computer Art Scene <sup>2004</sup>", "https://archive.org/details/notacon-artscene-2004-04-24",
			"Historian Jason Scott and ACiD founder RaD Man capture 100 years of computer art, the magic of the art scene, the demo scene, and a dozen other 'scenes' that have been with us as long as computers have.",
		},
	}
}

// videos returns a list of videos and films.
func videos() []Site {
	return Sites{
		Site{
			"Steal This Film <sup>2006</sup>", "https://stealthisfilm.com/",
			"Steal This Film is a series documenting the movement against intellectual property and released via BitTorrent.",
		},
		Site{
			"You're Stealing It Wrong <sup>2010</sup>", "https://vimeo.com/15400820",
			"Historian Jason Scott walks through the many-years story of software piracy and touches on the tired debates before going into a completely different direction - the interesting, informative, hilarious and occasionally obscene world of inter-pirate-group battles.",
		},
		Site{
			"Good Copy Bad Copy <sup>2007</sup>", "https://www.youtube.com/watch?v=ByY6j0qzOyM",
			"Good Copy Bad Copy is a documentary about the state of copyright and culture.",
		},
		Site{
			"No Copy <sup>2008</sup>", "https://www.youtube.com/watch?v=BXBqUBAv1ek",
			"A promotional movie for the book No Copy about copyright, warez and media.",
		},
		Site{
			"TPB AFK - The Pirate Bay Away From Keyboard <sup>2013</sup>", "https://www.youtube.com/watch?v=eTOKXCEwo_8",
			"TPB AFK is a documentary about the founders of the Pirate Bay, subtitled in English.",
		},
		Site{
			"The Scene <sup>2004</sup>", "https://www.youtube.com/watch?v=1ZKBCA6PQ_g",
			"Welcome to the Scene is a 20 part web series about people in the online movie piracy scene.",
		},
		Site{
			"Teh Scene <sup>2005</sup>", "https://archive.org/search?query=%22Teh%20Scene%22%20AND%20collection%3Acomputersandtechvideos",
			"Teh Scene is a parody of the online movie piracy scene.",
		},
	}
}

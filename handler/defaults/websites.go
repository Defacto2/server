package defaults

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Site is a website.
type Site struct {
	Title string // Title of the website.
	URL   string // URL of the website, including the HTTP or HTTPS protocol.
	Info  string // A short description of the website.
}

// Sites is a collection of websites.
type Sites = []Site

// Accordion is a collection of websites grouped by a category.
type Accordion = []struct {
	Name  string // Name of the category.
	ID    string // ID of the category.
	Open  bool   // Whether the category is displayed or closed.
	Sites Sites  // Sites are the websites shown in the category.
}

// List is a collection of websites grouped by a category.
func List() Accordion {
	return Accordion{
		{"The text art scene", "text", false, art()},
		{"Bulletin Board Systems", "bbs", false, bbs()},
		{"Crack and pirate scenes", "pirate", false, pir8()},
		{"The demoscene", "demo", false, demo()},
		{"Former groups", "exgroup", false, groups()},
		{"Podcasts", "podcast", false, podcasts()},
		{"Books", "book", false, books()},
		{"Videos and documentary", "video", false, video()},
		{"ama ~ ask me anything", "ama", false, ama()},
	}
}

// ama is a collection of ask me anything posts.
func ama() []Site {
	return Sites{
		Site{"Evil Current", "https://www.reddit.com/r/IAmA/comments/xusji/iama_former_member_of_razor_1911_amongst_many/",
			"August 2012, with a former head member of Razor 1911, and ex-member of The Cartel, Drink or Die, SCuM, Tyranny, Napalm, Pirates with Attitudes, and others."},
		Site{"The Playboy", "https://www.reddit.com/r/IAmA/comments/ckobg/iama_exbbs_warez_scene_guy_ama/",
			"July 2010, I worked for a major computer game company and was a member of Razor1911. I supplied and was responsible for releasing MAJOR releases such as C&C Red Alert, Z, and others."},
		Site{"BiGrAr", "https://yro.slashdot.org/story/02/10/04/144217/former-drinkordie-member-chris-tresco-answers",
			"October 2004, Slashdot AMA with ex Drink Or Die member, taken after being busted in Operation Buccaneer but before serving a 33 month jail sentence."},
		Site{"ex-MP3 scener", "https://www.reddit.com/r/IAmA/comments/c451i/iama_ex_warez_scene_member_ama/",
			"May 2010, I didn't have involvement with cracking etc. and was involved in the far less glamorous MP3 side of the scene, which in a way I'd consider to be a now redundant section."},
		Site{"ex-DVD scener", "https://www.reddit.com/r/IAmA/comments/9l1j3/iama_former_distributor_of_warez_on_the_top_level/",
			"September 2009, I was a member of the DVD ripping scene for a few years.",
		},
	}
}

// art is a collection of text art websites.
func art() []Site {
	return Sites{
		Site{"16colors", "https://16colo.rs/",
			"You're looking at retro computer graphics gallery. We make ANSI/ASCII art available for web display."},
		Site{"aSCII aRENA", "https://www.asciiarena.se/",
			"aSCII aRENA is a website dedicated to the art scene and the artists who create it."},
		Site{"Blocktronics", "http://blocktronics.org/",
			"We're an international network of digital textmode artists that releases ANSi artpacks periodically."},
		Site{"Art Scene", "http://artscene.textfiles.com/",
			"The textfiles.com computer art collection."},
	}
}

// bbs is a collection of BBS websites.
func bbs() []Site {
	return Sites{
		Site{"BBS Ads Collection", "https://mbox.bz/slurp/ascii/bbsads/",
			"One of the most complete BBS textmode ad collections, containing over 1,500 single ads from various platforms and scenes."},
		Site{"Break Into Chat", "https://breakintochat.com/blog/",
			"Break Into Chat is a blog about BBS history, retro computing and technology reminiscences."},
		Site{"BBS Documentary", "http://www.bbsdocumentary.com/",
			"Jason Scott's documentary about the history of the BBS."},
		Site{"The BBS Archives", "https://archives.thebbs.org/",
			"The BBS Archives is a collection of BBS files from the 80's and 90's."},
		Site{"The BBS Software Directory", "http://www.bbsdocumentary.com/software/",
			"A canonical list of BBS software packages for all platforms."},
	}
}

// books is a collection of books about the scene.
func books() []Site {
	return Sites{
		Site{"The Modem World", "https://yalebooks.yale.edu/book/9780300248142/modem-world/",
			"The Modem World is the first book to chronicle the history of the social, political, and technical changes wrought by the invention of the modem."},
		Site{"Exploding the Phone", "http://explodingthephone.com/",
			"Before smartphones, back even before the Internet and personal computer, a misfit group of technophiles, blind teenagers, hippies, and outlaws figured out how to hack the world's largest machine: the telephone system."},
	}
}

// demo is a collection of demoscene websites.
func demo() []Site {
	return Sites{
		Site{"Demozoo", "https://demozoo.org/",
			"Demozoo is the database of the demoscene."},
		Site{"Pouët", "https://www.pouet.net/",
			"Pouët is a website dedicated to the art of demoscene and the demoscene culture."},
		Site{"Scene.org", "https://www.scene.org/", ""},
		Site{"Demoscene Documentary", "https://www.youtube.com/user/demoscenedoc",
			"A ten episode documentary about the Finnish demoscene, subtitled in English."},
		Site{"DSP", "https://www.docsnyderspage.com/",
			"Commodore 64 cracker intros in your browser."},
		Site{"Flashtro", "http://www.flashtro.com/",
			"Amiga cracktros in your browser."},
		Site{"GDI Mayhem", "https://gdimayhem.com/",
			"Curated GDI/OpenGL/D3D effects from the PC cracking scene."},
		Site{"The Hornet Archive", "https://hornet.org/",
			"Digital art, rendered in realtime, from the dawn of the PC era."},
		Site{"Scenery", "https://www.exotica.org.uk/wiki/Scenery",
			"Scenery is the guide to the C64 and Amiga demoscenes with comprehensive information on releases, parties and groups."},
	}
}

// groups is a collection of scene group websites.
func groups() []Site {
	return Sites{
		Site{"ACiD Productions", "https://www.acid.org/",
			"The pioneering ANSI art group."},
		Site{"iCE", "https://www.ice.org/",
			"The other pioneering ANSI art group."},
		Site{"Deviance Demo", "http://deviance.untergrund.net/",
			"The former demo-division of the Deviance pirate group."},
		Site{"Fairlight", "http://www.fairlight.fi/",
			"The multi-platform pirate and demo group."},
		Site{"Quartex", "https://www.quartex.org/",
			"The Amiga and console pirate group, with a site under construction since 2001."},
		Site{"Scoopex", "http://www.scoopex1988.org/",
			"The Amiga pirate and demo group."},
		Site{"Razor 1911", "https://www.razor1911.com/",
			"The famed pirate and demo group, the website is down as of 2023."},
		Site{"Titan", "http://www.titancrew.org/",
			"The active, multi-platform demo group and former cracktro producers."},
		Site{"TRIAD", "https://www.triad.se/",
			"The Commodore 64 demo and cracking group, active since 1986."},
		Site{"TRSi", "http://www.trsi.org/",
			"The Amiga and PC demo and cracking group."},
	}
}

// pir8 is a collection of pirate websites.
func pir8() []Site {
	return Sites{
		Site{"Scize	classic collection", "https://scenelist.org/",
			"Browse and search through 1000s PC releases of the 1990s from the BBS and early internet era."},
		Site{"replacementdocs", "http://www.replacementdocs.com/",
			"The original web archive of game manuals for all platforms."},
		Site{"RECOLLECTION", "http://www.atlantis-prophecy.org/recollection",
			"Stories from the Commodore 64 underground scenes."},
		Site{"mp3Scene", "https://mp3scene.info/",
			"An archive of the MP3 scene."},
		Site{"GameCopyWorld", "https://gamecopyworld.com/",
			"An ad heavy website but with a massive, historical collection of game trainers and fixes."},
		Site{"ReScene", "https://rescene.wikidot.com/",
			"ReScene is a mechanism for backing up and restoring the metadata from Scene RAR and music releases."},
	}
}

// podcasts returns a list of podcasts.
func podcasts() []Site {
	return Sites{
		Site{"Modem Mischief Podcast", "https://modemmischief.com/",
			"Modem Mischief is a true cybercrime podcast."},
		Site{"Apple II pirate lore", "https://archive.org/details/Apple-II-Pirate-Lore",
			"Overview of the Apple II Piracy Community of the early to mid 1980's, presented by historian Jason Scott at the 5th Rubi-Con conference."},
		Site{"Open Apple #66:Glenda Adams", "https://www.open-apple.net/2016/12/28/show-066-glenda-the-atom-adams-software-piracy/",
			"Glenda Adams, also known as The Atom, was a cracker of some note back in the 1980s, and she shares great stories with us of her exploits in boot tracing, cracking, and distributing software in the glory days of the Apple II BBS scene."},
		Site{"100 Years of the Computer Art Scene", "https://archive.org/details/notacon-artscene-2004-04-24",
			"Historian Jason Scott and ACiD founder RaD Man capture 100 years of computer art, the magic of the art scene, the demo scene, and a dozen other 'scenes' that have been with us as long as computers have."},
	}
}

// video returns a list of videos and films.
func video() []Site {
	return Sites{
		Site{"Steal This Film", "https://stealthisfilm.com/",
			"Steal This Film and is a 2006-7 film series documenting the movement against intellectual property and released via BitTorrent."},
		Site{"You're Stealing It Wrong", "https://vimeo.com/15400820",
			"Historian Jason Scott walks through the many-years story of software piracy and touches on the tired debates before going into a completely different direction - the interesting, informative, hilarious and occasionally obscene world of inter-pirate-group battles."},
		Site{"Good Copy Bad Copy", "https://www.youtube.com/watch?v=ByY6j0qzOyM",
			"Good Copy Bad Copy is a 2007 documentary about the state of copyright and culture."},
		Site{"No Copy", "https://www.youtube.com/watch?v=BXBqUBAv1ek",
			"A promotional movie for the book No Copy about copywrite, warez and media."},
		Site{"TPB AFK - The Pirate Bay Away From Keyboard", "https://www.youtube.com/watch?v=eTOKXCEwo_8",
			"TPB AFK is a documentary about the founders of the Pirate Bay, subtitled in English."},
		Site{"The Scene", "https://www.youtube.com/watch?v=1ZKBCA6PQ_g",
			"Welcome to the Scene is a 20 part, 2004 web series about people in the online movie piracy scene."},
		Site{"Teh Scene", "https://archive.org/search?query=%22Teh%20Scene%22%20AND%20collection%3Acomputersandtechvideos",
			"Teh Scene is 2005 parody of the online movie piracy scene."},
	}
}

// Websites is the handler for the websites page.
// Open is the ID of the accordion section to open.
func Websites(s *zap.SugaredLogger, ctx echo.Context, open string) error {
	data := initData()
	data["title"] = "Websites"
	data["logo"] = "Websites, podcasts, videos, books and films"
	data["description"] = "A collection of websites, podcasts, videos, books and films about the scene."
	acc := List()

	// Open the accordion section.
	closeAll := true
	for i, site := range acc {
		if site.ID == open || open == "" {
			site.Open = true
			data["title"] = site.Name
			closeAll = false
			acc[i] = site
			if open == "" {
				continue
			}
			break
		}
	}
	// If a section was requested but not found, return a 404.
	if open != "hide" && closeAll {
		return echo.NewHTTPError(http.StatusNotFound, ErrTmpl)
	}

	// Render the page.
	data["accordion"] = acc
	err := ctx.Render(http.StatusOK, "websites", data)
	if err != nil {
		s.Errorf("%s: %s", ErrTmpl, err)
		return echo.NewHTTPError(http.StatusInternalServerError, ErrTmpl)
	}
	return nil
}

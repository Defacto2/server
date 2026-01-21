// Package site proves links and titles for recommended websites.
package site

import "sort"

// URI is the URL slug of the releaser.
type URI string

// Website is the historical URL of the releaser website.
type Website struct {
	// URL of the website, not working sites should exclude the protocol, e.g. www.example.com.
	// While working sites MUST include the protocol, e.g. https://www.example.com.
	URL string
	// Name of the website.
	Name string
	// NotWorking will not apply a hyperlink to the URL.
	NotWorking bool
}

// Groups is a map of releasers URIs mapped to their websites.
type Groups map[URI][]Website

// Websites returns a map of releasers URIs mapped to their websites.
//
//   - URL should be the absolute URL of the website including the required protocol.
//     But non-working sites should exclude the protocol, e.g. www.example.com.
//   - Name should be the name of the website.
//   - NotWorking will not apply a hyperlink to the URL.
func Websites() Groups {
	return Groups{
		"shade": {
			{URL: "www.suburbia.net/~shade", NotWorking: true},
		},
		"infinite-darkness-bbs": {
			{URL: "infidark.nws.net", NotWorking: true, Name: "former Telnet"},
		},
		"entropy-bbs": {
			{URL: "entropybbs.net", NotWorking: true, Name: "Telnet board"},
		},
		"sanctuary-bbs": {
			{URL: "https://www.brysk.se/sanctuary/index.htm", Name: "Connect to Sanctuary, former Fairlight HQ"},
		},
		"myth": []Website{
			{URL: "www.myth.org", NotWorking: true, Name: "Mentioned in 2000"},
		},
		"delusions-of-grandeur": []Website{
			{URL: "delusions.base.org", NotWorking: true},
		},
		"celebre": []Website{
			{URL: "www.celebre.net", NotWorking: true},
		},
		"dextrose": []Website{
			{URL: "www.dextrose.com", NotWorking: true},
			{URL: "https://web.archive.org/web/19980131194050/http://www.dextrose.com/", Name: "1998 mirror"},
		},
		"crc": []Website{{URL: "www.bgnett.no/~xbone", NotWorking: true}},
		"digital-corruption": []Website{
			{URL: "dc.denet.co.jp", NotWorking: true},
			{URL: "https://web.archive.org/web/19971224223854/http://dc.denet.co.jp/", Name: "1997 mirror"},
		},
		"fantastic-4-cracking-group": []Website{
			{URL: "www.f4cg.com", Name: "advertised in 1998", NotWorking: true},
			{
				URL:  "https://web.archive.org/web/20001204211300/http://www.f4cg.com/",
				Name: "f4cg.com mirror in 2000, used by the Hitmen",
			},
		},
		"trc-ware-report": []Website{
			{URL: "http://falcon.laker.net/chemist", Name: "1997 ad in CLASS.NFO", NotWorking: true},
		},
		"the-flame-arrows": []Website{
			{URL: "www.tfa.org", Name: "TFA", NotWorking: true},
			{URL: "www.euronet.nl/users/jdm/documents/members.html", Name: "The Flame Arrows", NotWorking: true},
			{URL: "https://web.archive.org/web/19990117024946/http://www.tfa.org/", Name: "TFA 1999 mirror"},
			{
				URL:  "https://web.archive.org/web/20000829080106/http://www.euronet.nl/users/jdm/documents/members.html",
				Name: "The Flame Arrows mirror",
			},
		},
		"byte-enforcerz": []Website{
			{URL: "www.ussuriisk.ru/~beg/", Name: "BEG", NotWorking: true},
		},
		"bios-systems": []Website{
			{URL: "www.bios-systems.yi.org", Name: "BIOS Systems", NotWorking: true},
		},
		"esprit-couriers": []Website{
			{URL: "www.esprit.org", Name: "Esprit", NotWorking: true},
		},
		"sea-shell-commando": []Website{
			{URL: "www.sedona.net/~weirdo", Name: "SSC application", NotWorking: true},
		},
		"breed": []Website{
			{URL: "opal.utu.fi/~ternur/breed/index.html", Name: "Breed", NotWorking: true},
		},
		"dead-pirates-society": []Website{
			{
				URL:        "www.cinecan.com/dps",
				Name:       "DSP by |Swine|",
				NotWorking: true,
			},
		},
		"paradox": []Website{
			{
				URL:        "www.paradogs.com",
				Name:       "Paradox",
				NotWorking: true,
			},
			{
				URL:  "https://en.wikipedia.org/wiki/Paradox_%28warez%29",
				Name: "Wikipedia - Paradox",
			},
		},
		"light-speed-warez": []Website{
			{
				URL:        "www.lsw.org",
				Name:       "Light Speed Warez",
				NotWorking: true,
			},
		},
		"high-society": []Website{
			{
				URL:        "www.high-society.org",
				Name:       "High Society",
				NotWorking: true,
			},
			{
				URL:  "https://www.high-society.at",
				Name: "High Society",
			},
		},
		"future-crew": []Website{
			{
				URL:  "https://en.wikipedia.org/wiki/Future_Crew",
				Name: "Wikipedia - Future Crew",
			},
			{
				URL:        "www.futurecrew.com",
				Name:       "Future Crew",
				NotWorking: true,
			},
		},
		"eagle-soft-incorporated": []Website{
			{
				URL:  "https://csdb.dk/group/?id=696",
				Name: "Eagle Soft Incorporated on CSDb",
			},
		},
		"myth-inc": []Website{
			{
				URL:  "https://demozoo.org/bbs/12549",
				Name: "Myth Inc BBS on Demozoo",
			},
		},
		"legion-of-doom": []Website{
			{
				URL:  "https://en.wikipedia.org/wiki/Legion_of_Doom_(hacker_group)",
				Name: "Wikipedia - Legion of Doom (hacker group)",
			},
			{
				URL:  "http://textfiles.com/magazines/LOD/",
				Name: "The Legion of Doom/Hackers Technical Journal",
			},
		},
		"the-acquisition": []Website{
			{
				URL:  "http://artscene.textfiles.com/acid/ARTPACKS/",
				Name: "ACiD Art Packs",
			},
		},
		"acid-productions": []Website{
			{
				URL:  "http://artscene.textfiles.com/acid/",
				Name: "The ACiD Collection",
			},
			{
				URL:  "https://www.acid.org",
				Name: "1996 ACiD webpage",
			},
			{
				URL:        "http://www.cyberspace.com/~aciddraw",
				Name:       "Original webpage",
				NotWorking: true,
			},
			{
				URL:  "https://en.wikipedia.org/wiki/ACiD_Productions",
				Name: "Wikipedia",
			},
			{
				URL:  "https://www.youtube.com/watch?v=oQrBbm5ZMlo",
				Name: "BBS The Documentary: Episode 5: Artscene",
			},
			{
				URL:  "https://archive.org/details/bbs-20020727-radman",
				Name: "Interview: RaD Man/ACiD",
			},
			{
				URL:  "https://archive.org/details/20040308-bbs-tracer",
				Name: "Interview: Tracer/ACiD",
			},
			{
				URL:  "https://archive.org/details/bbs-20030520-jed",
				Name: "Interview: JED/ACiD",
			},
		},
		"assault": []Website{
			{
				URL:        "www.nrg2000.com",
				Name:       "Assault",
				NotWorking: true,
			},
		},
		"chemical-reaction": []Website{
			{
				URL:        "www.creaction.com",
				Name:       "Chemical Reaction",
				NotWorking: true,
			},
		},
		"core": []Website{
			{
				URL:        "coremongos.home.ml.org",
				Name:       "CORE",
				NotWorking: true,
			},
		},
		"defacto2": []Website{
			{
				URL:  "https://defacto2.net",
				Name: "Defacto2",
			},
			{
				URL:  "https://wayback.defacto2.net/defacto2-from-2000-july-11/",
				Name: "from July 2000",
			},
			{
				URL:  "https://wayback.defacto2.net/defacto2-from-1999-september-26/",
				Name: "from September 1999",
			},
			{
				URL:  "https://wayback.defacto2.net/defacto2-from-1998-september-8/",
				Name: "from September 1998",
			},
			{
				URL:        "www.defacto2.com",
				Name:       "launch address",
				NotWorking: true,
			},
		},
		"defacto": []Website{
			{
				URL:        "www.jicom.jinr.ru/sodom",
				Name:       "Defacto",
				NotWorking: true,
			},
		},
		"deviance": []Website{
			{
				URL:  "https://deviance.untergrund.net",
				Name: "Deviance Demo Division",
			},
		},
		"devotion": []Website{
			{
				URL:        "www.devotion.pp.se",
				Name:       "Devotion",
				NotWorking: true,
			},
			{
				URL:        "www.dataplus.se/devotion",
				Name:       "by Strooper and Spy",
				NotWorking: true,
			},
		},
		"divine": []Website{
			{URL: "dvn.org", Name: "Divine", NotWorking: true},
			{URL: "www.divinegods.com", Name: "Divine Gods", NotWorking: true},
		},
		"drink-or-die": []Website{
			{
				URL:        "www.drinkordie.com",
				Name:       "Drink Or Die",
				NotWorking: true,
			},
			{
				URL:        "spl.co.il/zino",
				Name:       "INQ ad",
				NotWorking: true,
			},
		},
		"empress": []Website{
			{
				URL:  "https://www.reddit.com/r/HobbyDrama/comments/rowk83/digital_piracy_the_rise_of_empress_how_one_woman/",
				Name: "The rise of EMPRESS",
			},
			{
				URL:  "https://www.wired.com/story/empress-drm-cracking-denuvo-video-game-piracy/",
				Name: "WIRED interview",
			},
			{
				URL:        "www.reddit.com/r/EmpressEvolution",
				Name:       "EmpressEvolution",
				NotWorking: true,
			},
		},
		"fairlight": []Website{
			{
				URL:  "https://www.fairlight.to",
				Name: "Fairlight Commodore 64",
			},
			{
				URL:  "https://www.fairlight.fi",
				Name: "Fairlight Finland",
			},
			{
				URL:        "www.fairlight.org",
				Name:       "Fairlight",
				NotWorking: true,
			},
			{
				URL:        "www.ludd.luth.se/~watchman/fairlight",
				Name:       "Fairlight Sweden",
				NotWorking: true,
			},
			{
				URL:  "https://web.archive.org/web/19981201194626/http://www.fairlight.org/",
				Name: "1997 mirror",
			},
		},
		"fire-site-ftp": []Website{
			{
				URL:        "firesite.ml.org",
				Name:       "Fire Site FTP",
				NotWorking: true,
			},
		},
		"gorgeous-ladies-of-warez": []Website{
			{
				URL:        "www.glow.org",
				Name:       "Gorgeous Ladies Of Warez",
				NotWorking: true,
			},
		},
		"just-the-facts": []Website{
			{
				URL:        "www.mygale.org/~jtf98",
				Name:       "Just The Facts",
				NotWorking: true,
			},
			{
				URL:        "www.multimania.com/jtf98/",
				Name:       "Just The Facts",
				NotWorking: true,
			},
			{
				URL:  "https://web.archive.org/web/20010223130305/http://www.multimania.com/jtf98/index.html",
				Name: "JTF mirror",
			},
		},
		"hybrid": []Website{
			{
				URL:        "www.hybridism.com",
				Name:       "Hybrid 1998",
				NotWorking: true,
			},
			{
				URL:        "www.hybrid.to",
				Name:       "Hybrid",
				NotWorking: true,
			},
			{
				URL:        "www.hybreed.com",
				Name:       "Hybreed",
				NotWorking: true,
			},
			{
				URL:        "www.tripnet.se/~electro/home",
				Name:       "Hoson in INQ",
				NotWorking: true,
			},
		},
		"hybrid-christmas-e_mag": []Website{
			{
				URL:        "www.hybreed.com",
				Name:       "Hybreed",
				NotWorking: true,
			},
		},
		"insane-creators-enterprise": []Website{
			{
				URL:  "https://www.ice.org",
				Name: "iCE Advertisements",
			}, {
				URL:  "http://artscene.textfiles.com/ice",
				Name: "The iCE Collection",
			}, {
				URL:  "https://en.wikipedia.org/wiki/ICE_Advertisements",
				Name: "Wikipedia",
			}, {
				URL:  "https://www.youtube.com/watch?v=oQrBbm5ZMlo",
				Name: "BBS The Documentary: Episode 5: Artscene",
			},
		},
		"level4": []Website{
			{
				URL:        "www.level4.ml.org",
				Name:       "Level 4",
				NotWorking: true,
			},
		},
		"motiv8": []Website{
			{
				URL:        "www.motiv8.org",
				Name:       "Motiv8",
				NotWorking: true,
			},
			{
				URL:        "hipsworld.bridge.net/~tribal",
				Name:       "June 1996",
				NotWorking: true,
			},
		},
		"paradigm": []Website{
			{
				URL:        "www.pdmworld.com",
				Name:       "Paradigm",
				NotWorking: true,
			},
			{URL: "www.pdm97.com", NotWorking: true},
			{URL: "www.paradigm.org", NotWorking: true},
			{
				URL:        "www.pdmworld.com/dac",
				Name:       "DAC Paradigm art",
				NotWorking: true,
			},
		},
		"phrozen-crew": []Website{
			{
				URL:        "www.phrozencrew.org",
				Name:       "Phrozen Crew",
				NotWorking: true,
			},
		},
		"premiere": []Website{
			{
				URL:        "premiere97.com",
				Name:       "Premiere 97",
				NotWorking: true,
			},
			{
				URL:        "premiere.ttlc.net",
				Name:       "Premiere",
				NotWorking: true,
			},
		},
		"prestige": []Website{
			{
				URL:        "www.laker.net/prestige",
				Name:       "Prestige",
				NotWorking: true,
			},
			{
				URL:        "www.jet.laker.net/prestige",
				Name:       "June 1996",
				NotWorking: true,
			},
		},
		"quartex": []Website{
			{
				URL:  "https://www.quartex.org",
				Name: "Quartex",
			},
			{
				URL:        "www.quartex.demon.co.uk",
				Name:       "Quartex",
				NotWorking: true,
			},
		},
		"razor-1911": []Website{
			{
				URL:        "https://www.razor1911.com",
				Name:       "Razor 1911",
				NotWorking: false,
			},
			{
				URL:        "www.razor-1911.com",
				Name:       "Razor 1911",
				NotWorking: true,
			},
			{
				URL:        "www.laric.com/razor",
				Name:       "Razor 1911 Europe",
				NotWorking: true,
			},
			{
				URL:        "ionet.net/~razor/razor1911.html",
				Name:       "INQ ad",
				NotWorking: true,
			},
			{
				URL:        "www.ifi.unit.no/razor",
				Name:       "Razor 1911 Founder's website",
				NotWorking: true,
			},
			{
				URL:        "gplus.to/razor1911",
				Name:       "Google+",
				NotWorking: true,
			},
			{
				URL:  "https://wayback.defacto2.net/razor-1911-from-2002-july-1/",
				Name: "Flash site from 2002",
			},
			{
				URL:  "https://web.archive.org/web/19961227152420/http://www.razor1911.com/",
				Name: "Razor 1911 in 1995",
			},
			{
				URL:  "https://razor-1911.tumblr.com/",
				Name: "Tumblr",
			},
			{
				URL:  "https://vimeo.com/groups/razor1911",
				Name: "Vimeo",
			},
			{
				URL:  "http://en.wikipedia.org/wiki/Razor_1911",
				Name: "Wikipedia",
			},
			{
				URL:  "https://twitter.com/razor",
				Name: "Twitter",
			},
			{
				URL:  "http://www.textfiles.com/piracy/RAZOR/",
				Name: "textfiles.com",
			},
		},
		"razor-1911-demo": []Website{
			{
				URL:        "www.razor1911.co.uk",
				Name:       "Razor 1911 Demo Division",
				NotWorking: true,
			},
			{
				URL:        "www.razor1911.com/demo",
				Name:       "Razor 1911",
				NotWorking: true,
			},
		},
		"reality-check-network": []Website{
			{
				URL:        "https://web.archive.org/web/19961223125210/http://rcn.org/",
				Name:       "December 1996",
				NotWorking: false,
			},
			{
				URL:        "https://web.archive.org/web/19970219163852/http://rcn.org/",
				Name:       "1997 redesign",
				NotWorking: false,
			},
			{
				URL:        "www.rcn.org",
				Name:       "Reality Check Network",
				NotWorking: true,
			},
			{
				URL:        "www.shu.edu/~importmi/rcn",
				Name:       "RATM",
				NotWorking: true,
			},
			{
				URL:        "www.halcyon.com/sbecker",
				Name:       "File mirror #1",
				NotWorking: true,
			},
		},
		"rebels": []Website{
			{
				URL:        "www.rebels.dk",
				Name:       "Rebels Island",
				NotWorking: true,
			},
			{
				URL:        "www.rebels.org",
				Name:       "Rebels",
				NotWorking: true,
			},
			{
				URL:        "www.geocities.com/SunsetStrip/3491",
				Name:       "Geocities",
				NotWorking: true,
			},
		},
		"relativity": []Website{
			{
				URL:        "revp.home.ml.org",
				Name:       "Relativity",
				NotWorking: true,
			},
			{
				URL:        "www.cyberbeach.net/~jester/relativity",
				Name:       "Relativity",
				NotWorking: true,
			},
		},
		"risciso": []Website{
			{URL: "www.risc98.org", NotWorking: true},
			{URL: "www.risciso.com", NotWorking: true},
		},
		"scoopex": []Website{
			{
				URL:  "http://www.scoopex1988.org",
				Name: "Scoopex",
			},
		},
		"scenelink": []Website{
			{
				URL:  "/wayback/scenelink-from-1998-june-25/index.html",
				Name: "SceneLink mirror",
			},
			{
				URL:        "www.scenelink.org",
				Name:       "SceneLink",
				NotWorking: true,
			},
			{
				URL:        "www.scene-link.org",
				Name:       "SceneLink",
				NotWorking: true,
			},
		},
		"shock": []Website{
			{
				URL:        "www.shocking.net",
				Name:       "Shock",
				NotWorking: true,
			},
			{
				URL:        "www.shock97.com",
				Name:       "Shock",
				NotWorking: true,
			},
			{
				URL:        "www.shock.org",
				Name:       "Shock",
				NotWorking: true,
			},
		},
		"superior-art-creations": []Website{
			{
				URL:  "https://www.roysac.com/sac.html",
				Name: "RoySAC",
			},
			{
				URL:        "www.superiorartcreations.com",
				Name:       "SAC",
				NotWorking: true,
			},
			{
				URL:        "www.sac2000.home.ml.org",
				Name:       "SAC2000",
				NotWorking: true,
			},
			{
				URL:  "https://www.flickr.com/photos/cumbrowski/collections/72157612320706642/",
				Name: "Art releases",
			},
		},
		"titan": []Website{
			{
				URL:  "https://titandemo.org",
				Name: "Titan",
			},
			{
				URL:        "www.titancrew.org",
				Name:       "Titan",
				NotWorking: true,
			},
		},
		"the-council": []Website{
			{
				URL:        "www.the-council.org",
				Name:       "The Council",
				NotWorking: true,
			},
		},
		"the-game-review": []Website{
			{
				URL:        "www.thegamereview.com",
				Name:       "The Game Review #58",
				NotWorking: true,
			},
			{
				URL:        "www.monmouth.com/~jionin",
				Name:       "The Game Review",
				NotWorking: true,
			},
			{
				URL:        "www.aych-dee.com/tgr.html",
				Name:       "INQ ad",
				NotWorking: true,
			},
			{
				URL:        "ns2.clever.net/~ionizer",
				Name:       "RCN 17 ad",
				NotWorking: true,
			},
			{
				URL:        "https://web.archive.org/web/19990302094306/http://www.lookup.com/homepages/65443/tgr.htm",
				Name:       "incomplete 1995 mirror",
				NotWorking: false,
			},
		},
		"the-humble-guys": []Website{
			{
				URL:  "https://fabulousfurlough.blogspot.com",
				Name: "Fabulous Furlough's - My Life Behind The Patch",
			},
			{
				URL:        "www.thg.net",
				Name:       "1997 reunion",
				NotWorking: true,
			},
		},
		"tristar-ampersand-red-sector-inc": []Website{
			{
				URL:        "www.trsi.de",
				Name:       "TRSi",
				NotWorking: true,
			},
			{
				URL:        "www.trsi.org",
				Name:       "TRSI WHQ",
				NotWorking: true,
			},
			{
				URL:  "https://web.archive.org/web/19961227014238/http://www.trsi.de/",
				Name: "1996 mirror",
			},
		},
		"weapon": []Website{
			{
				URL:        "www.wpnworld.com",
				Name:       "Weapon",
				NotWorking: true,
			},
		},
		"united-cracking-force": []Website{
			{
				URL:        "www.ucf2000.com",
				Name:       "United Cracking Force",
				NotWorking: true,
			},
			{
				URL:        "www.ucf97.com",
				Name:       "UCF97",
				NotWorking: true,
			},
			{
				URL:        "ccnux.utm.my/ucf",
				Name:       "1996 #1",
				NotWorking: true,
			},
			{
				URL:        "ccnux.utm.my/edison",
				Name:       "1996 #2",
				NotWorking: true,
			},
			{
				URL:        "w3.darknet.com/~ucf96/ucf96.htm",
				Name:       "1996 #3",
				NotWorking: true,
			},
		},
		"x_force": []Website{
			{
				URL:        "www.xforce.org",
				Name:       "X-Force",
				NotWorking: true,
			},
		},
		"ultra-force": []Website{
			{
				URL:  "https://ultraforce.com/en/demogroup.html",
				Name: "Ultra Force",
			},
		},
		"x_pression-design": []Website{
			{
				URL:        "www.xpression.org",
				Name:       "X-Pression Design",
				NotWorking: true,
			},
		},
		"pirates-cove": []Website{
			{
				URL:  "https://phrack.org",
				Name: "Phrack Magazine",
			},
		},
		"the-naked-truth-magazine": []Website{
			{
				URL:        "ftp.giga.or.at/pub/pcmags/ntm",
				Name:       "FTP",
				NotWorking: true,
			},
		},
		"inquisition": []Website{
			{
				URL:        "www.openix.com/~apd/inq",
				Name:       "Inquisition",
				NotWorking: true,
			},
		},
		"the-week-in-warez": []Website{
			{
				URL:        "www.crl.com/~tails/wiw.html",
				Name:       "WWN",
				NotWorking: true,
			},
			{
				URL:        "www.hooked.net/users/tails/wwn",
				Name:       "WWN",
				NotWorking: true,
			},
		},
		"zillionz": []Website{
			{
				URL:        "www1.minn.net/~zillionz",
				Name:       "INQ ad",
				NotWorking: true,
			},
		},
		"fast-action-trading-elite": []Website{
			{
				URL:        "www.fatenet.net",
				Name:       "RCN 33 ad",
				NotWorking: true,
			},
			{
				URL:        "www.fate.net",
				Name:       "Original domain",
				NotWorking: true,
			},
			{
				URL:        "www.ceic.com/fate",
				Name:       "INQ ad",
				NotWorking: true,
			},
		},
		"sodom": []Website{
			{
				URL:        "https://wayback.defacto2.net/sodom-from-1998-january-5/",
				Name:       "thesodom mirrored",
				NotWorking: false,
			},
			{
				URL:        "www.thesodom.com",
				Name:       "Sodom",
				NotWorking: true,
			},
			{
				URL:        "jicom.jinr.ru/sodom",
				Name:       "1996",
				NotWorking: true,
			},
		},
		"request-to-send": []Website{
			{
				URL:        "www.request2send.com",
				Name:       "RTS",
				NotWorking: true,
			},
		},
		"reflux": []Website{
			{
				URL:        "addiction.altered.com/carbon8/reflux",
				Name:       "Reflux",
				NotWorking: true,
			},
		},
		"legends-never-die": []Website{
			{
				URL:        "awww.legendsneverdie.com",
				Name:       "Legends Never Die",
				NotWorking: true,
			},
		},
		"heritage": []Website{
			{
				URL:        "www.htg.net",
				Name:       "Heritage",
				NotWorking: true,
			},
		},
		"syndicate": []Website{
			{
				URL:        "www.syn.org",
				Name:       "Syndicate",
				NotWorking: true,
			},
		},
		"old-warez-inc": []Website{
			{
				URL:        "oldwarezinc.home.ml.org",
				Name:       "Old Warez Inc.",
				NotWorking: true,
			},
			{
				URL:        "emulation-world.ml.org/owi",
				Name:       "1998 site",
				NotWorking: true,
			},
		},
		"karma": []Website{
			{
				URL:        "www.karmanet.net",
				Name:       "Karma",
				NotWorking: true,
			},
		},
		"the-reviewers-guild": []Website{
			{
				URL:        "www.trguild.com",
				Name:       "TRGuild",
				NotWorking: true,
			},
			{
				URL:        "trguild.ml.org",
				Name:       "ml.org",
				NotWorking: true,
			},
			{
				URL:        "www.gil.net/~bleys/trg.html",
				Name:       "1997",
				NotWorking: true,
			},
		},
		"affinity": []Website{
			{
				URL:        "www.scenelink.org/aft",
				Name:       "Scenelink",
				NotWorking: true,
			},
			{
				URL:        "affinity.cns.net",
				Name:       "Mr. Mister",
				NotWorking: true,
			},
			{
				URL:        "futureone.com/~damftp/AFT",
				Name:       "DC hosted",
				NotWorking: true,
			},
			{
				URL:        "pages.ripco.com:8080/~devoid",
				Name:       "Devoid",
				NotWorking: true,
			},
			{
				URL:        "206.245.196.81/affinity/affinity.htm",
				Name:       "Issue 5",
				NotWorking: true,
			},
			{
				URL:        "www.iceonline.com/home/rahimk/aft.htm",
				Name:       "Rahimk",
				NotWorking: true,
			},
		},
		"cybermail": []Website{
			{
				URL:        "cybermail.home.ml.org",
				Name:       "ml.org",
				NotWorking: true,
			},
			{
				URL:        "www.dmn.com.au/~warez",
				Name:       "Cybermail",
				NotWorking: true,
			},
		},
		"email-compilation": []Website{
			{
				URL:        "www.xs4all.nl/~blahh",
				Name:       "Source Compilation",
				NotWorking: true,
			},
		},
		"scooby-snack-magazine": []Website{
			{
				URL:        "www.cris.com/~shadowkn",
				Name:       "Scooby Snack Magazine",
				NotWorking: true,
			},
		},
		"anemia": []Website{
			{
				URL:        "www.anemia.org",
				Name:       "Anemia",
				NotWorking: true,
			},
			{
				URL:        "anemia.base.org",
				Name:       "Anemia 97",
				NotWorking: true,
			},
			{
				URL:        "www.geocities.com/SunsetStrip/Towers/5435",
				Name:       "Anemia 96",
				NotWorking: true,
			},
		},
	}
}

// Find returns the website for the given uri.
// It returns an empty string if the uri is not known.
func Find(uri string) []Website {
	sites, groupExists := Websites()[URI(uri)]
	if !groupExists {
		return []Website{}
	}
	// sort using notworking listing as last
	sort.Slice(sites, func(i, j int) bool {
		return !sites[i].NotWorking && sites[j].NotWorking
	})
	return sites
}

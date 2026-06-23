// Package site proves links and titles for recommended websites.
package site

import "sort"

const (
	inqAD     = "INQ ad"
	razor1911 = "Razor 1911"
	shock     = "Shock"
	wikipedia = "Wikipedia"
	web98     = "1998 website"
	web99     = "1999 website"
	web2k     = "2000 website"
	web2006   = "2006 website"
)

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
//
//nolint:gochecknoglobals
var websites = Groups{
	"reverse-engineers-dream": []Website{
		{URL: "redcrew.astalavista.ms", NotWorking: true, Name: "RED 1"},
		{URL: "www.redcrew.net", NotWorking: true, Name: "RED 2"},
		{URL: "www.redteam.net.tc", NotWorking: true, Name: "RED 3"},
		{URL: "distro.webscene.ir/RED", NotWorking: true, Name: "RED 2018"},
	},
	"rise": []Website{
		{URL: "www.risen.org", NotWorking: true, Name: "RiSE"},
	},
	"2000ad": []Website{
		{URL: "www.loke.org/2000AD", NotWorking: true, Name: "2000AD"},
	},
	"prozacs-personal-gaming-report": []Website{
		{URL: "ppgr.webjump.com", NotWorking: true, Name: "PPGR"},
		{URL: "oldschool.io.com/ppgr", NotWorking: true, Name: web99},
		{URL: "ppgr.webjump.com/n64/n64-playable.htm", NotWorking: true, Name: "N64 and emulation Scene news"},
	},
	"crack-in-morocco": []Website{
		{URL: "www.cim.ht.st", NotWorking: true, Name: "website 1"},
		{URL: "cim.boxnet.net", NotWorking: true, Name: "website 2"},
		{URL: "www.cim-team.org", NotWorking: true, Name: "website 3"},
		{URL: "www.cim.astalavista.ms", NotWorking: true, Name: "website 4"},
		{URL: "cimteam.org", NotWorking: true, Name: "website 5"},
	},
	"arab-team-4-reverse-engineering": []Website{
		{URL: "www.at4re.com", NotWorking: true, Name: "at4re"},
	},
	"orion": []Website{
		{URL: "orion2000.dyndns.org", NotWorking: true, Name: web99},
		{URL: "www.oriongods.com", NotWorking: true, Name: web2k},
	},
	"lightforce": []Website{
		{URL: "www.thelightforce.home.ml.org", NotWorking: true, Name: web98},
		{URL: "www.lightforce.cjb.net", NotWorking: true, Name: web2k},
		{URL: "www.thelightforce.com", NotWorking: false, Name: "The LightForce"},
	},
	"linezer0": []Website{
		{URL: "www.coderz.net/linezer0", NotWorking: true, Name: "LineZer0"},
	},
	"orgasming-gaming-magazine": []Website{
		{URL: "www.orgasming.stc.cx", NotWorking: true, Name: "site #1"},
		{URL: "www.slushbucket.com/orgasming.html", NotWorking: true, Name: "site #2"},
		{URL: "www.orgasming.net", NotWorking: true, Name: "temporary domain"},
		{URL: "www.ogm.f2s.com", NotWorking: true, Name: "site #3"},
		{URL: "www.gamershell.com", NotWorking: true, Name: "often advertised"},
	},
	"the-net-monkey-weekly-report": []Website{
		{URL: "archive.pheared.com", NotWorking: true, Name: "NetMonkey"},
		{URL: "www.netmonkey.org", NotWorking: true, Name: "later, official website"},
	},
	"cybercrime-international-network": []Website{
		{URL: "www.meltdown.nu/cci ", NotWorking: true, Name: "CyberCrime"},
	},
	"insanity-couriers": []Website{
		{URL: "insanity.hax0r.org", Name: "Insanity", NotWorking: true},
	},
	"real-time-pirates": []Website{
		{URL: "rtp.home.ml.org", Name: "RTP", NotWorking: true},
	},
	"mortality": []Website{
		{URL: "www.mortality.com", Name: "Mortality", NotWorking: true},
	},
	"masons-ware-report": []Website{
		{URL: "www.cracking.net/mason", NotWorking: true, Name: "website"},
	},
	"the-crazed-asylum": []Website{
		{URL: "blackacid.pheared.com", NotWorking: true, Name: "TCA"},
		{URL: "cpu1058.adsl.bellglobal.com/tca", NotWorking: true, Name: "temp site in dec 1998"},
		{URL: "tca.ramwar.com", NotWorking: true, Name: web99},
		{URL: "tca.phatchicks.com", NotWorking: true, Name: web99},
	},
	"courier-weektop-scorecard": []Website{
		{URL: "www.scenelink.org/relativity/cws", NotWorking: true, Name: "magazine archive"},
		{URL: "www.couriers.org/cws", NotWorking: true, Name: "june " + web98},
		{URL: "rah.simplenet.com/cws", NotWorking: true, Name: "jan " + web99},
		{URL: "cws.couriers.org", NotWorking: true, Name: "dec " + web99},
	},
	"the-sabotage-rebellion-hackers": []Website{
		{URL: "zor.org/tsrh", NotWorking: true, Name: "2002 website"},
		{URL: "tsrh.be", NotWorking: true, Name: web2006},
	},
	"tport": {
		{URL: "www.tport.tk", NotWorking: true, Name: "2004 website"},
		{URL: "www.tport.antishate.net", NotWorking: true, Name: "2004 website"},
		{URL: "www.tport.com.ru", NotWorking: true, Name: web2006},
		{URL: "tport.be", NotWorking: true, Name: web2006},
		{URL: "tport.org", NotWorking: true, Name: "2008 website"},
		{URL: "tport.astalavista.ms", NotWorking: true, Name: "2008 website"},
	},
	"seek-n-destroy": {
		{URL: "zor.org/seekndestroy", NotWorking: true, Name: "first website"},
		{URL: "seekndestroy.host.sk", NotWorking: true, Name: "former website"},
		{URL: "www.seekndestroy.org", NotWorking: true, Name: "aspirational domain"},
	},
	"fighting-for-fun": {
		{
			URL:        "www.fighting-for-fun.fr.st",
			NotWorking: true, Name: "former website domain",
		},
		{
			URL:        "https://wayback.defacto2.net/fighting-for-fun_2002-april/",
			NotWorking: false, Name: "First website mirror",
		},
		{
			URL:        "https://wayback.defacto2.net/fighting-for-fun_circa-2005/",
			NotWorking: false, Name: "October 2002 refresh mirror",
		},
		{
			URL:        "https://wayback.defacto2.net/fighting-for-fun_circa-2003/",
			NotWorking: false, Name: "2003 website mirror",
		},
		{
			URL:        "https://wayback.defacto2.net/fighting-for-fun_circa-2002/",
			NotWorking: false, Name: "New Year's 2005 mirror",
		},
	},
	"class": {
		{URL: "class101.com", NotWorking: true, Name: "former domain"},
		{URL: "www.multimania.com/atm9x", NotWorking: true, Name: "coder ATM/Class"},
		{URL: "https://en.wikipedia.org/wiki/Class_(pirating_group)", Name: "Class (pirating group)", NotWorking: false},
	},
	"shade": {
		{URL: "www.suburbia.net/~shade", NotWorking: true, Name: ""},
	},
	"infinite-darkness-bbs": {
		{URL: "infidark.nws.net", NotWorking: true, Name: "former Telnet"},
	},
	"entropy-bbs": {
		{URL: "entropybbs.net", NotWorking: true, Name: "Telnet board"},
	},
	"sanctuary-bbs": {
		{
			URL:        "https://www.brysk.se/sanctuary/index.htm",
			Name:       "Connect to Sanctuary, former Fairlight HQ",
			NotWorking: false,
		},
	},
	"myth": []Website{
		{URL: "www.myth.org", NotWorking: true, Name: "Mentioned in 2000"},
	},
	"delusions-of-grandeur": []Website{
		{URL: "delusions.base.org", NotWorking: true, Name: ""},
	},
	"celebre": []Website{
		{URL: "www.celebre.net", NotWorking: true, Name: ""},
	},
	"dextrose": []Website{
		{URL: "www.dextrose.com", NotWorking: true, Name: ""},
		{URL: "https://web.archive.org/web/19980131194050/http://www.dextrose.com/", Name: "1998 mirror", NotWorking: false},
	},
	"crc": []Website{{URL: "www.bgnett.no/~xbone", NotWorking: true, Name: ""}},
	"digital-corruption": []Website{
		{URL: "dc.denet.co.jp", NotWorking: true, Name: ""},
		{URL: "https://web.archive.org/web/19971224223854/http://dc.denet.co.jp/", Name: "1997 mirror", NotWorking: false},
	},
	"fantastic-4-cracking-group": []Website{
		{URL: "www.f4cg.com", Name: "advertised in 1998", NotWorking: true},
		{
			URL:        "https://web.archive.org/web/20001204211300/http://www.f4cg.com/",
			Name:       "f4cg.com mirror in 2000, used by the Hitmen",
			NotWorking: false,
		},
	},
	"trc-ware-report": []Website{
		{URL: "http://falcon.laker.net/chemist", Name: "1997 ad in CLASS.NFO", NotWorking: true},
	},
	"the-flame-arrows": []Website{
		{URL: "www.tfa.org", Name: "TFA", NotWorking: true},
		{URL: "www.euronet.nl/users/jdm/documents/members.html", Name: "The Flame Arrows", NotWorking: true},
		{URL: "https://web.archive.org/web/19990117024946/http://www.tfa.org/", Name: "TFA 1999 mirror", NotWorking: false},
		{
			URL:        "https://web.archive.org/web/20000829080106/http://www.euronet.nl/users/jdm/documents/members.html",
			Name:       "The Flame Arrows mirror",
			NotWorking: false,
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
			URL:        "https://en.wikipedia.org/wiki/Paradox_%28warez%29",
			Name:       wikipedia + " - Paradox",
			NotWorking: false,
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
			URL:        "https://www.high-society.at",
			Name:       "High Society",
			NotWorking: false,
		},
	},
	"future-crew": []Website{
		{
			URL:        "https://en.wikipedia.org/wiki/Future_Crew",
			Name:       wikipedia + " - Future Crew",
			NotWorking: false,
		},
		{
			URL:        "www.futurecrew.com",
			Name:       "Future Crew",
			NotWorking: true,
		},
	},
	"eagle-soft-incorporated": []Website{
		{
			URL:        "https://csdb.dk/group/?id=696",
			Name:       "Eagle Soft Incorporated on CSDb",
			NotWorking: false,
		},
	},
	"myth-inc": []Website{
		{
			URL:        "https://demozoo.org/bbs/12549",
			Name:       "Myth Inc BBS on Demozoo",
			NotWorking: false,
		},
	},
	"legion-of-doom": []Website{
		{
			URL:        "https://en.wikipedia.org/wiki/Legion_of_Doom_(hacker_group)",
			Name:       wikipedia + " - Legion of Doom (hacker group)",
			NotWorking: false,
		},
		{
			URL:        "http://textfiles.com/magazines/LOD/",
			Name:       "The Legion of Doom/Hackers Technical Journal",
			NotWorking: false,
		},
	},
	"the-acquisition": []Website{
		{
			URL:        "http://artscene.textfiles.com/acid/ARTPACKS/",
			Name:       "ACiD Art Packs",
			NotWorking: false,
		},
	},
	"acid-productions": []Website{
		{
			URL:        "http://artscene.textfiles.com/acid/",
			Name:       "The ACiD Collection",
			NotWorking: false,
		},
		{
			URL:        "https://www.acid.org",
			Name:       "1996 ACiD webpage",
			NotWorking: false,
		},
		{
			URL:        "http://www.cyberspace.com/~aciddraw",
			Name:       "Original webpage",
			NotWorking: true,
		},
		{
			URL:        "https://en.wikipedia.org/wiki/ACiD_Productions",
			Name:       wikipedia,
			NotWorking: false,
		},
		{
			URL:        "https://www.youtube.com/watch?v=oQrBbm5ZMlo",
			Name:       "BBS The Documentary: Episode 5: Artscene",
			NotWorking: false,
		},
		{
			URL:        "https://archive.org/details/bbs-20020727-radman",
			Name:       "Interview: RaD Man/ACiD",
			NotWorking: false,
		},
		{
			URL:        "https://archive.org/details/20040308-bbs-tracer",
			Name:       "Interview: Tracer/ACiD",
			NotWorking: false,
		},
		{
			URL:        "https://archive.org/details/bbs-20030520-jed",
			Name:       "Interview: JED/ACiD",
			NotWorking: false,
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
			URL:        "https://defacto2.net",
			Name:       "Defacto2",
			NotWorking: false,
		},
		{
			URL:        "https://wayback.defacto2.net/defacto2-from-2000-july-11/",
			Name:       "from July 2000",
			NotWorking: false,
		},
		{
			URL:        "https://wayback.defacto2.net/defacto2-from-1999-september-26/",
			Name:       "from September 1999",
			NotWorking: false,
		},
		{
			URL:        "https://wayback.defacto2.net/defacto2-from-1998-september-8/",
			Name:       "from September 1998",
			NotWorking: false,
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
			URL:        "https://deviance.untergrund.net",
			Name:       "Deviance Demo Division",
			NotWorking: false,
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
			Name:       inqAD,
			NotWorking: true,
		},
	},
	"empress": []Website{
		{
			URL:        "https://www.reddit.com/r/HobbyDrama/comments/rowk83/digital_piracy_the_rise_of_empress_how_one_woman/",
			Name:       "The rise of EMPRESS",
			NotWorking: false,
		},
		{
			URL:        "https://www.wired.com/story/empress-drm-cracking-denuvo-video-game-piracy/",
			Name:       "WIRED interview",
			NotWorking: false,
		},
		{
			URL:        "www.reddit.com/r/EmpressEvolution",
			Name:       "EmpressEvolution",
			NotWorking: true,
		},
	},
	"fairlight": []Website{
		{
			URL:        "https://www.fairlight.to",
			Name:       "Fairlight Commodore 64",
			NotWorking: false,
		},
		{
			URL:        "https://www.fairlight.fi",
			Name:       "Fairlight Finland",
			NotWorking: false,
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
			URL:        "https://web.archive.org/web/19981201194626/http://www.fairlight.org/",
			Name:       "1997 mirror",
			NotWorking: false,
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
			URL:        "https://web.archive.org/web/20010223130305/http://www.multimania.com/jtf98/index.html",
			Name:       "JTF mirror",
			NotWorking: false,
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
			URL:        "https://www.ice.org",
			Name:       "iCE Advertisements",
			NotWorking: false,
		}, {
			URL:        "http://artscene.textfiles.com/ice",
			Name:       "The iCE Collection",
			NotWorking: false,
		}, {
			URL:        "https://en.wikipedia.org/wiki/ICE_Advertisements",
			Name:       wikipedia,
			NotWorking: false,
		}, {
			URL:        "https://www.youtube.com/watch?v=oQrBbm5ZMlo",
			Name:       "BBS The Documentary: Episode 5: Artscene",
			NotWorking: false,
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
		{URL: "www.pdm97.com", Name: "PDM 97", NotWorking: true},
		{URL: "www.paradigm.org", Name: "Paradigm", NotWorking: true},
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
			URL:        "https://www.quartex.org",
			Name:       "Quartex",
			NotWorking: false,
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
			Name:       razor1911,
			NotWorking: false,
		},
		{
			URL:        "www.razor-1911.com",
			Name:       razor1911,
			NotWorking: true,
		},
		{
			URL:        "www.laric.com/razor",
			Name:       "Razor 1911 Europe",
			NotWorking: true,
		},
		{
			URL:        "ionet.net/~razor/razor1911.html",
			Name:       inqAD,
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
			URL:        "https://wayback.defacto2.net/razor-1911-from-2002-july-1/",
			Name:       "Flash site from 2002",
			NotWorking: false,
		},
		{
			URL:        "https://web.archive.org/web/19961227152420/http://www.razor1911.com/",
			Name:       "Razor 1911 in 1995",
			NotWorking: false,
		},
		{
			URL:        "https://razor-1911.tumblr.com/",
			Name:       "Tumblr",
			NotWorking: false,
		},
		{
			URL:        "https://vimeo.com/groups/razor1911",
			Name:       "Vimeo",
			NotWorking: false,
		},
		{
			URL:        "http://en.wikipedia.org/wiki/Razor_1911",
			Name:       wikipedia,
			NotWorking: false,
		},
		{
			URL:        "https://twitter.com/razor",
			Name:       "Twitter",
			NotWorking: false,
		},
		{
			URL:        "http://www.textfiles.com/piracy/RAZOR/",
			Name:       "textfiles.com",
			NotWorking: false,
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
			Name:       razor1911,
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
		{URL: "www.scenelink.org/relativity", Name: web98, NotWorking: true},
	},
	"risciso": []Website{
		{URL: "www.risc98.org", Name: "RISC 98", NotWorking: true},
		{URL: "www.risciso.com", Name: "RISC ISO", NotWorking: true},
	},
	"scoopex": []Website{
		{
			URL:        "http://www.scoopex1988.org",
			Name:       "Scoopex",
			NotWorking: false,
		},
	},
	"scenelink": []Website{
		{
			URL:        "/wayback/scenelink-from-1998-june-25/index.html",
			Name:       "SceneLink mirror",
			NotWorking: false,
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
			Name:       shock,
			NotWorking: true,
		},
		{
			URL:        "www.shock97.com",
			Name:       shock,
			NotWorking: true,
		},
		{
			URL:        "www.shock.org",
			Name:       shock,
			NotWorking: true,
		},
	},
	"superior-art-creations": []Website{
		{
			URL:        "https://www.roysac.com/sac.html",
			Name:       "RoySAC",
			NotWorking: false,
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
			URL:        "https://www.flickr.com/photos/cumbrowski/collections/72157612320706642/",
			Name:       "Art releases",
			NotWorking: false,
		},
	},
	"titan": []Website{
		{
			URL:        "https://titandemo.org",
			Name:       "Titan",
			NotWorking: false,
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
			Name:       inqAD,
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
			URL:        "https://fabulousfurlough.blogspot.com",
			Name:       "Fabulous Furlough's - My Life Behind The Patch",
			NotWorking: false,
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
			URL:        "https://web.archive.org/web/19961227014238/http://www.trsi.de/",
			Name:       "1996 mirror",
			NotWorking: false,
		},
	},
	"weapon": []Website{
		{
			URL:        "www.wpnworld.com",
			Name:       "Weapon",
			NotWorking: true,
		},
		{URL: "www.weapon98.home.ml.org", NotWorking: true, Name: web98},
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
		{
			URL:        "www.geocities.com/SoHo/2680/cracking.html",
			Name:       "Geocities SOHO",
			NotWorking: true,
		},
		{
			URL: "www.geocities.com/Paris/9475/frame.htm", Name: "Geocities Paris", NotWorking: true,
		},
		{
			URL: "members.tripod.com/~ucf96", Name: "Tripod", NotWorking: true,
		},
		{
			URL: "www-pp.hogia.net/gabbah/trainers/", Name: "trainers", NotWorking: true,
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
			URL:        "https://ultraforce.com/en/demogroup.html",
			Name:       "Ultra Force",
			NotWorking: false,
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
			URL:        "https://phrack.org",
			Name:       "Phrack Magazine",
			NotWorking: false,
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
			Name:       inqAD,
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
			Name:       inqAD,
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
	"pirates-with-attitudes": []Website{
		{
			URL:        "https://en.wikipedia.org/wiki/Pirates_with_Attitudes",
			Name:       wikipedia,
			NotWorking: false,
		},
		{
			URL:        "https://files.mpoli.fi/software/DOS/BBS/",
			Name:       "Metropoli BBS including PPEs by PWA",
			NotWorking: false,
		},
		{
			URL:        "archives.thebbs.org/ra117a.htm",
			Name:       "153 PCBoard PPE's by PWA",
			NotWorking: true,
		},
	},
	"united-group-international": []Website{
		{
			URL:        "https://web.archive.org/web/20050228211151/https://xakep.ru/magazine/xa/009/026/1.asp",
			Name:       "Interview with the founder",
			NotWorking: false,
		},
	},
}

// Find returns the website for the given uri.
// It returns an empty string if the uri is not known.
func Find(uri string) []Website {
	sites, groupExists := websites[URI(uri)]
	if !groupExists {
		return []Website{}
	}
	// sort using notworking listing as last
	sort.Slice(sites, func(i, j int) bool {
		return !sites[i].NotWorking && sites[j].NotWorking
	})
	return sites
}

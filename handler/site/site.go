// Package site proves links and titles for recommended websites.
//
//nolint:gochecknoinits
package site

import (
	"slices"
)

const (
	bbsdoc    = "BBS Documentary downloads"
	dztag     = "BBS tagged on Demozoo"
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
	// Error will not apply a HTML hyperlink element to the URL.
	Error bool
}

// Groups is a map of releasers URIs mapped to their websites.
type Groups map[URI][]Website

// Websites returns a map of releasers URIs mapped to their websites.
//
//   - URL should be the absolute URL of the website including the required protocol.
//     But non-working sites should exclude the protocol, e.g. www.example.com.
//   - Name should be the name of the website.
//   - Error will not apply a hyperlink to the URL.
//
//nolint:gochecknoglobals
var websites = Groups{
	//			{URL: "", Error: false, Name: ""},
	//			{URL: "", Error: true, Name: ""},
	//
	"forum-3-bbs": []Website{
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/FORUM/", Error: false, Name: bbsdoc},
	},
	"acs-bbs": []Website{
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/ACS/", Error: false, Name: bbsdoc},
	},
	"pha-elite-bbs": []Website{
		{URL: "https://demozoo.org/groups/159728/", Error: false, Name: "PHA"},
	},
	"telegard-bbs": []Website{
		{URL: "https://www.telegard.net/", Error: false, Name: "Telegard History"},
		{URL: "https://en.wikipedia.org/wiki/Telegard", Error: false, Name: "Wikipedia"},
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/TELEGARD/", Error: false, Name: bbsdoc},
		{URL: "https://bbsfiles.org/browse/bbs-software/telegard", Error: false, Name: "BBS Files"},
		{URL: "https://joeldare.com/a-tour-of-the-telegard-bbs-system", Error: false, Name: "A Tour of the Telegard BBS System"},
	},
	"extreme-bbs": []Website{
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/EXTREME", Error: false, Name: bbsdoc},
	},
	"aftershock-bbs": []Website{
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/AFTERSHOCK/", Error: false, Name: bbsdoc},
	},
	"colossus-bbs": []Website{
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/COLOSSUS/", Error: false, Name: bbsdoc},
	},
	"dominous-productions": []Website{
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/DOMINION/", Error: false, Name: bbsdoc},
	},
	"evolution-software-bbs": []Website{
		{URL: "https://demozoo.org/productions/367052/", Error: false, Name: "TRSi Art/Coding"},
		{URL: "https://demozoo.org/productions/147959/", Error: false, Name: "Fairlight BBS Development"},
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/EVOLUTION/", Error: false, Name: bbsdoc},
	},
	"otera-bbs": []Website{
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/OTERA/", Error: false, Name: bbsdoc},
	},
	"liquid-bbs": []Website{
		{URL: "https://demozoo.org/groups/159800/", Error: false, Name: "CyberWorks"},
		{URL: "https://demozoo.org/bbs/tagged/liquid/", Error: false, Name: "Liquid " + dztag},
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/LIQUIDBBS/", Error: false, Name: bbsdoc},
	},
	"orion-software-bbs": []Website{
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/ORION/", Error: false, Name: bbsdoc},
	},
	"nochange-bbs": []Website{
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/NOCHANGE/", Error: false, Name: bbsdoc},
		{URL: "https://archive.org/details/nochange-bbs-7x/", Error: false, Name: "NOCHANGE v7 (1987)"},
	},
	"vision_x-bbs": []Website{
		{URL: "https://demozoo.org/bbs/tagged/vision-x/", Error: false, Name: "Vision-X " + dztag},
		{URL: "https://demozoo.org/bbs/tagged/vision2/", Error: false, Name: "Vision2 " + dztag},
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/VISION/", Error: false, Name: bbsdoc + " Vision"},
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/VISIONX/", Error: false, Name: bbsdoc + " VisionX"},
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/VISION2/", Error: false, Name: bbsdoc + " Vision2"},
	},
	"lush-software-designs-bbs": []Website{
		{URL: "https://demozoo.org/bbs/tagged/lsd-bbs/", Error: false, Name: "LSD " + dztag},
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/LSDBBS/", Error: false, Name: bbsdoc + " LSDBBS"},
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/LSD/", Error: false, Name: bbsdoc + " LSD"},
	},
	"emulex-bbs": []Website{
		{URL: "https://demozoo.org/bbs/tagged/emulex-bbs/", Error: false, Name: "Emulex " + dztag},
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/EMULEX2/", Error: false, Name: bbsdoc + " Emulex2"},
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/EMULEX/", Error: false, Name: bbsdoc + " Emulex"},
	},
	"monarch-bbs": []Website{
		{URL: "https://demozoo.org/bbs/tagged/monarch-bbs/", Error: false, Name: "Monarch " + dztag},
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/MONARCH2/", Error: false, Name: bbsdoc + " Monarch2"},
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/MONARCH/", Error: false, Name: bbsdoc + " Monarch"},
	},
	"oblivion-2-bbs": []Website{
		{URL: "https://demozoo.org/bbs/tagged/oblivion2", Error: false, Name: "Oblivion2 " + dztag},
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/OBLIVION2/", Error: false, Name: bbsdoc},
		{URL: "https://web.archive.org/web/20021016081152/http://members.darktech.org/obv2/", Error: false, Name: "Oblivion/2 website mirror"},
		{URL: "http://www.darktech.org/obv2/", Error: true, Name: "Oblivion/2 1999 website"},
		{URL: "http://oblivionbbs.home.ml.org/", Error: true, Name: "Oblivion/2 1998 website"},
	},
	"wwiv-bbs": []Website{
		{URL: "https://demozoo.org/bbs/tagged/wwiv", Error: false, Name: "WWIV " + dztag},
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/WWIV/", Error: false, Name: bbsdoc},
		{URL: "https://docs.wwivbbs.org/en/latest/misc/history/", Error: false, Name: "History of WWIV"},
		{URL: "https://www.wwivbbs.org/", Error: false, Name: "WWIV BBS Software"},
		{URL: "http://www.textfiles.com/bbs/WWIVNEWS/", Error: false, Name: "WWIV Newsletters (1991-94)"},
	},
	"genesis-pc-bbs": []Website{
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/GENESISPROJECT/", Error: false, Name: bbsdoc + " for Genesis"},
	},
	"forum_pc-bbs": []Website{
		{URL: "https://demozoo.org/bbs/tagged/forum-bbs", Error: false, Name: "Forum " + dztag},
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/FORUM/", Error: false, Name: bbsdoc + " for Forum"},
	},
	"tcs-bbs": []Website{
		{URL: "https://demozoo.org/groups/159601/", Error: false, Name: "TCS Staff on Demozoo"},
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/TCS/", Error: false, Name: bbsdoc},
	},
	"celerity-bbs": []Website{
		{URL: "https://demozoo.org/bbs/tagged/celerity-bbs", Error: false, Name: "Celerity " + dztag},
		{URL: "https://demozoo.org/groups/135333/", Error: false, Name: "CUD on Demozoo"},
		{URL: "http://software.bbsdocumentary.com/IBM/DOS/CELERITY/", Error: false, Name: bbsdoc},
		{URL: "https://www.infania.net/misc1/shareware/Simtel-MSDOS-Oct1995-CD1/DISC1/CELERITY/00_INDEX.HTM", Error: false, Name: "Simtel 1995 CD listing"},
	},
	"shining-8": []Website{
		{URL: "members.magic-web.net/home/Shining8", Error: true, Name: "1998 C-64 site"},
		{URL: "https://www.shining8.de", Error: false, Name: "Shining 8"},
	},
	"nightfall": []Website{
		{URL: "https://www.nightfallcrew.com", Error: false, Name: "Nightfall Blog"},
	},
	"oldskool": []Website{
		{URL: "www.crystalis.com/~oldskool", Error: true, Name: "OldSkool"},
		{URL: "www1.las.es/~calvo", Error: true, Name: "Calvo site"},
		{URL: "ftp://194.179.101.X/pub/console/oldskool", Error: true, Name: "ftp"},
	},
	"damaged-cybernetics": []Website{
		{URL: "www.futureone.com/~damaged", Error: true, Name: "Damaged Cybernetics"},
		{URL: "www.crystalis.com/~yoshi", Error: true, Name: "Yoshi"},
		{URL: "https://lngn.net/archaic-ruins/", Error: false, Name: "Archaic Ruins"},
		{
			URL: "https://patpend.net/articles/ar/damaged.html", Error: false,
			Name: "A record of the disbanding of Damaged Cybernetics",
		},
		{
			URL: "https://ia803104.us.archive.org/23/items/bloodlust_software/articles/" +
				"emulators.cz%20-%20The%20Real%20Story%20behind%20Damaged%20Cybernetics%20Closing....html",
			Error: false, Name: "The Real Truth",
		},
		{
			URL: "https://patpend.net/articles/zd/article2.html", Error: false,
			Name: "Damaged Cybernetics: A Year since its death",
		},
		{
			URL: "https://www.vice.com/en/article/" +
				"the-story-of-nesticle-the-ambitious-emulator-that-redefined-retro-gaming/",
			Error: false, Name: "VICE: The Story of NESticle",
		},
	},
	"skid-row": []Website{
		{URL: "www.skidrow.dk", Error: true, Name: "Skid Row Amiga"},
	},
	"alpha-flight": []Website{
		{URL: "alphaflight1970.de", Error: true, Name: "Alpha Flight 1970"},
	},
	"elitendo": []Website{
		{URL: "http://www.hotline9009.com/elitendo.htm", Error: false, Name: "Elitendo"},
	},
	"anthrox": []Website{
		{URL: "www.anthrox.com", Error: true, Name: "Anthrox"},
	},
	"mystic": []Website{
		{URL: "http://amigascne.org/mystic", Error: false, Name: "Mystic home page"},
	},
	"reverse-engineers-dream": []Website{
		{URL: "redcrew.astalavista.ms", Error: true, Name: "RED 1"},
		{URL: "www.redcrew.net", Error: true, Name: "RED 2"},
		{URL: "www.redteam.net.tc", Error: true, Name: "RED 3"},
		{URL: "distro.webscene.ir/RED", Error: true, Name: "RED 2018"},
	},
	"rise": []Website{
		{URL: "www.risen.org", Error: true, Name: "RiSE"},
	},
	"2000ad": []Website{
		{URL: "www.loke.org/2000AD", Error: true, Name: "2000AD"},
	},
	"prozacs-personal-gaming-report": []Website{
		{URL: "ppgr.webjump.com", Error: true, Name: "PPGR"},
		{URL: "oldschool.io.com/ppgr", Error: true, Name: web99},
		{URL: "ppgr.webjump.com/n64/n64-playable.htm", Error: true, Name: "N64 and emulation Scene news"},
	},
	"crack-in-morocco": []Website{
		{URL: "www.cim.ht.st", Error: true, Name: "website 1"},
		{URL: "cim.boxnet.net", Error: true, Name: "website 2"},
		{URL: "www.cim-team.org", Error: true, Name: "website 3"},
		{URL: "www.cim.astalavista.ms", Error: true, Name: "website 4"},
		{URL: "cimteam.org", Error: true, Name: "website 5"},
	},
	"arab-team-4-reverse-engineering": []Website{
		{URL: "www.at4re.com", Error: true, Name: "at4re"},
	},
	"orion": []Website{
		{URL: "orion2000.dyndns.org", Error: true, Name: web99},
		{URL: "www.oriongods.com", Error: true, Name: web2k},
	},
	"lightforce": []Website{
		{URL: "www.thelightforce.home.ml.org", Error: true, Name: web98},
		{URL: "www.lightforce.cjb.net", Error: true, Name: web2k},
		{URL: "www.thelightforce.com", Error: false, Name: "The LightForce"},
	},
	"linezer0": []Website{
		{URL: "www.coderz.net/linezer0", Error: true, Name: "LineZer0"},
	},
	"orgasming-gaming-magazine": []Website{
		{URL: "www.orgasming.stc.cx", Error: true, Name: "site #1"},
		{URL: "www.slushbucket.com/orgasming.html", Error: true, Name: "site #2"},
		{URL: "www.orgasming.net", Error: true, Name: "temporary domain"},
		{URL: "www.ogm.f2s.com", Error: true, Name: "site #3"},
		{URL: "www.gamershell.com", Error: true, Name: "often advertised"},
	},
	"the-net-monkey-weekly-report": []Website{
		{URL: "archive.pheared.com", Error: true, Name: "NetMonkey"},
		{URL: "www.netmonkey.org", Error: true, Name: "later, official website"},
	},
	"cybercrime-international-network": []Website{
		{URL: "www.meltdown.nu/cci ", Error: true, Name: "CyberCrime"},
	},
	"insanity-couriers": []Website{
		{URL: "insanity.hax0r.org", Name: "Insanity", Error: true},
	},
	"real-time-pirates": []Website{
		{URL: "rtp.home.ml.org", Name: "RTP", Error: true},
	},
	"mortality": []Website{
		{URL: "www.mortality.com", Name: "Mortality", Error: true},
	},
	"masons-ware-report": []Website{
		{URL: "www.cracking.net/mason", Error: true, Name: "website"},
	},
	"the-crazed-asylum": []Website{
		{URL: "blackacid.pheared.com", Error: true, Name: "TCA"},
		{URL: "cpu1058.adsl.bellglobal.com/tca", Error: true, Name: "temp site in dec 1998"},
		{URL: "tca.ramwar.com", Error: true, Name: web99},
		{URL: "tca.phatchicks.com", Error: true, Name: web99},
		{URL: "thecrazed.com", Error: true, Name: "unused domain"},
	},
	"courier-weektop-scorecard": []Website{
		{URL: "www.scenelink.org/relativity/cws", Error: true, Name: "magazine archive"},
		{URL: "www.couriers.org/cws", Error: true, Name: "june " + web98},
		{URL: "rah.simplenet.com/cws", Error: true, Name: "jan " + web99},
		{URL: "cws.couriers.org", Error: true, Name: "dec " + web99},
	},
	"the-sabotage-rebellion-hackers": []Website{
		{URL: "zor.org/tsrh", Error: true, Name: "2002 website"},
		{URL: "tsrh.be", Error: true, Name: web2006},
	},
	"tport": {
		{URL: "www.tport.tk", Error: true, Name: "2004 website"},
		{URL: "www.tport.antishate.net", Error: true, Name: "2004 website"},
		{URL: "www.tport.com.ru", Error: true, Name: web2006},
		{URL: "tport.be", Error: true, Name: web2006},
		{URL: "tport.org", Error: true, Name: "2008 website"},
		{URL: "tport.astalavista.ms", Error: true, Name: "2008 website"},
	},
	"seek-n-destroy": {
		{URL: "zor.org/seekndestroy", Error: true, Name: "first website"},
		{URL: "seekndestroy.host.sk", Error: true, Name: "former website"},
		{URL: "www.seekndestroy.org", Error: true, Name: "aspirational domain"},
	},
	"fighting-for-fun": {
		{
			URL:   "www.fighting-for-fun.fr.st",
			Error: true, Name: "former website domain",
		},
		{
			URL:   "https://wayback.defacto2.net/fighting-for-fun_2002-april/",
			Error: false, Name: "First website mirror",
		},
		{
			URL:   "https://wayback.defacto2.net/fighting-for-fun_circa-2005/",
			Error: false, Name: "October 2002 refresh mirror",
		},
		{
			URL:   "https://wayback.defacto2.net/fighting-for-fun_circa-2003/",
			Error: false, Name: "2003 website mirror",
		},
		{
			URL:   "https://wayback.defacto2.net/fighting-for-fun_circa-2002/",
			Error: false, Name: "New Year's 2005 mirror",
		},
	},
	"class": {
		{URL: "class101.com", Error: true, Name: "former domain"},
		{URL: "www.multimania.com/atm9x", Error: true, Name: "coder ATM/Class"},
		{URL: "https://en.wikipedia.org/wiki/Class_(pirating_group)", Name: "Class (pirating group)", Error: false},
	},
	"shade": {
		{URL: "www.suburbia.net/~shade", Error: true, Name: ""},
	},
	"infinite-darkness-bbs": {
		{URL: "infidark.nws.net", Error: true, Name: "former Telnet"},
	},
	"entropy-bbs": {
		{URL: "entropybbs.net", Error: true, Name: "Telnet board"},
	},
	"sanctuary-bbs": {
		{
			URL:   "https://www.brysk.se/sanctuary/index.htm",
			Name:  "Connect to Sanctuary, former Fairlight HQ",
			Error: false,
		},
	},
	"myth": []Website{
		{URL: "www.myth.org", Error: true, Name: "Mentioned in 2000"},
	},
	"delusions-of-grandeur": []Website{
		{URL: "delusions.base.org", Error: true, Name: ""},
	},
	"celebre": []Website{
		{URL: "www.celebre.net", Error: true, Name: ""},
	},
	"dextrose": []Website{
		{URL: "www.dextrose.com", Error: true, Name: ""},
		{URL: "https://web.archive.org/web/19980131194050/http://www.dextrose.com/", Name: "1998 mirror", Error: false},
	},
	"crc": []Website{{URL: "www.bgnett.no/~xbone", Error: true, Name: ""}},
	"digital-corruption": []Website{
		{URL: "dc.denet.co.jp", Error: true, Name: ""},
		{URL: "https://web.archive.org/web/19971224223854/http://dc.denet.co.jp/", Name: "1997 mirror", Error: false},
	},
	"fantastic-4-cracking-group": []Website{
		{URL: "www.f4cg.com", Name: "advertised in 1998", Error: true},
		{
			URL:   "https://web.archive.org/web/20001204211300/http://www.f4cg.com/",
			Name:  "f4cg.com mirror in 2000, used by the Hitmen",
			Error: false,
		},
	},
	"trc-ware-report": []Website{
		{URL: "http://falcon.laker.net/chemist", Name: "1997 ad in CLASS.NFO", Error: true},
	},
	"the-flame-arrows": []Website{
		{URL: "www.tfa.org", Name: "TFA", Error: true},
		{URL: "www.euronet.nl/users/jdm/documents/members.html", Name: "The Flame Arrows", Error: true},
		{URL: "https://web.archive.org/web/19990117024946/http://www.tfa.org/", Name: "TFA 1999 mirror", Error: false},
		{
			URL:   "https://web.archive.org/web/20000829080106/http://www.euronet.nl/users/jdm/documents/members.html",
			Name:  "The Flame Arrows mirror",
			Error: false,
		},
	},
	"byte-enforcerz": []Website{
		{URL: "www.ussuriisk.ru/~beg/", Name: "BEG", Error: true},
	},
	"bios-systems": []Website{
		{URL: "www.bios-systems.yi.org", Name: "BIOS Systems", Error: true},
	},
	"esprit-couriers": []Website{
		{URL: "www.esprit.org", Name: "Esprit", Error: true},
	},
	"sea-shell-commando": []Website{
		{URL: "www.sedona.net/~weirdo", Name: "SSC application", Error: true},
	},
	"breed": []Website{
		{URL: "opal.utu.fi/~ternur/breed/index.html", Name: "Breed", Error: true},
	},
	"dead-pirates-society": []Website{
		{
			URL:   "www.cinecan.com/dps",
			Name:  "DSP by |Swine|",
			Error: true,
		},
	},
	"paradox": []Website{
		{
			URL:   "www.paradogs.com",
			Name:  "Paradox",
			Error: true,
		},
		{
			URL:   "https://en.wikipedia.org/wiki/Paradox_%28warez%29",
			Name:  wikipedia + " - Paradox",
			Error: false,
		},
	},
	"light-speed-warez": []Website{
		{
			URL:   "www.lsw.org",
			Name:  "Light Speed Warez",
			Error: true,
		},
	},
	"high-society": []Website{
		{
			URL:   "www.high-society.org",
			Name:  "High Society",
			Error: true,
		},
		{
			URL:   "https://www.high-society.at",
			Name:  "High Society",
			Error: false,
		},
	},
	"future-crew": []Website{
		{
			URL:   "https://en.wikipedia.org/wiki/Future_Crew",
			Name:  wikipedia + " - Future Crew",
			Error: false,
		},
		{
			URL:   "www.futurecrew.com",
			Name:  "Future Crew",
			Error: true,
		},
	},
	"eagle-soft-incorporated": []Website{
		{
			URL:   "https://csdb.dk/group/?id=696",
			Name:  "Eagle Soft Incorporated on CSDb",
			Error: false,
		},
	},
	"myth-inc": []Website{
		{
			URL:   "https://demozoo.org/bbs/12549",
			Name:  "Myth Inc BBS on Demozoo",
			Error: false,
		},
	},
	"legion-of-doom": []Website{
		{
			URL:   "https://en.wikipedia.org/wiki/Legion_of_Doom_(hacker_group)",
			Name:  wikipedia + " - Legion of Doom (hacker group)",
			Error: false,
		},
		{
			URL:   "http://textfiles.com/magazines/LOD/",
			Name:  "The Legion of Doom/Hackers Technical Journal",
			Error: false,
		},
	},
	"the-acquisition": []Website{
		{
			URL:   "http://artscene.textfiles.com/acid/ARTPACKS/",
			Name:  "ACiD Art Packs",
			Error: false,
		},
	},
	"acid-productions": []Website{
		{
			URL:   "http://artscene.textfiles.com/acid/",
			Name:  "The ACiD Collection",
			Error: false,
		},
		{
			URL:   "https://www.acid.org",
			Name:  "1996 ACiD webpage",
			Error: false,
		},
		{
			URL:   "http://www.cyberspace.com/~aciddraw",
			Name:  "Original webpage",
			Error: true,
		},
		{
			URL:   "https://en.wikipedia.org/wiki/ACiD_Productions",
			Name:  wikipedia,
			Error: false,
		},
		{
			URL:   "https://www.youtube.com/watch?v=oQrBbm5ZMlo",
			Name:  "BBS The Documentary: Episode 5: Artscene",
			Error: false,
		},
		{
			URL:   "https://archive.org/details/bbs-20020727-radman",
			Name:  "Interview: RaD Man/ACiD",
			Error: false,
		},
		{
			URL:   "https://archive.org/details/20040308-bbs-tracer",
			Name:  "Interview: Tracer/ACiD",
			Error: false,
		},
		{
			URL:   "https://archive.org/details/bbs-20030520-jed",
			Name:  "Interview: JED/ACiD",
			Error: false,
		},
	},
	"assault": []Website{
		{
			URL:   "www.nrg2000.com",
			Name:  "Assault",
			Error: true,
		},
	},
	"chemical-reaction": []Website{
		{
			URL:   "www.creaction.com",
			Name:  "Chemical Reaction",
			Error: true,
		},
	},
	"core": []Website{
		{
			URL:   "coremongos.home.ml.org",
			Name:  "CORE",
			Error: true,
		},
	},
	"defacto2": []Website{
		{
			URL:   "https://defacto2.net",
			Name:  "Defacto2",
			Error: false,
		},
		{
			URL:   "https://wayback.defacto2.net/defacto2-from-2000-july-11/",
			Name:  "from July 2000",
			Error: false,
		},
		{
			URL:   "https://wayback.defacto2.net/defacto2-from-1999-september-26/",
			Name:  "from September 1999",
			Error: false,
		},
		{
			URL:   "https://wayback.defacto2.net/defacto2-from-1998-september-8/",
			Name:  "from September 1998",
			Error: false,
		},
		{
			URL:   "www.defacto2.com",
			Name:  "launch address",
			Error: true,
		},
	},
	"defacto": []Website{
		{
			URL:   "www.jicom.jinr.ru/sodom",
			Name:  "Defacto",
			Error: true,
		},
	},
	"deviance": []Website{
		{
			URL:   "https://deviance.untergrund.net",
			Name:  "Deviance Demo Division",
			Error: false,
		},
	},
	"devotion": []Website{
		{
			URL:   "www.devotion.pp.se",
			Name:  "Devotion",
			Error: true,
		},
		{
			URL:   "www.dataplus.se/devotion",
			Name:  "by Strooper and Spy",
			Error: true,
		},
	},
	"divine": []Website{
		{URL: "dvn.org", Name: "Divine", Error: true},
		{URL: "www.divinegods.com", Name: "Divine Gods", Error: true},
	},
	"drink-or-die": []Website{
		{
			URL:   "www.drinkordie.com",
			Name:  "Drink Or Die",
			Error: true,
		},
		{
			URL:   "spl.co.il/zino",
			Name:  inqAD,
			Error: true,
		},
	},
	"empress": []Website{
		{
			URL:   "https://www.reddit.com/r/HobbyDrama/comments/rowk83/digital_piracy_the_rise_of_empress_how_one_woman/",
			Name:  "The rise of EMPRESS",
			Error: false,
		},
		{
			URL:   "https://www.wired.com/story/empress-drm-cracking-denuvo-video-game-piracy/",
			Name:  "WIRED interview",
			Error: false,
		},
		{
			URL:   "www.reddit.com/r/EmpressEvolution",
			Name:  "EmpressEvolution",
			Error: true,
		},
	},
	"fairlight": []Website{
		{
			URL:   "https://www.fairlight.to",
			Name:  "Fairlight Commodore 64",
			Error: false,
		},
		{
			URL:   "https://www.fairlight.fi",
			Name:  "Fairlight Finland",
			Error: false,
		},
		{
			URL:   "www.fairlight.org",
			Name:  "Fairlight",
			Error: true,
		},
		{
			URL:   "www.ludd.luth.se/~watchman/fairlight",
			Name:  "Fairlight Sweden",
			Error: true,
		},
		{
			URL:   "https://web.archive.org/web/19981201194626/http://www.fairlight.org/",
			Name:  "1997 mirror",
			Error: false,
		},
	},
	"fire-site-ftp": []Website{
		{
			URL:   "firesite.ml.org",
			Name:  "Fire Site FTP",
			Error: true,
		},
	},
	"gorgeous-ladies-of-warez": []Website{
		{
			URL:   "www.glow.org",
			Name:  "Gorgeous Ladies Of Warez",
			Error: true,
		},
	},
	"just-the-facts": []Website{
		{
			URL:   "www.mygale.org/~jtf98",
			Name:  "Just The Facts",
			Error: true,
		},
		{
			URL:   "www.multimania.com/jtf98/",
			Name:  "Just The Facts",
			Error: true,
		},
		{
			URL:   "https://web.archive.org/web/20010223130305/http://www.multimania.com/jtf98/index.html",
			Name:  "JTF mirror",
			Error: false,
		},
	},
	"hybrid": []Website{
		{
			URL:   "www.hybridism.com",
			Name:  "Hybrid 1998",
			Error: true,
		},
		{
			URL:   "www.hybrid.to",
			Name:  "Hybrid",
			Error: true,
		},
		{
			URL:   "www.hybreed.com",
			Name:  "Hybreed",
			Error: true,
		},
		{
			URL:   "www.tripnet.se/~electro/home",
			Name:  "Hoson in INQ",
			Error: true,
		},
	},
	"hybrid-christmas-e_mag": []Website{
		{
			URL:   "www.hybreed.com",
			Name:  "Hybreed",
			Error: true,
		},
	},
	"insane-creators-enterprise": []Website{
		{
			URL:   "https://www.ice.org",
			Name:  "iCE Advertisements",
			Error: false,
		}, {
			URL:   "http://artscene.textfiles.com/ice",
			Name:  "The iCE Collection",
			Error: false,
		}, {
			URL:   "https://en.wikipedia.org/wiki/ICE_Advertisements",
			Name:  wikipedia,
			Error: false,
		}, {
			URL:   "https://www.youtube.com/watch?v=oQrBbm5ZMlo",
			Name:  "BBS The Documentary: Episode 5: Artscene",
			Error: false,
		},
	},
	"level4": []Website{
		{
			URL:   "www.level4.ml.org",
			Name:  "Level 4",
			Error: true,
		},
	},
	"motiv8": []Website{
		{
			URL:   "www.motiv8.org",
			Name:  "Motiv8",
			Error: true,
		},
		{
			URL:   "hipsworld.bridge.net/~tribal",
			Name:  "June 1996",
			Error: true,
		},
	},
	"paradigm": []Website{
		{
			URL:   "www.pdmworld.com",
			Name:  "Paradigm",
			Error: true,
		},
		{URL: "www.pdm97.com", Name: "PDM 97", Error: true},
		{URL: "www.paradigm.org", Name: "Paradigm", Error: true},
		{
			URL:   "www.pdmworld.com/dac",
			Name:  "DAC Paradigm art",
			Error: true,
		},
	},
	"phrozen-crew": []Website{
		{
			URL:   "www.phrozencrew.org",
			Name:  "Phrozen Crew",
			Error: true,
		},
		{URL: "www.false.com/phrozen", Name: "1996 website", Error: true},
	},
	"premiere": []Website{
		{
			URL:   "premiere97.com",
			Name:  "Premiere 97",
			Error: true,
		},
		{
			URL:   "premiere.ttlc.net",
			Name:  "Premiere",
			Error: true,
		},
	},
	"prestige": []Website{
		{
			URL:   "www.laker.net/prestige",
			Name:  "Prestige",
			Error: true,
		},
		{
			URL:   "www.jet.laker.net/prestige",
			Name:  "June 1996",
			Error: true,
		},
	},
	"quartex": []Website{
		{
			URL:   "https://www.quartex.org",
			Name:  "Quartex",
			Error: false,
		},
		{
			URL:   "www.quartex.demon.co.uk",
			Name:  "Quartex",
			Error: true,
		},
	},
	"razor-1911": []Website{
		{
			URL:   "https://www.razor1911.com",
			Name:  razor1911,
			Error: false,
		},
		{
			URL:   "www.razor-1911.com",
			Name:  razor1911,
			Error: true,
		},
		{
			URL:   "www.laric.com/razor",
			Name:  "Razor 1911 Europe",
			Error: true,
		},
		{
			URL:   "ionet.net/~razor/razor1911.html",
			Name:  inqAD,
			Error: true,
		},
		{
			URL:   "www.ifi.unit.no/razor",
			Name:  "Razor 1911 Founder's website",
			Error: true,
		},
		{
			URL:   "gplus.to/razor1911",
			Name:  "Google+",
			Error: true,
		},
		{
			URL:   "https://wayback.defacto2.net/razor-1911-from-2002-july-1/",
			Name:  "Flash site from 2002",
			Error: false,
		},
		{
			URL:   "https://web.archive.org/web/19961227152420/http://www.razor1911.com/",
			Name:  "Razor 1911 in 1995",
			Error: false,
		},
		{
			URL:   "https://razor-1911.tumblr.com/",
			Name:  "Tumblr",
			Error: false,
		},
		{
			URL:   "https://vimeo.com/groups/razor1911",
			Name:  "Vimeo",
			Error: false,
		},
		{
			URL:   "http://en.wikipedia.org/wiki/Razor_1911",
			Name:  wikipedia,
			Error: false,
		},
		{
			URL:   "https://twitter.com/razor",
			Name:  "Twitter",
			Error: false,
		},
		{
			URL:   "http://www.textfiles.com/piracy/RAZOR/",
			Name:  "textfiles.com",
			Error: false,
		},
	},
	"razor-1911-demo": []Website{
		{
			URL:   "www.razor1911.co.uk",
			Name:  "Razor 1911 Demo Division",
			Error: true,
		},
		{
			URL:   "www.razor1911.com/demo",
			Name:  razor1911,
			Error: true,
		},
	},
	"reality-check-network": []Website{
		{
			URL:   "https://web.archive.org/web/19961223125210/http://rcn.org/",
			Name:  "December 1996",
			Error: false,
		},
		{
			URL:   "https://web.archive.org/web/19970219163852/http://rcn.org/",
			Name:  "1997 redesign",
			Error: false,
		},
		{
			URL:   "www.rcn.org",
			Name:  "Reality Check Network",
			Error: true,
		},
		{
			URL:   "www.shu.edu/~importmi/rcn",
			Name:  "RATM",
			Error: true,
		},
		{
			URL:   "www.halcyon.com/sbecker",
			Name:  "File mirror #1",
			Error: true,
		},
	},
	"rebels": []Website{
		{
			URL:   "www.rebels.dk",
			Name:  "Rebels Island",
			Error: true,
		},
		{
			URL:   "www.rebels.org",
			Name:  "Rebels",
			Error: true,
		},
		{
			URL:   "www.geocities.com/SunsetStrip/3491",
			Name:  "Geocities",
			Error: true,
		},
	},
	"relativity": []Website{
		{
			URL:   "revp.home.ml.org",
			Name:  "Relativity",
			Error: true,
		},
		{
			URL:   "www.cyberbeach.net/~jester/relativity",
			Name:  "Relativity",
			Error: true,
		},
		{URL: "www.scenelink.org/relativity", Name: web98, Error: true},
	},
	"risciso": []Website{
		{URL: "www.risc98.org", Name: "RISC 98", Error: true},
		{URL: "www.risciso.com", Name: "RISC ISO", Error: true},
	},
	"scoopex": []Website{
		{
			URL:   "http://www.scoopex1988.org",
			Name:  "Scoopex",
			Error: false,
		},
	},
	"scenelink": []Website{
		{
			URL:   "/wayback/scenelink-from-1998-june-25/index.html",
			Name:  "SceneLink mirror",
			Error: false,
		},
		{
			URL:   "www.scenelink.org",
			Name:  "SceneLink",
			Error: true,
		},
		{
			URL:   "www.scene-link.org",
			Name:  "SceneLink",
			Error: true,
		},
	},
	"shock": []Website{
		{
			URL:   "www.shocking.net",
			Name:  shock,
			Error: true,
		},
		{
			URL:   "www.shock97.com",
			Name:  shock,
			Error: true,
		},
		{
			URL:   "www.shock.org",
			Name:  shock,
			Error: true,
		},
	},
	"superior-art-creations": []Website{
		{
			URL:   "https://www.roysac.com/sac.html",
			Name:  "RoySAC",
			Error: false,
		},
		{
			URL:   "www.superiorartcreations.com",
			Name:  "SAC",
			Error: true,
		},
		{
			URL:   "www.sac2000.home.ml.org",
			Name:  "SAC2000",
			Error: true,
		},
		{
			URL:   "https://www.flickr.com/photos/cumbrowski/collections/72157612320706642/",
			Name:  "Art releases",
			Error: false,
		},
	},
	"titan": []Website{
		{
			URL:   "https://titandemo.org",
			Name:  "Titan",
			Error: false,
		},
		{
			URL:   "www.titancrew.org",
			Name:  "Titan",
			Error: true,
		},
	},
	"the-council": []Website{
		{
			URL:   "www.the-council.org",
			Name:  "The Council",
			Error: true,
		},
	},
	"the-game-review": []Website{
		{
			URL:   "www.thegamereview.com",
			Name:  "The Game Review #58",
			Error: true,
		},
		{
			URL:   "www.monmouth.com/~jionin",
			Name:  "The Game Review",
			Error: true,
		},
		{
			URL:   "www.aych-dee.com/tgr.html",
			Name:  inqAD,
			Error: true,
		},
		{
			URL:   "ns2.clever.net/~ionizer",
			Name:  "RCN 17 ad",
			Error: true,
		},
		{
			URL:   "https://web.archive.org/web/19990302094306/http://www.lookup.com/homepages/65443/tgr.htm",
			Name:  "incomplete 1995 mirror",
			Error: false,
		},
	},
	"the-humble-guys": []Website{
		{
			URL:   "https://fabulousfurlough.blogspot.com",
			Name:  "Fabulous Furlough's - My Life Behind The Patch",
			Error: false,
		},
		{
			URL:   "www.thg.net",
			Name:  "1997 reunion",
			Error: true,
		},
	},
	"tristar-ampersand-red-sector-inc": []Website{
		{
			URL:   "www.trsi.de",
			Name:  "TRSi",
			Error: true,
		},
		{
			URL:   "www.trsi.org",
			Name:  "TRSI WHQ",
			Error: true,
		},
		{
			URL:   "https://web.archive.org/web/19961227014238/http://www.trsi.de/",
			Name:  "1996 mirror",
			Error: false,
		},
	},
	"weapon": []Website{
		{
			URL:   "www.wpnworld.com",
			Name:  "Weapon",
			Error: true,
		},
		{URL: "www.weapon98.home.ml.org", Error: true, Name: web98},
	},
	"united-cracking-force": []Website{
		{
			URL:   "www.ucf2000.com",
			Name:  "United Cracking Force",
			Error: true,
		},
		{
			URL:   "www.ucf97.com",
			Name:  "UCF97",
			Error: true,
		},
		{
			URL:   "ccnux.utm.my/ucf",
			Name:  "1996 #1",
			Error: true,
		},
		{
			URL:   "ccnux.utm.my/edison",
			Name:  "1996 #2",
			Error: true,
		},
		{
			URL:   "w3.darknet.com/~ucf96/ucf96.htm",
			Name:  "1996 #3",
			Error: true,
		},
		{
			URL:   "www.geocities.com/SoHo/2680/cracking.html",
			Name:  "Geocities SOHO",
			Error: true,
		},
		{
			URL: "www.geocities.com/Paris/9475/frame.htm", Name: "Geocities Paris", Error: true,
		},
		{
			URL: "members.tripod.com/~ucf96", Name: "Tripod", Error: true,
		},
		{
			URL: "www-pp.hogia.net/gabbah/trainers/", Name: "trainers", Error: true,
		},
	},
	"x_force": []Website{
		{
			URL:   "www.xforce.org",
			Name:  "X-Force",
			Error: true,
		},
	},
	"ultra-force": []Website{
		{
			URL:   "https://ultraforce.com/en/demogroup.html",
			Name:  "Ultra Force",
			Error: false,
		},
	},
	"x_pression-design": []Website{
		{
			URL:   "www.xpression.org",
			Name:  "X-Pression Design",
			Error: true,
		},
	},
	"pirates-cove": []Website{
		{
			URL:   "https://phrack.org",
			Name:  "Phrack Magazine",
			Error: false,
		},
	},
	"the-naked-truth-magazine": []Website{
		{
			URL:   "ftp.giga.or.at/pub/pcmags/ntm",
			Name:  "FTP",
			Error: true,
		},
	},
	"inquisition": []Website{
		{
			URL:   "www.openix.com/~apd/inq",
			Name:  "Inquisition",
			Error: true,
		},
	},
	"the-week-in-warez": []Website{
		{
			URL:   "www.crl.com/~tails/wiw.html",
			Name:  "WWN",
			Error: true,
		},
		{
			URL:   "www.hooked.net/users/tails/wwn",
			Name:  "WWN",
			Error: true,
		},
	},
	"zillionz": []Website{
		{
			URL:   "www1.minn.net/~zillionz",
			Name:  inqAD,
			Error: true,
		},
	},
	"fast-action-trading-elite": []Website{
		{
			URL:   "www.fatenet.net",
			Name:  "RCN 33 ad",
			Error: true,
		},
		{
			URL:   "www.fate.net",
			Name:  "Original domain",
			Error: true,
		},
		{
			URL:   "www.ceic.com/fate",
			Name:  inqAD,
			Error: true,
		},
	},
	"sodom": []Website{
		{
			URL:   "https://wayback.defacto2.net/sodom-from-1998-january-5/",
			Name:  "thesodom mirrored",
			Error: false,
		},
		{
			URL:   "www.thesodom.com",
			Name:  "Sodom",
			Error: true,
		},
		{
			URL:   "jicom.jinr.ru/sodom",
			Name:  "1996",
			Error: true,
		},
	},
	"request-to-send": []Website{
		{
			URL:   "www.request2send.com",
			Name:  "RTS",
			Error: true,
		},
	},
	"reflux": []Website{
		{
			URL:   "addiction.altered.com/carbon8/reflux",
			Name:  "Reflux",
			Error: true,
		},
	},
	"legends-never-die": []Website{
		{
			URL:   "awww.legendsneverdie.com",
			Name:  "Legends Never Die",
			Error: true,
		},
	},
	"heritage": []Website{
		{
			URL:   "www.htg.net",
			Name:  "Heritage",
			Error: true,
		},
	},
	"syndicate": []Website{
		{
			URL:   "www.syn.org",
			Name:  "Syndicate",
			Error: true,
		},
	},
	"old-warez-inc": []Website{
		{
			URL:   "oldwarezinc.home.ml.org",
			Name:  "Old Warez Inc.",
			Error: true,
		},
		{
			URL:   "emulation-world.ml.org/owi",
			Name:  "1998 site",
			Error: true,
		},
	},
	"karma": []Website{
		{
			URL:   "www.karmanet.net",
			Name:  "Karma",
			Error: true,
		},
	},
	"the-reviewers-guild": []Website{
		{
			URL:   "www.trguild.com",
			Name:  "TRGuild",
			Error: true,
		},
		{
			URL:   "trguild.ml.org",
			Name:  "ml.org",
			Error: true,
		},
		{
			URL:   "www.gil.net/~bleys/trg.html",
			Name:  "1997",
			Error: true,
		},
	},
	"affinity": []Website{
		{
			URL:   "www.scenelink.org/aft",
			Name:  "Scenelink",
			Error: true,
		},
		{
			URL:   "affinity.cns.net",
			Name:  "Mr. Mister",
			Error: true,
		},
		{
			URL:   "futureone.com/~damftp/AFT",
			Name:  "DC hosted",
			Error: true,
		},
		{
			URL:   "pages.ripco.com:8080/~devoid",
			Name:  "Devoid",
			Error: true,
		},
		{
			URL:   "206.245.196.81/affinity/affinity.htm",
			Name:  "Issue 5",
			Error: true,
		},
		{
			URL:   "www.iceonline.com/home/rahimk/aft.htm",
			Name:  "Rahimk",
			Error: true,
		},
	},
	"cybermail": []Website{
		{
			URL:   "cybermail.home.ml.org",
			Name:  "ml.org",
			Error: true,
		},
		{
			URL:   "www.dmn.com.au/~warez",
			Name:  "Cybermail",
			Error: true,
		},
	},
	"email-compilation": []Website{
		{
			URL:   "www.xs4all.nl/~blahh",
			Name:  "Source Compilation",
			Error: true,
		},
	},
	"scooby-snack-magazine": []Website{
		{
			URL:   "www.cris.com/~shadowkn",
			Name:  "Scooby Snack Magazine",
			Error: true,
		},
	},
	"anemia": []Website{
		{
			URL:   "www.anemia.org",
			Name:  "Anemia",
			Error: true,
		},
		{
			URL:   "anemia.base.org",
			Name:  "Anemia 97",
			Error: true,
		},
		{
			URL:   "www.geocities.com/SunsetStrip/Towers/5435",
			Name:  "Anemia 96",
			Error: true,
		},
	},
	"pirates-with-attitudes": []Website{
		{
			URL:   "https://en.wikipedia.org/wiki/Pirates_with_Attitudes",
			Name:  wikipedia,
			Error: false,
		},
		{
			URL:   "https://files.mpoli.fi/software/DOS/BBS/",
			Name:  "Metropoli BBS including PPEs by PWA",
			Error: false,
		},
		{
			URL:   "archives.thebbs.org/ra117a.htm",
			Name:  "153 PCBoard PPE's by PWA",
			Error: true,
		},
	},
	"united-group-international": []Website{
		{
			URL:   "https://web.archive.org/web/20050228211151/https://xakep.ru/magazine/xa/009/026/1.asp",
			Name:  "Interview with the founder",
			Error: false,
		},
	},
	"censor-design": []Website{
		{URL: "https://censordesign.com", Error: false, Name: "Censor Design"},
	},
}

func init() {
	for _, sites := range websites {
		slices.SortFunc(sites, func(a, b Website) int {
			if a.Error == b.Error {
				return 0
			}
			if !a.Error {
				return -1
			}
			return 1
		})
	}
}

// Find returns the website for the given uri.
// It returns an empty string if the uri is not known.
func Find(uri string) []Website {
	return websites[URI(uri)]
}

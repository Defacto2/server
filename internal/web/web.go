// Package web proves links and titles for recommended websites.
package web

// URI is the URL slug of the releaser.
type URI string

// URL is the historical URL of the releaser website.
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

func groups() Groups {
	return Groups{
		"acid-productions": []Website{
			{
				URL:  "https://www.acid.org",
				Name: "ACiD Productions",
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
		},
		"divine": []Website{
			{
				URL:        "dvn.org",
				Name:       "Divine",
				NotWorking: true,
			},
		},
		"drink-or-die": []Website{
			{
				URL:        "www.drinkordie.com",
				Name:       "Drink Or Die",
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
		},
		"hybrid": []Website{
			{
				URL:        "www.hybrid.to",
				Name:       "Hybrid",
				NotWorking: true,
			},
		},
		"insane-creators-enterprise": []Website{{
			URL:  "https://www.ice.org",
			Name: "iCE Advertisements",
		}},
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
		},
		"paradigm": []Website{
			{
				URL:        "www.pdmworld.com",
				Name:       "Paradigm",
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
				URL:        "www.razor1911.com",
				Name:       "Razor 1911",
				NotWorking: true,
			},
			{
				URL:        "www.laric.com/razor",
				Name:       "Razor 1911 Europe",
				NotWorking: true,
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
				URL:        "www.rcn.org",
				Name:       "Reality Check Network",
				NotWorking: true,
			},
		},
		"rebels": []Website{
			{
				URL:  "www.rebels.dk",
				Name: "Rebels Island",
			},
			{
				URL:        "www.rebels.org",
				Name:       "Rebels",
				NotWorking: true,
			},
		},
		"relativity": []Website{
			{
				URL:        "www.cyberbeach.net/~jester/relativity",
				Name:       "Relativity",
				NotWorking: true,
			},
		},
		"risciso": []Website{
			{
				URL:        "www.risc98.org",
				Name:       "RISCISO",
				NotWorking: true,
			},
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
		},
		"superior-art-creations": []Website{
			{
				URL:  "http://www.roysac.com/sac.html",
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
				URL:        "www.monmouth.com/~jionin",
				Name:       "The Game Review",
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
		},
		"x_force": []Website{
			{
				URL:        "www.xforce.org",
				Name:       "X-Force",
				NotWorking: true,
			},
		},
		"x_pression-design": []Website{
			{
				URL:        "www.xpression.org",
				Name:       "X-Pression Design",
				NotWorking: true,
			},
		},
	}
}

// Find returns the website for the given uri.
// It returns an empty string if the uri is not known.
func Find(uri string) []Website {
	if _, ok := groups()[URI(uri)]; ok {
		return groups()[URI(uri)]
	}
	return []Website{}
}

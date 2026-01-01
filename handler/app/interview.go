package app

// Package file interview.go contains the functions for the interview page.

// Groups is a collection of group interviews.
type Groups []Group

// Group is a collection of interviews with members of a group.
type Group struct {
	Name       string     // Name is the name of the group.
	Link       string     // Link is a local URL to the group.
	Interviews Interviews // Interviews is a list of interviews with members of the group.
}

// Interviews is a collection of Interviewee.
type Interviews []Interviewee

// Interviewee is a person who was interviewed with a link to the interview.
type Interviewee struct {
	Scener  string // Scener is the name of the person interviewed.
	Content string // Content is a short description of the interview.
	Link    string // Link is the URL to the interview.
	Year    int    // Year is the year the interview was conducted.
	Month   int    // Month is the month the interview was conducted.
}

// Interviewees returns a list of interviewees and their interviews.
// These are categorized by the group they were in at the time of the interview.
func Interviewees() Groups {
	i := Groups{
		{
			Name: "EMPRESS",
			Link: "empress",
			Interviews: Interviews{
				{
					Scener: "EMPRESS",
					Year:   2021, Month: 2,
					Content: "WIRED - The Woman Bulldozing Video Games' Toughest DRM.",
					Link:    "https://www.wired.com/story/empress-drm-cracking-denuvo-video-game-piracy/",
				},
			},
		},
		{
			Name: "The Humble Guys",
			Link: "the-humble-guys",
			Interviews: Interviews{
				{
					Scener: "Bryn Rogers",
					Year:   2012, Month: 8,
					Content: "The former member of The Humble Guys talks about his side group, \"Lamers of Power\".",
					Link:    "/f/ae2f55",
				},
				{
					Scener: "The Candy Man",
					Year:   1998, Month: 11,
					Content: "The co-founder of The Humble Guys talks about the group's history and their impact on the PC.",
					Link:    "https://wayback.defacto2.net/the-scene-news-from-1999-september-14/interview-002.html",
				},
			},
		},
		{
			Name: "Razor 1911",
			Link: "razor-1911",
			Interviews: Interviews{
				{
					Scener: "Evil Current",
					Year:   2012, Month: 8,
					Content: "The former head of Razor 1911 talks in a reddit IAmA.",
					Link:    "https://www.reddit.com/r/IAmA/comments/xusji/iama_former_member_of_razor_1911_amongst_many/",
				},
				{
					Scener: "The Playboy",
					Year:   2010, Month: 7,
					Content: "The former member of Razor 1911 talks in a reddit IAmA.",
					Link:    "https://www.reddit.com/r/IAmA/comments/ckobg/iama_exbbs_warez_scene_guy_ama/",
				},
				{
					Scener:  "Pitbull",
					Year:    2005,
					Content: "The former leader of Razor 1911 talks about life after being busted.",
					Link:    "/f/ab3914",
				},
				{
					Scener: "The Renegade Chemist",
					Year:   1996, Month: 6,
					Content: "A spotlight on the former leader of Razor 1911.",
					Link:    "/f/ac3d0c",
				},
				{
					Scener: "Doctor No",
					Year:   1992, Month: 3,
					Content: "Razor 1911 founder talks about the move to the PC scene.",
					Link:    "http://janeway.exotica.org.uk/target.php?idp=1873&idr=690&tgt=1",
				},
				{
					Scener: "Doctor No + Sector 9",
					Year:   1989, Month: 12,
					Content: "Founders of Razor 1911 talk about the group history, demos and the Amiga scene.",
					Link:    "http://janeway.exotica.org.uk/target.php?idp=1873&idr=690&tgt=1",
				},
			},
		},
		{
			Name: "Drink or Die",
			Link: "drink-or-die",
			Interviews: Interviews{
				{
					Scener: "BiGrAr",
					Year:   2002, Month: 10,
					Content: "Former member of Drink or Die and convicted pirate talks about his time in the group and prison.",
					Link:    "https://yro.slashdot.org/story/02/10/04/144217/former-drinkordie-member-chris-tresco-answers",
				},
				{
					Scener: "Bandido",
					Year:   1999, Month: 12,
					Content: "The council member of Drink or Die talks about life in The Scene.",
					Link:    "https://wayback.defacto2.net/apollo-x-demo-resources-1999-december-17/bandido.htm",
				},
			},
		},
		{
			Name: "Amnesia",
			Link: "amnesia",
			Interviews: Interviews{
				{
					Scener: "_TGK_",
					Year:   1998, Month: 6,
					Content: "The BBS courier group, Amnesia calls it quits.",
					Link:    "https://wayback.defacto2.net/scenelink-from-1998-june-25/features/issue/2/deathamnesia.html",
				},
			},
		},
		{
			Name: "International Network of Crackers",
			Link: "international-network-of-crackers",
			Interviews: Interviews{
				{
					Scener: "Coolhand",
					Year:   1998, Month: 6,
					Content: "Looking back at the history of the scene and INC.",
					Link:    "https://wayback.defacto2.net/scenelink-from-1998-june-25/features/issue/5/ch.html",
				},
				{
					Scener: "Line Noise",
					Year:   1993, Month: 6,
					Content: "Conducted with the current president of INC.",
					Link:    "/f/a72d0b",
				},
				{
					Scener: "Bar Manager",
					Year:   1993, Month: 2,
					Content: "The history of INC today.",
					Link:    "/f/a62913",
				},
			},
		},
		{
			Name: "The Dream Team",
			Link: "the-dream-team",
			Interviews: Interviews{
				{
					Scener: "T800",
					Year:   1996, Month: 1,
					Content: "The leader of the new The Dream Team.",
					Link:    "/f/aa3a34",
				},
				{
					Scener: "The Grim Reaper",
					Year:   1993, Month: 2,
					Content: "Member of The Dream Team and a coordinator of the ViSiON-X BBS software.",
					Link:    "/f/a1377e",
				},
				{
					Scener:  "Hard Core",
					Year:    1993,
					Content: "Founder of The Dream Team.",
					Link:    "/f/a729f9",
				},
				{
					Scener: "Belgarion",
					Year:   1991, Month: 10,
					Content: "Retired member of The Dream Team and the siteop of The Festering Pit (206).",
					Link:    "/f/a93a58",
				},
			},
		},
		{
			Name: "Public Enemy",
			Link: "public-enemy",
			Interviews: Interviews{
				{
					Scener: "Minor Threat",
					Year:   1994, Month: 9,
					Content: "\"This guy is a real Renaissance Man, hacker, programmer, " +
						"burglar, convict, star of stage and screen..\"",
					Link: "http://phrack.org/issues/46/5.html#article",
				},
			},
		},
		{
			Name: "Future Crew",
			Link: "future-crew",
			Interviews: Interviews{
				{
					Scener: "Purple Motion",
					Year:   1993, Month: 2,
					Content: "Member of the famous demoscene group Future Crew" +
						" talks about the group and working on the game, Max Payne" +
						" for Remedy Entertainment.",
					Link: "/f/a1377e",
				},
			},
		},
		{
			Name: "Fairlight",
			Link: "fairlight",
			Interviews: Interviews{
				{
					Scener:  "Ford Perfect",
					Year:    1993,
					Month:   1,
					Content: "Before the \"incident\"...",
					Link:    "/f/ac4680",
				},
				{
					Scener:  "Genesis",
					Year:    1991,
					Month:   10,
					Content: "Member of USA and Fairlight and the siteop of BBS-A-Holic (213).",
					Link:    "f/ad4af8",
				},
				{
					Scener: "Strider",
					Year:   1988, Month: 12,
					Content: "The co-founder of Fairlight talks about the group in its first year.",
					Link:    "http://janeway.exotica.org.uk/target.php?idp=6375&idr=1940&tgt=1",
				},
			},
		},
		{
			Name: "Apple II",
			Link: "appleii",
			Interviews: Interviews{
				Interviewee{
					Scener: "The Wizard",
					Year:   2012, Month: 10,
					Content: "Roland Gustafsson freelanced for Brøderbund and Strategic Simulations " +
						"supplying some of the first disk copy protections.",
					Link: "http://underground2e.free.fr/Underground/Deplombage/Interviews/The_Wizard/Roland_Gustafsson.html",
				},
				Interviewee{
					Scener: "Aldo Reset",
					Year:   2008, Month: 5,
					Content: "Beginning in 1982, he is one of the earliest crackers on any Scene, " +
						"Aldo Reset was both media savy and a famous French cracker.",
					Link: "https://textfiles.meulie.net/exhibits/frenchunderground/" +
						"Deplombage/Interviews/Aldo_Reset/Interview_Aldo_Reset.html",
				},
				Interviewee{
					Scener: "JPL",
					Year:   2006,
					Content: "Another early French cracker who got their start in 1983, " +
						"JPL often kept his cracks and self-branding to a minimum.",
					Link: "https://textfiles.meulie.net/exhibits/frenchunderground/Deplombage/Interviews/Jpl/Interview_JPL.html",
				},
			},
		},
		{
			Name: "Retirements",
			Link: "",
			Interviews: Interviews{
				{
					Scener: "ChinaBlue",
					Year:   1998, Month: 6,
					Content: "Talks about her retirement and the 'bust or be busted' paranoia of the scene.",
					Link:    "https://wayback.defacto2.net/scenelink-from-1998-june-25/features/issue/5/china-interview.html",
				},
			},
		},
	}
	return i
}

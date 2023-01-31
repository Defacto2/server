package tags

import "strings"

// URI is a unique URL slug for the tag.
type URI map[Tag]string

// Name is the tags displayed title.
type Name map[Tag]string

// Info is a short description of the tag.
type Info map[Tag]string

func (t Tag) String() string {
	return URIs[t]
}

var URIs = URI{
	Announcement: "announcements",
	ANSIEditor:   "ansieditor",
	AppleII:      "appleii",
	AtariST:      "atarist",
	BBS:          "bbs",
	Logo:         "logo",
	Bust:         "takedown",
	Drama:        "politics",
	Rule:         "scenerules",
	Tool:         "programmingtool",
	Intro:        "releaseadvert",
	Demo:         "demo",
	ForSale:      "forsale",
	Ftp:          "ftp",
	GameHack:     "gamehack",
	Job:          "groupapplication",
	Guide:        "guide",
	Interview:    "interview",
	Mag:          "magazine",
	News:         "newsmedia",
	Nfo:          "releaseinformation",
	NfoTool:      "nfotool",
	Pack:         "package",
	Proof:        "releaseproof",
	Restrict:     "internaldocument",
	Install:      "releaseinstall",
	ANSI:         "ansi",
	Audio:        "audio",
	DataB:        "database",
	DOS:          "dos",
	Markup:       "markup",
	Image:        "image",
	Java:         "java",
	Linux:        "linux",
	Mac:          "mac10",
	PCB:          "pcb",
	PDF:          "pdf",
	PHP:          "php",
	Text:         "text",
	TextAmiga:    "textamiga",
	Video:        "video",
	Windows:      "windows",
}

// TagByURI returns the tag belonging to the URI slug.
func TagByURI(slug string) Tag {
	for key, value := range URIs {
		if strings.ToLower(slug) == value {
			return key
		}
	}
	return -1
}

// For consistency, the names and descriptions should be in US English.

var Names = URI{
	Announcement: "Announcement",
	ANSIEditor:   "ANSI editor",
	AppleII:      "Apple II",
	AtariST:      "Atari ST",
	BBS:          "BBS",
	Logo:         "Brand art or logo",
	Bust:         "Bust or takedown",
	Drama:        "Community drama",
	Rule:         "Community standard",
	Tool:         "Computer tool",
	Intro:        "Cracktro or intro",
	Demo:         "Demo program",
	ForSale:      "For sale",
	Ftp:          "FTP",
	GameHack:     "Game hack",
	Job:          "Group role or job",
	Guide:        "Guides and how-tos",
	Interview:    "Interview",
	Mag:          "Magazine",
	News:         "Mainstream news",
	Nfo:          "NFO file or scene release",
	NfoTool:      "NFO tool",
	Pack:         "Filepack",
	Proof:        "Release proof",
	Restrict:     "Restricted",
	Install:      "Scene software install",
	ANSI:         "ANSI",
	Audio:        "Music",
	DataB:        "Database",
	DOS:          "DOS",
	Markup:       "HTML",
	Image:        "Image",
	Java:         "Java",
	Linux:        "Linux",
	Mac:          "macOS",
	PCB:          "PCBoard",
	PDF:          "PDF",
	PHP:          "Script",
	TextAmiga:    "Text for Amiga",
	Text:         "Text or ASCII",
	Video:        "Video",
	Windows:      "Windows",
}

var Infos = Info{
	Announcement: "Public announcements by Scene groups and organizations",
	ANSIEditor:   "Programs that enable you to create and edit ANSI and ASCII art",
	AppleII:      "Files about the Scene on the Apple II computer platform",
	AtariST:      "Files on the Scene on the Atari ST computer platform",
	BBS:          "Files about the Scene operating over telephone-based BBS (Bulletin Board System) systems",
	Logo:         "Branding logos used by scene groups and organizations",
	Bust: "First-hand accounts and third party reports on the arrest, " +
		"bust or take-down of an active person in the scene or a scene organizations",
	Drama: "Used for anything political that doesn't fall into the other categories. " +
		"It usually contains documents where people, groups, or organizations are complaining",
	Rule: "Various codes of conduct and agreements created by scene groups and organizations",
	Tool: "Miscellaneous tools, including fixes, intro generators, and BBS software",
	Intro: "A multimedia program designed to promote a scene group or organization, " +
		"known as a cracktro, crack intro, or loader",
	Demo:    "An artistic multimedia program designed to promote a demo group or collective",
	ForSale: "Adverts for commercial physical goods and online services, varying in legality",
	Ftp:     "Files about the scene operating over Internet-based FTP (File Transfer Protocol) servers",
	GameHack: "Trainers, dox, cheats, and walk-throughs, which include guides, " +
		"how-to documents, and tools to complete games",
	Job:       "Documents used by scene groups to advertise or enroll new members",
	Guide:     "Guides and how-to documents on how to hack and crack the workings of the scene",
	Interview: "Conversations with the personalities of the scene",
	Mag:       "Reports and written articles created by scene members about the scene",
	News:      "Mainstream media outlets report on the scene",
	Nfo:       "A text file or readme describes a scene release, group, or organization",
	NfoTool:   "Programs that enable you to create or view NFO text files",
	Proof:     "Evidence of the source media, usually a photo or scanned image",
	Restrict:  "Documents created by scene groups that are often never intended to be public",
	Install:   "A program to help an end-user install a scene release",
	ANSI:      "Colored, text-based computer art form widely used on Bulletin Board Systems",
	Audio:     "Music or audio sound clips",
	DataB: "A structured collection of data stored in particular formats, including " +
		"spreadsheets such as Microsoft Excel or databases such as MySQL",
	DOS:    "DOS programs require using Microsoft's DOS operating system on x86-compatible CPUs",
	Markup: "Web pages or documents in HTML or text documents in a mark-up language",
	Image:  "Digital art, pixel art, or photos",
	Java:   "Java software that requires the use of the Sun Microsystems or Oracle Java platform",
	Linux:  "Linux programs are software for the Linux platform, including server and desktop distributions",
	Mac:    "Programs for the various operating systems created by Apple under the Mac and Macintosh brands",
	Pack: "A curated bundle of scene-related files is stored and distributed in a " +
		"compressed archive file, often in ZIP or 7z formats",
	PCB: "Colored encoded text on Bulletin Board Systems " +
		"and plain text documents embedded with PCBoard control codes",
	PDF:       "A document compiled as PDF (Adobe Portable Document Format)",
	PHP:       "Scripts and interpreted programs that are in an interpreted programming language",
	TextAmiga: "Text documents and text-based computer art for the Amiga in the Latin-1 character set",
	Text:      "Text documents and text-based computer art that use an ASCII-compliant character set",
	Video:     "A film, video, or multimedia animation",
	Windows:   "These programs require the use of Microsoft's Windows operating system, working on Intel-compatible CPUs",
}

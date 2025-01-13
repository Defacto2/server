package tags

// Package file strings.go contains the tag strings and their descriptions.

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrPlatform = errors.New("invalid platform")
	ErrTag      = errors.New("invalid tag")
)

// URIS is a unique string for the tag.
type URIS map[Tag]string

// Name is the tags displayed title.
type Name map[Tag]string

// Info is a short description of the tag.
type Info map[Tag]string

func (t Tag) String() string {
	return URIs()[t]
}

// URIs returns the URI slugs for the tags.
func URIs() URIS {
	return URIS{
		AreaCodes:    "areacodes",
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
}

// TagByURI returns the tag belonging to the URI slug.
func TagByURI(slug string) Tag {
	for key, value := range URIs() {
		if strings.ToLower(slug) == value {
			return key
		}
	}
	return -1
}

// Names or titles of the URI tags.
// For consistency, the names and descriptions should be in US English and singular.
func Names() URIS {
	return URIS{
		Announcement: "announcement",
		ANSIEditor:   "ansi editor",
		AppleII:      "Apple II",
		AtariST:      "Atari ST",
		BBS:          "BBS",
		Logo:         "brand art or logo",
		Bust:         "bust or takedown",
		Drama:        "community drama",
		Rule:         "community standard",
		Tool:         "computer tool",
		Intro:        "cracktro or intro",
		Demo:         "demo program",
		ForSale:      "for sale",
		Ftp:          "FTP",
		GameHack:     "game hack",
		Job:          "group role or job",
		Guide:        "guide or how-to",
		Interview:    "interview",
		Mag:          "magazine",
		News:         "mainstream news",
		Nfo:          "nfo file or scene release",
		NfoTool:      "nfo tool",
		Pack:         "filepack",
		Proof:        "release proof",
		Restrict:     "restricted",
		Install:      "scene software installer",
		ANSI:         "ansi",
		Audio:        "music",
		DataB:        "database",
		DOS:          "Dos",
		Markup:       "HTML",
		Image:        "image",
		Java:         "Java",
		Linux:        "Linux",
		Mac:          "macOS",
		PCB:          "PCBoard",
		PDF:          "PDF",
		PHP:          "script",
		TextAmiga:    "text for the Amiga",
		Text:         "text or ascii",
		Video:        "video",
		Windows:      "Windows",
	}
}

// Determiner is the article used before the tag name.
func Determiner() URIS {
	return URIS{
		Announcement: "an",
		ANSIEditor:   "an",
		AppleII:      "an",
		AtariST:      "an",
		BBS:          "a",
		Logo:         "a",
		Bust:         "a",
		Drama:        "a",
		Rule:         "a",
		Tool:         "a",
		Intro:        "a",
		Demo:         "a",
		ForSale:      "a",
		Ftp:          "an",
		GameHack:     "a",
		Job:          "a",
		Guide:        "a",
		Interview:    "an",
		Mag:          "a",
		News:         "a",
		Nfo:          "a",
		NfoTool:      "a",
		Pack:         "a",
		Proof:        "a",
		Restrict:     "an",
		Install:      "a",
		ANSI:         "an",
		Audio:        "an",
		DataB:        "a",
		DOS:          "a",
		Markup:       "a",
		Image:        "an",
		Java:         "a",
		Linux:        "a",
		Mac:          "a",
		PCB:          "a",
		PDF:          "a",
		PHP:          "a",
		TextAmiga:    "a",
		Text:         "a",
		Video:        "a",
		Windows:      "a",
	}
}

// NameByURI returns the name of a tag belonging to the URI slug.
func NameByURI(slug string) string {
	slug = strings.ToLower(strings.TrimSpace(slug))
	for key, value := range Names() {
		if slug == key.String() {
			return value
		}
	}
	return fmt.Sprintf("error: unknown slug %q", slug)
}

// Infos returns short descriptions of the tags.
func Infos() Info {
	return Info{
		Announcement: "public announcements by Scene groups and organizations",
		ANSIEditor:   "programs that enable you to create and edit ansi and text art",
		AppleII:      "files about The Scene on the Apple II computer platform",
		AtariST:      "files on The Scene on the Atari ST computer platform",
		BBS:          "BBS (Bulletin Board System) files about The Scene operating over the telephone network",
		Logo:         "brand logos used by scene groups and organizations",
		Bust: "first-hand accounts and third party reports on the arrest, " +
			"bust or take-down of an active person in the scene or a scene organizations",
		Drama: "used for anything political that doesn't fall into the other categories. " +
			"It usually contains documents where people, groups, or organizations are complaining",
		Rule: "various codes of conduct and agreements created by scene groups and organizations",
		Tool: "miscellaneous tools, including fixes, intro generators, and BBS software",
		Intro: "a multimedia program designed to promote a scene group or organization, " +
			"known as a cracktro, crack intro, or loader",
		Demo:    "an artistic multimedia program designed to promote a demo group or collective",
		ForSale: "adverts for commercial physical goods and online services, varying in legality",
		Ftp:     "files about the scene operating over Internet-based FTP (File Transfer Protocol) servers",
		GameHack: "trainers, dox, cheats, and walk-throughs, which include guides, " +
			"how-to documents, and tools to complete games",
		Job:       "documents used by scene groups to advertise or enroll new members",
		Guide:     "guides and how-to documents on how to hack and crack the workings of The Scene",
		Interview: "conversations with the personalities of The Scene",
		Mag:       "reports and written articles created by scene members about The Scene",
		News:      "mainstream media outlets report on The Scene",
		Nfo:       "a text file or readme describes a scene release, group, or organization",
		NfoTool:   "programs that enable you to create or view nfo text files",
		Proof:     "evidence of the source media, usually a photo or scanned image",
		Restrict:  "documents created by scene groups that are often never intended to be public",
		Install:   "a program to help an end-user install a scene release",
		ANSI:      "colored, text-based computer art form widely used on Bulletin Board Systems",
		Audio:     "music or audio sound clips",
		DataB: "a structured collection of data stored in particular formats, including " +
			"spreadsheets such as Microsoft Excel or databases such as MySQL",
		DOS:    "Dos programs require using Microsoft's DOS operating system on x86-compatible CPUs",
		Markup: "web pages or documents in HTML or text documents in a mark-up language",
		Image:  "digital art, pixel art, or photos",
		Java:   "Java software that requires the use of the Sun Microsystems or Oracle Java platform",
		Linux:  "Linux programs are software for the Linux platform, including server and desktop distributions",
		Mac:    "programs for the various operating systems created by Apple under the Mac and Macintosh brands",
		Pack: "a curated bundle of scene-related files is stored and distributed in a " +
			"compressed archive file, often in ZIP or 7z formats",
		PCB: "colored encoded text on Bulletin Board Systems " +
			"and plain text documents embedded with PCBoard control codes",
		PDF:       "a document compiled as PDF (Adobe Portable Document Format)",
		PHP:       "scripts and interpreted programs that are in an interpreted programming language",
		TextAmiga: "text documents and text-based computer art for the Amiga in the Latin-1 character set",
		Text:      "text documents and text-based computer art that use an ASCII-compliant character set",
		Video:     "a film, video, or multimedia animation",
		Windows:   "these programs require the use of Microsoft's Windows operating system, working on Intel-compatible CPUs",
	}
}

// Description returns the short description of the tag.
func Description(tag string) (string, error) {
	t := TagByURI(tag)
	if t == -1 {
		return "", fmt.Errorf("%s: %w", tag, ErrTag)
	}
	s := Infos()[t]
	return s, nil
}

// Platform returns the human readable platform and tag name.
func Platform(platform, tag string) (string, error) {
	p, t := TagByURI(platform), TagByURI(tag)
	if p == -1 {
		return "", fmt.Errorf("%s: %w", platform, ErrPlatform)
	}
	if t == -1 {
		return "", fmt.Errorf("%s: %w", tag, ErrTag)
	}
	return Humanize(p, t), nil
}

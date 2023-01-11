package meta

import (
	"context"
	"log"

	"github.com/bengarrett/df2023/postgres/models"

	"github.com/bengarrett/df2023/postgres"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type Tag int

type URI map[Tag]string

type Name map[Tag]string

type Info map[Tag]string

type Count map[Tag]int

const (
	Announcement Tag = iota
	ANSIEditor
	AppleII
	AtariST
	BBS
	Logo
	Bust
	Drama
	Rule
	Tool
	Intro
	Demo
	ForSale
	Ftp
	GameHack
	Job
	Guide
	Interview
	Mag
	News
	Nfo
	NfoTool
	Proof
	Restrict
	Install
)

const CategoryCount = int(Install + 1)

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
	Proof:        "releaseproof",
	Restrict:     "internaldocument",
	Install:      "releaseinstall",
}

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
	Proof:        "Release proof",
	Restrict:     "Restricted",
	Install:      "Scene software install",
}

var Infos = Info{
	Announcement: "Public announcements by Scene groups and organisations",
	ANSIEditor:   "Programs that enable you to create and edit ANSI and ASCII art",
	AppleII:      "Files pertaining to the Scene on the Apple II computer platform",
	AtariST:      "Files pertaining to the Scene on the Atari ST computer platform",
	BBS:          "Files pertaining to the Scene operating over telephone based BBS (Bulletin Board System) systems",
	Logo:         "Branding logos used by scene groups and organisations",
	Bust: "First hand accounts and third party reports on the arrest, " +
		"bust or take-down of an active person in the scene or a scene organisation",
	Drama: "Used for anything political that doesn't fall into the other " +
		"categories. Usually contains documents where people, groups or " +
		"organisations are complaining",
	Rule: "Various codes of conduct and agreements created by scene groups and organisations",
	Tool: "Miscellaneous tools including fixes, intro generators and BBS software",
	Intro: "A multimedia program that is designed to promote a scene group or organisation. " +
		"Otherwise known as a cracktro, crack intro or loader",
	Demo:    "An artistic multimedia program that is designed to promote a demo group or collective",
	ForSale: "Adverts for commercial physical goods and online services, varying in legality",
	Ftp: "Files pertaining to the scene operating over Internet based " +
		"FTP (File Transfer Protocol) servers",
	GameHack: "Trainers, dox, cheats, and walk-throughs, which include guides, " +
		"how-to documents and tools to complete games",
	Job:       "Documents used by scene groups to advertise or enrol new members",
	Guide:     "Guides and how-to documents on how to hack and crack or on the workings of the scene",
	Interview: "Conversations conducted with scene personalities",
	Mag:       "Reports and written articles created by scene members about the scene",
	News:      "Mainstream media outlets reports on the scene",
	Nfo:       "A text file or readme used to describe a scene release, group or organisation",
	NfoTool:   "Programs that enable you to create or view NFO text files",
	Proof:     "Evidence of the source media, usually a photo or scanned image",
	Restrict:  "Documents created by scene groups that were often never intended to be public",
	Install:   "A program to help an end-user install a scene release",
}

var Counts = Count{
	Announcement: 0,
	ANSIEditor:   0,
	AppleII:      0,
	AtariST:      0,
	BBS:          0,
	Logo:         0,
	Bust:         0,
	Drama:        0,
	Rule:         0,
	Tool:         0,
	Intro:        0,
	Demo:         0,
	ForSale:      0,
	Ftp:          0,
	GameHack:     0,
	Job:          0,
	Guide:        0,
	Interview:    0,
	Mag:          0,
	News:         0,
	Nfo:          0,
	NfoTool:      0,
	Proof:        0,
	Restrict:     0,
	Install:      0,
}

type Meta struct {
	URI   string
	Name  string
	Info  string
	Count int
}

var Categories []Meta = New()

func New() []Meta {
	var m = make([]Meta, len(URIs))
	i := -1
	for key, val := range URIs {
		i++
		count := Counts[key]
		m[i] = Meta{
			URI:   val,
			Name:  Names[key],
			Info:  Infos[key],
			Count: count,
		}
		// TODO: cache the results and move the function / cache to /models/custom.go
		// https://stackoverflow.com/questions/67788292/add-a-cache-to-a-go-function-as-if-it-were-a-static-member
		if count == 0 {
			t := key
			defer func(i int, t Tag) {
				ctx := context.Background()
				db, err := postgres.ConnectDB()
				if err != nil {
					log.Fatalln(err)
				}
				val, err := models.Files(
					Where("section = ?", URIs[t])).Count(ctx, db)
				if err != nil {
					log.Fatalln(err)
				}
				m[i].Count = int(val)
			}(i, t)
		}
	}
	return m
}

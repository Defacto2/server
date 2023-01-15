// Package meta are categories and platform metadata used to classify served files.
package meta

import (
	"context"
	"log"
	"strings"

	"github.com/Defacto2/server/postgres/models"

	"github.com/Defacto2/server/postgres"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type Tag int

type Count map[Tag]int

type Meta struct {
	URI   string
	Name  string
	Info  string
	Count int
}

const (
	FirstCategory Tag = Announcement
	FirstPlatform Tag = ANSI
	LastCategory  Tag = Install
	LastPlatform  Tag = Windows
	CategoryCount     = int(FirstCategory + LastCategory + 1)
	PlatformCount     = int(LastPlatform - FirstPlatform + 1)
)

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
	Pack
	Proof
	Restrict
	Install
	ANSI
	Audio
	DataB
	DOS
	Markup
	Image
	Java
	Linux
	Mac
	PCB
	PDF
	PHP
	Text
	TextAmiga
	Video
	Windows
)

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
	ANSI:         0,
	Audio:        0,
	DataB:        0,
	DOS:          0,
	Markup:       0,
	Image:        0,
	Java:         0,
	Linux:        0,
	Mac:          0,
	Pack:         0,
	PCB:          0,
	PDF:          0,
	PHP:          0,
	TextAmiga:    0,
	Text:         0,
	Video:        0,
	Windows:      0,
}

var Tags []Meta = GetTags()

func GetApps() [5]string {
	return [5]string{
		URIs[DOS],
		URIs[Java],
		URIs[Linux],
		URIs[Windows],
		URIs[Mac]}
}

func GetMetaByName(name string) Meta {
	for _, m := range Tags {
		if strings.EqualFold(m.Name, name) {
			return m
		}
	}
	return Meta{}
}

func GetTags() []Meta {
	var m = make([]Meta, LastPlatform+1)
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
		if count > 0 {
			continue
		}
		t := key
		defer func(i int, t Tag) {
			ctx := context.Background()
			db, err := postgres.ConnectDB()
			if err != nil {
				log.Fatalln(err) // TODO: zap log
			}
			clause := "section = ?"
			if t >= FirstPlatform {
				clause = "platform = ?"
			}
			val, err := models.Files(
				Where(clause, URIs[t])).Count(ctx, db)
			if err != nil {
				log.Fatalln(err)
			}
			m[i].Count = int(val)
		}(i, t)
	}
	return m
}

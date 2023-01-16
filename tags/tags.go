// Package tags are categories and platform metadata used to classify served files.
package tags

import (
	"context"
	"log"
	"strings"

	"github.com/Defacto2/server/postgres/models"

	"github.com/Defacto2/server/postgres"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// TODO: cache the results and move the function / cache to /models/custom.go
// https://stackoverflow.com/questions/67788292/add-a-cache-to-a-go-function-as-if-it-were-a-static-member

// TagData holds the tag information.
type TagData struct {
	URI   string // URI is a unique URL slug for the tag.
	Name  string // Name is the tags displayed title.
	Info  string // Info is a short description of the tag.
	Count int    // Count is the results of file count query for the tag.
}

// Tag is the unique ID.
type Tag int

const (
	// FirstCategory is the first tag marked as a category.
	FirstCategory Tag = Announcement
	// FirstPlatform is the first tag marked as a platform.
	FirstPlatform Tag = ANSI
	// LastCategory is the final tag marked as a category.
	LastCategory Tag = Install
	// LastPlatform is the final tag marked as a platform.
	LastPlatform Tag = Windows
	// CategoryCount is the number of tags used as a category.
	CategoryCount = int(FirstCategory + LastCategory + 1)
	// PlatformCount is the number of tags used as a platform.
	PlatformCount = int(LastPlatform - FirstPlatform + 1)
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

// Sum the numbers of files with the tag.
type Sum map[Tag]int

// Sums stores the results of file count query for each tag.
var Sums = make(Sum, Windows+1)

// Tags contains data for all the tags used by the web application.
var Tags []TagData = All()

// OSTags returns the tags that flag an operating system.
func OSTags() [5]string {
	return [5]string{
		URIs[DOS],
		URIs[Java],
		URIs[Linux],
		URIs[Windows],
		URIs[Mac]}
}

// TagByName returns the named tag.
func TagByName(name string) TagData {
	for _, m := range Tags {
		if strings.EqualFold(m.Name, name) {
			return m
		}
	}
	return TagData{}
}

// All the tags and assoicated data.
func All() []TagData {
	var m = make([]TagData, LastPlatform+1)
	i := -1
	for key, val := range URIs {
		i++
		count := Sums[key]
		m[i] = TagData{
			URI:   val,
			Name:  Names[key],
			Info:  Infos[key],
			Count: count,
		}
		if count > 0 {
			continue
		}
		t := key
		defer func(i int, t Tag) {
			m[i].Count = int(counter(i, t))
		}(i, t)
	}
	return m
}

func counter(i int, t Tag) int64 {
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		log.Fatalln(err) // TODO: zap log
	}
	clause := "section = ?"
	if t >= FirstPlatform {
		clause = "platform = ?"
	}
	sum, err := models.Files(
		Where(clause, URIs[t])).Count(ctx, db)
	if err != nil {
		log.Fatalln(err)
	}
	return sum
}

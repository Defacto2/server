// Package tags are categories and platform metadata used to classify served files.
package tags

import (
	"context"
	"strings"
	"sync"

	"github.com/Defacto2/server/pkg/postgres"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/zap"
)

// TagData holds the tag information.
type TagData struct {
	URI   string // URI is a unique URL slug for the tag.
	Name  string // Name is the tags displayed title.
	Info  string // Info is a short description of the tag.
	Count int    // Count is the results of file count query for the tag.
}

// T is a lockable collection of tags, to stop potential race conditions
// when writing to the map containing the tagdata list.
type T struct {
	Mu   sync.RWMutex
	List []TagData
}

// ByName returns the data of the named tag.
func (t *T) ByName(name string, log *zap.SugaredLogger) TagData {
	if Tags.List == nil {
		t.Build(log)
	}
	for _, m := range Tags.List {
		if strings.EqualFold(m.Name, name) {
			return m
		}
	}
	return TagData{}
}

// Build the tags and collect the statistical data.
func (t *T) Build(log *zap.SugaredLogger) {
	t.List = make([]TagData, LastPlatform+1)
	i := -1
	for key, val := range URIs() {
		i++
		count := Sums[key]
		t.Mu.Lock()
		t.List[i] = TagData{
			URI:   val,
			Name:  Names()[key],
			Info:  Infos()[key],
			Count: count,
		}
		t.Mu.Unlock()
		if count > 0 {
			continue
		}
		tg := key
		defer func(i int, tg Tag) {
			t.Mu.Lock()
			t.List[i].Count = int(counter(tg, log))
			t.Mu.Unlock()
		}(i, tg)
	}
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
var Sums = make(Sum, Windows+1) //nolint:gochecknoglobals

// Tags contains data for all the tags used by the web application.
var Tags = T{} //nolint:gochecknoglobals

// OSTags returns the tags that flag an operating system.
func OSTags() [5]string {
	return [5]string{
		URIs()[DOS],
		URIs()[Java],
		URIs()[Linux],
		URIs()[Windows],
		URIs()[Mac],
	}
}

// count the number of files with the tag.
func counter(t Tag, log *zap.SugaredLogger) int64 {
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		log.Errorf("Could not connect to the database: %s.", err)
		return -1
	}
	clause := "section = ?"
	if t >= FirstPlatform {
		clause = "platform = ?"
	}
	sum, err := models.Files(
		qm.Where(clause, URIs()[t])).Count(ctx, db)
	if err != nil {
		log.Errorf("Could not sum the records associated with tags: %s.", err)
		return -1
	}
	return sum
}

// Package tags are categories and platform metadata used to classify served files.
package tags

import (
	"context"
	"fmt"
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

// Humanize returns the human readable name of the platform and section tags combined.
func Humanize(platform, section Tag) string {
	switch platform {
	case ANSI:
		switch section {
		case BBS:
			return "A BBS ansi advert"
		case Ftp:
			return "An ansi advert for an FTP site"
		case Logo:
			return "An ansi logo"
		case Nfo:
			return "An nfo text in ansi format"
		case Pack:
			return "An filepack of ansi files"
		}
	case DataB:
		switch section {
		case Nfo:
			return "A database of releases"
		default:
			return fmt.Sprintf("A %s database", Names()[section])
		}
	case Markup:
		return fmt.Sprintf("A %s in HTML", Names()[section])
	case Image:
		switch section {
		case BBS:
			return "A BBS advert image"
		case ForSale:
			return "An image advertisement"
		case Pack:
			return "A filepack of images"
		case Proof:
			return "A proof of release photo"
		}
	case PDF:
		return fmt.Sprintf("A %s as a PDF document", Names()[section])
	case Text:
		switch section {
		case AtariST:
			return "A textfile for the Atari ST"
		case AppleII:
			return "A textfile for the Apple II"
		case BBS:
			return "A text advert for a BBS"
		case ForSale:
			return "A textfile advert"
		case Ftp:
			return "A text advert for an FTP site"
		case Mag:
			return "A magazine textfile"
		case Nfo:
			return "An nfo textfile"
		case Pack:
			return "A filepack of textfiles"
		case Restrict:
			return "An textfile with restricted content"
		default:
			return fmt.Sprintf("A %s textfile", Names()[section])
		}
	case TextAmiga:
		switch section {
		case BBS:
			return "An Amiga text advert for a BBS"
		case ForSale:
			return "An Amiga textfile advert"
		case Ftp:
			return "An Amiga text advert for an FTP site"
		case Mag:
			return "An Amiga magazine textfile"
		case Nfo:
			return "An Amiga nfo textfile"
		case Restrict:
			return "An Amiga textfile with restricted content"
		}
	case Video:
		return fmt.Sprintf("A %s video", Names()[section])
	case Windows:
		switch section {
		case Demo:
			return "A demo on Windows"
		case Install:
			return "A Windows installer"
		case Intro:
			return "A Windows intro"
		case Job:
			return "A trial crackme for Windows"
		case Pack:
			return "A filepack of Windows programs"
		}
	case DOS:
		switch section {
		case BBS:
			return "A BBStro on MS-Dos"
		case Demo:
			return "A demo on MS-Dos"
		case ForSale:
			return "An advertisement on MS-Dos"
		case GameHack:
			return "A trainer or hack on MS-Dos"
		case Install:
			return "A MS-Dos installer"
		case Intro:
			return "A intro for MS-Dos"
		case Pack:
			return "A filepack of MS-Dos programs"
		}
	}
	return fmt.Sprintf("A %s %s", Names()[platform], Names()[section])
}

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

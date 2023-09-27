// Package tags are categories and platform metadata used to classify served files.
package tags

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.uber.org/zap"
)

// The dos, app funcmap handler must match the format and syntax of MS-DOS that's used here.
const msDos = "MS Dos"

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
func (t *T) ByName(z *zap.SugaredLogger, name string) TagData {
	if Tags.List == nil {
		t.Build(z)
	}
	for _, m := range Tags.List {
		if strings.EqualFold(m.Name, name) {
			return m
		}
	}
	return TagData{}
}

// Build the tags and collect the statistical data.
func (t *T) Build(z *zap.SugaredLogger) {
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
			t.List[i].Count = int(counter(z, tg))
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
	switch section {
	case News:
		return news(platform)
	case Restrict:
		return restrict(platform)
	}
	switch platform {
	case ANSI:
		switch section {
		case BBS:
			return "a BBS ansi advert"
		case Ftp:
			return "an ansi advert for an FTP site"
		case Logo:
			return "an ansi logo"
		case Nfo:
			return "an nfo text in ansi format"
		case Pack:
			return "an filepack of ansi files"
		}
	case Audio:
		switch section {
		case Intro:
			return "a chiptune or intro music"
		}
	case DataB:
		switch section {
		case Nfo:
			return "a database of releases"
		default:
			return fmt.Sprintf("a %s database", Names()[section])
		}
	case Markup:
		return fmt.Sprintf("a %s in HTML", Names()[section])
	case Image:
		switch section {
		case BBS:
			return "a BBS advert image"
		case ForSale:
			return "an image advertisement"
		case Pack:
			return "a filepack of images"
		case Proof:
			return "a proof of release photo"
		}
	case PDF:
		return fmt.Sprintf("a %s as a PDF document", Names()[section])
	case Text:
		switch section {
		case AtariST:
			return "a textfile for the Atari ST"
		case AppleII:
			return "a textfile for the Apple II"
		case BBS:
			return "a text advert for a BBS"
		case ForSale:
			return "a textfile advert"
		case Ftp:
			return "a text advert for an FTP site"
		case Mag:
			return "a magazine textfile"
		case Nfo:
			return "an nfo textfile"
		case Pack:
			return "a filepack of textfiles"
		default:
			return fmt.Sprintf("A %s textfile", Names()[section])
		}
	case TextAmiga:
		switch section {
		case BBS:
			return "an Amiga text advert for a BBS"
		case ForSale:
			return "an Amiga textfile advert"
		case Ftp:
			return "an Amiga text advert for an FTP site"
		case Mag:
			return "an Amiga magazine textfile"
		case Nfo:
			return "an Amiga nfo textfile"
		}
	case Video:
		switch section {
		case ForSale, Logo, Intro:
			return "a bumper video"
		}
		return fmt.Sprintf("A %s video", Names()[section])
	case Windows:
		switch section {
		case Demo:
			return "a demo on Windows"
		case Install:
			return "a Windows installer"
		case Intro:
			return "a Windows intro"
		case Job:
			return "a trial crackme for Windows"
		case Pack:
			return "a filepack of Windows programs"
		}
	case DOS:
		switch section {
		case BBS:
			return "a BBStro on " + msDos
		case Demo:
			return "a demo on " + msDos
		case ForSale:
			return "an advertisement on " + msDos
		case GameHack:
			return "a trainer or hack on " + msDos
		case Install:
			return "a " + msDos + " installer"
		case Intro:
			return "a intro for " + msDos
		case Pack:
			return "a filepack of " + msDos + " programs"
		}
	}
	return fmt.Sprintf("A %s %s", Names()[platform], Names()[section])
}

func news(platform Tag) string {
	switch platform {
	case Image:
		return "a screenshot of an article from a news outlet"
	case Markup:
		return "a HTML copy of an article from a news outlet"
	case PDF:
		return "a PDF of an article from a news outlet"
	case Text:
		return "a textfile copy of an article from a news outlet"
	case TextAmiga:
		return "an Amiga textfile copy of an article from a news outlet"
	default:
		return fmt.Sprintf("a %s from a news outlet", Names()[platform])
	}
}

func restrict(platform Tag) string {
	switch platform {
	case ANSI:
		return "a insider ansi textfile"
	case Text:
		return "a insider textfile"
	case TextAmiga:
		return "an insider Amiga textfile"
	default:
		return fmt.Sprintf("a insider %s file", Names()[platform])
	}
}

// Humanizes returns the human readable name plurals of the platform and section tags combined.
func Humanizes(platform, section Tag) string {
	switch platform {
	case ANSI:
		return ansi(section)
	case Audio:
		return "music, chiptunes and audio samples"
	case DataB:
		switch section {
		case Nfo:
			return "databases of releases"
		default:
			return fmt.Sprintf("%s databases", Names()[section])
		}
	case DOS:
		return dos(section)
	case Image:
		return image(section)
	case Java:
		return fmt.Sprintf("%s for Java", Names()[section])
	case Linux:
		return fmt.Sprintf("%s for Linux and BSD", Names()[section])
	case Markup:
		return fmt.Sprintf("%s as HTML files", Names()[section])
	case Mac:
		return fmt.Sprintf("%s for Macintosh and macOS", Names()[section])
	case PDF:
		return fmt.Sprintf("%s as PDF documents", Names()[section])
	case PHP:
		return fmt.Sprintf("%s for scripting languages", Names()[section])
	case Text:
		return text(section)
	case TextAmiga:
		return textAmiga(section)
	case Video:
		return "videos and animations"
	case Windows:
		return windows(section)
	}
	if platform < 0 {
		return emptyPlatform(section)
	}
	if section < 0 {
		return fmt.Sprintf("%ss", Names()[platform])
	}
	return fmt.Sprintf("%ss %ss", Names()[platform], Names()[section])
}

func ansi(section Tag) string {
	switch section {
	case BBS:
		return "BBS ansi adverts"
	case Ftp:
		return "FTP sites ansi adverts"
	case Logo:
		return "ansi format logos"
	case Nfo:
		return "ansi format nfo texts"
	case Pack:
		return "filepacks of ansi files"
	default:
		return "ansi format textfiles"
	}
}

func image(section Tag) string {
	switch section {
	case BBS:
		return "BBS advert images"
	case ForSale:
		return "image advertisements"
	case Pack:
		return "filepacks of images"
	case Proof:
		return "proof of release photos"
	default:
		return "images, pictures and photos"
	}
}

func text(section Tag) string {
	switch section {
	case AtariST:
		return "textfiles for the Atari ST"
	case AppleII:
		return "textfiles for the Apple II"
	case BBS:
		return "BBS text adverts"
	case ForSale:
		return "textfile adverts"
	case Ftp:
		return "textfile adverts for FTP sites"
	case Mag:
		return "magazine textfiles"
	case Nfo:
		return "nfo textfiles"
	case Pack:
		return "filepacks of textfiles"
	case Restrict:
		return "textfiles with restricted content"
	default:
		return fmt.Sprintf("%s textfiles", Names()[section])
	}
}

func textAmiga(section Tag) string {
	switch section {
	case BBS:
		return "BBS Amiga text adverts"
	case ForSale:
		return "Amiga textfile adverts"
	case Ftp:
		return "Amiga text adverts for FTP sites"
	case Mag:
		return "Amiga magazine textfiles"
	case Nfo:
		return "Amiga nfo textfiles"
	case Restrict:
		return "Amiga textfiles with restricted content"
	default:
		return fmt.Sprintf("%s textfiles for the Amiga", Names()[section])
	}
}

func windows(section Tag) string {
	switch section {
	case Demo:
		return "demos on Windows"
	case Install:
		return "Windows installers"
	case Intro:
		return "Windows intros"
	case Job:
		return "\"CrackMe\" tests for Windows"
	case Pack:
		return "filepacks of Windows programs"
	default:
		return fmt.Sprintf("%s for Windows", Names()[section])
	}
}

func dos(section Tag) string {
	switch section {
	case BBS:
		return "BBS intro adverts"
	case Demo:
		return "demos on " + msDos
	case ForSale:
		return "advertisements on " + msDos
	case GameHack:
		return "trainers or hacks on " + msDos
	case Install:
		return msDos + " installers"
	case Intro:
		return "intros for " + msDos
	case Pack:
		return "filepacks of " + msDos + " programs"
	default:
		return fmt.Sprintf("%s for %s", Names()[section], msDos)
	}
}

func emptyPlatform(section Tag) string {
	switch section {
	case BBS:
		return "BBS adverts"
	case Bust:
		return "busted releasers, sites and sceners"
	case Drama:
		return "drama between releasers or individuals"
	case ForSale:
		return "adverts for releasers or individuals"
	case Ftp:
		return "FTP site adverts"
	case Job:
		return "job adverts or new roles"
	case GameHack:
		return "game trainers or hacks"
	case Guide:
		return "guides, tutorials and how-to's"
	case Mag:
		return "magazine issues or ads"
	case News:
		return "articles from mainstream news outlets"
	case NfoTool:
		return "nfo file editors or tools"
	case Restrict:
		return "insider or restricted content"
	case Tool:
		return "software tools by the scene"
	}
	return fmt.Sprintf("%ss", Names()[section])
}

// Sum the numbers of files with the tag.
type Sum map[Tag]int

// Sums stores the results of file count query for each tag.
var Sums = make(Sum, Windows+1)

// Tags contains data for all the tags used by the web application.
var Tags = T{}

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
func counter(z *zap.SugaredLogger, t Tag) int64 {
	ctx := context.Background()
	db, err := postgres.ConnectDB()
	if err != nil {
		z.Errorf("Could not connect to the database: %s.", err)
		return -1
	}
	clause := "section = ?"
	if t >= FirstPlatform {
		clause = "platform = ?"
	}
	sum, err := models.Files(
		qm.Where(clause, URIs()[t])).Count(ctx, db)
	if err != nil {
		z.Errorf("Could not sum the records associated with tags: %s.", err)
		return -1
	}
	return sum
}

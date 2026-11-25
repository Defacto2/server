// Package fileslice provides functions that return model FileSlices, which are multiple artifact records.
//
//nolint:gochecknoglobals,nonamedreturns,wrapcheck
package fileslice

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/sqlboiler/v4/boil"
)

var (
	ErrCategory = errors.New("fileslice: unknown artifacts categories")
	ErrPagi     = errors.New("fileslice: pagination page and limit cannot be less than 1")
)

var uriMap = func() map[string]URI {
	m := make(map[string]URI)
	for val := range int(LastURI) {
		i := val + 1
		m[URI(i).String()] = URI(i)
	}
	return m
}()

// URI is a type for the files URI path.
type URI int

const LastURI = _maxURI - 1

const (
	root URI = iota
	advert
	animation
	announcement
	ansi
	ansiBBS
	ansiBrand
	ansiFTP
	ansiPack
	ansiNfo
	bbs
	bbstro
	bbsImage
	bbsText
	console
	database
	Deletions
	demoscene
	drama
	ForApproval
	ftp
	hack
	howTo
	htm
	java
	jobAdvert
	imageFile
	imagePack
	intro
	introMsdos
	introWindows
	installer
	linux
	magazine
	macos
	msdos
	msdosPack
	music
	Newest
	newsArticle
	NewUpdates
	NewUploads
	nfo
	nfoTool
	Oldest
	OneOffs
	OneOffsAZ
	pcb
	pcbPPE
	pcbText
	pdf
	proof
	restrict
	script
	Sensenstahl
	standards
	takedown
	text
	textAmiga
	textApple2
	textAtariST
	textPack
	tool
	trialCrackme
	Unwanted
	video
	windows
	windowsPack
	_maxURI // used to determine the last value
)

var uriStrings = [...]string{
	0:  "",
	1:  "advert",
	2:  "animation",
	3:  "announcement",
	4:  "ansi",
	5:  "ansi-bbs",
	6:  "ansi-brand",
	7:  "ansi-ftp",
	8:  "ansi-pack",
	9:  "ansi-nfo",
	10: "bbs",
	11: "bbstro",
	12: "bbs-image",
	13: "bbs-text",
	14: "console",
	15: "database",
	16: "deletions",
	17: "demoscene",
	18: "drama",
	19: "for-approval",
	20: "ftp",
	21: "hack",
	22: "how-to",
	23: "html",
	24: "java",
	25: "job-advert",
	26: "image",
	27: "image-pack",
	28: "intro",
	29: "intro-msdos",
	30: "intro-windows",
	31: "installer",
	32: "linux",
	33: "magazine",
	34: "macos",
	35: "msdos",
	36: "msdos-pack",
	37: "music",
	38: "newest",
	39: "news-article",
	40: "new-updates",
	41: "new-uploads",
	42: "nfo",
	43: "nfo-tool",
	44: "oldest",
	45: "one-offs",
	46: "one-offs-az",
	47: "pcboard",
	48: "pcboard-ppe",
	49: "pcboard-text",
	50: "pdf",
	51: "proof",
	52: "restrict",
	53: "script",
	54: "sensenstahl",
	55: "standards",
	56: "takedown",
	57: "text",
	58: "text-amiga",
	59: "text-apple2",
	60: "text-atari-st",
	61: "text-pack",
	62: "tool",
	63: "trial-crackme",
	64: "unwanted",
	65: "video",
	66: "windows",
	67: "windows-pack",
}

func (u URI) String() string {
	if uint(u) < uint(len(uriStrings)) {
		return uriStrings[u]
	}
	return ""
}

// Match path to a URI type or return -1 if not found.
func Match(path string) URI {
	if uri, ok := uriMap[path]; ok {
		return uri
	}
	return -1
}

// Valid returns true if path is a valid URI for the list of files.
func Valid(path string) bool {
	_, ok := uriMap[path]
	return ok
}

type fileMeta struct {
	logo  string
	h1sub string
	lead  string
}

var fileInfoMap = map[URI]fileMeta{
	NewUploads: {
		logo:  "new uploads",
		h1sub: "the new uploads",
		lead:  "These are the recent file artifacts that have been submitted to Defacto2.",
	},
	NewUpdates: {
		logo:  "new changes",
		h1sub: "the new changes",
		lead:  "These are the recent file artifacts that have been modified or submitted on Defacto2.",
	},
	ForApproval: {
		logo:  "new uploads",
		h1sub: "edit the new uploads for approval",
		lead:  "These are the recent file artifacts that have been submitted for approval on Defacto2.",
	},
	Deletions: {
		logo:  "deletions",
		h1sub: "edit the (hidden) deletions",
		lead:  "These are the file artifacts that have been removed from Defacto2.",
	},
	Unwanted: {
		logo:  "unwanted releases",
		h1sub: "edit the unwanted software releases",
		lead: "These are the file artifacts that have been marked " +
			"as potential unwanted software or containing viruses on Defacto2.",
	},
	Oldest: {
		logo:  "oldest releases",
		h1sub: "the oldest releases",
		lead:  "These are the earliest, historical file artifacts in the collection.",
	},
	OneOffs: {
		logo:  "one-off releases",
		h1sub: "the one-off releases",
		lead:  "Artifacts from groups, releasers, BBSes, and FTP sites with only a single documented file.",
	},
	OneOffsAZ: {
		logo:  "one-off releases, A-Z",
		h1sub: "the one-off releases, A–Z",
		lead: "Artifacts from groups, releasers, BBSes, and FTP sites with only a single documented file, " +
			"ordered alphabetically.",
	},
	Newest: {
		logo:  "newest releases",
		h1sub: "the newest releases",
		lead:  "These are the most recent file artifacts in the collection.",
	},
	Sensenstahl: {
		logo:  "sensenstahl 🎁",
		h1sub: "the bbstros for sensenstahl",
		lead:  "These are the newest BBStros added to the collection.",
	},
}

// FileInfo is a helper function for Files that returns the page title, h1 title and lead text.
func FileInfo(uri string) (logo, h1sub, lead string) {
	if meta, ok := fileInfoMap[Match(uri)]; ok {
		return meta.logo, meta.h1sub, meta.lead
	}

	s := RecordsSub(uri)
	return s, s, ""
}

// RecordsSub returns the records for the artifacts category URI.
func RecordsSub(uri string) string {
	if value, ok := recordsSubMap()[Match(uri)]; ok {
		return value
	}
	return "unknown uri"
}

var recordsSubMap = sync.OnceValue(func() map[URI]string {
	const none = tags.Tag(-1)
	return map[URI]string{
		advert:       none.Humanizes(tags.ForSale),
		animation:    "animations in an ansi format",
		announcement: none.Humanizes(tags.Announcement),
		ansi:         tags.ANSI.Humanizes(none),
		ansiBrand:    tags.ANSI.Humanizes(tags.Logo),
		ansiBBS:      tags.ANSI.Humanizes(tags.BBS),
		ansiFTP:      tags.ANSI.Humanizes(tags.Ftp),
		ansiNfo:      tags.ANSI.Humanizes(tags.Nfo),
		ansiPack:     tags.ANSI.Humanizes(tags.Pack),
		bbs:          none.Humanizes(tags.BBS),
		bbsImage:     tags.Image.Humanizes(tags.BBS),
		bbstro:       tags.DOS.Humanizes(tags.BBS),
		bbsText:      tags.Text.Humanizes(tags.BBS),
		console:      tags.Console.Humanizes(none),
		database:     none.Humanizes(tags.DataB),
		demoscene:    none.Humanizes(tags.Demo),
		drama:        none.Humanizes(tags.Drama),
		ftp:          none.Humanizes(tags.Ftp),
		hack:         none.Humanizes(tags.GameHack),
		htm:          "Websites and HTML documents",
		howTo:        none.Humanizes(tags.Guide),
		imageFile:    tags.Image.Humanizes(none),
		imagePack:    tags.Image.Humanizes(tags.Pack),
		installer:    none.Humanizes(tags.Install),
		intro:        none.Humanizes(tags.Intro),
		linux:        tags.Linux.Humanizes(none),
		java:         tags.Java.Humanizes(none),
		jobAdvert:    none.Humanizes(tags.Job),
		macos:        tags.Mac.Humanizes(none),
		msdosPack:    tags.DOS.Humanizes(tags.Pack),
		music:        tags.Audio.Humanizes(none),
		newsArticle:  none.Humanizes(tags.News),
		nfo:          none.Humanizes(tags.Nfo),
		nfoTool:      none.Humanizes(tags.NfoTool),
		standards:    none.Humanizes(tags.Rule),
		script:       tags.PHP.Humanizes(none),
		introMsdos:   tags.DOS.Humanizes(tags.Intro),
		introWindows: tags.Windows.Humanizes(tags.Intro),
		magazine:     none.Humanizes(tags.Mag),
		msdos:        tags.DOS.Humanizes(none),
		pcb:          tags.PCB.Humanizes(none),
		pcbPPE:       tags.PCB.Humanizes(tags.Tool),
		pcbText:      tags.PCB.Humanizes(tags.BBS),
		pdf:          tags.PDF.Humanizes(none),
		proof:        none.Humanizes(tags.Proof),
		restrict:     none.Humanizes(tags.Restrict),
		takedown:     none.Humanizes(tags.Bust),
		text:         tags.Text.Humanizes(none),
		textAmiga:    tags.TextAmiga.Humanizes(none),
		textApple2:   tags.Text.Humanizes(tags.AppleII),
		textAtariST:  tags.Text.Humanizes(tags.AtariST),
		textPack:     tags.Text.Humanizes(tags.Pack),
		tool:         none.Humanizes(tags.Tool),
		trialCrackme: tags.Windows.Humanizes(tags.Job),
		video:        tags.Video.Humanizes(none),
		windows:      tags.Windows.Humanizes(none),
		windowsPack:  tags.Windows.Humanizes(tags.Pack),
	}
})

type queryFunc func(context.Context, boil.ContextExecutor, int, int) (models.FileSlice, error)

var (
	a model.Artifacts

	// recordDispatch dispatch map for models.
	recordDispatch = map[URI]queryFunc{
		ForApproval:  model.OnlyApproval,
		Deletions:    model.OnlyHidden,
		Unwanted:     a.OnlyUnwanted,
		NewUploads:   a.ByKey,
		NewUpdates:   a.ByUpdated,
		Oldest:       a.ByOldest,
		Newest:       a.ByNewest,
		OneOffs:      list[model.OneOffs],
		OneOffsAZ:    list[model.OneOffsAZ],
		advert:       list[model.Advert],
		animation:    list[model.Animation],
		announcement: list[model.Announcement],
		ansi:         list[model.Ansi],
		ansiBrand:    list[model.AnsiBrand],
		ansiBBS:      list[model.AnsiBBS],
		ansiFTP:      list[model.AnsiFTP],
		ansiNfo:      list[model.AnsiNfo],
		ansiPack:     list[model.AnsiPack],
		bbs:          list[model.BBS],
		bbsImage:     list[model.BBSImage],
		bbstro:       list[model.BBStro],
		bbsText:      list[model.BBSText],
		console:      list[model.Console],
		database:     list[model.Database],
		demoscene:    list[model.Demoscene],
		drama:        list[model.Drama],
		ftp:          list[model.FTP],
		hack:         list[model.Hack],
		htm:          list[model.HTML],
		howTo:        list[model.HowTo],
		imageFile:    list[model.Image],
		imagePack:    list[model.ImagePack],
		installer:    list[model.Installer],
		intro:        list[model.Intro],
		linux:        list[model.Linux],
		java:         list[model.Java],
		jobAdvert:    list[model.JobAdvert],
		macos:        list[model.Macos],
		msdosPack:    list[model.MsDosPack],
		music:        list[model.Music],
		newsArticle:  list[model.NewsArticle],
		nfo:          list[model.Nfo],
		nfoTool:      list[model.NfoTool],
		standards:    list[model.Standards],
		script:       list[model.Script],
		introMsdos:   list[model.IntroMsDos],
		introWindows: list[model.IntroWindows],
		magazine:     list[model.Magazine],
		msdos:        list[model.MsDos],
		pcb:          list[model.PCBoard],
		pcbPPE:       list[model.PCBoardPPE],
		pcbText:      list[model.PCBoardText],
		pdf:          list[model.PDF],
		proof:        list[model.Proof],
		restrict:     list[model.Restrict],
		takedown:     list[model.Takedown],
		text:         list[model.Text],
		textAmiga:    list[model.TextAmiga],
		textApple2:   list[model.TextApple2],
		textAtariST:  list[model.TextAtariST],
		textPack:     list[model.TextPack],
		tool:         list[model.Tool],
		trialCrackme: list[model.TrialCrackme],
		video:        list[model.Video],
		windows:      list[model.Windows],
		windowsPack:  list[model.WindowsPack],
	}
)

func list[T any, PT interface {
	*T
	List(ctx context.Context, exec boil.ContextExecutor, page, limit int) (models.FileSlice, error)
}](ctx context.Context, exec boil.ContextExecutor, page, limit int) (models.FileSlice, error) {
	return PT(new(T)).List(ctx, exec, page, limit)
}

// Records returns the records for the artifacts category URI.
// Note that the record statistics and counts get cached.
func Records(ctx context.Context, exec boil.ContextExecutor, uri string, page, limit int) (models.FileSlice, error) {
	const format = "records %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return nil, fmt.Errorf(format, "fileslice check", err)
	}

	if page < 1 {
		return nil, fmt.Errorf(format, "page "+strconv.Itoa(page), ErrPagi)
	}
	if limit < 1 {
		return nil, fmt.Errorf(format, "page "+strconv.Itoa(limit), ErrPagi)
	}

	if fn, ok := recordDispatch[Match(uri)]; ok {
		return fn(ctx, exec, page, limit)
	}

	if Match(uri) == Sensenstahl {
		var r model.BBStro
		return r.Sensenstahl(ctx, exec, page, limit)
	}

	return nil, fmt.Errorf(format, "categories "+uri, ErrCategory)
}

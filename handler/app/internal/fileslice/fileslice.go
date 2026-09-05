// Package fileslice provides functions that return model FileSlices, which are multiple artifact records.
//
//nolint:exhaustive,gochecknoglobals,wrapcheck
package fileslice

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"

	"github.com/Defacto2/server/internal/nils"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/sqlboiler/v4/boil"
	"golang.org/x/sync/errgroup"
)

var ErrCategory = errors.New("unknown artifacts categories")

var uriMap = func() map[string]URI {
	m := make(map[string]URI)
	for val := range int(WindowsPack) {
		i := val + 1
		m[URI(i).String()] = URI(i)
	}
	return m
}()

// URI is a type for the files URI path.
type URI int

const (
	root URI = iota
	advert
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
	WindowsPack // last value needs to be a global to allow testing
)

var uriStrings = [...]string{
	0:  "",
	1:  "advert",
	2:  "announcement",
	3:  "ansi",
	4:  "ansi-bbs",
	5:  "ansi-brand",
	6:  "ansi-ftp",
	7:  "ansi-pack",
	8:  "ansi-nfo",
	9:  "bbs",
	10: "bbstro",
	11: "bbs-image",
	12: "bbs-text",
	13: "console",
	14: "database",
	15: "deletions",
	16: "demoscene",
	17: "drama",
	18: "for-approval",
	19: "ftp",
	20: "hack",
	21: "how-to",
	22: "html",
	23: "java",
	24: "job-advert",
	25: "image",
	26: "image-pack",
	27: "intro",
	28: "intro-msdos",
	29: "intro-windows",
	30: "installer",
	31: "linux",
	32: "magazine",
	33: "macos",
	34: "msdos",
	35: "msdos-pack",
	36: "music",
	37: "newest",
	38: "news-article",
	39: "new-updates",
	40: "new-uploads",
	41: "nfo",
	42: "nfo-tool",
	43: "oldest",
	44: "pcboard",
	45: "pcboard-ppe",
	46: "pcboard-text",
	47: "pdf",
	48: "proof",
	49: "restrict",
	50: "script",
	51: "sensenstahl",
	52: "standards",
	53: "takedown",
	54: "text",
	55: "text-amiga",
	56: "text-apple2",
	57: "text-atari-st",
	58: "text-pack",
	59: "tool",
	60: "trial-crackme",
	61: "unwanted",
	62: "video",
	63: "windows",
	64: "windows-pack",
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
func FileInfo(uri string) (string, string, string) {
	if meta, ok := fileInfoMap[Match(uri)]; ok {
		return meta.logo, meta.h1sub, meta.lead
	}

	s := RecordsSub(uri)
	return s, s, ""
}

// RecordsSub returns the records for the artifacts category URI.
func RecordsSub(uri string) string {
	const none = tags.Tag(-1)
	subs := map[URI]string{
		advert:       none.Humanizes(tags.ForSale),
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
		htm:          uri,
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
		WindowsPack:  tags.Windows.Humanizes(tags.Pack),
	}
	if value, found := subs[Match(uri)]; found {
		return value
	}

	return "unknown uri"
}

type queryFunc func(context.Context, boil.ContextExecutor, int, int) (models.FileSlice, error)

var (
	a model.Artifacts

	// recordDispatch dispatch map for models.
	recordDispatch = map[URI]queryFunc{
		ForApproval: model.OnlyApproval,
		Deletions:   model.OnlyHidden,
		Unwanted:    a.OnlyUnwanted,
		NewUploads:  a.ByKey,
		NewUpdates:  a.ByUpdated,
		Oldest:      a.ByOldest,
		Newest:      a.ByNewest,
	}
)

// Records returns the records for the artifacts category URI.
// Note that the record statistics and counts get cached.
func Records(ctx context.Context, exec boil.ContextExecutor, uri string, page, limit int) (models.FileSlice, error) {
	if err := nils.Check(ctx, exec); err != nil {
		return nil, fmt.Errorf("fileslice records check: %w", err)
	}

	if fn, ok := recordDispatch[Match(uri)]; ok {
		return fn(ctx, exec, page, limit)
	}

	return records00(ctx, exec, uri, page, limit)
}

func records00(ctx context.Context, exec boil.ContextExecutor, uri string, page, limit int) (models.FileSlice, error) {
	switch Match(uri) {
	case advert:
		var r model.Advert
		return r.List(ctx, exec, page, limit)
	case announcement:
		var r model.Announcement
		return r.List(ctx, exec, page, limit)
	case ansi:
		var r model.Ansi
		return r.List(ctx, exec, page, limit)
	case ansiBrand:
		var r model.AnsiBrand
		return r.List(ctx, exec, page, limit)
	case ansiBBS:
		var r model.AnsiBBS
		return r.List(ctx, exec, page, limit)
	case ansiFTP:
		var r model.AnsiFTP
		return r.List(ctx, exec, page, limit)
	case ansiNfo:
		var r model.AnsiNfo
		return r.List(ctx, exec, page, limit)
	case ansiPack:
		var r model.AnsiPack
		return r.List(ctx, exec, page, limit)
	case bbs:
		var r model.BBS
		return r.List(ctx, exec, page, limit)
	case bbsImage:
		var r model.BBSImage
		return r.List(ctx, exec, page, limit)
	case bbstro:
		var r model.BBStro
		return r.List(ctx, exec, page, limit)
	case bbsText:
		var r model.BBSText
		return r.List(ctx, exec, page, limit)
	}
	return records11(ctx, exec, uri, page, limit)
}

func records11(ctx context.Context, exec boil.ContextExecutor, uri string, page, limit int) (models.FileSlice, error) {
	switch Match(uri) {
	case database:
		var r model.Database
		return r.List(ctx, exec, page, limit)
	case demoscene:
		var r model.Demoscene
		return r.List(ctx, exec, page, limit)
	case drama:
		var r model.Drama
		return r.List(ctx, exec, page, limit)
	case ftp:
		var r model.FTP
		return r.List(ctx, exec, page, limit)
	case hack:
		var r model.Hack
		return r.List(ctx, exec, page, limit)
	case htm:
		var r model.HTML
		return r.List(ctx, exec, page, limit)
	case howTo:
		var r model.HowTo
		return r.List(ctx, exec, page, limit)
	case imageFile:
		var r model.Image
		return r.List(ctx, exec, page, limit)
	case imagePack:
		var r model.ImagePack
		return r.List(ctx, exec, page, limit)
	case installer:
		var r model.Installer
		return r.List(ctx, exec, page, limit)
	case intro:
		var r model.Intro
		return r.List(ctx, exec, page, limit)
	case linux:
		var r model.Linux
		return r.List(ctx, exec, page, limit)
	case java:
		var r model.Java
		return r.List(ctx, exec, page, limit)
	case jobAdvert:
		var r model.JobAdvert
		return r.List(ctx, exec, page, limit)
	}

	return records22(ctx, exec, uri, page, limit)
}

func records22(ctx context.Context, exec boil.ContextExecutor, uri string, page, limit int) (models.FileSlice, error) {
	switch Match(uri) {
	case macos:
		var r model.Macos
		return r.List(ctx, exec, page, limit)
	case msdosPack:
		var r model.MsDosPack
		return r.List(ctx, exec, page, limit)
	case music:
		var r model.Music
		return r.List(ctx, exec, page, limit)
	case newsArticle:
		var r model.NewsArticle
		return r.List(ctx, exec, page, limit)
	case nfo:
		var r model.Nfo
		return r.List(ctx, exec, page, limit)
	case nfoTool:
		var r model.NfoTool
		return r.List(ctx, exec, page, limit)
	case standards:
		var r model.Standards
		return r.List(ctx, exec, page, limit)
	case script:
		var r model.Script
		return r.List(ctx, exec, page, limit)
	case introMsdos:
		var r model.IntroMsDos
		return r.List(ctx, exec, page, limit)
	case introWindows:
		var r model.IntroWindows
		return r.List(ctx, exec, page, limit)
	case magazine:
		var r model.Magazine
		return r.List(ctx, exec, page, limit)
	case msdos:
		var r model.MsDos
		return r.List(ctx, exec, page, limit)
	case pcb:
		var r model.PCBoard
		return r.List(ctx, exec, page, limit)
	case pcbPPE:
		var r model.PCBoardPPE
		return r.List(ctx, exec, page, limit)
	case pcbText:
		var r model.PCBoardText
		return r.List(ctx, exec, page, limit)
	case pdf:
		var r model.PDF
		return r.List(ctx, exec, page, limit)
	}

	return records33(ctx, exec, uri, page, limit)
}

func records33(ctx context.Context, exec boil.ContextExecutor, uri string, page, limit int) (models.FileSlice, error) {
	switch Match(uri) {
	case proof:
		var r model.Proof
		return r.List(ctx, exec, page, limit)
	case restrict:
		var r model.Restrict
		return r.List(ctx, exec, page, limit)
	case takedown:
		var r model.Takedown
		return r.List(ctx, exec, page, limit)
	case text:
		var r model.Text
		return r.List(ctx, exec, page, limit)
	case textAmiga:
		var r model.TextAmiga
		return r.List(ctx, exec, page, limit)
	case textApple2:
		var r model.TextApple2
		return r.List(ctx, exec, page, limit)
	case textAtariST:
		var r model.TextAtariST
		return r.List(ctx, exec, page, limit)
	case textPack:
		var r model.TextPack
		return r.List(ctx, exec, page, limit)
	case tool:
		var r model.Tool
		return r.List(ctx, exec, page, limit)
	case trialCrackme:
		var r model.TrialCrackme
		return r.List(ctx, exec, page, limit)
	case video:
		var r model.Video
		return r.List(ctx, exec, page, limit)
	case windows:
		var r model.Windows
		return r.List(ctx, exec, page, limit)
	case WindowsPack:
		var r model.WindowsPack
		return r.List(ctx, exec, page, limit)
	case Sensenstahl:
		var r model.BBStro
		return r.Sensenstahl(ctx, exec, page, limit)
	case console:
		var r model.Console
		return r.List(ctx, exec, page, limit)
	default:
		const format = "artifacts category %s: %w"
		return nil, fmt.Errorf(format, uri, ErrCategory)
	}
}

// Counter returns the statistics for the artifacts categories.
func Counter(ctx context.Context, db *sql.DB) (Stats, error) {
	const format = "artifacts categories counter %s: %w"

	if err := nils.Check(ctx, db); err != nil {
		return Stats{}, fmt.Errorf(format, "check", err)
	}

	counter := newStats()
	if err := counter.Get(ctx, db); err != nil {
		return Stats{}, fmt.Errorf(format, "get", err)
	}

	return counter, nil
}

// Stats are the database statistics for the artifacts categories.
type Stats struct {
	Record    model.Artifacts
	Ansi      model.Ansi
	AnsiBBS   model.AnsiBBS
	BBS       model.BBS
	BBSText   model.BBSText
	BBStro    model.BBStro
	Console   model.Console
	Demoscene model.Demoscene
	MsDos     model.MsDos
	Intro     model.Intro
	IntroD    model.IntroMsDos
	IntroW    model.IntroWindows
	Installer model.Installer
	Java      model.Java
	Linux     model.Linux
	Magazine  model.Magazine
	Macos     model.Macos
	Nfo       model.Nfo
	NfoTool   model.NfoTool
	Proof     model.Proof
	Script    model.Script
	Text      model.Text
	Windows   model.Windows
}

// Era is used to provide sorted statistics for artifact categories.
type Era struct {
	Name    string
	MinYear int
	MaxYear int
}

// SortYear returns the Eras sorted from oldest to newest.
// If multiple Era share the same year, then the shortest to longest time span is used.
//
// For example:
//   - 1. 1980 - 1982
//   - 2. 1980 - 1990
//   - 3. 1981 - 1985
func (s *Stats) SortYear() [20]Era {
	items := s.eras()
	sort.Slice(items[:], func(i, j int) bool {
		if items[i].MinYear == items[j].MinYear {
			return items[i].MaxYear < items[j].MaxYear
		}
		return items[i].MinYear < items[j].MinYear
	})

	return items
}

// Item is used to provided sorted statistics for artifact categories.
type Item struct {
	Name  string
	Bytes int
	Count int
}

// SortByte returns the Items sorted from highest to lowest bytes.
func (s *Stats) SortByte() [21]Item {
	items := s.items()
	sort.Slice(items[:], func(i, j int) bool {
		if items[i].Bytes == items[j].Bytes {
			return items[i].Count > items[j].Count
		}
		return items[i].Bytes > items[j].Bytes
	})

	return items
}

// SortCount returns the Items sorted by highest to lowest counts.
func (s *Stats) SortCount() [21]Item {
	items := s.items()
	sort.Slice(items[:], func(i, j int) bool {
		if items[i].Count == items[j].Count {
			return items[i].Bytes > items[j].Bytes
		}
		return items[i].Count > items[j].Count
	})

	return items
}

// SortName returns the Items sorted alphabetically by their name.
func (s *Stats) SortName() [21]Item {
	items := s.items()
	sort.Slice(items[:], func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
	return items
}

// Statistics returns the empty database statistics for the artifacts categories.
func Statistics() Stats {
	return newStats()
}

// Get and store the database statistics for the artifacts categories.
func (s *Stats) Get(ctx context.Context, exec boil.ContextExecutor) error {
	const format = "category get stats %s: %w"
	if err := nils.Check(ctx, exec); err != nil {
		return fmt.Errorf(format, "check", err)
	}

	g, ctx := errgroup.WithContext(ctx)

	// fetch record
	g.Go(func() error {
		if err := s.Record.Public(ctx, exec); err != nil {
			return fmt.Errorf(format, "record", err)
		}
		return nil
	})

	// concurrent, individual category stats
	stats := []struct {
		name string
		fn   func(context.Context, boil.ContextExecutor) error
	}{
		{"ansi", s.Ansi.Stat},
		{"ansi bbs", s.AnsiBBS.Stat},
		{"bbs", s.BBS.Stat},
		{"bbs text", s.BBSText.Stat},
		{"bbstro", s.BBStro.Stat},
		{"console", s.Console.Stat},
		{"ms-dos", s.MsDos.Stat},
		{"intro", s.Intro.Stat},
		{"intro ms-dos", s.IntroD.Stat},
		{"intro windows", s.IntroW.Stat},
		{"installer", s.Installer.Stat},
		{"java", s.Java.Stat},
		{"linux", s.Linux.Stat},
		{"demoscene", s.Demoscene.Stat},
		{"macos", s.Macos.Stat},
		{"magazine", s.Magazine.Stat},
		{"nfo", s.Nfo.Stat},
		{"nfo tool", s.NfoTool.Stat},
		{"proof", s.Proof.Stat},
		{"script", s.Script.Stat},
		{"text", s.Text.Stat},
		{"windows", s.Windows.Stat},
	}

	for _, st := range stats {
		g.Go(func() error {
			if err := st.fn(ctx, exec); err != nil {
				return fmt.Errorf(format, st.name, err)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf(format, "wait", err)
	}

	return nil
}

func (s *Stats) eras() [20]Era {
	return [...]Era{
		{s.Ansi.String(), s.Ansi.MinYear, s.Ansi.MaxYear},
		{s.AnsiBBS.String(), s.AnsiBBS.MinYear, s.AnsiBBS.MaxYear},
		{s.BBS.String(), s.BBS.MinYear, s.BBS.MaxYear},
		{s.BBSText.String(), s.BBSText.MinYear, s.BBSText.MaxYear},
		{s.BBStro.String(), s.BBStro.MinYear, s.BBStro.MaxYear},
		//	{s.Console.String(), s.Console.MinYear, s.Console.MaxYear},
		{s.Demoscene.String(), s.Demoscene.MinYear, s.Demoscene.MaxYear},
		{s.MsDos.String(), s.MsDos.MinYear, s.MsDos.MaxYear},
		{s.Intro.String(), s.Intro.MinYear, s.Intro.MaxYear},
		{s.IntroD.String(), s.IntroD.MinYear, s.IntroD.MaxYear},
		{s.IntroW.String(), s.IntroW.MinYear, s.IntroW.MaxYear},
		{s.Installer.String(), s.Installer.MinYear, s.Installer.MaxYear},
		{s.Java.String(), s.Java.MinYear, s.Java.MaxYear},
		{s.Linux.String(), s.Linux.MinYear, s.Linux.MaxYear},
		{s.Magazine.String(), s.Magazine.MinYear, s.Magazine.MaxYear},
		{s.Macos.String(), s.Macos.MinYear, s.Macos.MaxYear},
		{s.Nfo.String(), s.Nfo.MinYear, s.Nfo.MaxYear},
		{s.NfoTool.String(), s.NfoTool.MinYear, s.NfoTool.MaxYear},
		{s.Proof.String(), s.Proof.MinYear, s.Proof.MaxYear},
		{s.Script.String(), s.Script.MinYear, s.Script.MaxYear},
		{s.Windows.String(), s.Windows.MinYear, s.Windows.MaxYear},
	}
}

func (s *Stats) items() [21]Item {
	return [...]Item{
		{s.Ansi.String(), s.Ansi.Bytes, s.Ansi.Count},
		{s.AnsiBBS.String(), s.AnsiBBS.Bytes, s.AnsiBBS.Count},
		{s.BBS.String(), s.BBS.Bytes, s.BBS.Count},
		{s.BBSText.String(), s.BBSText.Bytes, s.BBSText.Count},
		{s.BBStro.String(), s.BBStro.Bytes, s.BBStro.Count},
		{s.Console.String(), s.Console.Bytes, s.Console.Count},
		{s.Demoscene.String(), s.Demoscene.Bytes, s.Demoscene.Count},
		{s.MsDos.String(), s.MsDos.Bytes, s.MsDos.Count},
		{s.Intro.String(), s.Intro.Bytes, s.Intro.Count},
		{s.IntroD.String(), s.IntroD.Bytes, s.IntroD.Count},
		{s.IntroW.String(), s.IntroW.Bytes, s.IntroW.Count},
		{s.Installer.String(), s.Installer.Bytes, s.Installer.Count},
		{s.Java.String(), s.Java.Bytes, s.Java.Count},
		{s.Linux.String(), s.Linux.Bytes, s.Linux.Count},
		{s.Magazine.String(), s.Magazine.Bytes, s.Magazine.Count},
		{s.Macos.String(), s.Macos.Bytes, s.Macos.Count},
		{s.Nfo.String(), s.Nfo.Bytes, s.Nfo.Count},
		{s.NfoTool.String(), s.NfoTool.Bytes, s.NfoTool.Count},
		{s.Proof.String(), s.Proof.Bytes, s.Proof.Count},
		{s.Script.String(), s.Script.Bytes, s.Script.Count},
		{s.Windows.String(), s.Windows.Bytes, s.Windows.Count},
	}
}

// newStats returns a new Stats struct initialized with zero values.
func newStats() Stats {
	return Stats{} //nolint:exhaustruct_v5
}

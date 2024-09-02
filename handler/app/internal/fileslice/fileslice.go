// Package fileslice provides functions that return model FileSlices, which are multiple artifact records.
//
//nolint:wrapcheck
package fileslice

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

var ErrCategory = errors.New("unknown artifacts categories")

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
	ansiTool
	bbs
	bbstro
	bbsImage
	bbsText
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
	newest
	newsArticle
	NewUpdates
	NewUploads
	nfo
	nfoTool
	oldest
	pdf
	proof
	restrict
	script
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
)

func (u URI) String() string { //nolint:funlen
	return [...]string{
		"",
		"advert",
		"announcement",
		"ansi",
		"ansi-bbs",
		"ansi-brand",
		"ansi-ftp",
		"ansi-pack",
		"ansi-nfo",
		"ansi-tool",
		"bbs",
		"bbstro",
		"bbs-image",
		"bbs-text",
		"database",
		"deletions",
		"demoscene",
		"drama",
		"for-approval",
		"ftp",
		"hack",
		"how-to",
		"html",
		"java",
		"job-advert",
		"image",
		"image-pack",
		"intro",
		"intro-msdos",
		"intro-windows",
		"installer",
		"linux",
		"magazine",
		"macos",
		"msdos",
		"msdos-pack",
		"music",
		"newest",
		"news-article",
		"new-updates",
		"new-uploads",
		"nfo",
		"nfo-tool",
		"oldest",
		"pdf",
		"proof",
		"restrict",
		"script",
		"standards",
		"takedown",
		"text",
		"text-amiga",
		"text-apple2",
		"text-atari-st",
		"text-pack",
		"tool",
		"trial-crackme",
		"unwanted",
		"video",
		"windows",
		"windows-pack",
	}[u]
}

// Match path to a URI type or return -1 if not found.
func Match(path string) URI {
	for val := range int(windowsPack) {
		i := val + 1
		if URI(i).String() == path {
			return URI(i)
		}
	}
	return -1
}

// Valid returns true if path is a valid URI for the list of files.
func Valid(path string) bool {
	for val := range int(windowsPack) {
		i := val + 1
		if URI(i).String() == path {
			return true
		}
	}
	return false
}

// FileInfo is a helper function for Files that returns the page title, h1 title and lead text.
func FileInfo(uri string) (string, string, string) {
	var logo, h1sub, lead string
	switch Match(uri) {
	case NewUploads:
		logo = "new uploads"
		h1sub = "the new uploads"
		lead = "These are the recent file artifacts that have been submitted to Defacto2."
	case NewUpdates:
		logo = "new changes"
		h1sub = "the new changes"
		lead = "These are the recent file artifacts that have been modified or submitted on Defacto2."
	case ForApproval:
		logo = "new uploads"
		h1sub = "edit the new uploads"
		lead = "These are the recent file artifacts that have been submitted for approval on Defacto2."
	case Deletions:
		logo = "deletions"
		h1sub = "edit the (hidden) deletions"
		lead = "These are the file artifacts that have been removed from Defacto2."
	case Unwanted:
		logo = "unwanted releases"
		h1sub = "edit the unwanted software releases"
		lead = "These are the file artifacts that have been marked as potential unwanted software " +
			"or containing viruses on Defacto2."
	case oldest:
		logo = "oldest releases"
		h1sub = "the oldest releases"
		lead = "These are the earliest, historical file artifacts in the collection."
	case newest:
		logo = "newest releases"
		h1sub = "the newest releases"
		lead = "These are the most recent file artifacts in the collection."
	default:
		s := RecordsSub(uri)
		h1sub = s
		logo = s
	}
	return logo, h1sub, lead
}

// RecordsSub returns the records for the artifacts category URI.
func RecordsSub(uri string) string {
	const ignore = -1
	subs := map[URI]string{
		advert:       tags.Humanizes(ignore, tags.ForSale),
		announcement: tags.Humanizes(ignore, tags.Announcement),
		ansi:         tags.Humanizes(tags.ANSI, ignore),
		ansiBrand:    tags.Humanizes(tags.ANSI, tags.Logo),
		ansiBBS:      tags.Humanizes(tags.ANSI, tags.BBS),
		ansiFTP:      tags.Humanizes(tags.ANSI, tags.Ftp),
		ansiNfo:      tags.Humanizes(tags.ANSI, tags.Nfo),
		ansiPack:     tags.Humanizes(tags.ANSI, tags.Pack),
		bbs:          tags.Humanizes(ignore, tags.BBS),
		bbsImage:     tags.Humanizes(tags.Image, tags.BBS),
		bbstro:       tags.Humanizes(tags.DOS, tags.BBS),
		bbsText:      tags.Humanizes(tags.Text, tags.BBS),
		database:     tags.Humanizes(ignore, tags.DataB),
		demoscene:    tags.Humanizes(ignore, tags.Demo),
		drama:        tags.Humanizes(ignore, tags.Drama),
		ftp:          tags.Humanizes(ignore, tags.Ftp),
		hack:         tags.Humanizes(ignore, tags.GameHack),
		htm:          uri,
		howTo:        tags.Humanizes(ignore, tags.Guide),
		imageFile:    tags.Humanizes(tags.Image, ignore),
		imagePack:    tags.Humanizes(tags.Image, tags.Pack),
		installer:    tags.Humanizes(ignore, tags.Install),
		intro:        tags.Humanizes(ignore, tags.Intro),
		linux:        tags.Humanizes(tags.Linux, ignore),
		java:         tags.Humanizes(tags.Java, ignore),
		jobAdvert:    tags.Humanizes(ignore, tags.Job),
		macos:        tags.Humanizes(tags.Mac, ignore),
		msdosPack:    tags.Humanizes(tags.DOS, tags.Pack),
		music:        tags.Humanizes(tags.Audio, ignore),
		newsArticle:  tags.Humanizes(ignore, tags.News),
		nfo:          tags.Humanizes(ignore, tags.Nfo),
		nfoTool:      tags.Humanizes(ignore, tags.NfoTool),
		standards:    tags.Humanizes(ignore, tags.Rule),
		script:       tags.Humanizes(tags.PHP, ignore),
		introMsdos:   tags.Humanizes(tags.DOS, tags.Intro),
		introWindows: tags.Humanizes(tags.Windows, tags.Intro),
		magazine:     tags.Humanizes(ignore, tags.Mag),
		msdos:        tags.Humanizes(tags.DOS, ignore),
		pdf:          tags.Humanizes(tags.PDF, ignore),
		proof:        tags.Humanizes(ignore, tags.Proof),
		restrict:     tags.Humanizes(ignore, tags.Restrict),
		takedown:     tags.Humanizes(ignore, tags.Bust),
		text:         tags.Humanizes(tags.Text, ignore),
		textAmiga:    tags.Humanizes(tags.TextAmiga, ignore),
		textApple2:   tags.Humanizes(tags.Text, tags.AppleII),
		textAtariST:  tags.Humanizes(tags.Text, tags.AtariST),
		textPack:     tags.Humanizes(tags.Text, tags.Pack),
		tool:         tags.Humanizes(ignore, tags.Tool),
		trialCrackme: tags.Humanizes(tags.Windows, tags.Job),
		video:        tags.Humanizes(tags.Video, ignore),
		windows:      tags.Humanizes(tags.Windows, ignore),
		windowsPack:  tags.Humanizes(tags.Windows, tags.Pack),
	}
	if value, found := subs[Match(uri)]; found {
		return value
	}
	return "unknown uri"
}

// Records returns the records for the artifacts category URI.
// Note that the record statistics and counts get cached.
func Records(ctx context.Context, exec boil.ContextExecutor, uri string, page, limit int) (models.FileSlice, error) {
	switch Match(uri) {
	// pulldown editor menu matches
	case ForApproval:
		r := model.Artifacts{}
		return r.ByForApproval(ctx, exec, page, limit)
	case Deletions:
		r := model.Artifacts{}
		return r.ByHidden(ctx, exec, page, limit)
	case Unwanted:
		r := model.Artifacts{}
		return r.ByUnwanted(ctx, exec, page, limit)
	// pulldown menu matches
	case NewUploads:
		r := model.Artifacts{}
		return r.ByKey(ctx, exec, page, limit)
	case NewUpdates:
		r := model.Artifacts{}
		return r.ByUpdated(ctx, exec, page, limit)
	case oldest:
		r := model.Artifacts{}
		return r.ByOldest(ctx, exec, page, limit)
	case newest:
		r := model.Artifacts{}
		return r.ByNewest(ctx, exec, page, limit)
	}
	return recordsZ(ctx, exec, uri, page, limit)
}

func recordsZ(ctx context.Context, exec boil.ContextExecutor, uri string, page, limit int) (models.FileSlice, error) {
	switch Match(uri) {
	case advert:
		r := model.Advert{}
		return r.List(ctx, exec, page, limit)
	case announcement:
		r := model.Announcement{}
		return r.List(ctx, exec, page, limit)
	case ansi:
		r := model.Ansi{}
		return r.List(ctx, exec, page, limit)
	case ansiBrand:
		r := model.AnsiBrand{}
		return r.List(ctx, exec, page, limit)
	case ansiBBS:
		r := model.AnsiBBS{}
		return r.List(ctx, exec, page, limit)
	case ansiFTP:
		r := model.AnsiFTP{}
		return r.List(ctx, exec, page, limit)
	case ansiNfo:
		r := model.AnsiNfo{}
		return r.List(ctx, exec, page, limit)
	case ansiPack:
		r := model.AnsiPack{}
		return r.List(ctx, exec, page, limit)
	case bbs:
		r := model.BBS{}
		return r.List(ctx, exec, page, limit)
	case bbsImage:
		r := model.BBSImage{}
		return r.List(ctx, exec, page, limit)
	case bbstro:
		r := model.BBStro{}
		return r.List(ctx, exec, page, limit)
	case bbsText:
		r := model.BBSText{}
		return r.List(ctx, exec, page, limit)
	}
	return records0(ctx, exec, uri, page, limit)
}

func records0(ctx context.Context, exec boil.ContextExecutor, uri string, page, limit int) (models.FileSlice, error) {
	switch Match(uri) {
	case database:
		r := model.Database{}
		return r.List(ctx, exec, page, limit)
	case demoscene:
		r := model.Demoscene{}
		return r.List(ctx, exec, page, limit)
	case drama:
		r := model.Drama{}
		return r.List(ctx, exec, page, limit)
	case ftp:
		r := model.FTP{}
		return r.List(ctx, exec, page, limit)
	case hack:
		r := model.Hack{}
		return r.List(ctx, exec, page, limit)
	case htm:
		r := model.HTML{}
		return r.List(ctx, exec, page, limit)
	case howTo:
		r := model.HowTo{}
		return r.List(ctx, exec, page, limit)
	case imageFile:
		r := model.Image{}
		return r.List(ctx, exec, page, limit)
	case imagePack:
		r := model.ImagePack{}
		return r.List(ctx, exec, page, limit)
	case installer:
		r := model.Installer{}
		return r.List(ctx, exec, page, limit)
	case intro:
		r := model.Intro{}
		return r.List(ctx, exec, page, limit)
	case linux:
		r := model.Linux{}
		return r.List(ctx, exec, page, limit)
	case java:
		r := model.Java{}
		return r.List(ctx, exec, page, limit)
	case jobAdvert:
		r := model.JobAdvert{}
		return r.List(ctx, exec, page, limit)
	}
	return records1(ctx, exec, uri, page, limit)
}

func records1(ctx context.Context, exec boil.ContextExecutor, uri string, page, limit int) (models.FileSlice, error) {
	switch Match(uri) {
	case macos:
		r := model.Macos{}
		return r.List(ctx, exec, page, limit)
	case msdosPack:
		r := model.MsDosPack{}
		return r.List(ctx, exec, page, limit)
	case music:
		r := model.Music{}
		return r.List(ctx, exec, page, limit)
	case newsArticle:
		r := model.NewsArticle{}
		return r.List(ctx, exec, page, limit)
	case nfo:
		r := model.Nfo{}
		return r.List(ctx, exec, page, limit)
	case nfoTool:
		r := model.NfoTool{}
		return r.List(ctx, exec, page, limit)
	case standards:
		r := model.Standard{}
		return r.List(ctx, exec, page, limit)
	case script:
		r := model.Script{}
		return r.List(ctx, exec, page, limit)
	case introMsdos:
		r := model.IntroMsDos{}
		return r.List(ctx, exec, page, limit)
	case introWindows:
		r := model.IntroWindows{}
		return r.List(ctx, exec, page, limit)
	case magazine:
		r := model.Magazine{}
		return r.List(ctx, exec, page, limit)
	case msdos:
		r := model.MsDos{}
		return r.List(ctx, exec, page, limit)
	case pdf:
		r := model.PDF{}
		return r.List(ctx, exec, page, limit)
	case proof:
		r := model.Proof{}
		return r.List(ctx, exec, page, limit)
	}
	return records2(ctx, exec, uri, page, limit)
}

func records2(ctx context.Context, exec boil.ContextExecutor, uri string, page, limit int) (models.FileSlice, error) {
	switch Match(uri) {
	case restrict:
		r := model.Restrict{}
		return r.List(ctx, exec, page, limit)
	case takedown:
		r := model.Takedown{}
		return r.List(ctx, exec, page, limit)
	case text:
		r := model.Text{}
		return r.List(ctx, exec, page, limit)
	case textAmiga:
		r := model.TextAmiga{}
		return r.List(ctx, exec, page, limit)
	case textApple2:
		r := model.TextApple2{}
		return r.List(ctx, exec, page, limit)
	case textAtariST:
		r := model.TextAtariST{}
		return r.List(ctx, exec, page, limit)
	case textPack:
		r := model.TextPack{}
		return r.List(ctx, exec, page, limit)
	case tool:
		r := model.Tool{}
		return r.List(ctx, exec, page, limit)
	case trialCrackme:
		r := model.TrialCrackme{}
		return r.List(ctx, exec, page, limit)
	case video:
		r := model.Video{}
		return r.List(ctx, exec, page, limit)
	case windows:
		r := model.Windows{}
		return r.List(ctx, exec, page, limit)
	case windowsPack:
		r := model.WindowsPack{}
		return r.List(ctx, exec, page, limit)
	default:
		return nil, fmt.Errorf("artifacts category %w: %s", ErrCategory, uri)
	}
}

// Counter returns the statistics for the artifacts categories.
func Counter(db *sql.DB) (Stats, error) {
	ctx := context.Background()
	counter := Stats{}
	if err := counter.Get(ctx, db); err != nil {
		return Stats{}, fmt.Errorf("cartifacts categories counter get %w", err)
	}
	return counter, nil
}

// Stats are the database statistics for the artifacts categories.
type Stats struct {
	IntroW    model.IntroWindows
	Record    model.Artifacts
	Ansi      model.Ansi
	AnsiBBS   model.AnsiBBS
	BBS       model.BBS
	BBSText   model.BBSText
	BBStro    model.BBStro
	Demoscene model.Demoscene
	MsDos     model.MsDos
	Intro     model.Intro
	IntroD    model.IntroMsDos
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

// Statistics returns the empty database statistics for the artifacts categories.
func Statistics() Stats {
	return Stats{}
}

// Get and store the database statistics for the artifacts categories.
func (s *Stats) Get(ctx context.Context, exec boil.ContextExecutor) error {
	if err := s.Record.Public(ctx, exec); err != nil {
		return fmt.Errorf("category get record stat: %w", err)
	}
	if err := s.Ansi.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get ansi stat: %w", err)
	}
	if err := s.AnsiBBS.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get ansiBBS stat: %w", err)
	}
	if err := s.BBS.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get bbs stat: %w", err)
	}
	if err := s.BBSText.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get bbs trext stat: %w", err)
	}
	if err := s.BBStro.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get bbstro stat: %w", err)
	}
	if err := s.MsDos.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get msdos stat: %w", err)
	}
	if err := s.Intro.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get intro stat: %w", err)
	}
	if err := s.IntroD.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get introd stat: %w", err)
	}
	if err := s.IntroW.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get introw stat: %w", err)
	}
	if err := s.Installer.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get installer stat: %w", err)
	}
	if err := s.Java.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get java stat: %w", err)
	}
	if err := s.Linux.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get linux stat: %w", err)
	}
	if err := s.Demoscene.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get demoscene stat: %w", err)
	}
	return s.get(ctx, exec)
}

func (s *Stats) get(ctx context.Context, exec boil.ContextExecutor) error {
	if err := s.Macos.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get macos stat: %w", err)
	}
	if err := s.Magazine.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get magazine stat: %w", err)
	}
	if err := s.Nfo.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get nfo stat: %w", err)
	}
	if err := s.NfoTool.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get nfoTool stat: %w", err)
	}
	if err := s.Proof.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get proof stat: %w", err)
	}
	if err := s.Script.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get script stat: %w", err)
	}
	if err := s.Text.Stat(ctx, exec); err != nil {
		return fmt.Errorf("category get text stat: %w", err)
	}
	return s.Windows.Stat(ctx, exec)
}

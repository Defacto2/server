package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Defacto2/server/model"
	"github.com/Defacto2/server/pkg/postgres/models"
	"github.com/Defacto2/server/pkg/tags"
)

// Package file render_model.go contains the database queries for the renders.

// Records returns the records for the file category URI.
//
//nolint:maintidx,gocyclo
func Records(ctx context.Context, db *sql.DB, uri string, page, limit int) (models.FileSlice, error) {
	if db == nil {
		return nil, ErrDB
	}
	switch Match(uri) {
	// pulldown menu matches
	case newUploads:
		r := model.Files{}
		return r.List(ctx, db, page, limit)
	case newUpdates:
		r := model.Files{}
		return r.ListUpdates(ctx, db, page, limit)
	case oldest:
		r := model.Files{}
		return r.ListOldest(ctx, db, page, limit)
	case newest:
		r := model.Files{}
		return r.ListNewest(ctx, db, page, limit)
	// file categories matches
	case advert:
		r := model.Advert{}
		return r.List(ctx, db, page, limit)
	case announcement:
		r := model.Announcement{}
		return r.List(ctx, db, page, limit)
	case ansi:
		r := model.Ansi{}
		return r.List(ctx, db, page, limit)
	case ansiBrand:
		r := model.AnsiBrand{}
		return r.List(ctx, db, page, limit)
	case ansiBBS:
		r := model.AnsiBBS{}
		return r.List(ctx, db, page, limit)
	case ansiFTP:
		r := model.AnsiFTP{}
		return r.List(ctx, db, page, limit)
	case ansiNfo:
		r := model.AnsiNfo{}
		return r.List(ctx, db, page, limit)
	case ansiPack:
		r := model.AnsiPack{}
		return r.List(ctx, db, page, limit)
	case bbs:
		r := model.BBS{}
		return r.List(ctx, db, page, limit)
	case bbsImage:
		r := model.BBSImage{}
		return r.List(ctx, db, page, limit)
	case bbstro:
		r := model.BBStro{}
		return r.List(ctx, db, page, limit)
	case bbsText:
		r := model.BBSText{}
		return r.List(ctx, db, page, limit)
	case database:
		r := model.Database{}
		return r.List(ctx, db, page, limit)
	case demoscene:
		r := model.Demoscene{}
		return r.List(ctx, db, page, limit)
	case drama:
		r := model.Drama{}
		return r.List(ctx, db, page, limit)
	case ftp:
		r := model.FTP{}
		return r.List(ctx, db, page, limit)
	case hack:
		r := model.Hack{}
		return r.List(ctx, db, page, limit)
	case html:
		r := model.HTML{}
		return r.List(ctx, db, page, limit)
	case howTo:
		r := model.HowTo{}
		return r.List(ctx, db, page, limit)
	case image:
		r := model.Image{}
		return r.List(ctx, db, page, limit)
	case imagePack:
		r := model.ImagePack{}
		return r.List(ctx, db, page, limit)
	case installer:
		r := model.Installer{}
		return r.List(ctx, db, page, limit)
	case intro:
		r := model.Intro{}
		return r.List(ctx, db, page, limit)
	case linux:
		r := model.Linux{}
		return r.List(ctx, db, page, limit)
	case java:
		r := model.Java{}
		return r.List(ctx, db, page, limit)
	case jobAdvert:
		r := model.JobAdvert{}
		return r.List(ctx, db, page, limit)
	case macos:
		r := model.Macos{}
		return r.List(ctx, db, page, limit)
	case msdosPack:
		r := model.MsDosPack{}
		return r.List(ctx, db, page, limit)
	case music:
		r := model.Music{}
		return r.List(ctx, db, page, limit)
	case newsArticle:
		r := model.NewsArticle{}
		return r.List(ctx, db, page, limit)
	case nfo:
		r := model.Nfo{}
		return r.List(ctx, db, page, limit)
	case nfoTool:
		r := model.NfoTool{}
		return r.List(ctx, db, page, limit)
	case standards:
		r := model.Standard{}
		return r.List(ctx, db, page, limit)
	case script:
		r := model.Script{}
		return r.List(ctx, db, page, limit)
	case introMsdos:
		r := model.IntroMsDos{}
		return r.List(ctx, db, page, limit)
	case introWindows:
		r := model.IntroWindows{}
		return r.List(ctx, db, page, limit)
	case magazine:
		r := model.Magazine{}
		return r.List(ctx, db, page, limit)
	case msdos:
		r := model.MsDos{}
		return r.List(ctx, db, page, limit)
	case pdf:
		r := model.PDF{}
		return r.List(ctx, db, page, limit)
	case proof:
		r := model.Proof{}
		return r.List(ctx, db, page, limit)
	case restrict:
		r := model.Restrict{}
		return r.List(ctx, db, page, limit)
	case takedown:
		r := model.Takedown{}
		return r.List(ctx, db, page, limit)
	case text:
		r := model.Text{}
		return r.List(ctx, db, page, limit)
	case textAmiga:
		r := model.TextAmiga{}
		return r.List(ctx, db, page, limit)
	case textApple2:
		r := model.TextApple2{}
		return r.List(ctx, db, page, limit)
	case textAtariST:
		r := model.TextAtariST{}
		return r.List(ctx, db, page, limit)
	case textPack:
		r := model.TextPack{}
		return r.List(ctx, db, page, limit)
	case tool:
		r := model.Tool{}
		return r.List(ctx, db, page, limit)
	case trialCrackme:
		r := model.TrialCrackme{}
		return r.List(ctx, db, page, limit)
	case video:
		r := model.Video{}
		return r.List(ctx, db, page, limit)
	case windows:
		r := model.Windows{}
		return r.List(ctx, db, page, limit)
	case windowsPack:
		r := model.WindowsPack{}
		return r.List(ctx, db, page, limit)
	default:
		return nil, fmt.Errorf("unknown file category: %s", uri)
	}
}

// Records returns the records for the file category URI.
//
//nolint:maintidx,gocyclo,nolintlint
func RecordsSub(uri string) string {
	const ignore = -1
	switch Match(uri) {
	case advert:
		return tags.Humanizes(ignore, tags.ForSale)
	case announcement:
		return tags.Humanizes(ignore, tags.Announcement)
	case ansi:
		return tags.Humanizes(tags.ANSI, ignore)
	case ansiBrand:
		return tags.Humanizes(tags.ANSI, tags.Logo)
	case ansiBBS:
		return tags.Humanizes(tags.ANSI, tags.BBS)
	case ansiFTP:
		return tags.Humanizes(tags.ANSI, tags.Ftp)
	case ansiNfo:
		return tags.Humanizes(tags.ANSI, tags.Nfo)
	case ansiPack:
		return tags.Humanizes(tags.ANSI, tags.Pack)
	case bbs:
		return tags.Humanizes(ignore, tags.BBS)
	case bbsImage:
		return tags.Humanizes(tags.Image, tags.BBS)
	case bbstro:
		return tags.Humanizes(tags.DOS, tags.BBS)
	case bbsText:
		return tags.Humanizes(tags.Text, tags.BBS)
	case database:
		return tags.Humanizes(ignore, tags.DataB)
	case demoscene:
		return tags.Humanizes(ignore, tags.Demo)
	case drama:
		return tags.Humanizes(ignore, tags.Drama)
	case ftp:
		return tags.Humanizes(ignore, tags.Ftp)
	case hack:
		return tags.Humanizes(ignore, tags.GameHack)
	case html:
		return uri
	case howTo:
		return tags.Humanizes(ignore, tags.Guide)
	case image:
		return tags.Humanizes(tags.Image, ignore)
	case imagePack:
		return tags.Humanizes(tags.Image, tags.Pack)
	case installer:
		return tags.Humanizes(ignore, tags.Install)
	case intro:
		return tags.Humanizes(ignore, tags.Intro)
	case linux:
		return tags.Humanizes(tags.Linux, ignore)
	case java:
		return tags.Humanizes(tags.Java, ignore)
	case jobAdvert:
		return tags.Humanizes(ignore, tags.Job)
	case macos:
		return tags.Humanizes(tags.Mac, ignore)
	case msdosPack:
		return tags.Humanizes(tags.DOS, tags.Pack)
	case music:
		return tags.Humanizes(tags.Audio, ignore)
	case newsArticle:
		return tags.Humanizes(ignore, tags.News)
	case nfo:
		return tags.Humanizes(ignore, tags.Nfo)
	case nfoTool:
		return tags.Humanizes(ignore, tags.NfoTool)
	case standards:
		return tags.Humanizes(ignore, tags.Rule)
	case script:
		return tags.Humanizes(tags.PHP, ignore)
	case introMsdos:
		return tags.Humanizes(tags.DOS, tags.Intro)
	case introWindows:
		return tags.Humanizes(tags.Windows, tags.Intro)
	case magazine:
		return tags.Humanizes(ignore, tags.Mag)
	case msdos:
		return tags.Humanizes(tags.DOS, ignore)
	case pdf:
		return tags.Humanizes(tags.PDF, ignore)
	case proof:
		return tags.Humanizes(ignore, tags.Proof)
	case restrict:
		return tags.Humanizes(ignore, tags.Restrict)
	case takedown:
		return tags.Humanizes(ignore, tags.Bust)
	case text:
		return tags.Humanizes(tags.Text, ignore)
	case textAmiga:
		return tags.Humanizes(tags.TextAmiga, ignore)
	case textApple2:
		return tags.Humanizes(tags.Text, tags.AppleII)
	case textAtariST:
		return tags.Humanizes(tags.Text, tags.AtariST)
	case textPack:
		return tags.Humanizes(tags.Text, tags.Pack)
	case tool:
		return tags.Humanizes(ignore, tags.Tool)
	case trialCrackme:
		return tags.Humanizes(tags.Windows, tags.Job)
	case video:
		return tags.Humanizes(tags.Video, ignore)
	case windows:
		return tags.Humanizes(tags.Windows, ignore)
	case windowsPack:
		return tags.Humanizes(tags.Windows, tags.Pack)
	default:
		return "unknown uri"
	}
}

// Stats are the database statistics for the file categories.
type Stats struct { //nolint:gochecknoglobals
	Record    model.Files
	Ansi      model.Ansi
	AnsiBBS   model.AnsiBBS
	BBS       model.BBS
	BBSText   model.BBSText
	BBStro    model.BBStro
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

// Get and store the database statistics for the file categories.
func (s *Stats) Get(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return ErrDB
	}
	if err := s.Record.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Ansi.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.AnsiBBS.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.BBS.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.BBSText.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.BBStro.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.MsDos.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Intro.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.IntroD.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.IntroW.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Installer.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Java.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Linux.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Demoscene.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Macos.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Magazine.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Nfo.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.NfoTool.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Proof.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Script.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Text.Stat(ctx, db); err != nil {
		return err
	}
	if err := s.Windows.Stat(ctx, db); err != nil {
		return err
	}
	return nil
}

// Statistics returns the empty database statistics for the file categories.
func Statistics() Stats {
	return Stats{}
}

package querymod_test

import (
	"testing"

	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model/querymod"
	"github.com/nalgeon/be"
)

// Test section functions return valid null.String values.

func TestSAdvertReturnsValidTag(t *testing.T) {
	result := querymod.SAdvert()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.ForSale])
}

func TestSAnnouncementReturnsValidTag(t *testing.T) {
	result := querymod.SAnnouncement()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.Announcement])
}

func TestSAppleIIReturnsValidTag(t *testing.T) {
	result := querymod.SAppleII()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.AppleII])
}

func TestSAtariSTReturnsValidTag(t *testing.T) {
	result := querymod.SAtariST()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.AtariST])
}

func TestSBbsReturnsValidTag(t *testing.T) {
	result := querymod.SBbs()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.BBS])
}

func TestSBrandReturnsValidTag(t *testing.T) {
	result := querymod.SBrand()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.Logo])
}

func TestSDemoReturnsValidTag(t *testing.T) {
	result := querymod.SDemo()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.Demo])
}

func TestSDramaReturnsValidTag(t *testing.T) {
	result := querymod.SDrama()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.Drama])
}

func TestSFtpReturnsValidTag(t *testing.T) {
	result := querymod.SFtp()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.Ftp])
}

func TestSHackReturnsValidTag(t *testing.T) {
	result := querymod.SHack()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.GameHack])
}

func TestSHowToReturnsValidTag(t *testing.T) {
	result := querymod.SHowTo()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.Guide])
}

func TestSInstallReturnsValidTag(t *testing.T) {
	result := querymod.SInstall()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.Install])
}

func TestSIntroReturnsValidTag(t *testing.T) {
	result := querymod.SIntro()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.Intro])
}

func TestSJobAdvertReturnsValidTag(t *testing.T) {
	result := querymod.SJobAdvert()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.Job])
}

func TestSMagReturnsValidTag(t *testing.T) {
	result := querymod.SMag()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.Mag])
}

func TestSNewsReturnsValidTag(t *testing.T) {
	result := querymod.SNews()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.News])
}

func TestSNfoReturnsValidTag(t *testing.T) {
	result := querymod.SNfo()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.Nfo])
}

func TestSNfoToolReturnsValidTag(t *testing.T) {
	result := querymod.SNfoTool()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.NfoTool])
}

func TestSPackReturnsValidTag(t *testing.T) {
	result := querymod.SPack()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.Pack])
}

func TestSProofReturnsValidTag(t *testing.T) {
	result := querymod.SProof()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.Proof])
}

func TestSRestrictReturnsValidTag(t *testing.T) {
	result := querymod.SRestrict()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.Restrict])
}

func TestSStandardReturnsValidTag(t *testing.T) {
	result := querymod.SStandard()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.Rule])
}

func TestSTakedownReturnsValidTag(t *testing.T) {
	result := querymod.STakedown()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.Bust])
}

func TestSToolReturnsValidTag(t *testing.T) {
	result := querymod.STool()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.Tool])
}

// Test platform functions return valid null.String values.

func TestPAnsiReturnsValidTag(t *testing.T) {
	result := querymod.PAnsi()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.ANSI])
}

func TestPDatabaseReturnsValidTag(t *testing.T) {
	result := querymod.PDatabase()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.DataB])
}

func TestPDosReturnsValidTag(t *testing.T) {
	result := querymod.PDos()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.DOS])
}

func TestPHtmlReturnsValidTag(t *testing.T) {
	result := querymod.PHtml()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.Markup])
}

func TestPImageReturnsValidTag(t *testing.T) {
	result := querymod.PImage()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.Image])
}

func TestPLinuxReturnsValidTag(t *testing.T) {
	result := querymod.PLinux()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.Linux])
}

func TestPJavaReturnsValidTag(t *testing.T) {
	result := querymod.PJava()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.Java])
}

func TestPMacReturnsValidTag(t *testing.T) {
	result := querymod.PMac()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.Mac])
}

func TestPMusicReturnsValidTag(t *testing.T) {
	result := querymod.PMusic()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.Audio])
}

func TestPPdfReturnsValidTag(t *testing.T) {
	result := querymod.PPdf()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.PDF])
}

func TestPScriptReturnsValidTag(t *testing.T) {
	result := querymod.PScript()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.PHP])
}

func TestPTextReturnsValidTag(t *testing.T) {
	result := querymod.PText()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.Text])
}

func TestPTextAmigaReturnsValidTag(t *testing.T) {
	result := querymod.PTextAmiga()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.TextAmiga])
}

func TestPVideoReturnsValidTag(t *testing.T) {
	result := querymod.PVideo()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.Video])
}

func TestPWindowsReturnsValidTag(t *testing.T) {
	result := querymod.PWindows()
	be.True(t, result.Valid)
	be.Equal(t, result.String, tags.URIs()[tags.Windows])
}

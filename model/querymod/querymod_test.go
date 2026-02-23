package querymod_test

import (
	"testing"

	"github.com/Defacto2/server/model/querymod"
	"github.com/nalgeon/be"
)

// Test that each Expr function returns a valid QueryMod.
func TestAdvertExpr(t *testing.T) {
	expr := querymod.AdvertExpr()
	be.True(t, expr != nil)
}

func TestAnnouncementExpr(t *testing.T) {
	expr := querymod.AnnouncementExpr()
	be.True(t, expr != nil)
}

func TestAnsiExpr(t *testing.T) {
	expr := querymod.AnsiExpr()
	be.True(t, expr != nil)
}

func TestAnsiBBSExpr(t *testing.T) {
	expr := querymod.AnsiBBSExpr()
	be.True(t, expr != nil)
}

func TestAnsiFTPExpr(t *testing.T) {
	expr := querymod.AnsiFTPExpr()
	be.True(t, expr != nil)
}

func TestAnsiNfoExpr(t *testing.T) {
	expr := querymod.AnsiNfoExpr()
	be.True(t, expr != nil)
}

func TestAnsiPackExpr(t *testing.T) {
	expr := querymod.AnsiPackExpr()
	be.True(t, expr != nil)
}

func TestAnsiBrandExpr(t *testing.T) {
	expr := querymod.AnsiBrandExpr()
	be.True(t, expr != nil)
}

func TestAppleIIExpr(t *testing.T) {
	expr := querymod.AppleIIExpr()
	be.True(t, expr != nil)
}

func TestAtariSTExpr(t *testing.T) {
	expr := querymod.AtariSTExpr()
	be.True(t, expr != nil)
}

func TestBBSExpr(t *testing.T) {
	expr := querymod.BBSExpr()
	be.True(t, expr != nil)
}

func TestBBSImageExpr(t *testing.T) {
	expr := querymod.BBSImageExpr()
	be.True(t, expr != nil)
}

func TestBBStroExpr(t *testing.T) {
	expr := querymod.BBStroExpr()
	be.True(t, expr != nil)
}

func TestBBSTextExpr(t *testing.T) {
	expr := querymod.BBSTextExpr()
	be.True(t, expr != nil)
}

func TestDemoExpr(t *testing.T) {
	expr := querymod.DemoExpr()
	be.True(t, expr != nil)
}

func TestDramaExpr(t *testing.T) {
	expr := querymod.DramaExpr()
	be.True(t, expr != nil)
}

func TestFTPExpr(t *testing.T) {
	expr := querymod.FTPExpr()
	be.True(t, expr != nil)
}

func TestHackExpr(t *testing.T) {
	expr := querymod.HackExpr()
	be.True(t, expr != nil)
}

func TestInstallExpr(t *testing.T) {
	expr := querymod.InstallExpr()
	be.True(t, expr != nil)
}

func TestIntroExpr(t *testing.T) {
	expr := querymod.IntroExpr()
	be.True(t, expr != nil)
}

func TestIntroDOSExpr(t *testing.T) {
	expr := querymod.IntroDOSExpr()
	be.True(t, expr != nil)
}

func TestIntroWindowsExpr(t *testing.T) {
	expr := querymod.IntroWindowsExpr()
	be.True(t, expr != nil)
}

func TestJobAdvertExpr(t *testing.T) {
	expr := querymod.JobAdvertExpr()
	be.True(t, expr != nil)
}

func TestDatabaseExpr(t *testing.T) {
	expr := querymod.DatabaseExpr()
	be.True(t, expr != nil)
}

func TestDOSExpr(t *testing.T) {
	expr := querymod.DOSExpr()
	be.True(t, expr != nil)
}

func TestDosPackExpr(t *testing.T) {
	expr := querymod.DosPackExpr()
	be.True(t, expr != nil)
}

func TestHTMLExpr(t *testing.T) {
	expr := querymod.HTMLExpr()
	be.True(t, expr != nil)
}

func TestHowToExpr(t *testing.T) {
	expr := querymod.HowToExpr()
	be.True(t, expr != nil)
}

func TestLinuxExpr(t *testing.T) {
	expr := querymod.LinuxExpr()
	be.True(t, expr != nil)
}

func TestJavaExpr(t *testing.T) {
	expr := querymod.JavaExpr()
	be.True(t, expr != nil)
}

func TestImageExpr(t *testing.T) {
	expr := querymod.ImageExpr()
	be.True(t, expr != nil)
}

func TestImagePackExpr(t *testing.T) {
	expr := querymod.ImagePackExpr()
	be.True(t, expr != nil)
}

func TestMacExpr(t *testing.T) {
	expr := querymod.MacExpr()
	be.True(t, expr != nil)
}

func TestMagExpr(t *testing.T) {
	expr := querymod.MagExpr()
	be.True(t, expr != nil)
}

func TestMusicExpr(t *testing.T) {
	expr := querymod.MusicExpr()
	be.True(t, expr != nil)
}

func TestNewsArticleExpr(t *testing.T) {
	expr := querymod.NewsArticleExpr()
	be.True(t, expr != nil)
}

func TestNfoExpr(t *testing.T) {
	expr := querymod.NfoExpr()
	be.True(t, expr != nil)
}

func TestNfoToolExpr(t *testing.T) {
	expr := querymod.NfoToolExpr()
	be.True(t, expr != nil)
}

func TestPDFExpr(t *testing.T) {
	expr := querymod.PDFExpr()
	be.True(t, expr != nil)
}

func TestProofExpr(t *testing.T) {
	expr := querymod.ProofExpr()
	be.True(t, expr != nil)
}

func TestRestrictExpr(t *testing.T) {
	expr := querymod.RestrictExpr()
	be.True(t, expr != nil)
}

func TestScriptExpr(t *testing.T) {
	expr := querymod.ScriptExpr()
	be.True(t, expr != nil)
}

func TestStandardExpr(t *testing.T) {
	expr := querymod.StandardExpr()
	be.True(t, expr != nil)
}

func TestTakedownExpr(t *testing.T) {
	expr := querymod.TakedownExpr()
	be.True(t, expr != nil)
}

func TestTextExpr(t *testing.T) {
	expr := querymod.TextExpr()
	be.True(t, expr != nil)
}

func TestTextAmigaExpr(t *testing.T) {
	expr := querymod.TextAmigaExpr()
	be.True(t, expr != nil)
}

func TestTextPackExpr(t *testing.T) {
	expr := querymod.TextPackExpr()
	be.True(t, expr != nil)
}

func TestToolExpr(t *testing.T) {
	expr := querymod.ToolExpr()
	be.True(t, expr != nil)
}

func TestTrialCrackmeExpr(t *testing.T) {
	expr := querymod.TrialCrackmeExpr()
	be.True(t, expr != nil)
}

func TestVideoExpr(t *testing.T) {
	expr := querymod.VideoExpr()
	be.True(t, expr != nil)
}

func TestWindowsExpr(t *testing.T) {
	expr := querymod.WindowsExpr()
	be.True(t, expr != nil)
}

func TestWindowsPackExpr(t *testing.T) {
	expr := querymod.WindowsPackExpr()
	be.True(t, expr != nil)
}

package querymod_test

import (
	"testing"

	"github.com/Defacto2/server/model/querymod"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/nalgeon/be"
)

func TestExprs(t *testing.T) {
	t.Parallel()

	exprs := []struct {
		name string
		fn   func() qm.QueryMod
	}{
		{"Advert", querymod.AdvertExpr},
		{"Announcement", querymod.AnnouncementExpr},
		{"Animation", querymod.AnimationExpr},
		{"OneOffs", querymod.OneOffsExpr},
		{"Ansi", querymod.AnsiExpr},
		{"AnsiBBS", querymod.AnsiBBSExpr},
		{"AnsiFTP", querymod.AnsiFTPExpr},
		{"AnsiNfo", querymod.AnsiNfoExpr},
		{"AnsiPack", querymod.AnsiPackExpr},
		{"AnsiBrand", querymod.AnsiBrandExpr},
		{"AppleII", querymod.AppleIIExpr},
		{"AtariST", querymod.AtariSTExpr},
		{"BBS", querymod.BBSExpr},
		{"BBSImage", querymod.BBSImageExpr},
		{"BBStro", querymod.BBStroExpr},
		{"BBSText", querymod.BBSTextExpr},
		{"Demo", querymod.DemoExpr},
		{"Drama", querymod.DramaExpr},
		{"FTP", querymod.FTPExpr},
		{"Hack", querymod.HackExpr},
		{"Install", querymod.InstallExpr},
		{"Intro", querymod.IntroExpr},
		{"IntroDOS", querymod.IntroDOSExpr},
		{"IntroWindows", querymod.IntroWindowsExpr},
		{"JobAdvert", querymod.JobAdvertExpr},
		{"Console", querymod.ConsoleExpr},
		{"Database", querymod.DatabaseExpr},
		{"DOS", querymod.DOSExpr},
		{"DosPack", querymod.DosPackExpr},
		{"HTML", querymod.HTMLExpr},
		{"HowTo", querymod.HowToExpr},
		{"Linux", querymod.LinuxExpr},
		{"Java", querymod.JavaExpr},
		{"Image", querymod.ImageExpr},
		{"ImagePack", querymod.ImagePackExpr},
		{"Mac", querymod.MacExpr},
		{"Mag", querymod.MagExpr},
		{"Music", querymod.MusicExpr},
		{"NewsArticle", querymod.NewsArticleExpr},
		{"Nfo", querymod.NfoExpr},
		{"NfoTool", querymod.NfoToolExpr},
		{"PDF", querymod.PDFExpr},
		{"Proof", querymod.ProofExpr},
		{"Restrict", querymod.RestrictExpr},
		{"Script", querymod.ScriptExpr},
		{"Standard", querymod.StandardExpr},
		{"Takedown", querymod.TakedownExpr},
		{"Text", querymod.TextExpr},
		{"TextAmiga", querymod.TextAmigaExpr},
		{"TextPack", querymod.TextPackExpr},
		{"Tool", querymod.ToolExpr},
		{"TrialCrackme", querymod.TrialCrackmeExpr},
		{"Video", querymod.VideoExpr},
		{"Windows", querymod.WindowsExpr},
		{"WindowsPack", querymod.WindowsPackExpr},
	}

	for _, tc := range exprs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			be.True(t, tc.fn() != nil)
		})
	}
}

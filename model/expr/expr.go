// Package expr provides the query mod expressions for the file database.
package expr

import (
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// AdvertExpr is a the query mod expression for sale adverts.
func AdvertExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SAdvert()),
	)
}

// AnnouncementExpr is a the query mod expression for announcement releases.
func AnnouncementExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SAnnouncement()),
	)
}

// AnsiExpr is a the query mod expression for ANSI art releases.
func AnsiExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PAnsi()),
	)
}

// AnsiBBSExpr is a the query mod expression for ANSI BBS art releases.
func AnsiBBSExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PAnsi()),
		models.FileWhere.Section.EQ(SBbs()),
	)
}

// AnsiFTPExpr is a the query mod expression for ANSI FTP art releases.
func AnsiFTPExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PAnsi()),
		models.FileWhere.Section.EQ(SFtp()),
	)
}

// AnsiNfoExpr is a the query mod expression for ANSI NFO releases.
func AnsiNfoExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PAnsi()),
		models.FileWhere.Section.EQ(SNfo()),
	)
}

// AnsiPackExpr is a the query mod expression for ANSI art packs.
func AnsiPackExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PAnsi()),
		models.FileWhere.Section.EQ(SPack()),
	)
}

// AnsiBrandExpr is a the query mod expression for ANSI brand art releases.
func AnsiBrandExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PAnsi()),
		models.FileWhere.Section.EQ(SBrand()),
	)
}

// AppleIIExpr is a the query mod expression for Apple II text releases.
func AppleIIExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SAppleII()),
	)
}

// AtariSTExpr is a the query mod expression for Atari ST text releases.
func AtariSTExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SAtariST()),
	)
}

// BBSExpr is a the query mod expression for BBS releases.
func BBSExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SBbs()),
	)
}

// BBSImageExpr is a the query mod expression for BBS image releases.
func BBSImageExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SBbs()),
		models.FileWhere.Platform.EQ(PImage()),
	)
}

// BBStroExpr is a the query mod expression for BBS releases on MS-DOS.
func BBStroExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SBbs()),
		models.FileWhere.Platform.EQ(PDos()),
	)
}

// BBSTextExpr is a the query mod expression for BBS text releases.
func BBSTextExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SBbs()),
		models.FileWhere.Platform.EQ(PText()),
	)
}

// DemoExpr is a the query mod expression for demo releases.
func DemoExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SDemo()),
	)
}

// DramaExpr is a the query mod expression for community drama.
func DramaExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SDrama()),
	)
}

// FTPExpr is a the query mod expression for FTP releases.
func FTPExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SFtp()),
	)
}

// HackExpr is a the query mod expression for hack releases.
func HackExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SHack()),
	)
}

// InstallExpr is a the query mod expression for install releases.
func InstallExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SInstall()),
	)
}

// IntroExpr is a the query mod expression for intro releases.
func IntroExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SIntro()),
	)
}

// IntroDOSExpr is a the query mod expression for intro releases on MS-DOS.
func IntroDOSExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SIntro()),
		models.FileWhere.Platform.EQ(PDos()),
	)
}

// IntroWindowsExpr is a the query mod expression for Windows intro releases.
func IntroWindowsExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SIntro()),
		models.FileWhere.Platform.EQ(PWindows()),
	)
}

// JobAdvertExpr is a the query mod expression for job advert releases.
func JobAdvertExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SJobAdvert()),
	)
}

// DatabaseExpr is a the query mod expression for database releases.
func DatabaseExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PDatabase()),
	)
}

// DOSExpr is a the query mod expression for MS-DOS releases.
func DOSExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PDos()),
	)
}

// DosPackExpr is a the query mod expression for MS-DOS packs.
func DosPackExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PDos()),
		models.FileWhere.Section.EQ(SPack()),
	)
}

// HTMLExpr is a the query mod expression for HTML releases.
func HTMLExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PHtml()),
	)
}

// HowToExpr is a the query mod expression for guides and how-tos.
func HowToExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SHowTo()),
	)
}

// LinuxExpr is a the query mod expression for Linux releases.
func LinuxExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PLinux()),
	)
}

// JavaExpr is a the query mod expression for Java releases.
func JavaExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PJava()),
	)
}

// ImageExpr is a the query mod expression for images.
func ImageExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PImage()),
	)
}

// ImagePackExpr is a the query mod expression for image file packs.
func ImagePackExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PImage()),
		models.FileWhere.Section.EQ(SPack()),
	)
}

// MacExpr is a the query mod expression for MacOS releases.
func MacExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PMac()),
	)
}

// MagExpr is a the query mod expression for magazine releases.
func MagExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SMag()),
	)
}

// MusicExpr is a the query mod expression for music releases.
func MusicExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PMusic()),
	)
}

// NewsArticleExpr is a the query mod expression for news articles.
func NewsArticleExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SNews()),
	)
}

// NfoExpr is a the query mod expression for NFO releases.
func NfoExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SNfo()),
	)
}

// NfoToolExpr is a the query mod expression for NFO tool releases.
func NfoToolExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SNfoTool()),
	)
}

// PDFExpr is a the query mod expression for PDF releases.
func PDFExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PPdf()),
	)
}

// ProofExpr is a the query mod expression for proof releases.
func ProofExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SProof()),
	)
}

// RestrictExpr is a the query mod expression for restricted releases.
func RestrictExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SRestrict()),
	)
}

// ScriptExpr is a the query mod expression for script and shell releases.
func ScriptExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PScript()),
	)
}

// StandardExpr is a the query mod expression for release standards and rules.
func StandardExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SStandard()),
	)
}

// TakedownExpr is a the query mod expression for bust and takedowns.
func TakedownExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(STakedown()),
	)
}

// TextExpr is a the query mod expression for textfile releases.
func TextExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PText()),
	)
}

// TextAmigaExpr is a the query mod expression for Amiga text releases.
func TextAmigaExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PTextAmiga()),
	)
}

// TextPackExpr is a the query mod expression for text releases.
func TextPackExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PText()),
		models.FileWhere.Section.EQ(SPack()),
	)
}

// ToolExpr is a the query mod expression for computer tool releases.
func ToolExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(STool()),
	)
}

// TrialCrackmeExpr is a the query mod expression for trial crackme releases.
func TrialCrackmeExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Section.EQ(SJobAdvert()),
		models.FileWhere.Platform.EQ(PWindows()),
	)
}

// VideoExpr is a the query mod expression for video releases.
func VideoExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PVideo()),
	)
}

// WindowsExpr is a the query mod expression for Windows releases.
func WindowsExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PWindows()),
	)
}

// WindowsPackExpr is a the query mod expression for Windows file packs.
func WindowsPackExpr() qm.QueryMod {
	return qm.Expr(
		models.FileWhere.Platform.EQ(PWindows()),
		models.FileWhere.Section.EQ(SPack()),
	)
}

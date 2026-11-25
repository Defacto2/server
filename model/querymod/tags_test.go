package querymod_test

import (
	"testing"

	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model/querymod"
	"github.com/aarondl/null/v8"
	"github.com/nalgeon/be"
)

func TestTags(t *testing.T) {
	t.Parallel()

	uris := tags.URIs()
	cases := []struct {
		name string
		fn   func() null.String
		want string
	}{
		// Sections
		{"SAdvert", querymod.SAdvert, uris[tags.ForSale]},
		{"SAnnouncement", querymod.SAnnouncement, uris[tags.Announcement]},
		{"SAppleII", querymod.SAppleII, uris[tags.AppleII]},
		{"SAtariST", querymod.SAtariST, uris[tags.AtariST]},
		{"SBbs", querymod.SBbs, uris[tags.BBS]},
		{"SBrand", querymod.SBrand, uris[tags.Logo]},
		{"SDemo", querymod.SDemo, uris[tags.Demo]},
		{"SDrama", querymod.SDrama, uris[tags.Drama]},
		{"SFtp", querymod.SFtp, uris[tags.Ftp]},
		{"SHack", querymod.SHack, uris[tags.GameHack]},
		{"SHowTo", querymod.SHowTo, uris[tags.Guide]},
		{"SInstall", querymod.SInstall, uris[tags.Install]},
		{"SIntro", querymod.SIntro, uris[tags.Intro]},
		{"SJobAdvert", querymod.SJobAdvert, uris[tags.Job]},
		{"SMag", querymod.SMag, uris[tags.Mag]},
		{"SNews", querymod.SNews, uris[tags.News]},
		{"SNfo", querymod.SNfo, uris[tags.Nfo]},
		{"SNfoTool", querymod.SNfoTool, uris[tags.NfoTool]},
		{"SPack", querymod.SPack, uris[tags.Pack]},
		{"SProof", querymod.SProof, uris[tags.Proof]},
		{"SRestrict", querymod.SRestrict, uris[tags.Restrict]},
		{"SStandard", querymod.SStandard, uris[tags.Rule]},
		{"STakedown", querymod.STakedown, uris[tags.Bust]},
		{"STool", querymod.STool, uris[tags.Tool]},
		// Platforms
		{"PAnsi", querymod.PAnsi, uris[tags.ANSI]},
		{"PDatabase", querymod.PDatabase, uris[tags.DataB]},
		{"PDos", querymod.PDos, uris[tags.DOS]},
		{"PHtml", querymod.PHtml, uris[tags.Markup]},
		{"PImage", querymod.PImage, uris[tags.Image]},
		{"PLinux", querymod.PLinux, uris[tags.Linux]},
		{"PJava", querymod.PJava, uris[tags.Java]},
		{"PMac", querymod.PMac, uris[tags.Mac]},
		{"PMusic", querymod.PMusic, uris[tags.Audio]},
		{"PPdf", querymod.PPdf, uris[tags.PDF]},
		{"PScript", querymod.PScript, uris[tags.PHP]},
		{"PText", querymod.PText, uris[tags.Text]},
		{"PTextAmiga", querymod.PTextAmiga, uris[tags.TextAmiga]},
		{"PVideo", querymod.PVideo, uris[tags.Video]},
		{"PWindows", querymod.PWindows, uris[tags.Windows]},
		// Magic
		{"MAnimation", querymod.MAnimation, "ANSImation"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tc.fn()
			be.True(t, got.Valid)
			be.Equal(t, got.String, tc.want)
		})
	}
}

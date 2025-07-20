package model_test

import (
	"testing"

	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/nalgeon/be"
	"github.com/volatiletech/null/v8"
)

func TestSiteAd(t *testing.T) {
	empty := null.StringFrom("")
	rsomeone := null.StringFrom("someone")
	rftp := null.StringFrom("some site fTp ") // test case and whitespace
	rbbs := null.StringFrom("some board bBS") // test casing
	snfo := null.StringFrom(tags.Nfo.String())
	sexe := null.StringFrom(tags.Intro.String())
	sftp := null.StringFrom(tags.Ftp.String())
	sbbs := null.StringFrom(tags.BBS.String())

	got := model.SiteAd(empty, empty)
	be.Equal(t, got, empty)
	got = model.SiteAd(rftp, sexe)
	be.Equal(t, got, sexe)
	got = model.SiteAd(rftp, snfo)
	be.Equal(t, got, sftp)
	got = model.SiteAd(rbbs, snfo)
	be.Equal(t, got, sbbs)
	got = model.SiteAd(rsomeone, snfo)
	be.Equal(t, got, snfo)
}

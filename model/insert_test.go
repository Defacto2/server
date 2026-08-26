package model_test

import (
	"testing"

	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/null/v8"
	"github.com/nalgeon/be"
)

func TestSiteAd(t *testing.T) {
	t.Parallel()

	empty := null.StringFrom("")

	got := model.SiteAd(empty, empty)
	be.Equal(t, got, empty)

	sexe := null.StringFrom(tags.Intro.String())
	rftp := null.StringFrom("some site fTp ") // test case and white space
	got = model.SiteAd(rftp, sexe)
	be.Equal(t, got, sexe)

	sftp := null.StringFrom(tags.Ftp.String())
	snfo := null.StringFrom(tags.Nfo.String())
	got = model.SiteAd(rftp, snfo)
	be.Equal(t, got, sftp)

	rbbs := null.StringFrom("some board bBS") // test casing
	sbbs := null.StringFrom(tags.BBS.String())
	got = model.SiteAd(rbbs, snfo)
	be.Equal(t, got, sbbs)

	rsomeone := null.StringFrom("someone")
	got = model.SiteAd(rsomeone, snfo)
	be.Equal(t, got, snfo)
}

//nolint:testpackage
package api

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Defacto2/helper"
	"github.com/Defacto2/server/internal/logs"
	"github.com/Defacto2/server/internal/postgres"
	"github.com/Defacto2/server/internal/postgres/models"
	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
	"github.com/aarondl/null/v8"
	"github.com/labstack/echo/v5"
	"github.com/nalgeon/be"
)

func TestTagsAPIEmpty(t *testing.T) {
	t.Parallel()

	e := echo.New()
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v0/tags", nil)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)

	err := TagsAPI(c, &sql.DB{}, false, false)
	be.Err(t, err, nil)
	be.Equal(t, w.Code, http.StatusOK)
	be.Equal(t, strings.TrimSpace(w.Body.String()), "[]")
}

func TestBuildTags(t *testing.T) {
	t.Parallel()

	bbs := tags.BBS.String()
	dos := tags.DOS.String()
	sectionStats := map[string]model.TagStat{
		bbs: {Count: 3, Bytes: 1024},
	}
	platformStats := map[string]model.TagStat{
		dos: {Count: 5, Bytes: 2048},
	}

	cats := buildTags(tags.Infos(), true, false, sectionStats, nil)
	be.True(t, len(cats) > 0)
	var foundBBS, foundDOS bool
	for _, item := range cats {
		if item.Name == bbs {
			foundBBS = true
			be.Equal(t, item.Stats.TotalFiles, int64(3))
			be.Equal(t, item.Stats.TotalSizeBytes, int64(1024))
		}
		if item.Name == dos {
			foundDOS = true
		}
	}
	be.True(t, foundBBS)
	be.True(t, !foundDOS)

	plats := buildTags(tags.Infos(), false, true, nil, platformStats)
	be.True(t, len(plats) > 0)
	foundBBS, foundDOS = false, false
	for _, item := range plats {
		if item.Name == bbs {
			foundBBS = true
		}
		if item.Name == dos {
			foundDOS = true
			be.Equal(t, item.Stats.TotalFiles, int64(5))
			be.Equal(t, item.Stats.TotalSizeBytes, int64(2048))
		}
	}
	be.True(t, !foundBBS)
	be.True(t, foundDOS)
}

func TestExternalLinks(t *testing.T) {
	t.Parallel()

	be.Equal(t, linkSixteen("razor-1911"), "https://16colo.rs/tags/content/razor%201911")
	be.Equal(t, linkSixteen("not-a-group"), "")

	be.Equal(t, linkJaneway("razor-1911"), "https://janeway.exotica.org.uk/author.php?id=338")
	be.Equal(t, linkJaneway("not-a-group"), "")

	be.Equal(t, linkCsdb("razor-1911"), "https://csdb.dk/group/?id=431")
	be.Equal(t, linkCsdb("not-a-group"), "")

	be.Equal(t, linkDemozoo("defacto2"), "https://demozoo.org/groups/10000")
	be.Equal(t, linkDemozoo("not-a-group"), "")
}

func TestScenersAPIHelper(t *testing.T) {
	t.Parallel()

	got := scenersAPI(model.Sceners{{Name: "Zeta"}, {Name: "Alpha"}})
	be.Equal(t, len(got), 2)
	be.Equal(t, got[0].Name, "alpha")
	be.Equal(t, got[0].Title, "Alpha")
	be.Equal(t, got[0].URLs.API, APIBase+"/scener/alpha")
	be.True(t, got[0].ID != "")
	be.Equal(t, got[1].Name, "zeta")
}

func TestCacheHelpers(t *testing.T) {
	t.Parallel()

	const key = "test_cache_helper_key"
	t.Cleanup(func() {
		apiCacheMu.Lock()
		delete(apiCache, key)
		delete(apiCache, "tags_category=true_platform=true")
		apiCacheMu.Unlock()
	})

	_, ok := cachedResults(key)
	be.Equal(t, ok, false)

	cacheResults(key, "hello")
	val, ok := cachedResults(key)
	be.Equal(t, ok, true)
	be.Equal(t, val.(string), "hello")

	// Expired entry should miss.
	apiCacheMu.Lock()
	apiCache[key] = cacheQueryItem{data: "stale", expires: time.Now().Add(-time.Minute)}
	apiCacheMu.Unlock()
	_, ok = cachedResults(key)
	be.Equal(t, ok, false)

	// tagsCache + cachedTags + TagsAPI cache hit path.
	want := []tagAPI{{ID: 1, Name: "demo"}}
	tagsCache(true, true, want)
	gotTags, ok := cachedTags(true, true)
	be.Equal(t, ok, true)
	be.Equal(t, len(gotTags), 1)
	be.Equal(t, gotTags[0].Name, "demo")

	e := echo.New()
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, APIBase+"/tags", nil)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)
	be.Err(t, TagsAPI(c, &sql.DB{}, true, true), nil)
	be.Equal(t, w.Code, http.StatusOK)
	be.True(t, strings.Contains(w.Body.String(), `"demo"`))
}

func TestRoleAPIAndGroupsCountValidation(t *testing.T) {
	t.Parallel()

	_, err := groupsCount(nil, nil)
	be.Err(t, err)

	be.Err(t, roleAPI(nil, nil, nil, postgres.Artist))

	e := echo.New()
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, APIBase+"/sceners", nil)
	w := httptest.NewRecorder()
	c := e.NewContext(r, w)
	err = roleAPI(logs.Discard(), c, new(sql.DB), postgres.Role("unknown"))
	be.Err(t, err, nil)
	be.Equal(t, w.Code, http.StatusBadRequest)
}

func TestDatePublishedAndTimer(t *testing.T) {
	t.Parallel()

	be.Equal(t, datePublished(nil), publishedAPI{})
	be.Equal(t, datePublished(&models.File{}), publishedAPI{})
	be.Equal(t, datePublished(&models.File{
		DateIssuedYear: null.Int16From(1995),
	}), publishedAPI{Year: 1995})
	be.Equal(t, datePublished(&models.File{
		DateIssuedYear:  null.Int16From(1995),
		DateIssuedMonth: null.Int16From(6),
	}), publishedAPI{Year: 1995, Month: 6})
	be.Equal(t, datePublished(&models.File{
		DateIssuedYear:  null.Int16From(1995),
		DateIssuedMonth: null.Int16From(6),
		DateIssuedDay:   null.Int16From(15),
	}), publishedAPI{Year: 1995, Month: 6, Day: 15})

	be.True(t, timedTimer(null.Time{}) == nil)
	now := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	got := timedTimer(null.TimeFrom(now))
	be.True(t, got != nil && got.Equal(now))
}

func TestArtifactHelpers(t *testing.T) {
	t.Parallel()

	be.Equal(t, artifact(nil).Summary.Filename, "")
	be.Equal(t, artifactSum(nil).Filename, "")
	be.Equal(t, artifactSummary(nil).Filename, "")
	be.True(t, releasersAPI(nil) == nil)
	be.True(t, relationshipsAPI(nil) == nil)
	be.True(t, relationsAPI(nil) == nil)
	be.True(t, linksAPI(nil) == nil)

	now := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)
	obf := helper.ObfuscateID(123)
	file := &models.File{
		ID:               123,
		UUID:             null.StringFrom("abc-123"),
		Filename:         null.StringFrom("demo.zip"),
		Filesize:         null.Int64From(4096),
		RecordTitle:      null.StringFrom("Cool Demo"),
		GroupBrandBy:     null.StringFrom("Razor 1911"),
		GroupBrandFor:    null.StringFrom("Fairlight"),
		Section:          null.StringFrom("demo"),
		Platform:         null.StringFrom("dos"),
		DateIssuedYear:   null.Int16From(1993),
		Createdat:        null.TimeFrom(now),
		FileLastModified: null.TimeFrom(now),
		Comment:          null.StringFrom("<p>hello</p>"),
		WebIDDemozoo:     null.Int64From(42),
		WebIDPouet:       null.Int64From(99),
		WebID16colors:    null.StringFrom("pack/test"),
		WebIDYoutube:     null.StringFrom("dQw4w9WgXcQ"),
		WebIDGithub:      null.StringFrom("Defacto2/server"),
		ListRelations:    null.StringFrom("Related;" + obf + "|Malformed|Bad;notobf"),
		ListLinks:        null.StringFrom("Site;example.com|HTTP;http://textfiles.com|Bad|NoHost;http://"),
	}

	sum := artifactSum(file)
	be.Equal(t, sum.Filename, "demo.zip")
	be.Equal(t, sum.Size.Bytes, int64(4096))
	be.True(t, sum.PostedDate != nil && sum.PostedDate.Equal(now))
	be.Equal(t, sum.URLs.Download, "/d/"+obf)

	summaries := artifactSummaries([]*models.File{nil, file})
	be.Equal(t, len(summaries), 2)
	be.Equal(t, summaries[0].Filename, "")
	be.Equal(t, summaries[1].Filename, "demo.zip")

	full := artifact(file)
	be.Equal(t, full.Summary.Filename, "demo.zip")
	be.Equal(t, full.ArtMeta.Comment, "hello")
	be.Equal(t, len(full.ArtMeta.Releasers), 2)
	be.Equal(t, full.ArtMeta.Releasers[0].ID, "fairlight")
	be.Equal(t, full.ArtMeta.Releasers[0].HTML, "/g/fairlight")
	be.Equal(t, full.ArtMeta.Releasers[1].ID, "razor-1911")
	be.Equal(t, full.ArtMeta.Releasers[1].HTML, "/g/razor-1911")
	// 1 valid relation + 2 valid links + 5 external IDs = 8 relationships.
	be.Equal(t, len(full.Relationships), 8)
	var foundPouet, foundYouTube bool
	for _, rel := range full.Relationships {
		if rel.Desc == "Pouet production" {
			foundPouet = true
			be.Equal(t, rel.Link, "https://www.pouet.net/prod.php?which=99")
		}
		if rel.Desc == "YouTube video" {
			foundYouTube = true
			be.Equal(t, rel.Link, "https://www.youtube.com/watch?v=dQw4w9WgXcQ")
		}
	}
	be.True(t, foundPouet)
	be.True(t, foundYouTube)

	// Single-releaser artifact must not include a trailing empty releaser entry.
	single := releasersAPI(&models.File{GroupBrandBy: null.StringFrom("Razor 1911")})
	be.Equal(t, len(single), 1)
	be.Equal(t, single[0].ID, "razor-1911")
	be.Equal(t, single[0].HTML, "/g/razor-1911")
	be.Equal(t, len(releasersAPI(&models.File{})), 0)

	be.True(t, relationsAPI(&models.File{}) == nil)
	be.True(t, linksAPI(&models.File{}) == nil)
}

package api

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/tags"
	"github.com/Defacto2/server/model"
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

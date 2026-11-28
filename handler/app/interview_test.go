package app_test

import (
	"net/url"
	"strings"
	"testing"

	"github.com/Defacto2/server/handler/app"
	"github.com/nalgeon/be"
)

func TestInterviewees(t *testing.T) {
	t.Parallel()

	groups := app.Interviewees()
	const expectedGroups = 12
	be.Equal(t, len(groups), expectedGroups)

	seenGroups := make(map[string]bool, len(groups))
	for _, group := range groups {
		be.True(t, group.Name != "")
		be.True(t, !seenGroups[group.Name])
		seenGroups[group.Name] = true
		be.True(t, len(group.Interviews) > 0)

		prevYear := 0
		prevMonth := 0
		for j, item := range group.Interviews {
			be.True(t, item.Scener != "")
			be.True(t, item.Content != "")
			be.True(t, item.Year >= 1980)
			be.True(t, item.Month >= 0 && item.Month <= 12)
			be.True(t, validInterviewLink(item.Link))

			if j > 0 {
				be.True(t, item.Year < prevYear || (item.Year == prevYear && item.Month <= prevMonth))
			}
			prevYear = item.Year
			prevMonth = item.Month
		}
	}
}

func TestInterviewCount(t *testing.T) {
	t.Parallel()

	be.Equal(t, app.InterviewCount(nil), 0)
	be.Equal(t, app.InterviewCount(app.Groups{}), 0)
	be.Equal(t, app.InterviewCount(app.Groups{
		{Name: "Empty"},
		{Name: "Two", Interviews: app.Interviews{{Scener: "A"}, {Scener: "B"}}},
		{Name: "One", Interviews: app.Interviews{{Scener: "C"}}},
	}), 3)

	const expectedTotal = 28
	be.Equal(t, app.InterviewCount(app.Interviewees()), expectedTotal)
}

func validInterviewLink(raw string) bool {
	if raw == "" || strings.TrimSpace(raw) != raw {
		return false
	}
	if strings.HasPrefix(raw, "/f/") {
		return len(raw) > len("/f/") && !strings.ContainsAny(raw, " \t\n")
	}
	parsed, err := url.ParseRequestURI(raw)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}

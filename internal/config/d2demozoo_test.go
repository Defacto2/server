package config_test

import (
	"strings"
	"testing"

	"github.com/Defacto2/server/internal/config"
	"github.com/nalgeon/be"
)

func TestDemozooAuth(t *testing.T) {
	t.Parallel()

	var empty config.DemozooAuth
	be.Equal(t, "", empty.LogValue().String())
	be.True(t, strings.Contains(empty.Help(), "demozoo.org"))
	be.Equal(t, "", empty.Issue())

	short := config.DemozooAuth("12345678")
	be.Equal(t, "xxxxx", short.LogValue().String())
	be.Equal(t, "", short.Help())
	be.True(t, strings.Contains(short.Issue(), "too short"))

	noDot := config.DemozooAuth("1234567890")
	be.Equal(t, "", noDot.Help())
	be.True(t, strings.Contains(noDot.Issue(), "period"))

	valid := config.DemozooAuth("12345678.abcdefgh")
	be.Equal(t, "xxxxx", valid.LogValue().String())
	be.Equal(t, "", valid.Help())
	be.Equal(t, "", valid.Issue())
}

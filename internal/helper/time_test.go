package helper_test

import (
	"testing"
	"time"

	"github.com/Defacto2/server/internal/helper"
	"github.com/stretchr/testify/assert"
)

func TestLatency(t *testing.T) {
	result := helper.Latency()
	now := time.Now()
	assert.Less(t, *result, now)
}

func TestTimeDistance(t *testing.T) {
	now := time.Now()
	s := helper.TimeDistance(now, now, false)
	assert.Equal(t, "less than a minute", s)
	s = helper.TimeDistance(now, now.Add(time.Minute+time.Second), false)
	assert.Equal(t, "1 minute", s)
	s = helper.TimeDistance(now, now.Add(time.Second*2), true)
	assert.Equal(t, "less than 5 seconds", s)
	s = helper.TimeDistance(now, now.Add(time.Second*9), true)
	assert.Equal(t, "less than 10 seconds", s)
	s = helper.TimeDistance(now, now.Add(time.Second*19), true)
	assert.Equal(t, "less than 20 seconds", s)
	s = helper.TimeDistance(now, now.Add(time.Second*35), true)
	assert.Equal(t, "half a minute", s)
	s = helper.TimeDistance(now, now.Add(time.Second*60), true)
	assert.Equal(t, "1 minute", s)
	s = helper.TimeDistance(now, now.Add(time.Hour), true)
	assert.Equal(t, "about 1 hour", s)
	s = helper.TimeDistance(now, now.Add(time.Hour*24), true)
	assert.Equal(t, "1 day", s)
	s = helper.TimeDistance(now, now.Add(time.Hour*24*2), true)
	assert.Equal(t, "2 days", s)
	s = helper.TimeDistance(now, now.Add(time.Hour*24*30), true)
	assert.Equal(t, "about 1 month", s)
	s = helper.TimeDistance(now, now.Add(time.Hour*24*365), true)
	assert.Equal(t, "about 1 year", s)
	s = helper.TimeDistance(now, now.Add(time.Hour*24*500), true)
	assert.Equal(t, "over 1 year", s)
	s = helper.TimeDistance(now, now.Add(time.Hour*24*700), true)
	assert.Equal(t, "almost 2 years", s)
	s = helper.TimeDistance(now, now.Add(time.Hour*24*365*10), true)
	assert.Equal(t, "10 years", s)
}

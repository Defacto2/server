package helper

// Package time.go contains functions that return a time.Time value.

import (
	"fmt"
	"math/big"
	"time"
)

// Latency returns the stored, current local time.
func Latency() *time.Time {
	start := time.Now()
	r := new(big.Int)
	const n, k = 1000, 10
	r.Binomial(n, k)
	return &start
}

// TimeDistance describes the difference between two time values.
// The seconds parameter determines if the string should include seconds.
func TimeDistance(from, to time.Time, seconds bool) string {
	// This function is a port of a CFWheels framework function programmed in ColdFusion (CFML).
	// https://github.com/cfwheels/cfwheels/blob/cf8e6da4b9a216b642862e7205345dd5fca34b54/wheels/global/misc.cfm#L112

	delta := to.Sub(from)
	secs, mins, hrs := int(delta.Seconds()),
		int(delta.Minutes()),
		int(delta.Hours())

	const hours, days, months, year, years, twoyears = 1440, 43200, 525600, 657000, 919800, 1051200
	switch {
	case mins <= 1:
		if !seconds {
			return lessMin(secs)
		}
		return lessMinAsSec(secs)
	case mins < hours:
		return lessHours(mins, hrs)
	case mins < days:
		return lessDays(mins, hrs)
	case mins < months:
		return lessMonths(mins, hrs)
	case mins < year:
		return "about 1 year"
	case mins < years:
		return "over 1 year"
	case mins < twoyears:
		return "almost 2 years"
	default:
		y := mins / months

		return fmt.Sprintf("%d years", y)
	}
}
func lessMin(secs int) string {
	const minute = 60
	switch {
	case secs < minute:
		return "less than a minute"
	default:
		return "1 minute"
	}
}

func lessMinAsSec(secs int) string {
	const five, ten, twenty, forty = 5, 10, 20, 40
	switch {
	case secs < five:
		return "less than 5 seconds"
	case secs < ten:
		return "less than 10 seconds"
	case secs < twenty:
		return "less than 20 seconds"
	case secs < forty:
		return "half a minute"
	default:
		return "1 minute"
	}
}

func lessHours(mins, hrs int) string {
	const parthour, abouthour, hours = 45, 90, 1440

	switch {
	case mins < parthour:
		return fmt.Sprintf("%d minutes", mins)
	case mins < abouthour:
		return "about 1 hour"
	case mins < hours:
		return fmt.Sprintf("about %d hours", hrs)
	default:
		return ""
	}
}

func lessDays(mins, hrs int) string {
	const day, days = 2880, 43200
	switch {
	case mins < day:
		return "1 day"
	case mins < days:
		const hoursinaday = 24
		d := hrs / hoursinaday
		return fmt.Sprintf("%d days", d)
	default:
		return ""
	}
}

func lessMonths(mins, hrs int) string {
	const month, months = 86400, 525600
	switch {
	case mins < month:
		return "about 1 month"
	case mins < months:
		const hoursinamonth = 730
		m := hrs / hoursinamonth
		return fmt.Sprintf("%d months", m)
	default:
		return ""
	}
}

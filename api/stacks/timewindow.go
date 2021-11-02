package stacks

import (
	"time"

	"github.com/pkg/errors"
)

type timeWindow struct {
	start time.Time
	end   time.Time
}

func NewTimeWindow(start, end string) (*timeWindow, error) {
	startTime, err := time.Parse("15:04", start)
	if err != nil {
		return nil, errors.WithMessagef(err, "incorrect start time format")
	}

	endTime, err := time.Parse("15:04", end)
	if err != nil {
		return nil, errors.WithMessagef(err, "incorrect start time format")
	}

	return &timeWindow{start: startTime, end: endTime}, nil
}

// Within returns true if the provided time t is within the window (ignoring the date) using these rules
// start <= t < end
func (tw *timeWindow) Within(t time.Time) bool {
	// Convert times to minutes of the day
	var dayminutes = func(h, m int) int {
		return h*60 + m
	}

	startMins := dayminutes(tw.start.Hour(), tw.start.Minute())
	endMins := dayminutes(tw.end.Hour(), tw.end.Minute())
	tMins := dayminutes(t.Hour(), t.Minute())

	// Special case where the end time falls into the next day or
	// the start and end time are the same. Add 24hrs in minutes
	if endMins < startMins || startMins == endMins {
		if tMins < endMins {
			// and where the time to be tested falls into the next day
			tMins += dayminutes(24, 0)
		}

		endMins += dayminutes(24, 0)
	}

	if tMins >= startMins && tMins < endMins {
		return true
	}

	return false
}

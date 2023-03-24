package algo

import (
	"math"
	"time"
)

type Anki struct {
	now func() time.Time
}

func NewAnki(now func() time.Time) *Anki {
	return &Anki{now: now}
}

func (a Anki) CalculateNextDueDate(performance uint32, correctAnswers uint32) time.Time {
	next := a.now().
		UTC().
		Add(
			a.calculateShift(
				float64(performance),
				float64(correctAnswers),
			),
		).Truncate(24 * time.Hour)

	if next.Before(a.now()) {
		return a.now().
			Add(24 * time.Hour).
			Truncate(24 * time.Hour)
	} else {
		return next
	}
}

func (Anki) calculateShift(performance float64, correctAnswers float64) time.Duration {
	if correctAnswers == 0 {
		return durationFromDays(1)
	}

	difficulty := -0.8 + 0.28*float64(performance) + 0.02*math.Pow(performance, 2)

	return durationFromDays(6 * math.Pow(difficulty, correctAnswers-1))
}

func durationFromDays(days float64) time.Duration {
	return time.Duration(days) * 24 * time.Hour
}

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

const day = 24 * time.Hour

func (a Anki) CalculateNextDueDate(performance uint32, consecutiveCorrectAnswersNumber uint32) time.Time {
	next := a.now().
		UTC().
		Add(
			a.calculateShift(
				float64(performance),
				float64(consecutiveCorrectAnswersNumber),
			),
		).Truncate(day)

	if next.Before(a.now()) {
		return a.now().
			Add(day).
			Truncate(day)
	}
	return next
}

//nolint:mnd // magic numbers are used for calculations
func (Anki) calculateShift(performance float64, correctAnswers float64) time.Duration {
	if correctAnswers <= 1 {
		return durationFromDays(1)
	}

	difficulty := -0.8 + 0.28*performance + 0.02*performance*performance

	return durationFromDays(6 * math.Pow(difficulty, correctAnswers-1))
}

func durationFromDays(days float64) time.Duration {
	return time.Duration(days) * day
}

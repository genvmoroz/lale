package algo

import (
	"reflect"
	"testing"
	"time"
)

func TestAnkiCalculateNextDueDate(t *testing.T) {
	t.Parallel()

	type (
		field struct {
			now func() time.Time
		}
		input struct {
			performance    uint32
			correctAnswers uint32
		}
		want struct {
			nextDueDate time.Time
		}
	)
	testcases := map[string]struct {
		field field
		input input
		want  want
	}{
		"incorrect answer": {
			field: field{now: testNow},
			input: input{performance: 3, correctAnswers: 0},
			want:  want{nextDueDate: testNowTime.Add(24 * time.Hour).Truncate(24 * time.Hour)},
		},
		"1 correct answer, 1 performance": {
			field: field{now: testNow},
			input: input{performance: 1, correctAnswers: 1},
			want:  want{nextDueDate: testNowTime.Add(6 * 24 * time.Hour).Truncate(24 * time.Hour)},
		},
		"1 correct answer, 2 performance": {
			field: field{now: testNow},
			input: input{performance: 2, correctAnswers: 1},
			want:  want{nextDueDate: testNowTime.Add(6 * 24 * time.Hour).Truncate(24 * time.Hour)},
		},
		"1 correct answer, 3 performance": {
			field: field{now: testNow},
			input: input{performance: 3, correctAnswers: 1},
			want:  want{nextDueDate: testNowTime.Add(6 * 24 * time.Hour).Truncate(24 * time.Hour)},
		},
		"1 correct answer, 4 performance": {
			field: field{now: testNow},
			input: input{performance: 4, correctAnswers: 1},
			want:  want{nextDueDate: testNowTime.Add(6 * 24 * time.Hour).Truncate(24 * time.Hour)},
		},
		"1 correct answer, 5 performance": {
			field: field{now: testNow},
			input: input{performance: 5, correctAnswers: 1},
			want:  want{nextDueDate: testNowTime.Add(6 * 24 * time.Hour).Truncate(24 * time.Hour)},
		},
		"2 correct answer, 1 performance": {
			field: field{now: testNow},
			input: input{performance: 1, correctAnswers: 2},
			want:  want{nextDueDate: testNowTime.Add(-3 * 24 * time.Hour).Truncate(24 * time.Hour)},
		},
		"2 correct answer, 2 performance": {
			field: field{now: testNow},
			input: input{performance: 2, correctAnswers: 2},
			want:  want{nextDueDate: testNowTime.Truncate(24 * time.Hour)},
		},
		"0 correct answer, 2 performance": {
			field: field{now: testNow},
			input: input{performance: 2, correctAnswers: 0},
			want:  want{nextDueDate: testNowTime.Add(24 * time.Hour).Truncate(24 * time.Hour)},
		},
		"2 correct answer, 3 performance": {
			field: field{now: testNow},
			input: input{performance: 3, correctAnswers: 2},
			want:  want{nextDueDate: testNowTime.Add(24 * time.Hour).Truncate(24 * time.Hour)},
		},
		"2 correct answer, 4 performance": {
			field: field{now: testNow},
			input: input{performance: 4, correctAnswers: 2},
			want:  want{nextDueDate: testNowTime.Add(3 * 24 * time.Hour).Truncate(24 * time.Hour)},
		},
		"10 correct answer, 3 performance": {
			field: field{now: testNow},
			input: input{performance: 5, correctAnswers: 10},
			want:  want{nextDueDate: testNowTime.Add(14 * 24 * time.Hour).Truncate(24 * time.Hour)},
		},
	}
	for name, testcase := range testcases {
		name := name
		testcase := testcase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			a := NewAnki(testcase.field.now)

			got := a.CalculateNextDueDate(testcase.input.performance, testcase.input.correctAnswers)
			if !reflect.DeepEqual(got, testcase.want.nextDueDate) {
				t.Fatalf("CalculateNextDueDate() = %v, want %v", got, testcase.want.nextDueDate)
			}
		})
	}
}

var testNowTime = time.Now().UTC()

func testNow() time.Time {
	return testNowTime
}

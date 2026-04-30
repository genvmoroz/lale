package entity_test

import (
	"testing"
	"time"

	"github.com/genvmoroz/lale/service/pkg/entity"
)

func TestCard_NeedToLearn(t *testing.T) {
	t.Parallel()

	tnow := time.Now().UTC()

	tests := []struct {
		name string
		card entity.Card
		want bool
	}{
		{
			name: "new card",
			card: entity.Card{},
			want: true,
		},
		{
			name: "scheduled but not learnt",
			card: entity.Card{NextDueDate: tnow.Add(time.Hour)},
			want: false,
		},
		{
			name: "learnt with zero due date",
			card: entity.Card{Learnt: true},
			want: false,
		},
		{
			name: "learnt with due date set",
			card: entity.Card{Learnt: true, NextDueDate: tnow.Add(time.Hour)},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.card.NeedToLearn(); got != tt.want {
				t.Fatalf("NeedToLearn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCard_NeedToRepeat(t *testing.T) {
	t.Parallel()

	tnow := time.Now().UTC()

	tests := []struct {
		name string
		card entity.Card
		want bool
	}{
		{
			name: "new card",
			card: entity.Card{},
			want: false,
		},
		{
			name: "due for repeat",
			card: entity.Card{NextDueDate: tnow.Add(-time.Hour)},
			want: true,
		},
		{
			name: "not yet due",
			card: entity.Card{NextDueDate: tnow.Add(time.Hour)},
			want: false,
		},
		{
			name: "due but learnt",
			card: entity.Card{Learnt: true, NextDueDate: tnow.Add(-time.Hour)},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.card.NeedToRepeat(); got != tt.want {
				t.Fatalf("NeedToRepeat() = %v, want %v", got, tt.want)
			}
		})
	}
}

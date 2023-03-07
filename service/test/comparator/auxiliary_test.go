package comparator

import (
	"testing"

	"github.com/genvmoroz/lale/service/pkg/entity"
)

func TestContainDuplicatesByID(t *testing.T) {
	t.Parallel()

	type args struct {
		cards []entity.Card
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "without duplicates",
			args: args{cards: []entity.Card{{ID: "0"}, {ID: "1"}, {ID: "2"}, {ID: "3"}, {ID: "4"}, {ID: "5"}}},
			want: false,
		},
		{
			name: "with duplicates",
			args: args{cards: []entity.Card{{ID: "0"}, {ID: "1"}, {ID: "2"}, {ID: "0"}, {ID: "4"}, {ID: "5"}}},
			want: true,
		},
		{
			name: "empty",
			args: args{cards: []entity.Card{}},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := ContainDuplicatesByID(tt.args.cards); got != tt.want {
				t.Errorf("ContainDuplicatesByID() = %v, want %v", got, tt.want)
			}
		})
	}
}

package comparator

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/genvmoroz/lale/service/api"
	"github.com/genvmoroz/lale/service/internal/entity"
)

func ContainDuplicatesByID(cards []entity.Card) bool {
	for i := 0; i < len(cards); i++ {
		for j := i + 1; j < len(cards); j++ {
			if cards[i].ID == cards[j].ID {
				return true
			}
		}
	}
	return false
}

func Compare[T any, R any](t []T, r []R, equal func(t T, r R) bool) bool {
	if len(t) != len(r) {
		return false
	}

	for i, v := range t {
		if !equal(v, r[i]) {
			return false
		}
	}

	return true
}

func stringEqual(v0, v1 string) bool {
	return v0 == v1
}

func phoneticEqual(v0 entity.Phonetic, v1 *api.Phonetic) bool {
	return empty.ComparePhonetic(&v0, v1)
}

func meaningEqual(v0 entity.Meaning, v1 *api.Meaning) bool {
	return empty.CompareMeaning(&v0, v1)
}

func wordInformationEqual(v0 entity.WordInformation, v1 *api.WordInformation) bool {
	return empty.CompareWordInformation(&v0, v1)
}

func definitionEqual(v0 entity.Definition, v1 *api.Definition) bool {
	return empty.CompareDefinition(&v0, v1)
}

func timestampEqual(v0 time.Time, v1 *timestamppb.Timestamp) bool {
	if v1 == nil {
		v1 = timestamppb.New(time.Time{})
	}

	return v1.AsTime() == v0
}

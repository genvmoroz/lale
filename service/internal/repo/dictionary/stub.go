package dictionary

import (
	"github.com/genvmoroz/lale/service/pkg/entity"
	"golang.org/x/text/language"
)

type Stub struct{}

func NewStub() *Stub {
	return &Stub{}
}

func (s *Stub) GetWordInformation(word string, _ language.Tag) (entity.WordInformation, error) {
	return entity.WordInformation{Word: word}, nil
}

package transform

import (
	"github.com/genvmoroz/lale/service/api"
	"github.com/genvmoroz/lale/service/pkg/entity"
	"github.com/genvmoroz/lale/service/pkg/lang"
)

type (
	Transformer interface {
		ToCoreWordInformationList(list []*api.WordInformation) []entity.WordInformation
	}

	transformer struct{}
)

var DefaultTransformer Transformer = transformer{}

func (t transformer) ToCoreWordInformationList(list []*api.WordInformation) []entity.WordInformation {
	if len(list) == 0 {
		return nil
	}

	res := make([]entity.WordInformation, 0, len(list))
	for _, w := range list {
		if w != nil {
			res = append(res, t.toCoreWordInformation(w))
		}
	}

	return res
}

func (transformer) toCoreTranslation(t *api.Translation) *entity.Translation {
	if t == nil {
		return nil
	}

	return &entity.Translation{
		Language:     lang.Language(t.Language),
		Translations: t.Translations,
	}
}

func (t transformer) toCoreWordInformation(info *api.WordInformation) entity.WordInformation {
	return entity.WordInformation{
		Word:        info.Word,
		Translation: t.toCoreTranslation(info.Translation),
		Origin:      info.Origin,
		Phonetics:   t.toCorePhonetics(info.Phonetics),
		Meanings:    t.toCoreMeanings(info.Meanings),
		Sentences:   info.Sentences,
	}
}

func (t transformer) toCoreMeanings(meanings []*api.Meaning) []entity.Meaning {
	if len(meanings) == 0 {
		return nil
	}

	res := make([]entity.Meaning, 0, len(meanings))
	for _, m := range meanings {
		if m != nil {
			res = append(res, t.toCoreMeaning(m))
		}
	}

	return res
}

func (t transformer) toCoreMeaning(m *api.Meaning) entity.Meaning {
	return entity.Meaning{
		PartOfSpeech: m.PartOfSpeech,
		Definitions:  t.toCoreDefinitions(m.Definitions),
	}
}

func (t transformer) toCoreDefinitions(definitions []*api.Definition) []entity.Definition {
	if len(definitions) == 0 {
		return nil
	}

	res := make([]entity.Definition, 0, len(definitions))
	for _, d := range definitions {
		if d != nil {
			res = append(res, t.toCoreDefinition(d))
		}
	}

	return res
}

func (transformer) toCoreDefinition(d *api.Definition) entity.Definition {
	return entity.Definition{
		Definition: d.Definition,
		Example:    d.Example,
		Synonyms:   d.Synonyms,
		Antonyms:   d.Antonyms,
	}
}

func (t transformer) toCorePhonetics(phonetics []*api.Phonetic) []entity.Phonetic {
	if len(phonetics) == 0 {
		return nil
	}

	res := make([]entity.Phonetic, 0, len(phonetics))
	for _, p := range phonetics {
		if p != nil {
			res = append(res, t.toCorePhonetic(p))
		}
	}

	return res
}

func (transformer) toCorePhonetic(p *api.Phonetic) entity.Phonetic {
	return entity.Phonetic{
		Text:      p.Text,
		AudioLink: p.AudioLink,
	}
}

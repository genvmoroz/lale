package comparator

import (
	"github.com/genvmoroz/lale/service/api"
	"github.com/genvmoroz/lale/service/internal/entity"
)

type GRPCComparator struct {
}

var empty = &GRPCComparator{}

func NewGRPCComparator() *GRPCComparator { return empty }

func (c *GRPCComparator) CompareCard(card *entity.Card, target *api.Card) bool {
	if card == nil {
		return target == nil
	}

	if target != nil {
		return card.ID == target.Id &&
			card.UserID == target.UserID &&
			string(card.Language) == target.Language &&
			Compare(card.WordInformationList, target.WordInformationList, wordInformationEqual) &&
			timestampEqual(card.NextDueDate, target.NextDueDate)
	}

	return false
}

func (c *GRPCComparator) CompareWordInformation(word *entity.WordInformation, target *api.WordInformation) bool {
	if word == nil {
		return target == nil
	}
	if target != nil {
		return word.Word == target.Word &&
			c.CompareTranslation(word.Translation, target.Translation) &&
			word.Origin == target.Origin &&
			Compare(word.Phonetics, target.Phonetics, phoneticEqual) &&
			Compare(word.Meanings, target.Meanings, meaningEqual)
	}

	return false
}

func (*GRPCComparator) ComparePhonetic(phonetic *entity.Phonetic, target *api.Phonetic) bool {
	if phonetic == nil {
		return target == nil
	}
	if target != nil {
		return phonetic.Text == target.Text &&
			phonetic.AudioLink == target.AudioLink
	}

	return false
}

func (*GRPCComparator) CompareMeaning(meaning *entity.Meaning, target *api.Meaning) bool {
	if meaning == nil {
		return target == nil
	}
	if target != nil {
		return meaning.PartOfSpeech == target.PartOfSpeech &&
			Compare(meaning.Definitions, target.Definitions, definitionEqual)
	}

	return false
}

func (*GRPCComparator) CompareDefinition(definition *entity.Definition, target *api.Definition) bool {
	if definition == nil {
		return target == nil
	}
	if target != nil {
		return definition.Definition == target.Definition &&
			definition.Example == target.Example &&
			Compare(definition.Synonyms, target.Synonyms, stringEqual) &&
			Compare(definition.Antonyms, target.Antonyms, stringEqual)
	}

	return false
}

func (*GRPCComparator) CompareTranslation(Translation *entity.Translation, target *api.Translation) bool {
	if Translation == nil {
		return target == nil
	}
	if target != nil {
		return Translation.Language.EqualString(target.Language) &&
			Compare(Translation.Translations, target.Translations, stringEqual)
	}

	return false
}

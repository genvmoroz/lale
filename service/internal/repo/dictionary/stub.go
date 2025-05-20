package dictionary

import (
	"github.com/genvmoroz/lale/service/pkg/entity"
	"golang.org/x/text/language"
)

type Stub struct{}

func NewStub() *Stub {
	return &Stub{}
}

func (c *Stub) GetWordInformation(word string, lang language.Tag) (entity.WordInformation, error) {
	return entity.WordInformation{
		Word: word,
		Translation: &entity.Translation{
			Language: lang,
			Translations: []string{
				"test translation 1",
				"test translation 2",
				"test translation 3",
				"test translation 4",
			},
		},
		Origin: "test origin",
		Phonetics: []entity.Phonetic{
			{Text: "test phonetic 1"},
			{Text: "test phonetic 2"},
			{Text: "test phonetic 3"},
		},
		Meanings: []entity.Meaning{
			{
				PartOfSpeech: "test part of speech",
				Definitions: []entity.Definition{
					{
						Definition: "test definition 1",
						Example:    "test example 1",
						Synonyms:   []string{"test synonym 1", "test synonym 2", "test synonym 3"},
						Antonyms:   []string{"test antonym 1", "test antonym 2", "test antonym 3"},
					},
					{
						Definition: "test definition 2",
						Example:    "test example 2",
						Synonyms:   []string{"test synonym 4", "test synonym 5", "test synonym 6"},
						Antonyms:   []string{"test antonym 4", "test antonym 5", "test antonym 6"},
					},
					{
						Definition: "test definition 3",
						Example:    "test example 3",
						Synonyms:   []string{"test synonym 7", "test synonym 8", "test synonym 9"},
						Antonyms:   []string{"test antonym 7", "test antonym 8", "test antonym 9"},
					},
					{
						Definition: "test definition 4",
						Example:    "test example 4",
						Synonyms:   []string{"test synonym 10", "test synonym 11", "test synonym 12"},
						Antonyms:   []string{"test antonym 10", "test antonym 11", "test antonym 12"},
					},
				},
			},
		},
	}, nil
}

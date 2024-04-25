package pretty

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/genvmoroz/lale/service/api"
	"github.com/genvmoroz/lale/tg-client/internal/transform"
	"gopkg.in/yaml.v3"
)

func Card(card *api.Card, withWords bool) []string {
	if card == nil {
		return nil
	}

	var p []string

	meta := `
CardID: <code>%s</code>
UserID: <code>%s</code>
Language: <code>%s</code>
NextDueDate: <code>%s</code>
ConsecutiveCorrectAnswersNumber: <code>%s</code>
`
	p = append(p,
		fmt.Sprintf(
			meta,
			card.GetId(),
			card.GetUserID(),
			card.GetLanguage(),
			card.GetNextDueDate().AsTime().Format(time.RFC3339),
			strconv.Itoa(int(card.GetConsecutiveCorrectAnswersNumber())),
		),
	)

	if !withWords {
		return p
	}

	for _, word := range transform.DefaultTransformer.ToCoreWordInformationList(card.GetWordInformationList()) {
		out, err := yaml.Marshal(word)
		if err != nil {
			log.Fatalf("unexpected err: %s", err.Error())
		}
		p = append(p, string(out))
	}

	return p
}

func WordInformation(wordInfo *api.WordInformation) string {
	if wordInfo == nil {
		return ""
	}
	out, err := yaml.Marshal(transform.DefaultTransformer.ToCoreWordInformation(wordInfo))
	if err != nil {
		log.Fatalf("unexpected err: %s", err.Error())
	}

	return string(out)
}

func Translation(t *api.Translation) string {
	if t == nil {
		return ""
	}

	pattern := `
Language: <code>%s</code>
%s
`
	tmp := struct {
		Translations []string `yaml:"Translations"`
	}{
		Translations: t.GetTranslations(),
	}
	out, err := yaml.Marshal(tmp)
	if err != nil {
		log.Fatalf("unexpected err: %s", err.Error())
	}
	return fmt.Sprintf(pattern, t.GetLanguage(), string(out))
}

func Meaning(m *api.Meaning) string {
	if m == nil {
		return ""
	}
	out, err := yaml.Marshal(transform.DefaultTransformer.ToCoreMeaning(m))
	if err != nil {
		log.Fatalf("unexpected err: %s", err.Error())
	}

	return string(out)
}

func MeaningWithoutExamples(m *api.Meaning) string {
	if m == nil {
		return ""
	}

	copyM := &api.Meaning{}
	copyM.PartOfSpeech = m.GetPartOfSpeech()

	for _, def := range m.GetDefinitions() {
		if def == nil {
			continue
		}
		copyD := &api.Definition{
			Definition: def.GetDefinition(),
			Example:    "", // without example
			Synonyms:   def.GetSynonyms(),
			Antonyms:   def.GetAntonyms(),
		}
		copyM.Definitions = append(copyM.Definitions, copyD)
	}

	return Meaning(copyM)
}

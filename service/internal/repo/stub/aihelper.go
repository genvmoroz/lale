package stub

import (
	"fmt"
	"strings"

	"golang.org/x/text/language"
)

type AIHelper struct{}

func (repo *AIHelper) GenerateSentences(word string, count int) ([]string, error) {
	sentences := make([]string, count)
	for i := 0; i < count; i++ {
		sentences[i] = fmt.Sprintf("№%d. This is a test sentence with word %s.", i+1, word)
	}

	return sentences, nil
}

func (repo *AIHelper) GetFamilyWordsWithTranslation(word string, _ language.Tag) (map[string]string, error) {
	familyWords := make(map[string]string)
	familyWords[word] = "This is a test translation for word " + word

	return familyWords, nil
}

func (repo *AIHelper) GenStory(words []string, _ language.Tag) (string, error) {
	return "This is a test story with words: " + strings.Join(words, ", "), nil
}

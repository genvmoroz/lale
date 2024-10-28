package stub

import "golang.org/x/text/language"

type AIHelper struct{}

func (repo *AIHelper) GenerateSentences(_ string, _ uint32) ([]string, error) {
	return nil, nil
}

func (repo *AIHelper) GetFamilyWordsWithTranslation(_ string, _ language.Tag) (map[string]string, error) {
	return nil, nil
}

func (repo *AIHelper) GenStory(_ []string, _ language.Tag) (string, error) {
	return "", nil
}

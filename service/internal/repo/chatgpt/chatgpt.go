//nolint:gomnd,all // magic numbers are fine here
package chatgpt

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	goclient "github.com/amarnathcjd/chatgpt"
	"golang.org/x/text/language"
)

type (
	Client interface {
		Start() error
		Ask(ctx context.Context, prompt string, askOpts ...goclient.AskOpts) (*goclient.ChatResponse, error)
	}

	Repo struct {
		client Client
	}
)

func NewRepo(token string) (*Repo, error) {
	client := goclient.NewClient(&goclient.Config{ApiKey: token})
	if err := client.Start(); err != nil {
		return nil, err
	}

	return &Repo{client: client}, nil
}

var _ Client = &goclient.Client{}

type EnglishComplexity string

const (
	Unknown      EnglishComplexity = "Unknown"
	Beginner     EnglishComplexity = "Beginner"
	Intermediate EnglishComplexity = "Intermediate"
	Advanced     EnglishComplexity = "Advanced"
)

func (r *Repo) GenerateSentences(ctx context.Context, word string, size uint32) ([]string, error) {
	return r.generateSentencesWithRetry(ctx, word, size, Intermediate, 4)
}

func (r *Repo) generateSentencesWithRetry(ctx context.Context, word string, size uint32, complexity EnglishComplexity, retries uint) ([]string, error) {
	var (
		sents []string
		err   error
	)
	for retry := 0; retry <= int(retries); retry++ {
		sents, err = r.generateSentences(ctx, word, size, complexity)
		if err == nil {
			return sents, nil
		}
		time.Sleep(5 * time.Second)
	}

	return nil, err
}

func (r *Repo) generateSentences(ctx context.Context, word string, size uint32, complexity EnglishComplexity) ([]string, error) {
	if !utf8.ValidString(word) {
		return nil, fmt.Errorf("invalid utf8 string: %s", word)
	}
	if len(strings.TrimSpace(word)) == 0 {
		return nil, errors.New("word is blank")
	}

	req := fmt.Sprintf(
		"Generate %d random sentences with the word \"%s\" for %s English level with any topics.",
		size, strings.TrimSpace(word), complexity,
	)

	resp, err := r.client.Ask(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	if resp == nil || len(resp.Message) == 0 {
		return nil, errors.New("connection successful but response is empty")
	}

	content := strings.Split(resp.Message, "\n")
	var sentences []string

	for index := 0; index < len(content); index++ {
		rq := regexp.MustCompile(`^\d.`)
		ss := rq.ReplaceAll([]byte(content[index]), []byte{})
		sentence := strings.TrimSpace(strings.TrimRight(string(ss), "."))
		if len(sentence) != 0 {
			sentences = append(sentences, sentence)
		}
	}
	if len(sentences) > int(size) {
		return sentences[:size], nil
	}
	return sentences, nil
}

func (r *Repo) GetFamilyWordsWithTranslation(_ string, _ language.Tag) (map[string]string, error) {
	return nil, nil
}

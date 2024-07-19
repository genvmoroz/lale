//nolint:gomnd,lll,all // magic numbers are fine here
package openai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	basehttp "net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	http "github.com/hashicorp/go-retryablehttp"
	"golang.org/x/text/language"
)

type (
	Config struct {
		Addr    string        `envconfig:"APP_OPENAI_ADDR" default:"https://api.openai.com/v1/chat/completions" json:"addr,omitempty"`
		Token   string        `envconfig:"APP_OPENAI_TOKEN" required:"true" json:"token,omitempty"`
		Retries uint          `envconfig:"APP_OPENAI_RETRIES" default:"3" json:"retries,omitempty"`
		Timeout time.Duration `envconfig:"APP_OPENAI_TIMEOUT" default:"3s" json:"timeout,omitempty"`
	}

	Scraper struct {
		addr   string
		client *basehttp.Client
		token  string
	}
)

func NewHelper(cfg Config) (*Scraper, error) {
	_, err := url.ParseRequestURI(cfg.Addr)
	if err != nil {
		return nil, fmt.Errorf("parse addr [%s]: %w", cfg.Addr, err)
	}
	if cfg.Timeout < 0 {
		return nil, fmt.Errorf("timeout shouldn't be negative [%d]: %w", cfg.Timeout, err)
	}

	client := http.NewClient()
	client.RetryMax = int(cfg.Retries)
	client.Backoff = http.DefaultBackoff
	baseClient := client.StandardClient()
	baseClient.Timeout = cfg.Timeout

	scr := &Scraper{
		client: baseClient,
		addr:   cfg.Addr,
		token:  cfg.Token,
	}

	return scr, nil
}

type EnglishComplexity int32

const (
	Unknown EnglishComplexity = iota
	Beginner
	Intermediate
	Advanced
)

func (s *Scraper) GenerateSentences(word string, size uint32) ([]string, error) {
	return s.generateSentencesWithRetry(word, size, Intermediate, 4)
}

func (s *Scraper) generateSentencesWithRetry(word string, size uint32, complexity EnglishComplexity, retries uint) ([]string, error) {
	var (
		sents []string
		err   error
	)
	for retry := 0; retry <= int(retries); retry++ {
		sents, err = s.generateSentences(word, size, complexity)
		if err == nil {
			return sents, nil
		}
		time.Sleep(5 * time.Second)
	}

	return nil, err
}

func (s *Scraper) generateSentences(word string, size uint32, complexity EnglishComplexity) ([]string, error) {
	if !utf8.ValidString(word) {
		return nil, fmt.Errorf("invalid utf8 string: %s", word)
	}
	if len(strings.TrimSpace(word)) == 0 {
		return nil, errors.New("word is blank")
	}
	if !complexity.IsValid() {
		return nil, fmt.Errorf("invalid complexity: %d", int32(complexity))
	}

	body, err := s.prepareRequestBody(
		fmt.Sprintf(
			"Generate %d random sentences with the word \"%s\" in different meanings of the word for %s English level with any topics. Print only sentences.",
			size,
			strings.TrimSpace(word),
			complexity.String(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("prepare request body: %w", err)
	}

	req, err := basehttp.NewRequest(basehttp.MethodPost, s.addr, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	s.authorizeReq(req)

	resp, err := s.client.Do(req)
	defer func() {
		if resp != nil {
			if closeErr := resp.Body.Close(); closeErr != nil {
				log.Printf("close response body: %s", closeErr.Error())
			}
		}
	}()
	if err != nil {
		return nil, fmt.Errorf("executing request error: %w", err)
	}

	if resp.StatusCode != basehttp.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	var parsedResponse response
	if err = json.Unmarshal(respBody, &parsedResponse); err != nil {
		return nil, fmt.Errorf("parse response body: %w", err)
	}

	if len(parsedResponse.Choices) == 0 {
		return nil, errors.New("connection successful but response is empty")
	}

	content := strings.Split(parsedResponse.Choices[0].Message.Content, "\n")
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

func (s *Scraper) GetFamilyWordsWithTranslation(word string, lang language.Tag) (map[string]string, error) {
	return s.getFamilyWordsWithTranslationWithRetry(word, lang, 4)
}

func (s *Scraper) getFamilyWordsWithTranslationWithRetry(word string, lang language.Tag, retries uint) (map[string]string, error) {
	var (
		words map[string]string
		err   error
	)
	for retry := 0; retry <= int(retries); retry++ {
		words, err = s.getFamilyWordsWithTranslation(word, lang)
		if err == nil {
			return words, nil
		}
		time.Sleep(5 * time.Second)
	}

	return nil, err
}

func (s *Scraper) getFamilyWordsWithTranslation(word string, lang language.Tag) (map[string]string, error) {
	if !utf8.ValidString(word) {
		return nil, fmt.Errorf("invalid utf8 string: %s", word)
	}
	if len(strings.TrimSpace(word)) == 0 {
		return nil, errors.New("word is blank")
	}

	base, _ := lang.Base()

	body, err := s.prepareRequestBody(
		fmt.Sprintf(
			"Write all words which are in the one family with word \"%s\" and in use pretty often. "+
				"Include \"%s\" into beginning of your list. After each word write \"-\" and translation in %s language. "+
				"Write only words in your response.",
			strings.TrimSpace(word),
			strings.TrimSpace(word),
			base.ISO3(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("prepare request body: %w", err)
	}

	req, err := basehttp.NewRequest(basehttp.MethodPost, s.addr, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	s.authorizeReq(req)

	resp, err := s.client.Do(req)
	defer func() {
		if resp != nil {
			if closeErr := resp.Body.Close(); closeErr != nil {
				log.Printf("close response body: %s", closeErr.Error())
			}
		}
	}()
	if err != nil {
		return nil, fmt.Errorf("executing request error: %w", err)
	}

	if resp.StatusCode != basehttp.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var parsedResponse response
	if err = json.Unmarshal(respBody, &parsedResponse); err != nil {
		return nil, fmt.Errorf("parse response body: %w", err)
	}

	if len(parsedResponse.Choices) == 0 {
		return nil, errors.New("connection successful but response is empty")
	}

	words := map[string]string{}

	for _, line := range strings.Split(parsedResponse.Choices[0].Message.Content, "\n") {
		parts := strings.Split(line, "-")
		if len(parts) != 2 {
			continue
		}
		words[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return words, nil
}

func (s *Scraper) GenStory(words []string, lang language.Tag) (string, error) {
	if len(words) == 0 {
		return "", fmt.Errorf("words are missing")
	}

	return s.genStoryWithRetry(words, lang, 5)
}

func (s *Scraper) genStoryWithRetry(words []string, lang language.Tag, retries uint32) (string, error) {
	var (
		story string
		err   error
	)
	for retry := 0; retry <= int(retries); retry++ {
		story, err = s.genStory(words, lang)
		if err == nil {
			return story, nil
		}
		time.Sleep(5 * time.Second)
	}

	return story, err
}

func (s *Scraper) genStory(words []string, lang language.Tag) (string, error) {
	base, _ := lang.Base()

	body, err := s.prepareRequestBody(fmt.Sprintf("Generate a story using words %v in language %s. The story should contain only one word from that list per sentence.", words, base.ISO3()))
	if err != nil {
		return "", fmt.Errorf("prepare request body: %w", err)
	}

	req, err := basehttp.NewRequest(basehttp.MethodPost, s.addr, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	s.authorizeReq(req)

	resp, err := s.client.Do(req)
	defer func() {
		if resp != nil {
			if closeErr := resp.Body.Close(); closeErr != nil {
				log.Printf("close response body: %s", closeErr.Error())
			}
		}
	}()
	if err != nil {
		return "", fmt.Errorf("executing request error: %w", err)
	}

	if resp.StatusCode != basehttp.StatusOK {
		return "", fmt.Errorf("status code: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}

	var parsedResponse response
	if err = json.Unmarshal(respBody, &parsedResponse); err != nil {
		return "", fmt.Errorf("parse response body: %w", err)
	}

	if len(parsedResponse.Choices) == 0 {
		return "", errors.New("connection successful but response is empty")
	}

	return parsedResponse.Choices[0].Message.Content, nil
}

type (
	response struct {
		ID      string   `json:"id"`
		Object  string   `json:"object"`
		Created int64    `json:"created"`
		Model   string   `json:"model"`
		Usage   usage    `json:"usage"`
		Choices []choice `json:"choices"`
	}

	usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	}

	choice struct {
		Message      message `json:"message"`
		FinishReason string  `json:"finish_reason"`
		Index        int     `json:"index"`
	}

	message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	request struct {
		Model    string    `json:"model"`
		Messages []message `json:"messages"`
	}
)

const defaultModel = "gpt-4o-mini"

func (s *Scraper) prepareRequestBody(content string) ([]byte, error) {
	req := request{
		Model: defaultModel,
		Messages: []message{
			{Role: "user", Content: strings.TrimSpace(content)},
		},
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("unexpected marshal error: %w", err)
	}

	return body, nil
}

func (s *Scraper) authorizeReq(req *basehttp.Request) {
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept-Charset", "utf-8")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.token))
}

func (c EnglishComplexity) String() string {
	if val, ok := complexities[c]; ok {
		return val
	}

	return fmt.Sprintf("complexity %d", int32(c))
}

func (c EnglishComplexity) IsValid() bool {
	_, ok := complexities[c]
	return ok
}

var complexities = map[EnglishComplexity]string{ //nolint:gochecknoglobals // it's a map of constants
	Beginner:     "Beginner",
	Intermediate: "Intermediate",
	Advanced:     "Advanced",
}

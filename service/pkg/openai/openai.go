//nolint:gomnd,lll,all // magic numbers are fine here
package openai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/language"
)

type (
	Config struct {
		Addr    string        `envconfig:"APP_OPENAI_ADDR" default:"https://api.openai.com/v1/chat/completions" json:"addr,omitempty"`
		Token   string        `envconfig:"APP_OPENAI_TOKEN" required:"true" json:"token,omitempty"`
		Retries uint          `envconfig:"APP_OPENAI_RETRIES" default:"3" json:"retries,omitempty"`
		Timeout time.Duration `envconfig:"APP_OPENAI_TIMEOUT" default:"30s" json:"timeout,omitempty"`
	}

	Scraper struct {
		addr   string
		client *http.Client
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

	client := retryablehttp.NewClient()
	client.Logger = logrus.StandardLogger()
	client.RetryMax = int(cfg.Retries)
	client.Backoff = retryablehttp.DefaultBackoff
	baseClient := client.StandardClient()
	baseClient.Timeout = cfg.Timeout

	scr := &Scraper{
		client: baseClient,
		addr:   cfg.Addr,
		token:  cfg.Token,
	}

	return scr, nil
}

func (s *Scraper) GenerateSentences(word string, size int) ([]string, error) {
	switch {
	case !utf8.ValidString(word):
		return nil, fmt.Errorf("invalid utf8 string: %s", word)
	case size < 0:
		return nil, errors.New("size should be positive")
	case size == 0:
		return nil, nil
	}

	word = strings.TrimSpace(word)
	if len(word) == 0 {
		return nil, errors.New("word is blank")
	}

	body, err := s.prepareRequestBody(
		fmt.Sprintf(
			"Generate %d random sentences with the word \"%s\". "+
				"Include as much different meanings of this word in the sentence as possible. "+
				"Also leave an exaplanation of each meaning after the sentence. "+
				"Try to generate sentence for Intermediate English level with any topics. "+
				"Format of your output should be a JSON: [{\"sentence\":{\"value\":\"sentence itself\",\"explanations\":[\"explanation_1\",\"explanation_2\"]}}]"+
				"Also don't use line break for between a sentence and its explanations.",
			size,
			strings.TrimSpace(word),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("prepare request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.addr, bytes.NewReader(body))
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
		return nil, fmt.Errorf("request execution error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
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
	// todo: optimise it
	for index := 0; index < len(content); index++ {
		var generatedSentencesResp []generatedSentencesResponse
		if err = json.Unmarshal([]byte(content[index]), &generatedSentencesResp); err != nil {
			return nil, fmt.Errorf("parse response body: %w", err)
		}

		for _, sentence := range generatedSentencesResp {
			explanations := ""
			for _, explanation := range sentence.Sentence.Explanations {
				explanations += "\n"
				explanations += "- " + explanation
			}

			sentences = append(sentences, fmt.Sprintf("%s%s", sentence.Sentence.Value, explanations))
		}

	}
	if len(sentences) > int(size) {
		return sentences[:size], nil
	}
	return sentences, nil
}

type generatedSentencesResponse struct {
	Sentence struct {
		Value        string   `json:"value"`
		Explanations []string `json:"explanations"`
	} `json:"sentence"`
}

func (s *Scraper) GetFamilyWordsWithTranslation(word string, lang language.Tag) (map[string]string, error) {
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

	req, err := http.NewRequest(http.MethodPost, s.addr, bytes.NewReader(body))
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
		return nil, fmt.Errorf("request execution error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
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
	base, _ := lang.Base()

	body, err := s.prepareRequestBody(fmt.Sprintf("Generate a story using words %v in language %s. The story should contain only one word from that list per sentence.", words, base.ISO3()))
	if err != nil {
		return "", fmt.Errorf("prepare request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.addr, bytes.NewReader(body))
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
		return "", fmt.Errorf("request execution error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
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

func (s *Scraper) authorizeReq(req *http.Request) {
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept-Charset", "utf-8")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.token))
}

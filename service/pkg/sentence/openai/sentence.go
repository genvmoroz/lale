package openai

import (
	"encoding/json"
	"errors"
	"fmt"
	basehttp "net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/valyala/fasthttp"

	"github.com/genvmoroz/client-go/http"
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
		client *http.Client
		token  string
	}
)

func NewSentenceScraper(cfg Config) (*Scraper, error) {
	_, err := url.ParseRequestURI(cfg.Addr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse addr [%s]: %w", cfg.Addr, err)
	}
	if cfg.Timeout < 0 {
		return nil, fmt.Errorf("timeout shouldn't be negative [%d]: %w", cfg.Timeout, err)
	}

	client, err := http.NewClient(http.WithRetry(cfg.Retries, cfg.Timeout))
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	scr := &Scraper{
		client: client,
		addr:   cfg.Addr,
		token:  cfg.Token,
	}
	if err = scr.ping(); err != nil {
		return nil, fmt.Errorf("failed to ping the service: %w", err)
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

func (s *Scraper) ScrapeSentences(word string, size uint32) ([]string, error) {
	var result []string
	//sent, err := s.sentences(word, size, Beginner)
	//if err != nil {
	//	return nil, err
	//}

	//result = append(result, sent...)

	sent, err := s.sentences(word, size, Intermediate)
	if err != nil {
		return nil, err
	}
	result = append(result, sent...)

	//sent, err = s.sentences(word, size, Advanced)
	//if err != nil {
	//	return nil, err
	//}
	//
	//result = append(result, sent...)

	return result, nil
}

func (s *Scraper) sentences(word string, size uint32, complexity EnglishComplexity) ([]string, error) {
	if !utf8.ValidString(word) {
		return nil, fmt.Errorf("invalid utf8 string: %s", word)
	}
	if len(strings.TrimSpace(word)) == 0 {
		return nil, errors.New("word is blank")
	}

	req := http.AcquireRequest()
	defer http.ReleaseRequest(req)

	req.Header.SetRequestURI(s.addr)
	req.Header.SetMethod(basehttp.MethodPost)

	s.authorizeReq(req)

	body, err := s.prepareRequestBody(
		fmt.Sprintf(
			"Generate %d random sentences with the word \"%s\" for %s English level with any topics.",
			size,
			strings.TrimSpace(word),
			complexity.String(),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request body: %w", err)
	}

	req.AppendBody(body)

	resp, err := s.client.Do(req)
	defer func() {
		if resp != nil {
			http.ReleaseResponse(resp)
		}
	}()
	if err != nil {
		return nil, fmt.Errorf("executing request error: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode())
	}

	var parsedResponse response
	if err = json.Unmarshal(resp.Body(), &parsedResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w", err)
	}

	if len(parsedResponse.Choices) == 0 {
		return nil, errors.New("connection successful but response is empty")
	}

	sentences := strings.Split(parsedResponse.Choices[0].Message.Content, "\n")

	for index := 0; index < len(sentences); index++ {
		rq := regexp.MustCompile(`^\d.`)
		ss := rq.ReplaceAll([]byte(sentences[index]), []byte{})
		sentence := strings.TrimSpace(strings.TrimRight(string(ss), "."))
		if len(sentence) != 0 {
			sentences[index] = sentence
		}
	}
	return sentences, nil
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

const defaultModel = "gpt-3.5-turbo"

func (s *Scraper) ping() error {
	req := http.AcquireRequest()
	defer http.ReleaseRequest(req)

	req.Header.SetRequestURI(s.addr)
	req.Header.SetMethod(basehttp.MethodPost)

	s.authorizeReq(req)

	body, err := s.prepareRequestBody("Ping")
	if err != nil {
		return fmt.Errorf("failed to prepare request body: %w", err)
	}

	req.AppendBody(body)

	resp, err := s.client.Do(req)
	defer func() {
		if resp != nil {
			http.ReleaseResponse(resp)
		}
	}()
	if err != nil {
		return fmt.Errorf("executing request error: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("status code: %d", resp.StatusCode())
	}

	var parsedResponse response
	if err = json.Unmarshal(resp.Body(), &parsedResponse); err != nil {
		return fmt.Errorf("failed to parse response body: %w", err)
	}

	if len(parsedResponse.Choices) == 0 {
		return errors.New("connection successful but response is empty")
	}

	return nil
}

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

func (s *Scraper) authorizeReq(req *fasthttp.Request) {
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

var complexities = map[EnglishComplexity]string{
	Beginner:     "Beginner",
	Intermediate: "Intermediate",
	Advanced:     "Advanced",
}

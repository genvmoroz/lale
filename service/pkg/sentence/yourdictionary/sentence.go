package yourdictionary

import (
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/net/html/atom"

	web "github.com/genvmoroz/web-scraper"
)

const (
	sentencesXPath = "/html/body/div[1]/div/div/div[2]/div/div/div[1]/ul"
	sentenceXPath  = "/html/body/div[1]/div/div/div[2]/div/div/div[1]/ul/li[%d]/div/div[1]/div/span"
)

type (
	Config struct {
		Host    string        `envconfig:"APP_YOUR_DICTIONARY_SENTENCE_HOST" required:"true" default:"https://sentence.yourdictionary.com"`
		Retries uint          `envconfig:"APP_YOUR_DICTIONARY_SENTENCE_RETRIES" required:"true" default:"3"`
		Timeout time.Duration `envconfig:"APP_YOUR_DICTIONARY_SENTENCE_TIMEOUT" required:"true" default:"3s"`
	}

	Scraper struct {
		host *url.URL
		cli  web.HTTPClient
	}
)

func NewSentenceScraper(cfg Config) (*Scraper, error) {
	parsed, err := url.Parse(cfg.Host)
	if err != nil {
		return nil, fmt.Errorf("failed to parse host [%s]: %w", cfg.Host, err)
	}
	if cfg.Timeout < 0 {
		return nil, fmt.Errorf("timeout shouldn't be negative [%d]: %w", cfg.Timeout, err)
	}

	cli, err := web.NewHTTPClientWithRetry(cfg.Retries, cfg.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	return &Scraper{
		host: parsed,
		cli:  cli,
	}, nil
}

func (s Scraper) ScrapeSentences(word string, size uint32) ([]string, error) {
	if !utf8.ValidString(word) {
		return nil, fmt.Errorf("invalid utf8 string: %s", word)
	}

	query := fmt.Sprintf("%s/%s", s.host.String(), strings.TrimSpace(word))
	scr, err := web.New(query, s.cli)
	if err != nil {
		return nil, fmt.Errorf("failed to create web sentence scraper: %w", err)
	}

	sentences, err := scrapeSentences(sentencesXPath, scr)
	if err != nil {
		return nil, fmt.Errorf("failed to get all sentences: %w", err)
	}

	rand.Shuffle(len(sentences), func(i, j int) { sentences[i], sentences[j] = sentences[j], sentences[i] })

	if len(sentences) > int(size) {
		return sentences[:size], nil
	}

	return sentences, nil
}

func scrapeSentences(fullXPath string, scr *web.Scraper) ([]string, error) {
	node, err := scr.FindNode(fullXPath)
	if err != nil {
		if strings.Contains(err.Error(), "element not found") {
			return nil, nil
		}
		return nil, err
	}

	count := 0
	for n := node.FirstChild; n != nil; n = n.NextSibling {
		count++
	}

	sentences := make([]string, count)

	for i := 1; i <= count; i = i + 1 {
		nodes, err := scr.NextAfter(fmt.Sprintf(sentenceXPath, i))
		if err != nil {
			if strings.Contains(err.Error(), "element not found") {
				continue
			}
			return nil, err
		}

		b := strings.Builder{}
		for _, n := range nodes {
			if n.DataAtom != atom.Span && n.DataAtom != atom.Mark && len(n.Data) != 0 {
				b.WriteString(n.Data)
			}
		}
		sentences[i-1] = strings.TrimSpace(b.String())
	}

	return sentences, nil
}

package hippo

import (
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	web "github.com/genvmoroz/web-scraper"
)

const (
	sentencesXPath                    = "/html/body/div[2]/table/tbody/tr[2]/td/table/tbody/tr/td/div/table/tbody/tr[1]/td[2]/div/table/tbody/tr/td/div[2]/table/tbody"
	sentenceXPath                     = "/html/body/div[2]/table/tbody/tr[2]/td/table/tbody/tr/td/div/table/tbody/tr[1]/td[2]/div/table/tbody/tr/td/div[2]/table/tbody/tr[%d]/td[1]"
	classicalLiteratureSentencesXPath = "/html/body/div[2]/table/tbody/tr[2]/td/table/tbody/tr/td/div/table/tbody/tr[1]/td[2]/div/table/tbody/tr/td/div[4]/table/tbody"
	classicalLiteratureSentenceXPath  = "/html/body/div[2]/table/tbody/tr[2]/td/table/tbody/tr/td/div/table/tbody/tr[1]/td[2]/div/table/tbody/tr/td/div[4]/table/tbody/tr[%d]/td[1]"
)

type (
	Config struct {
		Host    string        `envconfig:"APP_HIPPO_SENTENCE_HOST" required:"true" default:"https://www.wordhippo.com" json:"host,omitempty"`
		Retries uint          `envconfig:"APP_HIPPO_SENTENCE_RETRIES" required:"true" default:"3" json:"retries,omitempty"`
		Timeout time.Duration `envconfig:"APP_HIPPO_SENTENCE_TIMEOUT" required:"true" default:"3s" json:"timeout,omitempty"`
	}

	Scraper struct {
		host *url.URL
		cli  web.HTTPClient
	}
)

func NewSentenceScraper(cfg Config) (*Scraper, error) {
	parsed, err := url.Parse(cfg.Host)
	if err != nil {
		return nil, fmt.Errorf("parse host [%s]: %w", cfg.Host, err)
	}
	if cfg.Timeout < 0 {
		return nil, fmt.Errorf("timeout shouldn't be negative [%d]: %w", cfg.Timeout, err)
	}

	cli, err := web.NewHTTPClientWithRetry(cfg.Retries, cfg.Timeout)
	if err != nil {
		return nil, fmt.Errorf("create http client: %w", err)
	}

	return &Scraper{
		host: parsed,
		cli:  cli,
	}, nil
}

func (s *Scraper) ScrapeSentences(word string, size uint32) ([]string, error) {
	if !utf8.ValidString(word) {
		return nil, fmt.Errorf("invalid utf8 string: %s", word)
	}

	query := fmt.Sprintf(
		"%s/what-is/sentences-with-the-word/%s.html",
		s.host.String(),
		strings.ReplaceAll(strings.TrimSpace(word), " ", "_"),
	)
	scr, err := web.New(query, s.cli)
	if err != nil {
		return nil, fmt.Errorf("create web sentence: %w", err)
	}

	sentences, err := scrapeSentences(sentencesXPath, sentenceXPath, scr)
	if err != nil {
		return nil, fmt.Errorf("scrape sentences XPath [%s]: %w", sentencesXPath, err)
	}

	classicalLiteratureSentences, err := scrapeSentences(classicalLiteratureSentencesXPath, classicalLiteratureSentenceXPath, scr)
	if err != nil {
		return nil, fmt.Errorf("scrape classical literarure sentences XPath [%s]: %w", classicalLiteratureSentencesXPath, err)
	}

	res := append(sentences, classicalLiteratureSentences...)

	rand.Shuffle(len(res), func(i, j int) { res[i], res[j] = res[j], res[i] })

	if len(res) > int(size) {
		return res[:size], nil
	}

	return res, nil
}

func scrapeSentences(tableFullXPath, sentenceFullXPath string, scr *web.Scraper) ([]string, error) {
	node, err := scr.FindNode(tableFullXPath)
	if err != nil {
		if strings.Contains(err.Error(), "element not found") {
			return nil, nil
		}
		return nil, fmt.Errorf("find Node: %w", err)
	}

	count := 0
	for n := node.FirstChild; n != nil; n = n.NextSibling {
		if n.DataAtom == atom.Tr {
			count++
		}
	}

	sentences := make([]string, 0, count)

	for i := 1; i <= count; i = i + 1 {
		xPath := fmt.Sprintf(sentenceFullXPath, i)
		nodes, err := scr.GetChildes(xPath)
		if err != nil {
			if strings.Contains(err.Error(), "element not found") {
				continue
			}
			return nil, fmt.Errorf("get childes by xpath [%s]: %w", xPath, err)
		}

		b := strings.Builder{}
		for _, n := range nodes {
			if n.Type == html.TextNode && len(n.Data) != 0 && n.Data != "\n" && n.Data != "Show More Sentences" { // exclude UI text
				b.WriteString(n.Data)
			}
		}

		sentence := strings.TrimSpace(b.String())
		if len(sentence) != 0 {
			sentences = append(sentences, sentence)
		}
	}

	return sentences, nil
}

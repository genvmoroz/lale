package dictionary

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/genvmoroz/lale/service/pkg/entity"
	clientHTTP "github.com/hashicorp/go-retryablehttp"
	"golang.org/x/text/language"
)

type (
	Config struct {
		Host    string
		Retries uint16
		Timeout time.Duration
	}

	Repo struct {
		client *clientHTTP.Client
		host   string
	}
)

func NewRepo(cfg Config) (*Repo, error) {
	httpClient := clientHTTP.NewClient()
	httpClient.RetryMax = int(cfg.Retries)

	return &Repo{client: httpClient, host: cfg.Host}, nil
}

const path = "api/v2/entries"

var ErrNotFound = errors.New("not found") // todo: move to core layer

func (c *Repo) GetWordInformation(word string, lang language.Tag) (entity.WordInformation, error) {
	target := fmt.Sprintf("%s/%s/%s/%s", c.host, path, lang.String(), url.PathEscape(word))

	req, err := clientHTTP.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		return entity.WordInformation{}, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept-Charset", "utf-8")

	resp, err := c.client.Do(req)
	defer func() {
		if resp != nil {
			if closeErr := resp.Body.Close(); closeErr != nil {
				log.Printf("close response body: %s", closeErr.Error())
			}
		}
	}()
	if err != nil {
		return entity.WordInformation{}, fmt.Errorf("execute request: %w", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return entity.WordInformation{}, fmt.Errorf("read response body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return c.decodeResponseBody(respBody)
	case http.StatusNotFound:
		return entity.WordInformation{}, ErrNotFound
	default:
		return entity.WordInformation{}, fmt.Errorf("status code: %d", resp.StatusCode)
	}
}

func (c *Repo) decodeResponseBody(body []byte) (entity.WordInformation, error) {
	var words []entity.WordInformation
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&words); err != nil {
		return entity.WordInformation{}, fmt.Errorf("decode response body: %w", err)
	}

	if len(words) == 0 {
		return entity.WordInformation{}, ErrNotFound
	}

	return words[0], nil
}

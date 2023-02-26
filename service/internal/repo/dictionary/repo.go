package dictionary

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	clientHTTP "github.com/genvmoroz/client-go/http"
	"github.com/genvmoroz/lale/service/internal/entity"
	"github.com/genvmoroz/lale/service/pkg/lang"
)

type (
	Config struct {
		Host    string
		Retries uint
		Timeout time.Duration
	}

	Repo struct {
		client *clientHTTP.Client
		host   string
	}
)

func NewRepo(cfg Config) (*Repo, error) {
	httpClient, err := clientHTTP.NewClient(clientHTTP.WithRetry(cfg.Retries, cfg.Timeout))
	if err != nil {
		return nil, err
	}

	req := clientHTTP.AcquireRequest()
	defer clientHTTP.ReleaseRequest(req)

	req.Header.SetRequestURI(cfg.Host)
	req.Header.SetMethod(http.MethodGet)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept-Charset", "utf-8")

	resp, err := httpClient.Do(req)
	defer func() {
		if resp != nil {
			clientHTTP.ReleaseResponse(resp)
		}
	}()
	if err != nil {
		return nil, fmt.Errorf("failed to reach host [%s]: %w", cfg.Host, err)
	}

	return &Repo{client: httpClient, host: cfg.Host}, nil
}

const path = "api/v2/entries"

func (c *Repo) GetWordInformation(language lang.Language, word string) (entity.WordInformation, error) {
	target := fmt.Sprintf("%s/%s/%s/%s", c.host, path, language, url.PathEscape(word))

	req := clientHTTP.AcquireRequest()
	defer clientHTTP.ReleaseRequest(req)

	req.Header.SetRequestURI(target)
	req.Header.SetMethod(http.MethodGet)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept-Charset", "utf-8")

	resp, err := c.client.Do(req)
	defer func() {
		if resp != nil {
			clientHTTP.ReleaseResponse(resp)
		}
	}()
	if err != nil {
		return entity.WordInformation{}, fmt.Errorf("failed to GET request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return entity.WordInformation{}, fmt.Errorf("status code: %d", resp.StatusCode())
	}

	var words []entity.WordInformation
	if err = json.NewDecoder(bytes.NewReader(resp.Body())).Decode(&words); err != nil {
		return entity.WordInformation{}, fmt.Errorf("failed to decode response body: %w", err)
	}

	if len(words) == 0 {
		return entity.WordInformation{}, fmt.Errorf("word [%s] not found in dictionary", word)
	}

	return words[0], nil
}

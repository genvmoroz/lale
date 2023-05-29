package dictionary

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	clientHTTP "github.com/genvmoroz/client-go/http"
	"github.com/genvmoroz/lale/service/pkg/entity"
	"golang.org/x/text/language"
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
		return nil, fmt.Errorf("reach host [%s]: %w", cfg.Host, err)
	}

	return &Repo{client: httpClient, host: cfg.Host}, nil
}

const path = "api/v2/entries"

var ErrNotFound = errors.New("not found")

func (c *Repo) GetWordInformation(word string, lang language.Tag) (entity.WordInformation, error) {
	target := fmt.Sprintf("%s/%s/%s/%s", c.host, path, lang.String(), url.PathEscape(word))

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
		return entity.WordInformation{}, fmt.Errorf("execute request: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return c.decodeResponseBody(resp.Body())
	case http.StatusNotFound:
		return entity.WordInformation{}, ErrNotFound
	default:
		return entity.WordInformation{}, fmt.Errorf("status code: %d", resp.StatusCode())
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

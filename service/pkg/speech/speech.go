package speech

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/text/language"
)

// todo: remove this abstraction level and use the client directly
type (
	Client interface {
		ToSpeech(ctx context.Context, req ToSpeechRequest) ([]byte, error)
		ListVoices(ctx context.Context, lang language.Tag) (ListVoicesResponse, error)
	}

	Repo struct {
		client Client
	}
)

func NewRepo(client Client) *Repo {
	return &Repo{client: client}
}

func (r *Repo) ToSpeech(ctx context.Context, req ToSpeechRequest) ([]byte, error) {
	if len(strings.TrimSpace(req.Input)) == 0 {
		return nil, errors.New("input is empty")
	}
	return r.client.ToSpeech(ctx, req)
}

func (r *Repo) ListVoices(ctx context.Context, lang language.Tag) (ListVoicesResponse, error) {
	return r.client.ListVoices(ctx, lang)
}

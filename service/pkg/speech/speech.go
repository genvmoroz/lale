package speech

import (
	"context"

	"github.com/genvmoroz/lale/service/pkg/lang"
)

type (
	Client interface {
		ToSpeech(ctx context.Context, req ToSpeechRequest) ([]byte, error)
		ListVoices(ctx context.Context, language lang.Language) (ListVoicesResponse, error)
	}

	Repo struct {
		client Client
	}
)

func NewRepo(client Client) *Repo {
	return &Repo{client: client}
}

func (r *Repo) ToSpeech(ctx context.Context, req ToSpeechRequest) ([]byte, error) {
	return r.client.ToSpeech(ctx, req)
}

func (r *Repo) ListVoices(ctx context.Context, language lang.Language) (ListVoicesResponse, error) {
	return r.client.ListVoices(ctx, language)
}

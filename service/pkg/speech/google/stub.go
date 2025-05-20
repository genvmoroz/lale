package google

import (
	"context"

	"github.com/genvmoroz/lale/service/pkg/speech"
	"golang.org/x/text/language"
)

// todo: improve it, add one extra func to Close a connection
type Stub struct {
}

func NewStub() *Stub {
	return &Stub{}
}

func (c *Stub) ToSpeech(_ context.Context, _ speech.ToSpeechRequest) ([]byte, error) {
	return []byte("test"), nil
}

func (c *Stub) ListVoices(_ context.Context, lang language.Tag) (speech.ListVoicesResponse, error) {
	return speech.ListVoicesResponse{
		Voices: []speech.Voice{
			{
				Languages:              []string{lang.String()},
				Name:                   "en-US-Standard-A",
				Gender:                 speech.Male,
				NaturalSampleRateHertz: 16000, //nolint:mnd // todo: remove this
			},
		},
	}, nil
}

func (c *Stub) Close() error {
	return nil
}

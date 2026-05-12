package stub

import (
	"context"

	"github.com/genvmoroz/lale/service/pkg/speech"
)

type SpeachStub struct{}

func (s *SpeachStub) ToSpeech(_ context.Context, _ speech.ToSpeechRequest) ([]byte, error) {
	return nil, nil
}

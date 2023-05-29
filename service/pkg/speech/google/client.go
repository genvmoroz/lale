package google

import (
	"context"
	"errors"
	"fmt"

	api "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"github.com/genvmoroz/lale/service/pkg/logger"
	"github.com/genvmoroz/lale/service/pkg/speech"
	"github.com/googleapis/gax-go/v2"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
)

type (
	Connection interface {
		SynthesizeSpeech(ctx context.Context, req *texttospeechpb.SynthesizeSpeechRequest, opts ...gax.CallOption) (*texttospeechpb.SynthesizeSpeechResponse, error)
		ListVoices(ctx context.Context, req *texttospeechpb.ListVoicesRequest, opts ...gax.CallOption) (*texttospeechpb.ListVoicesResponse, error)
		Close() error
	}

	TextToSpeechClient struct {
		open func(ctx context.Context) (Connection, error)
	}
)

var _ Connection = &api.Client{}

func NewTextToSpeechClient(ctx context.Context, cfg Config) (*TextToSpeechClient, error) {
	client := NewClientWithCustomConnection(googleTextToSpeechConnection(cfg))

	if err := client.ping(ctx); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return client, nil
}

func NewClientWithCustomConnection(connect func(ctx context.Context) (Connection, error)) *TextToSpeechClient {
	return &TextToSpeechClient{open: connect}
}

func (c *TextToSpeechClient) ToSpeech(ctx context.Context, req speech.ToSpeechRequest) ([]byte, error) {
	var audio []byte

	act := func(ctx context.Context, conn Connection) error {
		resp, err := conn.SynthesizeSpeech(ctx, toSynthesizeSpeechRequest(req))
		if err != nil {
			return fmt.Errorf("synthesize speech call: %w", err)
		}

		audio = resp.GetAudioContent()

		return nil
	}

	if err := c.execute(ctx, act); err != nil {
		return nil, err
	}

	return audio, nil
}

func (c *TextToSpeechClient) ListVoices(ctx context.Context, lang language.Tag) (speech.ListVoicesResponse, error) {
	var dResp *speech.ListVoicesResponse

	act := func(ctx context.Context, conn Connection) error {
		req := &texttospeechpb.ListVoicesRequest{LanguageCode: lang.String()}
		resp, err := conn.ListVoices(ctx, req)
		if err != nil {
			return fmt.Errorf("list voices call: %w", err)
		}

		dResp = toListVoicesResponse(resp)

		return nil
	}

	if err := c.execute(ctx, act); err != nil {
		return speech.ListVoicesResponse{}, err
	}

	if dResp != nil {
		return *dResp, nil
	}

	return speech.ListVoicesResponse{}, nil
}

func (c *TextToSpeechClient) execute(ctx context.Context, act func(ctx context.Context, conn Connection) error) error {
	conn, err := c.open(ctx)
	switch {
	case err != nil:
		return fmt.Errorf("new google connection: %w", err)
	case conn == nil:
		return errors.New("connection isn't established")
	}
	defer c.close(ctx, conn)

	return act(ctx, conn)
}

func (c *TextToSpeechClient) ping(ctx context.Context) error {
	ping := func(ctx context.Context, conn Connection) error { return nil }
	return c.execute(ctx, ping)
}

func googleTextToSpeechConnection(cfg Config) func(context.Context) (Connection, error) {
	return func(ctx context.Context) (Connection, error) {
		return api.NewClient(ctx, option.WithCredentialsFile(cfg.ProjectKeyFile))
	}
}

func (c *TextToSpeechClient) close(ctx context.Context, conn Connection) {
	if conn != nil {
		if err := conn.Close(); err != nil {
			logger.
				FromContext(ctx).
				Warnf("suppressed close google connection error: %s", err.Error())
		}
	}
}

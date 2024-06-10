package story

import (
	"context"
	"fmt"
	"strings"

	"github.com/genvmoroz/bot-engine/bot"
	"github.com/genvmoroz/lale-tg-client/internal/auxl"
	"github.com/genvmoroz/lale-tg-client/internal/repository"
	"github.com/genvmoroz/lale/service/api"
	"github.com/sirupsen/logrus"
)

type State struct {
	laleRepo *repository.LaleRepo
}

const Command = "/story"

func NewState(laleRepo *repository.LaleRepo) *State {
	return &State{laleRepo: laleRepo}
}

const initialMessage = `
Let the service generate a story with words you have learnt.
`

func (s *State) Process(ctx context.Context, client *bot.Client, chatID int64, updateChan bot.UpdatesChannel) error {
	if err := client.Send(chatID, initialMessage); err != nil {
		return err
	}

	language, userName, back, err := auxl.RequestInput[string](
		ctx,
		func(s string) bool {
			return len(strings.TrimSpace(s)) != 0
		},
		chatID,
		"Send the ISO 1 Letter Language Code of the story, ex: <code>en</code>",
		func(input string, _ int64, _ *bot.Client) (string, error) {
			return strings.TrimSpace(input), nil
		},
		client,
		updateChan,
	)
	if err != nil {
		return fmt.Errorf("request language: %w", err)
	}
	if back {
		return nil
	}

	req := &api.GenerateStoryRequest{
		UserID:   userName,
		Language: language,
	}

	resp, err := s.laleRepo.Client.GenerateStory(ctx, req)
	if err != nil {
		if sendErr := client.SendWithParseMode(chatID, fmt.Sprintf("<code>grpc [GenerateStory] err: %s</code>", err.Error()), "HTML"); sendErr != nil {
			logrus.
				WithField("grpc error", err.Error()).
				WithField("tg-bot error", sendErr.Error()).
				Errorf("internal error")
			return err
		}
		return err
	}

	return client.Send(chatID, resp.GetStory())
}

func (s *State) Command() string {
	return Command
}

func (s *State) Description() string {
	return "Repeat Card"
}

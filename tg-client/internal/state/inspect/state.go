package inspect

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/genvmoroz/bot-engine/bot"
	"github.com/genvmoroz/lale-tg-client/internal/pretty"
	"github.com/genvmoroz/lale-tg-client/internal/repository"
	"github.com/genvmoroz/lale/service/api"
)

type State struct {
	laleRepo *repository.LaleRepo
}

const Command = "/inspect"

func NewState(laleRepo *repository.LaleRepo) *State {
	return &State{laleRepo: laleRepo}
}

const initialMessage = `
Inspect Card State
Send the Word, Language to inspect the Card with that values
`

func (s *State) Process(ctx context.Context, client *bot.Client, chatID int64, updateChan bot.UpdatesChannel) error {
	if err := client.Send(chatID, initialMessage); err != nil {
		return err
	}

	req := &api.InspectCardRequest{}

	for len(req.Language) == 0 {
		if len(req.GetLanguage()) == 0 {
			if err := client.SendWithParseMode(chatID, "Send the ISO 1 Letter Language Code . Ex. <code>en</code>", "HTML"); err != nil {
				return err
			}
		}
		select {
		case <-ctx.Done():
			return nil
		case update, ok := <-updateChan:
			if !ok {
				return errors.New("updateChan is closed")
			}
			text := strings.TrimSpace(update.Message.Text)
			switch text {
			case "/back":
				return client.Send(chatID, "Back to previous state")
			case "":
				if err := client.Send(chatID, "Empty value is not allowed"); err != nil {
					return err
				}
			default:
				req.Language = text
			}
		}
	}

	for len(req.GetWord()) == 0 {
		if err := client.SendWithParseMode(chatID, "Send the word. Ex. <code>suspicion</code>", "HTML"); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return nil
		case update, ok := <-updateChan:
			if !ok {
				return errors.New("updateChan is closed")
			}
			text := strings.TrimSpace(update.Message.Text)
			switch text {
			case "/back":
				return client.Send(chatID, "Back to previous state")
			case "":
				if err := client.Send(chatID, "Empty value is not allowed"); err != nil {
					return err
				}
			default:
				req.UserID = strings.TrimSpace(update.Message.From.UserName)
				req.Word = text
			}
		}
	}

	resp, err := s.laleRepo.Client.InspectCard(ctx, req)
	if err != nil {
		if err = client.SendWithParseMode(chatID, fmt.Sprintf("<code>grpc [InspectCard] err: %s</code>", err.Error()), "HTML"); err != nil {
			return err
		}
	}

	for _, msg := range pretty.Card(resp, true) {
		if err = client.SendWithParseMode(chatID, msg, "HTML"); err != nil {
			return err
		}
	}

	return nil
}

func (s *State) Command() string {
	return Command
}

func (s *State) Description() string {
	return "Inspect Card"
}

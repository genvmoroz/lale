package getall

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/genvmoroz/bot-engine/bot"
	"github.com/genvmoroz/lale/service/api"
	"github.com/genvmoroz/lale/tg-client/internal/repository"
)

type State struct {
	laleRepo *repository.LaleRepo
}

const Command = "/getall"

func NewState(laleRepo *repository.LaleRepo) *State {
	return &State{laleRepo: laleRepo}
}

const initialMessage = `
Get All Cards State
`

func (s *State) Process(ctx context.Context, client *bot.Client, chatID int64, updateChan bot.UpdatesChannel) error {
	if err := client.Send(chatID, initialMessage); err != nil {
		return err
	}

	var req *api.GetCardsRequest

	for req == nil {
		if err := client.Send(chatID, "Send the language. Ex. en. Or all to request cards without filtering by language"); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return nil
		case update, ok := <-updateChan:
			if !ok {
				return errors.New("updateChan is closed")
			}
			text := strings.ToLower(strings.TrimSpace(update.Message.Text))
			switch text {
			case "/back":
				return client.Send(chatID, "Back to previous state")
			case "":
				if err := client.Send(chatID, "Empty value is not allowed"); err != nil {
					return err
				}
			case "all":
				req = &api.GetCardsRequest{
					UserID:   strings.TrimSpace(update.Message.From.UserName),
					Language: "",
				}
			default:
				req = &api.GetCardsRequest{
					UserID:   strings.TrimSpace(update.Message.From.UserName),
					Language: text,
				}
			}
		}
	}

	resp, err := s.laleRepo.GetAllCards(ctx, req)
	if err != nil {
		if err = client.SendWithParseMode(chatID, fmt.Sprintf("grpc [InspectCard] err: %s", err.Error()), "HTML"); err != nil {
			return err
		}
	}

	empJSON, err := json.MarshalIndent(resp.GetCards(), "", "\t\t\t")
	if err != nil {
		return err
	}

	if err = client.SendWithParseMode(chatID, fmt.Sprintf("%d Cards found:\n%s", len(resp.GetCards()), string(empJSON)), "HTML"); err != nil {
		return err
	}

	return nil
}

func (s *State) Command() string {
	return Command
}

func (s *State) Description() string {
	return "Inspect Card"
}

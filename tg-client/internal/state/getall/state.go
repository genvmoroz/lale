package getall

import (
	"context"
	"errors"
	"fmt"
	"github.com/genvmoroz/bot-engine/processor"
	"github.com/genvmoroz/bot-engine/tg"
	"strings"

	"github.com/genvmoroz/lale-tg-client/internal/pretty"
	"github.com/genvmoroz/lale-tg-client/internal/repository"
	"github.com/genvmoroz/lale/service/api"
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

func (s *State) Process(ctx context.Context, client processor.Client, chatID int64, updateChan tg.UpdatesChannel) error {
	if err := client.Send(chatID, initialMessage); err != nil {
		return err
	}

	var req *api.GetCardsRequest

	for req == nil {
		if err := client.SendWithParseMode(chatID, "Send the ISO 1 Letter Language Code. Ex. <code>en</code>. Or  <code>all</code> to request cards without filtering by language", tg.ModeHTML); err != nil {
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

	resp, err := s.laleRepo.Client.GetAllCards(ctx, req)
	if err != nil {
		if err = client.SendWithParseMode(chatID, fmt.Sprintf("<code>grpc [GetAllCards] err: %s</code>", err.Error()), tg.ModeHTML); err != nil {
			return err
		}
	}

	for _, card := range resp.GetCards() {
		for _, msg := range pretty.Card(card, true) {
			if err = client.SendWithParseMode(chatID, msg, tg.ModeHTML); err != nil {
				return err
			}
		}
	}

	if err = client.Send(chatID, fmt.Sprintf("Cards found %d", len(resp.GetCards()))); err != nil {
		return err
	}

	return nil
}

func (s *State) Command() string {
	return Command
}

func (s *State) Description() string {
	return "Get All Card"
}

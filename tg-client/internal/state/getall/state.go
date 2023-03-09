package getall

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/genvmoroz/bot-engine/bot"
	"github.com/genvmoroz/lale/service/api"
	"github.com/genvmoroz/lale/tg-client/internal/pretty"
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
		if err := client.SendWithParseMode(chatID, "Send the language. Ex. <code>en</code>. Or  <code>all</code> to request cards without filtering by language", "HTML"); err != nil {
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
		if err = client.Send(chatID, fmt.Sprintf("grpc [GetAllCards] err: %s", err.Error())); err != nil {
			return err
		}
	}

	for _, card := range resp.GetCards() {
		for _, msg := range pretty.Card(card) {
			if err = client.SendWithParseMode(chatID, msg, "HTML"); err != nil {
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
	return "Inspect Card"
}

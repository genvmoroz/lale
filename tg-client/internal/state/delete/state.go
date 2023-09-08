package delete

import (
	"context"
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

const Command = "/delete"

func NewState(laleRepo *repository.LaleRepo) *State {
	return &State{laleRepo: laleRepo}
}

const initialMessage = `
Delete Card State
`

func (s *State) Process(ctx context.Context, client *bot.Client, chatID int64, updateChan bot.UpdatesChannel) error {
	if err := client.Send(chatID, initialMessage); err != nil {
		return err
	}

	var req *api.DeleteCardRequest

	for req == nil {
		if err := client.Send(chatID, "Send the ID of the Card you want to be deleted"); err != nil {
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
			default:
				req = &api.DeleteCardRequest{
					UserID: strings.TrimSpace(update.Message.From.UserName),
					CardID: text,
				}
			}
		}
	}

	resp, err := s.laleRepo.Client.DeleteCard(ctx, req)
	if err != nil {
		if err = client.SendWithParseMode(chatID, fmt.Sprintf("<code>grpc [DeleteCard] err: %s</code>", err.Error()), "HTML"); err != nil {
			return err
		}
	}

	if err = client.SendWithParseMode(chatID, fmt.Sprintf("Card with ID <code>%s</code> deleted", resp.GetCard().GetId()), "HTML"); err != nil {
		return err
	}

	return nil
}

func (s *State) Command() string {
	return Command
}

func (s *State) Description() string {
	return "Delete Card"
}

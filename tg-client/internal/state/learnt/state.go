package learnt

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/genvmoroz/bot-engine/processor"
	"github.com/genvmoroz/bot-engine/tg"
	"github.com/genvmoroz/lale-tg-client/internal/repository"
	"github.com/genvmoroz/lale/service/api"
)

type State struct {
	laleRepo *repository.LaleRepo
}

const Command = "/learnt"

func NewState(laleRepo *repository.LaleRepo) *State {
	return &State{laleRepo: laleRepo}
}

const initialMessage = `
Mark card learnt
`

const promptCardID = "Send the card ID to mark as learnt. " +
	"It remains stored for statistics and is removed from learn/repeat queues."

func (s *State) Process(ctx context.Context, client processor.Client, chatID int64, updateChan tg.UpdatesChannel) error {
	if err := client.Send(chatID, initialMessage); err != nil {
		return err
	}

	var req *api.MarkCardLearntRequest

	for req == nil {
		if err := client.Send(chatID, promptCardID); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return nil
		case update, ok := <-updateChan:
			if !ok {
				return errors.New("updateChan is closed")
			}
			raw := strings.TrimSpace(update.Message.Text)
			switch strings.ToLower(raw) {
			case "/back":
				return client.Send(chatID, "Back to previous state")
			case "":
				if err := client.Send(chatID, "Empty value is not allowed"); err != nil {
					return err
				}
			default:
				req = &api.MarkCardLearntRequest{
					UserID: strings.TrimSpace(update.Message.From.UserName),
					CardID: raw,
				}
			}
		}
	}

	resp, err := s.laleRepo.Client.MarkCardLearnt(ctx, req)
	if err != nil {
		return client.SendWithParseMode(
			chatID,
			fmt.Sprintf("<code>grpc [MarkCardLearnt] err: %s</code>", err.Error()),
			tg.ModeHTML,
		)
	}

	return client.SendWithParseMode(
		chatID,
		fmt.Sprintf("Card <code>%s</code> marked as learnt", resp.GetId()),
		tg.ModeHTML,
	)
}

func (s *State) Command() string {
	return Command
}

func (s *State) Description() string {
	return "Mark card learnt"
}

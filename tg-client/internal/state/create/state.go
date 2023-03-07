package create

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

const Command = "/create"

func NewState(laleRepo *repository.LaleRepo) *State {
	return &State{laleRepo: laleRepo}
}

const initialMessage = `
Create New Card State.
`
const setWords = `
Send the Card you want to create. Examples:
suspect - вважати, підозрювати
suspicion - підозра, сумнів
suspicious - підозрілий, сумнівний
suspiciously - підозріло, сумнівно
`

func (s *State) Process(ctx context.Context, client *bot.Client, chatID int64, updateChan bot.UpdatesChannel) error {
	if err := client.Send(chatID, initialMessage); err != nil {
		return err
	}

	req := &api.CreateCardRequest{}

	if err := client.Send(chatID, "Send the language of translates, ex: en"); err != nil {
		return err
	}

	if len(req.Language) == 0 {
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
				req.Language = text
			}
		}
	}

	if err := client.Send(chatID, setWords); err != nil {
		return err
	}

	if len(req.WordInformationList) == 0 {
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
				lines := strings.Split(text, "\n")
				for _, line := range lines {
					parts := strings.Split(line, "-")
					if len(parts) != 2 {
						if err := client.Send(chatID, "Value is invalid, ex. word - translate"); err != nil {
							return err
						}
						return nil
					}

					req.WordInformationList = append(req.WordInformationList, &api.WordInformation{
						Word: strings.TrimSpace(parts[0]),
						Translate: &api.Translate{
							Translates: s.parseTranslates(parts[1]),
						},
					})
				}

			}
		}
	}

	if err := client.Send(chatID, "Do you want the service to enrich your Card with additional word information? [yes|no]"); err != nil {
		return err
	}

	for req.Params == nil {
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
			case "yes":
				req.Params = &api.CreateCardParameters{EnrichWordInformationFromDictionary: true}
			case "no":
				req.Params = &api.CreateCardParameters{EnrichWordInformationFromDictionary: false}
			default:
				if err := client.Send(chatID, "Invalid value, enter /back to go to previous state"); err != nil {
					return err
				}
			}
		}
	}

	resp, err := s.laleRepo.CreateCard(ctx, req)
	if err != nil {
		if err = client.SendWithParseMode(chatID, fmt.Sprintf("grpc [CreateCard] err: <code>%s</code>", err.Error()), "HTML"); err != nil {
			return err
		}
	}
	empJSON, err := json.MarshalIndent(resp.GetCard(), "", "\t\t\t")
	if err != nil {
		return err
	}

	if err = client.SendWithParseMode(chatID, fmt.Sprintf("Creation completed, card:\n<code>%s</code>", string(empJSON)), "HTML"); err != nil {
		return err
	}
	return nil

}

func (s *State) parseTranslates(raw string) []string {
	var translates []string

	parts := strings.Split(raw, ", ")
	for _, part := range parts {
		translate := strings.Trim(part, " .,()")
		if len(translate) != 0 {
			translates = append(translates, translate)
		}
	}

	return translates
}

func (s *State) Command() string {
	return Command
}

func (s *State) Description() string {
	return "Create new Card"
}

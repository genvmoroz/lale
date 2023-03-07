package create

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/genvmoroz/lale/tg-client/internal/auxl"
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
const createExample = `
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

	language, userName, back, err := auxl.RequestInput[string](
		ctx,
		func(s string) bool {
			return len(strings.TrimSpace(s)) != 0
		},
		chatID,
		"Send the language of Translations, ex: en",
		func(input string, _ int64, _ *bot.Client) (string, error) {
			return strings.TrimSpace(input), nil
		},
		client,
		updateChan,
	)
	if err != nil {
		return fmt.Errorf("failed to request language: %w", err)
	}
	if back {
		return nil
	}

	wordsList, _, back, err := auxl.RequestInput[[][2]string](
		ctx,
		func(s [][2]string) bool {
			return len(s) != 0
		},
		chatID,
		createExample,
		func(input string, chatID int64, client *bot.Client) ([][2]string, error) {
			lines := strings.Split(input, "\n")
			if len(lines) == 0 {
				return nil, client.Send(chatID, "Send at least one word with translation")
			}
			var wordList [][2]string

			for _, line := range lines {
				parts := strings.Split(line, "-")
				if len(parts) != 2 {
					return nil, client.Send(chatID, fmt.Sprintf("Value [%s] is invalid, ex. word - translation", line))
				}
				wordList = append(wordList, [2]string{parts[0], parts[1]})
			}
			return wordList, nil
		},
		client,
		updateChan,
	)
	if err != nil {
		return fmt.Errorf("failed to request words list: %w", err)
	}
	if back {
		return nil
	}

	for _, word := range wordsList {
		req.WordInformationList = append(req.WordInformationList, &api.WordInformation{
			Word: strings.TrimSpace(word[0]),
			Translation: &api.Translation{
				Language:     language,
				Translations: s.parseTranslations(word[1]),
			},
		})
	}

	enrichWithAdditionalInformation, _, back, err := auxl.RequestInput[*bool](
		ctx,
		func(s *bool) bool {
			return s != nil
		},
		chatID,
		"Do you want the service to enrich your Card with additional word information? [yes|no]",
		func(input string, chatID int64, client *bot.Client) (*bool, error) {
			text := strings.ToLower(strings.TrimSpace(input))
			switch text {
			case "/back":
				return nil, client.Send(chatID, "Back to previous state")
			case "":
				return nil, client.Send(chatID, "Empty value is not allowed")
			case "yes":
				t := true
				return &t, nil
			case "no":
				t := false
				return &t, nil
			default:
				return nil, client.Send(chatID, "Invalid value, enter /back to go to previous state")
			}
		},
		client,
		updateChan,
	)
	if err != nil {
		return fmt.Errorf("failed to request enrich with additional information: %w", err)
	}
	if back {
		return nil
	}

	req.UserID = userName
	req.Language = language
	if enrichWithAdditionalInformation != nil {
		req.Params = &api.CreateCardParameters{
			EnrichWordInformationFromDictionary: *enrichWithAdditionalInformation,
		}
	}

	resp, err := s.laleRepo.CreateCard(ctx, req)
	if err != nil {
		if err = client.SendWithParseMode(chatID, fmt.Sprintf("grpc [CreateCard] err: %s", err.Error()), "HTML"); err != nil {
			return err
		}
	}
	empJSON, err := json.MarshalIndent(resp.GetCard(), "", "\t\t\t")
	if err != nil {
		return err
	}

	if err = client.SendWithParseMode(chatID, fmt.Sprintf("Creation completed, card:\n%s", string(empJSON)), "HTML"); err != nil {
		return err
	}
	return nil

}

func (s *State) parseTranslations(raw string) []string {
	var Translations []string

	parts := strings.Split(raw, ", ")
	for _, part := range parts {
		Translation := strings.Trim(part, " .,()")
		if len(Translation) != 0 {
			Translations = append(Translations, Translation)
		}
	}

	return Translations
}

func (s *State) Command() string {
	return Command
}

func (s *State) Description() string {
	return "Create new Card"
}

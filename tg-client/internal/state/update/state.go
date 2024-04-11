package update

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/genvmoroz/bot-engine/bot"
	"github.com/genvmoroz/lale/service/api"
	"github.com/genvmoroz/lale/tg-client/internal/auxl"
	"github.com/genvmoroz/lale/tg-client/internal/repository"
	"slices"
)

type State struct {
	laleRepo *repository.LaleRepo
}

const Command = "/update"

const createExample = `
Send the Card you want to create. Examples:
<code>word - translate 1, translate 2</code>
`

func NewState(laleRepo *repository.LaleRepo) *State {
	return &State{laleRepo: laleRepo}
}

const initialMessage = `
Update Card State
`

func (s *State) Process(ctx context.Context, client *bot.Client, chatID int64, updateChan bot.UpdatesChannel) error {
	if err := client.Send(chatID, initialMessage); err != nil {
		return err
	}

	var req *api.UpdateCardRequest

	for req == nil {
		if err := client.Send(chatID, "Send the ID of the Card you want to update"); err != nil {
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
				req = &api.UpdateCardRequest{
					UserID: strings.TrimSpace(update.Message.From.UserName),
					CardID: text,
				}
			}
		}
	}

	language, _, back, err := auxl.RequestInput[string](
		ctx,
		func(s string) bool {
			return len(strings.TrimSpace(s)) != 0
		},
		chatID,
		"Send the ISO 1 Letter Language Code, ex: <code>en</code>",
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

	wordsList, _, back, err := auxl.RequestInput[[][2]string](
		ctx,
		func(s [][2]string) bool {
			return len(s) != 0
		},
		chatID,
		createExample,
		func(input string, chatID int64, client *bot.Client) ([][2]string, error) {
			lines := slices.DeleteFunc(strings.Split(input, "\n"),
				func(s string) bool {
					return len(strings.TrimSpace(s)) == 0
				},
			)
			if len(lines) == 0 {
				return nil, client.Send(chatID, "Send at least one word with translation")
			}
			var wordList [][2]string

			for _, line := range lines {
				parts := strings.Split(line, "-")
				if len(parts) != 2 {
					return nil, client.SendWithParseMode(chatID, fmt.Sprintf("Value [%s] is invalid, ex. <code>word - translation</code>", line), "HTML")
				}
				wordList = append(wordList, [2]string{parts[0], parts[1]})
			}
			return wordList, nil
		},
		client,
		updateChan,
	)
	if err != nil {
		return fmt.Errorf("request words list: %w", err)
	}
	if back {
		return nil
	}

	for _, word := range wordsList {
		req.WordInformationList = append(req.WordInformationList, &api.WordInformation{
			Word: strings.TrimSpace(word[0]),
			Translation: &api.Translation{
				Language:     strings.ToLower(language),
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
		"Do you want the service to enrich your Card with additional word information? [<code>yes</code>|<code>no</code>]",
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
				return nil, client.SendWithParseMode(chatID, fmt.Sprintf("Invalid value <code>%s</code>, enter [<code>yes</code>|<code>no</code>] or <code>/back</code> to go to the previous state", text), "HTML")
			}
		},
		client,
		updateChan,
	)
	if err != nil {
		return fmt.Errorf("request enrich with additional information: %w", err)
	}
	if back {
		return nil
	}

	if enrichWithAdditionalInformation != nil {
		req.Params = &api.Parameters{
			EnrichWordInformationFromDictionary: *enrichWithAdditionalInformation,
		}
	}

	resp, err := s.laleRepo.Client.UpdateCard(ctx, req)
	if err != nil {
		if err = client.SendWithParseMode(chatID, fmt.Sprintf("<code>grpc [UpdateCard] err: %s</code>", err.Error()), "HTML"); err != nil {
			return err
		}
	}

	if err = client.SendWithParseMode(chatID, fmt.Sprintf("Card with ID <code>%s</code> updated", resp.GetId()), "HTML"); err != nil {
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

func (s *State) parseTranslations(raw string) []string {
	var Translations []string

	parts := strings.Split(raw, ", ")
	for _, part := range parts {
		Translation := strings.Trim(part, " .,")
		if len(Translation) != 0 {
			Translations = append(Translations, Translation)
		}
	}

	return Translations
}

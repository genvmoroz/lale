package create

import (
	"context"
	"fmt"
	"strings"

	"github.com/genvmoroz/bot-engine/processor"
	"github.com/genvmoroz/bot-engine/tg"
	"github.com/genvmoroz/lale-tg-client/internal/auxl"
	"github.com/genvmoroz/lale-tg-client/internal/pretty"
	"github.com/genvmoroz/lale-tg-client/internal/repository"
	"github.com/genvmoroz/lale/service/api"
	"github.com/samber/lo"
	"slices"
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
<code>word - translate 1, translate 2</code>
`

func (s *State) Process(ctx context.Context, client processor.Client, chatID int64, updateChan tg.UpdatesChannel) error {
	if err := client.SendWithParseMode(chatID, initialMessage, tg.ModeHTML); err != nil {
		return err
	}

	req := &api.CreateCardRequest{}

	language, userName, back, err := auxl.RequestInput[string](
		ctx,
		func(s string) bool {
			return len(strings.TrimSpace(s)) != 0
		},
		chatID,
		"Send the ISO 1 Letter Language Code, ex: <code>en</code>",
		func(input string, _ int64, _ processor.Client) (string, error) {
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

	prompt, _, back, err := auxl.RequestInput[*bool](
		ctx,
		func(s *bool) bool {
			return s != nil
		},
		chatID,
		"Would you like me to prompt you possible complete card? [<code>yes</code>|<code>no</code>]",
		func(input string, chatID int64, client processor.Client) (*bool, error) {
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
				return nil, client.SendWithParseMode(chatID, fmt.Sprintf("Invalid value <code>%s</code>, enter [<code>yes</code>|<code>no</code>] or <code>/back</code> to go to the previous state", text), tg.ModeHTML)
			}
		},
		client,
		updateChan,
	)
	if err != nil {
		return fmt.Errorf("prompt card to the user: %w", err)
	}
	if back {
		return nil
	}

	if lo.FromPtr(prompt) {
		back, err = s.requestCardPrompt(ctx, language, client, chatID, updateChan)
		if err != nil {
			return fmt.Errorf("request card prompt: %w", err)
		}
		if back {
			return nil
		}
	}

	wordsList, _, back, err := auxl.RequestInput[[][2]string](
		ctx,
		func(s [][2]string) bool {
			return len(s) != 0
		},
		chatID,
		createExample,
		func(input string, chatID int64, client processor.Client) ([][2]string, error) {
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
					return nil, client.SendWithParseMode(chatID, fmt.Sprintf("Value [%s] is invalid, ex. <code>word - translation</code>", line), tg.ModeHTML)
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
		func(input string, chatID int64, client processor.Client) (*bool, error) {
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
				return nil, client.SendWithParseMode(chatID, fmt.Sprintf("Invalid value <code>%s</code>, enter [<code>yes</code>|<code>no</code>] or <code>/back</code> to go to the previous state", text), tg.ModeHTML)
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

	req.UserID = userName
	req.Language = strings.ToLower(language)
	if enrichWithAdditionalInformation != nil {
		req.Params = &api.Parameters{
			EnrichWordInformationFromDictionary: *enrichWithAdditionalInformation,
		}
	}

	resp, err := s.laleRepo.Client.CreateCard(ctx, req)
	if err != nil {
		if sendErr := client.SendWithParseMode(chatID, fmt.Sprintf("<code>grpc [CreateCard] err: %s</code>", err.Error()), tg.ModeHTML); sendErr != nil {
			return fmt.Errorf("send error [%s] message: %w", err.Error(), sendErr)
		}
		return err
	}

	if err = client.Send(chatID, "Card created"); err != nil {
		return err
	}

	for _, msg := range pretty.Card(resp, true) {
		if err = client.SendWithParseMode(chatID, msg, tg.ModeHTML); err != nil {
			return err
		}
	}

	return nil
}

func (s *State) requestCardPrompt(ctx context.Context, language string, client processor.Client, chatID int64, updateChan tg.UpdatesChannel) (bool, error) {
	word, userName, back, err := auxl.RequestInput[*string](
		ctx,
		func(s *string) bool {
			return len(strings.TrimSpace(lo.FromPtr(s))) != 0
		},
		chatID,
		"Send the word? ex: <code>suspicion</code>",
		func(input string, chatID int64, client processor.Client) (*string, error) {
			text := strings.ToLower(strings.TrimSpace(input))
			switch text {
			case "/back":
				return nil, client.Send(chatID, "Back to previous state")
			case "":
				return nil, client.Send(chatID, "Empty value is not allowed")
			default:
				return lo.ToPtr(text), nil
			}
		},
		client,
		updateChan,
	)
	if err != nil {
		return false, fmt.Errorf("request word: %w", err)
	}
	if back {
		return true, nil
	}

	req := &api.PromptCardRequest{
		UserID:              userName,
		Word:                *word,
		WordLanguage:        language,
		TranslationLanguage: "ukr",
	}

	resp, err := s.laleRepo.Client.PromptCard(ctx, req)
	if err != nil {
		return false, fmt.Errorf("invoke grpc PromptCard: %w ", err)
	}

	if len(resp.GetWords()) == 0 {
		return false, client.Send(chatID, fmt.Sprintf("card is not prompted with word %s", *word))
	}

	msg := ""
	for _, line := range resp.GetWords() {
		msg += line + "\n"
	}

	return false, client.Send(chatID, fmt.Sprintf("\n%s", msg))
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

func (s *State) Command() string {
	return Command
}

func (s *State) Description() string {
	return "Create new Card"
}

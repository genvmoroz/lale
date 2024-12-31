package repeat

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/genvmoroz/bot-engine/processor"
	"github.com/genvmoroz/bot-engine/tg"
	"github.com/genvmoroz/lale-tg-client/internal/auxl"
	"github.com/genvmoroz/lale-tg-client/internal/pretty"
	"github.com/genvmoroz/lale-tg-client/internal/repository"
	"github.com/genvmoroz/lale-tg-client/internal/state/cardseq"
	"github.com/genvmoroz/lale/service/api"
	"github.com/hako/durafmt"
	"github.com/sirupsen/logrus"
)

type State struct {
	laleRepo *repository.LaleRepo
}

const Command = "/repeat"

func NewState(laleRepo *repository.LaleRepo) *State {
	return &State{laleRepo: laleRepo}
}

const initialMessage = `
Repeat Words State
Send the ISO 1 Letter Language Code to repeat the Words with that language
`

func (s *State) Process(ctx context.Context, client processor.Client, chatID int64, updateChan tg.UpdatesChannel) error {
	if err := client.Send(chatID, initialMessage); err != nil {
		return err
	}

	language, userName, back, err := auxl.RequestInput[string](
		ctx,
		isStringNotBlank,
		chatID,

		"Send the language, ex: <code>en</code>",
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

	req := &api.GetCardsRequest{
		UserID:   userName,
		Language: language,
	}

	resp, err := s.laleRepo.Client.GetCardsToRepeat(ctx, req)
	if err != nil {
		if sendErr := client.SendWithParseMode(chatID, fmt.Sprintf("<code>grpc [GetCardsToRepeat] err: %s</code>", err.Error()), tg.ModeHTML); sendErr != nil {
			logrus.
				WithField("grpc error", err.Error()).
				WithField("tg-bot error", sendErr.Error()).
				Errorf("internal error")
			return err
		}
		return err
	}

	if err = client.SendWithParseMode(chatID, fmt.Sprintf("Found <code>%d</code> Cards to repeat", len(resp.GetCards())), tg.ModeHTML); err != nil {
		return err
	}

	cards := cardseq.NewCards(ctx, s.laleRepo, resp, 1, 1)

	for cards.HasNext() {
		card := cards.Next(ctx)

		isAnswerCorrect := true

		if card.Card.GetNextDueDate().AsTime().Equal(time.Time{}) {
			if back, err = s.processFirstRepeat(ctx, client, chatID, updateChan, card); err != nil {
				return err
			}
			if back {
				return nil
			}
			continue
		}

		if len(card.Words) == 0 {
			if err = client.SendWithParseMode(chatID, fmt.Sprintf("No words for Card <code>%s</code>. Inspect the Card and delete if empty", card.Card.GetId()), tg.ModeHTML); err != nil {
				return err
			}
			continue
		}

		for _, msg := range pretty.Card(card.Card, false) {
			if err = client.SendWithParseMode(chatID, msg, tg.ModeHTML); err != nil {
				return err
			}
		}

		for i, word := range card.Words {
			if err = client.Send(chatID, "Word:"); err != nil {
				return err
			}
			if err = client.SendWithParseMode(chatID, pretty.Translation(word.GetTranslation()), tg.ModeHTML); err != nil {
				return err
			}
			for _, meaning := range word.GetMeanings() {
				if err = client.Send(chatID, pretty.MeaningWithoutExamples(meaning)); err != nil {
					return err
				}
			}

			correct, _, back, err := auxl.RequestInput[*bool](
				ctx,
				func(u *bool) bool {
					return u != nil
				},
				chatID,
				"Send the Word",
				func(input string, chatID int64, client processor.Client) (*bool, error) {
					text := strings.ToLower(strings.TrimSpace(input))
					switch text {
					case "/back":
						return nil, client.Send(chatID, "Back to previous state")
					case "":
						return nil, client.Send(chatID, "Empty value is not allowed")
					default:
						if strings.EqualFold(text, word.GetWord()) {
							t := true
							return &t, nil
						}
						t := false
						return &t, nil
					}
				},
				client,
				updateChan,
			)
			if err != nil {
				return fmt.Errorf("request easiness level: %w", err)
			}
			if back {
				return nil
			}

			if correct != nil && *correct {
				if err = client.Send(chatID, "Correct"); err != nil {
					return err
				}
			} else {
				if err = client.SendWithParseMode(chatID, fmt.Sprintf("Incorrect, inspect word <code>%s</code> first", word.GetWord()), tg.ModeHTML); err != nil {
					return err
				}
				isAnswerCorrect = false
			}
			err = client.SendAudio(chatID, "pronunciation", word.GetAudio())
			if err != nil {
				if err = client.Send(chatID, fmt.Sprintf("sending audio error: %v", err.Error())); err != nil {
					return err
				}
			}
			if err = client.Send(chatID, "Sentences:"); err != nil {
				return err
			}

			task := card.Sentences[word.GetWord()]
			sentences, err := task.Get(time.Minute)
			if err != nil {
				if sendErr := client.Send(chatID, fmt.Sprintf("getting sentences error: %s", err.Error())); sendErr != nil {
					return fmt.Errorf("send error [%s]: %w", err.Error(), sendErr)
				}
			}
			for _, sentence := range sentences {
				if err = client.Send(chatID, sentence); err != nil {
					return err
				}
			}

			if i != len(card.Words)-1 {
				_, _, back, err = auxl.RequestInput[*bool](
					ctx,
					func(s *bool) bool {
						return s != nil
					},
					chatID,
					"Write <code>next</code> to repeat next word",
					func(input string, chatID int64, client processor.Client) (*bool, error) {
						text := strings.ToLower(strings.TrimSpace(input))
						switch text {
						case "":
							return nil, client.Send(chatID, "Empty value is not allowed")
						case "next":
							t := true
							return &t, nil
						default:
							return nil, client.SendWithParseMode(chatID, fmt.Sprintf("Invalid value <code>%s</code>, enter <code>/back</code> to go to the previous state", text), tg.ModeHTML)
						}
					},
					client,
					updateChan,
				)
				if err != nil {
					return fmt.Errorf("request input: %w", err)
				}
				if back {
					return nil
				}
			}
		}

		perfReq := &api.UpdateCardPerformanceRequest{
			UserID:         card.Card.GetUserID(),
			CardID:         card.Card.GetId(),
			IsInputCorrect: isAnswerCorrect,
		}

		resp, err := s.laleRepo.Client.UpdateCardPerformance(ctx, perfReq)
		if err != nil {
			if err = client.SendWithParseMode(chatID, fmt.Sprintf("<code>grpc [UpdateCardPerformance] err: %s</code>", err.Error()), tg.ModeHTML); err != nil {
				return err
			}
		}

		if err = client.Send(chatID, fmt.Sprintf("Repeat in %s", durafmt.ParseShort(resp.GetNextDueDate().AsTime().Sub(time.Now().UTC())))); err != nil {
			return err
		}
		if err = client.Send(chatID, fmt.Sprintf("At %s", resp.GetNextDueDate().AsTime())); err != nil {
			return err
		}
		if err = client.Send(chatID, fmt.Sprintf("Remaining %d cards to repeat", cards.Remaining())); err != nil {
			return err
		}
		if cards.Remaining() == 0 {
			break
		}

		_, _, back, err := auxl.RequestInput[*bool](
			ctx,
			func(s *bool) bool {
				return s != nil
			},
			chatID,
			"Write <code>next</code> to repeat next Card",
			func(input string, chatID int64, client processor.Client) (*bool, error) {
				text := strings.ToLower(strings.TrimSpace(input))
				switch text {
				case "":
					return nil, client.Send(chatID, "Empty value is not allowed")
				case "next":
					t := true
					return &t, nil
				default:
					return nil, client.SendWithParseMode(chatID, fmt.Sprintf("Invalid value <code>%s</code>, enter <code>/back</code> to go to the previous state", text), tg.ModeHTML)
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
	}

	if err = client.SendWithParseMode(chatID, fmt.Sprintf("Repeat finished, repeated <code>%d</code> cards", len(resp.GetCards())), tg.ModeHTML); err != nil {
		return err
	}

	return nil
}

func (s *State) processFirstRepeat(
	ctx context.Context,
	client processor.Client,
	chatID int64,
	updateChan tg.UpdatesChannel,
	card cardseq.Card,
) (bool, error) {
	if len(card.Words) == 0 {
		return false, client.SendWithParseMode(chatID, fmt.Sprintf("No words for Card <code>%s</code>. Inspect the Card and delete if empty", card.Card.GetId()), tg.ModeHTML)
	}

	for _, msg := range pretty.Card(card.Card, false) {
		if err := client.SendWithParseMode(chatID, msg, tg.ModeHTML); err != nil {
			return false, err
		}
	}

	for i, word := range card.Words {
		if err := client.Send(chatID, fmt.Sprintf("Word: %s", word.GetWord())); err != nil {
			return false, err
		}
		if err := client.SendWithParseMode(chatID, pretty.Translation(word.GetTranslation()), tg.ModeHTML); err != nil {
			return false, err
		}
		for _, meaning := range word.GetMeanings() {
			if err := client.Send(chatID, pretty.MeaningWithoutExamples(meaning)); err != nil {
				return false, err
			}
		}

		err := client.SendAudio(chatID, "pronunciation", word.GetAudio())
		if err != nil {
			if err = client.Send(chatID, fmt.Sprintf("sending audio error: %v", err.Error())); err != nil {
				return false, err
			}
		}

		task := card.Sentences[word.GetWord()]
		sentences, err := task.Get(time.Minute)
		if err != nil {
			if sendErr := client.Send(chatID, fmt.Sprintf("getting sentences error: %s", err.Error())); sendErr != nil {
				return false, fmt.Errorf("send error [%s]: %w", err.Error(), sendErr)
			}
		}
		for _, sentence := range sentences {
			if err = client.Send(chatID, sentence); err != nil {
				return false, err
			}
		}

		if i != len(card.Words)-1 {
			_, _, back, err := auxl.RequestInput[*bool](
				ctx,
				func(s *bool) bool {
					return s != nil
				},
				chatID,
				"Write <code>next</code> to repeat next word",
				func(input string, chatID int64, client processor.Client) (*bool, error) {
					text := strings.ToLower(strings.TrimSpace(input))
					switch text {
					case "":
						return nil, client.Send(chatID, "Empty value is not allowed")
					case "next":
						t := true
						return &t, nil
					default:
						return nil, client.SendWithParseMode(chatID, fmt.Sprintf("Invalid value <code>%s</code>, enter <code>/back</code> to go to the previous state", text), tg.ModeHTML)
					}
				},
				client,
				updateChan,
			)
			if err != nil {
				return false, fmt.Errorf("request input: %w", err)
			}
			if back {
				return true, nil
			}
		}
	}

	if err := client.SendWithParseMode(chatID, "Card repeated", tg.ModeHTML); err != nil {
		return false, err
	}

	perfReq := &api.UpdateCardPerformanceRequest{
		UserID:         card.Card.GetUserID(),
		CardID:         card.Card.GetId(),
		IsInputCorrect: false,
	}

	resp, err := s.laleRepo.Client.UpdateCardPerformance(ctx, perfReq)
	if err != nil {
		if err = client.SendWithParseMode(chatID, fmt.Sprintf("<code>grpc [UpdateCardPerformance] err: %s</code>", err.Error()), tg.ModeHTML); err != nil {
			return false, err
		}
	}

	if err = client.Send(chatID, fmt.Sprintf("Repeat in %s", durafmt.ParseShort(resp.GetNextDueDate().AsTime().Sub(time.Now().UTC())))); err != nil {
		return false, err
	}
	if err = client.Send(chatID, fmt.Sprintf("At %s", resp.GetNextDueDate().AsTime())); err != nil {
		return false, err
	}

	_, _, back, err := auxl.RequestInput[*bool](
		ctx,
		func(s *bool) bool {
			return s != nil
		},
		chatID,
		"Write <code>next</code> to repeat next Card",
		func(input string, chatID int64, client processor.Client) (*bool, error) {
			text := strings.ToLower(strings.TrimSpace(input))
			switch text {
			case "":
				return nil, client.Send(chatID, "Empty value is not allowed")
			case "next":
				t := true
				return &t, nil
			default:
				return nil, client.SendWithParseMode(chatID, fmt.Sprintf("Invalid value <code>%s</code>, enter <code>/back</code> to go to the previous state", text), tg.ModeHTML)
			}
		},
		client,
		updateChan,
	)
	if err != nil {
		return false, fmt.Errorf("request enrich with additional information: %w", err)
	}
	if back {
		return true, nil
	}
	return false, nil
}

func isStringNotBlank(s string) bool {
	return len(strings.TrimSpace(s)) != 0
}

func (s *State) Command() string {
	return Command
}

func (s *State) Description() string {
	return "Repeat Card"
}

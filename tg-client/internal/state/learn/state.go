package learn

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/genvmoroz/bot-engine/bot"
	"github.com/genvmoroz/lale/service/api"
	"github.com/genvmoroz/lale/tg-client/internal/auxl"
	"github.com/genvmoroz/lale/tg-client/internal/pretty"
	"github.com/genvmoroz/lale/tg-client/internal/repository"
	"github.com/genvmoroz/lale/tg-client/internal/state/cardseq"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

type State struct {
	laleRepo *repository.LaleRepo
}

const Command = "/learn"

func NewState(laleRepo *repository.LaleRepo) *State {
	return &State{laleRepo: laleRepo}
}

const initialMessage = `
Learn Words State
Send the ISO 1 Letter Language Code to learn the Words with that language
`

func (s *State) Process(ctx context.Context, client *bot.Client, chatID int64, updateChan bot.UpdatesChannel) error {
	if err := client.Send(chatID, initialMessage); err != nil {
		return err
	}

	language, userName, back, err := auxl.RequestInput[string](
		ctx,
		isStringNotBlank,
		chatID,

		"Send the language, ex: <code>en</code>",
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

	req := &api.GetCardsRequest{
		UserID:   userName,
		Language: language,
	}

	resp, err := s.laleRepo.Client.GetCardsToLearn(ctx, req)
	if err != nil {
		if sendErr := client.SendWithParseMode(chatID, fmt.Sprintf("<code>grpc [GetCardsToLearn] err: %s</code>", err.Error()), "HTML"); sendErr != nil {
			logrus.
				WithField("grpc error", err.Error()).
				WithField("tg-bot error", sendErr.Error()).
				Errorf("internal error")
			return err
		}
		return err
	}

	if err = client.SendWithParseMode(chatID, fmt.Sprintf("Found <code>%d</code> Cards to learn", len(resp.GetCards())), "HTML"); err != nil {
		return err
	}

	cards := cardseq.NewCards(ctx, s.laleRepo, resp, 1, 5)

	for cards.HasNext() {
		card := cards.Next(ctx)

		if card.Card.GetNextDueDate().AsTime().Equal(time.Time{}) {
			if back, err = s.processFirstReview(ctx, client, chatID, updateChan, card); err != nil {
				return err
			}
			if back {
				return nil
			}
			continue
		}

		if len(card.Words) == 0 {
			if err = client.SendWithParseMode(chatID, fmt.Sprintf("No words for Card <code>%s</code>. Inspect the Card and delete if empty", card.Card.GetId()), "HTML"); err != nil {
				return err
			}
			continue
		}

		for _, msg := range pretty.Card(card.Card, false) {
			if err = client.SendWithParseMode(chatID, msg, "HTML"); err != nil {
				return err
			}
		}

		var easiness []int

		for i, word := range card.Words {
			if err = client.Send(chatID, "Word:"); err != nil {
				return err
			}
			if err = client.SendWithParseMode(chatID, pretty.Translation(word.GetTranslation()), "HTML"); err != nil {
				return err
			}
			for _, meaning := range word.GetMeanings() {
				if err = client.Send(chatID, pretty.MeaningWithoutExamples(meaning)); err != nil {
					return err
				}
			}

			_, err = client.Bot.Send(tgbotapi.NewAudio(chatID, tgbotapi.FileBytes{
				Name:  "pronunciation",
				Bytes: word.GetAudio(),
			}))
			if err != nil {
				if err = client.Send(chatID, fmt.Sprintf("sending audio error: %v", err.Error())); err != nil {
					return err
				}
				//return fmt.Errorf("upload audio file: %w", err)
			}

			correct, _, back, err := auxl.RequestInput[*bool](
				ctx,
				func(u *bool) bool {
					return u != nil
				},
				chatID,
				"Send the Word",
				func(input string, chatID int64, client *bot.Client) (*bool, error) {
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
				easinessLevel, _, back, err := auxl.RequestInput[*uint32](
					ctx,
					func(u *uint32) bool {
						return u != nil
					},
					chatID,
					"Send Level Of Easiness. Ex. 3, range [0:5]",
					func(input string, chatID int64, client *bot.Client) (*uint32, error) {
						parsed, err := strconv.Atoi(strings.TrimSpace(input))
						switch {
						case err != nil:
							return nil, client.SendWithParseMode(chatID, fmt.Sprintf("Parsing error: %s", err.Error()), "HTML")
						case parsed < 0 || parsed > 5:
							return nil, client.Send(chatID, "The value is out of range [0:5]")
						default:
							v := uint32(parsed)
							return &v, nil
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
				if easinessLevel != nil {
					easiness = append(easiness, int(*easinessLevel))
				}
			} else {
				if err = client.SendWithParseMode(chatID, fmt.Sprintf("Incorrect, inspect word <code>%s</code> first", word.GetWord()), "HTML"); err != nil {
					return err
				}
				easiness = append(easiness, 0)
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
					"Write <code>next</code> to learn next word",
					func(input string, chatID int64, client *bot.Client) (*bool, error) {
						text := strings.ToLower(strings.TrimSpace(input))
						switch text {
						case "":
							return nil, client.Send(chatID, "Empty value is not allowed")
						case "next":
							t := true
							return &t, nil
						default:
							return nil, client.SendWithParseMode(chatID, fmt.Sprintf("Invalid value <code>%s</code>, enter <code>/back</code> to go to the previous state", text), "HTML")
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

		sum := 0

		for i := 0; i < len(easiness); i++ {
			sum += easiness[i]
		}
		avg := (float64(sum)) / (float64(len(easiness)))

		if err = client.SendWithParseMode(chatID, fmt.Sprintf("Card learnt, easiness level is <code>%d</code>", uint32(avg)), "HTML"); err != nil {
			return err
		}

		perfReq := &api.UpdateCardPerformanceRequest{
			UserID:            card.Card.GetUserID(),
			CardID:            card.Card.GetId(),
			PerformanceRating: uint32(avg),
		}

		resp, err := s.laleRepo.Client.UpdateCardPerformance(ctx, perfReq)
		if err != nil {
			if err = client.SendWithParseMode(chatID, fmt.Sprintf("<code>grpc [UpdateCardPerformance] err: %s</code>", err.Error()), "HTML"); err != nil {
				return err
			}
		}

		if err = client.Send(chatID, fmt.Sprintf("Learn in %s", resp.GetNextDueDate().AsTime().Sub(time.Now().UTC()))); err != nil {
			return err
		}
		if err = client.Send(chatID, fmt.Sprintf("At %s", resp.GetNextDueDate().AsTime())); err != nil {
			return err
		}

		if err = client.Send(chatID, fmt.Sprintf("Remaining %d cards to repeat", cards.Remaining())); err != nil {
			return err
		}

		_, _, back, err := auxl.RequestInput[*bool](
			ctx,
			func(s *bool) bool {
				return s != nil
			},
			chatID,
			"Write <code>next</code> to Learn next Card",
			func(input string, chatID int64, client *bot.Client) (*bool, error) {
				text := strings.ToLower(strings.TrimSpace(input))
				switch text {
				case "":
					return nil, client.Send(chatID, "Empty value is not allowed")
				case "next":
					t := true
					return &t, nil
				default:
					return nil, client.SendWithParseMode(chatID, fmt.Sprintf("Invalid value <code>%s</code>, enter <code>/back</code> to go to the previous state", text), "HTML")
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

	if err = client.SendWithParseMode(chatID, fmt.Sprintf("Learn finished, leanrt <code>%d</code> cards", len(resp.GetCards())), "HTML"); err != nil {
		return err
	}

	return nil
}

func (s *State) processFirstReview(
	ctx context.Context,
	client *bot.Client,
	chatID int64,
	updateChan bot.UpdatesChannel,
	card cardseq.Card,
) (bool, error) {
	if len(card.Words) == 0 {
		return false, client.SendWithParseMode(chatID, fmt.Sprintf("No words for Card <code>%s</code>. Inspect the Card and delete if empty", card.Card.GetId()), "HTML")
	}

	for _, msg := range pretty.Card(card.Card, false) {
		if err := client.SendWithParseMode(chatID, msg, "HTML"); err != nil {
			return false, err
		}
	}

	for i, word := range card.Words {
		if err := client.Send(chatID, fmt.Sprintf("Word: %s", word.GetWord())); err != nil {
			return false, err
		}
		if err := client.SendWithParseMode(chatID, pretty.Translation(word.GetTranslation()), "HTML"); err != nil {
			return false, err
		}
		for _, meaning := range word.GetMeanings() {
			if err := client.Send(chatID, pretty.MeaningWithoutExamples(meaning)); err != nil {
				return false, err
			}
		}

		_, err := client.Bot.Send(tgbotapi.NewAudio(chatID, tgbotapi.FileBytes{
			Name:  "pronunciation",
			Bytes: word.GetAudio(),
		}))
		if err != nil {
			if err = client.Send(chatID, fmt.Sprintf("sending audio error: %v", err.Error())); err != nil {
				return false, err
			}
			//return false, fmt.Errorf("upload audio file: %w", err)
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
				"Write <code>next</code> to learn next word",
				func(input string, chatID int64, client *bot.Client) (*bool, error) {
					text := strings.ToLower(strings.TrimSpace(input))
					switch text {
					case "":
						return nil, client.Send(chatID, "Empty value is not allowed")
					case "next":
						t := true
						return &t, nil
					default:
						return nil, client.SendWithParseMode(chatID, fmt.Sprintf("Invalid value <code>%s</code>, enter <code>/back</code> to go to the previous state", text), "HTML")
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

	if err := client.SendWithParseMode(chatID, "Card Learnt", "HTML"); err != nil {
		return false, err
	}

	perfReq := &api.UpdateCardPerformanceRequest{
		UserID:            card.Card.GetUserID(),
		CardID:            card.Card.GetId(),
		PerformanceRating: 0,
	}

	resp, err := s.laleRepo.Client.UpdateCardPerformance(ctx, perfReq)
	if err != nil {
		if err = client.SendWithParseMode(chatID, fmt.Sprintf("<code>grpc [UpdateCardPerformance] err: %s</code>", err.Error()), "HTML"); err != nil {
			return false, err
		}
	}

	if err = client.Send(chatID, fmt.Sprintf("Learn in %s", resp.GetNextDueDate().AsTime().Sub(time.Now().UTC()))); err != nil {
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
		"Write <code>next</code> to learn next Card",
		func(input string, chatID int64, client *bot.Client) (*bool, error) {
			text := strings.ToLower(strings.TrimSpace(input))
			switch text {
			case "":
				return nil, client.Send(chatID, "Empty value is not allowed")
			case "next":
				t := true
				return &t, nil
			default:
				return nil, client.SendWithParseMode(chatID, fmt.Sprintf("Invalid value <code>%s</code>, enter <code>/back</code> to go to the previous state", text), "HTML")
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
	return "Learn Card"
}
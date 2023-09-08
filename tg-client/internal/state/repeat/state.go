package repeat

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/genvmoroz/bot-engine/bot"
	"github.com/genvmoroz/lale/service/api"
	"github.com/genvmoroz/lale/service/pkg/future"
	"github.com/genvmoroz/lale/tg-client/internal/auxl"
	"github.com/genvmoroz/lale/tg-client/internal/pretty"
	"github.com/genvmoroz/lale/tg-client/internal/repository"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

type (
	repeatCards struct {
		index uint32
		cards []card

		cardsNumberToBeEnrichedWithSentencesInAdvance uint32

		laleRepo *repository.LaleRepo
	}

	card struct {
		Card      *api.Card
		Words     []*api.WordInformation
		Sentences map[string]future.Task[[]string]
	}
)

func newRepeatCards(ctx context.Context, laleRepo *repository.LaleRepo, resp *api.GetCardsResponse) repeatCards {
	cards := make([]card, 0, len(resp.GetCards()))
	for _, c := range resp.GetCards() {
		if c == nil {
			continue
		}
		cards = append(cards,
			card{
				Card:      c,
				Words:     c.GetWordInformationList(),
				Sentences: make(map[string]future.Task[[]string]),
			},
		)
	}
	repeatCards := repeatCards{
		index: 0,
		cards: cards,
		cardsNumberToBeEnrichedWithSentencesInAdvance: 1,
		laleRepo: laleRepo,
	}

	repeatCards.enrichCardsWithSentences(ctx)

	return repeatCards
}

func (r *repeatCards) next(ctx context.Context) card {
	if r.index >= uint32(len(r.cards)) {
		return card{}
	}

	next := r.cards[r.index]
	r.index++

	r.enrichCardsWithSentences(ctx)

	return next
}

func (r *repeatCards) hasNext() bool {
	return r.index < uint32(len(r.cards))
}

func (r *repeatCards) enrichCardsWithSentences(ctx context.Context) {
	for i := 0; i < int(r.cardsNumberToBeEnrichedWithSentencesInAdvance); i++ {
		r.enrichCardWithSentences(ctx, r.index+uint32(i))
	}
}

func (r *repeatCards) enrichCardWithSentences(ctx context.Context, i uint32) {
	if i >= uint32(len(r.cards)) || len(r.cards[i].Sentences) != 0 {
		return
	}
	if r.cards[i].Sentences == nil {
		r.cards[i].Sentences = make(map[string]future.Task[[]string])
	}
	for y, word := range r.cards[i].Words {
		y := y
		word := word

		run := func(innerCtx context.Context) ([]string, error) {
			time.Sleep(time.Duration(y*20) * time.Second)
			req := &api.GetSentencesRequest{
				UserID:         r.cards[i].Card.GetUserID(),
				Word:           word.GetWord(),
				SentencesCount: 3,
			}

			var (
				resp *api.GetSentencesResponse
				err  error
			)
			for index := 1; index <= 10; index++ {
				resp, err = r.laleRepo.Client.GetSentences(innerCtx, req)
				if err == nil {
					return resp.GetSentences(), nil
				}

				time.Sleep(time.Duration(rand.Intn(10)+5) * time.Second)
			}

			return nil, err
		}

		r.cards[i].Sentences[word.GetWord()] = future.NewTask[[]string](ctx, run)
	}
}

func (s *State) Process(ctx context.Context, client *bot.Client, chatID int64, updateChan bot.UpdatesChannel) error {
	if err := client.Send(chatID, initialMessage); err != nil {
		return err
	}

	language, userName, back, err := auxl.RequestInput[string](
		ctx,
		isStringNotBlank,
		chatID,

		"Send the language of translation, ex: <code>en</code>",
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

	req := &api.GetCardsForReviewRequest{
		UserID:   userName,
		Language: language,
	}

	resp, err := s.laleRepo.Client.GetCardsToReview(ctx, req)
	if err != nil {
		if sendErr := client.SendWithParseMode(chatID, fmt.Sprintf("<code>grpc [GetCardsToReview] err: %s</code>", err.Error()), "HTML"); sendErr != nil {
			logrus.
				WithField("grpc error", err.Error()).
				WithField("tg-bot error", sendErr.Error()).
				Errorf("internal error")
			return err
		}
		return err
	}

	if err = client.SendWithParseMode(chatID, fmt.Sprintf("Found <code>%d</code> Cards to repeat", len(resp.GetCards())), "HTML"); err != nil {
		return err
	}

	cards := newRepeatCards(ctx, s.laleRepo, resp)

	for cards.hasNext() {
		card := cards.next(ctx)

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
				return fmt.Errorf("upload audio file: %w", err)
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
					"Write <code>next</code> to review next word",
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

		if err = client.SendWithParseMode(chatID, fmt.Sprintf("Card reviewed, easiness level is <code>%d</code>", uint32(avg)), "HTML"); err != nil {
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

		if err = client.Send(chatID, fmt.Sprintf("Repeat in %s", resp.GetNextDueDate().AsTime().Sub(time.Now().UTC()))); err != nil {
			return err
		}
		if err = client.Send(chatID, fmt.Sprintf("Time %s", resp.GetNextDueDate().AsTime())); err != nil {
			return err
		}

		_, _, back, err := auxl.RequestInput[*bool](
			ctx,
			func(s *bool) bool {
				return s != nil
			},
			chatID,
			"Write <code>next</code> to review next Card",
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

	if err = client.SendWithParseMode(chatID, fmt.Sprintf("Review finished, reviewed <code>%d</code> cards", len(resp.GetCards())), "HTML"); err != nil {
		return err
	}

	return nil
}

func (s *State) processFirstReview(
	ctx context.Context,
	client *bot.Client,
	chatID int64,
	updateChan bot.UpdatesChannel,
	card card,
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
			return false, fmt.Errorf("upload audio file: %w", err)
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
				"Write <code>next</code> to review next word",
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

	if err := client.SendWithParseMode(chatID, "Card reviewed", "HTML"); err != nil {
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

	if err = client.Send(chatID, fmt.Sprintf("Repeat in %s", resp.GetNextDueDate().AsTime().Sub(time.Now().UTC()))); err != nil {
		return false, err
	}
	if err = client.Send(chatID, fmt.Sprintf("Time %s", resp.GetNextDueDate().AsTime())); err != nil {
		return false, err
	}

	_, _, back, err := auxl.RequestInput[*bool](
		ctx,
		func(s *bool) bool {
			return s != nil
		},
		chatID,
		"Write <code>next</code> to review next Card",
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
	return "Repeat Card"
}

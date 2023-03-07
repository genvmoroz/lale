package review

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/genvmoroz/bot-engine/bot"
	"github.com/genvmoroz/lale/service/api"
	"github.com/genvmoroz/lale/tg-client/internal/auxl"
	"github.com/genvmoroz/lale/tg-client/internal/repository"
	"github.com/genvmoroz/lale/tg-client/internal/transform"
)

type State struct {
	laleRepo *repository.LaleRepo
}

const Command = "/review"

func NewState(laleRepo *repository.LaleRepo) *State {
	return &State{laleRepo: laleRepo}
}

const initialMessage = `
Review Cards State
Send the Word, Language to inspect the Card with that values
`

func (s *State) Process(ctx context.Context, client *bot.Client, chatID int64, updateChan bot.UpdatesChannel) error {
	if err := client.Send(chatID, initialMessage); err != nil {
		return err
	}

	language, userName, back, err := auxl.RequestInput[string](
		ctx,
		isStringNotBlank,
		chatID,
		"Send the language of translation, ex: en",
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

	sentencesCount, _, back, err := auxl.RequestInput[*uint32](
		ctx,
		func(u *uint32) bool {
			return u != nil
		},
		chatID,
		"Send sentences count. Ex. 10",
		func(input string, chatID int64, client *bot.Client) (*uint32, error) {
			parsed, err := strconv.Atoi(strings.TrimSpace(input))
			switch {
			case err != nil:
				return nil, client.SendWithParseMode(chatID, fmt.Sprintf("Parsing error: %s", err.Error()), "HTML")
			case parsed < 0:
				return nil, client.Send(chatID, "The value cannot be negative")
			default:
				v := uint32(parsed)
				return &v, nil
			}
		},
		client,
		updateChan,
	)
	if err != nil {
		return fmt.Errorf("failed to request sentencesCount: %w", err)
	}
	if back {
		return nil
	}

	req := &api.GetCardsForReviewRequest{
		UserID:   userName,
		Language: language,
	}

	if sentencesCount != nil {
		req.SentencesCount = *sentencesCount
	}

	resp, err := s.laleRepo.GetCardsToReview(ctx, req)
	if err != nil {
		if sendErr := client.SendWithParseMode(chatID, fmt.Sprintf("grpc [GetCardsToReview] err: %s", err.Error()), "HTML"); sendErr != nil {
			logrus.
				WithField("grpc error", err.Error()).
				WithField("tg-bot error", sendErr.Error()).
				Errorf("internal error")
			return err
		}
	}

	for _, card := range resp.Cards {
		if len(card.GetWordInformationList()) == 0 {
			if err = client.Send(chatID, fmt.Sprintf("No words for Card [%s]: inspect the Card and delete if empty", card.GetId())); err != nil {
				return err
			}
			continue
		}
		for _, word := range transform.DefaultTransformer.ToCoreWordInformationList(card.GetWordInformationList()) {
			empJSON, err := yaml.Marshal(word)
			if err != nil {
				return err
			}
			if err = client.SendWithParseMode(chatID, fmt.Sprintf("%s", empJSON), "HTML"); err != nil {
				return err
			}
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
			return fmt.Errorf("failed to request easiness level: %w", err)
		}
		if back {
			return nil
		}

		perfReq := &api.UpdateCardPerformanceRequest{
			UserID: card.GetUserID(),
			CardID: card.GetId(),
		}

		if easinessLevel != nil {
			perfReq.PerformanceRating = *easinessLevel
		}

		resp, err := s.laleRepo.UpdateCardPerformance(ctx, perfReq)
		if err != nil {
			if err = client.SendWithParseMode(chatID, fmt.Sprintf("grpc [InspectCard] err: %s", err.Error()), "HTML"); err != nil {
				return err
			}
		}

		if err = client.Send(chatID, fmt.Sprintf("NextDueDate in %s", resp.GetNextDueDate().AsTime().Sub(time.Now().UTC()))); err != nil {
			return err
		}
	}

	if err = client.Send(chatID, fmt.Sprintf("Review finished, reviewed %d cards", len(resp.GetCards()))); err != nil {
		return err
	}

	return nil
}

func isStringNotBlank(s string) bool {
	return len(strings.TrimSpace(s)) != 0
}

func (s *State) Command() string {
	return Command
}

func (s *State) Description() string {
	return "Review Card"
}

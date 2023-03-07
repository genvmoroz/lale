package review

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/genvmoroz/bot-engine/bot"
	"github.com/genvmoroz/lale/service/api"
	"github.com/genvmoroz/lale/tg-client/internal/repository"
	"github.com/sirupsen/logrus"
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

	language, userName, back, err := requestInput[string](
		ctx,
		isStringNotBlank,
		chatID,
		"Send language. Ex. en",
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

	sentencesCount, _, back, err := requestInput[*uint32](
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
				if sendErr := client.SendWithParseMode(chatID, fmt.Sprintf("Parsing error: <code>%s</code>", err.Error()), "HTML"); sendErr != nil {
					return nil, sendErr
				}
			case parsed < 0:
				if sendErr := client.Send(chatID, "The value cannot be negative"); sendErr != nil {
					return nil, sendErr
				}
			default:
				v := uint32(parsed)
				return &v, nil
			}
			return nil, nil
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
		if sendErr := client.SendWithParseMode(chatID, fmt.Sprintf("grpc [GetCardsToReview] err: <code>%s</code>", err.Error()), "HTML"); sendErr != nil {
			logrus.
				WithField("grpc error", err.Error()).
				WithField("tg-bot error", sendErr.Error()).
				Errorf("internal error")
			return err
		}
	}

	for _, card := range resp.Cards {
		for _, word := range card.GetWordInformationList() {
			empJSON, err := json.MarshalIndent(word, "", "\t\t\t")
			if err != nil {
				return err
			}
			if err = client.SendWithParseMode(chatID, fmt.Sprintf("Word: %s<code>%s</code>", word.GetWord(), empJSON), "HTML"); err != nil {
				return err
			}
		}

		easinessLevel, _, back, err := requestInput[*uint32](
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
					if sendErr := client.SendWithParseMode(chatID, fmt.Sprintf("Parsing error: <code>%s</code>", err.Error()), "HTML"); sendErr != nil {
						return nil, sendErr
					}
				case parsed < 0:
					if sendErr := client.Send(chatID, "The value cannot be negative"); sendErr != nil {
						return nil, sendErr
					}
				case parsed > 5:
					if sendErr := client.Send(chatID, "The value is out of range [0:5]"); sendErr != nil {
						return nil, sendErr
					}
				default:
					v := uint32(parsed)
					return &v, nil
				}
				return nil, nil
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
			if err = client.SendWithParseMode(chatID, fmt.Sprintf("grpc [InspectCard] err: <code>%s</code>", err.Error()), "HTML"); err != nil {
				return err
			}
		}

		if err = client.Send(chatID, fmt.Sprintf("NextDueDate: %s", resp.GetNextDueDate().AsTime())); err != nil {
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

func requestInput[T any](
	ctx context.Context,
	until func(T) bool,
	chatID int64,
	message string,
	processInput func(input string, chatID int64, client *bot.Client) (T, error),
	client *bot.Client,
	updateChan bot.UpdatesChannel) (T, string, bool, error) {

	var (
		val      T
		userName string
	)

	if err := client.Send(chatID, message); err != nil {
		return val, userName, false, err
	}

	for !until(val) {
		select {
		case <-ctx.Done():
			return val, userName, false, nil
		case update, ok := <-updateChan:
			if !ok {
				return val, userName, false, errors.New("updateChan is closed")
			}
			text := strings.TrimSpace(update.Message.Text)
			switch text {
			case "/back":
				return val, userName, true, client.Send(chatID, "Back to previous state")
			case "":
				if err := client.Send(chatID, "Empty value is not allowed"); err != nil {
					return val, userName, false, err
				}
			default:
				input, err := processInput(text, chatID, client)
				if err != nil {
					return val, userName, false, fmt.Errorf("failed ")
				}

				val = input
				userName = update.Message.From.UserName
			}
		}
	}

	return val, userName, false, nil
}

func (s *State) Command() string {
	return Command
}

func (s *State) Description() string {
	return "Review Card"
}

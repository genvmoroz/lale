package repeat

import (
	"context"
	"fmt"
	"math/rand"
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

	language, userName, back, err := auxl.RequestInput(
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

			// todo: implement hinting on the server side
			if card.Card.ConsecutiveCorrectAnswersNumber <= 8 {
				hint := ""
				switch card.Card.ConsecutiveCorrectAnswersNumber {
				case 0, 1, 2:
					hint = shuffleLetters(word.GetWord())
				default:
					hint = maskWord(word.GetWord(), card.Card.ConsecutiveCorrectAnswersNumber)
				}
				if err = client.Send(chatID, "Hint: "+hint); err != nil {
					return err
				}
			}

			var lastIncorrectInput string
			checkWord := func(input string, chtID int64, cl processor.Client) (*bool, error) {
				text := strings.ToLower(strings.TrimSpace(input))
				switch text {
				case "/back":
					return nil, cl.Send(chtID, "Back to previous state")
				case "":
					return nil, cl.Send(chtID, "Empty value is not allowed")
				default:
					if strings.EqualFold(text, word.GetWord()) {
						t := true
						return &t, nil
					}
					lastIncorrectInput = text
					t := false
					return &t, nil
				}
			}

			const secondAttemptErrorThresholdPct = 20 // give second attempt only if error < 20%

			correct, _, back, err := auxl.RequestInput(
				ctx,
				func(u *bool) bool {
					return u != nil
				},
				chatID,
				"Send the Word",
				checkWord,
				client,
				updateChan,
			)
			if err != nil {
				return fmt.Errorf("request word: %w", err)
			}
			if back {
				return nil
			}

			if correct != nil && *correct {
				if err = client.Send(chatID, "Correct"); err != nil {
					return err
				}
			} else {
				errorPct := wordErrorPercent(lastIncorrectInput, word.GetWord())
				if errorPct >= secondAttemptErrorThresholdPct {
					if err = client.SendWithParseMode(chatID, fmt.Sprintf("Incorrect, inspect word <code>%s</code> first", word.GetWord()), tg.ModeHTML); err != nil {
						return err
					}
					isAnswerCorrect = false
				} else {
					if err = client.Send(chatID, "Incorrect, try again"); err != nil {
						return err
					}
					correct, _, back, err = auxl.RequestInput(
						ctx,
						func(u *bool) bool {
							return u != nil
						},
						chatID,
						"Send the Word (second attempt)",
						checkWord,
						client,
						updateChan,
					)
					if err != nil {
						return fmt.Errorf("request word second attempt: %w", err)
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
				}
			}
			err = auxl.SendAudioByLanguage(chatID, client, word.GetAudioByLanguage())
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

		err := auxl.SendAudioByLanguage(chatID, client, word.GetAudioByLanguage())
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

	_, _, back, err := auxl.RequestInput(
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

// wordErrorPercent returns how much the input differs from the correct word as 0–100.
// Uses Levenshtein distance; 0 = exact match, 100 = maximally different.
// Used to decide if a wrong answer is "close enough" to allow a second attempt.
func wordErrorPercent(input, correct string) float64 {
	a, b := []rune(strings.ToLower(input)), []rune(strings.ToLower(correct))
	if len(a) == 0 && len(b) == 0 {
		return 0
	}
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	if maxLen == 0 {
		return 0
	}
	d := levenshteinDistance(a, b)
	return float64(d) / float64(maxLen) * 100
}

func levenshteinDistance(a, b []rune) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := 0; j <= len(b); j++ {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min(curr[j-1]+1, min(prev[j]+1, prev[j-1]+cost))
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
}

// shuffleLetters shuffles pairs of letters in a word randomly
func shuffleLetters(word string) string {
	runes := []rune(word)

	// Split into pairs of letters
	var pairs [][]rune
	for i := 0; i < len(runes); i += 2 {
		if i+1 < len(runes) {
			pairs = append(pairs, []rune{runes[i], runes[i+1]})
		} else {
			// Handle odd-length word: keep last letter as single-rune pair
			pairs = append(pairs, []rune{runes[i]})
		}
	}

	// Fisher-Yates shuffle on pairs
	for i := len(pairs) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		pairs[i], pairs[j] = pairs[j], pairs[i]
	}

	// Reconstruct the word from shuffled pairs
	var result []rune
	for _, pair := range pairs {
		result = append(result, pair...)
	}

	return string(result)
}

// maskWord replaces some letters with asterisks based on consecutive correct answers.
// Progressive masking from level 3 to 8:
// - Level 3: shows half the word (masks half)
// - Level 8: shows only 2 letters (masks the rest)
// - Levels 4-7: linear progression between these two points
func maskWord(word string, consecutiveCorrectAnswers uint32) string {
	if word == "" {
		return ""
	}

	runes := []rune(word)
	wordLen := len(runes)

	// Calculate how many letters to reveal based on consecutive correct answers
	var visible int

	// At level 3: show half the word (rounded up)
	visibleAtLevel3 := (wordLen + 1) / 2
	// At level 8: show only 2 letters (or all if word is shorter)
	visibleAtLevel8 := 2
	if wordLen < 2 {
		visibleAtLevel8 = wordLen
	}

	// Handle edge case: if word is too short for progressive masking
	if visibleAtLevel8 >= visibleAtLevel3 {
		// Word is too short (e.g., 2-4 letters), just show half
		visible = visibleAtLevel3
	} else {
		// Linear interpolation between level 3 and level 8
		// visible = visibleAtLevel3 - (level - 3) * (visibleAtLevel3 - visibleAtLevel8) / 5
		level := int(consecutiveCorrectAnswers)
		if level < 3 {
			level = 3
		} else if level > 8 {
			level = 8
		}

		visible = visibleAtLevel3 - (level-3)*(visibleAtLevel3-visibleAtLevel8)/5

		// Ensure we don't reveal more than available or less than minimum
		if visible > wordLen {
			visible = wordLen
		} else if visible < visibleAtLevel8 {
			visible = visibleAtLevel8
		}
	}

	// Calculate how many letters to mask
	masked := wordLen - visible

	// If nothing to mask, return the word as-is
	if masked <= 0 {
		return word
	}

	// Create a slice to track which positions to mask
	positions := make([]int, wordLen)
	for i := range positions {
		positions[i] = i
	}

	// Shuffle positions using Fisher-Yates
	for i := len(positions) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		positions[i], positions[j] = positions[j], positions[i]
	}

	// Mark first 'masked' positions for masking
	maskSet := make(map[int]bool)
	for i := 0; i < masked; i++ {
		maskSet[positions[i]] = true
	}

	// Build result
	result := make([]rune, wordLen)
	for i, r := range runes {
		if maskSet[i] {
			result[i] = '*'
		} else {
			result[i] = r
		}
	}

	return string(result)
}

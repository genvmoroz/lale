package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/genvmoroz/lale/service/api"
	"github.com/liamg/clinch/prompt"
	"github.com/liamg/clinch/task"
	"github.com/liamg/tml"
	"github.com/samber/lo"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	lo.Must0(tml.Printf("<green><bold>Rewrite cards tool</bold></green>\n"))

	fmt.Println("")

	conn, err := buildDepsTask(ctx)
	if err != nil {
		exitWithError("build deps", err, 1)
	}

	cards, err := fetchCardsTask(ctx, conn)
	if err != nil {
		exitWithError("fetch cards", err, 2)
	}

	if err = updateCardsTask(ctx, lo.SliceToChannel(0, cards), conn); err != nil {
		exitWithError("rewrite cards", err, 3)
	}

	lo.Must0(tml.Printf("<green><bold>Done</bold></green>\n"))
}

func buildDepsTask(ctx context.Context) (api.LaleServiceClient, error) {
	host, port, err := askForLaleServiceAddr()
	if err != nil {
		return nil, fmt.Errorf("ask for Lale service addr: %w", err)
	}

	var conn api.LaleServiceClient
	err = task.New(
		"build deps",
		"connecting to Lale gRPC service...",
		func(t *task.Task) error {
			conn, err = connectToLaleService(ctx, host, port, time.Second*5)
			return err
		},
	).Run()
	if err != nil {
		return nil, fmt.Errorf("connect to Lale gRPC service: %w", err)
	}

	return conn, nil
}

func fetchCardsTask(ctx context.Context, conn api.LaleServiceClient) ([]*api.Card, error) {
	var err error

	req := &api.GetCardsRequest{}
	req.UserID, req.Language = askForUserIDAndLanguage()

	startCardID := prompt.EnterInput("If you want to jump to a specific card ID, enter it here (empty to skip): ")

	var resp *api.GetCardsResponse
	err = task.New(
		"get cards",
		"fetching...",
		func(t *task.Task) error {
			resp, err = conn.GetAllCards(ctx, req)
			return err
		},
	).Run()
	if err != nil {
		return nil, fmt.Errorf("fetch cards: %w", err)
	}

	cards := resp.GetCards()

	if len(startCardID) != 0 {
		_, index, found := lo.FindIndexOf(cards,
			func(card *api.Card) bool {
				return card != nil && card.GetId() == startCardID
			},
		)
		if !found {
			return nil, fmt.Errorf("card with ID %s not found", startCardID)
		}
		cards = cards[index:]
	}

	if err = tml.Printf("<green><bold>Found %d cards</bold></green>\n", countExcludingEmpty(cards)); err != nil {
		return nil, fmt.Errorf("tml: print cards count: %w", err)
	}

	return cards, nil
}

func updateCardsTask(ctx context.Context, cardCh <-chan *api.Card, conn api.LaleServiceClient) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case card, ok := <-cardCh:
			if !ok {
				return nil
			}

			if err := tml.Printf("<yellow><bold>CardID: %s</bold></yellow>\n", card.GetId()); err != nil {
				return fmt.Errorf("tml: print card ID: %w", err)
			}

			changed, err := askToChangeWordTranslations(ctx, card)
			if err != nil {
				return fmt.Errorf("ask to change word translations: %w", err)
			}
			if !changed {
				if err = tml.Printf("<yellow><bold>CardID: %s skipped</bold></yellow>\n\n", card.GetId()); err != nil {
					return fmt.Errorf("tml: print card ID skipped: %w", err)
				}
				continue
			}

			tsk := task.New(
				"update card",
				fmt.Sprintf("with ID %s", card.GetId()),
				func(t *task.Task) error {
					return tryToUpdateCard(ctx, card, conn)
				},
			)

			if err = tsk.Run(); err != nil {
				return fmt.Errorf("try to update card: %w", err)
			}
			if err = tml.Printf("<yellow><bold>%d cards left</bold></yellow>\n", len(cardCh)); err != nil {
				return fmt.Errorf("tml: print cards left: %w", err)
			}
		}
	}
}

func askToChangeWordTranslations(ctx context.Context, card *api.Card) (bool, error) {
	if card == nil {
		return false, nil
	}

	var changed bool
	for i, word := range card.GetWordInformationList() {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}

		if err := tml.Printf("<magenta><bold>%s - %s</bold></magenta>\n", word.GetWord(), translationsToString(word.GetTranslation())); err != nil {
			return false, fmt.Errorf("tml: print word and translation: %w", err)
		}

		translation := prompt.EnterInput("Enter new word translations (ex. translation1, translation2. Empty to skip current word): ")
		if len(strings.TrimSpace(translation)) == 0 {
			continue
		}

		card.WordInformationList[i].Translation = &api.Translation{
			Language:     card.WordInformationList[i].Translation.Language,
			Translations: parseTranslations(translation),
		}
		changed = true
	}

	return changed, nil
}

func tryToUpdateCard(ctx context.Context, card *api.Card, conn api.LaleServiceClient) error {
	if card == nil {
		return nil
	}

	req := &api.UpdateCardRequest{
		UserID:              card.GetUserID(),
		CardID:              card.GetId(),
		WordInformationList: card.GetWordInformationList(),
		Params: &api.Parameters{
			EnrichWordInformationFromDictionary: true,
		},
	}
	_, err := conn.UpdateCard(ctx, req)
	if err != nil {
		if !strings.Contains(err.Error(), "enrich card words from dictionary") {
			return fmt.Errorf("update card with the param to enrich word from dictionary: %w", err)
		}
		req.Params.EnrichWordInformationFromDictionary = false
		_, err = conn.UpdateCard(ctx, req)
		if err != nil {
			return fmt.Errorf("update card without the param to enrich word from dictionary: %w", err)
		}
	}

	return nil
}

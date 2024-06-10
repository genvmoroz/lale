package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/liamg/clinch/task"
	"update-cards-from-mongo-script/mongo"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg := mongo.Config{
		Protocol: "mongodb+srv",
		Host:     "dictionary.zxrao.mongodb.net",
		Params: map[string]string{
			"retryWrites": "true",
			"w":           "majority",
		},
		Database:   "dictionary",
		Collection: "cards",
		Creds: mongo.Creds{
			User: "genvmoroz",
			Pass: "",
		},
	}
	repo, err := mongo.NewRepo(ctx, cfg)
	if err != nil {
		log.Fatalf("create mongo repo: %v", err)
	}

	cards, err := repo.GetCardsForUser(ctx, "gennadiymoroz")
	if err != nil {
		log.Fatalf("get cards for user: %v", err)
	}

	updateCard := func(card *mongo.Card) error {
		return nil
	}

	if err = updateCards(ctx, updateCard, cards, repo); err != nil {
		log.Fatalf("update cards: %v", err)
	}
}

func updateCards(ctx context.Context, update func(card *mongo.Card) error, cards []mongo.Card, repo *mongo.Repo) error {
	if len(cards) == 0 {
		return fmt.Errorf("nothing to update")
	}
	updated := 0

	for _, card := range cards {
		card := card
		if err := update(&card); err != nil {
			return err
		}

		tsk := task.New(
			"update card",
			fmt.Sprintf("with ID %s", card.ID),
			func(t *task.Task) error {
				return repo.SaveCards(ctx, []mongo.Card{card})
			},
		)

		if err := tsk.Run(); err != nil {
			return fmt.Errorf("update card: %w", err)
		}
		updated++

		left := len(cards) - updated
		progressPercentage := 100 - (float32(left) / float32(len(cards)) * 100)
		log.Printf("%d cards left (Progress=%.1f%%)\n", left, progressPercentage)
	}

	return nil
}

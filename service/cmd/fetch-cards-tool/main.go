package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/genvmoroz/lale/service/api"
	"github.com/liamg/clinch/task"
	"github.com/liamg/tml"
	"github.com/samber/lo"
	"slices"
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

	log.Println("cards", len(cards))

	// go over all unlearnt cards
	cards = lo.Filter(cards, func(item *api.Card, index int) bool {
		return item.GetNextDueDate().AsTime() == time.Time{}
	})
	for _, card := range slices.Backward(cards) {
		fmt.Println()
		for _, word := range card.WordInformationList {
			fmt.Println(word.Word)
		}
		fmt.Println(card.GetId())
		fmt.Println("=====================================")
	}

	// finds duplicates
	for i := 0; i < len(cards); i++ {
		for j := i + 1; j < len(cards); j++ {
			for _, word := range cards[i].WordInformationList {
				if slices.ContainsFunc(cards[j].WordInformationList, func(item *api.WordInformation) bool {
					return strings.EqualFold(item.Word, word.Word)
				}) {
					fmt.Println(cards[i].GetId())
					fmt.Println(cards[i].WordInformationList[0].Word)
				}
			}
		}
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

	if err = tml.Printf("<green><bold>Found %d cards</bold></green>\n", len(cards)); err != nil {
		return nil, fmt.Errorf("tml: print cards count: %w", err)
	}

	return cards, nil
}

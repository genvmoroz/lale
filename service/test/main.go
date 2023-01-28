package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/genvmoroz/lale-service/api"
	"github.com/genvmoroz/lale-service/internal/entity"
	"github.com/genvmoroz/lale-service/pkg/lang"
	"github.com/genvmoroz/lale-service/test/client"
	"github.com/genvmoroz/lale-service/test/comparator"
	"github.com/genvmoroz/lale-service/test/options"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := options.FromEnv()
	if err != nil {
		log.Fatalf("failed to read envs: %s", err.Error())
	}

	user, err := NewUser(ctx, cfg.ClientConfig, 0)
	if err != nil {
		log.Fatalf("failed to read envs: %s", err.Error())
	}

	user.InspectCards(ctx, status.Error(codes.NotFound, "no card found"), false, nil)

	log.Println("testing finished")
}

type User struct {
	id    string
	cli   *client.Client
	words []WordPack
}

type WordPack struct {
	words    []entity.WordInformation
	language lang.Language
}

func (u *User) InspectCards(ctx context.Context, expErr error, skipComparison bool, expCard *entity.Card) {
	for _, pack := range u.words {
		for _, word := range pack.words {
			req := &api.InspectCardRequest{
				UserID:   u.id,
				Language: string(pack.language),
				Word:     word.Word,
			}
			card, err := u.cli.InspectCard(ctx, req)
			if err != nil && !errors.Is(expErr, err) {
				log.Fatalf("unexpected error: %s", err.Error())
			}
			if !skipComparison {
				if !comparator.NewGRPCComparator().CompareCard(expCard, card.GetCard()) {
					log.Fatalf("values are not equal")
				}
			}
		}
	}
}

func NewUser(ctx context.Context, cfg client.Config, count int) (User, error) {
	cli, err := client.NewClient(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to create client: %s", err.Error())
	}

	return User{
		id:  fmt.Sprintf("testID%d", count),
		cli: cli,
		words: []WordPack{
			{
				words: []entity.WordInformation{
					{Word: "anticipation", Translate: &entity.Translate{Language: "uk", Translates: []string{"очікування"}}},
					{Word: "anticipate", Translate: &entity.Translate{Language: "uk", Translates: []string{"передбачити", "очікувати", "передчувати"}}},
					{Word: "anticipated", Translate: &entity.Translate{Language: "uk", Translates: []string{"очікуваний"}}},
				},
				language: "en",
			},
			{
				words: []entity.WordInformation{
					{Word: "stir", Translate: &entity.Translate{Language: "uk", Translates: []string{"перемішати", "замішувати", "ворушіння", "метушня", "розмішування"}}},
				},
				language: "en",
			},
			{
				words: []entity.WordInformation{
					{Word: "spread", Translate: &entity.Translate{Language: "uk", Translates: []string{"поширювати", "поширення", "розкидати", "поширюватися"}}},
				},
				language: "en",
			},
			{
				words: []entity.WordInformation{
					{Word: "restrict", Translate: &entity.Translate{Language: "uk", Translates: []string{"обмеження", "обмежуватися", "засекречувати", "забороняти", "тримати в певних межах"}}},
					{Word: "restrictor", Translate: &entity.Translate{Language: "uk", Translates: []string{"обмежувач"}}},
					{Word: "restricted", Translate: &entity.Translate{Language: "uk", Translates: []string{"обмежений", "для службового користування"}}},
					{Word: "restrictively", Translate: &entity.Translate{Language: "uk", Translates: []string{"обмежено"}}},
				},
				language: "en",
			},
			{
				words: []entity.WordInformation{
					{Word: "object", Translate: &entity.Translate{Language: "uk", Translates: []string{"предмет", "об'єкт", "безглузда річ", "висловлювати несхвалення", "заперечувати", "протестувати"}}},
					{Word: "objective", Translate: &entity.Translate{Language: "uk", Translates: []string{"об'єктивний", "предметний", "дійсний", "мета"}}},
					{Word: "objectively", Translate: &entity.Translate{Language: "uk", Translates: []string{"об'єктивно", "неупереджено", "реально"}}},
				},
				language: "en",
			},
		},
	}, nil
}

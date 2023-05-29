package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/genvmoroz/lale/service/api"
	"github.com/genvmoroz/lale/service/pkg/entity"
	"github.com/genvmoroz/lale/service/test/client"
	"github.com/genvmoroz/lale/service/test/options"
	"golang.org/x/text/language"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := options.FromEnv()
	if err != nil {
		log.Fatalf("read envs: %s", err.Error())
	}

	user, err := NewUser(ctx, cfg.ClientConfig, 0)
	if err != nil {
		log.Fatalf("new user: %s", err.Error())
	}

	user.InspectCards(ctx, status.Error(codes.NotFound, "no card found"))

	log.Println("testing finished")
}

type User struct {
	id    string
	cli   *client.Client
	words []WordPack
}

type WordPack struct {
	words []entity.WordInformation
	lang  language.Tag
}

func (u *User) InspectCards(ctx context.Context, expErr error) {
	for _, pack := range u.words {
		for _, word := range pack.words {
			req := &api.InspectCardRequest{
				UserID:   u.id,
				Language: pack.lang.String(),
				Word:     word.Word,
			}
			_, err := u.cli.MustDo().InspectCard(ctx, req)
			if err != nil && !errors.Is(expErr, err) {
				log.Fatalf("unexpected error: %s", err.Error())
			}
		}
	}
}

func NewUser(ctx context.Context, cfg client.Config, count int) (User, error) {
	cli, err := client.NewClient(ctx, cfg)
	if err != nil {
		log.Fatalf("create client: %s", err.Error())
	}

	return User{
		id:  fmt.Sprintf("testID%d", count),
		cli: cli,
		words: []WordPack{
			{
				words: []entity.WordInformation{
					{Word: "anticipation", Translation: &entity.Translation{Language: language.Ukrainian, Translations: []string{"очікування"}}},
					{Word: "anticipate", Translation: &entity.Translation{Language: language.Ukrainian, Translations: []string{"передбачити", "очікувати", "передчувати"}}},
					{Word: "anticipated", Translation: &entity.Translation{Language: language.Ukrainian, Translations: []string{"очікуваний"}}},
				},
				lang: language.English,
			},
			{
				words: []entity.WordInformation{
					{Word: "stir", Translation: &entity.Translation{Language: language.Ukrainian, Translations: []string{"перемішати", "замішувати", "ворушіння", "метушня", "розмішування"}}},
				},
				lang: language.English,
			},
			{
				words: []entity.WordInformation{
					{Word: "spread", Translation: &entity.Translation{Language: language.Ukrainian, Translations: []string{"поширювати", "поширення", "розкидати", "поширюватися"}}},
				},
				lang: language.English,
			},
			{
				words: []entity.WordInformation{
					{Word: "restrict", Translation: &entity.Translation{Language: language.Ukrainian, Translations: []string{"обмеження", "обмежуватися", "засекречувати", "забороняти", "тримати в певних межах"}}},
					{Word: "restrictor", Translation: &entity.Translation{Language: language.Ukrainian, Translations: []string{"обмежувач"}}},
					{Word: "restricted", Translation: &entity.Translation{Language: language.Ukrainian, Translations: []string{"обмежений", "для службового користування"}}},
					{Word: "restrictively", Translation: &entity.Translation{Language: language.Ukrainian, Translations: []string{"обмежено"}}},
				},
				lang: language.English,
			},
			{
				words: []entity.WordInformation{
					{Word: "object", Translation: &entity.Translation{Language: language.Ukrainian, Translations: []string{"предмет", "об'єкт", "безглузда річ", "висловлювати несхвалення", "заперечувати", "протестувати"}}},
					{Word: "objective", Translation: &entity.Translation{Language: language.Ukrainian, Translations: []string{"об'єктивний", "предметний", "дійсний", "мета"}}},
					{Word: "objectively", Translation: &entity.Translation{Language: language.Ukrainian, Translations: []string{"об'єктивно", "неупереджено", "реально"}}},
				},
				lang: language.English,
			},
		},
	}, nil
}

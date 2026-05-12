package create_card

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/genvmoroz/lale/service/api"
	"github.com/genvmoroz/lale/service/test/stress/loader/internal/core"
	"github.com/genvmoroz/lale/service/test/stress/loader/internal/repository"
	"golang.org/x/text/language"
)

type Builder struct{}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) New(cfg core.PerformerConfig) (core.Performer, error) {
	laleRepo, err := repository.NewLaleRepo(
		repository.LaleRepoConfig{
			Host:    cfg.LaleServiceHost,
			Port:    cfg.LaleServicePort,
			Timeout: 10 * time.Second,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create lale repository: %w", err)
	}

	return &Performer{laleRepo: laleRepo}, nil
}

type Performer struct {
	laleRepo *repository.LaleRepo
}

func (p *Performer) Perform(ctx context.Context, env *core.Environment) error {
	return p.performCardsCreationForUsers(ctx, env.Users)
}

func (p *Performer) performCardsCreationForUsers(ctx context.Context, users []core.User) error {
	if len(users) == 0 {
		return nil
	}

	innerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg := &sync.WaitGroup{}
	errChan := make(chan error, len(users))

	for i := range users {
		wg.Add(1)
		go func(user *core.User) {
			defer wg.Done()

			if err := p.createCards(innerCtx, user); err != nil {
				errChan <- fmt.Errorf("create cards for user %s: %w", user.Name, err)
			}
		}(&users[i])
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err, ok := <-errChan:
		if ok {
			return err
		}
	}

	return nil
}

func (p *Performer) createCards(ctx context.Context, user *core.User) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	for i := range user.Cards {
		req := &api.CreateCardRequest{
			UserID:              user.Name,
			Language:            language.English.String(),
			WordInformationList: wordsToWordInformationList(user.Cards[i].Words),
		}
		resp, err := p.laleRepo.Client.CreateCard(ctx, req)
		if err != nil {
			return fmt.Errorf("create card: %w", err)
		}
		user.Cards[i].ID = resp.GetId()
	}

	return nil
}

// todo: try to change this loop to another one and compare the results
func wordsToWordInformationList(words []core.Word) []*api.WordInformation {
	wordInformationList := make([]*api.WordInformation, 0, len(words))
	for i := range words {
		wordInformationList = append(wordInformationList,
			&api.WordInformation{
				Word:        words[i].Word,
				Translation: toTranslation(words[i].Translation),
			},
		)
	}
	return wordInformationList
}

func toTranslation(translation []string) *api.Translation {
	return &api.Translation{
		Language:     language.English.String(),
		Translations: translation,
	}
}

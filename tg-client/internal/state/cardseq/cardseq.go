package cardseq

import (
	"context"
	"math/rand"
	"time"

	"github.com/genvmoroz/lale-tg-client/internal/repository"
	"github.com/genvmoroz/lale/service/api"
	"github.com/genvmoroz/lale/service/pkg/future"
	"github.com/sirupsen/logrus"
)

type (
	Cards struct {
		index uint32
		cards []Card

		cardsNumberToBeEnrichedWithSentencesInAdvance uint32
		sentencesCount                                uint32

		laleRepo *repository.LaleRepo
	}

	Card struct {
		Card      *api.Card
		Words     []*api.WordInformation
		Sentences map[string]future.Task[[]string]
	}
)

func NewCards(
	ctx context.Context,
	laleRepo *repository.LaleRepo,
	resp *api.GetCardsResponse,
	cardsNumberToBeEnrichedWithSentencesInAdvance uint32,
	sentencesCount uint32,
) Cards {
	cards := make([]Card, 0, len(resp.GetCards()))
	for _, c := range resp.GetCards() {
		if c == nil {
			continue
		}
		cards = append(cards,
			Card{
				Card:      c,
				Words:     c.GetWordInformationList(),
				Sentences: make(map[string]future.Task[[]string]),
			},
		)
	}
	res := Cards{
		index: 0,
		cards: cards,
		cardsNumberToBeEnrichedWithSentencesInAdvance: cardsNumberToBeEnrichedWithSentencesInAdvance,
		sentencesCount: sentencesCount,
		laleRepo:       laleRepo,
	}

	res.enrichCardsWithSentences(ctx)

	return res
}

func (r *Cards) Next(ctx context.Context) Card {
	if r.index >= uint32(len(r.cards)) {
		return Card{}
	}

	next := r.cards[r.index]
	r.index++

	r.enrichCardsWithSentences(ctx)

	return next
}

func (r *Cards) HasNext() bool {
	return r.index < uint32(len(r.cards))
}

func (r *Cards) enrichCardsWithSentences(ctx context.Context) {
	for i := 0; i < int(r.cardsNumberToBeEnrichedWithSentencesInAdvance); i++ {
		r.enrichCardWithSentences(ctx, r.index+uint32(i))
	}
}

func (r *Cards) enrichCardWithSentences(ctx context.Context, i uint32) {
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
				SentencesCount: r.sentencesCount,
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

				logrus.Errorf("failed to get sentences for word %s: %v", word.GetWord(), err)
				sleepDuration := time.Duration(rand.Intn(10)+5) * time.Second
				logrus.Infof("retrying in %s", sleepDuration)
				time.Sleep(sleepDuration)
			}

			return nil, err
		}

		r.cards[i].Sentences[word.GetWord()] = future.NewTask(ctx, run)
	}
}

func (r *Cards) Remaining() int {
	if int(r.index) >= len(r.cards) {
		return 0
	}

	return len(r.cards) - int(r.index)
}

package core

import (
    "slices"
    "testing"

    "math/rand/v2"

    "github.com/genvmoroz/lale/service/pkg/entity"
)

func cloneCards(cards []entity.Card) []entity.Card {
    out := make([]entity.Card, len(cards))
    copy(out, cards)
    return out
}

func makeCards(n int) []entity.Card {
    cards := make([]entity.Card, 0, n)
    for i := 1; i <= n; i++ {
        cards = append(cards, entity.Card{ID: string(rune('a' + i - 1)), ConsecutiveCorrectAnswersNumber: uint32(i)})
    }
    return cards
}

func TestSortShuffle_Determinism(t *testing.T) {
    t.Parallel()

    base := makeCards(10)

    tests := []struct {
        name      string
        chunkSize uint8
        seed      uint64
    }{
        {name: "single chunk", chunkSize: 10, seed: 1},
        {name: "multiple chunks", chunkSize: 3, seed: 42},
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()

            cardsA := cloneCards(base)
            cardsB := cloneCards(base)

            rngA := rand.New(rand.NewPCG(tt.seed, 0))
            rngB := rand.New(rand.NewPCG(tt.seed, 0))

            sortByConsecutiveCorrectAnswersAndShuffleInChunks(cardsA, tt.chunkSize, rngA)
            sortByConsecutiveCorrectAnswersAndShuffleInChunks(cardsB, tt.chunkSize, rngB)

            if !slices.EqualFunc(cardsA, cardsB, func(a, b entity.Card) bool { return a.ID == b.ID }) {
                t.Fatalf("expected deterministic order with same seed; got different")
            }
        })
    }
}

func TestSortShuffle_DifferentSeedsYieldDifferentOrder(t *testing.T) {
    t.Parallel()

    base := makeCards(10)

    tests := []struct {
        name      string
        chunkSize uint8
        seedA     uint64
        seedB     uint64
    }{
        {name: "single chunk", chunkSize: 10, seedA: 1, seedB: 2},
        {name: "multiple chunks", chunkSize: 4, seedA: 7, seedB: 8},
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            cardsA := cloneCards(base)
            cardsB := cloneCards(base)
            sortByConsecutiveCorrectAnswersAndShuffleInChunks(cardsA, tt.chunkSize, rand.New(rand.NewPCG(tt.seedA, 0)))
            sortByConsecutiveCorrectAnswersAndShuffleInChunks(cardsB, tt.chunkSize, rand.New(rand.NewPCG(tt.seedB, 0)))

            if slices.EqualFunc(cardsA, cardsB, func(a, b entity.Card) bool { return a.ID == b.ID }) {
                t.Fatalf("expected different order with different seeds; got same")
            }
        })
    }
}

func TestSortShuffle_ChunkMembershipPreserved(t *testing.T) {
    t.Parallel()

    base := makeCards(10)
    chunkSize := uint8(3)

    // Expected sorted (desc by ConsecutiveCorrectAnswersNumber)
    expected := cloneCards(base)
    slices.SortFunc(expected, func(a, b entity.Card) int {
        if a.ConsecutiveCorrectAnswersNumber == b.ConsecutiveCorrectAnswersNumber {
            return 0
        }
        if a.ConsecutiveCorrectAnswersNumber < b.ConsecutiveCorrectAnswersNumber {
            return 1
        }
        return -1
    })

    got := cloneCards(base)
    sortByConsecutiveCorrectAnswersAndShuffleInChunks(got, chunkSize, rand.New(rand.NewPCG(123, 0)))

    for i := 0; i < len(got); i += int(chunkSize) {
        end := i + int(chunkSize)
        if end > len(got) {
            end = len(got)
        }

        expChunk := expected[i:end]
        gotChunk := got[i:end]

        // Compare membership using IDs as set
        if !sameIDSet(expChunk, gotChunk) {
            t.Fatalf("chunk membership differs at range [%d:%d]", i, end)
        }
    }
}

func TestSortShuffle_ChunkSizeBelowOneTreatedAsOne(t *testing.T) {
    t.Parallel()

    base := makeCards(6)

    expected := cloneCards(base)
    slices.SortFunc(expected, func(a, b entity.Card) int {
        if a.ConsecutiveCorrectAnswersNumber == b.ConsecutiveCorrectAnswersNumber {
            return 0
        }
        if a.ConsecutiveCorrectAnswersNumber < b.ConsecutiveCorrectAnswersNumber {
            return 1
        }
        return -1
    })

    got := cloneCards(base)
    sortByConsecutiveCorrectAnswersAndShuffleInChunks(got, 0, rand.New(rand.NewPCG(999, 0)))

    if !slices.EqualFunc(expected, got, func(a, b entity.Card) bool { return a.ID == b.ID }) {
        t.Fatalf("expected pure sort when chunkSize<1; got different order")
    }
}

func sameIDSet(a, b []entity.Card) bool {
    if len(a) != len(b) {
        return false
    }
    m := make(map[string]int, len(a))
    for _, c := range a {
        m[c.ID]++
    }
    for _, c := range b {
        m[c.ID]--
    }
    for _, v := range m {
        if v != 0 {
            return false
        }
    }
    return true
}

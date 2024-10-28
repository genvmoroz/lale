package core

import (
	"fmt"
	"sync"
)

func enabledActions(req LoadRequest) []Action {
	actions := make([]Action, 0, 12)
	switch {
	case req.ActionCreateCardEnabled:
		actions = append(actions, ActionCreateCard)
	case req.ActionInspectCardEnabled:
		actions = append(actions, ActionInspectCard)
	case req.ActionPromptCardEnabled:
		actions = append(actions, ActionPromptCard)
	case req.ActionGetAllCardsEnabled:
		actions = append(actions, ActionGetAllCards)
	case req.ActionUpdateCardEnabled:
		actions = append(actions, ActionUpdateCard)
	case req.ActionUpdateCardPerformanceEnabled:
		actions = append(actions, ActionUpdateCardPerformance)
	case req.ActionGetCardsToLearnEnabled:
		actions = append(actions, ActionGetCardsToLearn)
	case req.ActionGetCardsToRepeatEnabled:
		actions = append(actions, ActionGetCardsToRepeat)
	case req.ActionGetSentencesEnabled:
		actions = append(actions, ActionGetSentences)
	case req.ActionGenerateStoryEnabled:
		actions = append(actions, ActionGenerateStory)
	case req.ActionDeleteCardEnabled:
		actions = append(actions, ActionDeleteCard)
	}

	return actions
}

func generateUsersInParallel(n uint32, workersCount uint32, cardsPerUser uint32, wordsPerCard uint32) []User {
	result := make([]User, 0, n)

	if n < workersCount {
		workersCount = n
	}

	usersChan := make(chan []User, workersCount)
	wg := &sync.WaitGroup{}
	usersPerWorker := n / workersCount
	remainder := n % workersCount

	// Launch workers
	for i := uint32(0); i < workersCount; i++ {
		wg.Add(1)
		go func(workerID uint32) {
			defer wg.Done()

			usersCount := usersPerWorker
			if workerID < remainder {
				usersCount++
			}

			usersChan <- generateUsers(workerID, usersCount, cardsPerUser, wordsPerCard)
		}(i)
	}

	// Close the channel once all workers are done
	go func() {
		wg.Wait()
		close(usersChan)
	}()

	// Collect results
	for users := range usersChan {
		result = append(result, users...)
	}

	return result
}

func generateUsers(workerID uint32, usersNumber uint32, cardsNumber uint32, wordsNumber uint32) []User {
	users := make([]User, 0, usersNumber)
	for n := range usersNumber {
		users = append(users, generateUser(workerID, n, cardsNumber, wordsNumber))
	}
	return users
}

func generateUser(workerID uint32, n uint32, cardsNumber uint32, wordsNumber uint32) User {
	return User{
		Name:  fmt.Sprintf("worker-%d-user-%d", workerID, n),
		Cards: generateCards(cardsNumber, wordsNumber),
	}
}

func generateCards(cardsNumber uint32, wordsNumber uint32) []Card {
	cards := make([]Card, 0, cardsNumber)
	for n := range cardsNumber {
		cards = append(cards, generateCard(n, wordsNumber))
	}
	return cards
}

func generateCard(cardN uint32, wordsNumber uint32) Card {
	return Card{
		Words: generateWords(cardN, wordsNumber),
	}
}

func generateWords(cardN uint32, wordsNumber uint32) []Word {
	words := make([]Word, 0, wordsNumber)
	for n := range wordsNumber {
		words = append(words, generateWord(cardN, n))
	}
	return words
}

func generateWord(cardN uint32, n uint32) Word {
	return Word{
		Word:        fmt.Sprintf("card-%d-word-%d", cardN, n),
		Translation: []string{"1", "2", "3", "4", "5"},
	}
}

package core

import (
	"context"
	"sync"

	"github.com/genvmoroz/bot-engine/bot"
	"github.com/genvmoroz/bot-engine/dispatcher"
)

type Service struct {
	dispatcher *dispatcher.Dispatcher
}

func NewService(baseBot *bot.Client, stateProvider dispatcher.StateProvider, timeout uint) (*Service, error) {
	baseDispatcher, err := dispatcher.New(baseBot, stateProvider, timeout)
	if err != nil {
		return nil, err
	}

	return &Service{dispatcher: baseDispatcher}, nil
}

func (s *Service) Run(ctx context.Context, wg *sync.WaitGroup) error {
	return s.dispatcher.Dispatch(ctx, wg, -1, 128)
}

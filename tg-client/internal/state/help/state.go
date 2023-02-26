package help

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/genvmoroz/bot-engine/bot"
	"github.com/genvmoroz/bot-engine/processor"
)

type State struct {
	*bot.Client
	states []processor.StateProcessor
}

const Command = "/help"

func NewState(client *bot.Client, states []processor.StateProcessor) *State {
	return &State{Client: client, states: states}
}

func (s *State) Process(ctx context.Context, updateChan bot.UpdatesChannel) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case update, ok := <-updateChan:
			if !ok {
				return errors.New("updateChan is closed")
			}
			if err := s.process(update.Message.Chat.ID); err != nil {
				return err
			}
		}
	}
}

func (s *State) process(chatID int64) error {
	b := strings.Builder{}

	for _, state := range s.states {
		b.WriteString(state.Command())
		b.WriteString(" - ")
		b.WriteString(state.Description())
	}

	if err := s.Send(chatID, b.String()); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (s *State) Command() string {
	return Command
}

func (s *State) Description() string {
	return "shows available commands"
}

package help

import (
	"context"
	"fmt"
	"github.com/genvmoroz/bot-engine/tg"
	"strings"

	"github.com/genvmoroz/bot-engine/processor"
)

type State struct {
	states []processor.StateProcessor
}

const Command = "/help"

func NewState(states []processor.StateProcessor) *State {
	return &State{states: states}
}

func (s *State) Process(ctx context.Context, client processor.Client, chatID int64, _ tg.UpdatesChannel) error {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	b := strings.Builder{}

	for _, state := range s.states {
		b.WriteString(state.Command())
		b.WriteString(" - ")
		b.WriteString(fmt.Sprintf("%s\n", state.Description()))
	}

	b.WriteString(fmt.Sprintf("ChatID: %d", chatID))

	if err := client.Send(chatID, b.String()); err != nil {
		return fmt.Errorf("send: %w", err)
	}

	return nil
}

func (s *State) Command() string {
	return Command
}

func (s *State) Description() string {
	return "Shows available commands"
}

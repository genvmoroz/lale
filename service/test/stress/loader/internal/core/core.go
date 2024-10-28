package core

import (
	"context"
	"fmt"
	"runtime"
	"slices"
	"time"
)

type (
	PerformerBuilder interface {
		New(cfg PerformerConfig) (Performer, error)
	}

	Performer interface {
		Perform(ctx context.Context, env *Environment) error
	}
)

type Loader struct {
	builders map[Action]PerformerBuilder
}

func NewLoader(builders map[Action]PerformerBuilder) (*Loader, error) {
	if len(builders) == 0 {
		return nil, fmt.Errorf("no performer builders provided")
	}
	return &Loader{builders: builders}, nil
}

func (l *Loader) Load(ctx context.Context, req LoadRequest) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("validate load request: %w", err)
	}

	actions := enabledActions(req)
	// Sort actions to ensure that they are performed in their logical order.
	// Their logical order is defined by the order in which they are defined in the Action enum.
	slices.Sort(actions)

	env := l.setupEnvironment(req)

	cfg := PerformerConfig{
		LaleServiceHost: req.LaleServiceHost,
		LaleServicePort: req.LaleServicePort,
	}
	for action := range slices.Values(actions) {
		if err := l.perform(ctx, action, cfg, env); err != nil {
			return err
		}
	}

	return nil
}

func (l *Loader) perform(ctx context.Context, action Action, cfg PerformerConfig, env *Environment) error {
	builder, ok := l.builders[action]
	if !ok {
		return fmt.Errorf("performer builder for action %s not found", action)
	}

	performer, err := builder.New(cfg)
	if err != nil {
		return fmt.Errorf("create performer for action %s: %w", action, err)
	}

	fmt.Println("Performing action", action)
	now := time.Now()
	if err = performer.Perform(ctx, env); err != nil {
		return fmt.Errorf("perform action %s: %w", action, err)
	}
	fmt.Println("Action", action, "completed in", time.Since(now))

	return nil
}

func (l *Loader) setupEnvironment(req LoadRequest) *Environment {
	fmt.Println("Setup environment, may take some time, please wait...")
	env := &Environment{
		Users: generateUsersInParallel(req.ParallelUsers, uint32(runtime.NumCPU()), req.CardsPerUser, req.WordsPerCard),
	}
	fmt.Println("Environment setup complete")
	fmt.Println("Environment:")
	fmt.Printf("  Parallel users: %d\n", len(env.Users))

	return env
}

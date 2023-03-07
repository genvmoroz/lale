package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/genvmoroz/bot-engine/bot"
	"github.com/genvmoroz/bot-engine/dispatcher"
	"github.com/genvmoroz/bot-engine/processor"
	"github.com/genvmoroz/lale/tg-client/internal/options"
	"github.com/genvmoroz/lale/tg-client/internal/repository"
	createstate "github.com/genvmoroz/lale/tg-client/internal/state/create"
	deletestate "github.com/genvmoroz/lale/tg-client/internal/state/delete"
	getallstate "github.com/genvmoroz/lale/tg-client/internal/state/getall"
	helpstate "github.com/genvmoroz/lale/tg-client/internal/state/help"
	inspectstate "github.com/genvmoroz/lale/tg-client/internal/state/inspect"
	reviewstate "github.com/genvmoroz/lale/tg-client/internal/state/review"
	"github.com/sirupsen/logrus"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("recovered: %s", r)
		}
	}()

	if err := launch(); err != nil {
		log.Fatalf("service error: %s", err.Error())
	}
}

func launch() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := options.FromEnv()
	if err != nil {
		return fmt.Errorf("failed to read env: %w", err)
	}

	logrus.SetLevel(cfg.LogLevel)

	baseBot, err := bot.NewClient(cfg.TelegramToken)
	if err != nil {
		return err
	}

	clientCfg := repository.ClientConfig{
		Host:    cfg.LaleService.Host,
		Port:    cfg.LaleService.Port,
		Timeout: cfg.LaleService.Timeout,
	}
	laleRepo, err := repository.NewLaleRepo(ctx, clientCfg)
	if err != nil {
		return fmt.Errorf("failed to connect to LaleRepo: %w", err)
	}

	states := map[string]processor.StateProcessor{
		createstate.Command:  createstate.NewState(laleRepo),
		inspectstate.Command: inspectstate.NewState(laleRepo),
		getallstate.Command:  getallstate.NewState(laleRepo),
		deletestate.Command:  deletestate.NewState(laleRepo),
		reviewstate.Command:  reviewstate.NewState(laleRepo),
		helpstate.Command: helpstate.NewState([]processor.StateProcessor{
			&createstate.State{},
			&inspectstate.State{},
			&getallstate.State{},
			&deletestate.State{},
			&reviewstate.State{},
			&helpstate.State{},
		}),
	}

	baseDispatcher, err := dispatcher.New(baseBot, states, cfg.TelegramUpdateTimeout)
	if err != nil {
		return fmt.Errorf("failed to create new sidpatcher: %w", err)
	}

	wg := sync.WaitGroup{}

	logrus.Info("service started")
	if err = baseDispatcher.Dispatch(ctx, &wg, -1, 256); err != nil {
		return fmt.Errorf("failed to dispatch: %w", err)
	}
	logrus.Debug("waiting until all members of sync.WaitGroup closes")
	wg.Wait()

	logrus.Info("stop the service")

	return nil
}

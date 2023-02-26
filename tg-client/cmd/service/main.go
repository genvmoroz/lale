package main

import (
	"context"
	"fmt"
	"github.com/genvmoroz/bot-engine/bot"
	"github.com/genvmoroz/bot-engine/dispatcher"
	"github.com/genvmoroz/bot-engine/processor"
	"github.com/genvmoroz/lale-tg-client/internal/options"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("recovered: %s", r)
		}
	}()

	if err := launch(); err != nil {
		log.Fatalf("launch error: %s", err.Error())
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

	/*
		inspectcard
		createcard
		getallcards
		updatecardperformance
		getcardstoreview
		deletecard
	*/
	states := map[string]processor.StateProcessor{}

	stateProvider := func(*bot.Client, int64) map[string]processor.StateProcessor {
		return states
	}

	baseDispatcher, err := dispatcher.New(baseBot, stateProvider, cfg.TelegramUpdateTimeout)
	if err != nil {
		return fmt.Errorf("failed to create new sidpatcher: %w", err)
	}

	wg := sync.WaitGroup{}

	if err = baseDispatcher.Dispatch(ctx, &wg, -1, 256); err != nil {
		return fmt.Errorf("failed to dispatch: %w", err)
	}

	logrus.Debug("waiting until all members of sync.WaitGroup close")
	wg.Wait()
	logrus.Debug("stop the service")

	return nil
}

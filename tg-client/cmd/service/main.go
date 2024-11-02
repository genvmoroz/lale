package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"github.com/genvmoroz/bot-engine/dispatcher"
	"github.com/genvmoroz/bot-engine/processor"
	"github.com/genvmoroz/bot-engine/tg"
	"github.com/genvmoroz/lale-tg-client/internal/options"
	"github.com/genvmoroz/lale-tg-client/internal/repository"
	createstate "github.com/genvmoroz/lale-tg-client/internal/state/create"
	deletestate "github.com/genvmoroz/lale-tg-client/internal/state/delete"
	getallstate "github.com/genvmoroz/lale-tg-client/internal/state/getall"
	helpstate "github.com/genvmoroz/lale-tg-client/internal/state/help"
	inspectstate "github.com/genvmoroz/lale-tg-client/internal/state/inspect"
	"github.com/genvmoroz/lale-tg-client/internal/state/learn"
	"github.com/genvmoroz/lale-tg-client/internal/state/repeat"
	"github.com/genvmoroz/lale-tg-client/internal/state/story"
	"github.com/genvmoroz/lale-tg-client/internal/state/update"
	"github.com/sirupsen/logrus"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("recovered: %s. Stack: %s", r, string(debug.Stack()))
		}
	}()

	if err := launch(); err != nil {
		log.Fatalf("service error: %s", err.Error())
	}
}

func launch() error {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := options.FromEnv()
	if err != nil {
		return fmt.Errorf("read env config: %w", err)
	}

	logrus.SetLevel(cfg.LogLevel)

	baseBot, err := tg.NewClient(cfg.TelegramToken)
	if err != nil {
		return err
	}

	clientCfg := repository.ClientConfig{
		Host:    cfg.LaleService.Host,
		Port:    cfg.LaleService.Port,
		Timeout: cfg.LaleService.Timeout,
	}
	laleRepo, err := repository.NewLaleRepo(clientCfg)
	if err != nil {
		return fmt.Errorf("create LaleRepo: %w", err)
	}

	states := map[string]processor.StateProcessor{
		createstate.Command:  createstate.NewState(laleRepo),
		inspectstate.Command: inspectstate.NewState(laleRepo),
		getallstate.Command:  getallstate.NewState(laleRepo),
		deletestate.Command:  deletestate.NewState(laleRepo),
		repeat.Command:       repeat.NewState(laleRepo),
		story.Command:        story.NewState(laleRepo),
		learn.Command:        learn.NewState(laleRepo),
		update.Command:       update.NewState(laleRepo),
		helpstate.Command: helpstate.NewState([]processor.StateProcessor{
			&createstate.State{},
			&inspectstate.State{},
			&getallstate.State{},
			&deletestate.State{},
			&helpstate.State{},
			&repeat.State{},
			&story.State{},
			&learn.State{},
			&update.State{},
		}),
	}

	baseDispatcher, err := dispatcher.New(baseBot, states, cfg.TelegramUpdateTimeout)
	if err != nil {
		return fmt.Errorf("create new dispatcher: %w", err)
	}

	wg := sync.WaitGroup{}

	logrus.Info("service started")
	if err = baseDispatcher.Dispatch(ctx, &wg, -1, 256); err != nil {
		return fmt.Errorf("dispatch: %w", err)
	}
	logrus.Debug("waiting until all members of sync.WaitGroup closes")
	wg.Wait()

	logrus.Info("stop the service")

	return nil
}

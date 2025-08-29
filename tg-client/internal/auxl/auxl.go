package auxl

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/genvmoroz/bot-engine/processor"
	"github.com/genvmoroz/bot-engine/tg"
)

func RequestInput[T any](
	ctx context.Context,
	until func(T) bool,
	chatID int64,
	message string,
	processInput func(input string, chatID int64, client processor.Client) (T, error),
	client processor.Client,
	updateChan tg.UpdatesChannel,
) (T, string, bool, error) {
	var (
		val      T
		userName string
	)

	if err := client.SendWithParseMode(chatID, message, tg.ModeHTML); err != nil {
		return val, userName, false, err
	}

	for !until(val) {
		select {
		case <-ctx.Done():
			return val, userName, false, nil
		case update, ok := <-updateChan:
			if !ok {
				return val, userName, false, errors.New("updateChan is closed")
			}
			text := strings.TrimSpace(update.Message.Text)
			switch text {
			case "/back":
				return val, userName, true, client.Send(chatID, "Back to previous state")
			case "":
				if err := client.Send(chatID, "Empty value is not allowed"); err != nil {
					return val, userName, false, err
				}
			default:
				input, err := processInput(text, chatID, client)
				if err != nil {
					return val, userName, false, fmt.Errorf("failed ")
				}

				val = input
				userName = update.Message.From.UserName
			}
		}
	}

	return val, userName, false, nil
}

func SendAudioByLanguage(chatID int64, client processor.Client, audioByLanguage map[string][]byte) error {
	tuple := []struct {
		language string
		audio    []byte
	}{}

	for language, audio := range audioByLanguage {
		tuple = append(tuple, struct {
			language string
			audio    []byte
		}{
			language: language,
			audio:    audio,
		})
	}

	slices.SortFunc(tuple, func(a, b struct {
		language string
		audio    []byte
	},
	) int {
		return strings.Compare(b.language, a.language)
	})

	for _, item := range tuple {
		err := client.SendAudio(chatID, item.language, item.audio)
		if err != nil {
			if err = client.Send(chatID, fmt.Sprintf("sending audio (%s) error: %v", item.language, err.Error())); err != nil {
				return err
			}
		}
	}

	return nil
}

package auxl

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/genvmoroz/bot-engine/bot"
)

func RequestInput[T any](
	ctx context.Context,
	until func(T) bool,
	chatID int64,
	message string,
	processInput func(input string, chatID int64, client *bot.Client) (T, error),
	client *bot.Client,
	updateChan bot.UpdatesChannel) (T, string, bool, error) {

	var (
		val      T
		userName string
	)

	if err := client.Send(chatID, message); err != nil {
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

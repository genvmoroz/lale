package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/genvmoroz/lale-service/api"
	"github.com/liamg/tml"
)

func countExcludingEmpty(cards []*api.Card) int { // todo: implement benchmark for it
	var count int
	for _, card := range cards {
		if card != nil && card.GetId() != "" {
			count++
		}
	}
	return count
}

func translationsToString(translation *api.Translation) string {
	if translation == nil {
		return ""
	}

	sb := strings.Builder{}
	for i, t := range translation.GetTranslations() {
		sb.WriteString(t)
		if i != len(translation.GetTranslations())-1 {
			sb.WriteString(", ")
		}
	}
	return sb.String()
}

func parseTranslations(raw string) []string {
	var translations []string

	parts := strings.Split(raw, ", ")
	for _, part := range parts {
		Translation := strings.Trim(part, " .,")
		if len(Translation) != 0 {
			translations = append(translations, Translation)
		}
	}

	return translations
}

func exitWithError(desc string, err error, exitCode int) {
	err = fmt.Errorf("%s: %w", desc, err)

	tmlErr := tml.Printf("<red><bold>Error: %s</bold></red>\n", err.Error())
	if tmlErr != nil {
		log.Printf("tml error: %s, cause: %s \n", tmlErr.Error(), err.Error())
	}
	os.Exit(exitCode)
}

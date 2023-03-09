package pretty

import (
	"fmt"
	"log"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/genvmoroz/lale/service/api"
	"github.com/genvmoroz/lale/tg-client/internal/transform"
)

func Card(card *api.Card) []string {
	if card == nil {
		return nil
	}

	var p []string

	meta := `
CardID: <code>%s</code>
UserID: <code>%s</code>
Language: <code>%s</code>
NextDueDate: <code>%s</code>
`
	p = append(p, fmt.Sprintf(meta, card.GetId(), card.GetUserID(), card.GetLanguage(), card.GetNextDueDate().AsTime().Format(time.RFC3339)))

	for _, word := range transform.DefaultTransformer.ToCoreWordInformationList(card.GetWordInformationList()) {
		out, err := yaml.Marshal(word)
		if err != nil {
			log.Fatalf("unexpected err: %s", err.Error())
		}
		p = append(p, string(out))
	}

	return p
}

package card

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/genvmoroz/lale/service/pkg/entity"
)

func cardToDoc(card entity.Card) bson.M {
	return bson.M{
		"id":                  card.ID,
		"userID":              card.UserID,
		"wordInformationList": card.WordInformationList,
		"language":            card.Language,
		"correctAnswers":      card.CorrectAnswers,
		"nextDueDate":         card.NextDueDate,
	}
}

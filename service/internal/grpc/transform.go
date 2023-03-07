package grpc

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/genvmoroz/lale/service/api"
	"github.com/genvmoroz/lale/service/internal/core"
	"github.com/genvmoroz/lale/service/internal/entity"
	"github.com/genvmoroz/lale/service/pkg/lang"
)

type (
	Transformer interface {
		ToCoreInspectCardRequest(req *api.InspectCardRequest) core.InspectCardRequest
		ToAPIInspectCardResponse(resp core.InspectCardResponse) *api.InspectCardResponse
		ToCoreCreateCardRequest(req *api.CreateCardRequest) core.CreateCardRequest
		ToAPICreateCardResponse(resp core.CreateCardResponse) *api.CreateCardResponse
		ToCoreGetCardsRequest(req *api.GetCardsRequest) core.GetCardsRequest
		ToAPIGetCardsResponse(resp core.GetCardsResponse) *api.GetCardsResponse
		ToCoreUpdateCardPerformanceRequest(req *api.UpdateCardPerformanceRequest) core.UpdateCardPerformanceRequest
		ToAPIUpdateCardPerformanceResponse(resp core.UpdateCardPerformanceResponse) *api.UpdateCardPerformanceResponse
		ToCoreGetCardsForReviewRequest(req *api.GetCardsForReviewRequest) core.GetCardsForReviewRequest
		ToCoreDeleteCardRequest(req *api.DeleteCardRequest) core.DeleteCardRequest
		ToAPIDeleteCardResponse(resp core.DeleteCardResponse) *api.DeleteCardResponse
	}

	transformer struct{}
)

var DefaultTransformer Transformer = transformer{}

func (transformer) ToCoreInspectCardRequest(req *api.InspectCardRequest) core.InspectCardRequest {
	return core.InspectCardRequest{
		UserID:   req.GetUserID(),
		Language: lang.Language(req.GetLanguage()),
		Word:     req.GetWord(),
	}
}

func (t transformer) ToAPIInspectCardResponse(resp core.InspectCardResponse) *api.InspectCardResponse {
	return &api.InspectCardResponse{
		Card: t.toAPICard(resp.Card),
	}
}

func (t transformer) ToCoreCreateCardRequest(req *api.CreateCardRequest) core.CreateCardRequest {
	return core.CreateCardRequest{
		UserID:              req.GetUserID(),
		Language:            lang.Language(req.GetLanguage()),
		WordInformationList: t.toCoreWordInformationList(req.GetWordInformationList()),
		Params:              t.toCoreCreateCardParameters(req.GetParams()),
	}
}

func (t transformer) ToAPICreateCardResponse(resp core.CreateCardResponse) *api.CreateCardResponse {
	return &api.CreateCardResponse{
		Card: t.toAPICard(resp.Card),
	}
}

func (transformer) ToCoreGetCardsRequest(req *api.GetCardsRequest) core.GetCardsRequest {
	return core.GetCardsRequest{
		UserID:   req.GetUserID(),
		Language: lang.Language(req.GetLanguage()),
	}
}

func (t transformer) ToAPIGetCardsResponse(resp core.GetCardsResponse) *api.GetCardsResponse {
	return &api.GetCardsResponse{
		UserID:   resp.UserID,
		Language: resp.Language.String(),
		Cards:    t.toAPICards(resp.Cards),
	}
}

func (transformer) ToCoreUpdateCardPerformanceRequest(req *api.UpdateCardPerformanceRequest) core.UpdateCardPerformanceRequest {
	return core.UpdateCardPerformanceRequest{
		UserID:            req.GetUserID(),
		CardID:            req.GetCardID(),
		PerformanceRating: req.GetPerformanceRating(),
	}
}

func (transformer) ToAPIUpdateCardPerformanceResponse(resp core.UpdateCardPerformanceResponse) *api.UpdateCardPerformanceResponse {
	return &api.UpdateCardPerformanceResponse{
		NextDueDate: timestamppb.New(resp.NextDueDate),
	}
}

func (transformer) ToCoreGetCardsForReviewRequest(req *api.GetCardsForReviewRequest) core.GetCardsForReviewRequest {
	return core.GetCardsForReviewRequest{
		UserID:         req.GetUserID(),
		Language:       lang.Language(req.GetLanguage()),
		SentencesCount: req.GetSentencesCount(),
	}
}

func (transformer) ToCoreDeleteCardRequest(req *api.DeleteCardRequest) core.DeleteCardRequest {
	return core.DeleteCardRequest{
		UserID: req.GetUserID(),
		CardID: req.GetCardID(),
	}
}

func (t transformer) ToAPIDeleteCardResponse(resp core.DeleteCardResponse) *api.DeleteCardResponse {
	return &api.DeleteCardResponse{
		Card: t.toAPICard(resp.Card),
	}
}

func (t transformer) toAPICards(cards []entity.Card) []*api.Card {
	if len(cards) == 0 {
		return nil
	}

	res := make([]*api.Card, len(cards))
	for i, c := range cards {
		res[i] = t.toAPICard(c)
	}

	return res
}

func (t transformer) toAPICard(card entity.Card) *api.Card {
	return &api.Card{
		Id:                  card.ID,
		UserID:              card.UserID,
		Language:            string(card.Language),
		WordInformationList: t.toAPIWordInformationList(card.WordInformationList),
		CorrectAnswers:      card.CorrectAnswers,
		NextDueDate:         timestamppb.New(card.NextDueDate),
	}
}

func (t transformer) toCoreCreateCardParameters(p *api.CreateCardParameters) core.CreateCardParameters {
	return core.CreateCardParameters{
		EnrichWordInformationFromDictionary: p.GetEnrichWordInformationFromDictionary(),
	}
}

func (t transformer) toAPIWordInformationList(list []entity.WordInformation) []*api.WordInformation {
	if len(list) == 0 {
		return nil
	}

	res := make([]*api.WordInformation, len(list))
	for i, w := range list {
		res[i] = t.toAPIWordInformation(w)
	}

	return res
}

func (t transformer) toCoreWordInformationList(list []*api.WordInformation) []entity.WordInformation {
	if len(list) == 0 {
		return nil
	}

	res := make([]entity.WordInformation, 0, len(list))
	for _, w := range list {
		if w != nil {
			res = append(res, t.toCoreWordInformation(w))

		}
	}

	return res
}

func (transformer) toCoreTranslation(t *api.Translation) *entity.Translation {
	if t == nil {
		return nil
	}

	return &entity.Translation{
		Language:     lang.Language(t.Language),
		Translations: t.Translations,
	}
}

func (t transformer) toAPIWordInformation(info entity.WordInformation) *api.WordInformation {
	return &api.WordInformation{
		Word:        info.Word,
		Translation: t.toAPITranslation(info.Translation),
		Origin:      info.Origin,
		Phonetics:   t.toAPIPhonetics(info.Phonetics),
		Meanings:    t.toAPIMeanings(info.Meanings),
		Sentences:   info.Sentences,
	}
}

func (t transformer) toCoreWordInformation(info *api.WordInformation) entity.WordInformation {
	return entity.WordInformation{
		Word:        info.Word,
		Translation: t.toCoreTranslation(info.Translation),
		Origin:      info.Origin,
		Phonetics:   t.toCorePhonetics(info.Phonetics),
		Meanings:    t.toCoreMeanings(info.Meanings),
		Sentences:   info.Sentences,
	}
}

func (t transformer) toAPIMeanings(meanings []entity.Meaning) []*api.Meaning {
	if len(meanings) == 0 {
		return nil
	}

	res := make([]*api.Meaning, len(meanings))
	for i, m := range meanings {
		res[i] = t.toAPIMeaning(m)
	}

	return res
}

func (t transformer) toAPIMeaning(m entity.Meaning) *api.Meaning {
	return &api.Meaning{
		PartOfSpeech: m.PartOfSpeech,
		Definitions:  t.toAPIDefinitions(m.Definitions),
	}
}

func (t transformer) toCoreMeanings(meanings []*api.Meaning) []entity.Meaning {
	if len(meanings) == 0 {
		return nil
	}

	res := make([]entity.Meaning, 0, len(meanings))
	for _, m := range meanings {
		if m != nil {
			res = append(res, t.toCoreMeaning(m))
		}
	}

	return res
}

func (t transformer) toCoreMeaning(m *api.Meaning) entity.Meaning {
	return entity.Meaning{
		PartOfSpeech: m.PartOfSpeech,
		Definitions:  t.toCoreDefinitions(m.Definitions),
	}
}

func (t transformer) toAPIDefinitions(definitions []entity.Definition) []*api.Definition {
	if len(definitions) == 0 {
		return nil
	}

	res := make([]*api.Definition, len(definitions))
	for i, d := range definitions {
		res[i] = t.toAPIDefinition(d)
	}

	return res
}

func (transformer) toAPIDefinition(d entity.Definition) *api.Definition {
	return &api.Definition{
		Definition: d.Definition,
		Example:    d.Example,
		Synonyms:   d.Synonyms,
		Antonyms:   d.Antonyms,
	}
}

func (t transformer) toCoreDefinitions(definitions []*api.Definition) []entity.Definition {
	if len(definitions) == 0 {
		return nil
	}

	res := make([]entity.Definition, 0, len(definitions))
	for _, d := range definitions {
		if d != nil {
			res = append(res, t.toCoreDefinition(d))
		}
	}

	return res
}

func (transformer) toCoreDefinition(d *api.Definition) entity.Definition {
	return entity.Definition{
		Definition: d.Definition,
		Example:    d.Example,
		Synonyms:   d.Synonyms,
		Antonyms:   d.Antonyms,
	}
}

func (t transformer) toAPIPhonetics(phonetics []entity.Phonetic) []*api.Phonetic {
	if len(phonetics) == 0 {
		return nil
	}

	res := make([]*api.Phonetic, len(phonetics))
	for i, p := range phonetics {
		res[i] = t.toAPIPhonetic(p)
	}

	return res
}

func (transformer) toAPIPhonetic(p entity.Phonetic) *api.Phonetic {
	return &api.Phonetic{
		Text:      p.Text,
		AudioLink: p.AudioLink,
	}
}

func (t transformer) toCorePhonetics(phonetics []*api.Phonetic) []entity.Phonetic {
	if len(phonetics) == 0 {
		return nil
	}

	res := make([]entity.Phonetic, 0, len(phonetics))
	for _, p := range phonetics {
		if p != nil {
			res = append(res, t.toCorePhonetic(p))
		}
	}

	return res
}

func (transformer) toCorePhonetic(p *api.Phonetic) entity.Phonetic {
	return entity.Phonetic{
		Text:      p.Text,
		AudioLink: p.AudioLink,
	}
}

func (transformer) toAPITranslation(p *entity.Translation) *api.Translation {
	if p == nil {
		return nil
	}
	return &api.Translation{Language: string(p.Language), Translations: p.Translations}
}

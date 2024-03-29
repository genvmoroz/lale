package grpc

import (
	"fmt"

	"github.com/genvmoroz/lale/service/api"
	"github.com/genvmoroz/lale/service/internal/core"
	"github.com/genvmoroz/lale/service/pkg/entity"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type (
	Transformer interface {
		ToAPICard(card entity.Card) *api.Card
		ToCoreInspectCardRequest(req *api.InspectCardRequest) (core.InspectCardRequest, error)
		ToCorePromptCardRequest(req *api.PromptCardRequest) (core.PromptCardRequest, error)
		ToAPIPromptCardResponse(resp core.PromptCardResponse) *api.PromptCardResponse
		ToCoreCreateCardRequest(req *api.CreateCardRequest) (core.CreateCardRequest, error)
		ToCoreGetCardsRequest(req *api.GetCardsRequest) (core.GetCardsRequest, error)
		ToAPIGetCardsResponse(resp core.GetCardsResponse) *api.GetCardsResponse
		ToCoreUpdateCardRequest(req *api.UpdateCardRequest) (core.UpdateCardRequest, error)
		ToCoreUpdateCardPerformanceRequest(req *api.UpdateCardPerformanceRequest) core.UpdateCardPerformanceRequest
		ToAPIUpdateCardPerformanceResponse(resp core.UpdateCardPerformanceResponse) *api.UpdateCardPerformanceResponse
		ToCoreGetSentencesRequest(req *api.GetSentencesRequest) core.GetSentencesRequest
		ToAPIGetSentencesResponse(resp core.GetSentencesResponse) *api.GetSentencesResponse
		ToCoreGenerateStoryRequest(req *api.GenerateStoryRequest) (core.GenerateStoryRequest, error)
		ToAPIGenerateStoryResponse(resp core.GenerateStoryResponse) *api.GenerateStoryResponse
		ToCoreDeleteCardRequest(req *api.DeleteCardRequest) core.DeleteCardRequest
	}

	transformer struct{}
)

func DefaultTransformer() Transformer {
	return transformer{}
}

func (transformer) ToCoreInspectCardRequest(req *api.InspectCardRequest) (core.InspectCardRequest, error) {
	if req == nil {
		return core.InspectCardRequest{}, nil
	}

	lang, err := language.Parse(req.GetLanguage())
	if err != nil {
		return core.InspectCardRequest{}, fmt.Errorf("invalid language (%s): %w", req.GetLanguage(), err)
	}

	return core.InspectCardRequest{
		UserID:   req.GetUserID(),
		Language: lang,
		Word:     req.GetWord(),
	}, nil
}

func (transformer) ToCorePromptCardRequest(req *api.PromptCardRequest) (core.PromptCardRequest, error) {
	if req == nil {
		return core.PromptCardRequest{}, nil
	}

	wLang, err := language.Parse(req.GetWordLanguage())
	if err != nil {
		return core.PromptCardRequest{}, fmt.Errorf("invalid language (%s): %w", req.GetWordLanguage(), err)
	}
	tLang, err := language.Parse(req.GetTranslationLanguage())
	if err != nil {
		return core.PromptCardRequest{}, fmt.Errorf("invalid language (%s): %w", req.GetTranslationLanguage(), err)
	}
	return core.PromptCardRequest{
		UserID:              req.GetUserID(),
		Word:                req.GetWord(),
		WordLanguage:        wLang,
		TranslationLanguage: tLang,
	}, nil
}

func (transformer) ToAPIPromptCardResponse(resp core.PromptCardResponse) *api.PromptCardResponse {
	return &api.PromptCardResponse{Words: resp.Words}
}

func (t transformer) ToCoreCreateCardRequest(req *api.CreateCardRequest) (core.CreateCardRequest, error) {
	if req == nil {
		return core.CreateCardRequest{}, nil
	}

	lang, err := language.Parse(req.GetLanguage())
	if err != nil {
		return core.CreateCardRequest{}, fmt.Errorf("invalid language (%s): %w", req.GetLanguage(), err)
	}

	words, err := t.toCoreWordInformationList(req.GetWordInformationList())
	if err != nil {
		return core.CreateCardRequest{}, err
	}

	return core.CreateCardRequest{
		UserID:              req.GetUserID(),
		Language:            lang,
		WordInformationList: words,
		Params:              t.toCoreCreateCardParameters(req.GetParams()),
	}, nil
}

func (transformer) ToCoreGetCardsRequest(req *api.GetCardsRequest) (core.GetCardsRequest, error) {
	if req == nil {
		return core.GetCardsRequest{}, nil
	}

	lang, err := language.Parse(req.GetLanguage())
	if err != nil {
		return core.GetCardsRequest{}, fmt.Errorf("invalid language (%s): %w", req.GetLanguage(), err)
	}
	return core.GetCardsRequest{
		UserID:   req.GetUserID(),
		Language: lang,
	}, nil
}

func (t transformer) ToAPIGetCardsResponse(resp core.GetCardsResponse) *api.GetCardsResponse {
	return &api.GetCardsResponse{
		UserID:   resp.UserID,
		Language: resp.Language.String(),
		Cards:    t.toAPICards(resp.Cards),
	}
}

func (transformer) ToCoreUpdateCardPerformanceRequest(
	req *api.UpdateCardPerformanceRequest,
) core.UpdateCardPerformanceRequest {
	return core.UpdateCardPerformanceRequest{
		UserID:            req.GetUserID(),
		CardID:            req.GetCardID(),
		PerformanceRating: req.GetPerformanceRating(),
	}
}

func (transformer) ToAPIUpdateCardPerformanceResponse(
	resp core.UpdateCardPerformanceResponse,
) *api.UpdateCardPerformanceResponse {
	return &api.UpdateCardPerformanceResponse{
		NextDueDate: timestamppb.New(resp.NextDueDate),
	}
}

func (transformer) ToCoreDeleteCardRequest(req *api.DeleteCardRequest) core.DeleteCardRequest {
	return core.DeleteCardRequest{
		UserID: req.GetUserID(),
		CardID: req.GetCardID(),
	}
}

func (t transformer) ToCoreGetSentencesRequest(req *api.GetSentencesRequest) core.GetSentencesRequest {
	return core.GetSentencesRequest{
		UserID:         req.GetUserID(),
		Word:           req.GetWord(),
		SentencesCount: req.GetSentencesCount(),
	}
}

func (t transformer) ToAPIGetSentencesResponse(resp core.GetSentencesResponse) *api.GetSentencesResponse {
	return &api.GetSentencesResponse{Sentences: resp.Sentences}
}

func (t transformer) ToCoreGenerateStoryRequest(req *api.GenerateStoryRequest) (core.GenerateStoryRequest, error) {
	if req == nil {
		return core.GenerateStoryRequest{}, nil
	}

	lang, err := language.Parse(req.GetLanguage())
	if err != nil {
		return core.GenerateStoryRequest{}, fmt.Errorf("invalid language (%s): %w", req.GetLanguage(), err)
	}
	return core.GenerateStoryRequest{
		UserID:   req.GetUserID(),
		Language: lang,
	}, nil
}

func (t transformer) ToAPIGenerateStoryResponse(resp core.GenerateStoryResponse) *api.GenerateStoryResponse {
	return &api.GenerateStoryResponse{Story: resp.Story}
}

func (t transformer) ToCoreUpdateCardRequest(req *api.UpdateCardRequest) (core.UpdateCardRequest, error) {
	words, err := t.toCoreWordInformationList(req.GetWordInformationList())
	if err != nil {
		return core.UpdateCardRequest{}, err
	}
	return core.UpdateCardRequest{
		UserID:              req.GetUserID(),
		CardID:              req.GetCardID(),
		WordInformationList: words,
		Params:              t.toCoreCreateCardParameters(req.GetParams()),
	}, nil
}

func (t transformer) toAPICards(cards []entity.Card) []*api.Card {
	if len(cards) == 0 {
		return nil
	}

	res := make([]*api.Card, len(cards))
	for i, c := range cards {
		res[i] = t.ToAPICard(c)
	}

	return res
}

func (t transformer) ToAPICard(card entity.Card) *api.Card {
	return &api.Card{
		Id:                  card.ID,
		UserID:              card.UserID,
		Language:            card.Language.String(),
		WordInformationList: t.toAPIWordInformationList(card.WordInformationList),
		CorrectAnswers:      card.CorrectAnswers,
		NextDueDate:         timestamppb.New(card.NextDueDate),
	}
}

func (t transformer) toCoreCreateCardParameters(p *api.Parameters) core.Parameters {
	return core.Parameters{
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

func (t transformer) toCoreWordInformationList(list []*api.WordInformation) ([]entity.WordInformation, error) {
	if len(list) == 0 {
		return nil, nil
	}

	res := make([]entity.WordInformation, 0, len(list))
	for _, w := range list {
		if w != nil {
			aw, err := t.toCoreWordInformation(w)
			if err != nil {
				return nil, fmt.Errorf("transform word (%s): %w", w.Word, err)
			}
			res = append(res, aw)
		}
	}

	return res, nil
}

func (transformer) toCoreTranslation(t *api.Translation) (*entity.Translation, error) {
	if t == nil {
		return nil, fmt.Errorf("empty translation")
	}

	lang, err := language.Parse(t.GetLanguage())
	if err != nil {
		return nil, fmt.Errorf("invalid language (%s): %w", t.GetLanguage(), err)
	}

	return &entity.Translation{
		Language:     lang,
		Translations: t.Translations,
	}, nil
}

func (t transformer) toAPIWordInformation(info entity.WordInformation) *api.WordInformation {
	return &api.WordInformation{
		Word:        info.Word,
		Translation: t.toAPITranslation(info.Translation),
		Origin:      info.Origin,
		Phonetics:   t.toAPIPhonetics(info.Phonetics),
		Meanings:    t.toAPIMeanings(info.Meanings),
		Audio:       info.Audio,
	}
}

func (t transformer) toCoreWordInformation(info *api.WordInformation) (entity.WordInformation, error) {
	out := entity.WordInformation{
		Word:      info.Word,
		Origin:    info.Origin,
		Phonetics: t.toCorePhonetics(info.Phonetics),
		Meanings:  t.toCoreMeanings(info.Meanings),
		Audio:     info.GetAudio(),
	}

	if info.Translation != nil {
		translation, err := t.toCoreTranslation(info.Translation)
		if err != nil {
			return entity.WordInformation{}, fmt.Errorf("translation: %w", err)
		}
		out.Translation = translation
	}

	return out, nil
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
	return &api.Phonetic{Text: p.Text}
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
	return entity.Phonetic{Text: p.Text}
}

func (transformer) toAPITranslation(p *entity.Translation) *api.Translation {
	if p == nil {
		return nil
	}
	return &api.Translation{
		Language:     p.Language.String(),
		Translations: p.Translations,
	}
}

package grpc

import (
	"reflect"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/genvmoroz/lale/service/api"
	"github.com/genvmoroz/lale/service/internal/core"
	"github.com/genvmoroz/lale/service/pkg/entity"
	"github.com/genvmoroz/lale/service/pkg/lang"
)

func TestAPICard(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  entity.Card
		want *api.Card
	}{
		{
			name: "with ID",
			arg: entity.Card{
				ID:          "someID",
				NextDueDate: time.Date(2022, 2, 24, 0, 0, 0, 0, time.UTC),
			},
			want: &api.Card{
				Id:          "someID",
				NextDueDate: timestamppb.New(time.Date(2022, 2, 24, 0, 0, 0, 0, time.UTC)),
			},
		},
		{
			name: "with UserID",
			arg: entity.Card{
				UserID:      "someUserID",
				NextDueDate: time.Date(2022, 2, 24, 1, 0, 0, 0, time.UTC),
			},
			want: &api.Card{
				UserID:      "someUserID",
				NextDueDate: timestamppb.New(time.Date(2022, 2, 24, 1, 0, 0, 0, time.UTC)),
			},
		},
		{
			name: "with Language",
			arg: entity.Card{
				Language:    lang.Language("uk"),
				NextDueDate: time.Date(2022, 2, 24, 2, 0, 0, 0, time.UTC),
			},
			want: &api.Card{
				Language:    "uk",
				NextDueDate: timestamppb.New(time.Date(2022, 2, 24, 2, 0, 0, 0, time.UTC)),
			},
		},
		{
			name: "with Word",
			arg: entity.Card{
				WordInformationList: []entity.WordInformation{
					{Word: "word1", Translation: &entity.Translation{Language: "en", Translations: []string{"Translation1"}}, Origin: "origin1"},
					{Word: "word2", Translation: &entity.Translation{Language: "en", Translations: []string{"Translation2"}}, Origin: "origin2"},
				},
				NextDueDate: time.Date(2022, 2, 24, 2, 0, 0, 0, time.UTC),
			},
			want: &api.Card{
				WordInformationList: []*api.WordInformation{
					{Word: "word1", Translation: &api.Translation{Language: "en", Translations: []string{"Translation1"}}, Origin: "origin1"},
					{Word: "word2", Translation: &api.Translation{Language: "en", Translations: []string{"Translation2"}}, Origin: "origin2"},
				},
				NextDueDate: timestamppb.New(time.Date(2022, 2, 24, 2, 0, 0, 0, time.UTC)),
			},
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := (transformer{}).toAPICard(tt.arg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("apiCard() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAPIWordInformation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  entity.WordInformation
		want *api.WordInformation
	}{
		{
			name: "with word",
			arg:  entity.WordInformation{Word: "someWord"},
			want: &api.WordInformation{Word: "someWord"},
		},
		{
			name: "with audio",
			arg:  entity.WordInformation{Audio: []byte("someAudio")},
			want: &api.WordInformation{Audio: []byte("someAudio")},
		},
		{
			name: "with Translation",
			arg: entity.WordInformation{
				Translation: &entity.Translation{
					Language:     "en",
					Translations: []string{"Translation1", "Translation2"},
				},
			},
			want: &api.WordInformation{
				Translation: &api.Translation{
					Language:     "en",
					Translations: []string{"Translation1", "Translation2"},
				},
			},
		},
		{
			name: "with origin",
			arg:  entity.WordInformation{Origin: "origin1"},
			want: &api.WordInformation{Origin: "origin1"},
		},
		{
			name: "with phonetics",
			arg: entity.WordInformation{
				Phonetics: []entity.Phonetic{
					{Text: "text1"},
					{Text: "text2"},
				},
			},
			want: &api.WordInformation{
				Phonetics: []*api.Phonetic{
					{Text: "text1"},
					{Text: "text2"},
				},
			},
		},
		{
			name: "with meanings",
			arg: entity.WordInformation{
				Meanings: []entity.Meaning{
					{
						PartOfSpeech: "part1",
						Definitions: []entity.Definition{{
							Definition: "definition1",
							Example:    "example1",
							Synonyms:   []string{"synonym1"},
							Antonyms:   []string{"antonym1"},
						}},
					},
					{
						PartOfSpeech: "part2",
						Definitions: []entity.Definition{{
							Definition: "definition2",
							Example:    "example2",
							Synonyms:   []string{"synonym2"},
							Antonyms:   []string{"antonym2"},
						}},
					},
				},
			},
			want: &api.WordInformation{
				Meanings: []*api.Meaning{
					{
						PartOfSpeech: "part1",
						Definitions: []*api.Definition{{
							Definition: "definition1",
							Example:    "example1",
							Synonyms:   []string{"synonym1"},
							Antonyms:   []string{"antonym1"},
						}},
					},
					{
						PartOfSpeech: "part2",
						Definitions: []*api.Definition{{
							Definition: "definition2",
							Example:    "example2",
							Synonyms:   []string{"synonym2"},
							Antonyms:   []string{"antonym2"},
						}},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := (transformer{}).toAPIWordInformation(tt.arg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("apiWord() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransformerToCoreInspectCardRequest(t *testing.T) {
	t.Parallel()

	inReq := &api.InspectCardRequest{
		UserID:   "UserID",
		Language: lang.English.String(),
		Word:     "Word",
	}

	expReq := core.InspectCardRequest{
		UserID:   "UserID",
		Language: lang.English,
		Word:     "Word",
	}

	tr := DefaultTransformer

	if got := tr.ToCoreInspectCardRequest(inReq); !reflect.DeepEqual(got, expReq) {
		t.Fatalf("ToCoreInspectCardRequest() = %v, want %v", got, expReq)
	}
}

func TestTransformerToAPIInspectCardResponse(t *testing.T) {
	t.Parallel()

	nextDueDate := time.Now().Add(time.Hour)

	inResp := core.InspectCardResponse{
		Card: entity.Card{
			ID:       "ID",
			UserID:   "UserID",
			Language: lang.English,
			WordInformationList: []entity.WordInformation{
				{
					Word: "Word_1",
					Translation: &entity.Translation{
						Language:     lang.English,
						Translations: []string{"Translation_1", "Translation_11"},
					},
					Origin: "Origin_1",
					Phonetics: []entity.Phonetic{
						{Text: "Text_11"},
						{Text: "Text_12"},
					},
					Meanings: []entity.Meaning{
						{
							PartOfSpeech: "PartOfSpeech_11",
							Definitions: []entity.Definition{
								{
									Definition: "Definition_111",
									Example:    "Example_111",
									Synonyms:   []string{"synonym_1111", "synonym_1112"},
									Antonyms:   []string{"antonym_1111", "antonym_1112"},
								},
								{
									Definition: "Definition_112",
									Example:    "Example_112",
									Synonyms:   []string{"synonym_1121", "synonym_1122"},
									Antonyms:   []string{"antonym_1121", "antonym_1122"},
								},
							},
						},
						{
							PartOfSpeech: "PartOfSpeech_12",
							Definitions: []entity.Definition{
								{
									Definition: "Definition_121",
									Example:    "Example_121",
									Synonyms:   []string{"synonym_1211", "synonym_1212"},
									Antonyms:   []string{"antonym_1211", "antonym_1212"},
								},
								{
									Definition: "Definition_122",
									Example:    "Example_122",
									Synonyms:   []string{"synonym_1221", "synonym_1222"},
									Antonyms:   []string{"antonym_1221", "antonym_1222"},
								},
							},
						},
					},
				},
				{
					Word: "Word_2",
					Translation: &entity.Translation{
						Language:     lang.English,
						Translations: []string{"Translation_2", "Translation_21"},
					},
					Origin: "Origin_2",
					Phonetics: []entity.Phonetic{
						{Text: "Text_21"},
						{Text: "Text_22"},
					},
					Meanings: []entity.Meaning{
						{
							PartOfSpeech: "PartOfSpeech_21",
							Definitions: []entity.Definition{
								{
									Definition: "Definition_211",
									Example:    "Example_211",
									Synonyms:   []string{"synonym_2111", "synonym_2112"},
									Antonyms:   []string{"antonym_2111", "antonym_2112"},
								},
								{
									Definition: "Definition_212",
									Example:    "Example_212",
									Synonyms:   []string{"synonym_2121", "synonym_2122"},
									Antonyms:   []string{"antonym_2121", "antonym_2122"},
								},
							},
						},
						{
							PartOfSpeech: "PartOfSpeech_22",
							Definitions: []entity.Definition{
								{
									Definition: "Definition_221",
									Example:    "Example_221",
									Synonyms:   []string{"synonym_2211", "synonym_2212"},
									Antonyms:   []string{"antonym_2211", "antonym_2212"},
								},
								{
									Definition: "Definition_122",
									Example:    "Example_122",
									Synonyms:   []string{"synonym_2221", "synonym_2222"},
									Antonyms:   []string{"antonym_2221", "antonym_2222"},
								},
							},
						},
					},
				},
			},
			CorrectAnswers: 1,
			NextDueDate:    nextDueDate,
		},
	}

	expResp := &api.InspectCardResponse{
		Card: &api.Card{
			Id:       "ID",
			UserID:   "UserID",
			Language: lang.English.String(),
			WordInformationList: []*api.WordInformation{
				{
					Word: "Word_1",
					Translation: &api.Translation{
						Language:     lang.English.String(),
						Translations: []string{"Translation_1", "Translation_11"},
					},
					Origin: "Origin_1",
					Phonetics: []*api.Phonetic{
						{Text: "Text_11"},
						{Text: "Text_12"},
					},
					Meanings: []*api.Meaning{
						{
							PartOfSpeech: "PartOfSpeech_11",
							Definitions: []*api.Definition{
								{
									Definition: "Definition_111",
									Example:    "Example_111",
									Synonyms:   []string{"synonym_1111", "synonym_1112"},
									Antonyms:   []string{"antonym_1111", "antonym_1112"},
								},
								{
									Definition: "Definition_112",
									Example:    "Example_112",
									Synonyms:   []string{"synonym_1121", "synonym_1122"},
									Antonyms:   []string{"antonym_1121", "antonym_1122"},
								},
							},
						},
						{
							PartOfSpeech: "PartOfSpeech_12",
							Definitions: []*api.Definition{
								{
									Definition: "Definition_121",
									Example:    "Example_121",
									Synonyms:   []string{"synonym_1211", "synonym_1212"},
									Antonyms:   []string{"antonym_1211", "antonym_1212"},
								},
								{
									Definition: "Definition_122",
									Example:    "Example_122",
									Synonyms:   []string{"synonym_1221", "synonym_1222"},
									Antonyms:   []string{"antonym_1221", "antonym_1222"},
								},
							},
						},
					},
				},
				{
					Word: "Word_2",
					Translation: &api.Translation{
						Language:     lang.English.String(),
						Translations: []string{"Translation_2", "Translation_21"},
					},
					Origin: "Origin_2",
					Phonetics: []*api.Phonetic{
						{Text: "Text_21"},
						{Text: "Text_22"},
					},
					Meanings: []*api.Meaning{
						{
							PartOfSpeech: "PartOfSpeech_21",
							Definitions: []*api.Definition{
								{
									Definition: "Definition_211",
									Example:    "Example_211",
									Synonyms:   []string{"synonym_2111", "synonym_2112"},
									Antonyms:   []string{"antonym_2111", "antonym_2112"},
								},
								{
									Definition: "Definition_212",
									Example:    "Example_212",
									Synonyms:   []string{"synonym_2121", "synonym_2122"},
									Antonyms:   []string{"antonym_2121", "antonym_2122"},
								},
							},
						},
						{
							PartOfSpeech: "PartOfSpeech_22",
							Definitions: []*api.Definition{
								{
									Definition: "Definition_221",
									Example:    "Example_221",
									Synonyms:   []string{"synonym_2211", "synonym_2212"},
									Antonyms:   []string{"antonym_2211", "antonym_2212"},
								},
								{
									Definition: "Definition_122",
									Example:    "Example_122",
									Synonyms:   []string{"synonym_2221", "synonym_2222"},
									Antonyms:   []string{"antonym_2221", "antonym_2222"},
								},
							},
						},
					},
				},
			},
			CorrectAnswers: 1,
			NextDueDate:    timestamppb.New(nextDueDate),
		},
	}

	tr := DefaultTransformer

	if got := tr.ToAPIInspectCardResponse(inResp); !reflect.DeepEqual(got, expResp) {
		t.Fatalf("ToAPIInspectCardResponse() = %v, want %v", got, expResp)
	}
}

func TestTransformerToCoreCreateCardRequest(t *testing.T) {
	t.Parallel()

	type (
		input struct{ req *api.CreateCardRequest }
		want  struct{ req core.CreateCardRequest }
	)
	testcases := map[string]struct {
		input input
		want  want
	}{
		"with word info list": {
			input: input{
				req: &api.CreateCardRequest{
					UserID:   "UserID",
					Language: lang.English.String(),
					WordInformationList: []*api.WordInformation{
						{
							Word: "Word_1",
							Translation: &api.Translation{
								Language:     lang.Ukrainian.String(),
								Translations: []string{"Translations_11", "Translations_12"},
							},
							Origin: "Origin_1",
							Phonetics: []*api.Phonetic{
								{Text: "Text_11"},
								{Text: "Text_12"},
							},
							Meanings: []*api.Meaning{
								{
									PartOfSpeech: "PartOfSpeech_11",
									Definitions: []*api.Definition{
										{
											Definition: "Definition_11",
											Example:    "Example_11",
											Synonyms:   []string{"Synonym_111", "Synonym_112"},
											Antonyms:   []string{"Antonym_111", "Antonym_112"},
										},
										{
											Definition: "Definition_12",
											Example:    "Example_12",
											Synonyms:   []string{"Synonym_121", "Synonym_122"},
											Antonyms:   []string{"Antonym_121", "Antonym_122"},
										},
									},
								},
								nil,
								{
									PartOfSpeech: "PartOfSpeech_12",
									Definitions: []*api.Definition{
										{
											Definition: "Definition_12",
											Example:    "Example_12",
											Synonyms:   []string{"Synonym_121", "Synonym_122"},
											Antonyms:   []string{"Antonym_121", "Antonym_122"},
										},
										{
											Definition: "Definition_13",
											Example:    "Example_13",
											Synonyms:   []string{"Synonym_131", "Synonym_132"},
											Antonyms:   []string{"Antonym_131", "Antonym_132"},
										},
									},
								},
							},
						},
						{
							Word: "Word_2",
							Translation: &api.Translation{
								Language:     lang.Ukrainian.String(),
								Translations: []string{"Translations_21", "Translations_22"},
							},
							Origin: "Origin_2",
							Phonetics: []*api.Phonetic{
								{Text: "Text_21"},
								{Text: "Text_22"},
							},
							Meanings: []*api.Meaning{
								{
									PartOfSpeech: "PartOfSpeech_21",
									Definitions: []*api.Definition{
										{
											Definition: "Definition_21",
											Example:    "Example_21",
											Synonyms:   []string{"Synonym_211", "Synonym_212"},
											Antonyms:   []string{"Antonym_211", "Antonym_212"},
										},
										{
											Definition: "Definition_22",
											Example:    "Example_22",
											Synonyms:   []string{"Synonym_221", "Synonym_222"},
											Antonyms:   []string{"Antonym_221", "Antonym_222"},
										},
									},
								},
								{
									PartOfSpeech: "PartOfSpeech_12",
									Definitions: []*api.Definition{
										{
											Definition: "Definition_12",
											Example:    "Example_12",
											Synonyms:   []string{"Synonym_121", "Synonym_122"},
											Antonyms:   []string{"Antonym_121", "Antonym_122"},
										},
										{
											Definition: "Definition_13",
											Example:    "Example_13",
											Synonyms:   []string{"Synonym_131", "Synonym_132"},
											Antonyms:   []string{"Antonym_131", "Antonym_132"},
										},
									},
								},
							},
						},
					},
					Params: &api.CreateCardParameters{EnrichWordInformationFromDictionary: true},
				},
			},
			want: want{
				req: core.CreateCardRequest{
					UserID:   "UserID",
					Language: lang.English,
					WordInformationList: []entity.WordInformation{
						{
							Word: "Word_1",
							Translation: &entity.Translation{
								Language:     lang.Ukrainian,
								Translations: []string{"Translations_11", "Translations_12"},
							},
							Origin: "Origin_1",
							Phonetics: []entity.Phonetic{
								{Text: "Text_11"},
								{Text: "Text_12"},
							},
							Meanings: []entity.Meaning{
								{
									PartOfSpeech: "PartOfSpeech_11",
									Definitions: []entity.Definition{
										{
											Definition: "Definition_11",
											Example:    "Example_11",
											Synonyms:   []string{"Synonym_111", "Synonym_112"},
											Antonyms:   []string{"Antonym_111", "Antonym_112"},
										},
										{
											Definition: "Definition_12",
											Example:    "Example_12",
											Synonyms:   []string{"Synonym_121", "Synonym_122"},
											Antonyms:   []string{"Antonym_121", "Antonym_122"},
										},
									},
								},
								{
									PartOfSpeech: "PartOfSpeech_12",
									Definitions: []entity.Definition{
										{
											Definition: "Definition_12",
											Example:    "Example_12",
											Synonyms:   []string{"Synonym_121", "Synonym_122"},
											Antonyms:   []string{"Antonym_121", "Antonym_122"},
										},
										{
											Definition: "Definition_13",
											Example:    "Example_13",
											Synonyms:   []string{"Synonym_131", "Synonym_132"},
											Antonyms:   []string{"Antonym_131", "Antonym_132"},
										},
									},
								},
							},
						},
						{
							Word: "Word_2",
							Translation: &entity.Translation{
								Language:     lang.Ukrainian,
								Translations: []string{"Translations_21", "Translations_22"},
							},
							Origin: "Origin_2",
							Phonetics: []entity.Phonetic{
								{Text: "Text_21"},
								{Text: "Text_22"},
							},
							Meanings: []entity.Meaning{
								{
									PartOfSpeech: "PartOfSpeech_21",
									Definitions: []entity.Definition{
										{
											Definition: "Definition_21",
											Example:    "Example_21",
											Synonyms:   []string{"Synonym_211", "Synonym_212"},
											Antonyms:   []string{"Antonym_211", "Antonym_212"},
										},
										{
											Definition: "Definition_22",
											Example:    "Example_22",
											Synonyms:   []string{"Synonym_221", "Synonym_222"},
											Antonyms:   []string{"Antonym_221", "Antonym_222"},
										},
									},
								},
								{
									PartOfSpeech: "PartOfSpeech_12",
									Definitions: []entity.Definition{
										{
											Definition: "Definition_12",
											Example:    "Example_12",
											Synonyms:   []string{"Synonym_121", "Synonym_122"},
											Antonyms:   []string{"Antonym_121", "Antonym_122"},
										},
										{
											Definition: "Definition_13",
											Example:    "Example_13",
											Synonyms:   []string{"Synonym_131", "Synonym_132"},
											Antonyms:   []string{"Antonym_131", "Antonym_132"},
										},
									},
								},
							},
						},
					},
					Params: core.CreateCardParameters{EnrichWordInformationFromDictionary: true},
				},
			},
		},
		"without word info list": {
			input: input{
				req: &api.CreateCardRequest{
					UserID:   "UserID",
					Language: lang.English.String(),
					Params:   &api.CreateCardParameters{EnrichWordInformationFromDictionary: true},
				},
			},
			want: want{
				req: core.CreateCardRequest{
					UserID:   "UserID",
					Language: lang.English,
					Params:   core.CreateCardParameters{EnrichWordInformationFromDictionary: true},
				},
			},
		},
		"without params": {
			input: input{
				req: &api.CreateCardRequest{
					UserID:   "UserID",
					Language: lang.English.String(),
				},
			},
			want: want{
				req: core.CreateCardRequest{
					UserID:   "UserID",
					Language: lang.English,
					Params:   core.CreateCardParameters{EnrichWordInformationFromDictionary: false},
				},
			},
		},
	}
	for name, testcase := range testcases {
		name := name
		testcase := testcase

		t.Run(name, func(t *testing.T) {
			if got := DefaultTransformer.ToCoreCreateCardRequest(testcase.input.req); !reflect.DeepEqual(got, testcase.want.req) {
				t.Fatalf("ToCoreCreateCardRequest() = %v, want %v", got, testcase.want.req)
			}
		})
	}
}

func TestTransformerToAPICreateCardResponse(t *testing.T) {
	t.Parallel()

	coreResp := core.CreateCardResponse{
		Card: entity.Card{
			ID:       "ID_1",
			UserID:   "UserID_1",
			Language: lang.English,
			WordInformationList: []entity.WordInformation{
				{
					Word: "Word_11",
					Translation: &entity.Translation{
						Language:     lang.Ukrainian,
						Translations: []string{"Translation_11", "Translation_12"},
					},
					Origin: "Origin_1",
					Phonetics: []entity.Phonetic{
						{Text: "Text_11"},
						{Text: "Text_12"},
					},
					Meanings: []entity.Meaning{
						{
							PartOfSpeech: "PartOfSpeech_11",
							Definitions: []entity.Definition{
								{
									Definition: "Definition_111",
									Example:    "Example_111",
									Synonyms:   []string{"synonym_1111", "synonym_1112"},
									Antonyms:   []string{"antonym_1111", "antonym_1112"},
								},
								{
									Definition: "Definition_112",
									Example:    "Example_112",
									Synonyms:   []string{"synonym_1121", "synonym_1122"},
									Antonyms:   []string{"antonym_1121", "antonym_1122"},
								},
							},
						},
						{
							PartOfSpeech: "PartOfSpeech_12",
							Definitions: []entity.Definition{
								{
									Definition: "Definition_121",
									Example:    "Example_121",
									Synonyms:   []string{"synonym_1211", "synonym_1212"},
									Antonyms:   []string{"antonym_1211", "antonym_1212"},
								},
								{
									Definition: "Definition_122",
									Example:    "Example_122",
									Synonyms:   []string{"synonym_1221", "synonym_1222"},
									Antonyms:   []string{"antonym_1221", "antonym_1222"},
								},
							},
						},
					},
				},
			},
			CorrectAnswers: 1,
			NextDueDate:    time.Date(2022, 01, 01, 01, 00, 00, 00, time.UTC),
		},
	}

	want := &api.CreateCardResponse{
		Card: &api.Card{
			Id:       "ID_1",
			UserID:   "UserID_1",
			Language: lang.English.String(),
			WordInformationList: []*api.WordInformation{
				{
					Word: "Word_11",
					Translation: &api.Translation{
						Language:     lang.Ukrainian.String(),
						Translations: []string{"Translation_11", "Translation_12"},
					},
					Origin: "Origin_1",
					Phonetics: []*api.Phonetic{
						{Text: "Text_11"},
						{Text: "Text_12"},
					},
					Meanings: []*api.Meaning{
						{
							PartOfSpeech: "PartOfSpeech_11",
							Definitions: []*api.Definition{
								{
									Definition: "Definition_111",
									Example:    "Example_111",
									Synonyms:   []string{"synonym_1111", "synonym_1112"},
									Antonyms:   []string{"antonym_1111", "antonym_1112"},
								},
								{
									Definition: "Definition_112",
									Example:    "Example_112",
									Synonyms:   []string{"synonym_1121", "synonym_1122"},
									Antonyms:   []string{"antonym_1121", "antonym_1122"},
								},
							},
						},
						{
							PartOfSpeech: "PartOfSpeech_12",
							Definitions: []*api.Definition{
								{
									Definition: "Definition_121",
									Example:    "Example_121",
									Synonyms:   []string{"synonym_1211", "synonym_1212"},
									Antonyms:   []string{"antonym_1211", "antonym_1212"},
								},
								{
									Definition: "Definition_122",
									Example:    "Example_122",
									Synonyms:   []string{"synonym_1221", "synonym_1222"},
									Antonyms:   []string{"antonym_1221", "antonym_1222"},
								},
							},
						},
					},
				},
			},
			CorrectAnswers: 1,
			NextDueDate:    timestamppb.New(time.Date(2022, 01, 01, 01, 00, 00, 00, time.UTC)),
		},
	}

	if got := DefaultTransformer.ToAPICreateCardResponse(coreResp); !reflect.DeepEqual(got, want) {
		t.Fatalf("ToAPICreateCardResponse() = %v, want %v", got, want)
	}
}

func TestTransformerToCoreGetCardsRequest(t *testing.T) {
	t.Parallel()

	type (
		input struct{ req *api.GetCardsRequest }
		want  struct{ req core.GetCardsRequest }
	)
	testcases := map[string]struct {
		input input
		want  want
	}{
		"positive case": {
			input: input{
				req: &api.GetCardsRequest{
					UserID:   "UserID",
					Language: lang.English.String(),
				},
			},
			want: want{
				req: core.GetCardsRequest{
					UserID:   "UserID",
					Language: lang.English,
				},
			},
		},
		"nullable input": {
			input: input{req: nil},
			want:  want{req: core.GetCardsRequest{}},
		},
	}
	for name, testcase := range testcases {
		name := name
		testcase := testcase

		t.Run(name, func(t *testing.T) {
			if got := DefaultTransformer.ToCoreGetCardsRequest(testcase.input.req); !reflect.DeepEqual(got, testcase.want.req) {
				t.Errorf("ToCoreGetCardsRequest() = %v, want %v", got, testcase.want.req)
			}
		})
	}
}

func TestTransformerToAPIGetCardsResponse(t *testing.T) {
	t.Parallel()

	type (
		input struct{ resp core.GetCardsResponse }
		want  struct{ resp *api.GetCardsResponse }
	)
	testcases := map[string]struct {
		input input
		want  want
	}{
		"positive case": {
			input: input{
				resp: core.GetCardsResponse{
					UserID:   "UserID",
					Language: lang.English,
					Cards: []entity.Card{
						{
							ID:       "ID_1",
							UserID:   "UserID_1",
							Language: lang.English,
							WordInformationList: []entity.WordInformation{
								{
									Word: "Word_11",
									Translation: &entity.Translation{
										Language:     lang.Ukrainian,
										Translations: []string{"Translation_11", "Translation_12"},
									},
									Origin: "Origin_1",
									Phonetics: []entity.Phonetic{
										{Text: "Text_11"},
										{Text: "Text_12"},
									},
									Meanings: []entity.Meaning{
										{
											PartOfSpeech: "PartOfSpeech_11",
											Definitions: []entity.Definition{
												{
													Definition: "Definition_111",
													Example:    "Example_111",
													Synonyms:   []string{"synonym_1111", "synonym_1112"},
													Antonyms:   []string{"antonym_1111", "antonym_1112"},
												},
												{
													Definition: "Definition_112",
													Example:    "Example_112",
													Synonyms:   []string{"synonym_1121", "synonym_1122"},
													Antonyms:   []string{"antonym_1121", "antonym_1122"},
												},
											},
										},
										{
											PartOfSpeech: "PartOfSpeech_12",
											Definitions: []entity.Definition{
												{
													Definition: "Definition_121",
													Example:    "Example_121",
													Synonyms:   []string{"synonym_1211", "synonym_1212"},
													Antonyms:   []string{"antonym_1211", "antonym_1212"},
												},
												{
													Definition: "Definition_122",
													Example:    "Example_122",
													Synonyms:   []string{"synonym_1221", "synonym_1222"},
													Antonyms:   []string{"antonym_1221", "antonym_1222"},
												},
											},
										},
									},
								},
							},
							CorrectAnswers: 1,
							NextDueDate:    time.Date(2022, 01, 01, 01, 00, 00, 00, time.UTC),
						},
						{
							ID:       "ID_2",
							UserID:   "UserID_2",
							Language: lang.Ukrainian,
							WordInformationList: []entity.WordInformation{
								{
									Word: "Word_21",
									Translation: &entity.Translation{
										Language:     lang.Ukrainian,
										Translations: []string{"Translation_21", "Translation_22"},
									},
									Origin: "Origin_2",
									Phonetics: []entity.Phonetic{
										{Text: "Text_21"},
										{Text: "Text_22"},
									},
									Meanings: []entity.Meaning{
										{
											PartOfSpeech: "PartOfSpeech_21",
											Definitions: []entity.Definition{
												{
													Definition: "Definition_211",
													Example:    "Example_211",
													Synonyms:   []string{"synonym_2111", "synonym_2112"},
													Antonyms:   []string{"antonym_2111", "antonym_2112"},
												},
												{
													Definition: "Definition_212",
													Example:    "Example_212",
													Synonyms:   []string{"synonym_2121", "synonym_2122"},
													Antonyms:   []string{"antonym_2121", "antonym_2122"},
												},
											},
										},
										{
											PartOfSpeech: "PartOfSpeech_22",
											Definitions: []entity.Definition{
												{
													Definition: "Definition_221",
													Example:    "Example_221",
													Synonyms:   []string{"synonym_2211", "synonym_2212"},
													Antonyms:   []string{"antonym_2211", "antonym_2212"},
												},
												{
													Definition: "Definition_222",
													Example:    "Example_222",
													Synonyms:   []string{"synonym_2221", "synonym_2222"},
													Antonyms:   []string{"antonym_2221", "antonym_2222"},
												},
											},
										},
									},
								},
							},
							CorrectAnswers: 1,
							NextDueDate:    time.Date(2022, 01, 02, 01, 00, 00, 00, time.UTC),
						},
					},
				},
			},
			want: want{
				resp: &api.GetCardsResponse{
					UserID:   "UserID",
					Language: lang.English.String(),
					Cards: []*api.Card{
						{
							Id:       "ID_1",
							UserID:   "UserID_1",
							Language: lang.English.String(),
							WordInformationList: []*api.WordInformation{
								{
									Word: "Word_11",
									Translation: &api.Translation{
										Language:     lang.Ukrainian.String(),
										Translations: []string{"Translation_11", "Translation_12"},
									},
									Origin: "Origin_1",
									Phonetics: []*api.Phonetic{
										{Text: "Text_11"},
										{Text: "Text_12"},
									},
									Meanings: []*api.Meaning{
										{
											PartOfSpeech: "PartOfSpeech_11",
											Definitions: []*api.Definition{
												{
													Definition: "Definition_111",
													Example:    "Example_111",
													Synonyms:   []string{"synonym_1111", "synonym_1112"},
													Antonyms:   []string{"antonym_1111", "antonym_1112"},
												},
												{
													Definition: "Definition_112",
													Example:    "Example_112",
													Synonyms:   []string{"synonym_1121", "synonym_1122"},
													Antonyms:   []string{"antonym_1121", "antonym_1122"},
												},
											},
										},
										{
											PartOfSpeech: "PartOfSpeech_12",
											Definitions: []*api.Definition{
												{
													Definition: "Definition_121",
													Example:    "Example_121",
													Synonyms:   []string{"synonym_1211", "synonym_1212"},
													Antonyms:   []string{"antonym_1211", "antonym_1212"},
												},
												{
													Definition: "Definition_122",
													Example:    "Example_122",
													Synonyms:   []string{"synonym_1221", "synonym_1222"},
													Antonyms:   []string{"antonym_1221", "antonym_1222"},
												},
											},
										},
									},
								},
							},
							CorrectAnswers: 1,
							NextDueDate:    timestamppb.New(time.Date(2022, 01, 01, 01, 00, 00, 00, time.UTC)),
						},
						{
							Id:       "ID_2",
							UserID:   "UserID_2",
							Language: lang.Ukrainian.String(),
							WordInformationList: []*api.WordInformation{
								{
									Word: "Word_21",
									Translation: &api.Translation{
										Language:     lang.Ukrainian.String(),
										Translations: []string{"Translation_21", "Translation_22"},
									},
									Origin: "Origin_2",
									Phonetics: []*api.Phonetic{
										{Text: "Text_21"},
										{Text: "Text_22"},
									},
									Meanings: []*api.Meaning{
										{
											PartOfSpeech: "PartOfSpeech_21",
											Definitions: []*api.Definition{
												{
													Definition: "Definition_211",
													Example:    "Example_211",
													Synonyms:   []string{"synonym_2111", "synonym_2112"},
													Antonyms:   []string{"antonym_2111", "antonym_2112"},
												},
												{
													Definition: "Definition_212",
													Example:    "Example_212",
													Synonyms:   []string{"synonym_2121", "synonym_2122"},
													Antonyms:   []string{"antonym_2121", "antonym_2122"},
												},
											},
										},
										{
											PartOfSpeech: "PartOfSpeech_22",
											Definitions: []*api.Definition{
												{
													Definition: "Definition_221",
													Example:    "Example_221",
													Synonyms:   []string{"synonym_2211", "synonym_2212"},
													Antonyms:   []string{"antonym_2211", "antonym_2212"},
												},
												{
													Definition: "Definition_222",
													Example:    "Example_222",
													Synonyms:   []string{"synonym_2221", "synonym_2222"},
													Antonyms:   []string{"antonym_2221", "antonym_2222"},
												},
											},
										},
									},
								},
							},
							CorrectAnswers: 1,
							NextDueDate:    timestamppb.New(time.Date(2022, 01, 02, 01, 00, 00, 00, time.UTC)),
						},
					},
				},
			},
		},
		"empty cards": {
			input: input{
				resp: core.GetCardsResponse{
					UserID:   "UserID",
					Language: lang.English,
				},
			},
			want: want{
				resp: &api.GetCardsResponse{
					UserID:   "UserID",
					Language: lang.English.String(),
				},
			},
		},
	}
	for name, testcase := range testcases {
		name := name
		testcase := testcase

		t.Run(name, func(t *testing.T) {
			if got := DefaultTransformer.ToAPIGetCardsResponse(testcase.input.resp); !reflect.DeepEqual(got, testcase.want.resp) {
				t.Fatalf("ToAPIGetCardsResponse() = %v, want %v", got, testcase.want.resp)
			}
		})
	}
}

func TestTransformerToCoreUpdateCardPerformanceRequest(t *testing.T) {
	t.Parallel()

	type (
		input struct {
			req *api.UpdateCardPerformanceRequest
		}
		want struct {
			req core.UpdateCardPerformanceRequest
		}
	)
	testcases := map[string]struct {
		input input
		want  want
	}{
		"nullable req": {
			input: input{req: nil},
			want:  want{req: core.UpdateCardPerformanceRequest{}},
		},
		"positive case": {
			input: input{
				req: &api.UpdateCardPerformanceRequest{
					UserID:            "UserID",
					CardID:            "CardID",
					PerformanceRating: 1,
				},
			},
			want: want{
				req: core.UpdateCardPerformanceRequest{
					UserID:            "UserID",
					CardID:            "CardID",
					PerformanceRating: 1,
				},
			},
		},
	}
	for name, testcase := range testcases {
		name := name
		testcase := testcase

		t.Run(name, func(t *testing.T) {
			if got := DefaultTransformer.ToCoreUpdateCardPerformanceRequest(testcase.input.req); !reflect.DeepEqual(got, testcase.want.req) {
				t.Fatalf("ToCoreUpdateCardPerformanceRequest() = %v, want %v", got, testcase.want.req)
			}
		})
	}
}

func TestTransformerToAPIUpdateCardPerformanceResponse(t *testing.T) {
	t.Parallel()

	type (
		input struct {
			resp core.UpdateCardPerformanceResponse
		}
		want struct {
			resp *api.UpdateCardPerformanceResponse
		}
	)
	testcases := map[string]struct {
		input input
		want  want
	}{
		"positive case": {
			input: input{
				resp: core.UpdateCardPerformanceResponse{
					NextDueDate: time.Date(2022, 2, 24, 0, 0, 0, 0, time.UTC),
				},
			},
			want: want{
				resp: &api.UpdateCardPerformanceResponse{
					NextDueDate: timestamppb.New(time.Date(2022, 2, 24, 0, 0, 0, 0, time.UTC)),
				},
			},
		},
	}
	for name, testcase := range testcases {
		name := name
		testcase := testcase

		t.Run(name, func(t *testing.T) {
			if got := DefaultTransformer.ToAPIUpdateCardPerformanceResponse(testcase.input.resp); !reflect.DeepEqual(got, testcase.want.resp) {
				t.Fatalf("ToAPIUpdateCardPerformanceResponse() = %v, want %v", got, testcase.want.resp)
			}
		})
	}
}

func TestTransformerToCoreGetCardsForReviewRequest(t *testing.T) {
	t.Parallel()

	type (
		input struct{ req *api.GetCardsForReviewRequest }
		want  struct{ req core.GetCardsForReviewRequest }
	)
	testcases := map[string]struct {
		input input
		want  want
	}{
		"nullable req": {
			input: input{req: nil},
			want:  want{req: core.GetCardsForReviewRequest{}},
		},
		"positive case": {
			input: input{
				req: &api.GetCardsForReviewRequest{
					UserID:   "UserID",
					Language: lang.English.String(),
				},
			},
			want: want{
				req: core.GetCardsForReviewRequest{
					UserID:   "UserID",
					Language: lang.English,
				},
			},
		},
	}
	for name, testcase := range testcases {
		name := name
		testcase := testcase

		t.Run(name, func(t *testing.T) {
			if got := DefaultTransformer.ToCoreGetCardsForReviewRequest(testcase.input.req); !reflect.DeepEqual(got, testcase.want.req) {
				t.Fatalf("ToCoreGetCardsForReviewRequest() = %v, want %v", got, testcase.want.req)
			}
		})
	}
}

func TestTransformerToCoreDeleteCardRequest(t *testing.T) {
	t.Parallel()

	type (
		input struct{ req *api.DeleteCardRequest }
		want  struct{ req core.DeleteCardRequest }
	)
	testcases := map[string]struct {
		input input
		want  want
	}{
		"nullable req": {
			input: input{req: nil},
			want:  want{req: core.DeleteCardRequest{}},
		},
		"positive case": {
			input: input{req: &api.DeleteCardRequest{UserID: "UserID", CardID: "CardID"}},
			want:  want{req: core.DeleteCardRequest{UserID: "UserID", CardID: "CardID"}},
		},
	}
	for name, testcase := range testcases {
		name := name
		testcase := testcase

		t.Run(name, func(t *testing.T) {
			if got := DefaultTransformer.ToCoreDeleteCardRequest(testcase.input.req); !reflect.DeepEqual(got, testcase.want.req) {
				t.Fatalf("ToCoreDeleteCardRequest() = %v, want %v", got, testcase.want.req)
			}
		})
	}
}

func TestTransformerToAPIDeleteCardResponse(t *testing.T) {
	t.Parallel()

	type (
		input struct{ resp core.DeleteCardResponse }
		want  struct{ resp *api.DeleteCardResponse }
	)
	testcases := map[string]struct {
		input input
		want  want
	}{
		"positive case": {
			input: input{
				resp: core.DeleteCardResponse{
					Card: entity.Card{
						ID:       "ID_1",
						UserID:   "UserID_1",
						Language: lang.English,
						WordInformationList: []entity.WordInformation{
							{
								Word: "Word_11",
								Translation: &entity.Translation{
									Language:     lang.Ukrainian,
									Translations: []string{"Translation_11", "Translation_12"},
								},
								Origin: "Origin_1",
								Phonetics: []entity.Phonetic{
									{Text: "Text_11"},
									{Text: "Text_12"},
								},
								Meanings: []entity.Meaning{
									{
										PartOfSpeech: "PartOfSpeech_11",
										Definitions: []entity.Definition{
											{
												Definition: "Definition_111",
												Example:    "Example_111",
												Synonyms:   []string{"synonym_1111", "synonym_1112"},
												Antonyms:   []string{"antonym_1111", "antonym_1112"},
											},
											{
												Definition: "Definition_112",
												Example:    "Example_112",
												Synonyms:   []string{"synonym_1121", "synonym_1122"},
												Antonyms:   []string{"antonym_1121", "antonym_1122"},
											},
										},
									},
									{
										PartOfSpeech: "PartOfSpeech_12",
										Definitions: []entity.Definition{
											{
												Definition: "Definition_121",
												Example:    "Example_121",
												Synonyms:   []string{"synonym_1211", "synonym_1212"},
												Antonyms:   []string{"antonym_1211", "antonym_1212"},
											},
											{
												Definition: "Definition_122",
												Example:    "Example_122",
												Synonyms:   []string{"synonym_1221", "synonym_1222"},
												Antonyms:   []string{"antonym_1221", "antonym_1222"},
											},
										},
									},
								},
								Audio: []byte("Audio_11"),
							},
						},
						CorrectAnswers: 1,
						NextDueDate:    time.Date(2022, 01, 01, 01, 00, 00, 00, time.UTC),
					},
				},
			},
			want: want{
				resp: &api.DeleteCardResponse{
					Card: &api.Card{
						Id:       "ID_1",
						UserID:   "UserID_1",
						Language: lang.English.String(),
						WordInformationList: []*api.WordInformation{
							{
								Word: "Word_11",
								Translation: &api.Translation{
									Language:     lang.Ukrainian.String(),
									Translations: []string{"Translation_11", "Translation_12"},
								},
								Origin: "Origin_1",
								Phonetics: []*api.Phonetic{
									{Text: "Text_11"},
									{Text: "Text_12"},
								},
								Meanings: []*api.Meaning{
									{
										PartOfSpeech: "PartOfSpeech_11",
										Definitions: []*api.Definition{
											{
												Definition: "Definition_111",
												Example:    "Example_111",
												Synonyms:   []string{"synonym_1111", "synonym_1112"},
												Antonyms:   []string{"antonym_1111", "antonym_1112"},
											},
											{
												Definition: "Definition_112",
												Example:    "Example_112",
												Synonyms:   []string{"synonym_1121", "synonym_1122"},
												Antonyms:   []string{"antonym_1121", "antonym_1122"},
											},
										},
									},
									{
										PartOfSpeech: "PartOfSpeech_12",
										Definitions: []*api.Definition{
											{
												Definition: "Definition_121",
												Example:    "Example_121",
												Synonyms:   []string{"synonym_1211", "synonym_1212"},
												Antonyms:   []string{"antonym_1211", "antonym_1212"},
											},
											{
												Definition: "Definition_122",
												Example:    "Example_122",
												Synonyms:   []string{"synonym_1221", "synonym_1222"},
												Antonyms:   []string{"antonym_1221", "antonym_1222"},
											},
										},
									},
								},
								Audio: []byte("Audio_11"),
							},
						},
						CorrectAnswers: 1,
						NextDueDate:    timestamppb.New(time.Date(2022, 01, 01, 01, 00, 00, 00, time.UTC)),
					},
				},
			},
		},
	}
	for name, testcase := range testcases {
		name := name
		testcase := testcase

		t.Run(name, func(t *testing.T) {
			if got := DefaultTransformer.ToAPIDeleteCardResponse(testcase.input.resp); !reflect.DeepEqual(got, testcase.want.resp) {
				t.Fatalf("ToAPIDeleteCardResponse() = %v, want %v", got, testcase.want.resp)
			}
		})
	}
}

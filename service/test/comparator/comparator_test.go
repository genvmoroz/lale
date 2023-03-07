package comparator

import (
	"reflect"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/genvmoroz/lale/service/api"
	"github.com/genvmoroz/lale/service/internal/entity"
)

func TestCompareCard(t *testing.T) {
	t.Parallel()

	type args struct {
		card   *entity.Card
		target *api.Card
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "both nullable",
			args: args{
				card:   nil,
				target: nil,
			},
			want: true,
		},
		{
			name: "card nullable",
			args: args{
				card:   nil,
				target: &api.Card{Id: "someID"},
			},
			want: false,
		},
		{
			name: "target nullable",
			args: args{
				card:   &entity.Card{ID: "someID"},
				target: nil,
			},
			want: false,
		},
		{
			name: "equal ID",
			args: args{
				card:   &entity.Card{ID: "someID"},
				target: &api.Card{Id: "someID"},
			},
			want: true,
		},
		{
			name: "not equal ID",
			args: args{
				card:   &entity.Card{ID: "someID"},
				target: &api.Card{Id: "anotherID"},
			},
			want: false,
		},
		{
			name: "equal UserID",
			args: args{
				card:   &entity.Card{UserID: "someID"},
				target: &api.Card{UserID: "someID"},
			},
			want: true,
		},
		{
			name: "not equal UserID",
			args: args{
				card:   &entity.Card{UserID: "someID"},
				target: &api.Card{UserID: "anotherID"},
			},
			want: false,
		},
		{
			name: "equal lang",
			args: args{
				card:   &entity.Card{Language: "lang"},
				target: &api.Card{Language: "lang"},
			},
			want: true,
		},
		{
			name: "not equal lang",
			args: args{
				card:   &entity.Card{Language: "lang"},
				target: &api.Card{Language: "anotherLanguage"},
			},
			want: false,
		},
		{
			name: "equal NextDueDate",
			args: args{
				card: &entity.Card{
					NextDueDate: time.Date(2022, 2, 24, 0, 0, 0, 0, time.UTC),
				},
				target: &api.Card{
					NextDueDate: timestamppb.New(
						time.Date(2022, 2, 24, 0, 0, 0, 0, time.UTC),
					),
				},
			},
			want: true,
		},
		{
			name: "not equal NextDueDate",
			args: args{
				card: &entity.Card{
					NextDueDate: time.Date(2022, 2, 24, 0, 0, 0, 0, time.UTC),
				},
				target: &api.Card{
					NextDueDate: timestamppb.New(
						time.Date(2023, 2, 24, 0, 0, 0, 0, time.UTC),
					),
				},
			},
			want: false,
		},
		{
			name: "equal Words",
			args: args{
				card: &entity.Card{
					WordInformationList: []entity.WordInformation{
						{
							Word:        "someWord0",
							Origin:      "origin",
							Translation: &entity.Translation{Language: "en", Translations: []string{"trans0"}},
						},
						{
							Word:        "someWord1",
							Origin:      "origin",
							Translation: &entity.Translation{Language: "en", Translations: []string{"trans1"}},
						},
					},
				},
				target: &api.Card{
					WordInformationList: []*api.WordInformation{
						{
							Word:        "someWord0",
							Origin:      "origin",
							Translation: &api.Translation{Language: "en", Translations: []string{"trans0"}},
						},
						{
							Word:        "someWord1",
							Origin:      "origin",
							Translation: &api.Translation{Language: "en", Translations: []string{"trans1"}},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "not equal Words",
			args: args{
				card: &entity.Card{
					WordInformationList: []entity.WordInformation{
						{
							Word:   "someWord0",
							Origin: "origin",
						},
						{
							Word:   "someWord1",
							Origin: "origin",
						},
					},
				},
				target: &api.Card{
					WordInformationList: []*api.WordInformation{
						{
							Word:   "someWord0",
							Origin: "origin",
						},
						{
							Word:   "someWord2",
							Origin: "origin",
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gr := &GRPCComparator{}
			if got := gr.CompareCard(tt.args.card, tt.args.target); got != tt.want {
				t.Errorf("CompareCard() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareDefinition(t *testing.T) {
	t.Parallel()

	type args struct {
		definition *entity.Definition
		target     *api.Definition
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "equal definitions",
			args: args{
				definition: &entity.Definition{
					Definition: "definition0",
					Example:    "example0",
					Synonyms:   []string{"synonym0"},
					Antonyms:   []string{"antonym0"},
				},
				target: &api.Definition{
					Definition: "definition0",
					Example:    "example0",
					Synonyms:   []string{"synonym0"},
					Antonyms:   []string{"antonym0"},
				},
			},
			want: true,
		},
		{
			name: "nullable synonyms",
			args: args{
				definition: &entity.Definition{
					Definition: "definition0",
					Example:    "example0",
					Synonyms:   []string{"synonym0"},
					Antonyms:   []string{"antonym0"},
				},
				target: &api.Definition{
					Definition: "definition0",
					Example:    "example0",
					Antonyms:   []string{"antonym0"},
				},
			},
			want: false,
		},
		{
			name: "nullable synonyms and antonyms",
			args: args{
				definition: &entity.Definition{
					Definition: "definition0",
					Example:    "example0",
				},
				target: &api.Definition{
					Definition: "definition0",
					Example:    "example0",
					Antonyms:   []string{"antonym0"},
				},
			},
			want: false,
		},
		{
			name: "nullable definition",
			args: args{
				definition: nil,
				target:     &api.Definition{},
			},
			want: false,
		},
		{
			name: "nullable target",
			args: args{
				definition: &entity.Definition{},
				target:     nil,
			},
			want: false,
		},
		{
			name: "nullable both",
			args: args{
				definition: nil,
				target:     nil,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gr := &GRPCComparator{}
			if got := gr.CompareDefinition(tt.args.definition, tt.args.target); got != tt.want {
				t.Errorf("CompareDefinition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareWord(t *testing.T) {
	t.Parallel()

	type args struct {
		word   *entity.WordInformation
		target *api.WordInformation
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "both nullable",
			args: args{
				word:   nil,
				target: nil,
			},
			want: true,
		},
		{
			name: "word nullable",
			args: args{
				word:   nil,
				target: &api.WordInformation{},
			},
			want: false,
		},
		{
			name: "target nullable",
			args: args{
				word:   &entity.WordInformation{},
				target: nil,
			},
			want: false,
		},
		{
			name: "equal word",
			args: args{
				word:   &entity.WordInformation{Word: "word"},
				target: &api.WordInformation{Word: "word"},
			},
			want: true,
		},
		{
			name: "not equal word",
			args: args{
				word:   &entity.WordInformation{Word: "word"},
				target: &api.WordInformation{Word: "anotherWord"},
			},
			want: false,
		},
		{
			name: "equal Translation",
			args: args{
				word:   &entity.WordInformation{Translation: &entity.Translation{Language: "en", Translations: []string{"trans"}}},
				target: &api.WordInformation{Translation: &api.Translation{Language: "en", Translations: []string{"trans"}}},
			},
			want: true,
		},
		{
			name: "not equal Translation",
			args: args{
				word:   &entity.WordInformation{Translation: &entity.Translation{Language: "en", Translations: []string{"trans"}}},
				target: &api.WordInformation{Translation: &api.Translation{Language: "en", Translations: []string{"anotherTrans"}}},
			},
			want: false,
		},
		{
			name: "equal origin",
			args: args{
				word:   &entity.WordInformation{Origin: "origin"},
				target: &api.WordInformation{Origin: "origin"},
			},
			want: true,
		},
		{
			name: "not equal origin",
			args: args{
				word:   &entity.WordInformation{Origin: "origin"},
				target: &api.WordInformation{Origin: "anotherOrigin"},
			},
			want: false,
		},
		{
			name: "equal phonetics",
			args: args{
				word: &entity.WordInformation{
					Phonetics: []entity.Phonetic{
						{Text: "text0", AudioLink: "link0"},
						{Text: "text1", AudioLink: "link1"},
					},
				},
				target: &api.WordInformation{
					Phonetics: []*api.Phonetic{
						{Text: "text0", AudioLink: "link0"},
						{Text: "text1", AudioLink: "link1"},
					},
				},
			},
			want: true,
		},
		{
			name: "not equal phonetics",
			args: args{
				word: &entity.WordInformation{
					Phonetics: []entity.Phonetic{
						{Text: "text0", AudioLink: "link0"},
					},
				},
				target: &api.WordInformation{
					Phonetics: []*api.Phonetic{
						{Text: "text0", AudioLink: "link0"},
						{Text: "text1", AudioLink: "link1"},
					},
				},
			},
			want: false,
		},
		{
			name: "equal meaning",
			args: args{
				word: &entity.WordInformation{
					Meanings: []entity.Meaning{
						{
							PartOfSpeech: "verb",
							Definitions: []entity.Definition{
								{
									Definition: "def0",
									Example:    "example0",
									Synonyms:   []string{"syn0", "syn1"},
									Antonyms:   []string{"an0", "an1", "an2"},
								},
								{
									Definition: "def1",
									Example:    "example1",
									Synonyms:   []string{"syn2"},
									Antonyms:   []string{"an3", "an4"},
								},
							},
						},
					},
				},
				target: &api.WordInformation{
					Meanings: []*api.Meaning{
						{
							PartOfSpeech: "verb",
							Definitions: []*api.Definition{
								{
									Definition: "def0",
									Example:    "example0",
									Synonyms:   []string{"syn0", "syn1"},
									Antonyms:   []string{"an0", "an1", "an2"},
								},
								{
									Definition: "def1",
									Example:    "example1",
									Synonyms:   []string{"syn2"},
									Antonyms:   []string{"an3", "an4"},
								},
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "not equal meaning 0",
			args: args{
				word: &entity.WordInformation{
					Meanings: []entity.Meaning{
						{
							PartOfSpeech: "verb",
							Definitions: []entity.Definition{
								{
									Definition: "def1",
									Example:    "example1",
									Synonyms:   []string{"syn2"},
									Antonyms:   []string{"an3", "an4"},
								},
								{
									Definition: "def0",
									Example:    "example0",
									Synonyms:   []string{"syn0", "syn1"},
									Antonyms:   []string{"an0", "an1", "an2"},
								},
							},
						},
					},
				},
				target: &api.WordInformation{
					Meanings: []*api.Meaning{
						{
							PartOfSpeech: "verb",
							Definitions: []*api.Definition{
								{
									Definition: "def0",
									Example:    "example0",
									Synonyms:   []string{"syn0", "syn1"},
									Antonyms:   []string{"an0", "an1", "an2"},
								},
								{
									Definition: "def1",
									Example:    "example1",
									Synonyms:   []string{"syn2"},
									Antonyms:   []string{"an3", "an4"},
								},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "not equal meaning 1",
			args: args{
				word: &entity.WordInformation{
					Meanings: []entity.Meaning{
						{
							PartOfSpeech: "verb",
							Definitions: []entity.Definition{
								{
									Definition: "def0",
									Example:    "example0",
									Synonyms:   []string{"syn0", "syn1"},
									Antonyms:   []string{"an0", "an1", "an2"},
								},
								{
									Definition: "def1",
									Example:    "example1",
									Synonyms:   []string{"syn2"},
									Antonyms:   []string{"an3"},
								},
							},
						},
					},
				},
				target: &api.WordInformation{
					Meanings: []*api.Meaning{
						{
							PartOfSpeech: "verb",
							Definitions: []*api.Definition{
								{
									Definition: "def0",
									Example:    "example0",
									Synonyms:   []string{"syn0"},
									Antonyms:   []string{"an0", "an1", "an2"},
								},
								{
									Definition: "def1",
									Example:    "example1",
									Synonyms:   []string{"syn2"},
									Antonyms:   []string{"an3", "an4"},
								},
							},
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gr := &GRPCComparator{}
			if got := gr.CompareWordInformation(tt.args.word, tt.args.target); got != tt.want {
				t.Errorf("CompareWord() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComparePhonetic(t *testing.T) {
	t.Parallel()

	type args struct {
		phonetic *entity.Phonetic
		target   *api.Phonetic
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "both nullable",
			args: args{
				phonetic: nil,
				target:   nil,
			},
			want: true,
		},
		{
			name: "phonetic nullable",
			args: args{
				phonetic: nil,
				target:   &api.Phonetic{},
			},
			want: false,
		},
		{
			name: "target nullable",
			args: args{
				phonetic: &entity.Phonetic{},
				target:   nil,
			},
			want: false,
		},
		{
			name: "equal text",
			args: args{
				phonetic: &entity.Phonetic{Text: "text"},
				target:   &api.Phonetic{Text: "text"},
			},
			want: true,
		},
		{
			name: "not equal text",
			args: args{
				phonetic: &entity.Phonetic{Text: "text"},
				target:   &api.Phonetic{Text: "anotherText"},
			},
			want: false,
		},
		{
			name: "equal audio link",
			args: args{
				phonetic: &entity.Phonetic{AudioLink: "link"},
				target:   &api.Phonetic{AudioLink: "link"},
			},
			want: true,
		},
		{
			name: "not equal audio link",
			args: args{
				phonetic: &entity.Phonetic{AudioLink: "link"},
				target:   &api.Phonetic{AudioLink: "anotherLink"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gr := &GRPCComparator{}
			if got := gr.ComparePhonetic(tt.args.phonetic, tt.args.target); got != tt.want {
				t.Errorf("ComparePhonetic() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareMeaning(t *testing.T) {
	t.Parallel()

	type args struct {
		meaning *entity.Meaning
		target  *api.Meaning
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "both nullable",
			args: args{
				meaning: nil,
				target:  nil,
			},
			want: true,
		},
		{
			name: "meaning nullable",
			args: args{
				meaning: nil,
				target:  &api.Meaning{},
			},
			want: false,
		},
		{
			name: "target nullable",
			args: args{
				meaning: &entity.Meaning{},
				target:  nil,
			},
			want: false,
		},
		{
			name: "equal part of speech",
			args: args{
				meaning: &entity.Meaning{PartOfSpeech: "part"},
				target:  &api.Meaning{PartOfSpeech: "part"},
			},
			want: true,
		},
		{
			name: "not equal part of speech",
			args: args{
				meaning: &entity.Meaning{PartOfSpeech: "part"},
				target:  &api.Meaning{PartOfSpeech: "anotherPart"},
			},
			want: false,
		},
		{
			name: "equal part of speech",
			args: args{
				meaning: &entity.Meaning{Definitions: []entity.Definition{{Definition: "def"}}},
				target:  &api.Meaning{Definitions: []*api.Definition{{Definition: "def"}}},
			},
			want: true,
		},
		{
			name: "not equal part of speech",
			args: args{
				meaning: &entity.Meaning{Definitions: []entity.Definition{{Definition: "def"}}},
				target:  &api.Meaning{Definitions: []*api.Definition{{Definition: "anotherDef"}}},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gr := &GRPCComparator{}
			if got := gr.CompareMeaning(tt.args.meaning, tt.args.target); got != tt.want {
				t.Errorf("CompareMeaning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewGRPCComparator(t *testing.T) {
	t.Parallel()

	want := empty

	if got := NewGRPCComparator(); !reflect.DeepEqual(got, want) {
		t.Errorf("NewGRPCComparator() = %v, want %v", got, want)
	}
}

func TestGRPCComparator_CompareTranslation(t *testing.T) {
	t.Parallel()

	type args struct {
		Translation *entity.Translation
		target      *api.Translation
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "both nullable",
			args: args{
				Translation: nil,
				target:      nil,
			},
			want: true,
		},
		{
			name: "Translation nullable",
			args: args{
				Translation: nil,
				target:      &api.Translation{},
			},
			want: false,
		},
		{
			name: "target nullable",
			args: args{
				Translation: &entity.Translation{},
				target:      nil,
			},
			want: false,
		},
		{
			name: "equal lang",
			args: args{
				Translation: &entity.Translation{Language: "uk"},
				target:      &api.Translation{Language: "uk"},
			},
			want: true,
		},
		{
			name: "not equal lang",
			args: args{
				Translation: &entity.Translation{Language: "uk"},
				target:      &api.Translation{Language: "gb"},
			},
			want: false,
		},
		{
			name: "equal Translations",
			args: args{
				Translation: &entity.Translation{Translations: []string{"trans0", "trans1"}},
				target:      &api.Translation{Translations: []string{"trans0", "trans1"}},
			},
			want: true,
		},
		{
			name: "not equal Translations",
			args: args{
				Translation: &entity.Translation{Translations: []string{"trans0", "trans1"}},
				target:      &api.Translation{Translations: []string{"trans0"}},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gr := &GRPCComparator{}
			if got := gr.CompareTranslation(tt.args.Translation, tt.args.target); got != tt.want {
				t.Errorf("CompareTranslation() = %v, want %v", got, tt.want)
			}
		})
	}
}

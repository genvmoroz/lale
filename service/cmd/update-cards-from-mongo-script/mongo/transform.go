package mongo

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/text/language"
)

type (
	Card struct {
		ID       string
		UserID   string
		Language language.Tag

		WordInformationList []WordInformation `yaml:"WordInformationList,omitempty"`

		ConsecutiveCorrectAnswersNumber uint32
		NextDueDate                     time.Time
	}

	WordInformation struct {
		Word        string       `yaml:"Word,omitempty"`
		Translation *Translation `yaml:"Translation,omitempty"`
		Origin      string       `yaml:"Origin,omitempty"`
		Phonetics   []Phonetic   `yaml:"Phonetics,omitempty"`
		Meanings    []Meaning    `yaml:"Meanings,omitempty"`
		Audio       []byte       `yaml:"Audio,omitempty"`
	}

	Translation struct {
		Language     language.Tag `yaml:"Language,omitempty"`
		Translations []string     `yaml:"Translations,omitempty"`
	}

	Phonetic struct {
		Text string `yaml:"Text,omitempty"`
	}

	Meaning struct {
		PartOfSpeech string       `yaml:"PartOfSpeech,omitempty"`
		Definitions  []Definition `yaml:"Definitions,omitempty"`
	}

	Definition struct {
		Definition string   `yaml:"Definition,omitempty"`
		Example    string   `yaml:"Example,omitempty"`
		Synonyms   []string `yaml:"Synonyms,omitempty"`
		Antonyms   []string `yaml:"Antonyms,omitempty"`
	}
)

func cardToDoc(card Card) ([]byte, error) {
	buf := new(bytes.Buffer)
	w, err := bsonrw.NewBSONValueWriter(buf)
	if err != nil {
		return nil, fmt.Errorf("new bson writer: %w", err)
	}
	encoder, err := bson.NewEncoder(w)
	if err != nil {
		return nil, fmt.Errorf("new bson writer: %w", err)
	}
	if err = encoder.SetRegistry(defaultCustomRegistry); err != nil {
		return nil, fmt.Errorf("set registry: %w", err)
	}
	if err = encoder.Encode(card); err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}

	return buf.Bytes(), nil
}

func unmarshalCursor(ctx context.Context, cursor *mongo.Cursor) ([]Card, error) {
	var cards []Card

	for cursor.Next(ctx) {
		if cursor.Err() != nil {
			return nil, fmt.Errorf("cursor error: %w", cursor.Err())
		}

		decoder, err := bson.NewDecoder(bsonrw.NewBSONDocumentReader(cursor.Current))
		if err != nil {
			return nil, fmt.Errorf("new decoder: %w", err)
		}
		if err = decoder.SetRegistry(defaultCustomRegistry); err != nil {
			return nil, fmt.Errorf("set registry: %w", err)
		}
		card := Card{}
		if err = decoder.Decode(&card); err != nil {
			return nil, fmt.Errorf("decode: %w", err)
		}
		cards = append(cards, card)
	}

	return cards, nil
}

var defaultCustomRegistry = createCustomRegistry() //nolint:gochecknoglobals // it's fine to have here

func createCustomRegistry() *bsoncodec.Registry {
	registry := bson.DefaultRegistry
	registry.RegisterTypeEncoder(
		reflect.TypeOf(language.Tag{}),
		bsoncodec.ValueEncoderFunc(func(_ bsoncodec.EncodeContext, writer bsonrw.ValueWriter, value reflect.Value) error {
			if lang, ok := value.Interface().(language.Tag); ok {
				return writer.WriteString(lang.String())
			}
			return fmt.Errorf("convert (%s) to (%T) error", value.Type().String(), language.Tag{})
		}),
	)
	registry.RegisterTypeDecoder(
		reflect.TypeOf(language.Tag{}),
		bsoncodec.ValueDecoderFunc(func(_ bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
			read, err := vr.ReadString()
			if err != nil {
				return err
			}
			lang, err := language.Parse(read)
			if err != nil {
				return err
			}
			val.Set(reflect.ValueOf(lang))
			return nil
		}),
	)

	return registry
}

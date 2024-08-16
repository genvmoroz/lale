package card

import (
	"bytes"
	"context"
	"fmt"
	"reflect"

	"github.com/genvmoroz/lale/service/pkg/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/text/language"
)

// todo: implement dto here
// todo: improve it

type transformer struct {
	registry *bsoncodec.Registry
}

func newTransformer() transformer {
	registry := bson.NewRegistry()
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

	return transformer{registry: registry}
}

func (t transformer) cardToDoc(card entity.Card) ([]byte, error) {
	buf := new(bytes.Buffer)
	w, err := bsonrw.NewBSONValueWriter(buf)
	if err != nil {
		return nil, fmt.Errorf("new bson writer: %w", err)
	}
	encoder, err := bson.NewEncoder(w)
	if err != nil {
		return nil, fmt.Errorf("new bson writer: %w", err)
	}
	if err = encoder.SetRegistry(t.registry); err != nil {
		return nil, fmt.Errorf("set registry: %w", err)
	}
	if err = encoder.Encode(card); err != nil {
		return nil, fmt.Errorf("encode: %w", err)
	}

	return buf.Bytes(), nil
}

func (t transformer) unmarshalCursor(ctx context.Context, cursor *mongo.Cursor) ([]entity.Card, error) {
	var cards []entity.Card

	for cursor.Next(ctx) {
		if cursor.Err() != nil {
			return nil, fmt.Errorf("cursor error: %w", cursor.Err())
		}

		decoder, err := bson.NewDecoder(bsonrw.NewBSONDocumentReader(cursor.Current))
		if err != nil {
			return nil, fmt.Errorf("new decoder: %w", err)
		}
		if err = decoder.SetRegistry(t.registry); err != nil {
			return nil, fmt.Errorf("set registry: %w", err)
		}
		card := entity.Card{}
		if err = decoder.Decode(&card); err != nil {
			return nil, fmt.Errorf("decode: %w", err)
		}
		cards = append(cards, card)
	}

	return cards, nil
}

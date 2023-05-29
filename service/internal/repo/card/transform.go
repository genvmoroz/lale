package card

import (
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

func cardToDoc(card entity.Card) ([]byte, error) {
	return bson.MarshalWithRegistry(defaultCustomRegistry, card)
}

func unmarshalCursor(ctx context.Context, cursor *mongo.Cursor) ([]entity.Card, error) {
	var cards []entity.Card

	for data := cursor.Current; cursor.Next(ctx); data = cursor.Current {
		if len(data) == 0 {
			continue
		}
		card := entity.Card{}
		if err := bson.UnmarshalWithRegistry(defaultCustomRegistry, data, &card); err != nil {
			return nil, fmt.Errorf("unmarshal: %w", err)
		}
		cards = append(cards, card)
	}

	return cards, nil
}

var defaultCustomRegistry = createCustomRegistry()

func createCustomRegistry() *bsoncodec.Registry {
	rb := bsoncodec.NewRegistryBuilder()
	bsoncodec.DefaultValueEncoders{}.RegisterDefaultEncoders(rb)
	bsoncodec.DefaultValueDecoders{}.RegisterDefaultDecoders(rb)
	rb.RegisterTypeEncoder(
		reflect.TypeOf(language.Tag{}),
		bsoncodec.ValueEncoderFunc(func(_ bsoncodec.EncodeContext, writer bsonrw.ValueWriter, value reflect.Value) error {
			if lang, ok := value.Interface().(language.Tag); ok {
				return writer.WriteString(lang.String())
			}
			return fmt.Errorf("convert (%s) to (%T) error", value.Type().String(), language.Tag{})
		}),
	)
	rb.RegisterTypeDecoder(
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

	bson.PrimitiveCodecs{}.RegisterPrimitiveCodecs(rb)

	return rb.Build()
}

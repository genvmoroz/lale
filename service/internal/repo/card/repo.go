package card

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/genvmoroz/lale/service/pkg/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type (
	Config struct {
		Protocol   string            `envconfig:"APP_MONGO_CARD_PROTOCOL" required:"true"`
		Host       string            `envconfig:"APP_MONGO_CARD_HOST" required:"true"`
		Port       *int              `envconfig:"APP_MONGO_CARD_PORT"`
		Params     map[string]string `envconfig:"APP_MONGO_CARD_URI_PARAMS" required:"true"`
		Database   string            `envconfig:"APP_MONGO_CARD_DATABASE" required:"true"`
		Collection string            `envconfig:"APP_MONGO_CARD_COLLECTION" required:"true"`

		Creds Creds
	}

	Creds struct {
		User string `envconfig:"APP_MONGO_USER" required:"true"`
		Pass string `envconfig:"APP_MONGO_PASS" required:"true"`
	}

	Repo struct {
		opts       *options.ClientOptions
		database   string
		collection string
	}
)

func NewRepo(ctx context.Context, cfg Config) (*Repo, error) {
	uri := prepareURI(cfg)

	repo := &Repo{
		opts:       options.Client().ApplyURI(uri),
		database:   cfg.Database,
		collection: cfg.Collection,
	}

	if err := repo.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping mongo: %w", err)
	}

	return repo, nil
}

func prepareURI(cfg Config) string {
	uri := strings.Builder{}

	uri.WriteString(fmt.Sprintf("%s://", cfg.Protocol))
	uri.WriteString(fmt.Sprintf("%s:%s", cfg.Creds.User, cfg.Creds.Pass))
	uri.WriteString(fmt.Sprintf("@%s", cfg.Host))

	if cfg.Port != nil {
		uri.WriteString(fmt.Sprintf(":%d", *cfg.Port))
	}

	uri.WriteString("/")

	if len(cfg.Params) != 0 {
		uri.WriteString("?")

		firstParam := true
		for k, v := range cfg.Params {
			if !firstParam {
				uri.WriteString("&")
			}
			uri.WriteString(fmt.Sprintf("%s=%s", k, v))
			firstParam = false
		}
	}

	return uri.String()
}

func (r *Repo) GetCardsForUser(ctx context.Context, userID string) ([]entity.Card, error) {
	if !utf8.ValidString(userID) {
		return nil, fmt.Errorf("userID [%s] is invalid utf8 string", userID)
	}

	client, err := mongo.Connect(ctx, r.opts)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	defer func() {
		disconnect(ctx, client)
	}()

	cardsCollection := client.
		Database(r.database).
		Collection(r.collection)

	query := bson.M{"userid": userID}
	count, err := cardsCollection.EstimatedDocumentCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("estimate document count: %w", err)
	}
	const extraSize = 100
	cursor, err := cardsCollection.Find(ctx, query, options.Find().SetBatchSize(int32(count)+extraSize))
	if err != nil {
		return nil, fmt.Errorf("find: %w", err)
	}
	defer func() {
		_ = cursor.Close(context.Background())
	}()

	return unmarshalCursor(ctx, cursor)
}

// TODO: implement search card by name on Repo side
// func (r *Repo) FindCardsByWord(ctx context.Context, userID, word string) ([]entity.Card, error) {
//	return nil, errors.New("not implemented yet")
// }

//nolint:gocognit // need to fix later
func (r *Repo) SaveCards(ctx context.Context, cards []entity.Card) error {
	if len(cards) == 0 {
		return nil
	}

	dupls := lo.FindDuplicatesBy[entity.Card, string](
		cards,
		func(item entity.Card) string {
			return item.ID
		},
	)
	if len(dupls) != 0 {
		return fmt.Errorf("provided cards contain duplicates: %v", dupls)
	}

	client, err := mongo.Connect(ctx, r.opts)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer func() {
		disconnect(ctx, client)
	}()

	cardsCollection := client.
		Database(r.database).
		Collection(r.collection)

	if err = client.UseSession(ctx, func(sessionContext mongo.SessionContext) error {
		if err = sessionContext.StartTransaction(); err != nil {
			return err
		}

		for _, card := range cards {
			var doc []byte
			doc, err = cardToDoc(card)
			if err != nil {
				return fmt.Errorf("marshal: %w", err)
			}

			// check card existence
			filter := bson.M{"id": card.ID}
			if err = cardsCollection.FindOne(ctx, filter).Err(); err != nil {
				if errors.Is(err, mongo.ErrNoDocuments) {
					// since card does not exist do insert
					_, err = cardsCollection.InsertOne(ctx, doc)
					if err != nil {
						abortTransaction(ctx, sessionContext)
						return fmt.Errorf("insert: %w", err)
					}
				} else {
					// checking existence failed with unpredictable error
					abortTransaction(ctx, sessionContext)
					return fmt.Errorf("check card existence: %w", err)
				}
			} else {
				// since card already exists do replacement
				_, err = cardsCollection.ReplaceOne(ctx, filter, doc)
				if err != nil {
					abortTransaction(ctx, sessionContext)
					return fmt.Errorf("replace card: %w", err)
				}
			}
		}

		return sessionContext.CommitTransaction(ctx)
	}); err != nil {
		return fmt.Errorf("perform transaction: %w", err)
	}

	return nil
}

func (r *Repo) DeleteCard(ctx context.Context, cardID string) error {
	if !utf8.ValidString(cardID) {
		return fmt.Errorf("cardID [%s] is invalid utf8 string", cardID)
	}

	client, err := mongo.Connect(ctx, r.opts)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	defer func() {
		disconnect(ctx, client)
	}()

	cardsCollection := client.
		Database(r.database).
		Collection(r.collection)

	query := bson.M{"id": cardID}
	_, err = cardsCollection.DeleteOne(ctx, query)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

func (r *Repo) Ping(ctx context.Context) error {
	client, err := mongo.Connect(ctx, r.opts)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	defer func() {
		disconnect(ctx, client)
	}()

	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("ping: %w", err)
	}

	return nil
}

func disconnect(ctx context.Context, client *mongo.Client) {
	if err := client.Disconnect(ctx); err != nil {
		logrus.Errorf("disconnect: %s", err.Error())
	}
}

func abortTransaction(ctx context.Context, sessionContext mongo.SessionContext) {
	if err := sessionContext.AbortTransaction(ctx); err != nil {
		logrus.Errorf("abort transaction: %s", err.Error())
	}
}

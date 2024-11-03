package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"golang.org/x/text/language"
)

type (

	// todo: move to core layer
	Card struct {
		ID       string
		UserID   string
		Language language.Tag

		//todo: add another one field like "StartedLearningAt" to store the date when the user started learning the word,
		//	so we can:
		// 		1. calculate the interval between the current date and the "StartedLearningAt".
		//		2. sort the words by the "StartedLearningAt" field for the repeat stag,
		//			later the date is the earlier the word will be repeated.

		//todo: add the "CreatedAt" field to store the date when the card was created.
		//	so we can:
		//		1. use this field to sort the words by the "CreatedAt" field for the learning stage,
		//			later the date is the earlier the word will be learned.

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

	/*	temporary unused, but may be useful in the future
		User struct {
				id      string
				created time.Time
			}
	*/

	UserSession struct {
		ID      string     `json:"id"`
		UserID  string     `json:"userID"`
		Started time.Time  `json:"started"`
		Closed  *time.Time `json:"closed"`
	}
)

func NewUserSession(userID string) UserSession {
	return UserSession{
		ID:      uuid.NewString(),
		UserID:  userID,
		Started: time.Now(),
	}
}

// todo: receive an user time zone. time.Now must be replaced with time.Now().In(userTimeZone)
func (c *Card) NeedToRepeat() bool {
	return !c.NextDueDate.IsZero() && time.Now().After(c.NextDueDate)
}

func (c *Card) NeedToLearn() bool {
	return c.NextDueDate.IsZero()
}

func (c *Card) AddAnswer(correct bool) {
	if correct {
		c.ConsecutiveCorrectAnswersNumber++
	} else {
		c.ConsecutiveCorrectAnswersNumber = 0
	}
}

func (c *Card) GetConsecutiveCorrectAnswersNumber() uint32 {
	return c.ConsecutiveCorrectAnswersNumber
}

func (s *UserSession) Duration() (time.Duration, error) {
	if s.IsClosed() {
		return s.Closed.Sub(s.Started), nil
	}

	return -1, errors.New("session is alive")
}

func (s *UserSession) IsClosed() bool {
	return s.Closed != nil
}

func (s *UserSession) Close() {
	if s.IsClosed() {
		return
	}
	s.Closed = lo.ToPtr(time.Now())
}

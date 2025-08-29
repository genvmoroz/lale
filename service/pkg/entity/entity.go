package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
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

	WordInformation struct { // todo: rename to Word
		Word            string            `yaml:"Word,omitempty"`
		Translation     *Translation      `yaml:"Translation,omitempty"`
		Origin          string            `yaml:"Origin,omitempty"`
		Phonetics       []Phonetic        `yaml:"Phonetics,omitempty"`
		Meanings        []Meaning         `yaml:"Meanings,omitempty"`
		AudioByLanguage map[string][]byte `yaml:"AudioByLanguage,omitempty"`

		// todo: add details field with the following fields:
		// Origin      string       `yaml:"Origin,omitempty"`
		// Phonetics   []Phonetic   `yaml:"Phonetics,omitempty"`
		// Meanings    []Meaning    `yaml:"Meanings,omitempty"`
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
		Started: time.Now().UTC(),
	}
}

func (c *Card) NeedToRepeat() bool {
	return !c.NextDueDate.IsZero() && time.Now().UTC().After(c.NextDueDate.UTC())
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
	s.Closed = lo.ToPtr(time.Now().UTC())
}

package entity

import (
	"errors"
	"github.com/samber/lo"
	"time"

	"github.com/google/uuid"
	"golang.org/x/text/language"
)

type (
	Card struct {
		ID       string
		UserID   string
		Language language.Tag

		WordInformationList []WordInformation `yaml:"WordInformationList,omitempty"`

		CorrectAnswers uint32
		NextDueDate    time.Time
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

	User struct {
		id      string    // nolint: unused
		created time.Time // nolint: unused
	}

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

func (c *Card) NeedToReview() bool {
	return time.Now().UTC().After(c.NextDueDate.UTC())
}

func (c *Card) GetAnswer(correct bool) uint32 {
	if correct {
		c.CorrectAnswers += 1
	} else {
		c.CorrectAnswers = 0
	}

	return c.CorrectAnswers
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
	s.Closed = lo.ToPtr(time.Now().UTC())
}

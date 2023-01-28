package lang

import "strings"

type Language string

const (
	English   Language = "en"
	Ukrainian Language = "uk"
)

func (l Language) EqualString(val string) bool {
	return l.Equal(Language(val))
}

func (l Language) Equal(val Language) bool {
	return strings.EqualFold(l.String(), val.String())
}

func (l Language) String() string {
	return string(l)
}

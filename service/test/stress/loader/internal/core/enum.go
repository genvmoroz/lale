package core

//go:generate stringer -output=enum_strings.go -type=Action

type Action int8

const (
	ActionInvalid Action = iota
	ActionCreateCard
	ActionInspectCard
	ActionPromptCard
	ActionGetAllCards
	ActionUpdateCard
	ActionUpdateCardPerformance
	ActionGetCardsToLearn
	ActionGetCardsToRepeat
	ActionGetSentences
	ActionGenerateStory
	ActionDeleteCard
)

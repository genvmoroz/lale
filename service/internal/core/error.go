package core

import "fmt"

type (
	RequestValidationError struct {
		baseErr error
	}

	CardNotFoundError struct {
		id   string
		word string
	}

	CardAlreadyExistsError struct {
		word string
	}
)

func NewRequestValidationError(err error) RequestValidationError {
	return RequestValidationError{baseErr: err}
}

func NewCardNotFoundError() CardNotFoundError {
	return CardNotFoundError{}
}

func NewCardAlreadyExistsError(word string) CardAlreadyExistsError {
	return CardAlreadyExistsError{word: word}
}

func (e RequestValidationError) Error() string {
	return fmt.Sprintf("validation failed: %s", e.baseErr.Error())
}

func (e CardNotFoundError) Error() string {
	switch {
	case e.id != "":
		return fmt.Sprintf("card with id %s not found", e.id)
	case e.word != "":
		return fmt.Sprintf("card with word %s not found", e.word)
	default:
		return "card not found"
	}
}

func (e CardAlreadyExistsError) Error() string {
	return fmt.Sprintf("card with word %s already exists", e.word)
}

func (e CardNotFoundError) WithID(id string) CardNotFoundError {
	e.id = id
	return e
}

func (e CardNotFoundError) WithWord(word string) CardNotFoundError {
	e.word = word
	return e
}

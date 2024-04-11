package core

import (
	"errors"
	"fmt"
)

var errValidation = fmt.Errorf("validation failed")

func NewValidationError() error {
	return errValidation
}

func IsValidationError(err error) bool {
	return errors.Is(err, errValidation)
}

var errNotFound = fmt.Errorf("not found")

func NewNotFoundError() error {
	return errNotFound
}

func IsNotFoundError(err error) bool {
	return errors.Is(err, errNotFound)
}

var errAlreadyExists = fmt.Errorf("already exists")

func NewAlreadyExistsError() error {
	return errAlreadyExists
}

func IsAlreadyExistsError(err error) bool {
	return errors.Is(err, errAlreadyExists)
}

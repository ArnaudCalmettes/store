package store

import (
	"errors"
)

type ErrorMap struct {
	ErrNotFound error
}

var (
	ErrNotFound = errors.New("not found")

	DefaultErrorMap = ErrorMap{
		ErrNotFound: ErrNotFound,
	}
)

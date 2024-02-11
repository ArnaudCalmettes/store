package store

import (
	"errors"
)

type ErrorMap struct {
	ErrNotFound error
	ErrEmptyKey error
}

var (
	ErrNotFound = errors.New("not found")
	ErrEmptyKey = errors.New("empty key")
)

func (e *ErrorMap) InitDefaultErrors() {
	if e.ErrNotFound == nil {
		e.ErrNotFound = ErrNotFound
	}
	if e.ErrEmptyKey == nil {
		e.ErrEmptyKey = ErrEmptyKey
	}
}

package store

import (
	"errors"
)

type ErrorMap struct {
	ErrNotFound      error
	ErrEmptyKey      error
	ErrSerialize     error
	ErrDeserialize   error
	ErrInvalidOption error
	ErrInvalidFilter error
}

var (
	ErrNotFound      = errors.New("not found")
	ErrEmptyKey      = errors.New("empty key")
	ErrSerialize     = errors.New("couldn't serialize object")
	ErrDeserialize   = errors.New("couldn't deserialize data")
	ErrInvalidOption = errors.New("invalid option")
	ErrInvalidFilter = errors.New("invalid filter")
)

func (e *ErrorMap) InitDefaultErrors() {
	if e.ErrNotFound == nil {
		e.ErrNotFound = ErrNotFound
	}
	if e.ErrEmptyKey == nil {
		e.ErrEmptyKey = ErrEmptyKey
	}
	if e.ErrSerialize == nil {
		e.ErrSerialize = ErrSerialize
	}
	if e.ErrDeserialize == nil {
		e.ErrDeserialize = ErrDeserialize
	}
	if e.ErrInvalidOption == nil {
		e.ErrInvalidOption = ErrInvalidOption
	}
	if e.ErrInvalidFilter == nil {
		e.ErrInvalidFilter = ErrInvalidFilter
	}
}

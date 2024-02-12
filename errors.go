package store

import (
	"errors"
)

type ErrorMap struct {
	ErrNotFound    error
	ErrEmptyKey    error
	ErrSerialize   error
	ErrDeserialize error
}

var (
	ErrNotFound    = errors.New("not found")
	ErrEmptyKey    = errors.New("empty key")
	ErrSerialize   = errors.New("couldn't serialize object")
	ErrDeserialize = errors.New("couldn't deserialize data")
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
}

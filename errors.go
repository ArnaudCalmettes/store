// Copyright (c) 2024 nohar
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

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

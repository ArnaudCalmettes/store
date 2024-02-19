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

package inspect

import (
	"sync"
	"testing"
	"time"

	"github.com/ArnaudCalmettes/store"
	. "github.com/ArnaudCalmettes/store/test/helpers"
)

func TestNewPredicateErrors(t *testing.T) {
	t.Parallel()
	t.Run("invalid filter", func(t *testing.T) {
		type Entry struct{}
		f, err := NewPredicate[Entry](&store.FilterSpec{})
		Expect(t,
			Equal(nil, f),
			IsError(errInvalidFilter, err),
		)
	})
	t.Run("not a struct", func(t *testing.T) {
		f, err := NewPredicate[string](store.Where("ID", "!=", 0))
		Expect(t,
			Equal(nil, f),
			IsError(errNotAStruct, err),
		)
		f, err = NewPredicate[string](store.All(store.Where("ID", "!=", int16(0))))
		Expect(t,
			Equal(nil, f),
			IsError(errNotAStruct, err),
		)
		f, err = NewPredicate[string](store.Any(store.Where("ID", "!=", int32(0))))
		Expect(t,
			Equal(nil, f),
			IsError(errNotAStruct, err),
		)
	})
	t.Run("no such field", func(t *testing.T) {
		type Entry struct{}
		f, err := NewPredicate[Entry](store.Where("ID", "!=", time.Now()))
		Expect(t,
			Equal(nil, f),
			IsError(errNoSuchField, err),
		)
		f, err = NewPredicate[Entry](store.All(store.Where("ID", "!=", false)))
		Expect(t,
			Equal(nil, f),
			IsError(errNoSuchField, err),
		)
		f, err = NewPredicate[Entry](store.Any(store.Where("ID", "!=", float32(0))))
		Expect(t,
			Equal(nil, f),
			IsError(errNoSuchField, err),
		)
	})
	t.Run("field type mismatch", func(t *testing.T) {
		type Entry struct {
			ID string
		}
		f, err := NewPredicate[Entry](store.Where("ID", "!=", int8(40)))
		Expect(t,
			Equal(nil, f),
			IsError(errTypeMismatch, err),
		)
		f, err = NewPredicate[Entry](store.All(store.Where("ID", "!=", uint(40))))
		Expect(t,
			Equal(nil, f),
			IsError(errTypeMismatch, err),
		)
		f, err = NewPredicate[Entry](store.Any(store.Where("ID", "!=", uint32(40))))
		Expect(t,
			Equal(nil, f),
			IsError(errTypeMismatch, err),
		)
	})
	t.Run("invalid operator", func(t *testing.T) {
		type Entry struct {
			Active bool
		}
		f, err := NewPredicate[Entry](store.Where("Active", ">", false))
		Expect(t,
			Equal(nil, f),
			IsError(errInvalidOperator, err),
		)
	})
	t.Run("unsupported type", func(t *testing.T) {
		type Entry struct {
			MaybeName *string
		}
		f, err := NewPredicate[Entry](store.Where("MaybeName", "=", sync.Mutex{}))
		Expect(t,
			Equal(nil, f),
			IsError(errTypeNotSupported, err),
		)
	})
}

func TestPredicates(t *testing.T) {
	t.Parallel()
	t.Run("ordered", func(t *testing.T) {
		type Entry struct {
			Name string
		}
		obj := &Entry{Name: "name"}
		testCases := []struct {
			Op     string
			Value  string
			Error  error
			Expect bool
		}{
			{Op: ">", Value: "address", Expect: true},
			{Op: ">", Value: "name", Expect: false},
			{Op: ">", Value: "zimbabwe", Expect: false},
			{Op: ">=", Value: "address", Expect: true},
			{Op: ">=", Value: "name", Expect: true},
			{Op: ">=", Value: "zimbabwe", Expect: false},
			{Op: "=", Value: "address", Expect: false},
			{Op: "=", Value: "name", Expect: true},
			{Op: "!=", Value: "address", Expect: true},
			{Op: "!=", Value: "name", Expect: false},
			{Op: "<=", Value: "address", Expect: false},
			{Op: "<=", Value: "name", Expect: true},
			{Op: "<=", Value: "zimbabwe", Expect: true},
			{Op: "<", Value: "address", Expect: false},
			{Op: "<", Value: "name", Expect: false},
			{Op: "<", Value: "zimbabwe", Expect: true},
			{Op: "LIKE", Value: "%name", Error: errInvalidOperator},
		}

		t.Parallel()
		for _, tc := range testCases {
			t.Run(tc.Op, func(t *testing.T) {
				f, err := NewPredicate[Entry](store.Where("Name", tc.Op, tc.Value))
				if tc.Error != nil {
					Expect(t,
						IsError(tc.Error, err),
					)
				} else {
					Expect(t,
						Equal(tc.Expect, f(obj)),
					)
				}
			})
		}
	})
	t.Run("pointer", func(t *testing.T) {
		type Entry struct {
			Ptr *string
			Int int
		}
		obj := &Entry{}
		f, err := NewPredicate[Entry](store.Where("Ptr", "=", nil))
		Expect(t,
			NoError(err),
			Equal(true, f(obj)),
		)
		f, err = NewPredicate[Entry](store.Where("Ptr", "!=", nil))
		Expect(t,
			NoError(err),
			Equal(false, f(obj)),
		)
		_, err = NewPredicate[int](store.Where("Ptr", "!=", nil))
		Expect(t,
			IsError(errNotAStruct, err),
		)
		_, err = NewPredicate[Entry](store.Where("Pointer", "=", nil))
		Expect(t,
			IsError(errNoSuchField, err),
		)
		_, err = NewPredicate[Entry](store.Where("Int", "!=", nil))
		Expect(t,
			IsError(errTypeMismatch, err),
		)
		_, err = NewPredicate[Entry](store.Where("Ptr", ">", nil))
		Expect(t,
			IsError(errInvalidOperator, err),
		)

	})
	t.Run("comparable", func(t *testing.T) {
		type Entry struct {
			Active bool
		}
		obj := &Entry{Active: true}
		testCases := []struct {
			Op     string
			Value  bool
			Error  error
			Expect bool
		}{
			{Op: "=", Value: false, Expect: false},
			{Op: "=", Value: true, Expect: true},
			{Op: "!=", Value: false, Expect: true},
			{Op: "!=", Value: true, Expect: false},
			{Op: "LIKE", Value: true, Error: errInvalidOperator},
		}

		t.Parallel()
		for _, tc := range testCases {
			t.Run(tc.Op, func(t *testing.T) {
				f, err := NewPredicate[Entry](store.Where("Active", tc.Op, tc.Value))
				if tc.Error != nil {
					Expect(t,
						IsError(tc.Error, err),
					)
				} else {
					Require(t,
						NoError(err),
					)
					Expect(t,
						Equal(tc.Expect, f(obj)),
					)
				}
			})
		}
	})
	t.Run("time", func(t *testing.T) {
		type Entry struct {
			CreatedAt time.Time
		}
		now := time.Now()
		tomorrow := now.Add(24 * time.Hour)
		yesterday := now.Add(-24 * time.Hour)
		obj := &Entry{CreatedAt: now}
		testCases := []struct {
			Op     string
			Value  time.Time
			Error  error
			Expect bool
		}{
			{Op: ">", Value: tomorrow, Expect: false},
			{Op: ">", Value: now, Expect: false},
			{Op: ">", Value: yesterday, Expect: true},
			{Op: ">=", Value: tomorrow, Expect: false},
			{Op: ">=", Value: now, Expect: true},
			{Op: ">=", Value: yesterday, Expect: true},
			{Op: "=", Value: tomorrow, Expect: false},
			{Op: "=", Value: now, Expect: true},
			{Op: "=", Value: yesterday, Expect: false},
			{Op: "!=", Value: tomorrow, Expect: true},
			{Op: "!=", Value: now, Expect: false},
			{Op: "!=", Value: yesterday, Expect: true},
			{Op: "<=", Value: tomorrow, Expect: true},
			{Op: "<=", Value: now, Expect: true},
			{Op: "<=", Value: yesterday, Expect: false},
			{Op: "<", Value: tomorrow, Expect: true},
			{Op: "<", Value: now, Expect: false},
			{Op: "<", Value: yesterday, Expect: false},
			{Op: "===", Value: now, Error: errInvalidOperator},
		}
		t.Parallel()
		for _, tc := range testCases {
			t.Run(tc.Op, func(t *testing.T) {
				f, err := NewPredicate[Entry](store.Where("CreatedAt", tc.Op, tc.Value))
				if tc.Error != nil {
					Expect(t,
						IsError(tc.Error, err),
					)
				} else {
					Require(t,
						NoError(err),
					)
					Expect(t,
						Equal(tc.Expect, f(obj)),
					)
				}
			})
		}
	})
}

func TestCompoundPredicate(t *testing.T) {
	type Entry struct {
		Length uint64
		Offset int64
		Weight float64
	}
	t.Run("all", func(t *testing.T) {
		f, err := NewPredicate[Entry](store.All(
			store.Where("Length", ">=", uint64(42)),
			store.Where("Offset", "=", int64(0)),
			store.Where("Weight", ">", float64(13.37)),
		))
		Require(t,
			NoError(err),
		)

		Expect(t,
			Equal(false, f(&Entry{})),
			Equal(true, f(&Entry{Length: 50, Weight: 38})),
		)
	})
	t.Run("any", func(t *testing.T) {
		f, err := NewPredicate[Entry](store.Any(
			store.Where("Length", ">=", uint64(42)),
			store.Where("Offset", "=", int64(0)),
			store.Where("Weight", ">", float64(13.37)),
		))
		Require(t,
			NoError(err),
		)

		Expect(t,
			Equal(true, f(&Entry{})),
			Equal(false, f(&Entry{Offset: 3})),
		)
	})
}

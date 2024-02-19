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

package test

import (
	"errors"
	"testing"

	//lint:ignore ST1001 common definitions
	. "github.com/ArnaudCalmettes/store"
	//lint:ignore ST1001 test vocabulary
	. "github.com/ArnaudCalmettes/store/test/helpers"
	"github.com/google/uuid"
)

type Entry struct {
	Float  float64
	Int    int
	Bool   bool
	String string
}

type baseStoreConstructor = func(*testing.T) BaseKeyValueStore[Entry]

func TestBaseKeyValueStore(t *testing.T, newStore baseStoreConstructor) {
	type TestFunc = func(*testing.T, baseStoreConstructor)
	run := func(t *testing.T, name string, testFunc TestFunc) {
		t.Run(name, func(t *testing.T) {
			testFunc(t, newStore)
		})
	}
	t.Parallel()
	run(t, "GetSetOne", testBaseKeyValueStoreGetSetOne)
	run(t, "GetSetMany", testBaseKeyValueStoreGetSetMany)
	run(t, "GetAll", testBaseKeyValueStoreGetAll)
	run(t, "UpdateOne", testBaseKeyValueStoreUpdateOne)
	run(t, "UpdateMany", testBaseKeyValueStoreUpdateMany)
	run(t, "Delete", testBaseKeyValueStoreDelete)
}

func testBaseKeyValueStoreGetSetOne(t *testing.T, newStore baseStoreConstructor) {
	store := newStore(t)
	ctx, cancel := NewTestContext()
	defer cancel()
	fixture := &Entry{
		Float:  43.5,
		Int:    1337,
		Bool:   true,
		String: "string",
	}

	t.Run("get non existent", func(t *testing.T) {
		result, err := store.GetOne(ctx, "does not exist")
		Expect(t,
			IsNilPointer(result),
			IsError(ErrNotFound, err),
		)
	})
	t.Run("get empty key", func(t *testing.T) {
		result, err := store.GetOne(ctx, "")
		Expect(t,
			IsNilPointer(result),
			IsError(ErrEmptyKey, err),
		)
	})
	t.Run("set empty key", func(t *testing.T) {
		err := store.SetOne(ctx, "", &Entry{})
		Expect(t,
			IsError(ErrEmptyKey, err),
		)
	})
	t.Run("nominal", func(t *testing.T) {
		key := uuid.NewString()
		err := store.SetOne(ctx, key, fixture)
		Require(t,
			NoErrorf(err, "setting up fixture"),
		)
		result, err := store.GetOne(ctx, key)
		Expect(t,
			Equal(fixture, result),
			NoError(err),
		)
	})
	t.Run("overwrite", func(t *testing.T) {
		key := uuid.NewString()
		err := store.SetOne(ctx, key, fixture)
		Require(t,
			NoError(err),
		)

		updatedValue := *fixture
		updatedValue.String = "updated"
		err = store.SetOne(ctx, key, &updatedValue)
		Require(t,
			NoError(err),
		)

		check, err := store.GetOne(ctx, key)
		Expect(t,
			NoError(err),
			Equalf(&updatedValue, check, "value should have been updated"),
		)
	})
}

func testBaseKeyValueStoreGetSetMany(t *testing.T, newStore baseStoreConstructor) {
	store := newStore(t)
	ctx, cancel := NewTestContext()
	defer cancel()

	t.Run("get many empty", func(t *testing.T) {
		result, err := store.GetMany(ctx, []string{})
		Expect(t,
			NoError(err),
			Equal(map[string]*Entry{}, result),
		)
	})
	t.Run("set many empty map", func(t *testing.T) {
		err := store.SetMany(ctx, map[string]*Entry{})
		Expect(t,
			NoError(err),
		)
	})
	t.Run("set many empty key", func(t *testing.T) {
		err := store.SetMany(ctx, map[string]*Entry{
			"": {String: "value"},
		})
		Require(t,
			NoError(err),
		)
		all, err := store.GetAll(ctx)
		Expect(t,
			NoError(err),
			Equal(map[string]*Entry{}, all),
		)
	})
	t.Run("nominal", func(t *testing.T) {
		err := store.SetMany(ctx, map[string]*Entry{
			"one":   {String: "one"},
			"two":   {String: "two"},
			"three": {String: "three"},
		})
		Require(t,
			NoError(err),
		)

		result, err := store.GetMany(ctx, []string{"one", "two", "three", "four"})
		Expect(t,
			NoError(err),
			Equal(
				map[string]*Entry{
					"one":   {String: "one"},
					"two":   {String: "two"},
					"three": {String: "three"},
				},
				result,
			),
		)
	})
	t.Run("overwrite", func(t *testing.T) {
		err := store.SetMany(ctx, map[string]*Entry{
			"one": {String: "ONE"},
			"two": {String: "TWO"},
		})
		Require(t,
			NoError(err),
		)

		result, err := store.GetMany(ctx, []string{"one", "two", "three", "four"})
		Require(t,
			NoError(err),
			Equal(
				map[string]*Entry{
					"one":   {String: "ONE"},
					"two":   {String: "TWO"},
					"three": {String: "three"},
				},
				result,
			),
		)
	})
}

func testBaseKeyValueStoreGetAll(t *testing.T, newStore baseStoreConstructor) {
	store := newStore(t)
	ctx, cancel := NewTestContext()
	defer cancel()

	t.Run("empty", func(t *testing.T) {
		all, err := store.GetAll(ctx)
		Expect(t,
			NoError(err),
			Equal(map[string]*Entry{}, all),
		)
	})
	t.Run("nominal", func(t *testing.T) {
		store.SetMany(ctx, map[string]*Entry{
			"one": {String: "one"},
			"two": {String: "two"},
		})
		all, err := store.GetAll(ctx)
		Expect(t,
			NoError(err),
			Equal(
				map[string]*Entry{
					"one": {String: "one"},
					"two": {String: "two"},
				},
				all,
			),
		)
	})
}

func testBaseKeyValueStoreUpdateOne(t *testing.T, newStore baseStoreConstructor) {
	store := newStore(t)
	ctx, cancel := NewTestContext()
	defer cancel()

	t.Run("empty key", func(t *testing.T) {
		err := store.UpdateOne(ctx, "", nil)
		Expect(t,
			IsError(ErrEmptyKey, err),
		)
	})
	t.Run("does not exist", func(t *testing.T) {
		errCalled := errors.New("called")
		updateFunc := func(_ string, value *Entry) (*Entry, error) {
			Require(t,
				IsNilPointerf(value, "non existent values should be passed as a nil pointer"),
			)
			return nil, errCalled
		}
		err := store.UpdateOne(ctx, uuid.NewString(), updateFunc)
		Expect(t,
			IsError(errCalled, err),
		)
	})
	t.Run("nominal", func(t *testing.T) {
		id := uuid.NewString()
		err := store.SetOne(ctx, id, &Entry{String: "initial value"})
		Require(t,
			NoError(err),
		)

		updateFunc := func(key string, value *Entry) (*Entry, error) {
			Require(t,
				Equal(id, key),
				Equal(&Entry{String: "initial value"}, value),
			)
			value.String = "updated value"
			return value, nil
		}
		err = store.UpdateOne(ctx, id, updateFunc)
		Require(t,
			NoError(err),
		)

		value, err := store.GetOne(ctx, id)
		Expect(t,
			NoError(err),
			Equal(&Entry{String: "updated value"}, value),
		)
	})
	t.Run("callback returns nil", func(t *testing.T) {
		id := uuid.NewString()
		err := store.SetOne(ctx, id, &Entry{String: "initial value"})
		Require(t,
			NoError(err),
		)
		var called bool
		updateFunc := func(_ string, value *Entry) (*Entry, error) {
			called = true
			value.String = "new value"
			return nil, nil
		}
		err = store.UpdateOne(ctx, id, updateFunc)
		Expect(t,
			NoError(err),
			Equal(true, called),
		)

		value, err := store.GetOne(ctx, id)
		Expect(t,
			NoError(err),
			Equal(&Entry{String: "initial value"}, value),
		)
	})
}

func testBaseKeyValueStoreUpdateMany(t *testing.T, newStore baseStoreConstructor) {
	store := newStore(t)
	ctx, cancel := NewTestContext()
	defer cancel()

	t.Run("no keys", func(t *testing.T) {
		err := store.UpdateMany(ctx, []string{}, nil)
		Expect(t,
			NoError(err),
		)
	})
	t.Run("empty keys", func(t *testing.T) {
		err := store.UpdateMany(ctx, []string{""}, nil)
		Expect(t,
			NoError(err),
		)
	})
	t.Run("key does not exist", func(t *testing.T) {
		id := uuid.NewString()
		var called bool
		updateFunc := func(key string, value *Entry) (*Entry, error) {
			called = true
			Require(t,
				Equal(id, key),
				IsNilPointer(value),
			)
			return &Entry{String: "inserted value"}, nil
		}
		err := store.UpdateMany(ctx, []string{id}, updateFunc)
		Expect(t,
			NoError(err),
			Equal(true, called),
		)

		value, err := store.GetOne(ctx, id)
		Expect(t,
			NoError(err),
			Equal(&Entry{String: "inserted value"}, value),
		)
	})
	t.Run("nominal", func(t *testing.T) {
		id := uuid.NewString()
		err := store.SetOne(ctx, id, &Entry{String: "initial value"})
		Require(t,
			NoError(err),
		)

		updateFunc := func(key string, value *Entry) (*Entry, error) {
			Require(t,
				Equal(id, key),
				Equal(&Entry{String: "initial value"}, value),
			)
			value.String = "updated value"
			return value, nil
		}
		err = store.UpdateMany(ctx, []string{id}, updateFunc)
		Require(t,
			NoError(err),
		)

		value, err := store.GetOne(ctx, id)
		Expect(t,
			NoError(err),
			Equal(&Entry{String: "updated value"}, value),
		)
	})
	t.Run("callback returns error", func(t *testing.T) {
		id := uuid.NewString()
		err := store.SetOne(ctx, id, &Entry{String: "initial value"})
		Require(t,
			NoError(err),
		)

		errUpdate := errors.New("update error")
		updateFunc := func(_ string, value *Entry) (*Entry, error) {
			value.String = "updated value"
			return value, errUpdate
		}
		err = store.UpdateMany(ctx, []string{id}, updateFunc)
		Require(t,
			IsError(errUpdate, err),
		)

		value, err := store.GetOne(ctx, id)
		Expect(t,
			NoError(err),
			Equal(&Entry{String: "initial value"}, value),
		)
	})
	t.Run("callback returns nil", func(t *testing.T) {
		id := uuid.NewString()
		err := store.SetOne(ctx, id, &Entry{String: "initial value"})
		Require(t,
			NoError(err),
		)

		updateFunc := func(_ string, value *Entry) (*Entry, error) {
			value.String = "updated value"
			return nil, nil
		}
		err = store.UpdateMany(ctx, []string{id}, updateFunc)
		Require(t,
			NoError(err),
		)

		value, err := store.GetOne(ctx, id)
		Expect(t,
			NoError(err),
			Equal(&Entry{String: "initial value"}, value),
		)
	})
}

func testBaseKeyValueStoreDelete(t *testing.T, newStore baseStoreConstructor) {
	store := newStore(t)
	ctx, cancel := NewTestContext()
	defer cancel()

	t.Run("empty keys", func(t *testing.T) {
		err := store.Delete(ctx)
		Expect(t,
			NoError(err),
		)
	})
	t.Run("nominal", func(t *testing.T) {
		err := store.SetMany(ctx, map[string]*Entry{
			"one": {String: "one"},
			"two": {String: "two"},
		})
		Require(t,
			NoError(err),
		)

		err = store.Delete(ctx, "three", "two")
		Expect(t,
			NoError(err),
		)

		all, err := store.GetMany(ctx, []string{"one", "two", "three"})
		Expect(t,
			NoError(err),
			Equal(map[string]*Entry{"one": {String: "one"}}, all),
		)
	})
}

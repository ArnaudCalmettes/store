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

package serializer

import (
	"context"
	"errors"
	"testing"

	. "github.com/ArnaudCalmettes/store"
	"github.com/ArnaudCalmettes/store/memory"
	. "github.com/ArnaudCalmettes/store/test"
	. "github.com/ArnaudCalmettes/store/test/helpers"
)

func TestSerializerKeyValueStore(t *testing.T) {
	newStore := func(*testing.T) BaseKeyValueStore[Entry] {
		return NewKeyValue(
			NewJSON[Entry](),
			memory.NewKeyValueMap(),
		)
	}
	TestBaseKeyValueStore(t, newStore)
}

func TestSerializerKeyValueStoreLister(t *testing.T) {
	newStore := func(*testing.T) TestListerInterface[Person] {
		return NewKeyValue(
			NewJSON[Person](),
			memory.NewKeyValueMap(),
		)
	}
	TestLister(t, newStore)
}

func TestKeyValueStoreCustomErrors(t *testing.T) {
	errTest := errors.New("test")
	store := NewKeyValue(NewJSON[Entry](), memory.NewKeyValueMap())
	store.SetErrorMap(ErrorMap{
		ErrNotFound: errTest,
	})
	_, err := store.GetOne(context.Background(), "does not exist")
	Require(t,
		IsError(errTest, err),
	)
}

func TestSerializationErrors(t *testing.T) {
	ctx, cancel := NewTestContext()
	defer cancel()
	mem := memory.NewKeyValueMap()
	store := NewKeyValue(NewJSON[Entry](), mem)
	mem.SetOne(ctx, "malformed", "}")

	t.Run("GetAll", func(t *testing.T) {
		_, err := store.GetAll(ctx)
		Expect(t,
			IsError(ErrDeserialize, err),
		)
	})
	t.Run("GetOne", func(t *testing.T) {
		_, err := store.GetOne(ctx, "malformed")
		Expect(t,
			IsError(ErrDeserialize, err),
		)
	})
	t.Run("GetMany", func(t *testing.T) {
		_, err := store.GetMany(ctx, []string{"malformed"})
		Expect(t,
			IsError(ErrDeserialize, err),
		)
	})
	t.Run("SetOne", func(t *testing.T) {
		err := store.SetOne(ctx, "item", nil)
		Expect(t,
			IsError(ErrSerialize, err),
		)
	})
	t.Run("SetMany", func(t *testing.T) {
		err := store.SetMany(ctx, map[string]*Entry{"item": nil})
		Expect(t,
			IsError(ErrSerialize, err),
		)
	})
	t.Run("UpdateOne", func(t *testing.T) {
		err := store.UpdateOne(ctx, "malformed",
			func(_ string, e *Entry) (*Entry, error) {
				return e, nil
			},
		)
		Expect(t,
			IsError(ErrDeserialize, err),
		)
	})
	t.Run("UpdateMany", func(t *testing.T) {
		err := store.UpdateMany(ctx, []string{"malformed"},
			func(_ string, e *Entry) (*Entry, error) {
				return e, nil
			},
		)
		Expect(t,
			IsError(ErrDeserialize, err),
		)
	})
}

func TestKeyValueStoreReset(t *testing.T) {
	store := NewKeyValue(NewJSON[Entry](), memory.NewKeyValueMap())
	err := store.SetMany(context.Background(), map[string]*Entry{
		"one":   {String: "one"},
		"three": {String: "three"},
	})
	Require(t,
		NoError(err),
	)

	err = store.Reset(context.Background())
	Require(t,
		NoError(err),
	)

	all, err := store.GetAll(context.Background())
	Require(t,
		NoError(err),
		Equal(map[string]*Entry{}, all),
	)
}

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

package sql

import (
	"context"
	"errors"
	"testing"

	. "github.com/ArnaudCalmettes/store"
	. "github.com/ArnaudCalmettes/store/test/helpers"
	"github.com/uptrace/bun"
)

func TestSQLNewKeyValueStore(t *testing.T) {
	var db *bun.DB
	t.Run("not a struct", func(t *testing.T) {
		Expect(t,
			ShouldPanic(func() {
				NewKeyValue[int](db)
			}),
		)
	})
	t.Run("not a bun model", func(t *testing.T) {
		type Model struct {
			ID string
		}
		Expect(t,
			ShouldPanic(func() {
				NewKeyValue[Model](db)
			}),
		)
	})
	t.Run("not a string key", func(t *testing.T) {
		type Model struct {
			bun.BaseModel `bun:"table:models"`

			ID   int `bun:",pk"`
			Name string
		}
		Expect(t,
			ShouldPanic(func() {
				NewKeyValue[Model](db)
			}),
		)
	})
	t.Run("nominal", func(t *testing.T) {
		type Model struct {
			bun.BaseModel `bun:"table:models"`

			ID   string `bun:",pk"`
			Name string
		}
		Expect(t,
			DoesNotPanic(func() {
				NewKeyValue[Model](db)
			}),
		)
	})
}

type Item struct {
	bun.BaseModel `bun:"table:entries"`

	ID   string `bun:",pk"`
	Name string
	Age  int
}

func TestKeyValueStoreCustomErrors(t *testing.T) {
	ctx, cancel := NewTestContext()
	defer cancel()

	db := newSQLite(t)
	db.ResetModel(ctx, (*Item)(nil))

	errTest := errors.New("test")
	store := NewKeyValue[Item](db)
	store.SetErrorMap(ErrorMap{
		ErrNotFound: errTest,
	})
	_, err := store.GetOne(context.Background(), "does not exist")
	Require(t,
		IsError(errTest, err),
	)
}

func TestSQLKeyValueStoreReset(t *testing.T) {
	ctx, cancel := NewTestContext()
	defer cancel()

	db := newSQLite(t)
	db.ResetModel(ctx, (*Item)(nil))

	store := NewKeyValue[Item](db)

	err := store.SetMany(context.Background(), map[string]*Item{
		"one":   {Name: "one"},
		"three": {Name: "three"},
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
		Equal(map[string]*Item{}, all),
	)
}

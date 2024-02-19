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
	"testing"

	//lint:ignore ST1001 shared definitions
	. "github.com/ArnaudCalmettes/store"
	//lint:ignore ST1001 test vocabulary
	. "github.com/ArnaudCalmettes/store/test/helpers"
)

type TestListerInterface[T any] interface {
	BaseKeyValueStore[T]
	Lister[T]
}

type Person struct {
	ID       string
	Name     string
	Age      int
	Referent *string
}

type listerConstructor func(*testing.T) TestListerInterface[Person]

func TestLister(t *testing.T, newLister listerConstructor) {
	store := newLister(t)
	ctx, cancel := NewTestContext()
	defer cancel()
	err := store.SetMany(ctx, map[string]*Person{
		"001": {ID: "001", Name: "John Doe", Age: 42},
		"002": {ID: "002", Name: "Willard", Age: 13},
		"003": {ID: "003", Name: "Jane Smith", Age: 20},
	})
	Require(t,
		NoError(err),
	)
	t.Run("no filter", func(t *testing.T) {
		result, err := store.List(ctx)
		Expect(t,
			NoError(err),
			SliceHasLength(3, result),
		)
	})
	t.Run("invalid filter", func(t *testing.T) {
		_, err := store.List(ctx, Filter(Where("BankAccount", "!=", 42)))
		Expect(t,
			IsError(ErrInvalidFilter, err),
		)
	})
	t.Run("filter nominal", func(t *testing.T) {
		result, err := store.List(ctx, Filter(Where("Age", "<", 18)))
		Expect(t,
			NoError(err),
			Equal(
				[]*Person{
					{ID: "002", Name: "Willard", Age: 13},
				},
				result,
			),
		)
	})
	t.Run("multiple filters", func(t *testing.T) {
		result, err := store.List(ctx,
			Filter(Where("Age", ">", 10)),
			Filter(Where("Age", "<", 40)),
			Filter(Where("Age", "!=", 13)),
		)
		Expect(t,
			NoError(err),
			Equal(
				[]*Person{
					{ID: "003", Name: "Jane Smith", Age: 20},
				},
				result,
			),
		)

	})
	t.Run("duplicate orderby option", func(t *testing.T) {
		_, err := store.List(ctx, Order(By("Name")), Order(By("Age").Desc()))
		Expect(t,
			IsError(ErrInvalidOption, err),
		)
	})
	t.Run("order by invalid field", func(t *testing.T) {
		_, err := store.List(ctx, Order(By("Profession")))
		Expect(t,
			IsError(ErrInvalidOption, err),
		)
	})
	t.Run("nominal", func(t *testing.T) {
		result, err := store.List(ctx, Order(By("Age").Desc()))
		Expect(t,
			NoError(err),
			Equal(
				[]*Person{
					{ID: "001", Name: "John Doe", Age: 42},
					{ID: "003", Name: "Jane Smith", Age: 20},
					{ID: "002", Name: "Willard", Age: 13},
				},
				result,
			),
		)
	})
	t.Run("filter and order", func(t *testing.T) {
		result, err := store.List(ctx,
			Filter(Where("Age", ">", 18)),
			Order(By("Age")),
		)
		Expect(t,
			NoError(err),
			Equal(
				[]*Person{
					{ID: "003", Name: "Jane Smith", Age: 20},
					{ID: "001", Name: "John Doe", Age: 42},
				},
				result,
			),
		)
	})
	t.Run("paginate", func(t *testing.T) {
		result, err := store.List(ctx, Order(By("Age")), Limit(2), Offset(1))
		Expect(t,
			NoError(err),
			Equal(
				[]*Person{
					{ID: "003", Name: "Jane Smith", Age: 20},
					{ID: "001", Name: "John Doe", Age: 42},
				},
				result,
			),
		)

		result, err = store.List(ctx, Limit(100), Offset(50))
		Expect(t,
			NoError(err),
			IsEmptySlice(result),
		)
	})
}

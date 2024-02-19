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
	"testing"
	"time"

	. "github.com/ArnaudCalmettes/store/test/helpers"
)

func TestFieldSelector(t *testing.T) {
	type MyStruct struct {
		Foo int
		Bar float64
		Baz string
		Biz time.Time
	}
	t.Run("not a struct type", func(t *testing.T) {
		result, err := FieldSelector[string, int]("Length")
		Expect(t,
			IsError(errNotAStruct, err),
			Equal(nil, result),
		)
	})
	t.Run("field does not exist", func(t *testing.T) {
		result, err := FieldSelector[MyStruct, int]("DoesNotExist")
		Expect(t,
			IsError(errNoSuchField, err),
			Equal(nil, result),
		)
	})
	t.Run("field type does not match", func(t *testing.T) {
		result, err := FieldSelector[MyStruct, string]("Foo")
		Expect(t,
			IsError(errTypeMismatch, err),
			Equal(nil, result),
		)
	})
	t.Run("nominal", func(t *testing.T) {
		fooSelector, err := FieldSelector[MyStruct, int]("Foo")
		Require(t,
			NoError(err),
		)
		barSelector, err := FieldSelector[MyStruct, float64]("Bar")
		Require(t,
			NoError(err),
		)
		bazSelector, err := FieldSelector[MyStruct, string]("Baz")
		Require(t,
			NoError(err),
		)
		bizSelector, err := FieldSelector[MyStruct, time.Time]("Biz")
		Require(t,
			NoError(err),
		)

		now := time.Now()
		myObj := &MyStruct{
			Foo: 42,
			Bar: 13.37,
			Baz: "basinga",
			Biz: now,
		}

		Expect(t,
			Equal(42, fooSelector(myObj)),
			Equal(13.37, barSelector(myObj)),
			Equal("basinga", bazSelector(myObj)),
			Equal(now, bizSelector(myObj)),
		)
	})

	t.Run("nested", func(t *testing.T) {
		type Metadata struct {
			Namespace string
			Name      string
		}
		type MyStruct struct {
			Metadata
		}

		nameSelector, err := FieldSelector[MyStruct, string]("Name")
		Require(t,
			NoError(err),
		)

		obj := &MyStruct{
			Metadata: Metadata{
				Namespace: "namespace",
				Name:      "name",
			},
		}
		Expect(t,
			Equal("name", nameSelector(obj)),
		)
	})
}

func TestStringFieldSetter(t *testing.T) {
	type MyStruct struct {
		ID   string
		Name string
		Age  int
	}
	t.Run("not a struct type", func(t *testing.T) {
		result, err := StringFieldSetter[int]("ID")
		Expect(t,
			IsError(errNotAStruct, err),
			Equal(nil, result),
		)
	})
	t.Run("field does not exist", func(t *testing.T) {
		result, err := StringFieldSetter[MyStruct]("Profession")
		Expect(t,
			IsError(errNoSuchField, err),
			Equal(nil, result),
		)
	})
	t.Run("field type does not match", func(t *testing.T) {
		result, err := StringFieldSetter[MyStruct]("Age")
		Expect(t,
			IsError(errTypeMismatch, err),
			Equal(nil, result),
		)
	})
	t.Run("nominal", func(t *testing.T) {
		obj := &MyStruct{
			ID:   "id",
			Name: "name",
			Age:  42,
		}
		setID, err := StringFieldSetter[MyStruct]("ID")
		Require(t,
			NoError(err),
		)

		setID(obj, "012345")
		Expect(t,
			Equal(
				&MyStruct{
					ID:   "012345",
					Name: "name",
					Age:  42,
				},
				obj,
			),
		)
	})
}

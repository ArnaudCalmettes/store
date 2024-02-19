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

package proxy

import (
	"context"
	"errors"
	"testing"

	. "github.com/ArnaudCalmettes/store"
	"github.com/ArnaudCalmettes/store/memory"
	. "github.com/ArnaudCalmettes/store/test"
	. "github.com/ArnaudCalmettes/store/test/helpers"
)

type EntryProxy struct {
	MyFloat  float64
	MyInt    int
	MyBool   bool
	MyString string
}

func toProxy(e *Entry) *EntryProxy {
	if e == nil {
		return nil
	}
	return &EntryProxy{
		MyFloat:  e.Float,
		MyInt:    e.Int,
		MyBool:   e.Bool,
		MyString: e.String,
	}
}

func fromProxy(p *EntryProxy) *Entry {
	if p == nil {
		return nil
	}
	return &Entry{
		Float:  p.MyFloat,
		Int:    p.MyInt,
		Bool:   p.MyBool,
		String: p.MyString,
	}
}

func TestProxyKeyValueStore(t *testing.T) {
	newStore := func(*testing.T) BaseKeyValueStore[Entry] {
		return NewKeyValueStoreWithProxy[Entry, EntryProxy](
			memory.NewKeyValueStore[EntryProxy](),
			toProxy,
			fromProxy,
		)
	}
	TestBaseKeyValueStore(t, newStore)
}

func TestProxyLister(t *testing.T) {
	type PersonProxy struct {
		Person
		Address string
	}
	fromProxy := func(p *PersonProxy) *Person {
		return &p.Person
	}
	toProxy := func(p *Person) *PersonProxy {
		return &PersonProxy{Person: *p}
	}
	newStore := func(*testing.T) TestListerInterface[Person] {
		return NewKeyValueStoreWithProxy[Person, PersonProxy](
			memory.NewKeyValueStore[PersonProxy](),
			toProxy,
			fromProxy,
		)
	}
	TestLister(t, newStore)
}

func TestKeyValueStoreCustomErrors(t *testing.T) {
	errTest := errors.New("test")
	store := NewKeyValueStoreWithProxy[Entry, EntryProxy](
		memory.NewKeyValueStore[EntryProxy](),
		toProxy,
		fromProxy,
	)

	store.SetErrorMap(ErrorMap{
		ErrNotFound: errTest,
	})
	_, err := store.GetOne(context.Background(), "does not exist")
	Require(t,
		IsError(errTest, err),
	)
}

func TestKeyValueStoreReset(t *testing.T) {
	store := NewKeyValueStoreWithProxy[Entry, EntryProxy](
		memory.NewKeyValueStore[EntryProxy](),
		toProxy,
		fromProxy,
	)
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

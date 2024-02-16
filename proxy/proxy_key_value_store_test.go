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

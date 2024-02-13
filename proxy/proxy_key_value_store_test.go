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

type Proxy struct {
	MyFloat  float64
	MyInt    int
	MyBool   bool
	MyString string
}

func toProxy(e *Entry) *Proxy {
	if e == nil {
		return nil
	}
	return &Proxy{
		MyFloat:  e.Float,
		MyInt:    e.Int,
		MyBool:   e.Bool,
		MyString: e.String,
	}
}

func fromProxy(p *Proxy) *Entry {
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
		return NewKeyValueStoreWithProxy[Entry, Proxy](
			memory.NewKeyValueStore[Proxy](),
			toProxy,
			fromProxy,
		)
	}
	TestBaseKeyValueStore(t, newStore)
}

func TestKeyValueStoreCustomErrors(t *testing.T) {
	errTest := errors.New("test")
	store := NewKeyValueStoreWithProxy[Entry, Proxy](
		memory.NewKeyValueStore[Proxy](),
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
	store := NewKeyValueStoreWithProxy[Entry, Proxy](
		memory.NewKeyValueStore[Proxy](),
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

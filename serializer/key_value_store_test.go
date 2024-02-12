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
	newStore := func() BaseKeyValueStore[Entry] {
		return NewKeyValueStore(
			NewJSON[Entry](),
			memory.NewKeyValueMap(),
		)
	}
	TestBaseKeyValueStore(t, newStore)
}

func TestKeyValueStoreCustomErrors(t *testing.T) {
	errTest := errors.New("test")
	store := NewKeyValueStore(NewJSON[Entry](), memory.NewKeyValueMap())
	store.SetErrorMap(ErrorMap{
		ErrNotFound: errTest,
	})
	_, err := store.GetOne(context.Background(), "does not exist")
	Require(t,
		IsError(errTest, err),
	)
}

func TestKeyValueStoreReset(t *testing.T) {
	store := NewKeyValueStore(NewJSON[Entry](), memory.NewKeyValueMap())
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

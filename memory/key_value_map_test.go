package memory

import (
	"context"
	"errors"
	"testing"

	. "github.com/ArnaudCalmettes/store"
	. "github.com/ArnaudCalmettes/store/test"
	. "github.com/ArnaudCalmettes/store/test/helpers"
)

func TestMemoryKeyValueMap(t *testing.T) {
	newKeyValueMap := func() BaseKeyValueMap { return NewKeyValueMap() }
	TestBaseKeyValueMap(t, newKeyValueMap)
}

func TestKeyValueMapCustomErrors(t *testing.T) {
	errTest := errors.New("test")
	store := NewKeyValueMap()
	store.SetErrorMap(ErrorMap{
		ErrNotFound: errTest,
	})
	_, err := store.GetOne(context.Background(), "does not exist")
	Require(t,
		IsError(errTest, err),
	)
}

func TestKeyValueMapReset(t *testing.T) {
	store := NewKeyValueMap()
	err := store.SetMany(context.Background(), map[string]string{
		"one":   "two",
		"three": "four",
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
		Equal(map[string]string{}, all),
	)
}

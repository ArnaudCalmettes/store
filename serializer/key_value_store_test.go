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
		return NewKeyValueStore(
			NewJSON[Entry](),
			memory.NewKeyValueMap(),
		)
	}
	TestBaseKeyValueStore(t, newStore)
}

func TestSerializerKeyValueStoreLister(t *testing.T) {
	newStore := func(*testing.T) TestListerInterface[Person] {
		return NewKeyValueStore(
			NewJSON[Person](),
			memory.NewKeyValueMap(),
		)
	}
	TestLister(t, newStore)
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

func TestSerializationErrors(t *testing.T) {
	ctx, cancel := NewTestContext()
	defer cancel()
	mem := memory.NewKeyValueMap()
	store := NewKeyValueStore(NewJSON[Entry](), mem)
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

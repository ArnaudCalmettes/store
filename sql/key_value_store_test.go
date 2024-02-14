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
				NewKeyValueStore[int](db)
			}),
		)
	})
	t.Run("not a bun model", func(t *testing.T) {
		type Model struct {
			ID string
		}
		Expect(t,
			ShouldPanic(func() {
				NewKeyValueStore[Model](db)
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
				NewKeyValueStore[Model](db)
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
				NewKeyValueStore[Model](db)
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
	store := NewKeyValueStore[Item](db)
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

	store := NewKeyValueStore[Item](db)

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

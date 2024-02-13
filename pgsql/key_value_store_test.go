package pgsql

import (
	"context"
	"errors"
	"testing"

	. "github.com/ArnaudCalmettes/store"
	. "github.com/ArnaudCalmettes/store/test/helpers"
	"github.com/google/uuid"
	"github.com/rubenv/pgtest"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

type Item struct {
	bun.BaseModel `bun:"table:entries"`

	ID   string `bun:",pk"`
	Name string
	Age  int
}

func TestPgsqlKeyValueStore(t *testing.T) {
	ctx, cancel := NewTestContext()
	defer cancel()

	db := newTestDB(t)
	db.ResetModel(ctx, (*Item)(nil))

	store := NewKeyValueStore[Item](db)
	entry := Item{
		ID:   uuid.NewString(),
		Name: "entry",
		Age:  42,
	}

	err := store.SetOne(ctx, entry.ID, &entry)
	Expect(t,
		NoError(err),
	)
	result, err := store.GetOne(ctx, entry.ID)
	Expect(t,
		NoError(err),
		Equal(&entry, result),
	)
	result, err = store.GetOne(ctx, "does not exist")
	Expect(t,
		IsError(ErrNotFound, err),
		IsNilPointer(result),
	)
}

func TestNewKeyValueStore(t *testing.T) {
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

func TestKeyValueStoreCustomErrors(t *testing.T) {
	ctx, cancel := NewTestContext()
	defer cancel()

	db := newTestDB(t)
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

func TestKeyValueStoreReset(t *testing.T) {
	ctx, cancel := NewTestContext()
	defer cancel()

	db := newTestDB(t)
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

func newTestDB(t *testing.T) *bun.DB {
	t.Helper()
	pg, err := pgtest.Start()
	Require(t, NoError(err))
	t.Cleanup(func() { pg.Stop() })
	return bun.NewDB(pg.DB, pgdialect.New())
}

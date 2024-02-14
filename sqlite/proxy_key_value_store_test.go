package sqlite

import (
	"context"
	"testing"

	. "github.com/ArnaudCalmettes/store"
	. "github.com/ArnaudCalmettes/store/test"
	. "github.com/ArnaudCalmettes/store/test/helpers"
	"github.com/uptrace/bun"
)

func TestProxySQLiteKeyValueStore(t *testing.T) {
	type Proxy struct {
		bun.BaseModel `bun:"table:entries,alias:e"`

		ID string `bun:",pk"`
		Entry
	}

	toProxy := func(e *Entry) *Proxy {
		if e == nil {
			return nil
		}
		return &Proxy{Entry: *e}
	}

	fromProxy := func(p *Proxy) *Entry {
		if p == nil {
			return nil
		}
		return &p.Entry
	}

	newStore := func(t *testing.T) BaseKeyValueStore[Entry] {
		db := newTestDB(t)
		err := db.ResetModel(context.Background(), (*Proxy)(nil))
		Require(t, NoError(err))
		return NewKeyValueStoreWithProxy[Entry, Proxy](db, toProxy, fromProxy)
	}
	TestBaseKeyValueStore(t, newStore)
}

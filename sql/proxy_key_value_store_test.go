package sql

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"testing"

	. "github.com/ArnaudCalmettes/store"
	. "github.com/ArnaudCalmettes/store/test"
	. "github.com/ArnaudCalmettes/store/test/helpers"
	"github.com/rubenv/pgtest"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

type Proxy struct {
	bun.BaseModel `bun:"table:entries,alias:e"`

	ID string `bun:",pk"`
	Entry
}

func toProxy(e *Entry) *Proxy {
	if e == nil {
		return nil
	}
	return &Proxy{Entry: *e}
}

func fromProxy(p *Proxy) *Entry {
	if p == nil {
		return nil
	}
	return &p.Entry
}

func TestSQLiteKeyValueStore(t *testing.T) {
	newStore := func(t *testing.T) BaseKeyValueStore[Entry] {
		db := newSQLite(t)
		err := db.ResetModel(context.Background(), (*Proxy)(nil))
		Require(t, NoError(err))
		return NewKeyValueStoreWithProxy[Entry, Proxy](db, toProxy, fromProxy)
	}
	TestBaseKeyValueStore(t, newStore)
}

func TestPGKeyValueStore(t *testing.T) {
	pg, err := pgtest.Start()
	Require(t,
		NoError(err),
	)
	t.Cleanup(func() { pg.Stop() })

	newStore := func(t *testing.T) BaseKeyValueStore[Entry] {
		db := newPostgres(t, pg)
		err := db.ResetModel(context.Background(), (*Proxy)(nil))
		Require(t, NoError(err))
		return NewKeyValueStoreWithProxy[Entry, Proxy](db, toProxy, fromProxy)
	}
	TestBaseKeyValueStore(t, newStore)
}

func newSQLite(t *testing.T) *bun.DB {
	t.Helper()
	sqldb, err := sql.Open(sqliteshim.ShimName, "file::memory:?cache=shared")
	Require(t, NoError(err))
	return bun.NewDB(sqldb, sqlitedialect.New())
}

func newPostgres(t *testing.T, pg *pgtest.PG) *bun.DB {
	t.Helper()
	suffix := make([]byte, 4)
	rand.Read(suffix)
	name := fmt.Sprintf("test%s", hex.EncodeToString(suffix))
	_, err := pg.DB.Exec("CREATE DATABASE " + name)
	Require(t,
		NoError(err),
	)
	dsn := fmt.Sprintf("host=%s dbname=%s", pg.Host, name)
	sqldb, err := sql.Open("postgres", dsn)
	Require(t,
		NoError(err),
	)
	db := bun.NewDB(sqldb, pgdialect.New())
	Expect(t,
		NoError(db.Ping()),
	)
	return db
}

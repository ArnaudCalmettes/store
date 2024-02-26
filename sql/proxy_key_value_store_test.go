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

type EntryProxy struct {
	bun.BaseModel `bun:"table:entries,alias:e"`

	ID string `bun:",pk"`
	Entry
}

func toEntryProxy(e *Entry) *EntryProxy {
	if e == nil {
		return nil
	}
	return &EntryProxy{Entry: *e}
}

func fromEntryProxy(p *EntryProxy) *Entry {
	if p == nil {
		return nil
	}
	return &p.Entry
}

func TestSQLiteKeyValueStore(t *testing.T) {
	newStore := func(t *testing.T) BaseKeyValueStore[Entry] {
		db := newSQLite(t)
		err := db.ResetModel(context.Background(), (*EntryProxy)(nil))
		Require(t,
			NoError(err),
		)
		return NewKeyValueStoreWithProxy[Entry, EntryProxy](db, toEntryProxy, fromEntryProxy)
	}
	TestBaseKeyValueStore(t, newStore)
}

type PersonProxy struct {
	bun.BaseModel `bun:"table:persons,alias:p"`

	ID       string `bun:",pk"`
	Name     string
	Age      int
	Referent *string
}

func toPersonProxy(p *Person) *PersonProxy {
	return &PersonProxy{
		ID: p.ID, Name: p.Name, Age: p.Age, Referent: p.Referent,
	}
}

func fromPersonProxy(p *PersonProxy) *Person {
	return &Person{
		ID: p.ID, Name: p.Name, Age: p.Age, Referent: p.Referent,
	}
}

func TestSQLiteProxyKeyValueLister(t *testing.T) {
	newStore := func(t *testing.T) TestListerInterface[Person] {
		db := newSQLite(t)
		err := db.ResetModel(context.Background(), (*PersonProxy)(nil))
		Require(t,
			NoError(err),
		)
		return NewKeyValueStoreWithProxy(db, toPersonProxy, fromPersonProxy)
	}
	TestLister(t, newStore)
}

func TestPGKeyValueStore(t *testing.T) {
	pg, err := pgtest.Start()
	Require(t,
		NoError(err),
	)
	t.Cleanup(func() { pg.Stop() })

	newStore := func(t *testing.T) BaseKeyValueStore[Entry] {
		db := newPostgres(t, pg)
		err := db.ResetModel(context.Background(), (*EntryProxy)(nil))
		Require(t, NoError(err))
		return NewKeyValueStoreWithProxy(db, toEntryProxy, fromEntryProxy)
	}
	TestBaseKeyValueStore(t, newStore)
}

func TestPGProxyKeyValueLister(t *testing.T) {
	pg, err := pgtest.Start()
	Require(t,
		NoError(err),
	)

	t.Cleanup(func() { pg.Stop() })
	newStore := func(t *testing.T) TestListerInterface[Person] {
		db := newPostgres(t, pg)
		err := db.ResetModel(context.Background(), (*PersonProxy)(nil))
		Require(t,
			NoError(err),
		)
		return NewKeyValueStoreWithProxy(db, toPersonProxy, fromPersonProxy)
	}
	TestLister(t, newStore)
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

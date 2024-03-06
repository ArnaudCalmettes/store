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
	"database/sql"
	"errors"
	"fmt"
	"slices"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
	"github.com/uptrace/bun/dialect/feature"

	//lint:ignore ST1001 common definitions
	. "github.com/ArnaudCalmettes/store"
	"github.com/ArnaudCalmettes/store/internal/inspect"
	"github.com/ArnaudCalmettes/store/internal/libbun"
	"github.com/ArnaudCalmettes/store/internal/options"
)

type KeyValueStore[T any] interface {
	BaseKeyValueStore[T]
	Lister[T]
	ErrorMapSetter
	Resetter
}

func NewKeyValue[T any](db *bun.DB) KeyValueStore[T] {
	k := &keyValueStore[T]{
		db: db,
		txOptions: &sql.TxOptions{
			Isolation: sql.LevelSerializable,
		},
	}
	var err error
	if k.spec, err = libbun.GetTableSpec[T](); err != nil {
		panic(err)
	}
	if err = k.spec.Validate(); err != nil {
		panic(err)
	}
	if k.getKey, err = inspect.FieldSelector[T, string](k.spec.KeyField); err != nil {
		panic(err)
	}
	k.setKey, _ = inspect.StringFieldSetter[T](k.spec.KeyField)
	k.InitDefaultErrors()
	return k
}

type keyValueStore[T any] struct {
	db   *bun.DB
	spec *libbun.TableSpec
	ErrorMap
	getKey    func(*T) string
	setKey    func(*T, string)
	txOptions *sql.TxOptions
}

func (k *keyValueStore[T]) SetErrorMap(errorMap ErrorMap) {
	k.ErrorMap = errorMap
	k.InitDefaultErrors()
}

func (k *keyValueStore[T]) List(ctx context.Context, opts ...*Options) ([]*T, error) {
	opt, err := options.Merge(opts...)
	if err != nil {
		return nil, errors.Join(k.ErrInvalidOption, err)
	}
	var items []*T
	query := k.db.NewSelect().Model(&items)
	if opt.Filter != nil {
		qb, err := libbun.BuilderForFilter(opt.Filter, k.spec)
		if err != nil {
			return nil, errors.Join(k.ErrInvalidFilter, err)
		}
		query.ApplyQueryBuilder(qb)
	}
	if order := opt.OrderBy; order != nil {
		column, ok := k.spec.ColumnNames[order.Field]
		if !ok {
			return nil, fmt.Errorf("%w: no such column: %s",
				k.ErrInvalidOption, order.Field,
			)
		}
		if order.Descending {
			column += " DESC"
		}
		query.Order(column)
	}
	if opt.Limit != 0 {
		query.Limit(opt.Limit).Offset(opt.Offset)
	}
	err = query.Scan(ctx)
	return items, err
}

func (k *keyValueStore[T]) GetOne(ctx context.Context, key string) (*T, error) {
	if key == "" {
		return nil, k.ErrEmptyKey
	}
	var item T
	query := k.db.NewSelect().Model(&item).Where("? = ?", bun.Ident(k.spec.KeySQL), key)
	err := query.Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = k.ErrNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (k *keyValueStore[T]) GetMany(ctx context.Context, keys []string) (map[string]*T, error) {
	var items []T
	err := k.db.NewSelect().Model(&items).
		Where("? IN (?)", bun.Ident(k.spec.KeySQL), bun.In(keys)).
		Scan(ctx)
	result := make(map[string]*T, len(items))
	for i := range items {
		item := &items[i]
		result[k.getKey(item)] = item
	}
	return result, err
}

func (k *keyValueStore[T]) GetAll(ctx context.Context) (map[string]*T, error) {
	var items []T
	err := k.db.NewSelect().Model(&items).Scan(ctx)
	result := make(map[string]*T, len(items))
	for i := range items {
		item := &items[i]
		result[k.getKey(item)] = item
	}
	return result, err
}

func (k *keyValueStore[T]) SetOne(ctx context.Context, key string, value *T) error {
	if key == "" {
		return k.ErrEmptyKey
	}
	k.setKey(value, key)
	return k.setRequest(ctx, value)
}

func (k *keyValueStore[T]) SetMany(ctx context.Context, items map[string]*T) error {
	values := make([]T, 0, len(items))
	for key, val := range items {
		if key == "" {
			continue
		}
		k.setKey(val, key)
		values = append(values, *val)
	}
	if len(values) == 0 {
		return nil
	}
	return k.setRequest(ctx, &values)
}

func (k *keyValueStore[T]) setRequest(ctx context.Context, model any) error {
	query := k.db.NewInsert().Model(model)
	k.handleInsertConflict(query)
	_, err := query.Exec(ctx)
	return err
}

func (k *keyValueStore[T]) UpdateOne(ctx context.Context, key string, update func(string, *T) (*T, error)) error {
	if key == "" {
		return k.ErrEmptyKey
	}
	return k.UpdateMany(ctx, []string{key}, update)
}

func (k *keyValueStore[T]) UpdateMany(ctx context.Context, keys []string, update func(string, *T) (*T, error)) error {
	if len(keys) == 0 {
		return nil
	}
	keys = slices.DeleteFunc(keys, func(e string) bool { return e == "" })
	return k.db.RunInTx(ctx, k.txOptions, func(ctx context.Context, tx bun.Tx) error {
		var rows []*T
		selectQuery := tx.NewSelect().Model(&rows).Where("? IN (?)", bun.Ident(k.spec.KeySQL), bun.In(keys))
		k.handleLocking(selectQuery)
		err := selectQuery.Scan(ctx)
		if err != nil {
			return err
		}
		initialRows := k.makeRowMap(keys, rows)
		updatedRows := make([]*T, 0, len(initialRows))
		for key, row := range initialRows {
			newRow, err := update(key, row)
			if err != nil {
				return err
			}
			if newRow == nil {
				continue
			}
			k.setKey(newRow, key)
			updatedRows = append(updatedRows, newRow)
		}
		if len(updatedRows) == 0 {
			return nil
		}
		insertQuery := tx.NewInsert().Model(&updatedRows)
		k.handleInsertConflict(insertQuery)
		_, err = insertQuery.Exec(ctx)
		return err
	})
}

func (k *keyValueStore[T]) makeRowMap(keys []string, rows []*T) map[string]*T {
	result := make(map[string]*T, len(keys))
	for _, key := range keys {
		result[key] = nil
	}
	for _, row := range rows {
		key := k.getKey(row)
		result[key] = row
	}
	return result
}

func (k *keyValueStore[T]) Delete(ctx context.Context, keys ...string) error {
	_, err := k.db.NewDelete().Table(k.spec.TableName).
		Where("? IN (?)", bun.Ident(k.spec.KeySQL), bun.In(keys)).
		Exec(ctx)
	return err
}

func (k *keyValueStore[T]) Reset(ctx context.Context) error {
	err := k.db.ResetModel(ctx, (*T)(nil))
	return err
}

func (k *keyValueStore[T]) handleInsertConflict(query *bun.InsertQuery) {
	if k.db.HasFeature(feature.InsertOnConflict) {
		query.On("CONFLICT (?) DO UPDATE", bun.Ident(k.spec.KeySQL))
		for _, column := range k.spec.ColumnNames {
			if column == k.spec.KeySQL {
				continue
			}
			query.Set("?0 = EXCLUDED.?0", bun.Ident(column))
		}
	}
	if k.db.HasFeature(feature.InsertOnDuplicateKey) {
		query.On("DUPLICATE KEY UPDATE")
	}
}

func (k *keyValueStore[T]) handleLocking(query *bun.SelectQuery) {
	switch k.db.Dialect().Name() {
	case dialect.SQLite:
		return
	case dialect.PG, dialect.MySQL:
		query.For("UPDATE")
	}
}

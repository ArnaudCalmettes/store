package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"

	"github.com/uptrace/bun"

	//lint:ignore ST1001 common definitions
	. "github.com/ArnaudCalmettes/store"
	"github.com/ArnaudCalmettes/store/internal/inspect"
	"github.com/ArnaudCalmettes/store/internal/libbun"
)

type KeyValueStore[T any] interface {
	BaseKeyValueStore[T]
	ErrorMapSetter
	Resetter
}

func NewKeyValueStore[T any](db *bun.DB) KeyValueStore[T] {
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

func (k *keyValueStore[T]) GetOne(ctx context.Context, key string) (*T, error) {
	if key == "" {
		return nil, k.ErrEmptyKey
	}
	var item T
	query := k.db.NewSelect().Model(&item).Where("? = ?", bun.Ident(k.spec.KeySQL), key)
	fmt.Println(query.String())
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
	query.On(fmt.Sprintf("CONFLICT (%s) DO UPDATE", k.spec.KeySQL))
	for _, column := range k.spec.ColumnNames {
		if column == k.spec.KeySQL {
			continue
		}
		query.Set(fmt.Sprintf("%s = EXCLUDED.%s", column, column))
	}
	_, err := query.Exec(ctx)
	return err
}

func (k *keyValueStore[T]) UpdateOne(ctx context.Context, key string, update UpdateFunc[T]) error {
	if key == "" {
		return k.ErrEmptyKey
	}
	return k.UpdateMany(ctx, []string{key}, update)
}

func (k *keyValueStore[T]) UpdateMany(ctx context.Context, keys []string, update UpdateFunc[T]) error {
	if len(keys) == 0 {
		return nil
	}
	keys = slices.DeleteFunc(keys, func(e string) bool { return e == "" })
	return k.db.RunInTx(ctx, k.txOptions, func(ctx context.Context, tx bun.Tx) error {
		var rows []T
		err := tx.NewSelect().Model(&rows).
			Where("? IN (?)", bun.Ident(k.spec.KeySQL), bun.In(keys)).
			Scan(ctx)
		if err != nil {
			return err
		}
		toUpdate := make(map[string]*T, len(keys))
		for _, key := range keys {
			toUpdate[key] = nil
		}
		for i := range rows {
			row := &rows[i]
			key := k.getKey(row)
			toUpdate[key] = row

		}
		updatedRows := make([]*T, 0, len(toUpdate))
		for key, row := range toUpdate {
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
		query := tx.NewInsert().Model(&updatedRows)
		query.On(fmt.Sprintf("CONFLICT (%s) DO UPDATE", k.spec.KeySQL))
		for _, column := range k.spec.ColumnNames {
			if column == k.spec.KeySQL {
				continue
			}
			query.Set(fmt.Sprintf("%s = EXCLUDED.%s", column, column))
		}
		_, err = query.Exec(ctx)
		return err
	})
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

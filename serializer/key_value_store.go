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

package serializer

import (
	"context"
	"errors"
	"slices"

	//lint:ignore ST1001 shared definitions
	. "github.com/ArnaudCalmettes/store"
	"github.com/ArnaudCalmettes/store/internal/inspect"
	"github.com/ArnaudCalmettes/store/internal/options"
)

type KeyValueStore[T any] interface {
	BaseKeyValueStore[T]
	Lister[T]
	Resetter
	ErrorMapSetter
}

func NewKeyValue[T any](serializer Serializer[T], storage Map) KeyValueStore[T] {
	k := &keyValueStore[T]{
		storage:    storage,
		Serializer: serializer,
	}
	k.InitDefaultErrors()
	return k
}

type Map interface {
	BaseKeyValueMap
	ErrorMapSetter
	Resetter
}

type keyValueStore[T any] struct {
	storage Map
	Serializer[T]
	ErrorMap
}

func (k *keyValueStore[T]) List(ctx context.Context, opts ...*Options) ([]*T, error) {
	opt, err := options.Merge(opts...)
	if err != nil {
		return nil, errors.Join(k.ErrInvalidOption, err)
	}
	predicate, err := k.getPredicate(opt)
	if err != nil {
		return nil, errors.Join(k.ErrInvalidFilter, err)
	}

	// FIXME: A better implementation would use some form of incremental scan.
	// TODO: Rework when a scanning interface is implemented.
	all, err := k.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*T, 0, len(all))
	for _, item := range all {
		if predicate(item) {
			result = append(result, item)
		}
	}

	if err := k.order(result, opt.OrderBy); err != nil {
		return nil, err
	}
	return k.paginate(result, opt), nil
}

func (k *keyValueStore[T]) getPredicate(opt *Options) (func(*T) bool, error) {
	filterPred := func(*T) bool { return true }
	if opt.Filter != nil {
		var err error
		filterPred, err = inspect.NewPredicate[T](opt.Filter)
		if err != nil {
			return nil, err
		}
	}
	return filterPred, nil
}

func (k *keyValueStore[T]) order(items []*T, order *OrderBySpec) error {
	if order == nil {
		return nil
	}
	cmp, err := inspect.NewCmp[T](order)
	if err != nil {
		return errors.Join(ErrInvalidOption, err)
	}
	slices.SortStableFunc(items, cmp)
	return nil
}

func (k *keyValueStore[T]) paginate(result []*T, opt *Options) []*T {
	if opt.Offset > len(result) {
		result = result[:0]
	} else {
		result = result[opt.Offset:]
	}
	if opt.Limit > 0 && opt.Limit < len(result) {
		result = result[:opt.Limit]
	}
	return result
}

func (k *keyValueStore[T]) SetErrorMap(errorMap ErrorMap) {
	k.storage.SetErrorMap(errorMap)
	k.ErrorMap = errorMap
	k.InitDefaultErrors()
}

func (k *keyValueStore[T]) GetOne(ctx context.Context, key string) (*T, error) {
	if key == "" {
		return nil, k.ErrEmptyKey
	}
	data, err := k.storage.GetOne(ctx, key)
	if err != nil {
		return nil, err
	}
	value, err := k.Deserialize(data)
	if err != nil {
		err = errors.Join(k.ErrDeserialize, err)
	}
	return value, err
}

func (k *keyValueStore[T]) GetMany(ctx context.Context, keys []string) (map[string]*T, error) {
	serializedItems, err := k.storage.GetMany(ctx, keys)
	if err != nil {
		return nil, err
	}
	return k.deserializeMap(serializedItems)
}

func (k *keyValueStore[T]) GetAll(ctx context.Context) (map[string]*T, error) {
	all, err := k.storage.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return k.deserializeMap(all)
}

func (k *keyValueStore[T]) SetOne(ctx context.Context, key string, value *T) error {
	data, err := k.Serialize(value)
	if err != nil {
		return errors.Join(k.ErrSerialize, err)
	}
	return k.storage.SetOne(ctx, key, data)
}

func (k *keyValueStore[T]) SetMany(ctx context.Context, items map[string]*T) error {
	serializedItems, err := k.serializeMap(items)
	if err != nil {
		return err
	}
	return k.storage.SetMany(ctx, serializedItems)
}

func (k *keyValueStore[T]) UpdateOne(ctx context.Context, key string, update func(string, *T) (*T, error)) error {
	return k.storage.UpdateOne(ctx, key, k.updateCallback(update))
}

func (k *keyValueStore[T]) UpdateMany(ctx context.Context, keys []string, update func(string, *T) (*T, error)) error {
	return k.storage.UpdateMany(ctx, keys, k.updateCallback(update))
}

func (k *keyValueStore[T]) Delete(ctx context.Context, keys ...string) error {
	return k.storage.Delete(ctx, keys...)
}

func (k *keyValueStore[T]) Reset(ctx context.Context) error {
	return k.storage.Reset(ctx)
}

func (k *keyValueStore[T]) serializeMap(in map[string]*T) (map[string]string, error) {
	out := make(map[string]string, len(in))
	for key, value := range in {
		data, err := k.Serialize(value)
		if err != nil {
			return nil, errors.Join(k.ErrSerialize, err)
		}
		out[key] = data
	}
	return out, nil
}

func (k *keyValueStore[T]) deserializeMap(in map[string]string) (map[string]*T, error) {
	out := make(map[string]*T, len(in))
	for key, data := range in {
		value, err := k.Deserialize(data)
		if err != nil {
			return nil, errors.Join(k.ErrDeserialize, err)
		}
		out[key] = value
	}
	return out, nil
}

func (k *keyValueStore[T]) updateCallback(in func(string, *T) (*T, error)) func(string, *string) (*string, error) {
	return func(id string, data *string) (*string, error) {
		var value *T
		var err error
		if data != nil {
			value, err = k.Deserialize(*data)
			if err != nil {
				return nil, errors.Join(k.ErrDeserialize, err)
			}
		}
		newValue, err := in(id, value)
		if err != nil || newValue == nil {
			return nil, err
		}
		newData, _ := k.Serialize(newValue)
		return &newData, err
	}
}

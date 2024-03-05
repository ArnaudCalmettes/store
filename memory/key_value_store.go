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

package memory

import (
	"context"
	"errors"
	"maps"
	"slices"
	"sync"

	//lint:ignore ST1001 common definitions
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

func NewKeyValueStore[T any]() KeyValueStore[T] {
	k := &keyValueStore[T]{
		items: make(map[string]T),
	}
	k.InitDefaultErrors()
	return k
}

type keyValueStore[T any] struct {
	items map[string]T
	mtx   sync.RWMutex
	ErrorMap
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
	predicate, err := k.getPredicate(opt)
	if err != nil {
		return nil, errors.Join(k.ErrInvalidFilter, err)
	}

	k.mtx.RLock()
	result := make([]*T, 0, len(k.items))
	for _, item := range k.items {
		if predicate(&item) {
			result = append(result, &item)
		}
	}
	k.mtx.RUnlock()

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

func (k *keyValueStore[T]) GetOne(ctx context.Context, key string) (*T, error) {
	k.mtx.RLock()
	defer k.mtx.RUnlock()
	if key == "" {
		return nil, k.ErrEmptyKey
	}
	value, ok := k.items[key]
	if !ok {
		return nil, k.ErrNotFound
	}
	return &value, nil
}

func (k *keyValueStore[T]) GetMany(ctx context.Context, keys []string) (map[string]*T, error) {
	k.mtx.RLock()
	defer k.mtx.RUnlock()
	items := make(map[string]*T, len(keys))
	for _, key := range keys {
		value, ok := k.items[key]
		if !ok {
			continue
		}
		items[key] = &value
	}
	return items, nil
}

func (k *keyValueStore[T]) GetAll(ctx context.Context) (map[string]*T, error) {
	k.mtx.RLock()
	defer k.mtx.RUnlock()
	items := make(map[string]*T, len(k.items))
	for key := range k.items {
		value := k.items[key]
		items[key] = &value
	}
	return items, nil
}

func (k *keyValueStore[T]) SetOne(ctx context.Context, key string, value *T) error {
	k.mtx.Lock()
	defer k.mtx.Unlock()
	if key == "" {
		return k.ErrEmptyKey
	}
	k.items[key] = *value
	return nil
}

func (k *keyValueStore[T]) SetMany(ctx context.Context, items map[string]*T) error {
	k.mtx.Lock()
	defer k.mtx.Unlock()
	for key, value := range items {
		if key == "" {
			continue
		}
		k.items[key] = *value
	}
	return nil
}

func (k *keyValueStore[T]) UpdateOne(ctx context.Context, key string, update func(string, *T) (*T, error)) error {
	k.mtx.Lock()
	defer k.mtx.Unlock()
	if key == "" {
		return k.ErrEmptyKey
	}
	var valuePtr *T
	value, ok := k.items[key]
	if ok {
		valuePtr = &value
	}
	newValue, err := update(key, valuePtr)
	if err != nil {
		return err
	}
	if newValue == nil {
		return nil
	}
	k.items[key] = *newValue
	return nil
}

func (k *keyValueStore[T]) UpdateMany(ctx context.Context, keys []string, update func(string, *T) (*T, error)) error {
	k.mtx.Lock()
	defer k.mtx.Unlock()

	updatedValues := make(map[string]T, len(keys))
	for _, key := range keys {
		if key == "" {
			continue
		}
		var valuePtr *T
		value, ok := k.items[key]
		if ok {
			valuePtr = &value
		}
		newValue, err := update(key, valuePtr)
		if err != nil {
			return err
		}
		if newValue != nil {
			updatedValues[key] = *newValue
		}
	}
	maps.Copy(k.items, updatedValues)
	return nil
}

func (k *keyValueStore[T]) Delete(ctx context.Context, keys ...string) error {
	k.mtx.Lock()
	defer k.mtx.Unlock()
	for _, key := range keys {
		delete(k.items, key)
	}
	return nil
}

func (k *keyValueStore[T]) Reset(ctx context.Context) error {
	k.mtx.Lock()
	defer k.mtx.Unlock()
	k.items = map[string]T{}
	return nil
}

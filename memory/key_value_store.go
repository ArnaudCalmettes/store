package memory

import (
	"context"
	"maps"
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
	opt := options.Merge(opts...)
	predicate, err := k.getPredicate(opt)
	if err != nil {
		return nil, err
	}
	k.mtx.RLock()
	defer k.mtx.RUnlock()
	result := make([]*T, 0, len(k.items))
	for _, item := range k.items {
		val := item
		if predicate(&val) {
			result = append(result, &item)
		}
	}
	return result, nil
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

func (k *keyValueStore[T]) UpdateOne(ctx context.Context, key string, update UpdateFunc[T]) error {
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

func (k *keyValueStore[T]) UpdateMany(ctx context.Context, keys []string, update UpdateFunc[T]) error {
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

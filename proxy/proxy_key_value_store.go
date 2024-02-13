package proxy

import (
	"context"

	//lint:ignore ST1001 common definitions
	. "github.com/ArnaudCalmettes/store"
)

type KeyValueStore[T any] interface {
	BaseKeyValueStore[T]
	ErrorMapSetter
	Resetter
}

func NewKeyValueStoreWithProxy[T, P any](
	inner KeyValueStore[P],
	toProxy func(*T) *P,
	fromProxy func(*P) *T,
) KeyValueStore[T] {
	return &keyValueStore[T, P]{
		inner:     inner,
		toProxy:   toProxy,
		fromProxy: fromProxy,
	}
}

type keyValueStore[T, P any] struct {
	inner     KeyValueStore[P]
	toProxy   func(*T) *P
	fromProxy func(*P) *T
}

func (k *keyValueStore[T, P]) Reset(ctx context.Context) error {
	return k.inner.Reset(ctx)
}

func (k *keyValueStore[T, P]) SetErrorMap(errorMap ErrorMap) {
	k.inner.SetErrorMap(errorMap)
}

func (k *keyValueStore[T, P]) GetOne(ctx context.Context, key string) (*T, error) {
	proxy, err := k.inner.GetOne(ctx, key)
	return k.fromProxy(proxy), err
}

func (k *keyValueStore[T, P]) GetMany(ctx context.Context, keys []string) (map[string]*T, error) {
	proxies, err := k.inner.GetMany(ctx, keys)
	items := make(map[string]*T, len(proxies))
	for key, proxy := range proxies {
		items[key] = k.fromProxy(proxy)
	}
	return items, err
}

func (k *keyValueStore[T, P]) GetAll(ctx context.Context) (map[string]*T, error) {
	proxies, err := k.inner.GetAll(ctx)
	items := make(map[string]*T, len(proxies))
	for key, proxy := range proxies {
		items[key] = k.fromProxy(proxy)
	}
	return items, err
}

func (k *keyValueStore[T, P]) SetOne(ctx context.Context, key string, value *T) error {
	return k.inner.SetOne(ctx, key, k.toProxy(value))
}

func (k *keyValueStore[T, P]) SetMany(ctx context.Context, items map[string]*T) error {
	proxies := make(map[string]*P, len(items))
	for key, item := range items {
		proxies[key] = k.toProxy(item)
	}
	return k.inner.SetMany(ctx, proxies)
}

func (k *keyValueStore[T, P]) UpdateOne(ctx context.Context, key string, f UpdateFunc[T]) error {
	return k.inner.UpdateOne(ctx, key, k.updateFunc(f))
}

func (k *keyValueStore[T, P]) UpdateMany(ctx context.Context, keys []string, f UpdateFunc[T]) error {
	return k.inner.UpdateMany(ctx, keys, k.updateFunc(f))
}

func (k *keyValueStore[T, P]) updateFunc(update UpdateFunc[T]) UpdateFunc[P] {
	return func(key string, proxy *P) (*P, error) {
		item := k.fromProxy(proxy)
		newItem, err := update(key, item)
		return k.toProxy(newItem), err
	}
}

func (k *keyValueStore[T, P]) Delete(ctx context.Context, keys ...string) error {
	return k.inner.Delete(ctx, keys...)
}

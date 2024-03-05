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

package proxy

import (
	"context"

	//lint:ignore ST1001 common definitions
	. "github.com/ArnaudCalmettes/store"
)

type KeyValueStore[T any] interface {
	BaseKeyValueStore[T]
	Lister[T]
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

func (k *keyValueStore[T, P]) List(ctx context.Context, opts ...*Options) ([]*T, error) {
	proxies, err := k.inner.List(ctx, opts...)
	items := make([]*T, len(proxies))
	for i, p := range proxies {
		items[i] = k.fromProxy(p)
	}
	return items, err
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

func (k *keyValueStore[T, P]) UpdateOne(ctx context.Context, key string, f func(string, *T) (*T, error)) error {
	return k.inner.UpdateOne(ctx, key, k.updateFunc(f))
}

func (k *keyValueStore[T, P]) UpdateMany(ctx context.Context, keys []string, f func(string, *T) (*T, error)) error {
	return k.inner.UpdateMany(ctx, keys, k.updateFunc(f))
}

func (k *keyValueStore[T, P]) updateFunc(update func(string, *T) (*T, error)) func(string, *P) (*P, error) {
	return func(key string, proxy *P) (*P, error) {
		item := k.fromProxy(proxy)
		newItem, err := update(key, item)
		return k.toProxy(newItem), err
	}
}

func (k *keyValueStore[T, P]) Delete(ctx context.Context, keys ...string) error {
	return k.inner.Delete(ctx, keys...)
}

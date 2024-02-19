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
	"maps"
	"sync"

	//lint:ignore ST1001 common definitions
	. "github.com/ArnaudCalmettes/store"
)

type KeyValueMap interface {
	BaseKeyValueMap
	Resetter
	ErrorMapSetter
}

func NewKeyValueMap() KeyValueMap {
	k := &keyValueMap{
		items: make(map[string]string),
	}
	k.InitDefaultErrors()
	return k
}

type keyValueMap struct {
	items map[string]string
	mtx   sync.RWMutex
	ErrorMap
}

func (k *keyValueMap) SetErrorMap(errorMap ErrorMap) {
	k.ErrorMap = errorMap
	k.InitDefaultErrors()
}

func (k *keyValueMap) SetOne(ctx context.Context, key string, value string) error {
	k.mtx.Lock()
	defer k.mtx.Unlock()
	if key == "" {
		return k.ErrEmptyKey
	}
	k.items[key] = value
	return nil
}

func (k *keyValueMap) SetMany(ctx context.Context, items map[string]string) error {
	k.mtx.Lock()
	defer k.mtx.Unlock()
	maps.Copy(k.items, items)
	delete(k.items, "")
	return nil
}

func (k *keyValueMap) GetOne(ctx context.Context, key string) (string, error) {
	k.mtx.RLock()
	defer k.mtx.RUnlock()
	value, ok := k.items[key]
	if !ok {
		return "", k.ErrNotFound
	}
	return value, nil
}

func (k *keyValueMap) GetMany(ctx context.Context, keys []string) (map[string]string, error) {
	k.mtx.RLock()
	defer k.mtx.RUnlock()
	items := make(map[string]string, len(keys))
	for _, key := range keys {
		value, ok := k.items[key]
		if !ok {
			continue
		}
		items[key] = value
	}
	return items, nil
}

func (k *keyValueMap) GetAll(ctx context.Context) (map[string]string, error) {
	k.mtx.RLock()
	defer k.mtx.RUnlock()
	items := maps.Clone(k.items)
	return items, nil
}

func (k *keyValueMap) UpdateOne(ctx context.Context, key string, update UpdateFunc[string]) error {
	k.mtx.Lock()
	defer k.mtx.Unlock()
	if key == "" {
		return k.ErrEmptyKey
	}
	var valuePtr *string
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

func (k *keyValueMap) UpdateMany(ctx context.Context, keys []string, update UpdateFunc[string]) error {
	k.mtx.Lock()
	defer k.mtx.Unlock()

	updatedValues := make(map[string]string, len(keys))
	for _, key := range keys {
		if key == "" {
			continue
		}
		var valuePtr *string
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

func (k *keyValueMap) Delete(ctx context.Context, keys ...string) error {
	k.mtx.Lock()
	defer k.mtx.Unlock()
	for _, key := range keys {
		delete(k.items, key)
	}
	return nil
}

func (k *keyValueMap) Reset(ctx context.Context) error {
	k.mtx.Lock()
	defer k.mtx.Unlock()
	k.items = map[string]string{}
	return nil
}

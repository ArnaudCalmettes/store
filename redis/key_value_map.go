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

package redis

import (
	"context"
	"errors"

	//lint:ignore ST1001 shared definitions
	. "github.com/ArnaudCalmettes/store"
	"github.com/go-redis/redis/v8"
)

type KeyValueMap interface {
	BaseKeyValueMap
	Resetter
	ErrorMapSetter
}

func NewKeyValueMap(rdb redis.UniversalClient, namespace string) KeyValueMap {
	k := &keyValueMap{
		rdb:       rdb,
		namespace: namespace,
	}
	k.InitDefaultErrors()
	return k
}

type keyValueMap struct {
	rdb       redis.UniversalClient
	namespace string
	ErrorMap
}

func (k *keyValueMap) SetErrorMap(errorMap ErrorMap) {
	k.ErrorMap = errorMap
	k.InitDefaultErrors()
}

func (k *keyValueMap) SetOne(ctx context.Context, key string, value string) error {
	if key == "" {
		return k.ErrEmptyKey
	}
	return k.rdb.HSet(ctx, k.namespace, key, value).Err()
}

func (k *keyValueMap) SetMany(ctx context.Context, items map[string]string) error {
	delete(items, "")
	if len(items) == 0 {
		return nil
	}
	return k.rdb.HSet(ctx, k.namespace, items).Err()
}

func (k *keyValueMap) GetOne(ctx context.Context, key string) (string, error) {
	value, err := k.rdb.HGet(ctx, k.namespace, key).Result()
	if err == redis.Nil {
		err = k.ErrNotFound
	}
	return value, err
}

func (k *keyValueMap) GetMany(ctx context.Context, keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return map[string]string{}, nil
	}
	values, err := k.rdb.HMGet(ctx, k.namespace, keys...).Result()
	result := make(map[string]string, len(keys))
	for i, key := range keys {
		value := values[i]
		if value == nil {
			continue
		}
		result[key] = value.(string)
	}
	return result, err
}

func (k *keyValueMap) GetAll(ctx context.Context) (map[string]string, error) {
	return k.rdb.HGetAll(ctx, k.namespace).Result()
}

func (k *keyValueMap) UpdateOne(ctx context.Context, key string, update UpdateFunc[string]) error {
	if key == "" {
		return k.ErrEmptyKey
	}
	txFunc := func(tx *redis.Tx) error {
		value, err := tx.HGet(ctx, k.namespace, key).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return err
		}
		var valuePtr *string
		if !errors.Is(err, redis.Nil) {
			valuePtr = &value
		}
		newValue, err := update(key, valuePtr)
		if err != nil {
			return err
		}
		if newValue == nil {
			return err
		}
		return tx.HSet(ctx, k.namespace, key, *newValue).Err()
	}
	var err error
	for i := 0; i < 10; i++ {
		err = k.rdb.Watch(ctx, txFunc, k.namespace)
		if err == redis.TxFailedErr {
			continue
		}
		break
	}
	return err
}

func (k *keyValueMap) UpdateMany(ctx context.Context, keys []string, update UpdateFunc[string]) error {
	if len(keys) == 0 {
		return nil
	}
	txFunc := func(tx *redis.Tx) error {
		values, err := tx.HMGet(ctx, k.namespace, keys...).Result()
		updated := make(map[string]string, len(keys))
		for i, value := range values {
			if keys[i] == "" {
				continue
			}

			var valuePtr *string
			if value != nil {
				val := value.(string)
				valuePtr = &val
			}
			newValue, err := update(keys[i], valuePtr)
			if err != nil {
				return err
			}
			if newValue != nil {
				updated[keys[i]] = *newValue
			}
		}
		if len(updated) == 0 {
			return err
		}
		return tx.HSet(ctx, k.namespace, updated).Err()
	}
	var err error
	for i := 0; i < 10; i++ {
		err = k.rdb.Watch(ctx, txFunc, k.namespace)
		if err == redis.TxFailedErr {
			continue
		}
		break
	}
	return err
}

func (k *keyValueMap) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return k.rdb.HDel(ctx, k.namespace, keys...).Err()
}

func (k *keyValueMap) Reset(ctx context.Context) error {
	return k.rdb.Del(ctx, k.namespace).Err()
}

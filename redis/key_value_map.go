package redis

import (
	"context"
	"errors"

	//lint:ignore ST1001 shared definitions
	. "github.com/ArnaudCalmettes/store"
	"github.com/go-redis/redis/v8"
)

func NewKeyValueMap(rdb redis.UniversalClient, namespace string) *keyValueMap {
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

func (k *keyValueMap) WithErrorMap(errorMap ErrorMap) *keyValueMap {
	k.ErrorMap = errorMap
	k.InitDefaultErrors()
	return k
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

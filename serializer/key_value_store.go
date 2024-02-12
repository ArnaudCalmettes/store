package serializer

import (
	"context"
	"errors"

	//lint:ignore ST1001 shared definitions
	. "github.com/ArnaudCalmettes/store"
)

var (
	_ BaseKeyValueStore[string] = (*keyValueStore[string])(nil)
	_ Resetter                  = (*keyValueStore[string])(nil)
)

func NewKeyValueStore[T any](serializer Serializer[T], storage MapInterface) *keyValueStore[T] {
	k := &keyValueStore[T]{
		storage:    storage,
		Serializer: serializer,
	}
	k.InitDefaultErrors()
	return k
}

type MapInterface interface {
	BaseKeyValueMap
	ErrorMapSetter
	Resetter
}

type keyValueStore[T any] struct {
	storage MapInterface
	Serializer[T]
	ErrorMap
}

func (k *keyValueStore[T]) SetErrorMap(errorMap ErrorMap) {
	k.storage.SetErrorMap(errorMap)
	k.ErrorMap = errorMap
	k.InitDefaultErrors()
}

func (k *keyValueStore[T]) GetOne(ctx context.Context, key string) (*T, error) {
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

func (k *keyValueStore[T]) UpdateOne(ctx context.Context, key string, update UpdateFunc[T]) error {
	return k.storage.UpdateOne(ctx, key, k.updateCallback(update))
}

func (k *keyValueStore[T]) UpdateMany(ctx context.Context, keys []string, update UpdateFunc[T]) error {
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

func (k *keyValueStore[T]) updateCallback(in UpdateFunc[T]) UpdateFunc[string] {
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
		newData, err := k.Serialize(newValue)
		if err != nil {
			return nil, err
		}
		return &newData, err
	}
}

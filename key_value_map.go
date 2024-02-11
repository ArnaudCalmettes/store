package store

import (
	"context"
)

type Resetter interface {
	Reset(ctx context.Context) error
}

type UpdateFunc[Key comparable, Value any] func(key Key, value *Value) (*Value, error)
type MapUpdateFunc = UpdateFunc[string, string]

type KeyValueMap interface {
	SetOne(ctx context.Context, key, value string) error
	SetMany(ctx context.Context, items map[string]string) error
	GetOne(ctx context.Context, key string) (string, error)
	GetMany(ctx context.Context, keys []string) (map[string]string, error)
	GetAll(ctx context.Context) (map[string]string, error)
	UpdateOne(ctx context.Context, key string, update MapUpdateFunc) error
	UpdateMany(ctx context.Context, keys []string, update MapUpdateFunc) error
	Delete(ctx context.Context, keys ...string) error
}

package store

import (
	"context"
)

type Resetter interface {
	Reset(ctx context.Context) error
}

type UpdateFunc[T any] func(key string, value *T) (*T, error)

type BaseKeyValueMap interface {
	SetOne(ctx context.Context, key, value string) error
	SetMany(ctx context.Context, items map[string]string) error
	GetOne(ctx context.Context, key string) (string, error)
	GetMany(ctx context.Context, keys []string) (map[string]string, error)
	GetAll(ctx context.Context) (map[string]string, error)
	UpdateOne(ctx context.Context, key string, update UpdateFunc[string]) error
	UpdateMany(ctx context.Context, keys []string, update UpdateFunc[string]) error
	Delete(ctx context.Context, keys ...string) error
}

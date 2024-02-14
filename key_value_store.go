package store

import "context"

type BaseKeyValueStore[T any] interface {
	GetOne(ctx context.Context, key string) (*T, error)
	GetMany(ctx context.Context, keys []string) (map[string]*T, error)
	GetAll(ctx context.Context) (map[string]*T, error)
	SetOne(ctx context.Context, key string, value *T) error
	SetMany(ctx context.Context, items map[string]*T) error
	UpdateOne(ctx context.Context, key string, update UpdateFunc[T]) error
	UpdateMany(ctx context.Context, keys []string, update UpdateFunc[T]) error
	Delete(ctx context.Context, keys ...string) error
}

type Options struct {
	Filter *Filter
}

type Lister[T any] interface {
	List(ctx context.Context, opts ...*Options)
}

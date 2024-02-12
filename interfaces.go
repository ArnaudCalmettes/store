package store

import "context"

type Resetter interface {
	Reset(ctx context.Context) error
}

type UpdateFunc[T any] func(key string, value *T) (*T, error)

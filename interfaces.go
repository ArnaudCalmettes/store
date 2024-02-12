package store

import "context"

type ErrorMapSetter interface {
	SetErrorMap(ErrorMap)
}

type Resetter interface {
	Reset(ctx context.Context) error
}

type Serializer[T any] interface {
	Serialize(*T) (string, error)
	Deserialize(string) (*T, error)
}

type UpdateFunc[T any] func(key string, value *T) (*T, error)

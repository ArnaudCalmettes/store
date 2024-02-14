package redis

import (
	//lint:ignore ST1001 common definitions
	. "github.com/ArnaudCalmettes/store"
	"github.com/ArnaudCalmettes/store/serializer"
	"github.com/go-redis/redis/v8"
)

type KeyValueStore[T any] interface {
	BaseKeyValueStore[T]
	Lister[T]
	ErrorMapSetter
	Resetter
}

func NewKeyValueStore[T any](rdb redis.UniversalClient, namespace string, s Serializer[T]) KeyValueStore[T] {
	return serializer.NewKeyValueStore[T](s, NewKeyValueMap(rdb, namespace))
}

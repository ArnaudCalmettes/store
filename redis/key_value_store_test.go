package redis

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/ArnaudCalmettes/store"
	"github.com/ArnaudCalmettes/store/serializer"
	"github.com/ArnaudCalmettes/store/test"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

func TestRedisKeyValueStore(t *testing.T) {
	newStoreConstructor := spawnNewKeyValueStore[test.Entry](t)
	newStore := func() store.BaseKeyValueStore[test.Entry] {
		return newStoreConstructor()
	}
	test.TestBaseKeyValueStore(t, newStore)
}

func spawnNewKeyValueStore[T any](t *testing.T) func() KeyValueStoreInterface[T] {
	t.Helper()
	s := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	return func() KeyValueStoreInterface[T] {
		suffix := make([]byte, 4)
		rand.Read(suffix)
		namespace := fmt.Sprintf("key_value_store_%s", hex.EncodeToString(suffix))
		return NewKeyValueStore[T](rdb, namespace, serializer.NewJSON[T]())
	}
}

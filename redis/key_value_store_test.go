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
	newStore := func(t *testing.T) store.BaseKeyValueStore[test.Entry] {
		return newStoreConstructor(t)
	}
	test.TestBaseKeyValueStore(t, newStore)
}

func spawnNewKeyValueStore[T any](t *testing.T) func(*testing.T) KeyValueStore[T] {
	t.Helper()
	s := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	return func(*testing.T) KeyValueStore[T] {
		suffix := make([]byte, 4)
		rand.Read(suffix)
		namespace := fmt.Sprintf("key_value_store_%s", hex.EncodeToString(suffix))
		t.Cleanup(func() {
			rdb.Del(context.Background(), namespace).Err()
		})
		return NewKeyValueStore[T](rdb, namespace, serializer.NewJSON[T]())
	}
}

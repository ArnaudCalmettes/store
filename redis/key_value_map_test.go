package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"

	. "github.com/ArnaudCalmettes/store"
	. "github.com/ArnaudCalmettes/store/test"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

func TestRedisKeyValueMap(t *testing.T) {
	TestBaseKeyValueMap(t, makeNewKeyValueMap(t))
}

func TestKeyValueMapCustomErrors(t *testing.T) {
	errTest := errors.New("test")
	s := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	store := NewKeyValueMap(rdb, "test_custom_errors")
	store.WithErrorMap(ErrorMap{
		ErrNotFound: errTest,
	})
	_, err := store.GetOne(context.Background(), "does not exist")
	Require(t,
		IsError(errTest, err),
	)
}

func TestKeyValueMapReset(t *testing.T) {
	s := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	store := NewKeyValueMap(rdb, "test_reset")

	err := store.SetMany(context.Background(), map[string]string{
		"one":   "two",
		"three": "four",
	})
	Require(t,
		NoError(err),
	)

	err = store.Reset(context.Background())
	Require(t,
		NoError(err),
	)

	all, err := store.GetAll(context.Background())
	Require(t,
		NoError(err),
		Equal(map[string]string{}, all),
	)
}

func makeNewKeyValueMap(t *testing.T) func() BaseKeyValueMap {
	t.Helper()
	s := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	newKeyValueMap := func() BaseKeyValueMap {
		suffix := make([]byte, 4)
		rand.Read(suffix)
		namespace := fmt.Sprintf("key_value_map_%s", hex.EncodeToString(suffix))
		return NewKeyValueMap(rdb, namespace)
	}
	return newKeyValueMap
}

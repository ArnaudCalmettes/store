package sqlite

import (
	"github.com/ArnaudCalmettes/store/proxy"
	"github.com/uptrace/bun"
)

func NewKeyValueStoreWithProxy[T, P any](
	db *bun.DB,
	toProxy func(*T) *P,
	fromProxy func(*P) *T,
) KeyValueStore[T] {
	return proxy.NewKeyValueStoreWithProxy[T, P](
		NewKeyValueStore[P](db),
		toProxy,
		fromProxy,
	)
}

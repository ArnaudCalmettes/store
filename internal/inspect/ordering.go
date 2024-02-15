package inspect

import (
	"reflect"
	"time"

	"cmp"

	"github.com/ArnaudCalmettes/store"
)

func NewCmp[T any](order *store.OrderBySpec) (func(*T, *T) int, error) {
	var zero T
	if _, ok := reflect.TypeOf(zero).FieldByName(order.Field); !ok {
		return nil, errNoSuchField
	}
	v := reflect.ValueOf(zero).FieldByName(order.Field).Interface()
	switch v.(type) {
	case string:
		return orderedCmp[T, string](order)
	case int:
		return orderedCmp[T, int](order)
	case int8:
		return orderedCmp[T, int8](order)
	case int16:
		return orderedCmp[T, int16](order)
	case int32:
		return orderedCmp[T, int32](order)
	case int64:
		return orderedCmp[T, int64](order)
	case uint:
		return orderedCmp[T, uint](order)
	case uint8:
		return orderedCmp[T, uint8](order)
	case uint16:
		return orderedCmp[T, uint16](order)
	case uint32:
		return orderedCmp[T, uint32](order)
	case uint64:
		return orderedCmp[T, uint64](order)
	case float32:
		return orderedCmp[T, float32](order)
	case float64:
		return orderedCmp[T, float64](order)
	case time.Time:
		return timeCmp[T](order)
	}
	return nil, errTypeNotSupported
}

func orderedCmp[T any, F cmp.Ordered](order *store.OrderBySpec) (func(*T, *T) int, error) {
	get, _ := FieldSelector[T, F](order.Field)
	if order.Descending {
		return func(a, b *T) int { return cmp.Compare(get(b), get(a)) }, nil
	}
	return func(a, b *T) int { return cmp.Compare(get(a), get(b)) }, nil
}

func timeCmp[T any](order *store.OrderBySpec) (func(*T, *T) int, error) {
	get, _ := FieldSelector[T, time.Time](order.Field)
	if order.Descending {
		return func(a, b *T) int { return get(b).Compare(get(a)) }, nil
	}
	return func(a, b *T) int { return get(a).Compare(get(b)) }, nil
}

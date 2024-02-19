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

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
	"errors"
	"fmt"
	"reflect"
)

var (
	errNotAStruct   = errors.New("not a struct type")
	errNoSuchField  = errors.New("no such field")
	errTypeMismatch = errors.New("type mismatch")
)

func FieldSelector[T, K any](name string) (func(*T) K, error) {
	var zeroStruct T
	var zeroValue K
	structType := reflect.TypeOf(zeroStruct)
	fieldType := reflect.TypeOf(zeroValue)

	if structType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%s is %w", structType.Name(), errNotAStruct)
	}
	field, ok := structType.FieldByName(name)
	if !ok {
		return nil, fmt.Errorf("%w: %q in type %s",
			errNoSuchField, name, structType.Name(),
		)
	}
	if field.Type != fieldType {
		return nil, fmt.Errorf("%w: field %s is of type %s, not %s",
			errTypeMismatch, name, field.Type.Name(), fieldType.Name(),
		)
	}
	selector := func(obj *T) K {
		val := reflect.ValueOf(*obj)
		return val.FieldByIndex(field.Index).Interface().(K)
	}
	return selector, nil
}

func StringFieldSetter[T any](name string) (func(*T, string), error) {
	var zeroStruct T
	typ := reflect.TypeOf(zeroStruct)
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%s is %w", typ.Name(), errNotAStruct)
	}
	field, ok := typ.FieldByName(name)
	if !ok {
		return nil, fmt.Errorf("%w: %q in type %s",
			errNoSuchField, name, typ.Name(),
		)
	}
	if field.Type != reflect.TypeOf("") {
		return nil, fmt.Errorf("%w: field %s is of type %s, not string",
			errTypeMismatch, name, field.Type.Name(),
		)
	}

	setter := func(obj *T, value string) {
		val := reflect.ValueOf(obj).Elem()
		val.FieldByIndex(field.Index).SetString(value)
	}
	return setter, nil
}

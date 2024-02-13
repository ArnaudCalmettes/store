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

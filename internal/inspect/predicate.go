package inspect

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"cmp"

	"github.com/ArnaudCalmettes/store"
)

var (
	errInvalidFilter = errors.New("invalid filter")
)

func NewPredicate[T any](filter *store.FilterSpec) (func(*T) bool, error) {
	switch {
	case filter.Where != nil:
		return predicateFromWhereClause[T](filter.Where)
	case filter.All != nil:
		return predicateAll[T](filter.All)
	case filter.Any != nil:
		return predicateAny[T](filter.Any)
	default:
		return nil, errInvalidFilter
	}
}

var (
	errTypeNotSupported = errors.New("type not supported")
)

func predicateFromWhereClause[T any](w *store.WhereClause) (func(*T) bool, error) {
	if w.Value == nil {
		return pointerPredicate[T](w.Field, w.Op)
	}
	switch t := w.Value.(type) {
	case string:
		return orderedPredicate[T](w.Field, w.Op, t)
	case int:
		return orderedPredicate[T](w.Field, w.Op, t)
	case int8:
		return orderedPredicate[T](w.Field, w.Op, t)
	case int16:
		return orderedPredicate[T](w.Field, w.Op, t)
	case int32:
		return orderedPredicate[T](w.Field, w.Op, t)
	case int64:
		return orderedPredicate[T](w.Field, w.Op, t)
	case uint:
		return orderedPredicate[T](w.Field, w.Op, t)
	case uint32:
		return orderedPredicate[T](w.Field, w.Op, t)
	case uint64:
		return orderedPredicate[T](w.Field, w.Op, t)
	case float32:
		return orderedPredicate[T](w.Field, w.Op, t)
	case float64:
		return orderedPredicate[T](w.Field, w.Op, t)
	case bool:
		return comparablePredicate[T](w.Field, w.Op, t)
	case time.Time:
		return timePredicate[T](w.Field, w.Op, t)
	}

	return nil, fmt.Errorf("%w: %s", errTypeNotSupported, reflect.TypeOf(w.Value).Name())
}

func predicateAll[T any](all []*store.FilterSpec) (func(*T) bool, error) {
	var err error
	preds := make([]func(*T) bool, len(all))
	for i, filter := range all {
		preds[i], err = NewPredicate[T](filter)
		if err != nil {
			return nil, err
		}
	}
	pred := func(obj *T) bool {
		for _, pred := range preds {
			if !pred(obj) {
				return false
			}
		}
		return true
	}
	return pred, nil
}

func predicateAny[T any](all []*store.FilterSpec) (func(*T) bool, error) {
	var err error
	preds := make([]func(*T) bool, len(all))
	for i, filter := range all {
		preds[i], err = NewPredicate[T](filter)
		if err != nil {
			return nil, err
		}
	}
	pred := func(obj *T) bool {
		for _, pred := range preds {
			if pred(obj) {
				return true
			}
		}
		return false
	}
	return pred, nil
}

var (
	errInvalidOperator = errors.New("invalid operator")
)

func orderedPredicate[T any, F cmp.Ordered](field string, op string, value F) (func(*T) bool, error) {
	f, err := FieldSelector[T, F](field)
	if err != nil {
		return nil, err
	}
	var pred func(*T) bool
	switch op {
	case ">":
		pred = func(obj *T) bool { return f(obj) > value }
	case ">=":
		pred = func(obj *T) bool { return f(obj) >= value }
	case "=":
		pred = func(obj *T) bool { return f(obj) == value }
	case "!=":
		pred = func(obj *T) bool { return f(obj) != value }
	case "<=":
		pred = func(obj *T) bool { return f(obj) <= value }
	case "<":
		pred = func(obj *T) bool { return f(obj) < value }
	default:
		err = fmt.Errorf("%w: %q not supported with type %s",
			errInvalidOperator, op, reflect.TypeOf(value).Name(),
		)
	}
	return pred, err
}

func comparablePredicate[T any, F comparable](field string, op string, value F) (func(*T) bool, error) {
	f, err := FieldSelector[T, F](field)
	if err != nil {
		return nil, err
	}
	var pred func(*T) bool
	switch op {
	case "=":
		pred = func(obj *T) bool { return f(obj) == value }
	case "!=":
		pred = func(obj *T) bool { return f(obj) != value }
	default:
		err = fmt.Errorf("%w: %q not supported with type %s",
			errInvalidOperator, op, reflect.TypeOf(value).Name())
	}
	return pred, err
}

func pointerPredicate[T any](field string, op string) (func(*T) bool, error) {
	var zero T
	tp := reflect.TypeOf(zero)
	if tp.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w: %s", errNotAStruct, tp.Name())
	}
	f, ok := reflect.TypeOf(zero).FieldByName(field)
	if !ok {
		return nil, fmt.Errorf("%w: %s", errNoSuchField, field)
	}
	if f.Type.Kind() != reflect.Pointer {
		return nil, fmt.Errorf("%w: %s is not a pointer", errTypeMismatch, f.Type.Name())
	}
	fieldIndex := f.Index

	var pred func(*T) bool
	var err error
	switch op {
	case "=":
		pred = func(obj *T) bool {
			return reflect.ValueOf(obj).Elem().FieldByIndex(fieldIndex).IsNil()
		}
	case "!=":
		pred = func(obj *T) bool {
			return !reflect.ValueOf(obj).Elem().FieldByIndex(fieldIndex).IsNil()
		}
	default:
		err = fmt.Errorf("%w: %q not supported with pointers",
			errInvalidOperator, op,
		)
	}
	return pred, err
}

func timePredicate[T any](field string, op string, value time.Time) (func(*T) bool, error) {
	f, err := FieldSelector[T, time.Time](field)
	if err != nil {
		return nil, err
	}
	var pred func(*T) bool
	switch op {
	case ">":
		pred = func(obj *T) bool { return f(obj).After(value) }
	case ">=":
		pred = func(obj *T) bool {
			t := f(obj)
			return t.Equal(value) || t.After(value)
		}
	case "=":
		pred = func(obj *T) bool { return f(obj).Equal(value) }
	case "!=":
		pred = func(obj *T) bool { return !f(obj).Equal(value) }
	case "<=":
		pred = func(obj *T) bool {
			t := f(obj)
			return t.Equal(value) || t.Before(value)
		}
	case "<":
		pred = func(obj *T) bool { return f(obj).Before(value) }
	default:
		err = fmt.Errorf("%w: %q not supported with type %s",
			errInvalidOperator, op, reflect.TypeOf(value).Name(),
		)
	}
	return pred, err
}

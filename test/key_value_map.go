package test

import (
	"errors"
	"testing"

	//!lint ignore ST10001 common definitions
	. "github.com/ArnaudCalmettes/store"
	"github.com/google/uuid"
)

func TestKeyValueMap(t *testing.T, newKeyValueMap func() KeyValueMap) {
	type TestFunc func(*testing.T, func() KeyValueMap)
	run := func(t *testing.T, name string, testFunc TestFunc) {
		t.Run(name, func(t *testing.T) {
			testFunc(t, newKeyValueMap)
		})
	}
	t.Parallel()
	run(t, "GetSetOne", testKeyValueMapGetSetOne)
	run(t, "GetSetMany", testKeyValueMapGetSetMany)
	run(t, "GetAll", testKeyValueMapGetAll)
	run(t, "UpdateOne", testKeyValueMapUpdateOne)
	run(t, "UpdateMany", testKeyValueMapUpdateMany)
	run(t, "Delete", testKeyValueMapDelete)
}

func testKeyValueMapGetSetOne(t *testing.T, newKeyValueMap func() KeyValueMap) {
	store := newKeyValueMap()
	ctx, cancel := NewTestContext()
	defer cancel()
	t.Run("get not found", func(t *testing.T) {
		result, err := store.GetOne(ctx, "does not exist")
		Expect(t,
			IsZero(result),
			IsError(ErrNotFound, err),
		)
	})
	t.Run("nominal", func(t *testing.T) {
		key := uuid.NewString()
		err := store.SetOne(ctx, key, "value")
		Require(t,
			NoErrorf(err, "setting up fixture"),
		)

		result, err := store.GetOne(ctx, key)
		Expect(t,
			Equal("value", result),
			NoError(err),
		)
	})
	t.Run("overwrite", func(t *testing.T) {
		key := uuid.NewString()
		err := store.SetOne(ctx, key, "initial value")
		Require(t,
			NoErrorf(err, "setting up initial value"),
		)

		err = store.SetOne(ctx, key, "updated value")
		Require(t,
			NoError(err),
		)

		check, err := store.GetOne(ctx, key)
		Require(t,
			NoError(err),
			Equalf("updated value", check, "value should have been updated"),
		)
	})
}

func testKeyValueMapGetSetMany(t *testing.T, newKeyValueMap func() KeyValueMap) {
	store := newKeyValueMap()
	ctx, cancel := NewTestContext()
	defer cancel()

	t.Run("get many empty", func(t *testing.T) {
		result, err := store.GetMany(ctx, []string{})
		Expect(t,
			NoError(err),
			Equal(map[string]string{}, result),
		)
	})
	t.Run("set many empty", func(t *testing.T) {
		err := store.SetMany(ctx, map[string]string{})
		Expect(t,
			NoError(err),
		)
	})
	t.Run("nominal", func(t *testing.T) {
		err := store.SetMany(ctx, map[string]string{
			"one":   "one",
			"two":   "two",
			"three": "three",
		})
		Require(t,
			NoError(err),
		)

		result, err := store.GetMany(ctx, []string{"one", "two", "three", "four"})
		Expect(t,
			NoError(err),
			Equal(
				map[string]string{
					"one":   "one",
					"two":   "two",
					"three": "three",
				},
				result,
			),
		)
	})

	t.Run("overwrite", func(t *testing.T) {
		err := store.SetMany(ctx, map[string]string{
			"one": "ONE",
			"two": "TWO",
		})
		Require(t,
			NoError(err),
		)

		result, err := store.GetMany(ctx, []string{"one", "two", "three", "four"})
		Require(t,
			NoError(err),
			Equal(
				map[string]string{
					"one":   "ONE",
					"two":   "TWO",
					"three": "three",
				},
				result,
			),
		)
	})
}

func testKeyValueMapGetAll(t *testing.T, newKeyValueMap func() KeyValueMap) {
	store := newKeyValueMap()
	ctx, cancel := NewTestContext()
	defer cancel()

	t.Run("empty", func(t *testing.T) {
		all, err := store.GetAll(ctx)
		Expect(t,
			NoError(err),
			Equal(map[string]string{}, all),
		)
	})
	t.Run("nominal", func(t *testing.T) {
		store.SetMany(ctx, map[string]string{
			"key-one": "value-one",
			"key-two": "value-two",
		})
		all, err := store.GetAll(ctx)
		Expect(t,
			NoError(err),
			Equal(
				map[string]string{
					"key-one": "value-one",
					"key-two": "value-two",
				},
				all,
			),
		)
	})
}

func testKeyValueMapUpdateOne(t *testing.T, newKeyValueMap func() KeyValueMap) {
	store := newKeyValueMap()
	ctx, cancel := NewTestContext()
	defer cancel()

	t.Run("does not exist", func(t *testing.T) {
		errCalled := errors.New("called")
		updateFunc := func(_ string, value *string) (*string, error) {
			Require(t,
				IsNilPointerf(value, "non existent values should be passed as a nil pointer"),
			)
			return nil, errCalled
		}
		err := store.UpdateOne(ctx, uuid.NewString(), updateFunc)
		Expect(t,
			IsError(errCalled, err),
		)
	})
	t.Run("nominal", func(t *testing.T) {
		id := uuid.NewString()
		err := store.SetOne(ctx, id, "initial value")
		Require(t,
			NoError(err),
		)

		updateFunc := func(key string, value *string) (*string, error) {
			Require(t,
				Equal(id, key),
				Equal(pointerTo("initial value"), value),
			)
			return pointerTo("updated value"), nil
		}
		err = store.UpdateOne(ctx, id, updateFunc)
		Require(t,
			NoError(err),
		)

		value, err := store.GetOne(ctx, id)
		Expect(t,
			NoError(err),
			Equal("updated value", value),
		)
	})
	t.Run("callback returns nil", func(t *testing.T) {
		id := uuid.NewString()
		err := store.SetOne(ctx, id, "initial value")
		Require(t,
			NoError(err),
		)
		var called bool
		updateFunc := func(_ string, value *string) (*string, error) {
			called = true
			*value = "new value"
			return nil, nil
		}
		err = store.UpdateOne(ctx, id, updateFunc)
		Expect(t,
			NoError(err),
			Equal(true, called),
		)

		value, err := store.GetOne(ctx, id)
		Expect(t,
			NoError(err),
			Equal("initial value", value),
		)
	})
}

func testKeyValueMapUpdateMany(t *testing.T, newKeyValueMap func() KeyValueMap) {
	store := newKeyValueMap()
	ctx, cancel := NewTestContext()
	defer cancel()

	t.Run("empty keys", func(t *testing.T) {
		err := store.UpdateMany(ctx, []string{}, nil)
		Expect(t,
			NoError(err),
		)
	})
	t.Run("key does not exist", func(t *testing.T) {
		id := uuid.NewString()
		var called bool
		updateFunc := func(key string, value *string) (*string, error) {
			called = true
			Require(t,
				Equal(id, key),
				IsNilPointer(value),
			)
			return pointerTo("value"), nil
		}
		err := store.UpdateMany(ctx, []string{id}, updateFunc)
		Expect(t,
			NoError(err),
			Equal(true, called),
		)

		value, err := store.GetOne(ctx, id)
		Expect(t,
			NoError(err),
			Equal("value", value),
		)
	})
	t.Run("nominal", func(t *testing.T) {
		id := uuid.NewString()
		err := store.SetOne(ctx, id, "initial value")
		Require(t,
			NoError(err),
		)

		updateFunc := func(key string, value *string) (*string, error) {
			Require(t,
				Equal(id, key),
				Equal(pointerTo("initial value"), value),
			)
			return pointerTo("updated value"), nil
		}
		err = store.UpdateMany(ctx, []string{id}, updateFunc)
		Require(t,
			NoError(err),
		)

		value, err := store.GetOne(ctx, id)
		Expect(t,
			NoError(err),
			Equal("updated value", value),
		)
	})
	t.Run("callback returns error", func(t *testing.T) {
		id := uuid.NewString()
		err := store.SetOne(ctx, id, "initial value")
		Require(t,
			NoError(err),
		)

		errUpdate := errors.New("update error")
		updateFunc := func(string, *string) (*string, error) {
			return pointerTo("updated value"), errUpdate
		}
		err = store.UpdateMany(ctx, []string{id}, updateFunc)
		Require(t,
			IsError(errUpdate, err),
		)

		value, err := store.GetOne(ctx, id)
		Expect(t,
			NoError(err),
			Equal("initial value", value),
		)
	})
	t.Run("callback returns nil", func(t *testing.T) {
		id := uuid.NewString()
		err := store.SetOne(ctx, id, "initial value")
		Require(t,
			NoError(err),
		)

		updateFunc := func(_ string, value *string) (*string, error) {
			*value = "new value"
			return nil, nil
		}
		err = store.UpdateMany(ctx, []string{id}, updateFunc)
		Require(t,
			NoError(err),
		)

		value, err := store.GetOne(ctx, id)
		Expect(t,
			NoError(err),
			Equal("initial value", value),
		)
	})
}

func testKeyValueMapDelete(t *testing.T, newKeyValueMap func() KeyValueMap) {
	store := newKeyValueMap()
	ctx, cancel := NewTestContext()
	defer cancel()

	t.Run("empty keys", func(t *testing.T) {
		err := store.Delete(ctx)
		Expect(t,
			NoError(err),
		)
	})
	t.Run("nominal", func(t *testing.T) {
		err := store.SetMany(ctx, map[string]string{
			"one": "one",
			"two": "two",
		})
		Require(t,
			NoError(err),
		)

		err = store.Delete(ctx, "three", "two")
		Expect(t,
			NoError(err),
		)

		all, err := store.GetMany(ctx, []string{"one", "two", "three"})
		Expect(t,
			NoError(err),
			Equal(map[string]string{"one": "one"}, all),
		)
	})
}

func pointerTo[T any](obj T) *T {
	return &obj
}

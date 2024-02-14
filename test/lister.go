package test

import (
	"testing"

	//lint:ignore ST1001 shared definitions
	. "github.com/ArnaudCalmettes/store"
	//lint:ignore ST1001 test vocabulary
	. "github.com/ArnaudCalmettes/store/test/helpers"
)

type TestListerInterface[T any] interface {
	BaseKeyValueStore[T]
	Lister[T]
}

type Person struct {
	ID   string
	Name string
	Age  int
}

type listerConstructor func(*testing.T) TestListerInterface[Person]

func TestLister(t *testing.T, newLister listerConstructor) {
	store := newLister(t)
	ctx, cancel := NewTestContext()
	defer cancel()
	err := store.SetMany(ctx, map[string]*Person{
		"001": {ID: "001", Name: "John Doe", Age: 42},
		"002": {ID: "002", Name: "Willard", Age: 13},
		"003": {ID: "003", Name: "Jane Smith", Age: 20},
	})
	Require(t,
		NoError(err),
	)

	t.Run("filter nominal", func(t *testing.T) {
		result, err := store.List(ctx, Filter(Where("Age", "<", 18)))
		Expect(t,
			NoError(err),
			Equal(
				[]*Person{
					{ID: "002", Name: "Willard", Age: 13},
				},
				result,
			),
		)
	})
}

package store

type Options struct {
	Filter  *FilterSpec
	OrderBy *OrderBySpec
	Limit   int
	Offset  int
}

// Filtering

func Filter(spec *FilterSpec) *Options {
	return &Options{Filter: spec}
}

type FilterSpec struct {
	Where *WhereClause
	All   []*FilterSpec
	Any   []*FilterSpec
}

type WhereClause struct {
	Field string
	Op    string
	Value any
}

func Where(fieldName string, op string, value any) *FilterSpec {
	return &FilterSpec{Where: &WhereClause{fieldName, op, value}}
}

func All(filters ...*FilterSpec) *FilterSpec {
	return &FilterSpec{All: filters}
}

func Any(filters ...*FilterSpec) *FilterSpec {
	return &FilterSpec{Any: filters}
}

// Ordering

func Order(order *OrderBySpec) *Options {
	return &Options{OrderBy: order}
}

func By(field string) *OrderBySpec {
	return &OrderBySpec{Field: field}
}

type OrderBySpec struct {
	Field      string
	Descending bool
}

func (o *OrderBySpec) Desc() *OrderBySpec {
	o.Descending = true
	return o
}

func (o *OrderBySpec) Asc() *OrderBySpec {
	o.Descending = false
	return o
}

// Pagination

func Paginate(limit, offset int) *Options {
	return &Options{Limit: limit, Offset: offset}
}

func Limit(n int) *Options {
	return &Options{Limit: n}
}

func Offset(n int) *Options {
	return &Options{Offset: n}
}

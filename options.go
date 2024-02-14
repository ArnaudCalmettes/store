package store

type Options struct {
	Filter *FilterSpec
}

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

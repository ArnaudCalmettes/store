package store

type Filter struct {
	Where *WhereClause
	All   []*Filter
	Any   []*Filter
}

type WhereClause struct {
	Field string
	Op    string
	Value any
}

func Where(fieldName string, op string, value any) *Filter {
	return &Filter{Where: &WhereClause{fieldName, op, value}}
}

func All(filters ...*Filter) *Filter {
	return &Filter{All: filters}
}

func Any(filters ...*Filter) *Filter {
	return &Filter{Any: filters}
}

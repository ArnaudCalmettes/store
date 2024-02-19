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

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

package libbun

import (
	"errors"
	"fmt"

	"github.com/ArnaudCalmettes/store"
	"github.com/uptrace/bun"
)

func BuilderForFilter(filter *store.FilterSpec, spec *TableSpec) (
	func(bun.QueryBuilder) bun.QueryBuilder,
	error,
) {
	if err := validateFilter(filter, spec); err != nil {
		return nil, err
	}
	builder := func(qb bun.QueryBuilder) bun.QueryBuilder {
		return applyFilter(qb, filter, spec)
	}
	return builder, nil
}

func applyFilter(qb bun.QueryBuilder, filter *store.FilterSpec, spec *TableSpec) bun.QueryBuilder {
	switch {
	case filter.Where != nil:
		w := filter.Where
		qb = qb.Where("? "+w.Op+" ?", bun.Ident(spec.ColumnNames[w.Field]), w.Value)
	case filter.All != nil:
		for _, sub := range filter.All {
			qb = applyFilter(qb, sub, spec)
		}
	case filter.Any != nil:
		for _, sub := range filter.Any {
			qb.WhereGroup(" OR ", func(q bun.QueryBuilder) bun.QueryBuilder {
				return applyFilter(qb, sub, spec)
			})
		}
	}
	return qb
}

var (
	errNoSuchField = errors.New("no such field")
)

func validateFilter(filter *store.FilterSpec, spec *TableSpec) error {
	switch {
	case filter.Where != nil:
		field := filter.Where.Field
		if _, ok := spec.ColumnNames[field]; !ok {
			return fmt.Errorf("%w: %s", errNoSuchField, field)
		}
	case filter.All != nil:
		errs := make([]error, len(filter.All))
		for i, sub := range filter.All {
			errs[i] = validateFilter(sub, spec)
		}
		return errors.Join(errs...)
	case filter.Any != nil:
		errs := make([]error, len(filter.Any))
		for i, sub := range filter.Any {
			errs[i] = validateFilter(sub, spec)
		}
		return errors.Join(errs...)
	}
	return nil
}

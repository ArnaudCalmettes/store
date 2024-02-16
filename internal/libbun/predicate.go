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

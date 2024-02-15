package options

import (
	"errors"

	"github.com/ArnaudCalmettes/store"
)

var (
	ErrMultipleOrderBy = errors.New("cannot have multiple OrderBy clauses")
)

func Merge(opts ...*store.Options) (*store.Options, error) {
	options := store.Options{}
	for _, opt := range opts {
		if opt.Filter != nil {
			if options.Filter == nil {
				options.Filter = opt.Filter
			} else if options.Filter.All != nil {
				options.Filter.All = append(options.Filter.All, opt.Filter)
			} else {
				options.Filter = store.All(options.Filter, opt.Filter)
			}
		}
		if opt.OrderBy != nil {
			if options.OrderBy != nil {
				return nil, ErrMultipleOrderBy
			}
			options.OrderBy = opt.OrderBy
		}
	}
	return &options, nil
}

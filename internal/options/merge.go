package options

import (
	"errors"
	"fmt"

	"github.com/ArnaudCalmettes/store"
)

var (
	ErrDuplicateOption = errors.New("duplicate option")
)

func duplicateOption(name string) error {
	return fmt.Errorf("%w: %s", ErrDuplicateOption, name)
}

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
				return nil, duplicateOption("OrderBy")
			}
			options.OrderBy = opt.OrderBy
		}
		if opt.Limit != 0 {
			if options.Limit != 0 {
				return nil, duplicateOption("Page")
			}
			options.Limit = opt.Limit
		}
		if opt.Offset != 0 {
			if options.Offset != 0 {
				return nil, duplicateOption("Offset")
			}
			options.Offset = opt.Offset
		}
	}
	return &options, nil
}

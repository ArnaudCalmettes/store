package options

import "github.com/ArnaudCalmettes/store"

func Merge(opts ...*store.Options) *store.Options {
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
	}
	return &options
}

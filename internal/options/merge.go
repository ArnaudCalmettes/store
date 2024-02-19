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

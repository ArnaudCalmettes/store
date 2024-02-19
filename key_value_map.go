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

import (
	"context"
)

type BaseKeyValueMap interface {
	GetOne(ctx context.Context, key string) (string, error)
	GetMany(ctx context.Context, keys []string) (map[string]string, error)
	GetAll(ctx context.Context) (map[string]string, error)
	SetOne(ctx context.Context, key, value string) error
	SetMany(ctx context.Context, items map[string]string) error
	UpdateOne(ctx context.Context, key string, update UpdateFunc[string]) error
	UpdateMany(ctx context.Context, keys []string, update UpdateFunc[string]) error
	Delete(ctx context.Context, keys ...string) error
}

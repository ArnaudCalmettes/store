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

package serializer

import (
	"encoding/json"

	"github.com/ArnaudCalmettes/store"
	"github.com/ArnaudCalmettes/store/internal/zerocopy"
)

func NewJSON[T any]() store.Serializer[T] {
	return jsonSerializer[T]{}
}

type jsonSerializer[T any] struct{}

func (j jsonSerializer[T]) Serialize(obj *T) (string, error) {
	if obj == nil {
		return "", ErrNilObject
	}
	data, err := json.Marshal(obj)
	return zerocopy.BytesToString(data), err
}

func (j jsonSerializer[T]) Deserialize(data string) (*T, error) {
	if len(data) == 0 {
		return nil, ErrEmptyData
	}
	var obj T
	err := json.Unmarshal(zerocopy.StringToBytes(data), &obj)
	return &obj, err
}

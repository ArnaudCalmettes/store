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
	"testing"

	"github.com/ArnaudCalmettes/store/internal/zerocopy"
	. "github.com/ArnaudCalmettes/store/test/helpers"
)

func TestSerializeJSON(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		s := NewJSON[any]()
		data, err := s.Serialize(nil)
		Expect(t,
			IsError(ErrNilObject, err),
			IsZero(data),
		)
	})
	t.Run("int", func(t *testing.T) {
		s := NewJSON[int]()
		input := 42
		data, err := s.Serialize(&input)
		Expect(t,
			NoError(err),
			IsNotZero(data),
		)

		var result int
		err = json.Unmarshal(zerocopy.StringToBytes(data), &result)
		Expect(t,
			NoError(err),
			Equal(input, result),
		)
	})
}

func TestDeserializeJSON(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		s := NewJSON[any]()
		result, err := s.Deserialize("")
		Expect(t,
			IsError(ErrEmptyData, err),
			IsNilPointer(result),
		)
	})
	t.Run("int", func(t *testing.T) {
		s := NewJSON[int]()
		input := 42
		data, err := json.Marshal(42)
		Expect(t,
			NoError(err),
			IsNotZero(data),
		)

		result, err := s.Deserialize(zerocopy.BytesToString(data))
		Expect(t,
			NoError(err),
			Equal(&input, result),
		)
	})
}

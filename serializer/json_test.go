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

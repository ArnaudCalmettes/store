package serializer

import (
	"testing"

	"github.com/ArnaudCalmettes/store/internal/zerocopy"
	. "github.com/ArnaudCalmettes/store/test/helpers"
	"github.com/vmihailenco/msgpack"
)

func TestSerializeMsgpack(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		s := NewMsgpack[any]()
		data, err := s.Serialize(nil)
		Expect(t,
			IsError(ErrNilObject, err),
			IsZero(data),
		)
	})
	t.Run("int", func(t *testing.T) {
		s := NewMsgpack[int]()
		input := 42
		data, err := s.Serialize(&input)
		Expect(t,
			NoError(err),
			IsNotZero(data),
		)

		var result int
		err = msgpack.Unmarshal(zerocopy.StringToBytes(data), &result)
		Expect(t,
			NoError(err),
			Equal(input, result),
		)
	})
}

func TestDeserializeMsgpack(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		s := NewMsgpack[any]()
		result, err := s.Deserialize("")
		Expect(t,
			IsError(ErrEmptyData, err),
			IsNilPointer(result),
		)
	})
	t.Run("int", func(t *testing.T) {
		s := NewMsgpack[int]()
		input := 42
		data, err := msgpack.Marshal(42)
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

package serializer

import (
	"github.com/ArnaudCalmettes/store"
	"github.com/ArnaudCalmettes/store/internal/zerocopy"
	"github.com/vmihailenco/msgpack"
)

func NewMsgpack[T any]() store.Serializer[T] {
	return msgpackSerializer[T]{}
}

type msgpackSerializer[T any] struct{}

func (j msgpackSerializer[T]) Serialize(obj *T) (string, error) {
	if obj == nil {
		return "", ErrNilObject
	}
	data, err := msgpack.Marshal(obj)
	return zerocopy.BytesToString(data), err
}

func (j msgpackSerializer[T]) Deserialize(data string) (*T, error) {
	if len(data) == 0 {
		return nil, ErrEmptyData
	}
	var obj T
	err := msgpack.Unmarshal(zerocopy.StringToBytes(data), &obj)
	return &obj, err
}

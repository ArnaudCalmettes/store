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

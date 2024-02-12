package serializer

import "errors"

var (
	ErrNilObject = errors.New("cannot serialze nil object")
	ErrEmptyData = errors.New("cannot deserialize empty data")
)

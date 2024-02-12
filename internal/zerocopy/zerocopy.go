package zerocopy

import "unsafe"

// StringToBytes converts a string to a []byte in a zero-copy operation.
func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}

// BytesToString converts a []byte to a string in a zero-copy operation.
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

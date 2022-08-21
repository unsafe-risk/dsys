package encoding2

import (
	"errors"
	"io"
	"reflect"
	"unsafe"

	"v8.run/go/exp/util/noescape"
)

func Bytes(w io.Writer, b []byte) (int, error) {
	var nn int
	var err error
	if uint64(len(b)) > MaxSize {
		return 0, ErrBytesTooLarge
	}
	nn, err = Uint(w, uint64(len(b)))
	if err != nil {
		return nn, err
	}
	n, err := noescape.Write(w, b)
	return nn + n, err
}

func String(w io.Writer, s string) (int, error) {
	var nn int
	var err error
	if uint64(len(s)) > MaxSize {
		return 0, ErrBytesTooLarge
	}
	nn, err = Uint(w, uint64(len(s)))
	if err != nil {
		return nn, err
	}
	sh := reflect.SliceHeader{Data: (*reflect.StringHeader)(unsafe.Pointer(&s)).Data, Len: len(s), Cap: len(s)}
	n, err := noescape.Write(w, *(*[]byte)(unsafe.Pointer(&sh)))
	return nn + n, err
}

// 16MB
var MaxSize = uint64(1 << 24)

var ErrBytesTooLarge = errors.New("bytes too large")

func LoadBytes(r io.Reader) ([]byte, error) {
	size, _, err := LoadUint(r)
	if err != nil {
		return nil, err
	}
	if size > MaxSize {
		return nil, ErrBytesTooLarge
	}
	b := make([]byte, size, size)
READ:
	n, err := noescape.Read(r, b)
	if err != nil {
		return nil, err
	}
	if n < len(b) {
		b = b[:n]
		goto READ
	}
	return b[:cap(b)], nil
}

func LoadString(r io.Reader) (string, error) {
	b, err := LoadBytes(r)
	return *(*string)(unsafe.Pointer(&b)), err
}

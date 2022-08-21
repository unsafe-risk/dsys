package encoding2

import (
	"encoding/binary"
	"errors"
	"io"

	"v8.run/go/exp/util/noescape"
)

func Uint(w io.Writer, v uint64) (n int, err error) {
	var b [9]byte
	// b[0]: length of the varint
	binary.LittleEndian.PutUint64(b[1:], v)
	switch {
	case v <= 1<<8-1:
		b[0] = 1
	case v <= 1<<16-1:
		b[0] = 2
	case v <= 1<<24-1:
		b[0] = 3
	case v <= 1<<32-1:
		b[0] = 4
	case v <= 1<<40-1:
		b[0] = 5
	case v <= 1<<48-1:
		b[0] = 6
	case v <= 1<<56-1:
		b[0] = 7
	case v <= 1<<64-1:
		b[0] = 8
	}
	return noescape.Write(w, b[:b[0]+1])
}

func Int(w io.Writer, v int64) (n int, err error) {
	return Uint(w, uint64(v))
}

var ErrOverflow = errors.New("overflow")

func LoadUint(r io.Reader) (v uint64, n int, err error) {
	var b [9]byte
	var nn int
READSIZE:
	n, err = noescape.Read(r, b[:1])
	if err != nil {
		return 0, n, err
	}
	nn += n
	if n != 1 {
		goto READSIZE
	}
	size := int(b[0])
	if size > 8 {
		err = ErrOverflow
		return v, nn, err
	}
	if size == 0 {
		return 0, nn, nil
	}
READDATA:
	rb := b[1 : size+1]
	n, err = noescape.Read(r, rb)
	if err != nil {
		return
	}
	nn += n
	if n != size {
		rb = rb[n:]
		goto READDATA
	}
	v = binary.LittleEndian.Uint64(b[1:])
	return v, nn, nil
}

func LoadInt(r io.Reader) (v int64, n int, err error) {
	vu, n, err := LoadUint(r)
	return int64(vu), n, err
}

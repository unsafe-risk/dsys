package encoding2

import (
	"bytes"
	"testing"
)

func TestUint(t *testing.T) {
	var tests = []uint64{
		0,
		1,
		2,
		125,
		126,
		127,
		128,
		129,
		253,
		254,
		255,
		256,
		257,
		258,
		65534,
		65535,
		65536,
		65537,
		65538,
		4294967294,
		4294967295,
		4294967296,
		4294967297,
		4294967298,
		281474976710654,
		281474976710655,
		281474976710656,
		281474976710657,
		281474976710658,
		72057594037927934,
		72057594037927935,
		72057594037927936,
		72057594037927937,
		72057594037927938,
		9223372036854775805,
		9223372036854775806,
		9223372036854775807,
		9223372036854775808,
		9223372036854775809,
		18446744073709551612,
		18446744073709551613,
		18446744073709551614,
		18446744073709551615,
	}
	var bb bytes.Buffer
	for _, v := range tests {
		n0, err := Uint(&bb, v)
		if err != nil {
			t.Errorf("Uint(%d) error: %v", v, err)
		}
		b := bb.Bytes()
		decoded, n1, err := LoadUint(&bb)
		if err != nil {
			t.Errorf("Uint(%d) error: %v", v, err)
		}
		if n0 != n1 {
			t.Errorf("Uint(%d) n0 != n1 (%d != %d), buf: %v", v, n0, n1, b)
		}
		if decoded != v {
			t.Errorf("Uint(%d) decoded != v (%d != %d), buf: %v", v, decoded, v, b)
		}
		bb.Reset()
	}
}

func TestInt(t *testing.T) {
	var tests = []int64{
		0,
		1,
		2,
		125,
		126,
		127,
		128,
		129,
		253,
		254,
		255,
		256,
		257,
		258,
		-1,
		-2,
		-125,
		-126,
		-127,
		-128,
		-129,
		-253,
		-254,
		-255,
		-256,
		-257,
		-258,
		-9223372036854775805,
		-9223372036854775806,
		-9223372036854775807,
		-9223372036854775808,
	}
	var bb bytes.Buffer
	for _, v := range tests {
		n0, err := Int(&bb, v)
		if err != nil {
			t.Errorf("Int(%d) error: %v", v, err)
		}
		b := bb.Bytes()
		decoded, n1, err := LoadInt(&bb)
		if err != nil {
			t.Errorf("Int(%d) error: %v", v, err)
		}
		if n0 != n1 {
			t.Errorf("Int(%d) n0 != n1 (%d != %d), buf: %v", v, n0, n1, b)
		}
		if decoded != v {
			t.Errorf("Int(%d) decoded != v (%d != %d ), buf: %v", v, decoded, v, b)
		}
		bb.Reset()
	}
}

func BenchmarkUint(b *testing.B) {
	var bb bytes.Buffer
	Uint(&bb, 9223372036854775806)
	bb.Reset()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bb.Reset()
		Uint(&bb, 9223372036854775806)
	}
}

func BenchmarkLoadUint(b *testing.B) {
	var data bytes.Buffer
	Uint(&data, 9223372036854775806)
	v := data.Bytes()
	var decodeBuf bytes.Buffer
	decodeBuf.Reset()
	decodeBuf.Write(v)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decodeBuf.Reset()
		decodeBuf.Write(v)
		LoadUint(&decodeBuf)
	}
}

package encoding2

import (
	"bytes"
	"math/rand"
	"testing"
)

func FuzzBytes(f *testing.F) {
	for i := 1; i < 100; i++ {
		b := make([]byte, i)
		rand.Read(b)
		f.Add(b)
	}

	f.Fuzz(func(t *testing.T, a []byte) {
		var bb bytes.Buffer
		if _, err := Bytes(&bb, a); err != nil {
			t.Fatal(err)
		}
		v, err := LoadBytes(&bb)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(a, v) {
			t.Fatal("not equal")
		}
	})
}

func TestBytes(t *testing.T) {
	b := []byte{1, 2, 3}
	var bb bytes.Buffer
	if _, err := Bytes(&bb, b); err != nil {
		t.Fatal(err)
	}
	v, err := LoadBytes(&bb)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, v) {
		t.Fatal("not equal")
	}
}

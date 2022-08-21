package encryption

import (
	"testing"
)

func TestXChaCha20Poly1305Box(t *testing.T) {
	key := []byte("this is a super secret key")
	box, err := NewXChaCha20Poly1305Box(key)
	if err != nil {
		t.Fatal(err)
	}
	plaintext := []byte("this is a super secret message")
	buffer := make([]byte, box.CalcSize(len(plaintext)))
	encrypted := box.Seal(buffer[:0], plaintext)
	if &buffer[0] != &encrypted[0] {
		t.Fatal("buffer and encrypted are not the same")
	}
	decrypted, err := box.Open(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if string(decrypted) != string(plaintext) {
		t.Fatal("decrypted and plaintext are not the same")
	}
}

func BenchmarkXChaCha20Poly1305BoxSeal(b *testing.B) {
	b.RunParallel(func(p *testing.PB) {
		key := []byte("this is a super secret key")
		box, err := NewXChaCha20Poly1305Box(key)
		if err != nil {
			b.Fatal(err)
		}
		plaintext := []byte("this is a super secret message")
		buffer := make([]byte, box.CalcSize(len(plaintext)))
		for p.Next() {
			box.Seal(buffer[:0], plaintext)
		}
	})
}

func BenchmarkXChaCha20Poly1305BoxOpen(b *testing.B) {
	b.RunParallel(func(p *testing.PB) {
		key := []byte("this is a super secret key")
		box, err := NewXChaCha20Poly1305Box(key)
		if err != nil {
			b.Fatal(err)
		}
		plaintext := []byte("this is a super secret message")
		buffer := make([]byte, box.CalcSize(len(plaintext)))
		encrypted := box.Seal(buffer[:0], plaintext)
		encrypted2 := make([]byte, len(encrypted))
		copy(encrypted2, encrypted)
		for p.Next() {
			copy(encrypted2[:len(encrypted)], encrypted)
			encrypted2 = encrypted2[:len(encrypted)]
			out, err := box.Open(encrypted2)
			if err != nil || string(out) != string(plaintext) {
				panic("failed to open: " + err.Error())
			}
		}
	})
}

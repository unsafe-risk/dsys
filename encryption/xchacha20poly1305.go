package encryption

import (
	"crypto/cipher"
	"errors"

	"github.com/lemon-mint/experiment/util/noescape"
	"github.com/zeebo/blake3"
	"golang.org/x/crypto/chacha20poly1305"
)

type XChaCha20Poly1305Box struct {
	aead cipher.AEAD
}

func NewXChaCha20Poly1305Box(key []byte) (*XChaCha20Poly1305Box, error) {
	key256 := blake3.Sum256(key)
	aead, err := chacha20poly1305.NewX(key256[:])
	if err != nil {
		// Unreachable. The output of blake3.Sum256 is always 32 bytes.
		panic(err)
	}
	return &XChaCha20Poly1305Box{aead}, nil
}

func (ab *XChaCha20Poly1305Box) CalcSize(plainSize int) int {
	return 24 + plainSize + ab.aead.Overhead()
}

func (ab *XChaCha20Poly1305Box) Seal(dst, plaintext []byte) []byte {
	var nonce [24]byte
	b := nonce[:]
	n, err := nonceReader.Read(b)
	if err != nil || n != len(b) {
		// If the nonce reader fails, we can't do anything about it.
		// Fatal Syscall errors are not recoverable.
		panic("xchacha20poly1305: failed to read random nonce")
	}
	dst = append(dst, b...)
	return ab.aead.Seal(dst, noescape.Bytes(&b), plaintext, nil)
}

var ErrInvalidCiphertext = errors.New("xchacha20poly1305: invalid ciphertext")

func (ab *XChaCha20Poly1305Box) Open(dst, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < 24 {
		return nil, ErrInvalidCiphertext
	}
	nonce := ciphertext[:24]
	ciphertext = ciphertext[24:]
	return ab.aead.Open(ciphertext[:0], nonce, ciphertext, nil)
}

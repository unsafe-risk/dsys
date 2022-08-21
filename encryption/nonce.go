package encryption

import (
	"bufio"
	"crypto/rand"
	"sync"

	"v8.run/go/exp/util/noescape"
)

type noncebox struct {
	br *bufio.Reader
	mu sync.Mutex
}

var nonceReader = &noncebox{
	br: bufio.NewReaderSize(rand.Reader, 8192),
}

func (nb *noncebox) Read(p []byte) (int, error) {
	size := len(p)
	nb.mu.Lock()
	var n int
	var err error
RETRY:
	n, err = noescape.Read(nb.br, p)
	if err == nil && n != len(p) {
		p = p[n:]
		goto RETRY
	}
	nb.mu.Unlock()
	return size, err
}

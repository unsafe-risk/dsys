package raft

import "io"

type State interface {
	GetState() any
	SetState(state any)

	GetDiff(dst []byte) []byte
	ApplyDiff(diff ...[]byte)

	GetSnapshot(w io.Writer) error
	SetSnapshot(r io.Reader) error
}

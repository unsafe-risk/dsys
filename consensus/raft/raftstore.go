package raft

import "v8.run/go/dsys/consensus/raft/raftproto"

type RaftStorage interface {
	// Last returns the last index and term.
	Last(e *raftproto.Entry) error
	// First returns the first index and term.
	First(e *raftproto.Entry) error

	// GetRange returns all entries in the range [start, end).
	GetRange(start, end uint64) ([]*raftproto.Entry, error)

	// Append appends a new entry to the storage.
	Append(e *raftproto.Entry) error
}

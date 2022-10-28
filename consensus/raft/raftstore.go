package raft

import "v8.run/go/dsys/consensus/raft/raftproto"

type RaftStorage interface {
	// Last returns the last entry in the log.
	Last(e *raftproto.Entry) error
	// First returns the first entry in the log.
	First(e *raftproto.Entry) error

	// Get returns the entry at a given index.
	Get(index uint64, e *raftproto.Entry) error

	// GetRange returns all entries in the range [start, end).
	GetRange(start, end uint64) ([]*raftproto.Entry, error)

	// Delete deletes the entry at a given index.
	Delete(index uint64) error

	// DeleteRange deletes all entries in the range [start, end).
	DeleteRange(start, end uint64) error

	// Set sets the entry at a given index.
	Set(e *raftproto.Entry) error

	// SetBatch sets a batch of entries.
	SetBatch(entries []*raftproto.Entry) error

	// StoreState stores the current state of the raft node.
	StoreState(state *raftproto.State) error

	// LoadState loads the current state of the raft node.
	LoadState(state *raftproto.State) error
}

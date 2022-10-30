package raft

import (
	"v8.run/go/dsys/consensus/raft/raftproto"
)

type StateMachine interface {
	// Apply applies a command to the state machine.
	Apply(command *raftproto.Entry) error

	// Snapshot returns a snapshot of the state machine.
	Snapshot() (<-chan *raftproto.Snapshot, error)

	// Restore restores the state machine from a snapshot.
	Restore(<-chan *raftproto.Snapshot) error
}

package vsr

import (
	"time"

	"v8.run/go/dsys/consensus/vsr/vrproto"
)

type VSRConfig struct {
	// ViewChangeTimeout is the timeout for view change
	ViewChangeTimeout time.Duration
}

type VOpMode uint32

const (
	// VOpModeNormal is the normal operation mode.
	VOpModeNormal VOpMode = iota

	// VOpModeViewChange is the view change mode.
	VOpModeViewChange

	// VOpModeRecovery is the recovery operation mode.
	VOpModeRecovery
)

type LogEntry struct {
}

type VSRState struct {
	// Nodes is the sorted list of nodes in the network
	Nodes []*vrproto.Node

	// ReplicaID is the ID of the replica
	ReplicaID uint64

	// Mode is the operation mode of the VSR.
	Mode VOpMode

	// View is the current view number.
	View uint64

	// OpNum is the current operation number.
	OpNum uint64

	// CommitNum is the current commit number.
	CommitNum uint64

	// Logs is the log entries.
	Logs []LogEntry

	// ClientTable
	ClientTable map[uint64]uint64
}

func (s *VSRState) Tick() {

}

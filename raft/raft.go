package raft

import (
	"sync"
	"sync/atomic"
	"time"
)

type NodeState uint32

const (
	NodeStateFollower NodeState = iota
	NodeStateCandidate
	NodeStateLeader
	NodeStateShutdown
)

type Config struct {
	ElectionTimeout  time.Duration
	HeartbeatTimeout time.Duration
	MaxLogEntries    int
}

type Raft struct {
	state  NodeState
	config Config

	mu sync.Mutex
}

func NewRaft(config Config) *Raft {
	return &Raft{
		state:  NodeStateFollower,
		config: config,
	}
}

func (r *Raft) State() NodeState {
	return NodeState(atomic.LoadUint32((*uint32)(&r.state)))
}

func (r *Raft) setState(state NodeState) {
	atomic.StoreUint32((*uint32)(&r.state), uint32(state))
}

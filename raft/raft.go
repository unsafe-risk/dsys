package raft

import (
	"net"
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

type Node struct {
	mu sync.Mutex
	c  net.Conn
}

type Raft struct {
	config Config

	// Persistent state
	currentTerm uint64
	votedFor    string
	state       State
	log         [][]byte

	// Volatile state
	role        NodeState
	commitIndex uint64
	lastApplied uint64

	// Leader state (Reinitialized after election)
	nextIndex  map[string]uint64
	matchIndex map[string]uint64

	// Nodes
	nodes map[string]*Node

	// Channels
	electionTimer  *time.Timer
	heartbeatTimer *time.Timer

	mu sync.Mutex
}

func NewRaft(config Config) *Raft {
	return &Raft{
		role:   NodeStateFollower,
		config: config,
	}
}

func (r *Raft) State() NodeState {
	return NodeState(atomic.LoadUint32((*uint32)(&r.role)))
}

func (r *Raft) setState(state NodeState) {
	atomic.StoreUint32((*uint32)(&r.role), uint32(state))
}

func (r *Raft) CurrentTerm() uint64 {
	return atomic.LoadUint64(&r.currentTerm)
}

/*
const currentPersistentVersion = uint64(1000)

func (r *Raft) DumpPersistentState(w io.Writer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, err := encoding2.Uint(w, currentPersistentVersion)
	if err != nil {
		return err
	}

	// Persistent state

	// Write currentTerm
	_, err = encoding2.Uint(w, r.currentTerm)
	if err != nil {
		return err
	}

	// Write votedFor
	_, err = encoding2.String(w, r.votedFor)
	if err != nil {
		return err
	}

	// Write logs
	// len(log)
	_, err = encoding2.Uint(w, uint64(len(r.log)))
	if err != nil {
		return err
	}
	for i := range r.log {
		encoding2.Bytes(w, r.log[i])
	}

	// Write state
	err = r.state.GetSnapshot(w)
	if err != nil {
		return err
	}
	return nil
}

var ErrInvalidPersistentState = errors.New("invalid persistent state")

func (r *Raft) LoadPersistentState(reader io.Reader) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	version, _, err := encoding2.LoadUint(reader)
	if err != nil {
		return err
	}
	if version != currentPersistentVersion {
		return ErrInvalidPersistentState
	}

	// Persistent state

	// Read currentTerm
	currentTerm, _, err := encoding2.LoadUint(reader)
	if err != nil {
		return err
	}

	// Read votedFor
	votedFor, err := encoding2.LoadString(reader)
	if err != nil {
		return err
	}

	// Read logs
	// len(log)
	lenLog, _, err := encoding2.LoadUint(reader)
	if err != nil {
		return err
	}
	log := make([][]byte, lenLog)
	for i := range log {
		log[i], err = encoding2.LoadBytes(reader)
		if err != nil {
			return err
		}
	}

	// Read state
	err = r.state.SetSnapshot(reader)
	if err != nil {
		return err
	}

	r.currentTerm = currentTerm
	r.votedFor = votedFor
	r.log = log
	return nil
}
*/

func (r *Raft) RequestVote(term uint64, candidateID string, lastLogIndex uint64, lastLogTerm uint64) (currentTerm uint64, voteGranted bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if term < r.CurrentTerm() {
		return r.CurrentTerm(), false
	}

	if r.votedFor != "" {
		return r.CurrentTerm(), false
	}

	if r.State() != NodeStateFollower {
		return r.CurrentTerm(), false
	}

	r.currentTerm = term
	r.votedFor = candidateID
	r.heartbeatTimer.Reset(r.config.HeartbeatTimeout)
	r.setState(NodeStateFollower)
	return r.CurrentTerm(), true
}

func (r *Raft) Heartbeat(term uint64, leaderID string) (currentTerm uint64, success bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if term < r.CurrentTerm() {
		return r.CurrentTerm(), false
	}

	if r.role != NodeStateFollower {
		return r.CurrentTerm(), false
	}

	r.currentTerm = term
	r.heartbeatTimer.Reset(r.config.HeartbeatTimeout)
	r.setState(NodeStateFollower)
	return r.CurrentTerm(), true
}

func (r *Raft) Start() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.setState(NodeStateFollower)
	r.heartbeatTimer = time.NewTimer(r.config.HeartbeatTimeout)
	r.electionTimer = time.NewTimer(r.config.ElectionTimeout)
	r.heartbeatTimer.Reset(r.config.HeartbeatTimeout)
	r.electionTimer.Reset(r.config.ElectionTimeout)

	go func() {
		select {
		case <-r.heartbeatTimer.C:
			r.mu.Lock()
			r.setState(NodeStateCandidate)
			r.mu.Unlock()
		}
	}()
}

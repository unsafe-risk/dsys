package raft

import (
	"errors"
	"sync"
	"sync/atomic"

	"v8.run/go/dsys/consensus/raft/raftproto"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=Role
type Role uint32

const (
	RoleFollower Role = iota
	RoleCandidate
	RoleLeader
)

type Group struct {
	GroupID uint64
	NodeID  uint64

	mu     sync.Mutex
	Config Config
	Role   Role
	State  raftproto.State

	CommitIndex uint64
	LastApplied uint64
	NextIndex   map[uint64]uint64
	MatchIndex  map[uint64]uint64

	// Tick Counters (decremented on every tick).
	HeartbeatTimeout uint64
	ElectionTimeout  uint64

	Storage      RaftStorage
	StateMachine StateMachine
}

func (g *Group) ResetHeartbeatTimeout() {
	atomic.StoreUint64(&g.HeartbeatTimeout, g.Config.HeartbeatTimeoutTicks)
}

func (g *Group) ResetElectionTimeout() {
	atomic.StoreUint64(&g.ElectionTimeout, g.Config.ElectionTimeoutTicks)
}

var ErrInvalidConsensusGroup = errors.New("invalid consensus group")

func (g *Group) HandleAppendEntries(a *raftproto.AppendEntriesRequest) (*raftproto.AppendEntriesResponse, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Term check.
	if a.Term < g.State.Term {
		return &raftproto.AppendEntriesResponse{
			ConsensusGroup: g.GroupID,
			Term:           g.State.Term,
			Success:        false,
		}, nil
	}

	e := raftproto.EntryFromVTPool()
	defer e.ReturnToVTPool()

	// Check that the previous log entry matches.
	err := g.Storage.Get(a.PrevLogIndex, e)
	if err != nil {
		return nil, err
	}

	if e.Term != a.PrevLogTerm {
		return &raftproto.AppendEntriesResponse{
			ConsensusGroup: g.GroupID,
			Term:           g.State.Term,
			Success:        false,
		}, nil
	}

	if len(a.Entries) > 0 {
		err = g.Storage.SetBatch(a.Entries)
		if err != nil {
			return nil, err
		}

		if a.LeaderCommit > g.CommitIndex {
			g.CommitIndex = min(g.CommitIndex, a.Entries[len(a.Entries)-1].Index)
		}
	}

	return &raftproto.AppendEntriesResponse{
		ConsensusGroup: g.GroupID,
		Term:           g.State.Term,
		Success:        true,
	}, nil
}

func (g *Group) HandleRequestVote(a *raftproto.RequestVoteRequest) (*raftproto.RequestVoteResponse, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Term check.
	if a.Term < g.State.Term {
		return &raftproto.RequestVoteResponse{
			ConsensusGroup: g.GroupID,
			Term:           g.State.Term,
			VoteGranted:    false,
		}, nil
	}

	// Check If Vote Already Granted.
	if g.State.VotedTerm == g.State.Term && g.State.VotedFor != a.CandidateID {
		return &raftproto.RequestVoteResponse{
			ConsensusGroup: g.GroupID,
			Term:           g.State.Term,
			VoteGranted:    false,
		}, nil
	}

	// Check If Candidate Log Is At Least As Up-To-Date As Receiver's Log.
	e := raftproto.EntryFromVTPool()
	defer e.ReturnToVTPool()

	err := g.Storage.Last(e)
	if err != nil {
		return nil, err
	}

	if e.Term > a.LastLogTerm || (e.Term == a.LastLogTerm && e.Index > a.LastLogIndex) {
		return &raftproto.RequestVoteResponse{
			ConsensusGroup: g.GroupID,
			Term:           g.State.Term,
			VoteGranted:    false,
		}, nil
	}

	// Grant Vote.
	g.State.VotedFor = a.CandidateID
	g.State.VotedTerm = a.Term
	err = g.Storage.StoreState(&g.State)
	if err != nil {
		return nil, err
	}

	return &raftproto.RequestVoteResponse{
		ConsensusGroup: g.GroupID,
		Term:           g.State.Term,
		VoteGranted:    true,
	}, nil
}

const minusOne = ^uint64(0)

// IsVoted requires the lock to be held.
func (g *Group) IsVoted() bool {
	return g.State.VotedTerm == g.State.Term && g.State.VotedFor != minusOne
}

func (g *Group) ConvertRole(term uint64, role Role) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.State.Term = term
	g.State.VotedFor = minusOne
	g.State.VotedTerm = minusOne
	g.Role = role

}

func (g *Group) Tick() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.CommitIndex > g.LastApplied {
		entries, err := g.Storage.GetRange(g.LastApplied+1, g.CommitIndex+1)
		if err != nil {
			return err
		}

		for i := range entries {
			err = g.StateMachine.Apply(entries[i])
			if err != nil {
				return err
			}
		}
	}

	switch g.Role {
	case RoleFollower:
		ht := atomic.AddUint64(&g.HeartbeatTimeout, minusOne)
		if ht <= 0 && !g.IsVoted() {
			// TODO: Convert to candidate.
		}
	case RoleLeader:
	case RoleCandidate:
	}

	return nil
}

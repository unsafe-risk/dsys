package raft

import (
	"errors"

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

	Role    Role
	State   raftproto.State
	Storage RaftStorage
}

var ErrInvalidConsensusGroup = errors.New("invalid consensus group")

func (g *Group) HandleAppendEntries(a *raftproto.AppendEntriesRequest) (*raftproto.AppendEntriesResponse, error) {
	// Check that the group is valid.
	if a.ConsensusGroup != g.GroupID {
		return nil, ErrInvalidConsensusGroup
	}

	// Term check.
	if a.Term < g.State.Persistent.Term {
		return &raftproto.AppendEntriesResponse{
			ConsensusGroup: g.GroupID,
			Term:           g.State.Persistent.Term,
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
			Term:           g.State.Persistent.Term,
			Success:        false,
		}, nil
	}

	if len(a.Entries) > 0 {
		err = g.Storage.SetBatch(a.Entries)
		if err != nil {
			return nil, err
		}

		if a.LeaderCommit > g.State.Volatile.CommitIndex {
			g.State.Volatile.CommitIndex = min(g.State.Volatile.CommitIndex, a.Entries[len(a.Entries)-1].Index)
		}
	}

	// TODO: RESET TIMEOUT

	return &raftproto.AppendEntriesResponse{
		ConsensusGroup: g.GroupID,
		Term:           g.State.Persistent.Term,
		Success:        true,
	}, nil
}

func (g *Group) HandleRequestVote(a *raftproto.RequestVoteRequest) (*raftproto.RequestVoteResponse, error) {
	// Check that the group is valid.
	if a.ConsensusGroup != g.GroupID {
		return nil, ErrInvalidConsensusGroup
	}

	// Term check.
	if a.Term < g.State.Persistent.Term {
		return &raftproto.RequestVoteResponse{
			ConsensusGroup: g.GroupID,
			Term:           g.State.Persistent.Term,
			VoteGranted:    false,
		}, nil
	}

	// Check If Vote Already Granted.
	if g.State.Persistent.VotedTerm == g.State.Persistent.Term && g.State.Persistent.VotedFor != a.CandidateID {
		return &raftproto.RequestVoteResponse{
			ConsensusGroup: g.GroupID,
			Term:           g.State.Persistent.Term,
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
			Term:           g.State.Persistent.Term,
			VoteGranted:    false,
		}, nil
	}

	// Grant Vote.
	g.State.Persistent.VotedFor = a.CandidateID
	g.State.Persistent.VotedTerm = a.Term
	err = g.Storage.StoreState(&g.State)
	if err != nil {
		return nil, err
	}

	return &raftproto.RequestVoteResponse{
		ConsensusGroup: g.GroupID,
		Term:           g.State.Persistent.Term,
		VoteGranted:    true,
	}, nil
}

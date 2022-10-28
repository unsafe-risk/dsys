package raft

import (
	"errors"

	"v8.run/go/dsys/consensus/raft/raftproto"
)

type Group struct {
	GroupID uint64
	NodeID  uint64
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

	logIndex := a.PrevLogIndex
	e := raftproto.EntryFromVTPool()
	defer e.ReturnToVTPool()

	// Check that the previous log entry matches.
	err := g.Storage.Get(logIndex, e)
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

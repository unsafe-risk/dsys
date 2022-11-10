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

	// Message Queues.
	RecvQueue []raftproto.RPCMessage
	SendQueue []raftproto.RPCMessage

	// Received Votes.
	Votes uint64

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

// convertRole requires the lock to be held.
func (g *Group) convertRole(term uint64, role Role) {
	g.State.Term = term
	g.State.VotedFor = minusOne
	g.State.VotedTerm = minusOne
	g.Role = role
}

// startElection requires the lock to be held.
func (g *Group) startElection() error {
	g.convertRole(g.State.Term+1, RoleCandidate)

	// Vote for self.
	g.State.VotedFor = g.NodeID
	g.State.VotedTerm = g.State.Term
	err := g.Storage.StoreState(&g.State)
	if err != nil {
		return err
	}

	// Reset election timeout.
	g.ResetElectionTimeout()

	// Send RequestVote RPCs to all other servers.
	for i := range g.State.Nodes {
		// TODO: Optimize this.
		if g.State.Nodes[i] == g.NodeID {
			continue
		}

		e := raftproto.EntryFromVTPool()
		err := g.Storage.Last(e)
		if err != nil {
			e.ReturnToVTPool()
			return err
		}
		lastIndex := e.Index
		lastTerm := e.Term
		e.ReturnToVTPool()

		g.SendQueue = append(g.SendQueue, raftproto.RPCMessage{
			ConsensusGroup: g.GroupID,
			From:           g.NodeID,
			To:             g.State.Nodes[i],
			Timestamp:      g.GetTimeStamp(),
			Message: &raftproto.RPCMessage_RequestVoteRequest{
				RequestVoteRequest: &raftproto.RequestVoteRequest{
					ConsensusGroup: g.GroupID,
					Term:           g.State.Term,
					CandidateID:    g.NodeID,
					LastLogIndex:   lastIndex,
					LastLogTerm:    lastTerm,
				},
			},
		})
	}

	return nil
}

func (g *Group) Tick() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// If RPCMessage's Term is greater than the current term, update the current term and convert to follower.
	for i := range g.RecvQueue {
		if g.RecvQueue[i].Term > g.State.Term {
			g.convertRole(g.RecvQueue[i].Term, RoleFollower)
			g.ResetHeartbeatTimeout()
		}
	}

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

ROLE_CONV:
	switch g.Role {
	case RoleFollower:
		ht := atomic.AddUint64(&g.HeartbeatTimeout, minusOne)
		if ht <= 0 && !g.IsVoted() {
			err := g.startElection()
			if err != nil {
				return err
			}
			goto ROLE_CONV
		}

		// Respond to RPCMessages.
		for i := range g.RecvQueue {
			switch v := g.RecvQueue[i].Message.(type) {
			case *raftproto.RPCMessage_AppendEntriesRequest:
				resp, err := g.HandleAppendEntries(v.AppendEntriesRequest)
				if err != nil {
					return err
				}
				if resp.Success {
					g.ResetHeartbeatTimeout()
					g.State.Leader = g.RecvQueue[i].From
				}
				g.SendQueue = append(g.SendQueue, raftproto.RPCMessage{
					ConsensusGroup: g.GroupID,
					From:           g.NodeID,
					To:             g.RecvQueue[i].From,
					Term:           g.State.Term,
					Timestamp:      g.GetTimeStamp(),
					Message: &raftproto.RPCMessage_AppendEntriesResponse{
						AppendEntriesResponse: resp,
					},
				})
			case *raftproto.RPCMessage_RequestVoteRequest:
				resp, err := g.HandleRequestVote(v.RequestVoteRequest)
				if err != nil {
					return err
				}
				if resp.VoteGranted {
					g.ResetHeartbeatTimeout()
				}
				g.SendQueue = append(g.SendQueue, raftproto.RPCMessage{
					ConsensusGroup: g.GroupID,
					From:           g.NodeID,
					To:             g.RecvQueue[i].From,
					Term:           g.State.Term,
					Timestamp:      g.GetTimeStamp(),
					Message: &raftproto.RPCMessage_RequestVoteResponse{
						RequestVoteResponse: resp,
					},
				})
			}
		}
	case RoleCandidate:
		et := atomic.AddUint64(&g.ElectionTimeout, minusOne)
		if et <= 0 {
			err := g.startElection()
			if err != nil {
				return err
			}
		}

	case RoleLeader:
	}

	return nil
}

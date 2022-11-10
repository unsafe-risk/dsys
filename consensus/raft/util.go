package raft

import "time"

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func (g *Group) GetTimeStamp() uint64 {
	// TODO: use a monotonic clock.
	// TODO: use a hybrid logical clock.
	return uint64(time.Now().UnixNano())
}

func (g *Group) Quorum() uint64 {
	q := (uint64(len(g.State.Nodes)) / 2) + 1
	// TODO: handle non-voting replicas.
	return q
}

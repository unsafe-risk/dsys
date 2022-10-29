package raft

type MultiRaft struct {
	Groups map[uint64]*Group
}

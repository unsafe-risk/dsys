package raft

type Group struct {
	ID      uint64      `msgpack:"id"`
	Storage RaftStorage `msgpack:"-"`
}

package raft

type Config struct {
	// The ID of the node.
	NodeID uint64

	// The ID of the consensus group.
	GroupID uint64

	// Election timeout in ticks.
	ElectionTimeoutTicks uint64

	// Heartbeat timeout in ticks.
	HeartbeatTimeoutTicks uint64

	// Initial Heartbeat timeout in ticks.
	InitialHeartbeatTimeoutTicks uint64
}

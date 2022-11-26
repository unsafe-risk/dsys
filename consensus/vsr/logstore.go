package vsr

type LogStore interface {
	// Get returns the log entry at the given index.
	Get(index uint64) (LogEntry, error)
}

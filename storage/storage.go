package storage

import "errors"

var ErrUnavailable = errors.New("storage: requested entry at index is unavailable")

type Storage interface {
	// GetRange returns all keys in the range [start, end) in lexicographical order.
	// If limit is non-zero, at most limit keys will be returned.
	// If callback returns an error, the iteration will stop and the error will be returned.
	// key and value are only valid for the duration of the callback.
	GetRange(start, end []byte, limit uint64, callback func(key, value []byte) error) error

	// Write sets the value for a key.
	Write(key, value []byte) error

	// WriteBatch writes a batch of key-value pairs to the storage.
	// If the storage is not available, the error ErrUnavailable will be returned.
	WriteBatch(keys [][]byte, values [][]byte) error

	// Delete deletes the key from the storage.
	Delete(key []byte) error

	// DeleteRange deletes all keys in the range [start, end).
	DeleteRange(start, end []byte) error
}

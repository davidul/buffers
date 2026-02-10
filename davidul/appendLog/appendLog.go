package linkedlist

import (
	"encoding/binary"
	"errors"
	"sync"
	"time"
)

// keyOffsets map of offsets
// offset last offset
// memory the actual values
type AppendOnly struct {
	mu         sync.RWMutex   // Protects concurrent access to keyOffsets, offset, and memory
	keyOffsets map[uint64]int // Maps keys to their offset positions in memory
	offset     int            // Current write position in the memory buffer
	memory     []uint8        // Byte array that stores all the data
}

// Record represents a single entry in the append-only log
type Record struct {
	Key       uint64
	Timestamp uint64
	Value     []byte
}

// Iterator provides sequential access to records in the append-only log
type Iterator struct {
	log           *AppendOnly
	currentOffset int
	maxOffset     int
	currentRecord *Record
	err           error
}

func NewLog() *AppendOnly {
	return NewLogWithSize(1024)
}

// NewLogWithSize creates a new append-only log with specified size
func NewLogWithSize(size int) *AppendOnly {
	only := &AppendOnly{
		keyOffsets: make(map[uint64]int),
		offset:     0,
		memory:     make([]uint8, size),
	}

	return only
}

// HasCapacity checks if the log has enough space for a record of given size
// recordSize includes: key(8) + timestamp(8) + length(4) + value
func (A *AppendOnly) HasCapacity(valueSize int) bool {
	A.mu.RLock()
	defer A.mu.RUnlock()

	requiredSize := 20 + valueSize // 8 + 8 + 4 + valueSize
	return A.offset+requiredSize <= len(A.memory)
}

// Contains checks if a key exists in this log segment
func (A *AppendOnly) Contains(key uint64) bool {
	A.mu.RLock()
	defer A.mu.RUnlock()

	_, exists := A.keyOffsets[key]
	return exists
}

// |key (8 bytes)|timestamp (8 bytes)|valueLength (4 bytes)|value (variable)|
// Key: 8-byte unique identifier (uint64)
// Timestamp: 8-byte Unix timestamp in nanoseconds
// Value Length: 4-byte integer indicating how many bytes the value occupies
// Value: Variable-length byte array containing the actual data
func (A *AppendOnly) Add(key uint64, v []byte) uint64 {
	A.mu.Lock()
	defer A.mu.Unlock()

	// key must be unique, if it already exists, generate a new one
	if _, ok := A.keyOffsets[key]; ok {
		key = A.GenerateKey()
	}
	// |key|timestamp|valueLength|value|
	A.keyOffsets[key] = A.offset
	binary.BigEndian.PutUint64(A.memory[A.offset:], key)
	A.offset += 8
	binary.BigEndian.PutUint64(A.memory[A.offset:], uint64(time.Now().UnixNano()))
	A.offset += 8
	valLength := len(v)
	binary.BigEndian.PutUint32(A.memory[A.offset:], uint32(valLength))
	A.offset += 4
	copy(A.memory[A.offset:], v)
	A.offset += valLength

	return key
}

// |key|timestamp|valueLength|value|
func (A *AppendOnly) Get(key uint64) []uint8 {
	A.mu.RLock()
	defer A.mu.RUnlock()

	start := A.keyOffsets[key]
	end := start + 8
	//key
	//k := binary.BigEndian.Uint64(A.memory[start:end])
	start = end
	end += 8
	//timestamp
	//ts := binary.BigEndian.Uint64(A.memory[start:end])
	start = end
	end += 4
	//value length
	vLength := binary.BigEndian.Uint32(A.memory[start:end])
	e := end + int(vLength)

	//value
	value := A.memory[end:e]
	return value
}

func (A *AppendOnly) GetOffset(key uint64) int {
	A.mu.RLock()
	defer A.mu.RUnlock()

	return A.keyOffsets[key]
}

func (A *AppendOnly) GetMemory() []uint8 {
	A.mu.RLock()
	defer A.mu.RUnlock()

	return A.memory
}

func (A *AppendOnly) GenerateKey() uint64 {
	return uint64(time.Now().UnixNano())
}

// NewIterator creates a new iterator positioned at the beginning of the log
// The iterator takes a snapshot of the current log size to ensure consistent iteration
func (A *AppendOnly) NewIterator() *Iterator {
	A.mu.RLock()
	defer A.mu.RUnlock()

	return &Iterator{
		log:           A,
		currentOffset: 0,
		maxOffset:     A.offset, // Snapshot the current end position
		currentRecord: nil,
		err:           nil,
	}
}

// Next advances the iterator to the next record
// Returns true if there is a next record, false if end-of-log is reached or an error occurs
// After Next returns false, check Error() to see if iteration stopped due to an error
func (iter *Iterator) Next() bool {
	// Check if we've reached the end of the snapshot
	if iter.currentOffset >= iter.maxOffset {
		return false
	}

	// If there was a previous error, don't continue
	if iter.err != nil {
		return false
	}

	// Lock for reading the memory buffer
	iter.log.mu.RLock()
	defer iter.log.mu.RUnlock()

	// Check bounds to prevent panic
	if iter.currentOffset+20 > len(iter.log.memory) {
		iter.err = errors.New("corrupted log: insufficient data for record header")
		return false
	}

	// Decode the record at current offset
	record := &Record{}

	// Read key (8 bytes)
	record.Key = binary.BigEndian.Uint64(iter.log.memory[iter.currentOffset : iter.currentOffset+8])
	iter.currentOffset += 8

	// Read timestamp (8 bytes)
	record.Timestamp = binary.BigEndian.Uint64(iter.log.memory[iter.currentOffset : iter.currentOffset+8])
	iter.currentOffset += 8

	// Read value length (4 bytes)
	valueLength := binary.BigEndian.Uint32(iter.log.memory[iter.currentOffset : iter.currentOffset+4])
	iter.currentOffset += 4

	// Check bounds for value
	if iter.currentOffset+int(valueLength) > len(iter.log.memory) {
		iter.err = errors.New("corrupted log: insufficient data for record value")
		return false
	}

	// Read value (variable length)
	record.Value = make([]byte, valueLength)
	copy(record.Value, iter.log.memory[iter.currentOffset:iter.currentOffset+int(valueLength)])
	iter.currentOffset += int(valueLength)

	iter.currentRecord = record
	return true
}

// Value returns the current record
// Should only be called after Next() returns true
func (iter *Iterator) Value() *Record {
	return iter.currentRecord
}

// Error returns any error that occurred during iteration
func (iter *Iterator) Error() error {
	return iter.err
}
